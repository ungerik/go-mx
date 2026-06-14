---
title: html how-to guides
---

# `html` how-to guides

Task-oriented recipes for the [`html`](https://pkg.go.dev/github.com/ungerik/go-mx/html)
package. Each one assumes you can already build elements (`html.Div(...)`); if
not, start with the **[tutorial](tutorial.html)**.

- [Render a component to a string](#render-a-component-to-a-string)
- [Serve a full HTML document over HTTP](#serve-a-full-html-document-over-http)
- [Add CSS and JavaScript to a page](#add-css-and-javascript-to-a-page)
- [Set boolean attributes and keyword enums](#set-boolean-attributes-and-keyword-enums)
- [Add data-attributes and ARIA](#add-data-attributes-and-aria)
- [Insert trusted raw HTML safely](#insert-trusted-raw-html-safely)
- [Render a list and conditional markup](#render-a-list-and-conditional-markup)
- [Build a form from a struct](#build-a-form-from-a-struct)
- [Use an element or attribute the package doesn't define](#use-an-element-or-attribute-the-package-doesnt-define)
- [Pretty-print the output](#pretty-print-the-output)

## Render a component to a string

Useful for tests, email bodies, or anywhere you need the HTML as a value rather
than streamed to a response.

```go
import (
    "context"
    "strings"

    "github.com/ungerik/go-mx"
    "github.com/ungerik/go-mx/html"
)

func renderString(c mx.Component) (string, error) {
    var b strings.Builder
    w := mx.NewCheckedWriter(&b)
    if err := c.Render(context.Background(), w); err != nil {
        return "", err
    }
    return b.String(), nil
}
```

**Verification:** `renderString(html.P("hi"))` returns `<p>hi</p>`, nil.

**Shortcut for tests:** `(*mx.Element).String()` renders the element with
single-quoted attributes and swallows nothing — a render error comes back as the
string `mx.Element.String: <error>`. Handy in assertions:

```go
got := html.A(html.HRef("/"), "Home").String() // <a href='/'>Home</a>
```

## Serve a full HTML document over HTTP

```go
page := html.NewDocument("Dashboard",
    html.H1("Dashboard"),
    html.P("Welcome back."),
)
http.HandleFunc("/", page.HandleHTTP) // HandleHTTP is an http.HandlerFunc
log.Fatal(http.ListenAndServe(":8080", nil))
```

`HandleHTTP` renders with an indenting, escaping writer and sets
`Content-Type: text/html; charset=utf-8`. For a one-document program,
`page.Serve(":8080")` is even shorter. To serve any component (not a full
document), use `html.Serve(addr, component)`.

**Verification:** `curl -s localhost:8080` shows the `<!DOCTYPE html>` page with
your `<h1>` and `<p>` inside `<body>`.

**Troubleshooting:** a render error becomes a `500` with a generic message (the
error string is never sent to the client). Log it server-side if you need the
detail.

## Add CSS and JavaScript to a page

`Document` carries stylesheets and inline style as fields; add scripts as body
or head children.

```go
doc := html.NewDocument("Styled", html.H1("Hi"))
doc.Stylesheets = []string{"/static/app.css"}    // <link rel="stylesheet" href=...>
doc.Style = "body{font-family:system-ui}"          // inline <style> after the links
doc.HeadCustom = html.Script(html.Src("/static/app.js"), html.Defer)
```

For an inline stylesheet element anywhere, use `html.StyleElem(css)` (it wraps
the CSS in `Raw`, so it is not escaped). For inline JS, put the code in a
`html.Script(html.Raw(js))` — and remember `Raw` is unescaped, so never build it
from user input.

## Set boolean attributes and keyword enums

Boolean attributes are constants — include them to turn them on:

```go
html.Input(html.Type("checkbox"), html.Checked, html.Disabled)
// <input type="checkbox" checked="checked" disabled="disabled"/>
```

Closed-keyword attributes are typed enums; use the named constant:

```go
html.Div(html.DirRTL, html.SpellCheckFalse)
html.Img(html.Src("/a.png"), html.LoadingLazy, html.DecodingAsync)
html.Form(html.MethodPOST, html.EncTypeMultipartFormData)
```

For a value computed at runtime, convert a string: `html.Dir(userDir)`. If the
string is not a valid keyword, rendering the element fails with a clear error
instead of emitting bad markup — check it up front with `.Valid()` if you need to
branch:

```go
if d := html.Dir(userDir); d.Valid() {
    el = html.Div(d, content)
}
```

## Add data-attributes and ARIA

```go
html.Div(
    html.DataAttr("user-id", "42"),       // data-user-id="42"
    html.Attrib("aria-live", "polite"),    // any name/value pair
    html.Role("status"),                    // role="status"
    "Saved",
)
```

`html.DataAttr(name, value)` prefixes `data-`; `html.Attrib(name, value)` emits
any literal name/value pair for attributes the package doesn't name (most ARIA
attributes).

## Insert trusted raw HTML safely

A bare `string` child is always escaped. To emit HTML you already trust (a
Markdown render, a stored snippet you control), wrap it in `Raw`:

```go
html.Article(html.Raw(renderedMarkdownHTML))   // emitted verbatim
html.P("<script>alert(1)</script>")             // escaped: &lt;script&gt;…
```

**Never** pass user input to `Raw` — that is an XSS hole. Escape user-supplied
strings by passing them as plain `string` children (automatic) or with
`html.Escape(s)` when you need the value.

## Render a list and conditional markup

```go
items := []string{"Alpha", "Beta", "Gamma"}

html.UL(
    html.ForEach(items, func(s string) *mx.Element {
        return html.LI(s)
    }),
)

html.If(loggedIn,
    html.Span("Welcome back"),
).Else(
    html.A(html.HRef("/login"), "Log in"),
)
```

`ForEach` maps a slice to components; `ForEachIter` does the same for an
`iter.Seq`. `If(...).Else(...)` also offers `.ElseIf(cond, …)`; `Iff(func() bool,
…)` takes the condition as a func that is called immediately, not at render time.

## Build a form from a struct

Render, parse, and validate a form for a struct type with
`mx.ReflectFormHandler` plus `html.FieldDecider`.

**Prerequisites:** a `*T` to bind, and an `onSubmit` to persist it.

```go
import (
    "context"
    "net/http"

    "github.com/ungerik/go-mx"
    "github.com/ungerik/go-mx/html"
)

type Signup struct {
    Email string `form:"required,placeholder=you@example.com"`
    Age   int    `form:"min=13"`
    Bio   string `form:"widget=textarea,help=Tell us about yourself"`
}

func main() {
    handler := mx.ReflectFormHandler(
        nil, // load: nil for a submit-only form (seeded with new(Signup))
        func(ctx context.Context, s *Signup) error {
            // persist s here; return mx.FieldErrors{"Email": err} for inline errors
            return nil
        },
        html.FieldDecider,
    )
    http.Handle("/signup", handler)
    http.ListenAndServe(":8080", nil)
}
```

The handler renders the `<form>` on GET and, on POST, parses **only the fields it
rendered** (an allowlist against mass-assignment), validates, and either
re-renders with inline errors or 303-redirects on success.

**Editing an existing record:** pass a non-nil `load` that returns the current
`*T`; the form is seeded with its values.

**App-wide decider:** instead of passing `html.FieldDecider` to every handler,
install it once with middleware:

```go
mux := http.NewServeMux()
mux.Handle("/signup", mx.ReflectFormHandler(nil, onSubmit)) // no decider arg
http.ListenAndServe(":8080", mx.Middleware(html.FieldDecider)(mux))
```

**Troubleshooting:** if the page renders the literal text *"no FieldDecider in
request context"*, you neither passed a decider nor wrapped the handler in
`mx.Middleware(html.FieldDecider)`. Do one of the two.

## Use an element or attribute the package doesn't define

`Element`, `VoidElement`, and `Attrib` are the escape hatches for custom tags
(web components) and attributes the package doesn't name:

```go
html.Element("my-widget", html.Attrib("variant", "compact"), "child text")
// <my-widget variant="compact">child text</my-widget>

html.VoidElement("custom-input", html.ID("x"))
// <custom-input id="x"/>
```

For a custom boolean attribute, declare a `BoolAttrib` constant:

```go
const HxBoost = html.BoolAttrib("hx-boost")
html.Div(HxBoost) // <div hx-boost></div>
```

(For htmx specifically, the [`hx`](https://pkg.go.dev/github.com/ungerik/go-mx/hx)
package already provides typed constructors.)

## Pretty-print the output

Wrap the writer with `WithIndent` for human-readable, indented HTML:

```go
w := mx.NewCheckedWriter(os.Stdout).WithIndent("", "  ")
html.NewDocument("Demo", html.H1("Hi")).Render(context.Background(), w)
```

`Document.HandleHTTP` and `html.Serve` already indent with two spaces. Drop
`WithIndent` for compact, single-line output (smaller payloads in production).

## See also

- **[Tutorial](tutorial.html)** — build your first page from scratch.
- **[Package reference (README)](https://github.com/ungerik/go-mx/blob/main/html/README.md)**
  — every element, attribute, and helper.
- **[API on pkg.go.dev](https://pkg.go.dev/github.com/ungerik/go-mx/html)**
