package hx

import (
	"strings"

	"github.com/ungerik/go-mx"
)

// SSEConnect opens an SSE connection to url. The events of that connection are
// available to this element and its descendants, so it is the element the
// [SSESwap] targets sit under:
//
//	html.Div(hx.Ext("sse"), hx.SSEConnect("/chat/stream"),
//		html.Div(hx.SSESwap("message")),
//	)
//
// htmx 2.0 moved SSE out of core, so unlike the hx-* attributes in
// attributes.go, SSEConnect, [SSESwap] and [SSEClose] are spelled without the
// hx- prefix — that is how the extension defines them, not an omission. Load
// the extension on the connecting element (or an ancestor) with Ext("sse"), and
// the extension script itself with [ScriptSSEFromCDN].
//
// The extension raises [EventSSEError] on a transport failure, while failures
// the server reports in-band arrive as an ordinary named event, which
// [mx.SSEResponse.SendError] sends as [mx.SSEEventError].
//
// Nesting is not checked at all: an [SSESwap] element with no sse-connect above
// it subscribes to nothing and never updates, with no event and no console
// message. That silence is worth knowing about, because it looks exactly like a
// server that is not sending.
//
// An empty url defers an error to render time: the extension guards on the
// attribute's truthiness, so sse-connect="" opens no connection at all and every
// [SSESwap] below it goes silent with nothing reported.
//
// See https://htmx.org/extensions/sse/
func SSEConnect(url string) mx.Attrib {
	if url == "" {
		return mx.ErrAttribf("sse-connect", "hx: SSEConnect needs a URL")
	}
	return mx.NewAttrib("sse-connect", url)
}

// SSESwap swaps the content of the named events into this element, using the
// element's hx-swap and hx-target like a normal htmx response. Several event
// names are joined with a comma, mirroring how [SwapOOB] takes variadic
// selectors.
//
// Calling it with no event names, or with an empty one, defers an error to
// render time. Both produce a subscription to nothing — the extension skips a
// falsy sse-swap outright, and an empty name in the comma list matches no event
// — which looks exactly like a server that never sends.
func SSESwap(events ...string) mx.Attrib {
	if len(events) == 0 {
		return mx.ErrAttribf("sse-swap", "hx: SSESwap needs at least one event name")
	}
	for _, event := range events {
		if strings.TrimSpace(event) == "" {
			return mx.ErrAttribf("sse-swap", "hx: SSESwap event names must not be empty, got %q", events)
		}
	}
	return mx.NewAttrib("sse-swap", strings.Join(events, ","))
}

// SSEClose closes the SSE connection when the named event arrives, so a server
// that is done streaming can stop the client from reconnecting. Put it on the
// same element as [SSEConnect] and send event as the last event of the stream.
//
// Closing raises [EventSSEClose], which the extension also raises when the
// connecting element leaves the DOM — its detail carries a type distinguishing
// the two, so a handler that only cares about a finished stream has to check it.
//
// An empty event defers an error to render time: the extension guards on the
// attribute's truthiness, so sse-close="" never closes anything and the client
// reconnects to a finished stream forever.
func SSEClose(event string) mx.Attrib {
	if strings.TrimSpace(event) == "" {
		return mx.ErrAttribf("sse-close", "hx: SSEClose needs an event name")
	}
	return mx.NewAttrib("sse-close", event)
}
