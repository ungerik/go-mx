---
title: go-mx
---

# go-mx

**Type-safe, component-based HTML generation for Go.** Every HTML element is a Go
function call, and calls compose into a tree of components that renders to
escaped, validated HTML. No templates and no separate template language: markup
is ordinary Go, so it composes and refactors with the rest of your code.

```go
html.Div(html.Class("card"),
    html.H2("Hello, go-mx"),
    html.P("Rendered in Go, escaped and validated."),
)
```

## Where to go next

- **[html package docs](html/)** — a tutorial, task-oriented how-to guides, and
  the full reference for the `html` package: the core HTML vocabulary everything
  else builds on.
- **[Component gallery](shadcn/gallery/)** — every ported shadcn/ui component, rendered
  server-side in Go and shown next to its source. A pre-rendered static copy of
  the live `cmd/shadcn-gallery` app.
- **[shadcn package docs](shadcn/)** — a tutorial, task-oriented how-to guides,
  and the full component reference for the `shadcn` package.
- **[Why go-mx?](why-go-mx.html)** — server-rendered HTML versus the SPA default:
  an honest advantages-and-disadvantages look at where this approach fits.
- **[API reference on pkg.go.dev](https://pkg.go.dev/github.com/ungerik/go-mx)**
- **[Source on GitHub](https://github.com/ungerik/go-mx)**

## Install

```sh
go get github.com/ungerik/go-mx
```

Requires Go 1.26 or newer.

## Hello, go-mx

A complete HTTP server that renders a page:

```go
package main

import (
    "log"
    "net/http"

    "github.com/ungerik/go-mx/html"
)

func main() {
    page := html.NewDocument("Hello",
        html.H1("Hello, go-mx"),
        html.P("Type-safe HTML, rendered in Go — no templates."),
    )
    http.HandleFunc("/", page.HandleHTTP)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

`html.NewDocument` returns an `*html.Document` whose `HandleHTTP` method is a
plain `http.HandlerFunc`. It renders the `<!DOCTYPE html>`, `<head>` and `<body>`
for you with an escaping, indenting writer.

## Core concepts

**Component.** The one interface everything implements:

```go
type Component interface {
    Render(ctx context.Context, w Writer) error
}
```

Anything that can write HTML is a `Component`. Elements, text, documents, and
your own types all satisfy it.

**Elements.** Regular elements take attributes and children as variadic `any`;
void elements take only attributes:

```go
html.Div(html.Class("container"), html.P("Hello"))   // <div class="container"><p>Hello</p></div>
html.Img(html.Src("/logo.png"), html.Alt("Logo"))    // void: no children
```

**Children conversion.** Children are converted at render-build time: a `string`
becomes escaped text, a `Component` passes through, a `func` is wrapped, `nil`
renders nothing, an `error` or `fmt.Stringer` renders its text, and anything
else is rendered with [go-pretty](https://github.com/domonda/go-pretty) as a
compact, single-line, type-tagged representation, escaped as text. So you can
mix strings and components freely, and a stringified value can never inject
markup — escaping is done by the writer, so it is the same for HTML, XHTML, SVG
and XML. (Because the fallback is silent, a value you *meant* as markup but that
does not implement `Component` renders as its escaped pretty representation;
convert such children to a `Component` explicitly.) The conversion is the
package-level `mx.AsComponent` variable, which you can reassign to change this
behavior program-wide.

**Attributes.** `html.Class`, `html.ID`, `html.HRef`, `html.Type`, and the rest
map one-to-one to HTML attributes. Values are escaped. You can also pass a slice,
map, or a struct with `attr` tags and let go-mx expand it.

**Conditional rendering and iteration:**

```go
mx.If(loggedIn, logoutButton).Else(loginButton)

mx.ForEach(items, func(s string) mx.Component {
    return html.LI(s)
})
```

**Writer.** `mx.NewCheckedWriter(w)` wraps any `io.Writer`. It escapes text and
attribute values, validates structure (for example, it rejects two attributes
with the same name on one element), and can pretty-print with
`.WithIndent("", "  ")`.

## Packages

| Package      | What it gives you                                                     |
|--------------|----------------------------------------------------------------------|
| `html`       | HTML5 elements (`Div`, `Span`, `Input`, …) and attributes (`Class`, `ID`, …) |
| `svg`        | SVG elements and attributes, with `xmlns` handling and numeric values |
| `xml`        | Generic XML markup: `Element`/`Attrib` take any name (no fixed vocabulary), plus `Comment`, `CDATA`, `ProcInst`, the XML declaration and a `Document` |
| `hx`         | [htmx](https://htmx.org) attributes (`hx.Get`, `hx.Post`, …), with typed values where htmx constrains them (a `hx.SwapStyle` enum for `hx.Swap`, `bool` args for `hx.Boost`), `htmx:` event and `htmx-*` class constants, plus request readers and response-header setters for handlers |
| `shadcn`     | A Go port of [shadcn/ui](https://ui.shadcn.com) components, plus the `Cn` class-merge helper and ports of `clsx`, `tailwind-merge` and `cva` |
| `highlight`  | A dependency-free Go syntax highlighter built from go-mx components   |
| `pdf`        | Composable PDF rendering primitives (a separate module) mirroring the `html` component model on [fpdf](https://codeberg.org/go-pdf/fpdf) |

The root `mx` package holds the core abstractions (`Component`, `Element`,
`Writer`, `If`/`ForEach`) that the others build on. Higher-level packages
(`web`, `doc`) are partially implemented.

## The shadcn port

The [`shadcn`](shadcn/) package reproduces shadcn/ui's markup and Tailwind
classes in Go and renders them server-side with no client framework. Behavior
React delegates to Radix is re-expressed with web-platform primitives — native
`<dialog>`, the Popover API, CSS Anchor Positioning, native form controls — so
most components ship zero or near-zero JavaScript.

See the **[component gallery](shadcn/gallery/)** for every component rendered live next
to its Go source, and the **[shadcn docs](shadcn/)** to start building.

## Documentation map

This site follows the [Diátaxis](https://diataxis.fr) framework — four kinds of
documentation for four reader needs:

| Need                         | Where                                            |
|------------------------------|--------------------------------------------------|
| **Learn** (tutorial)         | [Build your first page](html/tutorial.html) · [first shadcn page](shadcn/tutorial.html) |
| **Do a task** (how-to)       | [html how-to](html/how-to.html) · [shadcn how-to](shadcn/how-to.html) |
| **Look up** (reference)      | [html reference](html/#reference) · [shadcn reference](shadcn/#reference) · [pkg.go.dev](https://pkg.go.dev/github.com/ungerik/go-mx) |
| **Understand** (explanation) | [Why go-mx](why-go-mx.html) · [html design notes](html/#explanation) · [shadcn design notes](shadcn/#explanation) |
