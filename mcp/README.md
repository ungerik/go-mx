# mcp

The public pages of a [`web.Site`](../web) as read-only
[Model Context Protocol](https://modelcontextprotocol.io) tools, so an AI agent
can search and read a site the way a crawler reads its sitemap. Built on the
official Go SDK,
[`github.com/modelcontextprotocol/go-sdk`](https://github.com/modelcontextprotocol/go-sdk).

It is a nested module, `github.com/ungerik/go-mx/mcp`, so the SDK and its
dependencies are only pulled into programs that serve MCP. The SDK package is
called `mcp` too, so import this one under another name:

```go
import (
    "github.com/modelcontextprotocol/go-sdk/mcp"
    mxmcp "github.com/ungerik/go-mx/mcp"
)
```

## What it exposes

`mxmcp.AddPageTools(server, source, site, opts)` adds three tools to a server
you created:

| Tool                                      | Returns |
|-------------------------------------------|---------|
| `search_pages(query, language?, limit?)`  | Pages containing every word of the query in their title, description or Markdown source, best match first, each with a snippet. 10 results unless asked for another number, at most 25. A query is at most 16 distinct words and 256 bytes, since each word is searched in the text of every page. |
| `list_pages(language?)`                   | Every page, without snippets. |
| `get_page(path)`                          | The [Markdown alternate](../web/markdown.go) of one page as text, with its metadata as structured content. `path` is the page path or any URL the other tools return. Truncated at `MaxResponseBytes` (64 KiB by default) with a note pointing at the full page. |

Each result carries the title, description, canonical URL, Markdown URL,
language and last update of the page. All three tools are annotated read-only;
the module has no tool that changes anything.

**A page is returned only if it is public:** `Page.Indexable()` (published, not
scheduled, not `NoIndex`) and with a Markdown source (`Page.MarkdownSource()`:
Markdown content type, content a `string` or `[]byte`). The same rule decides
the sitemap and `llms.txt`, so an agent sees exactly what a search engine may.
`get_page` answers a page that does not exist and one that is not public with
the same "page not found", so a client cannot probe for drafts.

**Indexable is no access control.** A page for logged-in readers only that
search engines may still index — content behind a paywall — has to be left
out of the source you pass, or the tools give its text to anyone. And the
Markdown source is published as it is, HTML comments included, which your HTML
renderer may drop from the page.

**The `.md` URLs have to exist.** `markdown_url` and the note on a truncated
page point at `Site.MarkdownAlternateURL`, so serve the alternates there with
`Site.ServeMarkdown`.

**Nothing is exposed without a source.** There is no default: passing a source
is the explicit act of publishing it, and a nil source registers nothing and
returns an error. The source is usually the site itself, but can be any
`web.PageSource`, for example only the documentation of a site.

The source is read with content on every tool call. A site whose source is
expensive should pass a cached one.

### Language

The tools report and filter by a page's language: `Page.Lang`, else
`Site.Lang` (`Site.PageLang`), the language the page is rendered with, so an
agent is told the same language the HTML and the Markdown alternate declare. A
language filter matches regional variants, so `en` finds `en-GB` pages too,
and a filter that is no BCP 47 language tag is refused with an error. A
site keeping its translations under `/de/…` sets `Page.Lang` on them (the
`lang` front matter of a `GlobPageSource` page), or decides the language with
`PageToolsOptions.Language`, which `get_page` then puts into the Markdown
header too.

### Errors

A failure on the server's side — a source that cannot be read, a page whose
Markdown alternate cannot be built, a panic in the source — reaches the client
only as "the pages are unavailable". The error itself goes to
`PageToolsOptions.OnError` for your log. The SDK runs tool handlers on
goroutines of its own, outside `net/http`'s per-request recover, so the page
tools recover their panics themselves; tools of your own on the same server
need the same.
A page whose URLs cannot be built is left out of the results and reported
there, rather than taking every other page offline with it.

## Your own tools on the same server

The module never constructs the server or the mux. Product-specific tools —
prices, availability, a booking — are added to the same server with the SDK, so
an agent sees one MCP endpoint:

```go
server := mcp.NewServer(&mcp.Implementation{Name: "example", Version: "v1.0.0"}, nil)

err := mxmcp.AddPageTools(server, site, site, mxmcp.PageToolsOptions{
    OnError: func(ctx context.Context, err error) { log.Println("mcp:", err) },
})
if err != nil {
    log.Fatal(err)
}

mcp.AddTool(server, &mcp.Tool{
    Name:        "opening_hours",
    Description: "The opening hours of the shop.",
    Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
}, openingHours)
```

## Serving it over HTTP

Serve the server with the SDK's Streamable HTTP handler in stateless mode — the
page tools keep no state between calls, so no session has to survive between
requests or be pinned to one instance — and wrap it in `GuardHandler`:

```go
handler := mcp.NewStreamableHTTPHandler(
    func(*http.Request) *mcp.Server { return server },
    &mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true},
)
mux.Handle("/mcp", mxmcp.GuardHandler(handler, mxmcp.GuardOptions{
    AllowedOrigins: []string{"https://example.com"},
}))
```

`GuardHandler` does what every public MCP endpoint needs before a request
reaches the SDK:

- **Origin check.** A request with an `Origin` header not in `AllowedOrigins`
  gets 403, as the MCP specification requires, so a web page cannot make its
  visitors' browsers call the endpoint. Requests without an `Origin` pass by
  default, because desktop clients and server-side agents do not send one; set
  `RequireOrigin` to refuse them. Browsers send an `Origin` with same-origin
  POSTs too, so list the site's own origin if its pages call the endpoint. The
  guard only rejects and sends no CORS headers: a page on another origin needs
  CORS set up by your own middleware as well.
- **Host check.** With `AllowedHosts` set, a request for any other `Host` gets
  403. That stops DNS rebinding, where a page points its own domain at an
  address in its visitors' network: those requests are same-origin and a GET
  carries no `Origin`, but the `Host` is still the page's domain. The SDK
  handler already does this for a server on a loopback address; set
  `AllowedHosts` for one on a private network.
- **Body limit.** A request body above `MaxRequestBytes` (64 KiB by default)
  gets 413 before the handler reads any of it.
- **No batches.** A JSON-RPC batch gets 400. MCP dropped batching in its
  2025-06-18 revision, but the SDK still accepts one from a client that sends
  no protocol version, and a single request with hundreds of tool calls in it
  would bypass every per-call limit. `AllowBatches` lets them through for
  older clients that batch.
- **No caching.** Every response is marked `Cache-Control: no-store,
  no-transform`, replacing the `no-cache` the SDK sets on its own responses;
  `no-transform` keeps a proxy from holding back an event stream.

Rate limiting is deliberately not part of it: put your own middleware in front,
it knows your clients. With `AllowBatches` set, have it count tool calls rather
than HTTP requests.

The complete wiring is in [`example_test.go`](example_test.go).
