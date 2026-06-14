---
title: Build your first shadcn page
---

# Build your first shadcn page

You'll go from an empty Go module to a **styled sign-in card served over HTTP** —
a real page you open in a browser. Along the way you'll see how a shadcn
component renders to HTML, how to serve it, and how to wire up the Tailwind CSS
the components need. By the end you'll understand the whole loop: write Go, get
styled HTML, no client framework.

## What you'll need

- Go 1.26 or newer (`go version`)
- A web browser
- An internet connection (Tailwind is loaded from a CDN in this tutorial)

No Node, no bundler, no build step.

## Step 1: Create the module

```bash
mkdir signin && cd signin
go mod init example.com/signin
go get github.com/ungerik/go-mx
```

You now have a Go module that can import go-mx.

## Step 2: Render a component to your terminal

Before serving anything, let's see what a component *is*. Create `main.go`:

```go
package main

import (
    "context"
    "os"

    "github.com/ungerik/go-mx"
    "github.com/ungerik/go-mx/shadcn"
)

func main() {
    button := shadcn.Button(shadcn.ButtonDefault, shadcn.SizeDefault, "Sign in")
    button.Render(context.Background(), mx.NewCheckedWriter(os.Stdout))
}
```

Run it:

```bash
go run .
```

You'll see the HTML the component produced:

```html
<button data-slot="button" data-variant="default" data-size="default" type="button" class="...">Sign in</button>
```

That's the whole idea: `shadcn.Button(...)` is a Go value that knows how to write
its own HTML. `mx.NewCheckedWriter` is the writer that escapes and validates as
it goes.

## Step 3: Serve it in a browser

Now put that button on a real page. Replace `main.go` with:

```go
package main

import (
    "log"
    "net/http"

    "github.com/ungerik/go-mx/html"
    "github.com/ungerik/go-mx/shadcn"
)

func main() {
    page := html.NewDocument("Sign in",
        shadcn.Button(shadcn.ButtonDefault, shadcn.SizeDefault, "Sign in"),
    )
    http.HandleFunc("/", page.HandleHTTP)
    log.Println("serving on http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

```bash
go run .
```

Open <http://localhost:8080>. You'll see a button — but an **unstyled** one. The
component emitted all the right Tailwind classes; nothing is interpreting them
yet. That's the next step.

## Step 4: Add Tailwind v4 so it's styled

shadcn components emit Tailwind v4 utility classes and rely on shadcn's CSS
variables (`--background`, `--primary`, …). Something has to turn those classes
into CSS. The no-build way (great for development and demos) is the
`@tailwindcss/browser` CDN build, which compiles classes in the page at runtime.

First, grab the shadcn theme tokens. Copy
[`cmd/shadcn-gallery/theme.css`](https://github.com/ungerik/go-mx/blob/main/cmd/shadcn-gallery/theme.css)
from the go-mx repo into your project as `theme.css` — it's the shadcn
new-york-v4 `globals.css` (the `--background`, `--primary`, … variable
definitions the components reference).

Then embed it and wire the `<head>`:

```go
package main

import (
    "log"
    "net/http"

    _ "embed"

    "github.com/ungerik/go-mx"
    "github.com/ungerik/go-mx/html"
    "github.com/ungerik/go-mx/shadcn"
)

//go:embed theme.css
var themeCSS string

// head loads Tailwind v4 from the CDN and injects the shadcn theme tokens.
func head() mx.Component {
    return mx.Components{
        html.Element("style", html.Type("text/tailwindcss"), html.Raw(themeCSS)),
        html.Script(html.Src("https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4")),
    }
}

func main() {
    page := html.NewDocument("Sign in",
        html.Div(html.Class("flex min-h-screen items-center justify-center bg-background"),
            shadcn.Button(shadcn.ButtonDefault, shadcn.SizeDefault, "Sign in"),
        ),
    )
    page.HeadCustom = head()

    http.HandleFunc("/", page.HandleHTTP)
    log.Println("serving on http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

`html.Raw` inserts the CSS verbatim (no escaping); `page.HeadCustom` is rendered
into the document `<head>` after the title and meta tags.

```bash
go run .
```

Reload <http://localhost:8080>. The button is now a proper shadcn button,
centered on the page. There's a brief flash before Tailwind compiles — that's the
CDN trade-off, fine for development. (For production you'd run the Tailwind v4
CLI over your Go source instead; see the [how-to guides](how-to.html).)

## Step 5: Build the sign-in card

A lone button isn't a sign-in screen. Compose a `Card` with `Label`s, `Input`s,
and the button. Replace the body passed to `NewDocument`:

```go
page := html.NewDocument("Sign in",
    html.Div(html.Class("flex min-h-screen items-center justify-center bg-background p-4"),
        shadcn.Card(html.Class("w-full max-w-sm"),
            shadcn.CardHeader(
                shadcn.CardTitle("Sign in to your account"),
                shadcn.CardDescription("Enter your email below to sign in."),
            ),
            shadcn.CardContent(html.Class("grid gap-4"),
                html.Div(html.Class("grid gap-2"),
                    shadcn.Label(html.For("email"), "Email"),
                    shadcn.Input(html.Type("email"), html.ID("email"),
                        html.Placeholder("m@example.com")),
                ),
                html.Div(html.Class("grid gap-2"),
                    shadcn.Label(html.For("password"), "Password"),
                    shadcn.Input(html.Type("password"), html.ID("password")),
                ),
            ),
            shadcn.CardFooter(
                shadcn.Button(shadcn.ButtonDefault, shadcn.SizeDefault,
                    html.Class("w-full"), "Sign in"),
            ),
        ),
    ),
)
page.HeadCustom = head()
```

```bash
go run .
```

Reload. You have a centered sign-in card: a titled header, two labeled inputs,
and a full-width button — all styled, all server-rendered, zero client
JavaScript.

Notice the two ways you passed classes:
- The components carry their own base classes (`Card` already looks like a card).
- You added layout classes — `w-full max-w-sm`, `grid gap-2` — by passing
  `html.Class("…")` alongside the children. go-mx merges them into one `class`
  attribute, with your classes winning any Tailwind conflict.

## What you built

A complete, styled, server-rendered sign-in page in one Go file plus a theme,
with no templates and no front-end build. You saw a component render to raw HTML,
serve over HTTP, pick up Tailwind styling, and compose into a larger layout.

From here:

- **[How-to guides](how-to.html)** — production Tailwind setup, overriding
  classes, htmx, confirm dialogs, static export.
- **[Component gallery](gallery/)** — every component live next to its source,
  to copy from.
- **[Component reference](https://github.com/ungerik/go-mx/blob/main/shadcn/README.md)**
  — every signature and variant.
