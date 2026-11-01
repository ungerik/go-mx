---
title: Build a live feed with hx and SSE
---

# Build a live feed with `hx` and SSE

You'll build a page with a text box and a feed. Posting sends the text to the
server with htmx, without a page reload, and every browser that has the page
open sees the post appear a moment later, pushed over Server-Sent Events. Along
the way you'll use the [`hx`](https://pkg.go.dev/github.com/ungerik/go-mx/hx)
request attributes, stream components with `mx.SSEResponse`, make the stream
survive a dropped connection, and keep the feed scrolled to its newest post.

## What you'll need

- **Go 1.27 or newer**.
- A terminal and a browser.
- The basics of building elements with `html` — if `html.Div(html.Class("x"),
  "text")` is new to you, do the **[html tutorial](../html/tutorial.html)**
  first.

## Step 1: Create a module

```sh
mkdir live-feed && cd live-feed
go mod init example.com/live-feed
go get github.com/ungerik/go-mx
```

## Step 2: A page that loads htmx

htmx is a script; go-mx ships it as ready-made `<script>` elements that load it
from a CDN with a Subresource Integrity hash. SSE support is a separate htmx
extension, so the page loads both.

Create `main.go`:

```go
package main

import (
    "log"
    "net/http"

    "github.com/ungerik/go-mx"
    "github.com/ungerik/go-mx/html"
    "github.com/ungerik/go-mx/hx"
)

func page() *html.Document {
    doc := html.NewDocument("Live feed",
        html.H1("Live feed"),
    )
    doc.HeadCustom = mx.Components{hx.ScriptFromCDN, hx.ScriptSSEFromCDN}
    return doc
}

func main() {
    http.HandleFunc("GET /{$}", page().HandleHTTP)
    log.Println("listening on http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

`go run .` and open <http://localhost:8080>. You see a heading; view the source
to find the two `<script>` tags in the `<head>`.

## Step 3: Post without a reload

The feed itself is an append-only list in memory. It also needs a way to wake
every stream that is waiting for the next post, which a channel that is closed
and replaced on every post does for any number of waiters at once. Add it to
`main.go` (and `"slices"`, `"strings"` and `"sync"` to the imports):

```go
// feed is an append-only list of posts that wakes every waiting stream when a
// post arrives.
type feed struct {
    mtx   sync.Mutex
    posts []string
    wake  chan struct{}
}

func newFeed() *feed {
    return &feed{wake: make(chan struct{})}
}

func (f *feed) post(text string) {
    f.mtx.Lock()
    defer f.mtx.Unlock()
    f.posts = append(f.posts, text)
    close(f.wake) // wakes every stream waiting on the old channel
    f.wake = make(chan struct{})
}

// since returns the posts from index next on, and a channel that is closed
// when the next post arrives.
func (f *feed) since(next int) ([]string, <-chan struct{}) {
    f.mtx.Lock()
    defer f.mtx.Unlock()
    next = min(next, len(f.posts))
    return slices.Clone(f.posts[next:]), f.wake
}
```

Now the form. `hx.Post` makes htmx submit it with a background request instead
of navigating. The post will come back to this page through the stream, so the
response has nothing to swap: `hx.Swap(hx.SwapNone)` says so, and `hx.OnHTMX`
clears the box once the request has finished.

```go
html.Form(
    hx.Post("/post"),
    hx.Swap(hx.SwapNone),
    hx.OnHTMX("after-request", "this.reset()"),
    html.Input(html.Type("text"), html.Name("text"), html.Required),
    html.Button(html.Type("submit"), "Post"),
),
```

Put it in `page()` after the heading, and add the handler:

```go
func postHandler(f *feed) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        text := strings.TrimSpace(r.FormValue("text"))
        if text == "" {
            http.Error(w, "empty post", http.StatusBadRequest)
            return
        }
        f.post(text)
        w.WriteHeader(http.StatusNoContent)
    }
}
```

and register it in `main`:

```go
f := newFeed()
http.HandleFunc("GET /{$}", page().HandleHTTP)
http.HandleFunc("POST /post", postHandler(f))
```

Typing something and pressing *Post* now clears the box and nothing else — the
browser's network tab shows the `POST /post` answered with `204`. The post is
stored; the next step shows it.

`hx.Swap` takes a typed `hx.SwapStyle`, so a misspelt style is a compile error
rather than an attribute htmx silently ignores.

## Step 4: Stream the feed

A streaming handler creates an `mx.SSEResponse`, then sends one component per
event for as long as the client stays connected:

```go
func streamHandler(f *feed) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        sse, err := mx.NewSSEResponse(w, nil)
        if err != nil {
            mx.RespondNonContextError(w, r, err)
            return
        }
        defer sse.Close()

        next := 0
        for {
            posts, wake := f.since(next)
            for _, text := range posts {
                if err := sse.Send(r.Context(), "post", html.P(text)); err != nil {
                    return // the client went away
                }
                next++
            }
            select {
            case <-r.Context().Done():
                return
            case <-wake:
            }
        }
    }
}
```

`NewSSEResponse` writes the `text/event-stream` header and commits the `200`.
Before that the handler can still fail with a clean `500` — which is why it
checks it can flush before writing anything. Every `Send` renders the component
into a buffer first, so a render error is reported before a byte of the event is
on the wire. The request context ends when the browser leaves, which ends the
loop.

On the page, three attributes connect to it: `hx.Ext("sse")` enables the
extension, `hx.SSEConnect` opens the connection, and `hx.SSESwap` swaps every
`post` event into the element — appended at the end, because of
`hx.Swap(hx.SwapBeforeEnd)`:

```go
html.Div(
    hx.Ext("sse"),
    hx.SSEConnect("/stream"),
    html.Div(
        hx.SSESwap("post"),
        hx.Swap(hx.SwapBeforeEnd),
    ),
),
```

Add it to `page()` after the form, register the handler in `main`, and run
again:

```go
http.HandleFunc("GET /stream", streamHandler(f))
```

Open the page in two windows and post in one. The post appears in both.

## Step 5: Survive a dropped connection

The browser reconnects on its own whenever the connection drops — a sleeping
laptop, a proxy timeout, a flaky network. Each reconnect is a new request, and
this handler answers every request from the first post, so after a reconnect
every post is appended to the feed a second time.

The fix is an id per event. The browser remembers the id of the last event it
received and sends it back as the `Last-Event-ID` header when it reconnects;
the handler resumes after it. Replace the `next := 0` line with these, and move
them to the top of the handler, before `NewSSEResponse`:

```go
next := 0
if id, err := strconv.Atoi(mx.LastEventID(r)); err == nil && id >= 0 {
    next = id + 1
}
```

and the `Send` call with `SendEvent`, which takes the id along with the name
and the component:

```go
err := sse.SendEvent(r.Context(), mx.SSEEvent{
    Name: "post",
    ID:   strconv.Itoa(next),
    Comp: html.P(text),
})
if err != nil {
    return // the client went away
}
```

Request input is read before `NewSSEResponse`, while a bad request could still
get an error status. The id comes back from the client, so it is untrusted:
anything that does not parse simply means "a new reader". Add `"strconv"` to
the imports.

A feed is quiet most of the time, and proxies and load balancers close
connections that stay idle. One line keeps it open, right after `defer
sse.Close()` (add `"time"` to the imports):

```go
go sse.KeepaliveLoop(r.Context(), 15*time.Second)
```

## Step 6: Follow the newest post

Once the feed is longer than the window, new posts arrive out of sight. Wrap the
feed in a `shadcn.ScrollArea` with `shadcn.StickToBottom`: it keeps the area
scrolled to the bottom as content arrives, stops following when the reader
scrolls up to read, and follows again once they scroll back down. `role="log"`
with `aria-live` does the same job for a screen reader, announcing each post.

```go
html.Div(
    hx.Ext("sse"),
    hx.SSEConnect("/stream"),
    shadcn.ScrollArea(
        html.Style("height: 12rem; overflow: auto; border: 1px solid #ccc; padding: 0 .5rem"),
        shadcn.StickToBottom,
        html.Div(
            html.Role("log"),
            html.Attrib("aria-live", "polite"),
            hx.SSESwap("post"),
            hx.Swap(hx.SwapBeforeEnd),
        ),
    ),
    html.Div(html.Role("alert"), hx.SSESwap(mx.SSEEventError)),
),
```

The inline height and overflow are what make the area scroll on a page without
a Tailwind build; the `shadcn` classes add the styled scrollbar when there is
one. Import `"github.com/ungerik/go-mx/shadcn"`.

The last `Div` is the error sink. Once the `200` is committed a server-side
failure can only travel as an event, and `sse.SendError` sends it under the name
`mx.SSEEventError`. Binding it costs one line and turns a log that silently
stops into one that says why. It must sit inside the `sse-connect` element: an
`sse-swap` outside it subscribes to nothing, without an error.

## Step 7: Test the stream

A stream is an HTTP response like any other, so `httptest` reads it. Create
`main_test.go`:

```go
package main

import (
    "bufio"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
)

func TestStreamSendsAPost(t *testing.T) {
    f := newFeed()
    f.post("hello")

    srv := httptest.NewServer(streamHandler(f))
    defer srv.Close()

    req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL, nil)
    if err != nil {
        t.Fatal(err)
    }
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        t.Fatal(err)
    }
    defer resp.Body.Close()

    // Read one frame: the lines up to the first blank line.
    var frame strings.Builder
    r := bufio.NewReader(resp.Body)
    for {
        line, err := r.ReadString('\n')
        if err != nil {
            t.Fatal(err)
        }
        if line == "\n" {
            break
        }
        frame.WriteString(line)
    }
    want := "id: 0\nevent: post\ndata: <p>hello</p>\n"
    if frame.String() != want {
        t.Errorf("frame = %q, want %q", frame.String(), want)
    }
}
```

`go test` passes. The stream never ends on its own, so the test reads exactly
the frame it wants and lets the deferred `Close` and the end of `t.Context()`
disconnect.

## What you built

A page that posts with htmx and receives with SSE, where:

- request attributes (`hx.Post`, `hx.Swap`, `hx.OnHTMX`) replace a form
  submission with a background request,
- `mx.SSEResponse` pushes one rendered component per event, with every event
  buffered before it is written,
- event ids and `mx.LastEventID` make a reconnect resume instead of replay, and
  `KeepaliveLoop` keeps an idle connection open,
- `shadcn.StickToBottom` follows the feed and an `mx.SSEEventError` sink shows a
  failure.

## Next steps

- **[How-to guides](how-to.html)** — steer the client from the response, append
  to an element an earlier event rendered with `mx.KeyedID`, end a stream for
  good with `hx.SSEClose`, and stream a searchable log with `logview`.
- **[hx reference](https://github.com/ungerik/go-mx/blob/main/hx/README.md)** —
  every attribute and header helper.
- **[`cmd/example-sse`](https://github.com/ungerik/go-mx/tree/main/cmd/example-sse)**
  — the same ideas with a reply that grows token by token, and flags that drop
  the connection or fail mid-stream.
