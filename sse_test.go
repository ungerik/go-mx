package mx

import (
	"bufio"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// nonFlushingWriter is an http.ResponseWriter that neither implements
// http.Flusher nor unwraps to one.
type nonFlushingWriter struct{ http.ResponseWriter }

// unwrappingWriter wraps a ResponseWriter the way middleware does: it hides the
// http.Flusher behind an Unwrap method, which is the chain http.ResponseController
// walks.
type unwrappingWriter struct{ inner http.ResponseWriter }

func (w unwrappingWriter) Header() http.Header         { return w.inner.Header() }
func (w unwrappingWriter) Write(b []byte) (int, error) { return w.inner.Write(b) }
func (w unwrappingWriter) WriteHeader(code int)        { w.inner.WriteHeader(code) }
func (w unwrappingWriter) Unwrap() http.ResponseWriter { return w.inner }

// newSSETest returns an SSEResponse writing into a recorder, failing the test if
// it cannot be created.
func newSSETest(t *testing.T) (*SSEResponse, *httptest.ResponseRecorder) {
	t.Helper()
	rec := httptest.NewRecorder()
	sse, err := NewSSEResponse(rec, nil)
	if err != nil {
		t.Fatalf("NewSSEResponse: %v", err)
	}
	return sse, rec
}

// body returns the response body written after the headers.
func body(rec *httptest.ResponseRecorder) string { return rec.Body.String() }

func TestNewSSEResponse_RejectsNonFlushableWriterBeforeWriting(t *testing.T) {
	// The whole point of failing here is that the caller can still answer with a
	// 500: a writer that silently buffers would turn the stream into a single
	// response delivered when the handler returns, which looks like a hang.
	rec := httptest.NewRecorder()
	sse, err := NewSSEResponse(nonFlushingWriter{rec}, nil)
	if err == nil {
		t.Fatal("NewSSEResponse accepted a non-flushable ResponseWriter; a silently buffered stream is worse than a 500")
	}
	if sse != nil {
		t.Error("NewSSEResponse returned a non-nil SSEResponse together with an error")
	}
	if rec.Code != http.StatusOK || body(rec) != "" || rec.Header().Get("Content-Type") != "" {
		t.Error("NewSSEResponse wrote to the response before the flush check, leaving the caller no clean 500")
	}
}

func TestNewSSEResponse_FindsFlusherThroughUnwrap(t *testing.T) {
	// Middleware routinely wraps the ResponseWriter. If canFlush only did a
	// direct type assertion, every wrapped handler would be rejected even though
	// http.ResponseController can flush it fine.
	rec := httptest.NewRecorder()
	if _, err := NewSSEResponse(unwrappingWriter{rec}, nil); err != nil {
		t.Fatalf("NewSSEResponse rejected a wrapped flushable writer: %v", err)
	}
}

func TestNewSSEResponse_Headers(t *testing.T) {
	_, rec := newSSETest(t)
	want := map[string]string{
		"Content-Type":      ContentTypeEventStream,
		"Cache-Control":     "no-cache",
		"Connection":        "keep-alive",
		"X-Accel-Buffering": "no",
	}
	for name, value := range want {
		if got := rec.Header().Get(name); got != value {
			t.Errorf("header %s = %q, want %q", name, got, value)
		}
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !rec.Flushed {
		// Without this flush the client's EventSource does not see the response
		// open, so a slow first event is indistinguishable from a dead server.
		t.Error("NewSSEResponse did not flush the headers")
	}
}

func TestSSEResponse_SendSplitsMultilineMarkupIntoDataLines(t *testing.T) {
	// "data:" is line-delimited: a client concatenates the data lines of a frame
	// and stops the field at the first newline. Emitting indented markup as one
	// data line would deliver a truncated fragment, and indentation is a
	// supported writer option (CheckedWriter.WithIndent), not an exotic case.
	sse, rec := newSSETest(t)
	comp := Raw("<ul>\n  <li>a</li>\n  <li>b</li>\n</ul>")
	if err := sse.Send(t.Context(), "message", comp); err != nil {
		t.Fatalf("Send: %v", err)
	}
	want := "event: message\n" +
		"data: <ul>\n" +
		"data:   <li>a</li>\n" +
		"data:   <li>b</li>\n" +
		"data: </ul>\n" +
		"\n"
	if got := body(rec); got != want {
		t.Errorf("frame =\n%q\nwant\n%q", got, want)
	}
}

func TestSSEResponse_SendSplitsCRAndCRLF(t *testing.T) {
	// Rendered markup only contains LF, but Raw passes arbitrary bytes through.
	// An unsplit CR is a line break to the client and would truncate the
	// fragment just as silently as an LF.
	sse, rec := newSSETest(t)
	if err := sse.Send(t.Context(), "", Raw("a\r\nb\rc\nd")); err != nil {
		t.Fatalf("Send: %v", err)
	}
	want := "data: a\ndata: b\ndata: c\ndata: d\n\n"
	if got := body(rec); got != want {
		t.Errorf("frame = %q, want %q", got, want)
	}
}

func TestSSEResponse_SendRenderErrorWritesNothing(t *testing.T) {
	// This is the per-event replacement for go-mx's per-response buffering: a
	// deferred computation failing mid-render must not reach the client as a
	// half-rendered fragment, because there is no status code left to correct it.
	sse, rec := newSSETest(t)
	wantErr := errors.New("deferred computation failed")
	// The first child renders fine, so this fails part-way through — exactly the
	// case buffering exists for.
	comp := Components{
		Text("partial output"),
		ComponentFunc(func(context.Context, Writer) error { return wantErr }),
	}
	err := sse.Send(t.Context(), "message", comp)
	if !errors.Is(err, wantErr) {
		t.Fatalf("Send error = %v, want %v", err, wantErr)
	}
	if got := body(rec); got != "" {
		t.Errorf("Send wrote %q despite the render error; a partial frame is undetectable by the client", got)
	}
}

func TestSSEResponse_SendRejectsEventNameWithLineBreak(t *testing.T) {
	// A line break in the event name would end the field and let the rest of the
	// name forge further SSE fields — frame injection, the SSE analogue of
	// header injection.
	sse, rec := newSSETest(t)
	for _, name := range []string{"a\nb", "a\rb", "a\r\nb"} {
		if err := sse.Send(t.Context(), name, Text("x")); err == nil {
			t.Errorf("Send(%q) succeeded, want an error", name)
		}
	}
	if got := body(rec); got != "" {
		t.Errorf("Send wrote %q for a rejected event name", got)
	}
}

func TestSSEResponse_SendEmptyOutputStillDispatches(t *testing.T) {
	// A frame carrying no data field at all is discarded by the client without
	// dispatching, so a component that legitimately renders nothing would
	// silently drop the event instead of delivering an empty one.
	sse, rec := newSSETest(t)
	if err := sse.Send(t.Context(), "tick", nil); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if got, want := body(rec), "event: tick\ndata:\n\n"; got != want {
		t.Errorf("frame = %q, want %q", got, want)
	}
}

func TestSSEResponse_SendEmptyEventNameOmitsEventField(t *testing.T) {
	// Omitting the field (rather than sending "event: ") is what makes the client
	// dispatch the default "message" event.
	sse, rec := newSSETest(t)
	if err := sse.Send(t.Context(), "", Text("hi")); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if got, want := body(rec), "data: hi\n\n"; got != want {
		t.Errorf("frame = %q, want %q", got, want)
	}
}

func TestSSEResponse_SendReportsCanceledContextWithoutWriting(t *testing.T) {
	// A canceled request context is how a streaming loop learns the client
	// disconnected; without this check the loop would keep rendering into a dead
	// connection until a write finally failed.
	sse, rec := newSSETest(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := sse.Send(ctx, "message", Text("x")); !errors.Is(err, context.Canceled) {
		t.Fatalf("Send error = %v, want context.Canceled", err)
	}
	if got := body(rec); got != "" {
		t.Errorf("Send wrote %q for a canceled context", got)
	}
}

func TestSSEResponse_SendErrorHidesTheErrorByDefault(t *testing.T) {
	// Same contract as RespondNonContextError: the detail stays server-side.
	// SendError exists because after the first flush there is no 500 left, not
	// to become a new way of leaking internals to the client.
	sse, rec := newSSETest(t)
	if err := sse.SendError(t.Context(), errors.New("connection to 10.0.0.4 refused")); err != nil {
		t.Fatalf("SendError: %v", err)
	}
	got := body(rec)
	if strings.Contains(got, "10.0.0.4") {
		t.Errorf("SendError leaked the error detail: %q", got)
	}
	if want := "event: " + SSEEventError + "\ndata: Internal Server Error\n\n"; got != want {
		t.Errorf("frame = %q, want %q", got, want)
	}
}

func TestSSEResponse_SendErrorRevealsWhenConfigured(t *testing.T) {
	defer func(prev bool) { RevealInternalServerErrors = prev }(RevealInternalServerErrors)
	RevealInternalServerErrors = true

	sse, rec := newSSETest(t)
	if err := sse.SendError(t.Context(), errors.New("boom")); err != nil {
		t.Fatalf("SendError: %v", err)
	}
	if want := "event: " + SSEEventError + "\ndata: boom\n\n"; body(rec) != want {
		t.Errorf("frame = %q, want %q", body(rec), want)
	}
}

func TestSSEResponse_SendErrorEscapesTheMessage(t *testing.T) {
	// The message is documented as swappable into the DOM, so an error string
	// carrying markup (a parse error quoting its input, say) must not become
	// markup. Text routes it through the writer's escaper.
	defer func(prev bool) { RevealInternalServerErrors = prev }(RevealInternalServerErrors)
	RevealInternalServerErrors = true

	sse, rec := newSSETest(t)
	if err := sse.SendError(t.Context(), errors.New(`bad token <script>`)); err != nil {
		t.Fatalf("SendError: %v", err)
	}
	if got := body(rec); strings.Contains(got, "<script>") {
		t.Errorf("SendError did not escape the message: %q", got)
	}
}

func TestSSEResponse_SendErrorIsSilentForContextErrors(t *testing.T) {
	// The client that would read the report is the one that disconnected, so
	// there is nobody to tell — mirroring RespondNonContextError.
	sse, rec := newSSETest(t)
	for _, err := range []error{context.Canceled, context.DeadlineExceeded} {
		if got := sse.SendError(t.Context(), err); got != nil {
			t.Errorf("SendError(%v) = %v, want nil", err, got)
		}
	}
	if got := body(rec); got != "" {
		t.Errorf("SendError wrote %q for a context error", got)
	}
}

func TestSSEResponse_SendErrorDeliversDespiteCanceledContext(t *testing.T) {
	// A cancellation is often what caused the error being reported. Passing the
	// request context straight through would make SendError a no-op exactly when
	// it is needed, so the cancellation is stripped for the render.
	sse, rec := newSSETest(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := sse.SendError(ctx, errors.New("query failed")); err != nil {
		t.Fatalf("SendError: %v", err)
	}
	if !strings.Contains(body(rec), "event: "+SSEEventError) {
		t.Errorf("SendError wrote nothing for a canceled context: %q", body(rec))
	}
}

func TestSSEResponse_KeepaliveWritesAnIgnoredComment(t *testing.T) {
	// It must put bytes on the wire (so an idle proxy does not drop the
	// connection) without changing anything the client renders.
	sse, rec := newSSETest(t)
	if err := sse.Keepalive(); err != nil {
		t.Fatalf("Keepalive: %v", err)
	}
	got := body(rec)
	if !strings.HasPrefix(got, ":") {
		t.Errorf("keepalive frame %q is not an SSE comment, so a client would try to interpret it", got)
	}
	if strings.Contains(got, "data:") || strings.Contains(got, "event:") {
		t.Errorf("keepalive frame %q carries fields, so it would dispatch an event", got)
	}
}

func TestSSEResponse_CloseIsIdempotentAndBlocksLaterWrites(t *testing.T) {
	// Close is meant to be deferred and may also be called explicitly, and a
	// write after the handler returned is a panic in net/http, so it has to fail
	// as an error instead.
	sse, rec := newSSETest(t)
	if err := sse.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if err := sse.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
	if err := sse.Send(t.Context(), "message", Text("x")); err == nil {
		t.Error("Send succeeded after Close")
	}
	if err := sse.Keepalive(); err == nil {
		t.Error("Keepalive succeeded after Close")
	}
	if got := body(rec); got != "" {
		t.Errorf("wrote %q after Close", got)
	}
}

func TestSSEResponse_ConcurrentWritesProduceIntactFrames(t *testing.T) {
	// Keepalive existing as a separate method implies a ticker goroutine running
	// alongside the producer. Unsynchronized writes would interleave into frames
	// no client can parse, and the corruption would be intermittent.
	sse, rec := newSSETest(t)
	const n = 50
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for range n {
			if err := sse.Send(t.Context(), "message", Text("0123456789")); err != nil {
				t.Errorf("Send: %v", err)
				return
			}
		}
	}()
	go func() {
		defer wg.Done()
		for range n {
			if err := sse.Keepalive(); err != nil {
				t.Errorf("Keepalive: %v", err)
				return
			}
		}
	}()
	wg.Wait()

	frames := strings.SplitAfter(body(rec), "\n\n")
	var events, pings int
	for _, frame := range frames {
		switch {
		case frame == "":
		case frame == string(sseKeepaliveFrame):
			pings++
		case frame == "event: message\ndata: 0123456789\n\n":
			events++
		default:
			t.Fatalf("corrupt frame %q — writes interleaved", frame)
		}
	}
	if events != n || pings != n {
		t.Errorf("got %d events and %d pings, want %d of each", events, pings, n)
	}
}

func TestSSELines(t *testing.T) {
	for _, tc := range []struct {
		data string
		want []string
	}{
		// Always at least one line, so an empty render still dispatches.
		{"", []string{""}},
		{"a", []string{"a"}},
		{"a\nb", []string{"a", "b"}},
		{"a\r\nb", []string{"a", "b"}},
		{"a\rb", []string{"a", "b"}},
		// A trailing terminator means a real trailing empty line, not "no line".
		{"a\n", []string{"a", ""}},
		{"\n", []string{"", ""}},
		// A lone CR at the end must not be read as the first half of a CRLF.
		{"a\r", []string{"a", ""}},
	} {
		var got []string
		for line := range sseLines([]byte(tc.data)) {
			got = append(got, string(line))
		}
		if len(got) != len(tc.want) {
			t.Errorf("sseLines(%q) = %q, want %q", tc.data, got, tc.want)
			continue
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("sseLines(%q) = %q, want %q", tc.data, got, tc.want)
				break
			}
		}
	}
}

func TestSSEResponse_SendEventWritesTheIDField(t *testing.T) {
	// Without an id the client has nothing to send back on reconnect, so the
	// handler cannot tell a fresh reader from a returning one.
	sse, rec := newSSETest(t)
	err := sse.SendEvent(t.Context(), SSEEvent{Name: "message", ID: "42", Comp: Text("hi")})
	if err != nil {
		t.Fatalf("SendEvent: %v", err)
	}
	if got, want := body(rec), "id: 42\nevent: message\ndata: hi\n\n"; got != want {
		t.Errorf("frame = %q, want %q", got, want)
	}
}

func TestSSEResponse_SendEventOmitsAnEmptyID(t *testing.T) {
	// An empty "id:" field would reset the client's remembered id to the empty
	// string, so a later reconnect would silently lose its resume point. The
	// field has to be absent, not empty.
	sse, rec := newSSETest(t)
	if err := sse.SendEvent(t.Context(), SSEEvent{Name: "message", Comp: Text("hi")}); err != nil {
		t.Fatalf("SendEvent: %v", err)
	}
	if strings.Contains(body(rec), "id:") {
		t.Errorf("frame %q carries an id field for an empty SSEEvent.ID", body(rec))
	}
}

func TestSSEResponse_SendEventRejectsUnusableIDs(t *testing.T) {
	// A line break forges extra fields. A NUL is worse than that: the SSE
	// specification tells clients to ignore an id containing one, so resumption
	// would break with nothing on the wire looking wrong.
	sse, rec := newSSETest(t)
	for _, id := range []string{"a\nb", "a\rb", "a\x00b"} {
		if err := sse.SendEvent(t.Context(), SSEEvent{ID: id, Comp: Text("x")}); err == nil {
			t.Errorf("SendEvent with id %q succeeded, want an error", id)
		}
	}
	if got := body(rec); got != "" {
		t.Errorf("SendEvent wrote %q for a rejected id", got)
	}
}

func TestSSEResponse_SendIsSendEventWithoutAnID(t *testing.T) {
	// Send is the shorthand; if the two drifted apart the shorthand would be a
	// second wire-format implementation to keep correct.
	plain, plainRec := newSSETest(t)
	full, fullRec := newSSETest(t)
	if err := plain.Send(t.Context(), "message", Text("hi")); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if err := full.SendEvent(t.Context(), SSEEvent{Name: "message", Comp: Text("hi")}); err != nil {
		t.Fatalf("SendEvent: %v", err)
	}
	if body(plainRec) != body(fullRec) {
		t.Errorf("Send wrote %q, SendEvent wrote %q", body(plainRec), body(fullRec))
	}
}

func TestLastEventID(t *testing.T) {
	// This is the whole server-side half of resumption: without reading the
	// header the handler cannot know where the client left off.
	req := httptest.NewRequest(http.MethodGet, "/stream", nil)
	if got := LastEventID(req); got != "" {
		t.Errorf("LastEventID of a fresh request = %q, want empty", got)
	}
	req.Header.Set(HeaderLastEventID, "42")
	if got := LastEventID(req); got != "42" {
		t.Errorf("LastEventID = %q, want %q", got, "42")
	}
}

func TestSSEResponse_SetRetry(t *testing.T) {
	sse, rec := newSSETest(t)
	if err := sse.SetRetry(2500 * time.Millisecond); err != nil {
		t.Fatalf("SetRetry: %v", err)
	}
	if got, want := body(rec), "retry: 2500\n\n"; got != want {
		t.Errorf("frame = %q, want %q", got, want)
	}
}

func TestSSEResponse_SetRetryRejectsSubMillisecond(t *testing.T) {
	// The SSE specification requires the field to be ASCII digits and tells
	// clients to ignore anything else, so a zero or negative delay would be a
	// silent no-op rather than an error the caller can see.
	sse, rec := newSSETest(t)
	for _, d := range []time.Duration{0, -time.Second, 999 * time.Microsecond} {
		if err := sse.SetRetry(d); err == nil {
			t.Errorf("SetRetry(%s) succeeded, want an error", d)
		}
	}
	if got := body(rec); got != "" {
		t.Errorf("SetRetry wrote %q for a rejected duration", got)
	}
}

func TestSSEResponse_KeepaliveLoopStopsWithTheContext(t *testing.T) {
	// A client going away is how a stream normally ends, so the loop must
	// return without an error — a caller logging that error would log one per
	// disconnect.
	sse, rec := newSSETest(t)
	ctx, cancel := context.WithCancel(t.Context())

	done := make(chan error, 1)
	go func() { done <- sse.KeepaliveLoop(ctx, time.Millisecond) }()

	// Let a few ticks land before stopping it.
	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("KeepaliveLoop returned %v on a canceled context, want nil", err)
		}
	case <-time.After(time.Second):
		t.Fatal("KeepaliveLoop did not return after its context was canceled")
	}
	if !strings.HasPrefix(body(rec), ": ") {
		t.Errorf("KeepaliveLoop wrote %q, want SSE comments", body(rec))
	}
}

func TestSSEResponse_KeepaliveLoopReturnsWriteErrors(t *testing.T) {
	// Once the stream is closed the loop has to stop on its own, or a goroutine
	// spins for the lifetime of the request context writing failures.
	sse, _ := newSSETest(t)
	if err := sse.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	done := make(chan error, 1)
	go func() { done <- sse.KeepaliveLoop(t.Context(), time.Millisecond) }()
	select {
	case err := <-done:
		if err == nil {
			t.Error("KeepaliveLoop returned nil after Close, want the write error")
		}
	case <-time.After(time.Second):
		t.Fatal("KeepaliveLoop did not return after the stream was closed")
	}
}

func TestSSEResponse_KeepaliveLoopRejectsNonPositiveInterval(t *testing.T) {
	// time.NewTicker panics on a non-positive interval, which would take down
	// the whole server from a goroutine the caller cannot recover in.
	sse, _ := newSSETest(t)
	for _, interval := range []time.Duration{0, -time.Second} {
		if err := sse.KeepaliveLoop(t.Context(), interval); err == nil {
			t.Errorf("KeepaliveLoop(%s) returned nil, want an error", interval)
		}
	}
}

// readFrame reads one SSE frame (up to and including the blank line that ends
// it) with a timeout, so a response that is buffered instead of flushed fails
// the test rather than hanging it.
func readFrame(t *testing.T, r *bufio.Reader) string {
	t.Helper()
	type result struct {
		frame string
		err   error
	}
	ch := make(chan result, 1)
	go func() {
		var frame strings.Builder
		for {
			line, err := r.ReadString('\n')
			if err != nil {
				ch <- result{frame.String(), err}
				return
			}
			frame.WriteString(line)
			if line == "\n" {
				ch <- result{frame.String(), nil}
				return
			}
		}
	}()
	select {
	case res := <-ch:
		if res.err != nil {
			t.Fatalf("reading an SSE frame: %v (got %q)", res.err, res.frame)
		}
		return res.frame
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for an SSE frame — the response was buffered instead of flushed")
		return ""
	}
}

func TestSSEResponse_FlushesThroughARealServer(t *testing.T) {
	// Every other test here writes into an httptest.ResponseRecorder, which
	// "flushes" by doing nothing, so none of them can tell a working flush loop
	// from one that buffers the whole response until the handler returns — the
	// exact failure NewSSEResponse's flusher check exists to prevent. This test
	// reads the first event off a real socket while the handler is still
	// blocked, which is only possible if the flush actually reached the client.
	release := make(chan struct{})
	handlerDone := make(chan error, 1)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sse, err := NewSSEResponse(w, nil)
		if err != nil {
			handlerDone <- err
			return
		}
		defer sse.Close()
		if err := sse.SendEvent(r.Context(), SSEEvent{Name: "message", ID: "1", Comp: Text("first")}); err != nil {
			handlerDone <- err
			return
		}
		select {
		case <-release:
		case <-r.Context().Done():
			handlerDone <- r.Context().Err()
			return
		}
		handlerDone <- sse.Send(r.Context(), "message", Text("second"))
	}))
	defer srv.Close()

	resp, err := srv.Client().Get(srv.URL)
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	if got := resp.Header.Get("Content-Type"); got != ContentTypeEventStream {
		t.Errorf("Content-Type = %q, want %q", got, ContentTypeEventStream)
	}
	if got := resp.Header.Get("X-Accel-Buffering"); got != "no" {
		t.Errorf("X-Accel-Buffering = %q, want %q — a reverse proxy will buffer the stream", got, "no")
	}
	if resp.ContentLength != -1 {
		// A Content-Length means net/http buffered the whole body to measure
		// it, which is exactly the failure mode this type exists to avoid.
		t.Errorf("Content-Length = %d, want unknown (-1): the response was not streamed", resp.ContentLength)
	}

	reader := bufio.NewReader(resp.Body)
	if got, want := readFrame(t, reader), "id: 1\nevent: message\ndata: first\n\n"; got != want {
		t.Errorf("first frame = %q, want %q", got, want)
	}

	// Only now let the handler continue — the assertion above already proved the
	// frame arrived before it did.
	close(release)
	if got, want := readFrame(t, reader), "event: message\ndata: second\n\n"; got != want {
		t.Errorf("second frame = %q, want %q", got, want)
	}
	if err := <-handlerDone; err != nil {
		t.Errorf("handler: %v", err)
	}
}

func TestSSEResponse_ResumesFromLastEventIDThroughARealServer(t *testing.T) {
	// The round trip only works if all three halves agree: the server writes
	// "id:", the browser echoes it in the Last-Event-ID header, and the handler
	// reads it back. A unit test can check any one of them and still miss that
	// they disagree.
	messages := []string{"one", "two", "three"}
	seen := make(chan string, 1)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		last := LastEventID(r)
		seen <- last

		sse, err := NewSSEResponse(w, nil)
		if err != nil {
			return
		}
		defer sse.Close()

		// Resume after the last id the client acknowledged. An unknown id means
		// "start from the beginning", never an error — it is client-controlled.
		from := 0
		if i, err := strconv.Atoi(last); err == nil && i >= 0 && i < len(messages) {
			from = i + 1
		}
		for i := from; i < len(messages); i++ {
			event := SSEEvent{Name: "message", ID: strconv.Itoa(i), Comp: Text(messages[i])}
			if err := sse.SendEvent(r.Context(), event); err != nil {
				return
			}
		}
	}))
	defer srv.Close()

	// First connection: no Last-Event-ID, so the client gets everything.
	resp, err := srv.Client().Get(srv.URL)
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	if got := <-seen; got != "" {
		t.Errorf("first request carried Last-Event-ID %q, want none", got)
	}
	reader := bufio.NewReader(resp.Body)
	if got, want := readFrame(t, reader), "id: 0\nevent: message\ndata: one\n\n"; got != want {
		t.Errorf("first frame = %q, want %q", got, want)
	}
	resp.Body.Close()

	// Reconnect the way a browser does, echoing the last id it received.
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.Header.Set(HeaderLastEventID, "0")
	resp2, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("reconnect: %v", err)
	}
	defer resp2.Body.Close()
	if got := <-seen; got != "0" {
		t.Errorf("handler saw Last-Event-ID %q, want %q", got, "0")
	}

	// The resumed stream must start after the acknowledged event, not replay it.
	reader2 := bufio.NewReader(resp2.Body)
	if got, want := readFrame(t, reader2), "id: 1\nevent: message\ndata: two\n\n"; got != want {
		t.Errorf("resumed stream started with %q, want %q — it replayed instead of resuming", got, want)
	}
}
