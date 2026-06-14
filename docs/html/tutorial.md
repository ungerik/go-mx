---
title: Build your first page with html
---

# Build your first page with `html`

You'll build a small web page — a navigation bar, a heading, a dynamic list, and
a login link that changes with state — and serve it over HTTP. Everything is
written in Go with the [`html`](https://pkg.go.dev/github.com/ungerik/go-mx/html)
package: no template files, no template language. By the end you'll understand
how elements, attributes, iteration, and documents fit together.

## What you'll need

- **Go 1.26 or newer** (`go version` to check).
- A terminal and a browser.
- No prior go-mx knowledge.

## Step 1: Create a module

```sh
mkdir hello-mx && cd hello-mx
go mod init example.com/hello-mx
go get github.com/ungerik/go-mx
```

The last command adds go-mx to your `go.mod`.

## Step 2: Write a server that renders a page

Create `main.go`:

```go
package main

import (
    "log"
    "net/http"

    "github.com/ungerik/go-mx/html"
)

func main() {
    page := html.NewDocument("My First Page",
        html.H1("Hello, go-mx"),
        html.P("This page is built entirely in Go."),
    )

    http.HandleFunc("/", page.HandleHTTP)
    log.Println("listening on http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

`html.NewDocument(title, body...)` returns an `*html.Document` — a complete
`<!DOCTYPE html>` page. Its `HandleHTTP` method *is* an `http.HandlerFunc`, so it
plugs straight into the standard library.

## Step 3: Run it and see the HTML

```sh
go run .
```

Open <http://localhost:8080> (or `curl -s localhost:8080`). You'll see:

```html
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8"/>
  <title>My First Page</title>
</head>
<body>
  <h1>Hello, go-mx</h1>
  <p>This page is built entirely in Go.</p>
</body>
</html>
```

That's the whole loop: Go function calls in, escaped and indented HTML out. The
`<head>`, charset, and title were written for you. Stop the server with `Ctrl+C`
before the next step.

## Step 4: Add structure and a dynamic list

Real pages have nesting and repetition. Elements take other elements as children,
and `html.ForEach` turns a slice into a list of elements. Replace the body of
`main` with a builder function:

```go
package main

import (
    "log"
    "net/http"
    "strings"

    "github.com/ungerik/go-mx"
    "github.com/ungerik/go-mx/html"
)

func homePage() *html.Document {
    sections := []string{"Home", "About", "Contact"}

    return html.NewDocument("My First Page",
        html.Nav(
            html.UL(
                html.ForEach(sections, func(name string) *mx.Element {
                    return html.LI(
                        html.A(html.HRef("/"+strings.ToLower(name)), name),
                    )
                }),
            ),
        ),
        html.H1("Hello, go-mx"),
        html.P(html.Class("lead"), "This page is built entirely in Go."),
    )
}

func main() {
    http.HandleFunc("/", homePage().HandleHTTP)
    log.Println("listening on http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

Two new ideas:

- **Children nest.** `html.Nav(html.UL(...))` puts the `<ul>` inside the `<nav>`.
- **`html.ForEach(slice, fn)`** calls `fn` for each element and collects the
  results as children. Here each section name becomes an `<li><a>…</a></li>`.
- **`html.Class("lead")`** is an attribute. go-mx sorts arguments at build time:
  `mx.Attrib` values become attributes, everything else becomes a child — so you
  pass them in any order.

Run `go run .` again: the nav list now renders with three links.

## Step 5: Render different markup based on state

Pages change with data. `html.If(cond, ...).Else(...)` picks between two
branches at build time. Add a login control to the nav:

```go
func homePage(loggedIn bool) *html.Document {
    sections := []string{"Home", "About", "Contact"}

    loginLink := html.If(loggedIn,
        html.A(html.HRef("/logout"), "Log out"),
    ).Else(
        html.A(html.HRef("/login"), "Log in"),
    )

    return html.NewDocument("My First Page",
        html.Nav(
            html.UL(
                html.ForEach(sections, func(name string) *mx.Element {
                    return html.LI(html.A(html.HRef("/"+strings.ToLower(name)), name))
                }),
            ),
            loginLink,
        ),
        html.H1("Hello, go-mx"),
        html.P(html.Class("lead"), "This page is built entirely in Go."),
    )
}
```

Update `main` to pass a value, for example `homePage(false).HandleHTTP`. Flip it
to `true` and the link changes to "Log out". Because the page is plain Go, the
condition can be anything — a session lookup, a feature flag, a database result.

## Step 6: Test the markup without a server

Since a page is just a value, you can render it to a string and assert on it —
no HTTP needed. Create `main_test.go`:

```go
package main

import (
    "context"
    "strings"
    "testing"

    "github.com/ungerik/go-mx"
)

func TestHomePageHasLoginLink(t *testing.T) {
    var b strings.Builder
    if err := homePage(false).Render(context.Background(), mx.NewCheckedWriter(&b)); err != nil {
        t.Fatal(err)
    }
    if !strings.Contains(b.String(), `href="/login"`) {
        t.Errorf("expected a login link, got:\n%s", b.String())
    }
}
```

```sh
go test .
```

`mx.NewCheckedWriter` escapes text and attribute values and validates structure
(it rejects, for example, two attributes with the same name on one element). The
same writer backs `HandleHTTP`, so what you test is what you serve.

## What you built

A server that renders a structured, stateful HTML page entirely from Go — nav,
heading, a list generated from a slice, and a branch that changes with state —
plus a test that checks the output directly. You used:

- `html.NewDocument` and `HandleHTTP` to assemble and serve a full page.
- Element constructors (`Div`, `Nav`, `UL`, `LI`, `A`, `H1`, `P`) that nest.
- Attributes (`Class`, `HRef`) mixed freely with children.
- `html.ForEach` for iteration and `html.If(...).Else(...)` for conditionals.
- `mx.NewCheckedWriter` to render to a string for testing.

## Where to go next

- **[How-to guides](how-to.html)** — forms from structs, `data-*`/ARIA
  attributes, raw HTML, custom elements, pretty-printing, and more.
- **[Package reference (README)](https://github.com/ungerik/go-mx/blob/main/html/README.md)**
  — the complete element/attribute/helper catalog.
- **[`shadcn` package](../shadcn/)** — styled, ready-made components built on
  these same primitives.
