---
title: hx and streaming how-to guides
---

# `hx` and streaming how-to guides

Task-oriented recipes for the [`hx`](https://pkg.go.dev/github.com/ungerik/go-mx/hx)
package and for streaming components with `mx.SSEResponse`. Each one assumes
you can build elements with `html`; for a guided start, do the
**[tutorial](tutorial.html)**.

htmx requests:

- [Load htmx on a page](#load-htmx-on-a-page)
- [Swap a fragment on click](#swap-a-fragment-on-click)
- [Answer htmx and plain requests from one handler](#answer-htmx-and-plain-requests-from-one-handler)
- [Steer the client from the response](#steer-the-client-from-the-response)
- [Update several places with one response](#update-several-places-with-one-response)
- [Show that a request is in flight](#show-that-a-request-is-in-flight)

Streaming:

- [Stream components to a page](#stream-components-to-a-page)
- [Make a stream resumable](#make-a-stream-resumable)
- [Keep an idle stream open](#keep-an-idle-stream-open)
- [End a stream, or report a failure after the 200](#end-a-stream-or-report-a-failure-after-the-200)
- [Append to an element an earlier event rendered](#append-to-an-element-an-earlier-event-rendered)
- [Follow a growing transcript](#follow-a-growing-transcript)
- [Stream a log](#stream-a-log)
- [Search the whole log, not just the screen](#search-the-whole-log-not-just-the-screen)
- [Find out why a stream shows nothing](#find-out-why-a-stream-shows-nothing)

## Load htmx on a page

```go
doc := html.NewDocument("App", body)
doc.HeadCustom = mx.Components{
    hx.ScriptFromCDN,    // htmx itself, minified
    hx.ScriptSSEFromCDN, // the sse extension, only for streaming
}
```

Both load from jsDelivr with a Subresource Integrity hash. During development
`hx.ScriptDebugFromCDN` loads the unminified build instead of
`hx.ScriptFromCDN`.

## Swap a fragment on click

The element says what to request and where the answer goes:

```go
html.Button(
    hx.Get("/stats"),
    hx.Target("#stats"),
    hx.Swap(hx.SwapInnerHTML),
    "Refresh",
),
html.Div(html.ID("stats")),
```

The handler renders just the fragment. Build the component per request, then
let `mx.ComponentHTTPHandler` render and write it, which turns a render error
into a 500 instead of a half-written fragment:

```go
func stats(w http.ResponseWriter, r *http.Request) {
    fragment := html.P(fmt.Sprintf("%d invoices open", countOpen(r.Context())))
    mx.ComponentHTTPHandler(fragment, mx.DefaultWriterFactory, nil)(w, r)
}
```

`hx.Swap` takes a typed `hx.SwapStyle` (`hx.SwapOuterHTML`,
`hx.SwapBeforeEnd`, …); modifiers such as `"scroll:bottom"` follow as strings.
An out-of-set style renders as an error instead of a broken attribute.

## Answer htmx and plain requests from one handler

A boosted link or a bookmarked URL can reach the same handler with or without
htmx. Answer the fragment to htmx and the whole page to everyone else:

```go
func invoices(w http.ResponseWriter, r *http.Request) {
    list := invoiceList(r.Context())
    if hx.IsRequest(r) {
        mx.ComponentHTTPHandler(list, mx.DefaultWriterFactory, nil)(w, r)
        return
    }
    html.NewDocument("Invoices", list).HandleHTTP(w, r)
}
```

`hx.IsBoosted(r)` tells a boosted navigation apart, and
`hx.IsHistoryRestoreRequest(r)` a history restore after a cache miss.

## Steer the client from the response

Response headers tell htmx to do something besides swapping. Set them before
the body is written:

```go
hx.SetTrigger(w, "invoice-saved")   // fire an event on the client
hx.SetRedirect(w, "/invoices")      // full client-side redirect
hx.SetRetarget(w, "#errors")        // swap somewhere else than hx-target
if err := hx.SetReswap(w, hx.SwapOuterHTML); err != nil {
    // only an invalid style fails, and leaves the header unset
}
```

`hx.SetTriggerJSON(w, map[string]any{"invoice-saved": map[string]any{"id": id}})`
sends an event with a detail payload. The full set, including `HX-Push-Url` and
`HX-Refresh`, is in the
[hx reference](https://github.com/ungerik/go-mx/blob/main/hx/README.md#server-side-http-headers).

## Update several places with one response

Besides the main fragment, a response can carry elements marked for an
out-of-band swap. htmx takes each one out of the response and swaps it by id:

```go
mx.Components{
    invoiceRow(invoice),                  // the main swap
    html.Span(html.ID("open-count"), hx.SwapOOBTrue, strconv.Itoa(open)),
}
```

`hx.SwapOOBTrue` replaces the element with the same id. `hx.SwapOOB(style,
selector)` swaps with a given style into a selector, for example
`hx.SwapOOB(hx.SwapBeforeEnd, "#notifications")` to append. The same marking
works inside a streamed event — see
[below](#append-to-an-element-an-earlier-event-rendered).

## Show that a request is in flight

htmx adds `htmx-request` to the element making a request, or to its indicator,
for as long as the request runs. The class names are constants:

```go
html.Button(hx.Post("/save"), hx.Indicator("#saving"), "Save"),
html.Span(html.ID("saving"), html.Class(hx.ClassIndicator), "Saving…"),
```

htmx's default stylesheet shows an `hx.ClassIndicator` element only while
`hx.ClassRequest` is on it or an ancestor.

## Stream components to a page

The handler sends one component per event:

```go
func stream(w http.ResponseWriter, r *http.Request) {
    // Everything that may fail with an error status happens before this line.
    sse, err := mx.NewSSEResponse(w, nil)
    if err != nil {
        mx.RespondNonContextError(w, r, err) // still a clean 500
        return
    }
    defer sse.Close()

    for msg := range messages(r.Context()) {
        if err := sse.Send(r.Context(), "message", html.P(msg)); err != nil {
            return // the client went away
        }
    }
}
```

and the page connects to it and swaps the named event:

```go
html.Div(hx.Ext("sse"), hx.SSEConnect("/stream"),
    html.Div(hx.SSESwap("message"), hx.Swap(hx.SwapBeforeEnd)),
)
```

`hx.Swap(hx.SwapBeforeEnd)` appends each event; without it every event replaces
the last. One `sse-connect` serves every `sse-swap` inside it, so put several
subscribers under one connection rather than a connection on each.

Each `Send` renders into a buffer before writing, so a component that fails to
render is reported without a partial event reaching the client. Markup that
spans several lines is split across `data:` lines and rejoined by the browser,
so an indenting writer is fine.

## Make a stream resumable

A browser reconnects on its own and silently whenever a stream drops. Give
every event an id the handler can resume after, and read the id the browser
sends back:

```go
after := mx.LastEventID(r) // read before NewSSEResponse; "" for a new reader
// ...
for _, msg := range store.Since(after) {
    err := sse.SendEvent(r.Context(), mx.SSEEvent{
        Name: "message",
        ID:   msg.ID,
        Comp: renderMessage(msg),
    })
    if err != nil {
        return
    }
}
```

The id is echoed by the client, so it is untrusted input: validate it, and treat
an unknown one as "start from the beginning" rather than as an error. Choose ids
so that "everything after this" is answerable — a sequence number or a message
id, not a hash.

`sse.SetRetry(500 * time.Millisecond)`, sent once before the first event, sets
how long the browser waits before reconnecting; the default is about three
seconds.

**Verify it:** run
[`cmd/example-sse`](https://github.com/ungerik/go-mx/tree/main/cmd/example-sse)
with `-drop`. It cuts the first connection mid-reply; the transcript still
completes exactly once.

## Keep an idle stream open

Proxies, load balancers and browsers each close a connection that stays quiet
for too long. A log or a notification stream is quiet most of the time:

```go
go sse.KeepaliveLoop(r.Context(), 15*time.Second)
```

It writes an SSE comment, which every client ignores, at the interval, and ends
with the request context. Pick an interval below the shortest idle timeout on
the path. `sse.SetWriteTimeout` bounds how long one frame may take to write
(`mx.DefaultSSEWriteTimeout` by default), so a client that stops reading cannot
hold the handler forever.

## End a stream, or report a failure after the 200

SSE has no close frame: when the handler returns, the browser reconnects. To end
a stream for good, send an event the page has bound to `hx.SSEClose`:

```go
html.Div(hx.Ext("sse"), hx.SSEConnect("/stream"), hx.SSEClose("done"),
    html.Div(hx.SSESwap("message"), hx.Swap(hx.SwapBeforeEnd)),
    html.Div(html.Role("alert"), hx.SSESwap(mx.SSEEventError)),
)
```

```go
_ = sse.Send(ctx, "done", nil) // last thing the handler sends
```

Once `NewSSEResponse` has committed the 200, a failure can only travel in band.
`sse.SendError(ctx, err)` sends it as the `mx.SSEEventError` event, which the
`role="alert"` element above swaps in. The client sees a generic "Internal
Server Error" unless `mx.RevealInternalServerErrors` is set; the error itself
goes to `mx.ReportError`. Send the close event after it, or the browser
reconnects and the failure repeats.

`mx.SSEEventError` is `"mx-error"`, not `"error"`: the browser already uses
`error` for its own transport failures. Those reach htmx as the
`hx.EventSSEError` event, which a page can handle with
`hx.OnHTMX("sse-error", …)`.

## Append to an element an earlier event rendered

A streamed reply that grows token by token is one element that later events
append to. Those events find it by id, so the id must come out the same in every
render — and in a reconnect handled by another process. `mx.UniqueID` draws from
a counter and cannot do that; `mx.KeyedID` derives the id from key parts:

```go
// The event that opens the reply renders the container.
sse.Send(ctx, "message", html.Div(
    html.Strong("assistant: "),
    html.Span(mx.KeyedID("reply", msgID)),
))

// Every later event appends a token to it out of band.
target := "#" + mx.KeyedIDValue("reply", msgID)
for token := range tokens {
    sse.Send(ctx, "message", html.Span(hx.SwapOOB(hx.SwapBeforeEnd, target), token))
}
```

`KeyedIDValue` returns the same string `KeyedID` renders, which is what makes
the selector match. The value is reduced to characters valid in an id; when that
reduction changes a part, a short digest keeps distinct keys distinct.

## Follow a growing transcript

Content that streams in arrives below the visible area. `shadcn.StickToBottom`
on a `shadcn.ScrollArea` keeps it scrolled to the bottom, stops following when
the reader scrolls up, and follows again once they return:

```go
shadcn.ScrollArea(
    html.Style("height: 20rem; overflow: auto"),
    shadcn.StickToBottom,
    html.Div(
        html.Role("log"),
        html.Attrib("aria-live", "polite"),
        hx.SSESwap("message"),
        hx.Swap(hx.SwapBeforeEnd),
    ),
)
```

The inline height and overflow make the area scroll even without a Tailwind
build. `shadcn.StickToBottomThreshold(px)` replaces the default distance from
the bottom that still counts as following (at most
`shadcn.StickToBottomDefaultThresholdPx`). `role="log"` with `aria-live` is what
announces the new content to a screen reader.

## Stream a log

[`logview`](https://github.com/ungerik/go-mx/blob/main/logview/README.md) is a
complete log viewer: JSON records render as their fields (time and level
promoted, `key=value` colored by type), plain-text lines get their level word
colored, and the view adds a filter, a pause button and scroll anchoring. The
page:

```go
html.Head(
    hx.ScriptFromCDN,
    hx.ScriptSSEFromCDN,
    logview.StyleElement(),
),
html.Body(
    logview.View(hx.Ext("sse"), hx.SSEConnect("/logs/stream"), hx.SSEClose("done")),
)
```

The handler sends each raw line through `logview.Line`, under the event name the
view listens for:

```go
for line := range tail(r.Context(), mx.LastEventID(r)) {
    err := sse.SendEvent(r.Context(), mx.SSEEvent{
        Name: logview.Default.EventName(),
        ID:   line.Seq,
        Comp: logview.Line(line.Text),
    })
    if err != nil {
        return
    }
}
```

Send a backlog as one event with `logview.Lines(texts...)`: htmx inserts its
lines individually, for one flush instead of one per line. Configure a
`*logview.Config` instead of using `logview.Default` to set the message field
(`MessageKey: "msg"` for `log/slog`), the line cap (`MaxLines`), or the theme
(`logview.LightTheme`). Terminal color codes are stripped from every line;
`KeepANSI` keeps them as text.

## Search the whole log, not just the screen

The view's filter hides the lines on screen that do not match; it cannot see the
lines `MaxLines` dropped or the backlog behind the stream. With `SearchParam`
set, a change of filter reconnects the stream with the filter as a query
parameter, and the handler answers with what matches:

```go
var logs = &logview.Config{SearchParam: "q"}

func stream(w http.ResponseWriter, r *http.Request) {
    query := r.URL.Query().Get(logs.SearchParam)
    // ...
    for line := range tail(r.Context(), mx.LastEventID(r)) {
        if !logs.Matches(line.Text, query) {
            continue
        }
        // send it as above, with logs.Line
    }
}
```

`Matches` compares the line as the view renders it, case-insensitively, so the
server sends exactly what the filter would show: a reader typing `status=500`
finds the record that carries `"status":500`. A search is answered like a
resume — the matching history, then matching live lines — and a reconnect of a
search sends `Last-Event-ID` too, so ids must stay those of the log itself.

**Verify it:** run
[`cmd/example-logstream`](https://github.com/ungerik/go-mx/tree/main/cmd/example-logstream),
wait for a `panic` line to scroll past, and type `panic` into the filter.

## Find out why a stream shows nothing

- **The `sse-swap` element is outside the `sse-connect` element.** It then
  subscribes to nothing, and the extension reports that neither as an event nor
  on the console. Every subscriber must be a descendant of the connecting
  element (or the element itself).
- **The event names differ.** The handler's event name must equal the
  `hx.SSESwap` name exactly. An unnamed event (`Send(ctx, "", comp)`) is
  `message` on the client.
- **The extension is missing.** `sse-*` attributes need `hx.ScriptSSEFromCDN`
  in addition to htmx, and `hx.Ext("sse")` on the connecting element or an
  ancestor.
- **Something buffers the response.** A proxy that buffers responses holds the
  events back until it has enough of them. `NewSSEResponse` sets
  `Cache-Control: no-cache` and `X-Accel-Buffering: no` (which nginx honors) and
  fails early if the writer cannot flush at all, but other buffering proxies or
  compressing middleware in front of the handler need their own configuration.
- **The browser tab is in the background.** Browsers throttle background tabs,
  and a reconnect may not happen until the tab is visible again.
