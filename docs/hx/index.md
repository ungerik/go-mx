---
title: hx and streaming for go-mx
---

# The `hx` package and streaming

[htmx](https://htmx.org) lets a server-rendered page update itself: an attribute
on an element says which URL to request and where to put the HTML that comes
back. The `hx` package spells those attributes as Go functions, typed where htmx
constrains the value, and adds the handler side: request readers and
response-header setters.

```go
import "github.com/ungerik/go-mx/hx"

html.Button(
    hx.Post("/clicked"),
    hx.Target("#result"),
    hx.Swap(hx.SwapOuterHTML),
    "Click me",
)
// <button hx-post="/clicked" hx-target="#result" hx-swap="outerHTML">Click me</button>
```

**Streaming** is the same idea in the other direction: the server pushes HTML
after the page has loaded. go-mx renders one component per
[Server-Sent Event](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events),
and htmx's `sse` extension swaps each one into the page. The pieces:

| Piece | Package | Job |
|-------|---------|-----|
| `mx.SSEResponse` | root `mx` | renders one component per event over a still-open response, with resumable ids, keepalives and in-band errors |
| `hx.SSEConnect`, `hx.SSESwap`, `hx.SSEClose` | `hx` | the extension's attributes: open the connection, swap named events, close it |
| `mx.KeyedID`, `mx.KeyedIDValue` | root `mx` | an element id a later event can reproduce, so it can append to what an earlier event rendered |
| `shadcn.StickToBottom` | `shadcn` | makes a `ScrollArea` follow the stream until the reader scrolls up |
| `logview` | `logview` | a complete log viewer on top of all of the above: JSON records rendered as fields, filter or server-side search, pause |

This documentation follows the [Diátaxis](https://diataxis.fr) framework.

## Tutorial

New to htmx or streaming? Start here.

- **[Build a live feed](tutorial.html)** — a page where a form posts with htmx
  and every open browser sees the post arrive over SSE: request attributes, a
  streaming handler, resuming after a dropped connection, scroll anchoring,
  and a test that reads the stream.

## How-to guides

Task-oriented recipes for things you'll actually need.

- **[hx and streaming how-to guides](how-to.html)** — load htmx, swap a
  fragment, answer htmx and plain requests from one handler, steer the client
  from the response, update several places at once, stream components, make a
  stream resumable, keep it alive, report a failure after the 200, append to an
  element an earlier event rendered, follow a growing transcript, stream and
  search a log, and find out why a stream shows nothing.

## Reference

The complete reference lives next to the code, so it can't drift:

- **[hx package reference](https://github.com/ungerik/go-mx/blob/main/hx/README.md)**
  — every request attribute, the typed values, `hx-on`, out-of-band swaps, the
  SSE attributes, event and class constants, and the header protocol.
- **[logview reference](https://github.com/ungerik/go-mx/blob/main/logview/README.md)**
  — how a line is rendered, configuration, themes, the view's behavior and
  server-side search.
- **[`mx.SSEResponse` on pkg.go.dev](https://pkg.go.dev/github.com/ungerik/go-mx#SSEResponse)**
  — buffering, reconnection, keepalives and errors, with the rest of the root
  package.
- **Full API on pkg.go.dev:**
  [hx](https://pkg.go.dev/github.com/ungerik/go-mx/hx) ·
  [logview](https://pkg.go.dev/github.com/ungerik/go-mx/logview) ·
  [shadcn.ScrollArea](https://pkg.go.dev/github.com/ungerik/go-mx/shadcn#ScrollArea)

Two worked commands in the repository run everything end to end:
[`cmd/example-sse`](https://github.com/ungerik/go-mx/tree/main/cmd/example-sse)
streams a chat transcript that grows token by token (with `-drop` to cut the
connection and watch it resume, and `-fail` for an in-band error), and
[`cmd/example-logstream`](https://github.com/ungerik/go-mx/tree/main/cmd/example-logstream)
streams a simulated production log into a searchable `logview`.

## Explanation

Why streaming works the way it does:

- **Buffered events, committed status.** go-mx buffers every response so that a
  render error becomes a clean 500 instead of a truncated page. An
  `SSEResponse` keeps that guarantee per event — a failed render is reported
  before a byte of its frame is written — but `NewSSEResponse` commits the 200,
  so anything that should fail with a 500 has to happen before it, and later
  failures travel in band as the `mx.SSEEventError` event.
- **Why the error event is `mx-error`, not `error`.** A browser already
  dispatches its own transport failures at the `EventSource` under the name
  `error`, with no data. A subscription to `error` would fire on every dropped
  connection and give htmx nothing to swap. Bind `mx.SSEEventError` for
  failures the server reports, and `hx.EventSSEError` for the transport.
- **Why ids matter.** A browser reconnects on its own and silently, after a
  dropped connection, a sleeping laptop or a proxy timeout. Without event ids a
  stream that appends replays everything on every reconnect.
- **Why server-rendered at all** — see [Why go-mx?](../why-go-mx.html).
