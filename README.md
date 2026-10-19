# go-mx

[![Go Reference](https://pkg.go.dev/badge/github.com/ungerik/go-mx.svg)](https://pkg.go.dev/github.com/ungerik/go-mx)
[![Go Report Card](https://goreportcard.com/badge/github.com/ungerik/go-mx)](https://goreportcard.com/report/github.com/ungerik/go-mx)
[![Go version](https://img.shields.io/github/go-mod/go-version/ungerik/go-mx)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

go-mx builds HTML in Go. Every HTML element is a Go function call, and calls
compose into a tree of components. Rendering walks the tree and writes HTML,
escaping text and attribute values and checking structural validity as it goes.

No templates and no separate template language: markup is ordinary Go, so it
composes and refactors with the rest of your code.

For the *why* — server-rendered HTML versus the SPA default, an honest
advantages-vs-disadvantages overview, and the use cases go-mx is built for —
see [docs/why-go-mx.md](docs/why-go-mx.md).

## Documentation

- **[Documentation site](https://ungerik.github.io/go-mx/)** — overview, the
  `html` and `shadcn` tutorials and how-to guides, the public knowledge layer
  guide, and the rendered component gallery. (GitHub Pages, served from
  [`docs/`](docs/).)
- **[Component gallery](https://ungerik.github.io/go-mx/shadcn/gallery/)** — every
  ported shadcn/ui component shown next to its Go source.
- **[`html` package](html/README.md)** — the core HTML vocabulary: elements,
  attributes, typed keyword enums, and the `html/entity` character set.
- **[`shadcn` package](shadcn/README.md)** — the shadcn/ui port: full component
  reference and design notes.
- **[API reference](https://pkg.go.dev/github.com/ungerik/go-mx)** on pkg.go.dev.
- **[CHANGELOG.md](CHANGELOG.md)** — what changed. Nothing is tagged yet, so
  everything is still under `Unreleased` and the API is free to change.
- **[TODOS.md](TODOS.md)** — deferred work by area and priority (the
  shadcn/ui port build order is in [shadcn/TODOS.md](shadcn/TODOS.md)).

## Install

```sh
go get github.com/ungerik/go-mx
```

Requires Go 1.27 or newer.

## Example

```go
package main

import (
	"fmt"

	. "github.com/ungerik/go-mx/html"
)

func main() {
	page := Div(Class("greeting"),
		H1("Hello, world"),
		P("Built with go-mx."),
	)
	fmt.Println(page) // *mx.Element implements fmt.Stringer
}
```

For streaming output or serving over HTTP, render a `Component` into an
`mx.Writer` with `Component.Render(ctx, writer)`, or use
`mx.ComponentHTTPHandler`.

## Packages

- **`mx`** (root) — core abstractions: `Component` (anything that renders),
  `Element`, `Attrib`, `Writer`, and `CheckedWriter` (escaping, structural
  validation, optional indentation). Plus conditional rendering (`If`,
  `ForEach`), struct-reflection helpers, and `ReflectFormHandler` (see
  below).
- **`html`** — HTML5 element and attribute constructors: `html.Div`,
  `html.Span`, `html.Class`, `html.ID`, and the rest of the HTML5 surface,
  with typed keyword enums that validate at render time and numeric
  attribute constructors that format Go values for you. Provides
  `html.FieldDecider` for plain-HTML form rendering. The
  [`html/entity`](html/entity) subpackage adds named character references
  (`entity.Copyright`, `entity.Heart`, …).
- **`svg`** — SVG element and attribute constructors, mapped the same way as
  `html`: `xmlns` namespace handling, spec-typed numeric attribute values, and
  typed keyword enums. Use `svg.Root` for a standalone document or `svg.SVG`
  for inline `<svg>`. See [svg/README.md](svg/README.md).
- **`xml`** — generic XML markup: `xml.Element`/`xml.EmptyElement` and
  `xml.Attrib` take any name (XML has no fixed vocabulary), with `…NS`
  namespace-prefix shortcuts, plus the XML-specific constructs `xml.Comment`,
  `xml.CDATA`, `xml.ProcInst`, `xml.Declaration`/`xml.Decl` and `xml.Doctype`,
  and an `xml.Document` for full documents. See [xml/README.md](xml/README.md).
- **`hx`** — HTMX integration: request attributes (`hx.Get`, `hx.Post`,
  `hx.Trigger`, …), with typed values where htmx constrains them — a
  `hx.SwapStyle` enum for `hx.Swap`, `bool` arguments for `hx.Boost`/
  `hx.History`/`hx.Validate`, and boolean attributes like `hx.Disable`/
  `hx.Preserve`. Adds `htmx:` event and `htmx-*` CSS class name constants, the
  `sse` extension's attributes (`hx.SSEConnect`, `hx.SSESwap`, `hx.SSEClose`)
  for consuming an `mx.SSEResponse`, plus a server side for HTTP handlers:
  `hx.IsRequest`/`hx.IsBoosted` request readers and
  `hx.SetRedirect`/`hx.SetTrigger`/… response-header setters.
  Provides `hx.FieldDecider` that wraps `html.FieldDecider` and adds
  `hx-trigger="change"` to live inputs. See [hx/README.md](hx/README.md).
- **`shadcn`** — `Cn`, a faithful Go port of tailwind-merge v3, plus ported
  shadcn/ui components and `shadcn.FieldDecider` for Tailwind/shadcn form
  rendering. See [shadcn/README.md](shadcn/README.md).
- **`highlight`** — Go syntax highlighter: tokenizes Go source and renders it as
  highlighted HTML components, or as the go-mx source that builds that markup,
  plus a `Theme` that emits CSS. Depends only on `mx` and `html`. See
  [highlight/README.md](highlight/README.md).
- **`logview`** — streamed log lines rendered as a readable surface: a JSON
  record becomes promoted time and level plus `key=value` pairs colored by JSON
  type, a non-JSON line becomes plain text with its level word colored, and
  `logview.View` adds the filter, pause and scroll-anchored viewport they stream
  into. Fed by `mx.SSEResponse`. See [logview/README.md](logview/README.md).
- **`pdf`** — native PDF rendering mirroring the `html` component model, with
  the [`codeberg.org/go-pdf/fpdf`](https://codeberg.org/go-pdf/fpdf) engine
  inlined so there is no external PDF dependency: `pdf.Document`,
  text/vector/image primitives, typed enums, a best-effort `pdf.SVG` that
  draws `svg`-package element trees as native vector graphics, and embedded
  files plus XMP metadata (PDF/A-3 associated files for ZUGFeRD/Factur-X
  e-invoices). The previous
  wrapper around the external fpdf module lives on as the **`fpdf`** legacy
  module, whose parity test suite asserts byte-identical output from both
  packages — apart from a few [documented intentional
  divergences](pdf/README.md#divergences-from-the-legacy-engine) where the
  inlined engine deliberately behaves better — until the legacy module is
  deleted. See
  [pdf/README.md](pdf/README.md) (and [fpdf/README.md](fpdf/README.md) for the
  legacy module).
- **`pdf/pdftable`** — data tables for `pdf`: measured fixed/weighted/auto
  columns, wrapped and aligned cell text with a style cascade, header rows
  repeated across automatic page breaks, grid rules, and custom-drawn cells
  (e.g. SVG icons). See [pdf/pdftable/README.md](pdf/pdftable/README.md).
- **`wordpress`** — imports a WordPress WXR export into a typed Go model and
  renders it as a static shadcn site plus an import-diagnostics report. Its own
  nested module, so its `golang.org/x/net/html` dependency stays out of core. See
  [wordpress/README.md](wordpress/README.md).
- **`web`** — a site above `html`: pages from a `PageSource`, rendered by a
  `PageRenderer`, and a `Site` that derives what crawlers and agents read from
  its public pages: sitemap and robots.txt, canonical and hreflang links (in
  the HTML head and as `Link` headers from one list), Markdown alternates of
  pages, and an `llms.txt` index. The
  [`web/schemaorg`](web/schemaorg) subpackage is a typed schema.org subset for
  structured data, embedded with `html.JSONLD`. See
  [docs/web/knowledge-layer.md](docs/web/knowledge-layer.md).
- **`mcp`** — the public pages of a `web.Site` as read-only
  [Model Context Protocol](https://modelcontextprotocol.io) tools for AI
  agents, added to a server of the official Go SDK, plus a guard for the HTTP
  endpoint. Its own nested module, so the SDK stays out of core. See
  [mcp/README.md](mcp/README.md).
- **`doc`** — higher-level abstractions, partially implemented.

## Streaming

`mx.SSEResponse` streams components to a client as Server-Sent Events over a
still-open HTTP response, rendering one component per event. go-mx otherwise
buffers a whole response before writing it, so that a deferred computation
failing mid-render becomes a clean 500 instead of a truncated page;
`SSEResponse` keeps that guarantee per *event* rather than per response, and
`SendError` carries a later failure in band, since after the first flush the 200
is committed and there is no 500 left to send.

A browser reconnects to a dropped stream on its own and does it silently, so
`SSEResponse.SendEvent` takes an `SSEEvent.ID` and `mx.LastEventID(request)`
reads back what the client acknowledged — without them a reconnect replays the
whole stream or skips what it missed. `SetRetry` tunes the reconnect delay and
`KeepaliveLoop` keeps an idle connection from being dropped at all. A stream
that never ends needs `http.Server.WriteTimeout` cleared, which would leave
writes unbounded, so `SetWriteTimeout` bounds a single frame instead: a client
that stops reading cannot pin the producer goroutine forever.

Around it: `mx.KeyedID` derives an element id that is stable across renders and
processes (which `mx.UniqueID`'s counter cannot), so a later event can address
an element an earlier one rendered; `hx.SSEConnect` / `hx.SSESwap` /
`hx.SSEClose` are the htmx SSE-extension attributes; and
`shadcn.StickToBottom` makes a `ScrollArea` follow growing content unless the
user has scrolled up.

See [`cmd/example-sse`](cmd/example-sse/main.go) for a worked example that
streams a chat transcript and appends tokens to one message out of band by its
`KeyedID`. Its `-drop` flag cuts the first connection mid-reply so the browser
reconnects and the handler resumes, and `-fail` reports an error in band.
[`cmd/example-logstream`](cmd/example-logstream/main.go) streams a simulated
production log through the `logview` package, resumable from a bounded history.

## Reflected forms

`mx.ReflectFormHandler[T]` builds a full http.Handler for a struct
type — rendering, parsing, validation, and load-then-apply against
your record store, in one call:

```go
import (
    "net/http"
    "github.com/ungerik/go-mx"
    "github.com/ungerik/go-mx/shadcn"
)

mux := http.NewServeMux()
mux.Handle("/admin/profile", mx.ReflectFormHandler(loadProfile, saveProfile))
http.ListenAndServe(":8080", mx.Middleware(shadcn.FieldDecider)(mux))
```

The decider lives in request context. A custom `FieldDecider` can be
installed via `mx.Middleware(...)` once per route subtree, or passed
as an optional variadic to a single handler. Validation runs a
richest-first chain (`Normalize() []error` → `Normalize() error` →
`Validate() error` → `Valid() bool`), per-field tags drive widget
choice, and POST parses only the fields the form actually rendered —
mass-assignment-safe by construction.

On success the handler 303-redirects to `cfg.Redirect` (or the request
path when nil). When the destination is only known after submitting — a
Stripe Checkout Session, a signed upload URL, an OAuth handoff — use
`mx.ReflectFormHandlerRedirecting`, whose `onSubmit` returns
`(redirect string, err error)`; the returned URL wins, and an empty
string falls back to `cfg.Redirect` then the request path.

The `date`, `datetime` and `time` widgets render the matching input for a
`time.Time` and also for a field that holds its value as text: a `string`, a
`*string`, or a named string type such as a date type. A type can declare its
widget once with `FormWidget() string` (`mx.FormWidgetHint`) instead of every
field carrying `form:"widget=date"`. The input shows the stored value: a
`Format(string) string` method is asked for it, and otherwise text in the
input's own form is shown. That form has no zone and carries seconds only when
they are not zero: `2006-01-02`, `2006-01-02T15:04[:05]` and `15:04[:05]`. On
POST the submitted value is stored in that same form, through `UnmarshalText`
when the type has it, and an empty value stores the zero value, which
`required` rejects. An input cannot show stored text in another spelling
(`06.07.2026`), an invalid value or a time with a zone, and a blank one would
clear it on the next save, so such text renders in a text input instead, with
the format to enter as placeholder and hint. Saving it unchanged then fails
with a field error until it is corrected, and the stored value stays as it
was. A type that keeps its own spelling implements `Format` and
`UnmarshalText` to convert.

Slice-of-struct fields tagged `form:"repeatable"` (`[]T` or `[]*T`)
render as a dynamic list of rows — one per element, with JavaScript-free
"Add row" / "Remove" buttons — and bind the submitted rows back into the
slice. A `form:"hidden"` sub-field round-trips a per-row key so `onSubmit`
can map rows to records.

Option-driven fields (selects, radio groups, checkbox sets) resolve their
option lists via `mx.CollectOptions`: `Enums()`/`EnumStrings()` methods, the
static `mx.OptionsProvider`/`mx.NamedOptionsProvider` interfaces, the
context-aware `mx.NamedOptionsContextProvider`, or a `form:"options=name"`
tag naming a provider registered with `mx.RegisterNamedOptions`. The last
two are resolved at render time with the request context, so a dropdown can
be per-request (a tenant-scoped partner list from the database) without the
field type carrying any UI state:

```go
mx.RegisterNamedOptions("partners", func(ctx context.Context) ([]mx.NamedOption, error) {
    return queryPartnerOptions(ctx) // tenant + DB from the request context
})

type InvoiceDraft struct {
    PartnerID uuid.UUID `form:"options=partners,label=Partner,required"`
}
```

The per-request option list is a trust boundary for values the browser
submits: on POST the handler re-resolves the list with the request context
and rejects submitted values outside it as a normal field validation error,
so a tenant-scoped dropdown cannot be bypassed by posting another tenant's
ID directly. On render, a select whose current value is no longer in that
list shows a generic disabled placeholder (the filtered-out value is not
echoed). A sensitive field uses that same generic placeholder: the value
stays out of the markup, and an unchanged submit does not fall through to
the first listed option. Because the placeholder is disabled, an unchanged submit sends no
value for the field: required fields fail validation until a value is picked
from the list, while non-required fields are cleared.

When the option list has no empty value, a select with no current value
gets a leading empty option so the browser does not submit the first real
option. Required fields use a disabled "Choose…" option, which also makes
the `required` attribute fire; optional fields use a selectable "(none)".
An optional select that already has a listed value also gets "(none)"
when the list has no empty value, so the choice can be cleared. A
context-dependent list that comes back empty still renders: a stored value
shows the same generic placeholder, and a blank required field shows
"Choose…". An empty answer is a real "nothing to offer", so it does not
fail the rest of the form. A provider that returns an error still does.

An enum-set checkbox group renders only the options in the current list.
Members already stored but absent from a context-dependent list are not
shown and not echoed, and a save does not keep them: the checked boxes are
the whole set that is stored, so a value the list stopped offering cannot
be written back, and unchecking a box that was on the page removes it even
if the list changes before the POST. The group says that some current
values are not available, without naming them. A static enum list is the
whole domain, so parse replaces the set from the submission. A blank
optional enum is no choice: the type's `Validate` does not reject it, and
a decider `Validate` still runs.

The rendered `<form>` carries an explicit `action` attribute. By default it
self-submits: the action is the request URI that served the form (path+query),
forced same-origin (a leading `//` is collapsed so the action can never become
a protocol-relative off-origin URL). This is what lets a form loaded as an HTMX
fragment (via `hx-get` into a GET-only page) post back to its own handler
rather than to the embedding document — without it a native submit would 405.
Set `ReflectFormConfig.Action func(*http.Request) string` to point the form
somewhere else; a non-nil `Action` is trusted and emitted as-is (escaped, but
not scheme/host allow-listed), so do not feed it untrusted input.

See [`cmd/example-form`](cmd/example-form/main.go) for a complete
worked example covering every supported field kind, section grouping,
the repeatable invoice-line editor, and the `FieldErrors` cross-field
error routing path.

## To Do

The `mx` reflection to-do list. Everything else deferred — per package, with
priorities and context — is in [TODOS.md](TODOS.md).

- [ ] ReflectMarkup()
- [ ] More ReflectInputOptions
- [x] Slice-of-struct fields (`form:"repeatable"`)
- [ ] Rich file upload widget (preview, multi-file, progress)
- [ ] HTMX OOB fragment responses
- [ ] OptionsProvider registry sub-package (ISO 4217, ISO 639, country codes)

## License

go-mx is licensed under the [MIT License](LICENSE), with one exception: the
[`shadcn/cva`](shadcn/cva/) subpackage is a port of
[class-variance-authority](https://github.com/joe-bell/cva) and is licensed
under the Apache License, Version 2.0 (see [`shadcn/cva/LICENSE`](shadcn/cva/LICENSE)).

Parts of the project port or adapt other open-source work (shadcn/ui, clsx,
tailwind-merge, and GitHub's primer syntax palette); their license texts and
copyright notices are reproduced in
[THIRD-PARTY-LICENSES.md](THIRD-PARTY-LICENSES.md).
