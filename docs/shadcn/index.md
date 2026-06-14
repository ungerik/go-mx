---
title: shadcn for go-mx
---

# The `shadcn` package

A Go port of [shadcn/ui](https://ui.shadcn.com), built on go-mx's `html`
primitives. shadcn/ui ships React components; go-mx renders HTML on the server
with no client runtime, so this is a *port*, not a wrapper: the markup and
Tailwind classes are reproduced in Go, and behavior React delegates to Radix is
re-expressed with web-platform primitives (native `<dialog>`, the Popover API,
CSS Anchor Positioning, native form controls).

See every component rendered live next to its Go source in the
**[component gallery](gallery/)**.

```go
import "github.com/ungerik/go-mx/shadcn"

shadcn.Button(shadcn.ButtonDefault, shadcn.SizeDefault, "Sign in")
```

This documentation follows the [Diátaxis](https://diataxis.fr) framework.

## Tutorial

New to the package? Start here.

- **[Build your first shadcn page](tutorial.html)** — from an empty module to a
  styled sign-in card served over HTTP, including the Tailwind v4 setup the
  components require.

## How-to guides

Task-oriented recipes for things you'll actually need.

- **[shadcn how-to guides](how-to.html)** — set up Tailwind v4, override a
  component's classes, wire a component to htmx, build a confirm dialog, give a
  non-button the button look, and export a page to static HTML.

## Reference

The complete component reference lives in the package README, kept next to the
code so it can't drift:

- **[shadcn package reference](https://github.com/ungerik/go-mx/blob/main/shadcn/README.md)**
  — every component's signature, variants, and the React-to-Go mapping, tier by
  tier (structural, interactive, floating).
- **[`Cn`, `clsx`, `twmerge`, `cva`](https://github.com/ungerik/go-mx/blob/main/shadcn/README.md#cn--the-class-merging-helper)**
  — the class-handling helpers the components build on.
- **[Full API on pkg.go.dev](https://pkg.go.dev/github.com/ungerik/go-mx/shadcn)**

## Explanation

Why the port works the way it does — the design rationale for replacing Radix
with web-platform primitives:

- **[AlertDialog: native `<dialog>` instead of Radix](https://github.com/ungerik/go-mx/blob/main/shadcn/README.md#alertdialog-native-dialog-instead-of-radix)**
- **[Tier 3 floating components: Popover API + CSS Anchor Positioning](https://github.com/ungerik/go-mx/blob/main/shadcn/README.md#tier-3-components)**
- **[CSS pseudo-element indicators on void inputs](https://github.com/ungerik/go-mx/blob/main/shadcn/README.md#css-pseudo-element-indicators-on-void-inputs)**
- **[Why server-rendered HTML at all](../why-go-mx.html)**
