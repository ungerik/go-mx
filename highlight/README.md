# highlight

[![Go Reference](https://pkg.go.dev/badge/github.com/ungerik/go-mx/highlight.svg)](https://pkg.go.dev/github.com/ungerik/go-mx/highlight)

A Go syntax highlighter for [go-mx](../). It tokenizes Go source with the
standard-library `go/scanner` (no third-party dependencies) and renders it two
ways from the same tokens:

- **Highlighted HTML**, built through `mx`/`html` components — every meaningful
  token becomes a `<span class="hl-…">`.
- **Generated Go source** that uses the `html` package to build that same
  markup — a code generator, not an echo of the input.

Colors live in a separate [`Theme`](theme.go) that emits CSS, so the same HTML
works with any theme.

The package depends only on the root `mx` package and the `html` element
helpers; it does **not** depend on `shadcn`, so it can be used on its own or
dropped into any go-mx markup, including the shadcn UI.

For why it is built this way (two backends, `go/scanner`, byte-faithful
round-trip, trade-offs), see [DESIGN.md](DESIGN.md).

## Tutorial: highlight Go on a web page

This walks from nothing to a styled HTML page you open in a browser. You need a
Go module that can import go-mx.

### Step 1: write the program

Create `main.go`. It highlights a snippet, puts the theme's CSS in the page
`<head>`, and writes the whole page to stdout:

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/highlight"
	"github.com/ungerik/go-mx/html"
)

func main() {
	src := `package main

import "fmt"

func main() {
	fmt.Println("hello, highlight")
}
`
	page := html.HTML(
		html.Head(highlight.LightTheme.StyleElement("")),
		html.Body(highlight.Component(src)),
	)

	fmt.Println("<!DOCTYPE html>")
	if err := page.Render(context.Background(), mx.NewCheckedWriter(os.Stdout)); err != nil {
		panic(err)
	}
}
```

Three calls do the work: `highlight.Component(src)` builds the `<pre><code>`
tree, `highlight.LightTheme.StyleElement("")` builds the `<style>` that colors
it, and `mx.NewCheckedWriter` renders the tree to bytes (it is non-indenting, so
the code layout inside `<pre>` survives).

### Step 2: run it into a file

```bash
go run . > highlighted.html
```

`highlighted.html` is now a complete page: `<!DOCTYPE html>`, a `<head>` with
`pre.hl { … }` plus one `.hl-keyword { … }` rule per token class, and a `<body>`
whose `<pre>` holds spans like `<span class="hl-keyword">func</span>`.

### Step 3: open it

```bash
open highlighted.html      # macOS; use xdg-open on Linux
```

You see the snippet with red keywords, purple function names, and a gray italic
comment on a light background. Switch the look by swapping `LightTheme` for
`DarkTheme` in both the `StyleElement` call and (if you want) nothing else: the
markup is theme-independent, so only the `<style>` changes.

### What you built

A self-contained highlighter page with no client-side JavaScript and no
third-party dependencies. From here, `Component`/`Inline`/`HTML` cover the markup
and `GoSource` emits the go-mx code that builds it; the [Usage](#usage) section
below has each one.

## Usage

```go
import "github.com/ungerik/go-mx/highlight"

const src = `func main() { fmt.Println("hi") }`
```

### Highlighted HTML

```go
block := highlight.Component(src) // *mx.Element: <pre class="hl"><code>…</code></pre>
inline := highlight.Inline(src)   // *mx.Element: <code class="hl">… (no <pre>)
s, err := highlight.HTML(src)     // render directly to an HTML string
```

`Component` renders to:

```html
<pre class="hl"><code><span class="hl-keyword">func</span> <span class="hl-function">main</span>() { fmt.<span class="hl-function">Println</span>(<span class="hl-string">&quot;hi&quot;</span>) }</code></pre>
```

Render with a non-indenting writer (the default `mx.NewCheckedWriter`); an
indenting writer would inject whitespace inside `<pre>` and corrupt the layout.

### Generated go-mx source

```go
code := highlight.GoSource(src)
```

returns gofmt-formatted Go source that reproduces the markup above:

```go
html.Pre(html.Class("hl"),
	html.Code(
		html.Span(html.Class("hl-keyword"), "func"),
		" ",
		html.Span(html.Class("hl-function"), "main"),
		"() { fmt.",
		html.Span(html.Class("hl-function"), "Println"),
		"(",
		html.Span(html.Class("hl-string"), "\"hi\""),
		") }",
	),
)
```

### CSS

```go
css := highlight.LightTheme.CSS("")          // stylesheet string
style := highlight.DarkTheme.StyleElement("") // *mx.Element: <style>…</style>
```

Built-in themes: `LightTheme` and `DarkTheme` (GitHub-like palettes). Pass `""`
to match the default `hl-` class prefix.

## Customizing

```go
h := &highlight.Highlighter{
	Prefix:      "syn-",                                      // class prefix; block class becomes "syn"
	Highlighted: map[highlight.TokenClass]bool{               // which classes get a <span>
		highlight.ClassKeyword:  true,
		highlight.ClassString:   true,
		highlight.ClassOperator: true, // off by default
	},
}
out := h.Component(src)
```

By default the highlighted classes are keyword, type, function, builtin,
constant, string, number and comment; operators, punctuation and plain
identifiers render as text to keep the markup small.

## Token classes

`keyword`, `type`, `function`, `builtin`, `constant`, `string`, `number`,
`comment`, `operator`, `punctuation`, `ident`. Each maps to a `Style` in a
`Theme` and to a CSS class `<prefix><class>` (e.g. `hl-keyword`).

## Demo

```bash
go run ./cmd/highlight-demo
# browse to http://localhost:8080  (add ?theme=dark to switch)
```

The demo highlights a sample file, shows the generated `GoSource` output
(itself highlighted) next to it, and injects the theme CSS into the page head.

## Design

[DESIGN.md](DESIGN.md) explains the why: tokenize-once with two backends, the
`go/scanner` choice, the byte-faithful round-trip that preserves `<pre>` layout,
why classification is lexical, and the trade-offs behind each call.
