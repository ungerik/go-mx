# go-mx

go-mx builds HTML in Go. Every HTML element is a Go function call, and calls
compose into a tree of components. Rendering walks the tree and writes HTML,
escaping text and attribute values and checking structural validity as it goes.

No templates and no separate template language: markup is ordinary Go, so it
composes and refactors with the rest of your code.

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
  `ForEach`) and struct-reflection helpers.
- **`html`** — HTML5 element and attribute constructors: `html.Div`,
  `html.Span`, `html.Class`, `html.ID`, and the rest of the HTML5 surface.
- **`hx`** — HTMX integration: `hx.Get`, `hx.Post`, `hx.Trigger`, ...
- **`shadcn`** — `Cn`, a faithful Go port of tailwind-merge v3, plus ported
  shadcn/ui components: `Button`, `Alert`, `AlertDialog`. See
  [shadcn/README.md](shadcn/README.md).
- **`web`**, **`doc`**, **`pdf`** — higher-level abstractions, partially
  implemented.

## To Do

- [ ] ReflectMarkup()
- [ ] Change ReflectFormComponents to general purpose ReflectComponents
- [ ] More ReflectInputOptions
- [ ] ReflectFormHandler
  - validation
  - status code / redirect
  - custom headers
- [ ] select options, get options from struct tag, config using radio buttons instead of select
