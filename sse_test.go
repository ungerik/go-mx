package mx

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
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
		case frame == sseKeepaliveFrame:
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
