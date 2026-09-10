package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/ungerik/go-mx"
)

// streamBody runs the stream handler to completion and returns the raw SSE
// wire bytes, which is the level the interesting mistakes are visible at.
func streamBody(t *testing.T, fail bool) string {
	t.Helper()
	return streamFrom(t, fail, false, "")
}

// streamFrom runs the stream handler as a client with the given Last-Event-ID
// would see it, so a reconnect can be replayed in a test.
func streamFrom(t *testing.T, fail, drop bool, lastEventID string) string {
	t.Helper()
	prev := tickInterval
	tickInterval = 0
	t.Cleanup(func() { tickInterval = prev })

	req := httptest.NewRequest(http.MethodGet, "/stream", nil)
	if lastEventID != "" {
		req.Header.Set(mx.HeaderLastEventID, lastEventID)
	}
	rec := httptest.NewRecorder()
	streamHandler(fail, drop)(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	return rec.Body.String()
}

// eventIDs returns the "id:" field of every frame in body, in order.
func eventIDs(body string) []string {
	var ids []string
	for _, line := range strings.Split(body, "\n") {
		if after, ok := strings.CutPrefix(line, "id: "); ok {
			ids = append(ids, after)
		}
	}
	return ids
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
	// A line that is not a known field (or a ":" comment) is either a broken
	// split or an injected frame; both reach the client as garbage.
	for _, line := range strings.Split(body, "\n") {
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}
		switch {
		case strings.HasPrefix(line, "data:"),
			strings.HasPrefix(line, "event:"),
			strings.HasPrefix(line, "id:"),
			strings.HasPrefix(line, "retry:"):
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
	if got, want := strings.Count(body, `hx-swap-oob="beforeend:#`+id+`"`), len(replyTokens); got != want {
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
	// By the time the failure happens the 200 and the whole transcript are
	// already on the wire, so there is no status code left to report it with.
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

func TestStreamTagsEveryEventWithItsIndex(t *testing.T) {
	// Without ids a client has nothing to send back, so a dropped connection
	// can only be answered by replaying the whole transcript.
	ids := eventIDs(streamBody(t, false))
	want := len(streamSteps())
	if len(ids) != want {
		t.Fatalf("%d events carry an id, want %d", len(ids), want)
	}
	for i, id := range ids {
		if id != strconv.Itoa(i) {
			t.Errorf("event %d has id %q, want %q", i, id, strconv.Itoa(i))
		}
	}
}

func TestStreamResumesAfterTheAcknowledgedEvent(t *testing.T) {
	// The point of resumption: the client already has everything up to and
	// including the id it sent back, so re-sending any of it would duplicate
	// messages in the transcript.
	ids := eventIDs(streamFrom(t, false, false, "5"))
	if len(ids) == 0 {
		t.Fatal("the resumed stream sent no events")
	}
	if ids[0] != "6" {
		t.Errorf("the resumed stream started at id %q, want %q", ids[0], "6")
	}
}

func TestStreamRestartsFromAnUnusableLastEventID(t *testing.T) {
	// Last-Event-ID is echoed back by the client, so it is client-controlled
	// input. An unrecognized value has to mean "start from the beginning" —
	// never an error, and never an index into anything.
	for _, id := range []string{"not-a-number", "-1", "999999", "0x10"} {
		ids := eventIDs(streamFrom(t, false, false, id))
		if len(ids) == 0 || ids[0] != "0" {
			t.Errorf("Last-Event-ID %q resumed at %v, want a restart from 0", id, ids)
		}
	}
}

func TestDroppedConnectionResumesWithoutDuplicates(t *testing.T) {
	// The whole reason resumption exists, end to end: a browser reconnects to a
	// stream that merely stops, and what it gets back has to continue the
	// transcript rather than repeat it. A stream without event ids fails this
	// by delivering every message twice.
	first := streamFrom(t, false, true, "")
	firstIDs := eventIDs(first)
	if len(firstIDs) == 0 {
		t.Fatal("the dropped connection sent nothing")
	}
	if strings.Contains(first, "event: "+eventDone) {
		t.Fatal("the dropped connection sent the close event, so the browser would not reconnect")
	}

	// Reconnect the way a browser does, echoing the last id it received.
	second := streamFrom(t, false, true, firstIDs[len(firstIDs)-1])
	all := append(firstIDs, eventIDs(second)...)

	seen := make(map[string]bool, len(all))
	for _, id := range all {
		if seen[id] {
			t.Errorf("event %q was delivered twice across the reconnect", id)
		}
		seen[id] = true
	}
	if len(seen) != len(streamSteps()) {
		t.Errorf("%d distinct events delivered across the reconnect, want %d", len(seen), len(streamSteps()))
	}
	if !strings.Contains(second, "event: "+eventDone) {
		t.Error("the resumed stream did not close, so the browser would reconnect again")
	}
}

func TestStreamSetsTheReconnectDelay(t *testing.T) {
	// A browser waits about three seconds by default, which makes a dropped
	// connection look like a hang rather than a reconnect.
	if !strings.HasPrefix(streamBody(t, false), "retry: 500\n\n") {
		t.Error("the stream did not set the reconnect delay before its first event")
	}
}

// nonFlushingWriter is an http.ResponseWriter that cannot flush, which is what
// a handler sees behind middleware that wraps the writer without forwarding
// Flush.
type nonFlushingWriter struct{ http.ResponseWriter }

func TestStreamAnswersANonFlushableWriterWithACleanError(t *testing.T) {
	// This is the shape the whole handler is arranged around: everything that
	// can fail with a status code has to happen before NewSSEResponse commits
	// the 200. A stream that cannot flush would otherwise be delivered as one
	// buffered response when the handler returns, which looks like a hang the
	// client can only time out on.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/stream", nil)
	streamHandler(false, false)(nonFlushingWriter{rec}, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	if strings.Contains(rec.Body.String(), "data:") {
		t.Errorf("the handler streamed into a writer that cannot flush:\n%s", rec.Body.String())
	}
}

func TestStreamStopsWhenTheClientDisconnects(t *testing.T) {
	// A canceled request context is how a handler learns the browser navigated
	// away, reloaded, or closed the stream. Without checking it the loop renders
	// every remaining step into a socket nobody reads — and with a real tick
	// interval it holds the goroutine for the rest of the transcript.
	prev := tickInterval
	tickInterval = 0
	t.Cleanup(func() { tickInterval = prev })

	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	req := httptest.NewRequest(http.MethodGet, "/stream", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	streamHandler(false, false)(rec, req)

	if got := len(eventIDs(rec.Body.String())); got != 0 {
		t.Errorf("%d events were written for a disconnected client:\n%s", got, rec.Body.String())
	}
	// The close event is for a client that is still there; a disconnected one
	// cannot read it, and reaching it would mean the loop ran to the end.
	if strings.Contains(rec.Body.String(), "event: "+eventDone) {
		t.Errorf("the handler ran to completion for a disconnected client:\n%s", rec.Body.String())
	}
}
