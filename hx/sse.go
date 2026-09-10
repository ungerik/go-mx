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
// The extension raises [EventSSEError] on a transport failure and
// [EventNoSSESourceError] when an sse-swap element has no source above it;
// failures the server reports in-band arrive as an ordinary named event, which
// [mx.SSEResponse.SendError] sends as [mx.SSEEventError].
//
// See https://htmx.org/extensions/sse/
func SSEConnect(url string) mx.Attrib { return mx.NewAttrib("sse-connect", url) }

// SSESwap swaps the content of the named events into this element, using the
// element's hx-swap and hx-target like a normal htmx response. Several event
// names are joined with a comma, mirroring how [SwapOOB] takes variadic
// selectors.
//
// Calling it with no event names defers an error to render time: an empty
// sse-swap subscribes to nothing, which would look like a server that never
// sends.
func SSESwap(events ...string) mx.Attrib {
	if len(events) == 0 {
		return mx.ErrAttribf("sse-swap", "hx: SSESwap needs at least one event name")
	}
	return mx.NewAttrib("sse-swap", strings.Join(events, ","))
}

// SSEClose closes the SSE connection when the named event arrives, so a server
// that is done streaming can stop the client from reconnecting. Put it on the
// same element as [SSEConnect] and send event as the last event of the stream.
func SSEClose(event string) mx.Attrib { return mx.NewAttrib("sse-close", event) }
