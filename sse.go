package mx

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"iter"
	"net/http"
	"strings"
	"sync"

	"github.com/domonda/go-errs"
)

// SSEEventError is the event name [SSEResponse.SendError] emits. It is the
// in-band error contract of an SSE stream: once the response headers are
// flushed the status is committed, so a later failure can no longer be reported
// as a 500 and has to travel as an event the client binds to.
//
// With htmx's sse extension a client subscribes to it like any other event, for
// example hx.SSESwap(mx.SSEEventError) into an alert region. That covers
// server-reported failures only — transport failures (a dropped connection, an
// unparsable stream) are raised by htmx itself as [hx.EventSSEError], and a
// complete UI handles both.
const SSEEventError = "error"

// sseKeepaliveFrame is an SSE comment (a line starting with ':'), which every
// client ignores. Writing one periodically keeps a proxy from closing the
// connection as idle.
const sseKeepaliveFrame = ": ping\n\n"

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
// An SSEResponse is safe for concurrent use: every frame is written under a
// mutex, so a producer goroutine and a [SSEResponse.Keepalive] ticker cannot
// interleave into a corrupt frame.
type SSEResponse struct {
	mtx     sync.Mutex
	writer  http.ResponseWriter
	ctrl    *http.ResponseController
	factory WriterFactory
	closed  bool
}

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
// Calling it commits the 200 status — see the [SSEResponse] docs on what that
// means for error reporting.
func NewSSEResponse(w http.ResponseWriter, factory WriterFactory) (*SSEResponse, error) {
	if !canFlush(w) {
		return nil, errs.Errorf("mx: http.ResponseWriter of type %T can't flush, which an SSE response requires", w)
	}
	if factory == nil {
		factory = DefaultWriterFactory
	}
	header := w.Header()
	header.Set("Content-Type", ContentTypeEventStream)
	header.Set("Cache-Control", "no-cache")
	header.Set("Connection", "keep-alive")
	header.Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	ctrl := http.NewResponseController(w)
	if err := ctrl.Flush(); err != nil {
		return nil, errs.Errorf("mx: flushing the SSE response headers: %w", err)
	}
	return &SSEResponse{writer: w, ctrl: ctrl, factory: factory}, nil
}

// canFlush reports whether w, or any http.ResponseWriter it wraps, implements
// http.Flusher. http.ResponseController walks the same Unwrap chain, but only
// reports http.ErrNotSupported from Flush — by then the headers are written and
// there is no 500 left to send, so [NewSSEResponse] probes up front instead.
func canFlush(w http.ResponseWriter) bool {
	for {
		switch t := w.(type) {
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
func (r *SSEResponse) Send(ctx context.Context, event string, comp Component) error {
	if strings.ContainsAny(event, "\r\n") {
		// A malformed event name is a caller mistake detected up front,
		// not a runtime failure, so it carries no callstack.
		return fmt.Errorf("mx: SSE event name must not contain a line break, got %q", event)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var buf bytes.Buffer
	if comp != nil {
		if err := comp.Render(ctx, r.factory.NewWriter(&buf)); err != nil {
			return err
		}
	}
	return r.writeFrame(event, buf.Bytes())
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
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return nil
	}
	message := "Internal Server Error"
	if RevealInternalServerErrors {
		message = err.Error()
	}
	return r.Send(context.WithoutCancel(ctx), SSEEventError, Text(message))
}

// Keepalive writes an SSE comment and flushes it. Clients ignore comments, so it
// changes nothing in the DOM; its only job is to put bytes on the wire, which
// stops a proxy (or a load balancer) from dropping the connection as idle.
// Call it periodically from a ticker whenever events can be far apart.
func (r *SSEResponse) Keepalive() error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	return r.writeAndFlush([]byte(sseKeepaliveFrame))
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

// writeFrame assembles the complete SSE frame for a rendered event and writes
// it in one call, so a write error can never leave a partial frame on the wire.
func (r *SSEResponse) writeFrame(event string, data []byte) error {
	var frame bytes.Buffer
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

	r.mtx.Lock()
	defer r.mtx.Unlock()

	return r.writeAndFlush(frame.Bytes())
}

// writeAndFlush writes a complete frame and flushes it. The caller must hold mtx.
func (r *SSEResponse) writeAndFlush(frame []byte) error {
	if r.closed {
		return errs.New("mx: SSEResponse is closed")
	}
	if _, err := r.writer.Write(frame); err != nil {
		return err
	}
	return r.ctrl.Flush()
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
