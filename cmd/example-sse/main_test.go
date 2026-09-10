package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ungerik/go-mx"
)

// streamBody runs the stream handler to completion and returns the raw SSE
// wire bytes, which is the level the interesting mistakes are visible at.
func streamBody(t *testing.T, fail bool) string {
	t.Helper()
	prev := tickInterval
	tickInterval = 0
	t.Cleanup(func() { tickInterval = prev })

	rec := httptest.NewRecorder()
	streamHandler(fail)(rec, httptest.NewRequest(http.MethodGet, "/stream", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	return rec.Body.String()
}

func TestPageWiresTheClientToTheStream(t *testing.T) {
	// Each of these is a silent failure if missing: htmx just never connects,
	// never subscribes, or never follows, with no error anywhere.
	var b strings.Builder
	if err := page().Render(t.Context(), mx.NewCheckedWriter(&b)); err != nil {
		t.Fatalf("render: %v", err)
	}
	out := b.String()
	for _, want := range []string{
		`hx-ext="sse"`,
		`sse-connect="/stream"`,
		`sse-swap="` + eventMessage + `"`,
		`sse-swap="` + mx.SSEEventError + `"`,
		`sse-close="` + eventDone + `"`,
		`data-stick-to-bottom=""`,
		"window.mxStickToBottom",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in the page", want)
		}
	}
}

func TestStreamSplitsMultilineMarkupIntoDataLines(t *testing.T) {
	// The transcript lines contain newlines on purpose. A client concatenates
	// the data fields of a frame and stops each at the newline, so an unsplit
	// fragment arrives truncated.
	body := streamBody(t, false)
	if !strings.Contains(body, "<strong>user: </strong>\ndata: <span>How much") {
		t.Errorf("multi-line message was not split across data: lines:\n%s", firstFrames(body, 2))
	}
	for _, line := range strings.Split(body, "\n") {
		if line == "" || strings.HasPrefix(line, "data:") || strings.HasPrefix(line, "event:") {
			continue
		}
		t.Fatalf("stray line %q — every non-blank line must be an SSE field", line)
	}
}

func TestStreamAppendsTokensToTheKeyedContainer(t *testing.T) {
	// This is the whole reason KeyedID exists next to UniqueID: the id rendered
	// in one event has to be the selector a later event targets. If the two
	// drifted apart the tokens would silently land nowhere.
	body := streamBody(t, false)
	id := mx.KeyedIDValue(replyKey)
	if !strings.Contains(body, `id="`+id+`"`) {
		t.Fatalf("the reply container was not rendered with id %q", id)
	}
	if !strings.Contains(body, `hx-swap-oob="beforeend:#`+id+`"`) {
		t.Fatalf("no token targets #%s out of band", id)
	}
	if got, want := strings.Count(body, `hx-swap-oob="beforeend:#`+id+`"`), 9; got != want {
		t.Errorf("%d tokens target the container, want %d", got, want)
	}
}

func TestStreamEndsWithTheCloseEvent(t *testing.T) {
	// hx.SSEClose binds this name; without the event the browser reconnects and
	// replays the whole transcript forever.
	body := streamBody(t, false)
	if !strings.Contains(body, "event: "+eventDone+"\n") {
		t.Error("the stream never sent the close event")
	}
	if strings.Contains(body, "event: "+mx.SSEEventError) {
		t.Error("a successful stream sent an error event")
	}
}

func TestStreamFailureTravelsInBand(t *testing.T) {
	// By the time the failure happens the 200 and four messages are already on
	// the wire, so there is no status code left to report it with.
	body := streamBody(t, true)
	if !strings.Contains(body, "event: "+mx.SSEEventError+"\n") {
		t.Fatalf("the failure was not reported as an %s event:\n%s", mx.SSEEventError, body)
	}
	// The failure path must close too: a browser reconnects to a stream that
	// merely ends, so without this it would replay and re-report forever.
	if !strings.Contains(body, "event: "+eventDone+"\n") {
		t.Error("a failed stream did not close, so the client will reconnect and replay it")
	}
	if strings.LastIndex(body, "event: "+mx.SSEEventError) > strings.LastIndex(body, "event: "+eventDone) {
		t.Error("the error event came after the close event")
	}
	// mx.RevealInternalServerErrors is false by default, so the detail must not
	// reach the client.
	if strings.Contains(body, "ledger query") {
		t.Errorf("the error detail leaked to the client:\n%s", body)
	}
}

// firstFrames returns the first n SSE frames of body, for readable failures.
func firstFrames(body string, n int) string {
	frames := strings.SplitAfter(body, "\n\n")
	if len(frames) > n {
		frames = frames[:n]
	}
	return strings.Join(frames, "")
}
