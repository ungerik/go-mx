# go-mx

go-mx builds HTML in Go. Every HTML element is a Go function call, and calls
compose into a tree of components. Rendering walks the tree and writes HTML,
escaping text and attribute values and checking structural validity as it goes.

No templates and no separate template language: markup is ordinary Go, so it
composes and refactors with the rest of your code.

For the *why* — server-rendered HTML versus the SPA default, an honest
advantages-vs-disadvantages overview, and the use cases go-mx is built for —
see [docs/why-go-mx.md](docs/why-go-mx.md).

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
  `html.Span`, `html.Class`, `html.ID`, and the rest of the HTML5 surface.
  Provides `html.FieldDecider` for plain-HTML form rendering.
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
