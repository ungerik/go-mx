---
title: Magic UI for go-mx
---

# The `shadcn/magicui` package

A Go port of [Magic UI](https://magicui.design), the animated components and
effects built on shadcn/ui's design tokens: marquees, text that blurs in word
by word, shimmering and rainbow buttons, meteor showers, dot and grid
backgrounds, device mockups and more. Upstream ships React components that
animate through the motion library; this port renders HTML on the server and
animates it with CSS alone, so a Magic UI page loads no script of its own.

See all 41 components rendered live next to their Go source in the
**[component gallery](../gallery/)** (the Magic UI group).

```go
import "github.com/ungerik/go-mx/shadcn/magicui"

magicui.Marquee(magicui.MarqueeOptions{PauseOnHover: true},
    html.Span("Go"), html.Span("HTML"), html.Span("CSS"),
)
```

The animations live in one stylesheet, `magicui.CSS`, which the app adds to
its Tailwind v4 input. Without it the components render, but stand still.

This documentation follows the [Diátaxis](https://diataxis.fr) framework.

## Tutorial

- **[Build an animated landing hero](tutorial.html)**: from an empty module to
  a hero section with five animated components served over HTTP, including the
  CSS setup and a look at what reduced motion shows.

## How-to guides

- **[Magic UI how-to guides](how-to.html)**: build the CSS for production
  (with the Tailwind `@source` line the Go module cache needs), change speeds
  and gaps, play entrances on scroll, give a link the rainbow button look,
  render the same markup on every request, check reduced motion, and fix a
  component that renders an error.

## Reference

The component reference lives in the package README, next to the code:

- **[Magic UI package reference](https://github.com/ungerik/go-mx/blob/main/shadcn/magicui/README.md)**:
  setup, the React-to-Go component model, and every component's signature
  and options, grouped into text effects, buttons, borders, backgrounds,
  motion and layout, data and media, and device mockups.
- **[Differences from upstream](https://github.com/ungerik/go-mx/blob/main/shadcn/magicui/README.md#differences-from-upstream)**:
  reduced motion, motion rewrites, scroll-driven animations, seeded
  randomness, ids, and accessibility.
- **[Not yet ported](https://github.com/ungerik/go-mx/blob/main/shadcn/TODOS.md#magic-ui-magicui)**:
  the components that need a script or a third-party library.
- **[Full API on pkg.go.dev](https://pkg.go.dev/github.com/ungerik/go-mx/shadcn/magicui)**

## Explanation

- **[Why Magic UI animates with CSS alone](explanation.html)**: how the port
  replaces the motion library, `Math.random()` and `useId()`, and what it
  gives up (exit animations, moving sparkles, Firefox's scroll-driven
  animations).
- **[The `shadcn` package docs](../)**: the shadcn/ui port these components
  build on, with its own tutorial and Tailwind setup.
