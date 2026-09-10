package mx

import (
	"bytes"
	"context"
	"fmt"
	"iter"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/domonda/go-errs"
)

// SSEEventError is the event name [SSEResponse.SendError] emits. It is the
// in-band error contract of an SSE stream: once the response headers are
// flushed the status is committed, so a later failure can no longer be reported
// as a 500 and has to travel as an event the client binds to.
//
// A client subscribes to it like any other named event. The name is
// deliberately not "error": a browser dispatches its own transport failures at
// the EventSource under that name, so a subscriber to "error" would also fire
// on every dropped connection — with an Event that carries no data, which
// htmx's sse extension then throws on while trying to swap it.
//
// It therefore covers server-reported failures only: a transport failure (a
// dropped connection, an unparsable stream) never reaches the stream as an
// event, so a complete UI also handles whatever its client reports for that.
// The hx package documents both halves for htmx, where it can name the symbols
// on either side.
const SSEEventError = "mx-error"

// sseKeepaliveFrame is an SSE comment (a line starting with ':'), which every
// client ignores. Writing one periodically keeps a proxy from closing the
// connection as idle. It is a []byte because it goes straight to the wire on
// every keepalive tick, with no conversion.
var sseKeepaliveFrame = []byte(": ping\n\n")

// HeaderLastEventID is the request header a client sends when it reconnects to
// a stream it was previously reading, carrying the [SSEEvent.ID] of the last
// event it received. Read it with [LastEventID].
//
// A browser reconnects on its own whenever a stream ends without the client
// having closed it — a dropped connection, a sleeping laptop, a proxy idle
// timeout — so a stream that ignores this header replays from the beginning or
// silently skips whatever was sent while the client was away.
const HeaderLastEventID = "Last-Event-ID"

// LastEventID returns the [HeaderLastEventID] value of request, or "" when the
// request is not a reconnect (or the stream never sent ids).
//
// The value is whatever the server previously put in [SSEEvent.ID], echoed back
// verbatim, so it is client-controlled input: validate it before using it to
// address anything, and treat an unknown value as "start from the beginning"
// rather than as an error.
func LastEventID(request *http.Request) string {
	return request.Header.Get(HeaderLastEventID)
}

// SSEEvent is one event for [SSEResponse.SendEvent]. Only Comp is required.
type SSEEvent struct {
	// Name is the "event:" field, which a client subscribes to by name
	// (hx.SSESwap(name) with htmx). Empty sends an unnamed event, which
	// clients receive as "message".
	Name string

	// ID is the "id:" field. Setting it makes the client remember the event
	// and send it back as [HeaderLastEventID] when it reconnects, which is
	// what lets a handler resume instead of replaying. Empty omits the field
	// and leaves the client's remembered id unchanged.
	//
	// It only has to be meaningful to the handler that reads it back — a
	// sequence number or a message id, not a hash. Choose it so that "give me
	// everything after this" is answerable.
	ID string

	// Comp is rendered into the "data:" field(s). A nil Comp sends an event
	// with empty data, which still dispatches on the client.
	Comp Component
}

// SSEResponse streams components to a client as Server-Sent Events over a
// still-open HTTP response. Where [ComponentHTTPHandler] renders one component
// into a buffer and writes it as the whole response, an SSEResponse renders one
// component per event and flushes it immediately, so a page can be updated
// after it was delivered.
//
// # Buffering
//
// go-mx buffers HTTP responses on purpose: a deferred computation (a
// context-dependent option provider, [ErrAttrib], [NewErrElement]) can fail
// mid-render, and streaming the render directly would already have sent a 200
// and a truncated page. SSEResponse keeps that guarantee, but moves it from the
// response to the event: [SSEResponse.Send] renders into a buffer and returns a
// render error before writing a single byte of the frame, so a failed event is
// never partially delivered.
//
// What genuinely changes is the response-level contract. [NewSSEResponse]
// commits the 200 status, so from then on there is no 500 left to send — do
// whatever can fail with a 500 (authentication, loading the record) before
// creating the SSEResponse, and report later failures in-band with
// [SSEResponse.SendError].
//
// # Usage
//
// The type owns the flush loop, so a caller cannot half-use it — there is no
// raw flushing [Writer] to render into directly:
//
//	sse, err := mx.NewSSEResponse(w, nil)
//	if err != nil {
//		mx.RespondNonContextError(w, err)
//		return
//	}
//	defer sse.Close()
//	for token := range tokens {
//		if err := sse.Send(r.Context(), "message", html.Span(token)); err != nil {
//			return
//		}
//	}
//
// The request context carries cancellation for a disconnected client, so
// passing it to Send ends the loop when the client goes away.
//
// # Reconnection
//
// A browser reconnects on its own whenever a stream ends without the client
// having closed it, and it does so silently. A stream that sends no event ids
// therefore replays from the beginning on every dropped connection, sleeping
// laptop or proxy idle timeout — or skips what it sent while the client was
// away, depending on how the handler resumes.
//
// To make that survivable, give each event an id with
// [SSEResponse.SendEvent] and start from [LastEventID] on the way in:
//
//	for _, msg := range conversation.Since(mx.LastEventID(r)) {
//		err := sse.SendEvent(r.Context(), mx.SSEEvent{
//			Name: "message",
//			ID:   msg.ID,
//			Comp: renderMessage(msg),
//		})
//		if err != nil {
//			return
//		}
//	}
//
// [SSEResponse.SetRetry] adjusts how long the client waits before reconnecting,
// and [SSEResponse.KeepaliveLoop] keeps an idle connection from being dropped
// in the first place.
//
// An SSEResponse is safe for concurrent use: every frame is written under a
// mutex, so a producer goroutine and a [SSEResponse.Keepalive] ticker cannot
// interleave into a corrupt frame.
type SSEResponse struct {
	mtx     sync.Mutex
	writer  http.ResponseWriter
	ctrl    *http.ResponseController
	factory WriterFactory
	closed  bool

	// writeTimeout bounds a single frame write. See SetWriteTimeout.
	writeTimeout time.Duration
}

// DefaultSSEWriteTimeout bounds how long one frame may take to reach the client
// before [SSEResponse] gives up on it. It exists because clearing the
// response-wide deadline (which [NewSSEResponse] must do — a stream never ends,
// so http.Server.WriteTimeout would cut it off) otherwise leaves a write with no
// bound at all: a client that stays connected but stops reading fills the TCP
// window, the write blocks forever holding the frame mutex, and because the
// socket is still open the request context is never canceled. That pins the
// producer goroutine and traps [SSEResponse.KeepaliveLoop] in the mutex where it
// can no longer reach its ctx.Done. A bounded write turns all of that into an
// ordinary error the caller already handles.
const DefaultSSEWriteTimeout = 10 * time.Second

// NewSSEResponse writes and flushes the Server-Sent Events response headers on
// w and returns an [SSEResponse] that renders events with a [Writer] from
// factory. A nil factory uses [DefaultWriterFactory].
//
// It fails if w cannot flush, so that surfaces at handler entry — as a plain
// error the caller can still answer with a 500 — rather than as a stream that
// is silently buffered into uselessness. It does not fail for a client that has
// already disconnected; that is reported by the first [SSEResponse.Send].
//
// The headers are Content-Type: text/event-stream, Cache-Control: no-cache,
// Connection: keep-alive and X-Accel-Buffering: no. The last one matters for
// deployments behind a reverse proxy the caller does not control: nginx (and
// several proxies that copied the header) otherwise buffers the response and
// defeats the point of streaming.
//
// It also clears the write deadline that http.Server.WriteTimeout puts on the
// response. That deadline is meant for a response which ends, and would
// otherwise cut every stream off at the timeout no matter how alive it is —
// producing exactly the symptom [SSEResponse.KeepaliveLoop] addresses, from a
// different cause, which makes it easy to misdiagnose.
//
// Calling it commits the 200 status — see the [SSEResponse] docs on what that
// means for error reporting.
func NewSSEResponse(w http.ResponseWriter, factory WriterFactory) (*SSEResponse, error) {
	if !canFlush(w) {
		return nil, errs.Errorf("mx: http.ResponseWriter of type %T can't flush, which an SSE response requires", w)
	}
	if factory == nil {
		factory = DefaultWriterFactory
	}
	ctrl := http.NewResponseController(w)
	// A stream has no length to finish within, so the response-wide write
	// deadline has to go before anything is written. http.ErrNotSupported means
	// the writer has no deadline to clear in the first place. Individual writes
	// stay bounded by [DefaultSSEWriteTimeout] — see [SSEResponse.SetWriteTimeout].
	_ = ctrl.SetWriteDeadline(time.Time{})

	header := w.Header()
	header.Set("Content-Type", ContentTypeEventStream)
	header.Set("Cache-Control", "no-cache")
	header.Set("Connection", "keep-alive")
	header.Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	// Flushing is what puts the headers on the wire and starts the stream on
	// the client. canFlush already ruled out a writer that cannot flush at all,
	// so a failure here is a write failure — the client is already gone — and
	// the 200 above is committed by now, leaving no 500 to report it with.
	// Reporting it as a constructor error would only make the caller answer a
	// committed response; the first Send reports it instead, which is where the
	// documented contract puts it and which the caller already checks.
	_ = ctrl.Flush()

	return &SSEResponse{writer: w, ctrl: ctrl, factory: factory, writeTimeout: DefaultSSEWriteTimeout}, nil
}

// canFlush reports whether w, or any http.ResponseWriter it wraps, can be
// flushed. http.ResponseController walks the same Unwrap chain, but only reports
// http.ErrNotSupported from Flush — by then the headers are written and there is
// no 500 left to send, so [NewSSEResponse] probes up front instead.
//
// The switch mirrors ResponseController.Flush in net/http/responsecontroller.go,
// including the FlushError case it prefers over http.Flusher, and has to be kept
// in sync with it: a case missing here rejects a writer that Flush would accept.
func canFlush(w http.ResponseWriter) bool {
	for {
		switch t := w.(type) {
		case interface{ FlushError() error }:
			return true
		case http.Flusher:
			return true
		case interface{ Unwrap() http.ResponseWriter }:
			w = t.Unwrap()
		default:
			return false
		}
	}
}

// Send renders comp with ctx into a buffer and writes it as one SSE event named
// event, then flushes. A nil comp sends an event with empty data, and an empty
// event name sends an unnamed event (which a client receives as "message").
//
// A render error is returned before anything is written, so a failed event
// never reaches the client half-rendered. A canceled ctx is reported without
// writing, which is how a loop notices a disconnected client.
//
// The rendered markup is re-split into one "data:" line per line of output,
// because "data:" is line-delimited in the SSE wire format: a client
// concatenates the data lines of a frame and would otherwise see the fragment
// truncated at the first newline. That matters because [Writer.Newline] and
// CheckedWriter.WithIndent both put newlines in the middle of ordinary markup.
//
// Send fails if event contains a line break, which would otherwise let a
// caller-supplied name forge additional frames.
//
// Send never sets an event id. Use [SSEResponse.SendEvent] for a stream a
// client should be able to resume after reconnecting.
func (r *SSEResponse) Send(ctx context.Context, event string, comp Component) error {
	return r.SendEvent(ctx, SSEEvent{Name: event, Comp: comp})
}

// SendEvent renders event.Comp with ctx into a buffer and writes it as one SSE
// event, then flushes. It is the full form of [SSEResponse.Send], which cannot
// set [SSEEvent.ID].
//
// Everything Send documents applies: a render error is returned before anything
// is written, a canceled ctx is reported without writing, and the rendered
// markup is re-split into one "data:" line per line of output.
//
// SendEvent fails if Name or ID contains a line break, or if ID contains a NUL,
// rather than emitting a frame the client would misparse or silently drop: a
// line break would let a caller-supplied value forge additional fields, and the
// SSE specification tells clients to ignore an id containing a NUL, which would
// break resumption invisibly.
//
// A panic while rendering is returned as an error, like [ComponentHTTPHandler]
// turns it into a 500: letting it unwind would abort the connection mid-stream
// with nothing said, and a browser answers that by reconnecting straight back
// into the same panic.
func (r *SSEResponse) SendEvent(ctx context.Context, event SSEEvent) error {
	// Malformed field values are caller mistakes detected up front,
	// not runtime failures, so they carry no callstack.
	if strings.ContainsAny(event.Name, "\r\n") {
		return fmt.Errorf("mx: SSE event name must not contain a line break, got %q", event.Name)
	}
	if strings.ContainsAny(event.ID, "\r\n") {
		return fmt.Errorf("mx: SSE event id must not contain a line break, got %q", event.ID)
	}
	if strings.ContainsRune(event.ID, 0) {
		return fmt.Errorf("mx: SSE event id must not contain a NUL, got %q", event.ID)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	data, err := renderEventData(ctx, r.factory, event.Comp)
	if err != nil {
		return err
	}
	return r.writeFrame(sseFrame(event.Name, event.ID, data))
}

// renderEventData renders comp into a buffer, converting a panic into an error
// the way [ComponentHTTPHandler] does. Rendering happens off the wire either
// way, so a failed event — panicking or not — is reported to the caller with
// nothing written, and the stream stays usable for the error it sends next.
func renderEventData(ctx context.Context, factory WriterFactory, comp Component) (data []byte, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = errs.AsErrorWithDebugStack(p)
		}
	}()
	if comp == nil {
		return nil, nil
	}
	var buf bytes.Buffer
	err = comp.Render(ctx, factory.NewWriter(&buf))
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// SendError reports err to the client as an event named [SSEEventError], the
// in-band replacement for the 500 that is no longer available once the stream
// has started. The message is escaped text, so it can be swapped into the DOM
// directly.
//
// Like [RespondNonContextError] it sends a generic "Internal Server Error"
// unless [RevealInternalServerErrors] is set, and it does nothing (returning a
// nil error) when err is a context.Canceled or context.DeadlineExceeded,
// because the client that would read it has disconnected.
//
// ctx is used only to render the message, with its cancellation stripped: a
// cancellation is often what caused err, and the report should still reach a
// client that is still reading.
func (r *SSEResponse) SendError(ctx context.Context, err error) error {
	message, report := nonContextErrorMessage(err)
	if !report {
		return nil
	}
	return r.Send(context.WithoutCancel(ctx), SSEEventError, Text(message))
}

// Keepalive writes an SSE comment and flushes it. Clients ignore comments, so it
// changes nothing in the DOM; its only job is to put bytes on the wire, which
// stops a proxy (or a load balancer) from dropping the connection as idle.
// Call it periodically from a ticker whenever events can be far apart.
func (r *SSEResponse) Keepalive() error {
	return r.writeFrame(sseKeepaliveFrame)
}

// KeepaliveLoop calls [SSEResponse.Keepalive] every interval until ctx is done
// or a write fails, and is meant to run in its own goroutine alongside the
// producer:
//
//	go sse.KeepaliveLoop(r.Context(), 20*time.Second)
//
// It returns nil when ctx is done, because a client going away is how a stream
// normally ends, and the write error otherwise. Concurrent use is safe: every
// frame is written under the same mutex, so a keepalive cannot interleave into
// an event the producer is writing.
//
// Pick an interval below the shortest idle timeout on the path — proxy, load
// balancer and browser all have one, and the shortest wins.
func (r *SSEResponse) KeepaliveLoop(ctx context.Context, interval time.Duration) error {
	if interval <= 0 {
		return fmt.Errorf("mx: SSE keepalive interval must be positive, got %s", interval)
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := r.Keepalive(); err != nil {
				return err
			}
		}
	}
}

// SetRetry tells the client how long to wait before reconnecting after the
// stream ends, overriding its default (three seconds in most browsers). It
// applies to every later reconnect until it is set again, so it is normally
// sent once, right after [NewSSEResponse].
//
// Raise it for a stream that is expensive to restart, and note that it does not
// stop a client reconnecting — only closing the stream client-side does that
// (with htmx, the event bound by hx.SSEClose).
//
// The duration is sent as whole milliseconds, so it must be at least 1ms; the
// SSE specification requires the field to be ASCII digits and tells clients to
// ignore anything else, which would make a zero or negative value a silent
// no-op.
func (r *SSEResponse) SetRetry(reconnectDelay time.Duration) error {
	millis := reconnectDelay.Milliseconds()
	if millis <= 0 {
		return fmt.Errorf("mx: SSE retry must be at least a millisecond, got %s", reconnectDelay)
	}
	return r.writeFrame([]byte("retry: " + strconv.FormatInt(millis, 10) + "\n\n"))
}

// Close marks the stream finished and flushes what is pending. It is idempotent,
// so it is safe to defer and also call explicitly, and any later Send,
// SendError or Keepalive fails rather than writing into a response the handler
// has returned from.
//
// It sends no event of its own: SSE has no close frame, and htmx's sse-close
// binds a caller-chosen event name. To make the browser close the connection,
// Send that event before calling Close. Otherwise the connection ends when the
// handler returns, and the client reconnects unless it was told not to.
func (r *SSEResponse) Close() error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if r.closed {
		return nil
	}
	r.closed = true
	return r.ctrl.Flush()
}

// sseFrame assembles the complete SSE frame for a rendered event. It is a pure
// function so the wire format can be tested directly, and so frame assembly
// happens off the mutex that [SSEResponse.writeFrame] takes.
func sseFrame(event, id string, data []byte) []byte {
	var frame bytes.Buffer
	if id != "" {
		frame.WriteString("id: ")
		frame.WriteString(id)
		frame.WriteByte('\n')
	}
	if event != "" {
		frame.WriteString("event: ")
		frame.WriteString(event)
		frame.WriteByte('\n')
	}
	for line := range sseLines(data) {
		if len(line) == 0 {
			frame.WriteString("data:\n")
			continue
		}
		frame.WriteString("data: ")
		frame.Write(line)
		frame.WriteByte('\n')
	}
	frame.WriteByte('\n')
	return frame.Bytes()
}

// writeFrame writes one complete frame and flushes it, taking mtx itself so
// every writer shares one lock discipline. The frame is written in a single
// call, so a write error can never leave a partial frame on the wire.
func (r *SSEResponse) writeFrame(frame []byte) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if r.closed {
		return errs.New("mx: SSEResponse is closed")
	}
	if r.writeTimeout > 0 {
		// Bound this frame only. A failure here is http.ErrNotSupported on a
		// writer without deadlines, which is not a reason to refuse the write.
		_ = r.ctrl.SetWriteDeadline(time.Now().Add(r.writeTimeout))
		defer func() { _ = r.ctrl.SetWriteDeadline(time.Time{}) }()
	}
	if _, err := r.writer.Write(frame); err != nil {
		return err
	}
	return r.ctrl.Flush()
}

// SetWriteTimeout bounds how long a single frame write may block before it
// fails, replacing the response-wide http.Server.WriteTimeout that
// [NewSSEResponse] has to clear. It defaults to [DefaultSSEWriteTimeout]; a
// zero or negative timeout removes the bound, which means one unresponsive
// client can pin a goroutine for the life of the process.
//
// Raise it for a slow link or large frames, lower it to reclaim stalled
// connections sooner.
func (r *SSEResponse) SetWriteTimeout(timeout time.Duration) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.writeTimeout = timeout
}

// sseLines splits data on the line terminators of the SSE wire format — CRLF,
// CR and LF — yielding one line per "data:" field. Rendered markup only ever
// contains LF, but [Raw] passes arbitrary bytes through, and an unsplit CR
// would be a line break to the client and silently truncate the fragment.
//
// It always yields at least one line, so empty data still produces a "data:"
// field and the event dispatches on the client instead of being discarded.
func sseLines(data []byte) iter.Seq[[]byte] {
	return func(yield func([]byte) bool) {
		for {
			i := bytes.IndexAny(data, "\r\n")
			if i < 0 {
				yield(data)
				return
			}
			if !yield(data[:i]) {
				return
			}
			if data[i] == '\r' && i+1 < len(data) && data[i+1] == '\n' {
				i++
			}
			data = data[i+1:]
		}
	}
}
