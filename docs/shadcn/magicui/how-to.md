---
title: Magic UI how-to guides
---

# Magic UI how-to guides

Task-oriented recipes for the `shadcn/magicui` package. They assume you have
rendered a Magic UI component before; if not, start with the
[tutorial](tutorial.html). Every component's options are in the
[component reference](https://github.com/ungerik/go-mx/blob/main/shadcn/magicui/README.md#components).

- [Build the Magic UI CSS for production](#build-the-magic-ui-css-for-production)
- [Change the speed or spacing of an animation](#change-the-speed-or-spacing-of-an-animation)
- [Use the components under a strict Content-Security-Policy](#use-the-components-under-a-strict-content-security-policy)
- [Play an entrance when it scrolls into view](#play-an-entrance-when-it-scrolls-into-view)
- [Give a link the rainbow button look](#give-a-link-the-rainbow-button-look)
- [Render the same markup on every request](#render-the-same-markup-on-every-request)
- [Check what reduced motion shows](#check-what-reduced-motion-shows)
- [Fix a component that renders an error](#fix-a-component-that-renders-an-error)

---

## Build the Magic UI CSS for production

The tutorial compiles Tailwind in the browser from a CDN. For production,
compile one static stylesheet with the Tailwind v4 CLI. Magic UI adds two
things to the
[shadcn production setup](../how-to.html#set-up-tailwind-v4-for-production):
the animations in `magicui.CSS`, and an `@source` line, because Tailwind does
not look in the Go module cache where the components' class names live.

### Prerequisites

- Node.js, only to run the Tailwind CLI; the app stays pure Go
- `theme.css` with the shadcn tokens in `assets/` (see the
  [tutorial](tutorial.html#step-3-serve-it-and-watch-it-move))

### Steps

1. Install Tailwind and its CLI in the project:

   ```bash
   npm install tailwindcss@4 @tailwindcss/cli@4
   ```

2. Create `gencss.go` in the project root. It writes the two Tailwind inputs
   that come from go-mx: `magicui.CSS`, and an `@source` line pointing at the
   go-mx source wherever `go` keeps it.

   ```go
   //go:build ignore

   // gencss writes the Tailwind inputs that come from go-mx: the Magic UI
   // animations, and an @source line that makes Tailwind scan the go-mx
   // source in the Go module cache for the classes its components emit.
   package main

   import (
       "log"
       "os"
       "os/exec"
       "path/filepath"
       "strings"

       "github.com/ungerik/go-mx/shadcn/magicui"
   )

   func main() {
       err := os.WriteFile("assets/magicui.css", []byte(magicui.CSS), 0o644)
       if err != nil {
           log.Fatal(err)
       }
       dir, err := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/ungerik/go-mx").Output()
       if err != nil {
           log.Fatal(err)
       }
       shadcnDir := filepath.ToSlash(filepath.Join(strings.TrimSpace(string(dir)), "shadcn"))
       err = os.WriteFile("assets/go-mx-source.css", []byte(`@source "`+shadcnDir+`";`+"\n"), 0o644)
       if err != nil {
           log.Fatal(err)
       }
   }
   ```

   The `//go:build ignore` line keeps it out of your app's build; you run it
   with `go run`. Pointing `@source` at `shadcn` covers both `shadcn` and
   `shadcn/magicui` (`HeroVideoDialog` uses shadcn's `Dialog` classes).

3. Create `assets/input.css`:

   ```css
   @import "./theme.css";        /* begins with @import "tailwindcss" */
   @import "./magicui.css";
   @import "./go-mx-source.css";
   ```

4. Generate the inputs and compile:

   ```bash
   go run gencss.go
   npx @tailwindcss/cli -i assets/input.css -o public/app.css --minify
   ```

   Run both again after upgrading go-mx: the version is part of the module
   cache path, and `magicui.CSS` may have changed.

5. Link the stylesheet instead of the CDN script, and serve it:

   ```go
   page := html.NewDocument("Hero", hero())
   page.Stylesheets = []string{"/app.css"}

   http.Handle("/app.css", http.FileServer(http.Dir("public")))
   http.HandleFunc("/", page.HandleHTTP)
   ```

### Verification

```bash
grep -c '@keyframes marquee' public/app.css
```

prints `1` or more. Load a page with a `Marquee` on it: it scrolls, and the
page loads no script.

### Troubleshooting

- **Components are styled but don't move**, and `@keyframes marquee` is
  missing from `app.css`: Tailwind did not see the go-mx source. Tailwind only
  emits a theme animation such as `animate-marquee` when it finds the class in
  a scanned file, and that class is written in `shadcn/magicui/marquee.go`,
  not in your code. Check that `assets/go-mx-source.css` exists and names a
  directory that exists, and that `input.css` imports it.
- **`Can't resolve 'tailwindcss'`**: the CLI runs, but the `tailwindcss`
  package is not installed next to it. Run step 1 in the project root.
- **`app.css` is bigger than you expected**: scanning the whole `shadcn`
  directory emits the classes of every component, not only those you use. If
  you use only `magicui` components other than `HeroVideoDialog`, point
  `@source` at `shadcn/magicui` instead.

---

## Change the speed or spacing of an animation

Speeds and gaps come in two kinds, and the component's entry in the
[reference](https://github.com/ungerik/go-mx/blob/main/shadcn/magicui/README.md#components)
says which one it uses.

**An option field.** Many components take durations as `time.Duration` and
colors and lengths as CSS strings:

```go
magicui.ShimmerButton(magicui.ShimmerButtonOptions{
    ShimmerDuration: 2 * time.Second,
    ShimmerColor:    "oklch(80% 0.15 250)",
    BorderRadius:    "12px",
}, "Join the beta")
```

**A CSS variable with a class.** Where upstream sets a variable with a class,
so does the port, and your class wins over the component's own, merged with
`shadcn.Cn`. A Marquee runs one loop in 40 seconds with a 1rem gap; to run it
twice as fast with a wider gap:

```go
magicui.Marquee(magicui.MarqueeOptions{},
    html.Class("[--duration:20s] [--gap:2rem]"),
    html.Span("Go"), html.Span("HTML"), html.Span("CSS"),
)
```

A `style` attribute also wins over the component's own declarations: the
component writes its variables first and your `style` after them. A
Content-Security-Policy without `'unsafe-inline'` in `style-src` drops every
`style` attribute, options included; under one, use classes (see the next
guide).

```go
magicui.ShimmerButton(magicui.ShimmerButtonOptions{}, html.Style("--speed:1s"), "Join the beta")
```

### Verification

Render the component and look at its root element: your class appears after
the component's (`… flex-row [--duration:20s] [--gap:2rem]`) and replaces its
`[--duration:40s] [--gap:1rem]`; your `style` declarations come last.

---

## Use the components under a strict Content-Security-Policy

A site whose `Content-Security-Policy` has `style-src 'self'` (no
`'unsafe-inline'`) loads CSS only from its stylesheets, and the browser
ignores every `style` attribute. A component left at its defaults writes
none, so it works as it is; the
[reference](https://github.com/ungerik/go-mx/blob/main/shadcn/magicui/README.md#strict-content-security-policy)
names the 15 components that need a `style` attribute anyway.

### Steps

1. Leave the options that take CSS values at their zero values: an option
   that is set goes into `style`.
2. Set the value with a class instead: the variable or property the option
   sets, as a Tailwind arbitrary property. It replaces the component's
   default class, merged with `shadcn.Cn`.

   ```go
   // Instead of ShimmerButtonOptions{ShimmerColor: "#3b82f6", ShimmerDuration: 1500 * time.Millisecond}:
   magicui.ShimmerButton(magicui.ShimmerButtonOptions{},
       html.Class("[--shimmer-color:#3b82f6] [--speed:1.5s]"),
       "Join the beta",
   )
   ```

3. Write the class as a literal in your Go source, so Tailwind finds it and
   generates its rule.
4. For a BorderBeam's border width, put the class on the BorderBeam's parent
   (`[--border-beam-width:2px]`): the BorderBeam's own classes go to the
   beam, not to the border around it.

### Verification

Render the component: its elements carry no `style` attribute, and the root's
class list has your `[--shimmer-color:#3b82f6]` in place of the default
`[--shimmer-color:#ffffff]`. In the browser, the console shows no
Content-Security-Policy violation for a style attribute.

---

## Play an entrance when it scrolls into view

By default an entrance animation plays once, when the page loads, so content
below the fold has finished animating before anyone scrolls to it. Two
components can tie their entrance to the scroll position instead, through
scroll-driven animations (`animation-timeline: view()`):

```go
// Fade each photo in while it scrolls into view.
magicui.BlurFade(magicui.BlurFadeOptions{InView: true},
    html.Img("/photo.jpg", html.Alt("Our office")),
)

// Animate a heading in word by word as it scrolls into view.
magicui.TextAnimate(magicui.TextAnimateOptions{
    Animation:   magicui.TextAnimateSlideUp,
    StartOnView: true,
}, "Built for teams")
```

The animation follows the scrollbar: it runs while the element enters the
viewport, so `Duration` and `Delay` don't apply. `TextAnimate` also runs
backwards when the element scrolls out again.

### Verification

Put the component below the fold and scroll slowly: it fades or slides in as
it enters the viewport, and stops halfway if you stop scrolling halfway.

### Troubleshooting

- **The content is there at once, with no animation**: the browser has no
  scroll-driven animations (Firefox, unless its flag is set). The rules sit in
  `@supports (animation-timeline: view())`, so those browsers show the final
  state. That fallback is deliberate.
- **Nothing animates in any browser**: the `@source` line is missing from the
  production build (see the first guide). `animate-blur-fade-in-view` is a
  `magicui.CSS` utility like the others.

---

## Give a link the rainbow button look

Upstream renders a link as a rainbow button with `asChild`. go-mx has no Slot
component, so the package exports the button's classes instead:

```go
html.A(html.HRef("/pricing"),
    html.Class(magicui.RainbowButtonClasses(magicui.RainbowButtonOutline, magicui.RainbowButtonSizeLG)),
    "See pricing",
)
```

An empty variant or size means the default, so
`RainbowButtonClasses("", "")` is the plain rainbow button. The colors are the
CSS variables `--rainbow-color-1` to `--rainbow-color-5` (upstream's
`--color-1` … `--color-5`, renamed out of Tailwind's color namespace), defined
by `magicui.CSS` on `:root` in the base layer; set them unlayered or in a later
layer to change them.

### Verification

The link looks exactly like `magicui.RainbowButton(…)` and navigates on click.

---

## Render the same markup on every request

Same input, same HTML: that makes a page cacheable by an ETag and lets a
golden-file test compare it byte for byte. Two kinds of options decide it.

**`Seed`.** `Meteors`, `LightRays`, `SparklesText`, `WarpBackground`,
`DotPattern` (with `Glow`) and `Text3DFlip` (with a random stagger) draw
positions and timings that upstream takes from `Math.random()`. Here they come
from a generator seeded by `Seed` on the server, so the markup is already the
same on every request, even when `Seed` is 0. Set different seeds to make two
instances on one page look different:

```go
magicui.Meteors(magicui.MeteorsOptions{Seed: 1}), // left panel
magicui.Meteors(magicui.MeteorsOptions{Seed: 2}), // right panel
```

**`ID`.** The SVG components (`DotPattern`, `GridPattern`, `HexagonPattern`,
`StripedPattern`, `NoiseTexture`, `Backlight`, and the `Iphone`, `Android` and
`Safari` mockups) and `HeroVideoDialog` need an id that other markup refers
to. Left empty it is an `mx.UniqueID`, drawn from a process-wide counter, which
differs on every render. Pass a fixed id that is unique in the page:

```go
magicui.Iphone(magicui.IphoneOptions{Src: "/screenshot.png", ID: "hero-phone"})
```

### Verification

Render the page twice and compare the output: with fixed ids, the two renders
are identical. An `Iphone` without an `ID` differs between two renders; with
`ID: "hero-phone"` it doesn't.

### Troubleshooting

- **`magicui: invalid id "…": only letters, digits, '_' and '-' are
  allowed`**: the id ends up inside `url(#…)` in CSS, where other characters
  could end the reference. Use a plain slug.

---

## Check what reduced motion shows

Under `prefers-reduced-motion: reduce` the package switches off every
animation it defines, and each component shows its final state: text fully
visible, a marquee standing still, a progress bar at its value, sparkles
hidden. Hover transitions stay.

### Steps

1. Open the page in Chrome and open DevTools.
2. Press <kbd>Cmd</kbd>+<kbd>Shift</kbd>+<kbd>P</kbd>
   (<kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>P</kbd> on Windows and Linux) and
   run **Show Rendering**.
3. Set **Emulate CSS media feature prefers-reduced-motion** to `reduce`.

On macOS you can test the real setting instead: System Settings, Accessibility,
Display, **Reduce motion**.

### Verification

Nothing moves, and nothing is missing that the animation would have revealed.
Text that animates in by words or letters is complete, and your layout still
reads as designed.

### Troubleshooting

- **Your own animation keeps playing**: the package only switches off the
  animations in `magicui.CSS`. An `animate-*` utility you pass in yourself,
  for example through `TextAnimateOptions.SegmentClass`, needs its own
  `@media (prefers-reduced-motion: reduce)` rule.

---

## Fix a component that renders an error

Magic UI constructors never panic. A bad option becomes an error that
`Render` returns, and `html.Document.HandleHTTP` answers with
`500 Internal Server Error` instead of sending a half-rendered page. The
error says what to fix:

| Error | Cause | Fix |
|---|---|---|
| `magicui: invalid CSS value "…" for --shimmer-color` | A color, length or other CSS option holds `;`, a brace, an angle bracket, a backslash, a quote, a comment delimiter, `url(` or an unbalanced parenthesis or bracket, any of which could break out of the `style` attribute | Pass a plain CSS value such as `#3b82f6`, `oklch(70% 0.2 250)` or `12px`. Check user-supplied values before passing them in. |
| `magicui: invalid id "…": only letters, digits, '_' and '-' are allowed` | An `ID` option holds another character | Use a slug like `hero-phone` |
| `magicui: Marquee Repeat -1 is negative` | A count option is negative | Use 0 for the default, or a positive count |
| `magicui: … children must be text, got …` | `TextAnimate`, `TextReveal`, `SpinningText`, `Text3DFlip` or `LineShadowText` got a non-text child | Pass strings; put attributes such as `html.Class` before them |

### Verification

Render the component on its own and check the error:

```go
err := magicui.ShimmerButton(magicui.ShimmerButtonOptions{
    ShimmerColor: "oklch(70% 0.2 250)",
}, "Buy").Render(context.Background(), mx.NewCheckedWriter(io.Discard))
// err == nil
```

A value that would inject CSS, such as `"red;background:url(https://x)"`,
returns `magicui: invalid CSS value "red;background:url(https://x)" for
--shimmer-color`.
