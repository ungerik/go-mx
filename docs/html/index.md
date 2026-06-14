---
title: html for go-mx
---

# The `html` package

The core HTML vocabulary for go-mx. Every HTML5 element is a Go function
returning a `*mx.Element` (`Div`, `Span`, `Img`, …) and every attribute is a
function or constant returning an `mx.Attrib` (`Class`, `ID`, `HRef`, …). Calls
compose into a tree that renders to escaped, validated HTML — no templates, no
separate template language.

```go
import "github.com/ungerik/go-mx/html"

html.Div(html.Class("card"),
    html.H2("Hello, go-mx"),
    html.P("Rendered in Go, escaped and validated."),
)
```

It is the foundation the other packages build on:
[`svg`](https://github.com/ungerik/go-mx/blob/main/svg/README.md) mirrors it for
SVG, [`hx`](https://pkg.go.dev/github.com/ungerik/go-mx/hx) adds htmx attributes,
and [`shadcn`](../shadcn/) ports styled components on top.

This documentation follows the [Diátaxis](https://diataxis.fr) framework.

## Tutorial

New to the package? Start here.

- **[Build your first page](tutorial.html)** — from an empty module to a served,
  stateful page with a nav, a list generated from a slice, a conditional login
  link, and a test that asserts on the rendered HTML.

## How-to guides

Task-oriented recipes for things you'll actually need.

- **[html how-to guides](how-to.html)** — render a component to a string, serve a
  document, add CSS/JS, set boolean and keyword-enum attributes, add `data-*` and
  ARIA, insert trusted raw HTML, iterate and branch, build a form from a struct,
  use custom elements, and pretty-print output.

## Reference

The complete reference lives in the package README, kept next to the code so it
can't drift:

- **[html package reference](https://github.com/ungerik/go-mx/blob/main/html/README.md)**
  — elements, attributes (free values, booleans, keyword enums, events,
  `data-*`), text and escaping, conditionals, documents and serving, forms, and
  the `Attr`/`Elem` naming conventions.
- **[Full API on pkg.go.dev](https://pkg.go.dev/github.com/ungerik/go-mx/html)**

## Explanation

Why the package works the way it does:

- **[Strict keyword enums](https://github.com/ungerik/go-mx/blob/main/html/README.md#strict-keyword-enums)**
  — why closed-set attributes (`dir`, `loading`, `method`, …) are typed enums
  instead of loose strings, and how invalid values are caught.
- **[The deferred-error pattern](https://github.com/ungerik/go-mx/blob/main/html/README.md#the-deferred-error-pattern)**
  — how constructors stay chainable yet still report invalid input, at render
  time.
- **[Why server-rendered HTML at all](../why-go-mx.html)** — the approach in
  context, with its trade-offs.
