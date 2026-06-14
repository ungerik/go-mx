# go-mx

[![Go Reference](https://pkg.go.dev/badge/github.com/ungerik/go-mx.svg)](https://pkg.go.dev/github.com/ungerik/go-mx)
[![Go Report Card](https://goreportcard.com/badge/github.com/ungerik/go-mx)](https://goreportcard.com/report/github.com/ungerik/go-mx)
[![Go version](https://img.shields.io/github/go-mod/go-version/ungerik/go-mx)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

go-mx builds HTML in Go. Every HTML element is a Go function call, and calls
compose into a tree of components. Rendering walks the tree and writes HTML,
escaping text and attribute values and checking structural validity as it goes.

No templates and no separate template language: markup is ordinary Go, so it
composes and refactors with the rest of your code.

For the *why* — server-rendered HTML versus the SPA default, an honest
advantages-vs-disadvantages overview, and the use cases go-mx is built for —
see [docs/why-go-mx.md](docs/why-go-mx.md).

## Documentation

- **[Documentation site](https://ungerik.github.io/go-mx/)** — overview, the
  `html` and `shadcn` tutorials and how-to guides, and the rendered component
  gallery. (GitHub Pages, served from [`docs/`](docs/).)
- **[Component gallery](https://ungerik.github.io/go-mx/gallery/)** — every
  ported shadcn/ui component shown next to its Go source.
- **[`html` package](html/README.md)** — the core HTML vocabulary: elements,
  attributes, typed keyword enums, and the `html/entity` character set.
- **[`shadcn` package](shadcn/README.md)** — the shadcn/ui port: full component
  reference and design notes.
- **[API reference](https://pkg.go.dev/github.com/ungerik/go-mx)** on pkg.go.dev.

## Install

```sh
go get github.com/ungerik/go-mx
```

Requires Go 1.26 or newer.

## Example

```go
package main

import (
	"fmt"

	"github.com/ungerik/go-mx/html"
)

func main() {
	page := html.Div(html.Class("greeting"),
		html.H1("Hello, world"),
		html.P("Built with go-mx."),
	)
	fmt.Println(page) // *mx.Element implements fmt.Stringer
}
```

For streaming output or serving over HTTP, render a `Component` into an
`mx.Writer` with `Component.Render(ctx, writer)`, or use
`mx.ComponentHTTPHandler`.

## Packages

- **`mx`** (root) — core abstractions: `Component` (anything that renders),
  `Element`, `Attrib`, `Writer`, and `CheckedWriter` (escaping, structural
  validation, optional indentation). Plus conditional rendering (`If`,
  `ForEach`), struct-reflection helpers, and `ReflectFormHandler` (see
  below).
- **`html`** — HTML5 element and attribute constructors: `html.Div`,
  `html.Span`, `html.Class`, `html.ID`, and the rest of the HTML5 surface,
  with typed keyword enums that validate at render time and numeric
  attribute constructors that format Go values for you. Provides
  `html.FieldDecider` for plain-HTML form rendering. The
  [`html/entity`](html/entity) subpackage adds named character references
  (`entity.Copyright`, `entity.Heart`, …).
- **`svg`** — SVG element and attribute constructors, mapped the same way as
  `html`: `xmlns` namespace handling, spec-typed numeric attribute values, and
  typed keyword enums. Use `svg.Root` for a standalone document or `svg.SVG`
  for inline `<svg>`. See [svg/README.md](svg/README.md).
- **`hx`** — HTMX integration: `hx.Get`, `hx.Post`, `hx.Trigger`. Provides
  `hx.FieldDecider` that wraps `html.FieldDecider` and adds
  `hx-trigger="change"` to live inputs.
- **`shadcn`** — `Cn`, a faithful Go port of tailwind-merge v3, plus ported
  shadcn/ui components and `shadcn.FieldDecider` for Tailwind/shadcn form
  rendering. See [shadcn/README.md](shadcn/README.md).
- **`highlight`** — Go syntax highlighter: tokenizes Go source and renders it as
  highlighted HTML components, or as the go-mx source that builds that markup,
  plus a `Theme` that emits CSS. Depends only on `mx` and `html`. See
  [highlight/README.md](highlight/README.md).
- **`web`**, **`doc`**, **`pdf`** — higher-level abstractions, partially
  implemented.

## Reflected forms

`mx.ReflectFormHandler[T]` builds a full http.Handler for a struct
type — rendering, parsing, validation, and load-then-apply against
your record store, in one call:

```go
import (
    "net/http"
    "github.com/ungerik/go-mx"
    "github.com/ungerik/go-mx/shadcn"
)

mux := http.NewServeMux()
mux.Handle("/admin/profile", mx.ReflectFormHandler(loadProfile, saveProfile))
http.ListenAndServe(":8080", mx.Middleware(shadcn.FieldDecider)(mux))
```

The decider lives in request context. A custom `FieldDecider` can be
installed via `mx.Middleware(...)` once per route subtree, or passed
as an optional variadic to a single handler. Validation runs a
richest-first chain (`Normalize() []error` → `Normalize() error` →
`Validate() error` → `Valid() bool`), per-field tags drive widget
choice, and POST parses only the fields the form actually rendered —
mass-assignment-safe by construction.

See [`cmd/example-form`](cmd/example-form/main.go) for a complete
worked example covering every supported field kind, section grouping,
and the `FieldErrors` cross-field error routing path.

## To Do

- [ ] ReflectMarkup()
- [ ] More ReflectInputOptions
- [ ] Slice-of-struct fields (`form:"repeatable"`)
- [ ] Rich file upload widget (preview, multi-file, progress)
- [ ] HTMX OOB fragment responses
- [ ] OptionsProvider registry sub-package (ISO 4217, ISO 639, country codes)

## License

go-mx is licensed under the [MIT License](LICENSE), with one exception: the
[`shadcn/cva`](shadcn/cva/) subpackage is a port of
[class-variance-authority](https://github.com/joe-bell/cva) and is licensed
under the Apache License, Version 2.0 (see [`shadcn/cva/LICENSE`](shadcn/cva/LICENSE)).

Parts of the project port or adapt other open-source work (shadcn/ui, clsx,
tailwind-merge, and GitHub's primer syntax palette); their license texts and
copyright notices are reproduced in
[THIRD-PARTY-LICENSES.md](THIRD-PARTY-LICENSES.md).
