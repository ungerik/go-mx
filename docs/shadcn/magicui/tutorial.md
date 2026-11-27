---
title: Build an animated landing hero with Magic UI
---

# Build an animated landing hero with Magic UI

You'll build the hero section of a landing page in Go: a shiny announcement
pill, a headline that blurs in word by word, a rainbow call-to-action button, a
row of words scrolling past, and a meteor shower behind it all. Every animation
is CSS, and the page itself loads no JavaScript of its own. Along the way you'll
see what a `magicui` component renders, why its animations need `magicui.CSS`,
and what a visitor who has turned motion off sees instead.

## What you'll need

- Go 1.27 or newer (`go version`)
- Chrome, Edge or another Chromium browser (Firefox and Safari work too, but
  Step 5 uses Chrome's DevTools)
- An internet connection (this tutorial loads Tailwind from a CDN)

No Node, no bundler, no build step. If you have done the
[shadcn tutorial](../tutorial.html), Steps 1 and 3 will look familiar.

## Step 1: Create the module

```bash
mkdir hero && cd hero
go mod init example.com/hero
go get github.com/ungerik/go-mx
```

`shadcn/magicui` is a package of the go-mx module, so this one `go get` brings
it in.

## Step 2: Render a Marquee to your terminal

Create `main.go`:

```go
package main

import (
    "context"
    "os"

    "github.com/ungerik/go-mx"
    "github.com/ungerik/go-mx/html"
    "github.com/ungerik/go-mx/shadcn/magicui"
)

func main() {
    marquee := magicui.Marquee(magicui.MarqueeOptions{Repeat: 2},
        html.Span("Go"), html.Span("HTML"), html.Span("CSS"),
    )
    marquee.Render(context.Background(), mx.NewCheckedWriter(os.Stdout))
}
```

Run it:

```bash
go run .
```

You'll see the markup the component produced (wrapped here for reading):

```html
<div data-slot="marquee" class="group flex gap-(--gap) overflow-hidden p-2 [--duration:40s] [--gap:1rem] flex-row">
  <div class="flex shrink-0 justify-around gap-(--gap) animate-marquee flex-row"><span>Go</span><span>HTML</span><span>CSS</span></div>
  <div class="flex shrink-0 justify-around gap-(--gap) animate-marquee flex-row" aria-hidden="true"><span>Go</span><span>HTML</span><span>CSS</span></div>
</div>
```

Three things to notice:

- Every Magic UI component is a Go function that takes an options struct and
  then attributes and children, like `shadcn` components. The zero value of an
  option is upstream's default, so `MarqueeOptions{}` is a sensible marquee.
- `Repeat: 2` rendered your children twice, and the second copy is
  `aria-hidden`, so a screen reader reads the words once.
- The movement is only a class name: `animate-marquee`. Nothing in this markup
  moves until a stylesheet defines what `animate-marquee` means. That's the
  next step.

## Step 3: Serve it and watch it move

The components emit Tailwind v4 classes on shadcn's theme tokens. Two pieces of
CSS turn them into a moving page:

- **The shadcn theme.** Copy
  [`cmd/shadcn-gallery/theme.css`](https://github.com/ungerik/go-mx/blob/main/cmd/shadcn-gallery/theme.css)
  from the go-mx repo into your project as `theme.css`. It is shadcn's
  `globals.css`: the `--background`, `--primary`, … tokens.
- **`magicui.CSS`.** A Go string constant in the package holding every
  animation: the `@keyframes`, the `animate-*` utilities and the
  reduced-motion rules.

Replace `main.go` with:

```go
package main

import (
    "log"
    "net/http"

    _ "embed"

    "github.com/ungerik/go-mx"
    "github.com/ungerik/go-mx/html"
    "github.com/ungerik/go-mx/shadcn/magicui"
)

//go:embed theme.css
var themeCSS string

// head loads Tailwind v4 from the CDN and compiles the shadcn theme tokens
// plus the Magic UI animations in the browser.
func head() mx.Component {
    return mx.Components{
        html.Element("style", html.Type("text/tailwindcss"),
            html.Raw(themeCSS+"\n"+magicui.CSS)),
        html.ScriptSrc("https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4"),
    }
}

func main() {
    page := html.NewDocument("Hero",
        html.Div(html.Class("flex min-h-screen items-center bg-background"),
            magicui.Marquee(magicui.MarqueeOptions{},
                html.Class("text-4xl font-bold"),
                html.Span("Go"), html.Span("HTML"), html.Span("CSS"),
            ),
        ),
    )
    page.HeadCustom = head()

    http.HandleFunc("/", page.HandleHTTP)
    log.Println("serving on http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

```bash
go run .
```

Open <http://localhost:8080>. The three words scroll from right to left in an
endless loop. The `<style type="text/tailwindcss">` block is compiled in the
browser by the `@tailwindcss/browser` script, and `magicui.CSS` goes after the
theme, because the theme begins with `@import "tailwindcss"` and the
animations build on it.

Try leaving `magicui.CSS` out of `head` and reload: the words are still there
and styled, but they stand still. If a Magic UI component ever renders but
doesn't move, this is the first thing to check.

## Step 4: Build the hero

Now compose the full hero. Replace `main` and add a `hero` function:

```go
func main() {
    page := html.NewDocument("Hero", hero())
    page.HeadCustom = head()

    http.HandleFunc("/", page.HandleHTTP)
    log.Println("serving on http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func hero() mx.Component {
    return html.Section(
        html.Class("relative flex min-h-screen flex-col items-center justify-center gap-8 overflow-hidden bg-background p-8"),
        magicui.Meteors(magicui.MeteorsOptions{Seed: 1}),
        html.Div(html.Class("rounded-full border bg-muted"),
            magicui.AnimatedShinyText(magicui.AnimatedShinyTextOptions{},
                html.Class("px-4 py-1 text-sm"),
                "✨ Rendered on the server",
            ),
        ),
        magicui.TextAnimate(magicui.TextAnimateOptions{
            Animation: magicui.TextAnimateBlurInUp,
            By:        magicui.TextAnimateByWord,
            As:        magicui.TextAnimateH1,
        }, html.Class("text-5xl font-bold tracking-tight"), "Pages that move, written in Go"),
        magicui.RainbowButton(magicui.RainbowButtonOptions{}, "Get started"),
        magicui.Marquee(magicui.MarqueeOptions{PauseOnHover: true},
            html.Class("w-full max-w-2xl text-muted-foreground [--duration:20s]"),
            html.Span("Go"), html.Span("HTML"), html.Span("CSS"),
            html.Span("Tailwind"), html.Span("htmx"),
        ),
    )
}
```

```bash
go run .
```

Reload. Meteors streak across the background, the pill's text is swept by a
glare, the headline blurs in one word after another, the button's border
cycles through the rainbow, and the bottom row scrolls past and stops while
your pointer is on it.

What each piece shows you:

- **`Meteors` draws its positions from `Seed`.** Upstream calls
  `Math.random()` in the browser; here the meteors are placed on the server,
  so the same seed renders the same markup on every request. The meteors are
  absolutely positioned, which is why the section is `relative` and
  `overflow-hidden`.
- **Your classes win.** `[--duration:20s]` on the Marquee overrides its base
  `[--duration:40s]`, because caller classes are merged over the component's
  own with `shadcn.Cn`. That's how you tune speeds and gaps.
- **`TextAnimate` splits the text on the server** into one `<span>` per word,
  each delayed through a CSS variable, plus a screen-reader-only copy of the
  whole sentence. Its children have to be plain text.

## Step 5: See what reduced motion shows

Some visitors turn animations off in their operating system. Magic UI's
upstream ignores that setting; this port honors it.

1. In Chrome, open DevTools, press <kbd>Cmd</kbd>+<kbd>Shift</kbd>+<kbd>P</kbd>
   (<kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>P</kbd> on Windows and Linux) and run
   **Show Rendering**.
2. Set **Emulate CSS media feature prefers-reduced-motion** to `reduce`.

The page stops moving, and nothing is lost: the headline is fully there, the
marquee stands still with its words visible, the button keeps its rainbow
border, and the meteors stop where they start. Every component's resting style is its
final state, so switching the animation off leaves the finished picture. Set
the option back to "No emulation" when you're done.

## What you built

A complete animated hero section in one Go file plus the shadcn theme: five
Magic UI components composed like any other go-mx markup, animated by CSS,
deterministic on the server, and still readable with motion turned off.

From here:

- **[How-to guides](how-to.html)**: the production Tailwind build (the CDN is
  for development), tuning speeds, scroll-triggered entrances, stable markup
  for caching and tests, and what to do when a component renders an error.
- **[Why CSS alone](explanation.html)**: how the port replaces the motion
  library and `Math.random()`, and what that costs.
- **[Component gallery](../gallery/)**: all 41 components next to their Go
  source, to copy from.
- **[Component reference](https://github.com/ungerik/go-mx/blob/main/shadcn/magicui/README.md#components)**:
  every component's signature and how it differs from upstream.
