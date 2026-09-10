// Command example-logstream simulates a production log and streams it into a
// [logview.View], to dogfood the log renderer and the [mx.SSEResponse]
// lifecycle together.
//
// Unlike cmd/example-sse, which replays a fixed script and ends, this one runs
// indefinitely: a generator goroutine produces randomly composed log lines at
// irregular intervals — exponentially distributed gaps, occasional bursts of
// correlated lines, occasional quiet stretches of several seconds — because a
// log that arrives on a fixed tick exercises none of the things a real one
// does. The generated lines cover every rendering path: JSON records with each
// value type, nested objects, arrays, a level the theme does not define, plain
// text lines with and without a level word, and the occasional Go panic with a
// stack trace.
//
//	go run ./cmd/example-logstream
//
// Then browse to http://localhost:8080 and try the surface:
//
//   - Press Split to put a second viewer beside the first. Each opens its own
//     connection to the same stream, so filtering or pausing one leaves the
//     other streaming — two subscribers, one source.
//   - Type in the filter to hide the lines that do not match and highlight the
//     match in the ones that do. It filters the lines already received, not the
//     backlog behind the stream.
//   - Scroll up to read something. That pauses on its own: lines keep arriving
//     and are held back, counted by the badge, and the button becomes Resume.
//   - Press Pause instead to hold the stream without leaving the bottom.
//   - Scroll back to the bottom, or press Resume, to release the held lines in
//     order and follow again. A pause the button caused only the button undoes.
//   - Stop the server and start it again. The browser reconnects on its own and
//     sends back the last id it saw, so the log continues instead of repeating
//     the lines it already has.
//
// Two flags show what happens when a stream does not simply keep running:
//
//   - -fail reports an error in band after a few lines instead of continuing,
//     because by then the 200 status is committed and there is no 500 left.
//   - -rate sets the mean number of lines per second, to see the view under a
//     firehose or at a trickle.
package main

import (
	"context"
	"flag"
	"log"
	"math/rand/v2"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/domonda/go-errs"
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/hx"
	"github.com/ungerik/go-mx/logview"
)

const (
	// eventDone is the event name wired to hx.SSEClose, so the client stops
	// reconnecting when the stream ends for good. SSE has no close frame.
	eventDone = "done"
	// ringSize is how many lines the simulated log keeps for readers that
	// reconnect. A real one would tail a file or query a store; the property
	// that matters is the same, that a sequence number can be answered with
	// "everything after this".
	ringSize = 500
	// backlogLines is how much of the log a fresh reader is shown before the
	// live lines start, the way `tail -n` opens with the recent past.
	backlogLines = 40
	// keepaliveInterval has to be below the shortest idle timeout on the path —
	// proxy, load balancer and browser each have one. A log stream is idle most
	// of the time, which is what makes this mandatory rather than an
	// optimization.
	keepaliveInterval = 10 * time.Second
	// failAfter is how many lines -fail streams before reporting an error.
	failAfter = 8
)

// view configures the renderer. MessageKey is set because the generated records
// are shaped like log/slog's, where the message field is "msg". Height is left
// at its default because the page below makes the view a flex item filling the
// window, which overrides it.
var view = &logview.Config{
	MessageKey: "msg",
	MaxLines:   400,
	Theme:      themeWithFatalIcon(),
}

// themeWithFatalIcon demonstrates [logview.LevelStyle.Image]: the two levels
// that should stop the reader get an icon instead of their text. The image is a
// data URL so this example stays one file with no assets to serve.
func themeWithFatalIcon() logview.Theme {
	theme := logview.DarkTheme
	levels := make(map[string]logview.LevelStyle, len(theme.Levels))
	for name, style := range theme.Levels {
		levels[name] = style
	}
	const icon = "data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCAxNiAxNiI+PGNpcmNsZSBjeD0iOCIgY3k9IjgiIHI9IjciIGZpbGw9IiNmODUxNDkiLz48cGF0aCBkPSJNOCA0djVNOCAxMXYxIiBzdHJva2U9IiNmZmYiIHN0cm9rZS13aWR0aD0iMiIvPjwvc3ZnPg=="
	for _, name := range []string{"fatal", "panic"} {
		style := levels[name]
		style.Image = icon
		levels[name] = style
	}
	theme.Levels = levels
	return theme
}

// entry is one produced log line and the sequence number that addresses it.
type entry struct {
	seq  int64
	text string
}

// source is the simulated log: a generator, a bounded history for readers that
// reconnect, and a set of live subscribers.
type source struct {
	mtx  sync.Mutex
	ring []entry
	next int64
	subs map[chan entry]struct{}
}

func newSource() *source {
	return &source{subs: make(map[chan entry]struct{})}
}

// publish appends a line to the history and hands it to every live subscriber.
func (s *source) publish(text string) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	e := entry{seq: s.next, text: text}
	s.next++
	s.ring = append(s.ring, e)
	if len(s.ring) > ringSize {
		s.ring = s.ring[len(s.ring)-ringSize:]
	}
	for ch := range s.subs {
		select {
		case ch <- e:
		default:
			// A reader too slow to keep up loses lines rather than stalling the
			// generator, which is what a real log does too. It notices on its
			// next reconnect, when the gap in the ids shows.
		}
	}
}

// subscribe returns the lines after seq that are still in the history, plus a
// channel of everything published from now on. Both are taken under one lock,
// so no line can fall between them or arrive on both.
func (s *source) subscribe(after int64, all bool) (backlog []entry, live <-chan entry, cancel func()) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	for _, e := range s.ring {
		if all && e.seq <= after {
			continue
		}
		backlog = append(backlog, e)
	}
	if !all && len(backlog) > backlogLines {
		backlog = backlog[len(backlog)-backlogLines:]
	}

	ch := make(chan entry, 256)
	s.subs[ch] = struct{}{}
	return backlog, ch, func() {
		s.mtx.Lock()
		defer s.mtx.Unlock()
		delete(s.subs, ch)
	}
}

// run generates lines until ctx is done. The pacing is the point: gaps are
// exponentially distributed around the mean rather than fixed, a burst emits
// several correlated lines at once the way one handled request does, and every
// so often the log goes quiet for seconds — which is when the keepalive earns
// its place.
func (s *source) run(ctx context.Context, rate float64, rnd *rand.Rand) {
	for {
		for _, line := range generate(rnd) {
			s.publish(line)
		}
		gap := time.Duration(rnd.ExpFloat64() / rate * float64(time.Second))
		if rnd.IntN(20) == 0 {
			gap += time.Duration(2+rnd.IntN(5)) * time.Second
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(gap):
		}
	}
}

// page is the initial document, with panes independent log views side by side.
// Every log line arrives over the stream.
func page(panes int) mx.Component {
	views := make(mx.Components, panes)
	for i := range views {
		// Each pane opens its own connection to the same stream: its own
		// hx-ext, its own sse-connect, its own EventSource. That is the point of
		// splitting — two subscribers reading one source, each with its own
		// backlog, filter and pause. Sharing one connection between them would
		// need a single sse-connect ancestor, and then pausing would be a
		// property of the page rather than of a viewer.
		views[i] = view.View(
			hx.Ext("sse"),
			hx.SSEConnect("/stream"),
			hx.SSEClose(eventDone),
		)
	}
	return mx.Components{
		mx.Raw("<!DOCTYPE html>"),
		html.HTML(html.Lang("en"),
			html.Head(
				html.Meta(html.CharSet("utf-8")),
				html.TitleElem("go-mx log stream example"),
				hx.ScriptFromCDN,
				hx.ScriptSSEFromCDN,
				// The one stylesheet the view needs. The scroll area's own
				// classes are Tailwind, which this example deliberately has no
				// build for, so it scrolls but does not get its thin scrollbar.
				view.StyleElement(),
				// The log fills the window: a full-height flex column with the
				// panes as the growing item, laid out in a row. Because a view
				// is itself a flex column, its scroll area then takes whatever
				// the toolbar and the error sink leave — no height in the
				// Config needed.
				html.StyleElem( /*css*/ `
html, body { height: 100% }
body { margin: 0; display: flex; flex-direction: column; font: 14px/1.5 system-ui, sans-serif }
header { flex: none; display: flex; align-items: center; gap: 1rem; padding: 0.5rem 0.875rem; border-bottom: 1px solid #d8dee4 }
header div { flex: 1; min-width: 0 }
h1 { margin: 0; font-size: 0.9375rem }
p { margin: 0; color: #57606a; font-size: 0.8125rem }
a.split { flex: none; padding: 0.25rem 0.75rem; border: 1px solid #d8dee4; border-radius: 4px; color: inherit; text-decoration: none; font-size: 0.8125rem; white-space: nowrap }
a.split:hover { background: #f3f4f6 }
main { flex: 1; min-height: 0; display: flex; gap: 1px; background: #d8dee4 }
.log-view { flex: 1; min-width: 0; min-height: 0; border-radius: 0 }
`),
			),
			html.Body(
				html.Header(
					html.Div(
						html.H1("go-mx log stream example"),
						html.P("A simulated production log. Filter it, pause it, scroll it."),
					),
					splitLink(panes),
				),
				html.Main(views),
			),
		),
	}
}

// splitLink toggles the number of panes. It is a link rather than a button
// because it navigates: each pane has to open its own connection, and building
// the page again is the shortest way to say that without a second route and an
// out-of-band swap to keep the control in step.
func splitLink(panes int) mx.Component {
	if panes > 1 {
		return html.A(html.Class("split"), html.HRef("/"), "Unsplit")
	}
	return html.A(html.Class("split"), html.HRef("/?panes=2"), "Split")
}

// resumeAfter returns the sequence number to continue after, given the id the
// client acknowledged, and whether it recognized one.
//
// The value is echoed back by the client, so it is client-controlled input: an
// unparsable one means "this reader is new", not an error.
func resumeAfter(lastEventID string) (int64, bool) {
	seq, err := strconv.ParseInt(lastEventID, 10, 64)
	if err != nil || seq < 0 {
		return 0, false
	}
	return seq, true
}

// streamHandler serves the log. Everything that could fail with a 500 has to
// happen before NewSSEResponse, which commits the 200 status.
func streamHandler(src *source, fail bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Read this before creating the response: it is the only thing that
		// tells a reconnect from a fresh reader.
		after, resuming := resumeAfter(mx.LastEventID(r))

		backlog, live, unsubscribe := src.subscribe(after, resuming)
		defer unsubscribe()

		sse, err := mx.NewSSEResponse(w, nil)
		if err != nil {
			// Still a clean 500: NewSSEResponse checks it can flush before
			// writing anything.
			mx.RespondNonContextError(w, err)
			return
		}
		defer sse.Close()

		ctx := r.Context()
		// An idle log stream is the normal case, so this is what keeps the
		// connection from being dropped by whatever is in front of it.
		go sse.KeepaliveLoop(ctx, keepaliveInterval) //nolint:errcheck // ends with ctx
		if err := sse.SetRetry(time.Second); err != nil {
			return
		}

		// sent counts log lines rather than events, so -fail fires after the same
		// amount of log whether it arrived as a backlog batch or live.
		sent := len(backlog)

		// The backlog goes out as one event rather than one per line: htmx
		// inserts the element children of a fragment individually, so a batch
		// costs one flush and one swap instead of forty.
		if len(backlog) > 0 {
			texts := make([]string, len(backlog))
			for i, e := range backlog {
				texts[i] = e.text
			}
			event := mx.SSEEvent{
				Name: view.EventName(),
				ID:   strconv.FormatInt(backlog[len(backlog)-1].seq, 10),
				Comp: view.Lines(texts...),
			}
			if err := sse.SendEvent(ctx, event); err != nil {
				return
			}
		}

		for ; ; sent++ {
			if fail && sent >= failAfter {
				// A 200 and a few lines are already on the wire, so this can
				// only travel in band. It lands in the view's error sink.
				_ = sse.SendError(ctx, errs.New("log shipper lost its connection to the collector"))
				// Close the stream for good, or the browser reconnects and
				// re-reports the same error forever.
				_ = sse.Send(ctx, eventDone, nil)
				return
			}
			// A canceled context means the client navigated away, reloaded or
			// lost the connection, which is the normal way this loop ends.
			select {
			case <-ctx.Done():
				return
			case e := <-live:
				event := mx.SSEEvent{
					Name: view.EventName(),
					ID:   strconv.FormatInt(e.seq, 10),
					Comp: view.Line(e.text),
				}
				if err := sse.SendEvent(ctx, event); err != nil {
					return
				}
			}
		}
	}
}

// pageHandler serves the one-pane and two-pane documents. Both are built once:
// a page is a component, and nothing in it depends on the request.
func pageHandler() http.HandlerFunc {
	header := http.Header{"Content-Type": {mx.ContentTypeHTML}}
	single := mx.ComponentHTTPHandler(page(1), mx.DefaultWriterFactory, header)
	split := mx.ComponentHTTPHandler(page(2), mx.DefaultWriterFactory, header)
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("panes") == "2" {
			split(w, r)
			return
		}
		single(w, r)
	}
}

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	rate := flag.Float64("rate", 3, "mean log lines per second")
	fail := flag.Bool("fail", false, "report an in-band error after a few lines instead of continuing")
	seed := flag.Uint64("seed", 1, "generator seed")
	flag.Parse()

	if *rate <= 0 {
		log.Fatal("-rate must be positive")
	}

	src := newSource()
	go src.run(context.Background(), *rate, rand.New(rand.NewPCG(*seed, *seed+1)))

	mux := http.NewServeMux()
	mux.HandleFunc("/stream", streamHandler(src, *fail))
	mux.HandleFunc("/", pageHandler())

	log.Printf("listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}
