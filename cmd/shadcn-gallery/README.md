# shadcn-gallery

A go-mx rebuild of the [shadcn/ui component docs](https://ui.shadcn.com/docs/components):
a sidebar of every ported component, one page per component, and each example
shown as a **live preview next to the Go source** that produced it. It doubles
as a visual dogfood and an integration test for the `shadcn` package.

The same executable either **serves** the gallery or **writes it as static
files**:

```bash
go run ./cmd/shadcn-gallery               # serve on http://localhost:8080
go run ./cmd/shadcn-gallery -out ./dist   # write static HTML to ./dist, then exit
```

The static output is the same pages the server serves — `dist/index.html` plus
`dist/components/<slug>/index.html`. By default the pages link to each other with
root-absolute URLs (`/`, `/components/…`), so serve the directory from a web root
(e.g. `python3 -m http.server` inside `./dist`).

To host under a URL **sub-path** (such as a GitHub Pages project page), pass
`-base` so every in-gallery link is prefixed:

```bash
go run ./cmd/shadcn-gallery -out docs/shadcn/gallery -base /go-mx/shadcn/gallery -static-highlight
```

This is exactly how the gallery committed under [`docs/shadcn/gallery/`](../../docs/shadcn/gallery/)
(served at `https://ungerik.github.io/go-mx/shadcn/gallery/`) is generated.

An internet connection is required: Tailwind v4 is loaded from a CDN (see below).

## How it fits together

| File             | Responsibility                                                        |
|------------------|-----------------------------------------------------------------------|
| `main.go`        | The component catalog (`docs()`), the flags, and the HTTP routes.    |
| `export.go`      | `-out` static-site generation: render every page to a file.          |
| `shell.go`       | The page shell: `<head>` wiring, sidebar + main layout, Preview/Code  |
|                  | tab blocks. Embeds `theme.css`.                                       |
| `registry.go`    | Types (`Example`, `ComponentDoc`, `Registry`) and Go-source           |
|                  | extraction. Embeds `examples/*.go`.                                   |
| `examples/`      | One function per labeled preview — the live component. Their bodies   |
|                  | are the snippets shown in the Code tab.                               |
| `theme.css`      | shadcn/ui new-york-v4 `globals.css` theme tokens.                     |

## Tailwind v4 via the browser CDN build

The `shadcn` components emit only Tailwind v4 utility classes and rely on the
shadcn CSS variables (`--background`, `--primary`, …). Rather than add a node
toolchain to this pure-Go repo, the gallery loads
[`@tailwindcss/browser`](https://www.npmjs.com/package/@tailwindcss/browser):
the `theme.css` tokens go in a `<style type="text/tailwindcss">` block and the
CDN script compiles them plus the classes it finds in the live DOM at runtime.
Zero build step; the trade-off is a CDN dependency and a brief first-paint
compile, which is fine for a demo. A production app would run the Tailwind v4
CLI over the Go source instead.

## Showing the source next to the preview

Each preview is a `func() mx.Component` in `examples/`. The files are embedded
and parsed once at startup (`registry.go`); each example's function body is
extracted and shown verbatim in the Code tab. The live component and its
displayed source therefore come from the **same function** and cannot drift.

## Status

The gallery covers every ported `shadcn` component — 48 in all, one page
each, listed alphabetically like the shadcn/ui docs (see `docs()` in
`main.go` for the catalog).

The only shadcn/ui component without a gallery page is **Chart**, which is
deferred (a design decision is pending; see `shadcn/TODOS.md`).

## Syntax highlighting

The Code tab can be highlighted two ways, selected at startup:

```bash
go run ./cmd/shadcn-gallery                    # dynamic (default): Shiki in the browser
go run ./cmd/shadcn-gallery -static-highlight  # static: highlight package, server-side
```

**Dynamic (default).** The Code tab is highlighted by
[Shiki](https://shiki.style) loaded as a deferred ESM module from a CDN
(`head()`), matching the no-build-step, CDN-first approach of the Tailwind
setup. Shiki uses TextMate grammars (the same engine as VS Code), so call sites,
types and properties are colored — not just keywords and strings the way a regex
highlighter manages. For each `<pre><code class="language-go">` it produces its
own themed `<pre>` (Shiki's bundled monokai theme, with inline background and token colors), onto
which the layout classes are re-applied before swapping it in; this covers Code
tabs that start hidden. `dynamicCodeBlock`'s placeholder `<pre>` carries a dark
background so the block looks right for the moment before Shiki paints.

**Static (`-static-highlight`).** The Go source is highlighted server-side by the
repo's own [`highlight`](../../highlight) package: `staticCodeBlock` renders the
source to `<span class="hl-…">` markup and `head()` injects the package's dark
theme stylesheet. Every page then ships already-colored code with no client-side
JavaScript and no CDN dependency for the code blocks. The trade-off is a
lexer-based highlighter (Go `go/scanner`) rather than Shiki's full TextMate
grammar, so the coloring is a little coarser.
