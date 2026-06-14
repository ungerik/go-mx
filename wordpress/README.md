# wordpress

[![Go Reference](https://pkg.go.dev/badge/github.com/ungerik/go-mx/wordpress.svg)](https://pkg.go.dev/github.com/ungerik/go-mx/wordpress)

Import the content of a WordPress site from a **WXR export** (the XML produced by
WordPress under *Tools → Export*) and re-render it with go-mx + shadcn as a clean
static site — same structure as the original (posts, pages, archives, navigation,
comments), not its pixel-perfect theme.

It is its own nested module so its dependencies (`x/net/html` for parsing) never
touch the core go-mx module.

```bash
# install the CLI
go install github.com/ungerik/go-mx/wordpress/cmd/wordpress-import@latest

# convert an export to a static site, then preview it
wordpress-import -in export.xml -out ./dist -serve :8080
```

## What you get

- A typed, JSON-serializable Go model of the site (`*Site` → posts, pages, terms,
  authors, comments, attachments, menus). This is the durable, reusable asset.
- A static shadcn site: single posts, pages, home/category/tag/author archives,
  a 404, a site shell with the primary menu, and threaded comments.
- An **import report** (stdout summary + `import-report.json` + `import-report.html`)
  listing everything that didn't translate cleanly — unknown shortcodes, plugin
  blocks, removed markup, blocked URLs — each tied to the source post. This is how
  you find what needs manual attention after a migration.
- Every logical component as a composable `mx.Component`, so you can embed a post,
  an archive, or a menu in your own go-mx page.

## Library: import in three lines

```go
site, report, err := wordpress.ParseFile("export.xml")
if err != nil { log.Fatal(err) }
_, err = wordpress.WriteStatic(site, "./dist", wordpress.Options{})
fmt.Print(report.Summary())
```

`Parse(io.Reader)`, `ParseFile(path)` and `ParseFiles(paths...)` (for a split
multi-file export) all return `(*Site, *Report, error)`.

## Embedding a single component

Every view is a plain `mx.Component`. Build the render context once, then drop a
piece into your own page — install the theme with `wordpress.HeadComponents()` so
it’s styled:

```go
v := site.Views(wordpress.Options{})
doc := html.NewDocument("My blog", v.PostView(site.Posts[0]))
doc.HeadCustom = wordpress.HeadComponents()
```

## Options (each mirrored by a CLI flag)

| Option        | CLI flag       | Default   | Meaning                              |
|---------------|----------------|-----------|--------------------------------------|
| `Permalinks`  | `-permalinks`  | `slug`    | `slug` (`/x/`), `dated` (`/2024/05/x/`), `id` (`/p/12/`) |
| `Statuses`    | `-status`      | `publish` | comma-separated statuses to include  |
| `BasePath`    | `-base`        | `""`      | URL sub-path, e.g. `/blog`           |
| `SiteTitle`   | `-title`       | from WXR  | overrides the rendered site title    |

The CLI also takes `-in <file>` (or positional files), `-out <dir>`, and
`-serve <addr>` (serve `-out` after writing).

## Security

Body HTML from the export is parsed with `x/net/html` and re-emitted through the
go-mx `html` constructors under an allowlist — `<script>`/`<style>`/`<iframe>`,
event handlers, inline styles and dangerous URL schemes are removed, and body
`<h1>`s are demoted so the page keeps one heading. Raw HTML is never passed
through. Output slugs are sanitized and every write is containment-checked, so a
hostile slug cannot escape the output directory.

## Caveats (v1)

- The site loads **Tailwind v4 from a CDN**, so viewing it needs internet. The
  article-body typography (`.wp-content`) is plain CSS and works offline.
- Links are **root-absolute** — serve the output from a web root (`python3 -m
  http.server`), don’t open the files directly.
- **Light theme only** (raw WP HTML carries inline colors that break on dark).
- **Page builders** (Elementor, Divi, WPBakery) are detected and flagged in the
  report, never faked. **Gutenberg** blocks render as their inner HTML.
- Non-ASCII slugs fall back to a `post-<id>` path (a deliberate safety choice).
- The live **MySQL database** and a dynamic HTTP handler are out of scope for v1;
  a REST adapter is the documented future seam.

## Planned (not yet implemented)

These are designed-for but not built in v1. The serializable model and the
`Source`/`Views` seams are shaped so each is an additive change, not a rewrite.

**Sources & output**
- **Dynamic HTTP handler** — serve the model per request instead of writing
  static files. v1 is static export only; the model makes this additive.
- **REST API source** — ingest from `wp-json/wp/v2` (public read) as an
  alternative to a WXR file. WXR is the only source today.
- **Markdown (+ frontmatter) output** — for migrating off WordPress into another
  SSG. v1 renders shadcn HTML only.
- **RSS / Atom feed** generation.

**Content pipeline**
- **Gutenberg block → shadcn mapping** — v1 strips block comments and renders the
  inner HTML; a block dispatcher (heading, image, gallery, columns, …) is future.
- **More core shortcodes** — v1 strips unknown shortcode delimiters and keeps the
  inner content; only a few core media shortcodes are recognized.
- **Optional `bluemonday` sanitize path** — for untrusted / multi-author exports.
  v1 defaults to trusted input (the structural allowlist always runs, but the
  `bluemonday` integration and a `TrustedRawHTML` escape hatch are not wired yet).
- **Input size / count limits** (`MaxInputBytes`, `MaxItems`) to bound memory on
  untrusted or pathological exports.
- **PHP-`unserialize`** for the serialized postmeta/options values that need it.

**Assets & rendering**
- **Media download + local URL rewriting** — v1 keeps absolute source-site media
  URLs; downloading and rewriting them for a self-contained export is future.
- **Dark mode / theme toggle** — v1 is light-only (see Caveats).
- **Offline asset bundling** — vendor a precompiled Tailwind stylesheet so the
  output renders without the CDN.

**Hardening (Tier-B)**
- Symlink-safe writes (v1's containment check is lexical; it does not yet reject
  symlinks in a pre-existing output tree).
- Parse-time skipped-item diagnostics in the written report files (today they
  appear in the CLI stdout summary, not in `import-report.json`/`.html`).
