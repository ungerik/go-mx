package main

import (
	"bufio"
	"context"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/logview"
)

// collect reads the raw wire bytes of a stream until want frames have arrived,
// then disconnects. The stream never ends on its own, so a test has to say how
// much of it it wants rather than reading to EOF.
func collect(t *testing.T, src *source, fail bool, lastEventID string, want int) string {
	t.Helper()
	return collectQuery(t, src, fail, "", lastEventID, want)
}

// collectQuery is collect with a query string, such as the search parameter
// the view adds when its filter changes.
func collectQuery(t *testing.T, src *source, fail bool, query, lastEventID string, want int) string {
	t.Helper()

	srv := httptest.NewServer(streamHandler(src, fail))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+query, nil)
	if err != nil {
		t.Fatal(err)
	}
	if lastEventID != "" {
		req.Header.Set(mx.HeaderLastEventID, lastEventID)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if got := resp.Header.Get("Content-Type"); got != mx.ContentTypeEventStream {
		t.Errorf("Content-Type is %q, want %q", got, mx.ContentTypeEventStream)
	}

	var b strings.Builder
	r := bufio.NewReader(resp.Body)
	for frames := 0; frames < want; {
		line, err := r.ReadString('\n')
		b.WriteString(line)
		if err != nil {
			break
		}
		if line == "\n" {
			frames++
		}
	}
	return b.String()
}

// sourceWith returns a source pre-filled with lines, and no generator running,
// so what the handler sends is exactly what the test put in.
func sourceWith(lines ...string) *source {
	src := newSource()
	for _, line := range lines {
		src.publish(line)
	}
	return src
}

func TestStreamSendsTheBacklogAsOneEvent(t *testing.T) {
	// A fresh reader opens with the recent past, the way `tail -n` does. It goes
	// out as one event because htmx inserts a fragment's element children
	// individually — so a batch costs one flush and one swap instead of forty.
	src := sourceWith(`{"level":"info","msg":"one"}`, `{"level":"warn","msg":"two"}`)
	out := collect(t, src, false, "", 2) // retry frame, then the backlog

	if got := strings.Count(out, "event: "+logview.DefaultEvent+"\n"); got != 1 {
		t.Errorf("backlog arrived as %d events, want 1:\n%s", got, out)
	}
	for _, want := range []string{
		"retry: 1000\n",
		"id: 1\n", // the id of the last line in the batch
		`class="log-line"`,
		">one<",
		">two<",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestStreamFramesAreWellFormed(t *testing.T) {
	// Every line of an SSE response has to be a field or a frame terminator. A
	// stray line would make the client drop the event, or worse, misread the
	// next one.
	src := sourceWith(`{"level":"error","msg":"boom","stack":"a\n\tb\n\tc"}`)
	out := collect(t, src, false, "", 2)

	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}
		switch {
		case strings.HasPrefix(line, "data:"),
			strings.HasPrefix(line, "event: "),
			strings.HasPrefix(line, "id: "),
			strings.HasPrefix(line, "retry: "),
			strings.HasPrefix(line, ": "):
		default:
			t.Errorf("stray line %q in:\n%s", line, out)
		}
	}
	// A multi-line value survives as several data lines, which the client
	// rejoins — that is why the <pre> keeps its indentation across the wire.
	if !strings.Contains(out, "data: \tb") {
		t.Errorf("the stack trace's indentation did not survive:\n%s", out)
	}
}

func TestStreamResumesFromLastEventID(t *testing.T) {
	// Without this a reconnect re-appends everything the reader already has,
	// because the lines are appended, not replaced. That is the failure a
	// stream without event ids has on every dropped connection.
	src := sourceWith(
		`{"msg":"zero"}`, `{"msg":"one"}`, `{"msg":"two"}`, `{"msg":"three"}`,
	)
	out := collect(t, src, false, "1", 2)

	for _, want := range []string{">two<", ">three<"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q after resuming from id 1:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{">zero<", ">one<"} {
		if strings.Contains(out, unwanted) {
			t.Errorf("resuming from id 1 replayed %q:\n%s", unwanted, out)
		}
	}
}

func TestStreamIgnoresAnUnusableLastEventID(t *testing.T) {
	// The id is echoed back by the client, so it is client-controlled input. An
	// unparsable one means "this reader is new", not an error.
	src := sourceWith(`{"msg":"zero"}`, `{"msg":"one"}`)
	for _, id := range []string{"not-a-number", "-5", "../../etc/passwd"} {
		out := collect(t, src, false, id, 2)
		if !strings.Contains(out, ">zero<") {
			t.Errorf("id %q did not fall back to a fresh read:\n%s", id, out)
		}
	}
}

func TestStreamLiveLinesFollowTheBacklog(t *testing.T) {
	// The backlog snapshot and the live subscription are taken under one lock,
	// so a line published in between can neither be lost nor sent twice.
	src := sourceWith(`{"msg":"backlog"}`)

	srv := httptest.NewServer(streamHandler(src, false))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	r := bufio.NewReader(resp.Body)
	readFrame := func() string {
		t.Helper()
		var b strings.Builder
		for {
			line, err := r.ReadString('\n')
			if err != nil {
				t.Fatalf("reading a frame: %v\ngot so far:\n%s", err, b.String())
			}
			b.WriteString(line)
			if line == "\n" {
				return b.String()
			}
		}
	}

	readFrame() // retry
	if got := readFrame(); !strings.Contains(got, ">backlog<") {
		t.Fatalf("first event is not the backlog: %s", got)
	}
	src.publish(`{"msg":"live"}`)
	live := readFrame()
	if !strings.Contains(live, ">live<") {
		t.Errorf("a line published after subscribing did not arrive: %s", live)
	}
	if !strings.Contains(live, "id: 1\n") {
		t.Errorf("the live line carries no resumable id: %s", live)
	}
}

func TestStreamSearchAnswersWithTheMatchingHistory(t *testing.T) {
	// A search reaches past the backlog a fresh reader is shown: the match is
	// older than the last backlogLines lines, which is the line a client-side
	// filter could never have found.
	lines := []string{`{"level":"error","msg":"panic: nil map"}`}
	for i := range backlogLines + 5 {
		lines = append(lines, `{"level":"info","msg":"line `+strconv.Itoa(i)+`"}`)
	}
	src := sourceWith(lines...)

	if out := collect(t, src, false, "", 2); strings.Contains(out, "panic: nil map") {
		t.Fatalf("the test needs the match outside the plain backlog, but it was sent:\n%s", out)
	}
	out := collectQuery(t, src, false, "?q=PANIC", "", 2)
	if !strings.Contains(out, "panic: nil map") {
		t.Errorf("the search did not find the old match:\n%s", out)
	}
	if strings.Contains(out, ">line ") {
		t.Errorf("the search sent lines that do not match:\n%s", out)
	}
	// The id is the matching line's own sequence number, so a reconnect of the
	// search resumes after it.
	if !strings.Contains(out, "id: 0\n") {
		t.Errorf("the search result does not carry the match's id:\n%s", out)
	}
}

func TestStreamSearchMatchesTheRenderedLine(t *testing.T) {
	// The reader searches for what the view shows, "status=500", not for the
	// raw JSON, which reads "status":500.
	src := sourceWith(`{"msg":"request finished","status":500}`, `{"msg":"request finished","status":200}`)
	out := collectQuery(t, src, false, "?q=status%3D500", "", 2)
	if !strings.Contains(out, ">500<") || strings.Contains(out, ">200<") {
		t.Errorf("the search did not match on the rendered text:\n%s", out)
	}
}

func TestStreamSearchResumes(t *testing.T) {
	// A browser reconnect of a search keeps the parameter and sends the last id,
	// and gets the matches after it, not the ones it already has.
	src := sourceWith(`{"msg":"hit one"}`, `{"msg":"miss"}`, `{"msg":"hit two"}`)
	out := collectQuery(t, src, false, "?q=hit", "0", 2)
	if !strings.Contains(out, "hit two") || strings.Contains(out, "hit one") || strings.Contains(out, ">miss<") {
		t.Errorf("resuming a search sent the wrong lines:\n%s", out)
	}
}

func TestStreamSearchFiltersLiveLines(t *testing.T) {
	src := sourceWith(`{"msg":"hit backlog"}`)

	srv := httptest.NewServer(streamHandler(src, false))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"?q=hit", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	r := bufio.NewReader(resp.Body)
	readFrame := func() string {
		t.Helper()
		var b strings.Builder
		for {
			line, err := r.ReadString('\n')
			if err != nil {
				t.Fatalf("reading a frame: %v\ngot so far:\n%s", err, b.String())
			}
			b.WriteString(line)
			if line == "\n" {
				return b.String()
			}
		}
	}

	readFrame() // retry
	readFrame() // the backlog
	src.publish(`{"msg":"miss live"}`)
	src.publish(`{"msg":"hit live"}`)
	got := readFrame()
	if !strings.Contains(got, "hit live") {
		t.Errorf("the next event is not the matching live line: %s", got)
	}
	if strings.Contains(got, "miss live") {
		t.Errorf("a live line that does not match was sent: %s", got)
	}
}

func TestStreamFailureTravelsInBandAndCloses(t *testing.T) {
	// By the time this fails, a 200 and several lines are on the wire, so the
	// failure can only be an event. And it has to close the stream: a browser
	// reconnects on its own, and would re-report the same error forever.
	lines := make([]string, failAfter)
	for i := range lines {
		lines[i] = `{"msg":"line ` + strconv.Itoa(i) + `"}`
	}
	out := collect(t, sourceWith(lines...), true, "", 4)

	errIndex := strings.Index(out, "event: "+mx.SSEEventError+"\n")
	doneIndex := strings.Index(out, "event: "+eventDone+"\n")
	if errIndex < 0 || doneIndex < 0 {
		t.Fatalf("missing the error or close event:\n%s", out)
	}
	if errIndex > doneIndex {
		t.Errorf("the close event came before the error:\n%s", out)
	}
	// mx.RevealInternalServerErrors is off by default, so the message must not
	// reach the client.
	if strings.Contains(out, "collector") {
		t.Errorf("the internal error message leaked to the client:\n%s", out)
	}
}

func TestGeneratedLinesCoverEveryRenderingPath(t *testing.T) {
	// The example exists to exercise the renderer, so the generator has to
	// actually produce each shape rather than only the common one.
	rnd := rand.New(rand.NewPCG(1, 2))
	var all strings.Builder
	for range 400 {
		for _, line := range generate(rnd) {
			all.WriteString(render(t, view.Line(line)))
		}
	}
	out := all.String()
	for _, want := range []string{
		"log-string", "log-number", "log-bool", "log-null", // every value type
		"log-time", "log-message", "log-key",
		"log-punct\">{", // a nested object
		"log-punct\">[", // an array
		"log-multiline", // a stack trace
		"log-text",      // a plain-text line
		"log-level-info", "log-level-warn", "log-level-error", "log-level-debug",
		"log-level-undefined",            // a level the theme does not define
		"<img class=\"log-level-image\"", // a level rendered as an icon
	} {
		if !strings.Contains(out, want) {
			t.Errorf("the generator never produced %q", want)
		}
	}
}

func TestGeneratedLinesAreParsableRecords(t *testing.T) {
	// A generated JSON line that does not parse would silently fall back to the
	// plain-text path and quietly stop testing what it was written to test.
	rnd := rand.New(rand.NewPCG(7, 11))
	for range 400 {
		for _, line := range generate(rnd) {
			if !strings.HasPrefix(line, "{") {
				continue
			}
			if out := render(t, view.Line(line)); !strings.Contains(out, "log-key") {
				t.Fatalf("a generated JSON line did not render as a record:\n%s\n%s", line, out)
			}
		}
	}
}

func TestPageWiresTheViewToTheStream(t *testing.T) {
	out := render(t, page(1))
	for _, want := range []string{
		`hx-ext="sse"`,
		`sse-connect="/stream"`,
		`sse-close="` + eventDone + `"`,
		`sse-swap="` + logview.DefaultEvent + `"`,
		`sse-swap="` + mx.SSEEventError + `"`,
		`data-mx-log-view=""`,
		`data-mx-log-search="q"`, // the filter is a search the handler answers
		`data-stick-to-bottom=""`,
		".log-line[data-mx-log-pending]", // the stylesheet the view needs
		"window.mxLogView",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in the page", want)
		}
	}
	if !strings.Contains(out, `href="/?panes=2"`) {
		t.Errorf("the page offers no way to split: %s", out)
	}
}

func TestSplitPaneSubscribesIndependently(t *testing.T) {
	// The point of splitting is two subscribers reading one source. Sharing a
	// connection would mean one sse-connect above both panes, and then pausing
	// or filtering would be a property of the page rather than of a viewer.
	out := render(t, page(2))
	for what, want := range map[string]string{
		"log view":     `data-mx-log-view=""`,
		"connection":   `sse-connect="/stream"`,
		"extension":    `hx-ext="sse"`,
		"line target":  `data-mx-log-lines=""`,
		"error sink":   `sse-swap="` + mx.SSEEventError + `"`,
		"filter input": `data-mx-log-filter=""`,
	} {
		if got := strings.Count(out, want); got != 2 {
			t.Errorf("split page has %d of the %s (%q), want 2", got, what, want)
		}
	}
	// The script defines itself once per instance behind its own guard, and
	// each instance wires only its own root.
	if got := strings.Count(out, "if(!window.mxLogView)"); got != 2 {
		t.Errorf("script guard appears %d times, want 2", got)
	}
	if !strings.Contains(out, `href="/"`) {
		t.Errorf("the split page offers no way back: %s", out)
	}
}

func TestResumeAfter(t *testing.T) {
	for _, tc := range []struct {
		id    string
		want  int64
		found bool
	}{
		{"", 0, false},
		{"0", 0, true},
		{"41", 41, true},
		{"-1", 0, false},
		{"nope", 0, false},
		{strconv.FormatInt(1<<62, 10), 1 << 62, true},
	} {
		got, found := resumeAfter(tc.id)
		if got != tc.want || found != tc.found {
			t.Errorf("resumeAfter(%q) = (%d, %v), want (%d, %v)", tc.id, got, found, tc.want, tc.found)
		}
	}
}

// render renders a component and fails the test on any render error.
func render(t *testing.T, c mx.Component) string {
	t.Helper()
	var b strings.Builder
	if err := c.Render(context.Background(), mx.NewCheckedWriter(&b)); err != nil {
		t.Fatalf("render error: %v\npartial output:\n%s", err, b.String())
	}
	return b.String()
}
