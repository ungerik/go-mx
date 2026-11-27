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
`dist/components/<slug>/index.html` — and the fragments the examples load
after the page, `dist/fragments/<path>`: the calendar examples set
`CalendarOptions.MonthURL`, and every month they can navigate to is
pre-rendered there (`examples/fragments.go`), so month navigation works on a
static host. The export replaces the whole `fragments` directory. The pages link to each other with
relative URLs that name the `index.html` files (`components/button/index.html`,
`../../index.html`), so the directory works served from any URL path, such as a
GitHub Pages project page, and opened straight from disk. The server answers
the same `…/index.html` paths. From disk, browsers refuse the `fetch` of a
calendar's month fragments, so month navigation needs the directory served
(e.g. `python3 -m http.server` inside `./dist`).

```bash
go run ./cmd/shadcn-gallery -out ./dist -static-highlight -home ../../index.html
```

This is how the documentation site
([`tools/gen-docs-site`](../../tools/gen-docs-site/main.go)) writes the gallery
it publishes at <https://ungerik.github.io/go-mx/shadcn/gallery/>
whenever `main` changes; the output is not committed. `-home` adds a link from
the sidebar back to the site's home page, given relative to the gallery index.

An internet connection is required: Tailwind v4 is loaded from a CDN (see below).

## How it fits together

| File             | Responsibility                                                        |
|------------------|-----------------------------------------------------------------------|
| `main.go`        | The component catalog (`docs()`), the flags, and the HTTP routes.    |
| `magicui.go`     | The Magic UI part of the catalog (`magicUIDocs()`), its own group.   |
| `export.go`      | `-out` static-site generation: render every page to a file.          |
| `shell.go`       | The page shell: `<head>` wiring, sidebar + main layout grouped by     |
|                  | library, Preview/Code tab blocks. The `<head>` comes from             |
|                  | `internal/shadcntheme`, which the documentation site shares.         |
| `registry.go`    | Types (`Example`, `ComponentDoc`, `DocGroup`, `Registry`) and Go-     |
|                  | source extraction. Embeds `examples/*.go` and                         |
|                  | `examples/magicui/*.go`.                                              |
| `examples/`      | One function per labeled preview — the live component. Their bodies   |
|                  | are the snippets shown in the Code tab. The Magic UI examples are     |
|                  | the `examples/magicui` package.                                       |

## Tailwind v4 via the browser CDN build

The `shadcn` components emit only Tailwind v4 utility classes and rely on the
shadcn CSS variables (`--background`, `--primary`, …). Rather than add a node
toolchain to this pure-Go repo, the gallery loads
[`@tailwindcss/browser`](https://www.npmjs.com/package/@tailwindcss/browser):
the shadcn/ui new-york-v4 theme tokens
([`internal/shadcntheme/theme.css`](../../internal/shadcntheme/theme.css)) go in a
`<style type="text/tailwindcss">` block and the
CDN script compiles them plus the classes it finds in the live DOM at runtime.
The script is pinned to an exact version and loaded with a Subresource
Integrity hash (`shadcntheme.TailwindBrowserURL` and
`TailwindBrowserIntegrity`), so a new release cannot change the gallery
without a commit. Zero build step; the trade-off is a CDN dependency and a brief first-paint
compile, which is fine for a demo. A production app would run the Tailwind v4
CLI over the Go source instead.

## Showing the source next to the preview

Each preview is a `func() mx.Component` in `examples/`. The files are embedded
and parsed once at startup (`registry.go`); each example's function body is
extracted and shown verbatim in the Code tab. The live component and its
displayed source therefore come from the **same function** and cannot drift.
An example whose month fragments are rendered by a month-taking function
names that function as `Example.ShowSource`, and the Code tab shows it.

## Status

The gallery covers the whole shadcn/ui registry: 65 pages, one per
component, listed alphabetically like the shadcn/ui docs (see `shadcnDocs()`
in `main.go` for the catalog). A second section holds the 41 pages of the
Magic UI port, `shadcn/magicui` (`magicUIDocs()` in `magicui.go`, examples in
the `examples/magicui` package); `head()` adds `magicui.CSS` to the Tailwind
input so their animations compile. Date Picker and Data Table are compositions
rather than package functions, so their pages show the recipes, as upstream
ships them. Combobox is a recipe upstream too, but here it is the
`shadcn.Combobox` form field, as a recipe cannot submit a value.

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
