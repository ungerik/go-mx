# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Test Commands

```bash
# Run all tests of the root module
go test ./...

# ./... stops at nested modules even in workspace mode; CI tests each one
go test -C mcp ./...

# Run tests for a specific package
go test ./html/...

# Run a single test
go test -run TestFunctionName ./...

# The tools module: gen-docs-site and gen-stdoptions; -short skips the full
# site build (gallery and example snapshots), which the Pages workflow runs
go test -C tools -short ./...

# After changing an htmx version or hash in hx/htmx.go: check the SRI hashes
# against the CDN (needs network, so it is skipped unless MX_CHECK_CDN is set)
MX_CHECK_CDN=1 go test ./hx -run TestScriptFromCDNIntegrity

# The same check for the pinned @tailwindcss/browser build of the demo pages,
# after changing its version or hash in internal/shadcntheme/shadcntheme.go
MX_CHECK_CDN=1 go test ./internal/shadcntheme -run TestTailwindBrowserIntegrity

# Build check
go build ./...

# Regenerate the go-enum methods of the root module; like go test, ../...
# stops at nested modules (CI runs go-enum -validate on each)
go -C tools tool go-enum ../...
```

## Repository documentation

- [`README.md`](README.md) — what go-mx is, the package overview, and the
  reflected-forms guide.
- [`CHANGELOG.md`](CHANGELOG.md) — [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)
  format. Nothing is tagged yet, so every entry is still under
  `## [Unreleased]`; record user-visible changes there.
- [`TODOS.md`](TODOS.md) — deferred work, grouped by area and prioritised
  P0–P4, each item as **What / Why / Context / Effort / Priority / Depends on**,
  with finished items moved to `## Completed`. Deliberate non-goals of a change
  belong here rather than being silently skipped. The shadcn/ui port build order
  has its own list in [`shadcn/TODOS.md`](shadcn/TODOS.md).
- [`V1-AUDIT.md`](V1-AUDIT.md) — the v1.0 readiness audit of 2026-11-16: what to
  change before the API is frozen, with evidence; its open decisions and work
  items are tracked in `TODOS.md` under `## v1.0 release`, `## hx` (the htmx
  major, the event constants) and `## doc` (the unfinished `doc` package).
- Per-package READMEs carry the detail: `html/`, `svg/`, `xml/`, `hx/`,
  `shadcn/`, `highlight/`, `logview/`, `pdf/`, `pdf/pdftable/`, `mcp/`,
  `stdoptions/` (plus `highlight/DESIGN.md` and `shadcn/RESEARCH.md`).
- Former nested modules now in their own repositories:
  [`go-mx-wordpress`](https://github.com/ungerik/go-mx-wordpress) (the
  WordPress WXR importer) and [`go-mx-fpdf`](https://github.com/ungerik/go-mx-fpdf)
  (the legacy fpdf wrapper, parity baseline of `pdf`).

## Documentation site

The docs site at <https://ungerik.github.io/go-mx/> is built by go-mx itself
and deployed from GitHub Actions (`.github/workflows/pages.yml`); nothing of
it is committed. `tools/gen-docs-site` (in the `tools` module, so goldmark
stays out of the root `go.mod`) renders every Markdown file of the repository
with goldmark into pages built from `html` and `shadcn` in the gallery's
theme (`internal/shadcntheme`), Go code blocks highlighted by `highlight`:
`docs/**` at the URLs Jekyll used to serve (`docs/html/tutorial.md` →
`/html/tutorial`), every other `*.md` below `/repo/` (`html/README.md` →
`/repo/html/`), except `TODOS.md`, `V1-AUDIT.md`, `AGENTS.md`, `CLAUDE.md`
and `testdata/`. Links between Markdown files become relative page links,
links to other repository files go to GitHub, and a link that resolves to
nothing fails the build. It also writes the component gallery (`go run
./cmd/shadcn-gallery -static-highlight -home …`) to `/shadcn/gallery/`, a
static snapshot of each example server (their `-snapshot DIR` mode, with
`-snapshot-index` linking back to the examples page,
`snapshot.go` next to `main.go`, helpers in `internal/snapshot`) to
`/examples/`, the reference PDFs of the pdf golden tests, and through `web`
`sitemap.xml`, `llms.txt` and a `.md` alternate of every page, then checks
every internal link and fragment of the output.

Build and look at the site locally (the output directory is emptied first,
which the command only does for a directory an earlier build wrote):

```bash
go -C tools run ./gen-docs-site -out ../_site

# Only the Markdown pages, skipping the slow gallery and example snapshot builds
go -C tools run ./gen-docs-site -out ../_site -gallery=false -examples=false
```

Pull requests run the same build in the `Pages` workflow; a push to `main`
deploys it. A new example command gets a `-snapshot` mode and an entry in
`examples` in `tools/gen-docs-site/examples.go`; a new Markdown file is picked
up on its own, and so is any other file below `docs/` (an image, a download),
published at the same path.

## Project Overview

go-mx is a Go library for generating HTML markup programmatically using a component-based architecture. It provides type-safe, composable HTML generation with automatic escaping and validation.

## Architecture

### Core Abstractions (mx package - root)

- **Component** (`component.go`): The fundamental interface - everything that can render HTML implements `Render(context.Context, Writer) error`
- **Element** (`element.go`): Represents an HTML element with name, attributes, and children. Void elements have `nil` Children, regular elements have empty slice for no children. An `Err` field carries a deferred construction error: if non-nil, `Render` returns it; `NewErrElement(err)` builds such an element.
- **Attrib** (`attrib.go`): Interface for HTML attributes; `AttribValue(context.Context) (string, error)` so a value can be computed from the context and fail. Implementations: `Attribute` (name/value pair), `ConstAttrib` (format "name=value", or a bare "name" for a boolean attribute), `BoolAttrib` (boolean HTML attributes), and `ErrAttrib` (defers an invalid-value error to render time). See `doc.go` for the deferred-error pattern. `AttribString` formats any `Attrib` as escaped `name='value'` for `String` methods and debugging (a failing value as `!ERROR: …`); `AttribStringContext` does the same with a render context and returns the value's error. Check an `Attrib` for nil with `IsNilAttrib`, which also catches a typed nil pointer or func; `a == nil` misses it.
- **Writer** (`writer.go`): Interface for HTML output with element lifecycle methods
- **CheckedWriter** (`checkedwriter.go`): Writer implementation with validation, escaping, and optional indentation (never inside `pre`, `textarea`, `script` or `style`, whose whitespace is content)
- **Error policy** (`errors.go`): where go-mx hides an error behind a generic text, it reports it with `ReportHiddenError` (to the application's `ReportError` hook; context errors skipped), and where it decides whether the user may read an error, it asks `UserMessage` (any `UserMessage() string` method in the chain, `FormMessage` included). New code that hides an error does both, and the substitute text it shows is a `FormMessage`, so a later stage neither hides nor reports it again. `FormErrorText` and the sensitive-field mask `MaskSensitiveError` apply both, a field renders its errors with `FieldErrorTexts`; `RespondNonContextError(w, r, err)` and `SSEResponse.SendError` show no user message and report every error they hide.
- **SSEResponse** (`sse.go`): a `text/event-stream` response that renders one `Component` per event over a still-open HTTP response. `NewSSEResponse` probes flushability before writing a header, so a non-flushable writer is still answerable with a 500; `Send`/`SendEvent` render into a buffer first, so a render error or panic is reported before a byte of the frame is written, and re-split the markup into one `data:` line per output line. After the first flush the 200 is committed, so a later failure goes out in band as the `SSEEventError` (`"mx-error"`) event via `SendError`. `SSEEvent.ID` plus `LastEventID(request)` make a browser's silent reconnect resumable, `SetRetry` tunes the reconnect delay, `KeepaliveLoop` defeats proxy idle timeouts, and `SetWriteTimeout` bounds a single frame (`DefaultSSEWriteTimeout`). The content type is `ContentTypeEventStream` (`contenttypes.go`). Worked example in `cmd/example-sse`, which streams a chat transcript, appends tokens out of band by `KeyedID`, and can drop the connection mid-reply (`-drop`) or fail in band (`-fail`).
- **UniqueID / KeyedID** (`uniqueid.go`): `UniqueID()` draws from a process-lifetime counter; `KeyedID(parts...)` derives the id deterministically from key parts, so the same entity gets the same id in the initial render and in every later event — what an out-of-band or SSE swap needs to find its target. `KeyedIDValue` returns the string for a selector. `ValidIDRune` is the shared definition of which characters an id may hold (`shadcn.validateID` uses it too).

### Packages

- **html/**: HTML5 elements and attributes - `Div()`, `Span()`, `Class()`, `ID()`, etc. Includes typed keyword enum attributes that validate at render time (`enums.go`, generated by go-enum like `svg`) and numeric attribute constructors. The `html/entity` subpackage provides named character references (`entity.Copyright`, `entity.Heart`, …) as `mx.Raw` constants.
- **svg/**: SVG elements and attributes mapped like `html`, with `xmlns` handling, spec-typed numeric attribute values, and typed keyword enums (`enums.go`, generated by go-enum pinned in the nested `tools/` module). See `svg/README.md`.
- **xml/**: Generic XML markup on the same component model as `html`/`svg`, but because XML has no fixed vocabulary it uses **generic** constructors (`Element`, `EmptyElement`, `Attrib`, with `…NS` namespace-prefix variants) instead of one function per tag. Adds the XML-specific constructs (`Comment`, `CDATA`, `ProcInst`, `Declaration`/`Decl`, `Doctype`) and an `xml.Document` for full documents. `Comment`/`CDATA` defer an error rather than emit malformed markup. See `xml/README.md`.
- **hx/**: HTMX integration - request attributes (`hx.Get()`, `hx.Post()`, `hx.Trigger()`, …), with typed values where htmx constrains them: the `SwapStyle` keyword enum for `Swap()` (`enums.go`, generated by go-enum like `html`/`svg`), `bool` arguments for `Boost`/`History`/`Validate`, and boolean attributes (`Disable`, `Preserve`, `HistoryElt`). Plus htmx event-name constants (`events.go`), request-lifecycle CSS class constants (`classes.go`), and HTTP header constants with request readers (`IsRequest`, `IsBoosted`) and response-header setters (`SetRedirect`, `SetTrigger`, …) in `headers.go`. htmx 2.0 moved SSE out of core, so `sse.go` carries the extension's unprefixed attributes (`SSEConnect`, `SSESwap`, `SSEClose`) — they need `Ext("sse")` on the connecting element and `ScriptSSEFromCDN` in the head, alongside `ScriptFromCDN`/`ScriptDebugFromCDN` (all three load from jsDelivr with an SRI hash). See `hx/README.md`.
- **shadcn/**: a Go port of [shadcn/ui](https://ui.shadcn.com) — 65 components, the whole shadcn/ui registry (every one of them in the gallery) plus the `Cn` class-merging helper, with `clsx/`, `twmerge/` and `cva/` subpackages porting the npm packages its class handling builds on (`cva.Config.Composes` and `cva.DefineConfig` from cva v1.0.0-beta.12). Upstream ships React, so this is a *port*, not a wrapper: the new-york-v4 markup and Tailwind classes are transcribed into Go and everything React delegates to Radix is re-expressed with web-platform primitives — native `<dialog>` for the modal overlays (`AlertDialog`, `Dialog`, `Sheet`, `Drawer`), the Popover API plus CSS anchor positioning for the floating ones (`Popover`, `Tooltip`, `HoverCard`, `DropdownMenu`, `ContextMenu`, `Menubar`, `NavigationMenu`), native form controls and `<details>` for the interactive ones (`Checkbox`, `Switch`, `RadioGroup`, `Slider`, `InputOTP`, `Collapsible`, `Accordion`, and a `Select` that is a real `<select>` styled with `appearance: base-select`), CSS scroll-snap and styled scrollbars for `Carousel` and `ScrollArea`, and one small shared inline script only where the platform has no equivalent (`Tabs`, `ToggleGroup`, `Command`, `Resizable`, `Toaster`, `Sidebar`, Drawer's drag-to-close, `CodeBlock`'s copy button, and the select mode and month navigation of `Calendar` that make the `DatePicker` recipe work, including the reflected `form:"widget=datepicker"`, and `Combobox`). `Chart` draws its bar, line, area, pie, radar and radial SVG in Go instead of wrapping Recharts. `Questionnaire` makes every step a server round trip. Upstream's shimmer and scroll-fade `@utility` rules ship as `UtilitiesCSS`, emitted only with `UseUtilities`. The scripts are tested in headless Chrome through `internal/headless` (skipped without a browser, required in CI via `MX_REQUIRE_CHROME`). `DatePicker` and `DataTable` are compositions rather than primitives, so — exactly as upstream ships them — they exist as gallery examples, not as functions in the package. `Combobox` is a recipe upstream too, but here it is a form field (`Combobox(ComboboxOptions{…})`, the reflected `form:"widget=combobox"`): its value lives in a single-option `<select>` laid over the trigger, so `Required` takes part in the browser's constraint validation, which a hidden input never does. Components ported after shadcn/ui's July 2026 restructure (`Field`, `Kbd`, `Spinner`, `Empty`, `Item`, `ButtonGroup`, `InputGroup`, `NativeSelect`, and the chat/agent set `Message`, `Bubble`, `Marker`, `Attachment`, `Trace`, `CodeBlock`, `MessageScroller`) reconstruct their class strings from the new `cn-*` registry instead of transcribing them; upstream's `shimmer` and `scroll-fade-*` come from shadcn's own stylesheet, not Tailwind, so they become `animate-pulse` or are dropped. `CodeBlock` takes highlighted markup as children rather than importing `highlight`. Components finish their element with the exported `StyleSlot` (caller classes merged over the base with `Cn`, one `data-slot`), which add-on packages such as `shadcn/magicui` use too. Plus `shadcn.FieldDecider` (the Tailwind/shadcn form renderer for `mx.ReflectFormHandler`) and the `StickToBottom` attrib that makes a `ScrollArea` (and every `MessageScrollerViewport`) follow a stream. Requires Tailwind CSS v4. See `shadcn/README.md` for the per-component reference and design rationale, and `shadcn/TODOS.md` for the port build order.
- **shadcn/magicui/**: a Go port of [Magic UI](https://magicui.design), the animated components and effects on shadcn's tokens, pinned to `magicui.UpstreamCommit` — 41 components: those that are markup and CSS upstream (`Marquee`, `ShimmerButton`, `OrbitingCircles`, the SVG patterns, the `Iphone`/`Safari`/`Android` mockups, …) and the motion-library ones re-expressed in CSS (`BorderBeam` on `offset-path`, `BlurFade`/`TextAnimate` keyframes with per-segment `--i` delays, `ScrollProgress`/`TextReveal` scroll-driven animations inside `@supports`, `HeroVideoDialog` on `shadcn.Dialog`). Components finish their root with `finish` (wraps `shadcn.StyleSlot` and also merges `style`); an option left at its default is a class on the element (an arbitrary property such as `[--blur-fade-duration:0.4s]`, or an `@utility` in the component's CSS), never `style`, so a component at its defaults works under a Content-Security-Policy without `'unsafe-inline'` (`TestDefaultsWithoutStyleAttribute` in `magicui_test.go` enforces it, its `needStyleAttribute` lists the components that need a value per element, and the README's "Strict Content-Security-Policy" names them); options that are set and carry a CSS value go through `styleVars`, which rejects anything that could end the declaration; `Math.random` becomes a `Seed` drawn through `newRand`; `useId` becomes an `ID` option. Each component's CSS is its own `css/<name>.css` (the `@theme` `--animate-*` entries, keyframes, and a required `prefers-reduced-motion` block switching it off, the element's base style being its final state); they are joined into `magicui.CSS`, which the app adds to its Tailwind input, and `magicui_test.go` enforces that every animation is defined, used and switched off. No inline scripts yet: the scripted and library-based components are open in `shadcn/TODOS.md`. Gallery examples live in the `cmd/shadcn-gallery/examples/magicui` package. See `shadcn/magicui/README.md`.
- **highlight/**: a Go syntax highlighter with no third-party dependencies. `TokenizeGo` scans the source once with the standard library's `go/scanner`, and two backends render from the same tokens: `Component`/`Inline`/`HTML` build highlighted markup through `mx`/`html` (every meaningful token a `<span class="hl-…">`; `Highlighter.Lines` cuts it into one `mx.Components` per source line), while `GoSource` emits the Go source that would build that markup — a code generator, not an echo of the input. Colors live in a separate `Theme` (`theme.go`, `LightTheme`/`DarkTheme`/`OneMonokaiTheme`; `vscodetheme.go` loads any VS Code color theme file as one) that emits the CSS, so the same markup works with any theme, and a `Highlighter` tunes the class prefix and which token classes are highlighted at all. It depends only on the root `mx` package and `html`, **not** on `shadcn`, so it drops into any go-mx markup; `cmd/shadcn-gallery -static-highlight` uses it to render the gallery's code panes server-side. See `highlight/README.md` and `highlight/DESIGN.md` (two backends, `go/scanner`, byte-faithful round-trip).
- **pdf/**: Native PDF rendering mirroring the `html` component model, with the `codeberg.org/go-pdf/fpdf` engine inlined as root-module source so there is no external PDF dependency (`Document`, text/vector/image primitives, typed enums in `constants.go`, and a best-effort SVG renderer: `pdf.SVG` draws `svg`-package element trees as native vector graphics, visual reference in `pdf/testdata/svg_reference.pdf`; plus embedded files and document-level XMP metadata via `Document.Attachments`/`Document.XMP` (`XMPMetadata`, `FacturX`, `AFRelationship`) for PDF/A-3 hybrid formats like ZUGFeRD/Factur-X). The previous wrapper around the external fpdf module lives on in its own repository, `github.com/ungerik/go-mx-fpdf`, as a byte-for-byte parity baseline (apart from the intentional divergences listed in `pdf/README.md`, where the inlined engine deliberately behaves better); `pdf/testdata/legacy_reference.pdf` is its reference PDF, which `TestEngineGoldenParity` compares the inlined engine against, and `TestLegacyParity` (`pdf/parity_test.go`) keeps the scenarios of its parity suite against reference PDFs the legacy stack rendered into `pdf/testdata/parity/` (uncompressed; never regenerate them from `pdf` itself). See `pdf/README.md`.
- **pdf/pdftable/**: Data tables for `pdf` (`Table` is a retained, two-pass `pdf.Component`): fixed/weighted/auto-measured columns, wrapped cells with a cell→row→column→table style cascade, header rows repeated across automatic page breaks, `Grid` rule combinations, and custom-drawn cells via a `Draw` callback. Visual reference in `pdf/pdftable/testdata/table_reference.pdf`. See `pdf/pdftable/README.md`.
- **mcp/**: Exposes the public pages of a `web.Site` to AI agents as read-only MCP tools (`search_pages`, `list_pages` paged by `limit`/`cursor`, `get_page`) on a caller-owned server of `github.com/modelcontextprotocol/go-sdk`, plus `GuardHandler` (Origin and optional Host allowlists, body limit, JSON-RPC batches rejected unless `AllowBatches`, `Cache-Control: no-store, no-transform`). Its own nested module (in `go.work`) so the SDK never touches the core module; import it as `mxmcp`, since the SDK package is `mcp` too. Tool errors shown to clients are plain `errors.New` values, because a go-errs message carries the call stack; internal errors go to `PageToolsOptions.OnError`. See `mcp/README.md`.
- **stdoptions/**: Standard option lists for reflected form fields: ISO 4217 currencies in circulation, ISO 639-1 languages and ISO 3166-1 alpha-2 countries, registered with `mx.RegisterNamedOptions` as `iso4217`, `iso639-1` and `iso3166-1` by an explicit `stdoptions.Register()`. Values are the codes and labels the English or German CLDR names, chosen from the render context (`ContextWithLanguage`, falling back to English) and presorted by CLDR collation. `cldrdata.go` is generated by `tools/gen-stdoptions` from cldr-json at the release pinned in its `cldrRelease` constant (`go generate ./stdoptions`); it is a separate package so only binaries that import it carry the data. See `stdoptions/README.md`.
- **web/**: Website-level abstractions above `html`: a `Page` (metadata plus content) comes from a `PageSource` and is rendered by a `PageRenderer`, while `Site` holds what all pages share (`BaseURL`, title, language) and derives the site-level files from its pages: a `Sitemap`/`SitemapIndex` (sitemaps.org 0.9, rendered through the `xml` package) and a `Robots` robots.txt pointing at it, plus the per-page head metadata (canonical link, description, Open Graph, `noindex` for drafts) added by `Site.AddPageMetadata`. Pages come from a `GlobPageSource` (a glob matched below its `Dir`, which `Confine` turns into a symlink-proof boundary, markdown front matter parsed, URL paths derived from file paths) or any other `PageSource`; serving the pages themselves is still up to the caller's routes, while `Site.RegisterRoutes` serves robots.txt and the sitemap, and with `Site.LLMsTxtOptions`/`Site.MarkdownAlternates` also llms.txt and the `.md` alternates. The public knowledge layer builds on it, with `Page.Indexable()` as the one predicate for "public" that every surface filters on: `Page.Alternates` (translations get a self hreflang link from `Page.Lang`/`Site.Lang` and an x-default from `Site.DefaultLang`) rendered from one list into both the HTML head (`AddPageMetadata`) and `Link` headers (`Site.PageHeaders`); Markdown alternates (`MarkdownPath`, `Page.HasMarkdownSource`, `Site.MarkdownAlternate`, `Site.ServeMarkdown`, `Site.ServeMarkdownDocument` for an alternate extended with `Sections`, `AcceptsMarkdown`); and `LLMsTxt` with `Site.LLMsTxt`. **web/schemaorg/** is a typed schema.org subset for JSON-LD (prices as `Price(cents)` strings, never floats), embedded with `html.JSONLD`. Nothing in `web` may know what a site sells (prices, offerings, operators): that stays downstream. The package has no README; its guide is `docs/web/knowledge-layer.md`.
- **logview/**: Streamed log lines rendered as a readable, searchable surface, fed by `mx.SSEResponse`. `Config.Line` parses one raw line: a single JSON object becomes its fields in source order — time and level promoted without labels, everything else `key=value` with the value colored by its JSON type, nested objects in braces, arrays in brackets, and a multi-line string value in a `<pre>` so a stack trace keeps its indentation — while anything else becomes a plain-text line with its leading level word colored. ANSI escape sequences are stripped before parsing unless `Config.KeepANSI`; `Config.OverwriteCR` keeps only what follows a line's last carriage return, as a terminal shows a redrawn progress bar, and `Config.Expandable` wraps each record in a `<details>` whose body pretty-prints it. `Config.View` is the surface: a filter over the received lines that `Config.SearchParam` turns into a server-side search (the view reconnects by swapping its inner `data-mx-log-stream` connecting element, which carries `sse-connect`/`sse-close`; handlers answer with `Config.Matches`), a pause toggle that holds the view without dropping anything, `shadcn.ScrollArea`+`StickToBottom` following the stream, and an `mx.SSEEventError` sink. One inline script drives pause/filter/line-cap through fixed `data-mx-log-*` attributes. A level value never reaches a class attribute: only tokens the `Theme` defines are emitted, everything else is `log-level-undefined`. Worked example in `cmd/example-logstream`. See `logview/README.md`.
- **doc/**: Higher-level abstractions (partially implemented)

### Key Patterns

**Element Creation**:
```go
// Regular elements take attribs and children as variadic any
html.Div(html.Class("container"), html.P("Hello"))

// Void elements take only Attrib (no children allowed)
html.Input(html.Name("q"), html.Placeholder("Search"))

// An attribute the HTML spec requires is the first argument and its only source
html.Img("/logo.png", html.Alt("Logo"))
```

Constructors with a required first argument: `Img(src)`, `Track(src)`,
`OptGroup(label)`, `Data(value)`, `Map(name)`, `BDo(dir Dir)`. The same
attribute among the others is a duplicate-attribute render error, and so is a
value the spec forbids (empty `src`, empty or whitespace `Map` name, a `BDo`
dir other than `DirLTR`/`DirRTL`), deferred with `mx.ErrAttribf`. A new
element whose spec requires one attribute follows the same pattern; see
`html/README.md` "Required attributes come first".

**Attribs before children**: in every `attribsChildren ...any` constructor
(`html.Div`, `NewElement`, the `shadcn` components, …) always pass attributes
before children. `NewElement` sorts the variadic into the separate `Attribs`
and `Children` slices regardless of order, but the source must list attribs
first, then children — including when appending to the variadic, e.g.
`html.Div(append(attribsChildren, child)...)` (child added last).

**Component Conversion** (`DefaultAsComponent` in `component.go`):
Children are passed as `...any` and converted dynamically at render-build
time, not checked at compile time.
- `nil` → nil (renders nothing)
- `string` → `Text` (escaped)
- `Component` → passes through
- attributes (anything `AsAttribs` accepts: an `Attrib`, `[]Attrib`,
  `Attribs` map, `AttribProvider`, struct with `attr` tags) → space-separated
  `name='value'` texts from `AttribStringContext` (name and value already
  escaped, a nil `Attrib` left out), valued on render with the render
  context, so a failing value makes `Render` return its error. Element
  constructors sort attributes out first, so this case catches attributes
  passed where children are converted directly (`AsComponents`,
  `ModifyOnRender`). A `Component` that also converts to attributes stays a
  `Component`.
- Functions → wrapped as `ComponentFunc`
- `error` / `fmt.Stringer` → `Text` of `Error()` / `String()`
- anything else → `Text(pretty.Sprint(value))` using
  `github.com/domonda/go-pretty`, rendered as escaped text. `pretty` gives
  primitives their plain form; structs and pointers get a type-tagged
  single-line dump (e.g. `` Item{Name:`x`;Count:3} `` instead of fmt's
  anonymous `{x 3}`) while slices and maps stay literal (`[1,2]`); it
  dereferences pointers, keeps the text single-line, and bounds the length. Because it becomes a `Text` node the writer escapes
  it on render — so a stringified value can never inject markup, escaping is
  the same for HTML, XHTML, SVG and XML (one `TextEscaper`; escaping is a
  writer concern, so `AsComponent` stays target-agnostic), and escaping the
  data content is still mandatory. The trade-off is that a value passed by
  mistake (a struct that does not implement `Component`, a `*T` whose methods
  are on `T`, …) is rendered as its escaped pretty representation with no
  compile error; convert non-obvious children to a `Component` explicitly.
  The package-level `AsComponent` var (`config.go`) can be reassigned to
  change this behavior program-wide (e.g. panic/log on unexpected types, or
  use `fmt.Sprint`). The `pdf` package mirrors all of this except the
  `Attrib` case — same `error`/
  `fmt.Stringer`/`pretty.Sprint` cases and a reassignable `pdf.AsComponent`
  var (`pdf/config.go`) — except there is no escaping step (a PDF is not
  markup).

**Attribute Conversion** (`DefaultAsAttribs` in `attribs.go`):
- Single `Attrib` or `Attribute`, slices, iterators and constructor funcs of
  those, `AttribProvider`, maps, structs with `attr` tags. A nil func or
  iterator, or a typed-nil `AttribProvider` (nil pointer or func), gives no
  attributes instead of being called; a typed-nil `Attrib` is kept, for
  `Element.Render` to refuse (`IsNilAttrib`).
- `AsAttribsOrComponent` (`element.go`) sorts one `attribsChildren` argument
  into attributes or a child; `NewElement` and `html.NewDocument` share it,
  so a new constructor that sorts its arguments calls it too.

**Conditional Rendering** (`conditional.go`):
```go
mx.If(condition, component1, component2).Else(fallback)
mx.ForEach(slice, func(v T) Component { ... })
```

### Reflection Features

- `ReflectAttribs()`: Extract attributes from struct fields using `attr` tag
- `ReflectFormFields()`: Iterator over form-relevant struct fields (`form:"…"` tags), the walk behind `mx.ReflectFormHandler` and the `FieldDecider` chain
- `ReflectStructFields()`: Iterator over flattened struct fields (handles embedded)

## Code Conventions

- Use `any` instead of `interface{}`
- Go file names have no underscores except where the suffix has a meaning: `_test.go`, a `_GOOS`/`_GOARCH` build-constraint suffix, or a platform pair split by `//go:build` like `globpagesource_unix.go`/`globpagesource_other.go`. Join the words instead (`validationchain.go`, `animatedcircularprogressbar.go`).
- Use `github.com/domonda/go-errs` for errors (`errs.New`, `errs.Errorf`)
- Use the standard library `uuid` package for UUIDs (`uuid.UUID`, `uuid.New()`, `uuid.Nil()`; `Nil` is a function, not a variable)
- Closed sets of typed constants are go-enum types: `type X string //#enum` plus the `//go:generate` line of e.g. `html/enums.go`. Check and list values with the generated `Valid`/`Validate`/`Enums`/`EnumStrings` methods; use `ValidOr(default)` to resolve an empty or unknown value to its default where that is the desired behavior.
- Enum options in a JSON Schema come from `//#enum,jsonschema`. Its `JSONSchema()` returns an invopop schema, which the MCP SDK's `google/jsonschema-go` ignores, so an `mcp/` tool parameter needs `jsonschema.ForOptions.TypeSchemas` filled from `EnumStrings()`.
- SQL strings: prefix with `/*sql*/` and use backticks and start on a new line when following an argument
- HTML strings: prefix with `/*html*/` and use backticks and start on a new line when following an argument
