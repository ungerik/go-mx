# highlight: design notes

Why the `highlight` package is built the way it is. For the API and usage, see
[README.md](README.md) and the godoc; this file is the "why".

## The problem

go-mx builds HTML programmatically from Go. Showing Go source on a go-mx page
needs three things that pull in different directions:

1. **Highlighted markup** that fits the component model, not a wall of
   string-concatenated `<span>`s.
2. **The go-mx source that produces that markup.** go-mx's whole pitch is "write
   HTML as Go", so a docs site wants to show the `html.Pre(html.Span(...))` code
   a snippet compiles to, not just the rendered output.
3. **Theme independence.** The same highlighted page should restyle from light
   to dark without re-tokenizing or re-rendering.

A single function that returns an HTML string solves none of these well. It
can't be composed into a component tree, it can't show its own source, and it
bakes colors into the markup.

## The approach

Tokenize once, render many ways. One pass over the source produces a flat token
stream; everything downstream is a projection of that stream.

```
                       ┌─ Component / Inline / HTML ─► mx.Component  ─► <pre><span>…
  src ─► TokenizeGo ─►  │
        []Token         └─ GoSource ──────────────────► string       ─► html.Pre(html.Span(…))

  Theme ─► CSS / StyleElement ─► .hl-keyword { color: … }   (independent of the above)
```

- `TokenizeGo` is the only thing that reads Go. It returns `[]Token`, each a
  slice of the source tagged with a semantic `TokenClass`.
- The two HTML/source backends are pure functions of that token slice plus a
  `Highlighter` config. They never re-parse.
- `Theme` never touches tokens. It maps classes to colors and emits CSS. The
  markup carries class names (`hl-keyword`), the theme carries colors, and they
  meet only in the browser.

This is why both backends produce identical structure: `GoSource` is literally
the program that, run through the html package, yields what `Component` renders.

### Why `go/scanner`

The standard library already contains a correct, fast Go lexer. Using it means:

- **Zero third-party dependencies.** The package imports only `go/scanner`,
  `go/token`, `go/format`, and the go-mx packages.
- **Correct tokenization for free** — raw strings, rune literals, imaginary
  numbers, every operator, are handled by the same code the compiler's
  front end is modeled on.
- **Leniency.** Initialized with a nil error handler, the scanner keeps going
  past syntax errors instead of bailing. A half-typed snippet still highlights.

The cost is that `go/scanner` is a lexer, not a type checker. It can tell a
keyword from an identifier, but it cannot tell that `MyType` is a type. See
[classification](#classification-is-lexical-not-semantic).

### Byte-faithful round-trip

The scanner reports tokens but skips whitespace and comments-as-trivia, and it
inserts synthetic semicolons at line ends that have no bytes in the source. A
highlighter that emitted only the scanned tokens would drop indentation and
blank lines, which is fatal inside `<pre>` where whitespace is the layout.

`TokenizeGo` reconstructs the original byte-for-byte. It records each token's
byte offset, then for every token emits the gap `src[prevEnd:tokenStart]` as a
plain-text token before the token itself. Auto-inserted semicolons (reported
with literal `"\n"`) contribute no bytes; the newline they stand for is picked
up as part of the next gap.

The invariant: **concatenating every `Token.Text` reproduces the input exactly.**
This holds for CRLF files, raw strings with embedded carriage returns,
unterminated literals, NUL bytes, BOMs, and multibyte identifiers. The
round-trip test (`highlight_test.go`) and a fuzz pass over hostile inputs both
enforce it, and it is what lets the HTML backend preserve source layout and the
`GoSource` backend stay lossless.

### Classification is lexical, not semantic

`classify` works from token kind plus a small amount of local context:

- Comments, strings, numbers, and keywords come straight from the token kind.
- Identifiers are refined against three sets of Go's predeclared names
  (`predeclared.go`): types (`int`, `error`, …), constants (`true`, `nil`,
  `iota`), and builtins (`make`, `len`, …).
- A one-token lookahead marks an identifier followed by `(` as a function or
  method call, and an identifier right after `func` as a declaration name.

That is the whole heuristic. It deliberately stops short of type resolution: a
user-defined `MyType` lexes as a plain identifier because proving it is a type
would require loading the package and running the type checker, which is a
different order of dependency and cost. The trade is a little under-coloring of
user types for a tokenizer that runs on a snippet with no build context.

### Why operators and punctuation are plain by default

`DefaultHighlighted` wraps keyword, type, function, builtin, constant, string,
number, and comment. Operators, punctuation, and plain identifiers render as
text. Two reasons:

- **Smaller markup.** Punctuation is the most frequent token kind in Go.
  Wrapping every `(`, `)`, `,`, and `.` in a `<span>` multiplies output size for
  little visual gain.
- **It matches what readers expect.** GitHub and most editors leave punctuation
  in the base text color.

The classes still exist, so a caller who wants colored operators sets
`Highlighted[ClassOperator] = true` and adds a theme color. Nothing is lost,
the default is just lean.

### The `GoSource` backend and gofmt

`GoSource` walks the same tokens and prints `html.Pre(...)` calls. Two details
make the output clean:

- **Each child on its own line.** gofmt normalizes indentation but never breaks
  a single-line call into multiple lines, so the generator emits the newlines
  and lets `go/format` handle the tabs.
- **Statement-wrap to format an expression.** `format.Source` formats files or
  statement lists, not bare expressions, so the generator wraps the expression
  as `_ = <expr>`, formats that, and strips the `_ = ` prefix. If formatting
  ever fails, it returns the unformatted-but-valid expression rather than an
  error.

Adjacent non-highlighted tokens are merged into one string literal (`"() { fmt."`)
so the generated code reads like something a person would write.

### The non-indenting-writer constraint

The HTML backend builds real `html.Span` elements and lets the `CheckedWriter`
render them. That writer, when configured to indent, inserts a newline and
padding before each element's start tag. Inside `<pre>` those injected bytes are
visible and would shift the code. So the markup is only correct under a
non-indenting writer, which is the default `mx.NewCheckedWriter`. `HTML()`
hardcodes that writer; `Component`/`Inline` document the requirement.

This is the one sharp edge of building the output from components instead of
writing raw bytes. The alternative — emitting spans through `mx.Raw` to bypass
the writer — was rejected because the explicit goal was markup built from the
html package, composable and inspectable like any other go-mx tree.

## Trade-offs

| Choice | Gained | Gave up |
|--------------------------------------------|------------------------------------------|------------------------------------------|
| `go/scanner` over a hand-written lexer     | Correctness, zero deps, leniency         | No type info; user types stay plain      |
| Lexical classification, no type checker    | Runs on any snippet, no build context    | `MyType` not colored as a type           |
| Two backends over one token stream         | HTML and go-mx source always agree       | A second backend to keep in sync         |
| Components over raw byte output            | Composable, inspectable, escapes safely  | Requires a non-indenting writer          |
| Plain operators/punctuation by default     | Lean markup, editor-like look            | Opt-in needed for colored operators      |
| gofmt the generated source                 | Output reads hand-written                | A `go/format` pass per `GoSource` call   |

## Alternatives considered

- **Return an HTML string instead of components.** Simplest, but un-composable,
  can't show its own source, and re-implements escaping. Rejected; the component
  path gives all three for the cost of the writer constraint.
- **Bake colors into the HTML (inline styles).** Removes the separate CSS step
  but kills theme switching and bloats every span. Rejected for the
  class-plus-`Theme` split.
- **Resolve types with `go/types`.** Would color user-defined types and method
  receivers correctly, at the cost of needing a loadable package and a full type
  pass. Out of scope for a snippet highlighter; revisit if a "highlight a whole
  package" mode is ever added.
