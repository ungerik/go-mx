# TODOS

Deferred work, grouped by component and sorted by priority. The older `mx`
reflection to-do list lives in `README.md`; the shadcn/ui port build order lives
in `shadcn/TODOS.md`.

Priorities: **P0** blocking · **P1** critical, this cycle · **P2** important ·
**P3** nice-to-have · **P4** someday.

## shadcn/cva

Follow-ons from the initial `cva` port (class-variance-authority v0.7.1). The
subpackage currently exports only `New(config Config) Variants`
(`shadcn/cva/cva.go:65`), which covers the variant resolution the ported
components actually use.

### Port cva's `compose`

**What:** Merge several variant resolvers into one, as npm cva's `compose` does.

**Why:** Lets a component build on another component's variants instead of
restating them, which is how upstream shares a base look across a family.

**Context:** No component in this repo needs it yet — every port declares its own
`Config`. Worth doing when the first component would otherwise copy another's
variant table.

**Effort:** S
**Priority:** P3
**Depends on:** None

### Decide on a Go equivalent for cva's `VariantProps`

**What:** Provide a Go equivalent of the `VariantProps` type helper, or document
that a typed props struct replaces it.

**Why:** `VariantProps` is how TypeScript callers get compile-time checking of
variant names. Go's answer is a typed struct per component, and the port should
say so explicitly rather than leave readers looking for the missing helper.

**Context:** This is mostly a documentation decision — the existing components
already take typed variant parameters (`ButtonVariant`, `ButtonSize`), which is
the Go equivalent in practice. Write it down in `shadcn/cva/README.md` and close
the item unless a real generics-based helper turns out to be worth the surface.

**Effort:** S
**Priority:** P3
**Depends on:** None

### Port cva's `defineConfig` / `onComplete` hook

**What:** Port the `defineConfig` factory and its `onComplete` hook.

**Why:** Upstream uses it to post-process every generated class string in one
place — the natural seam for wiring `twmerge` in globally instead of per call
site.

**Context:** The most speculative of the three: this repo already merges classes
through `Cn` at the component boundary, so the hook has no current job. Revisit
only if a caller needs to intercept class generation program-wide.

**Effort:** S
**Priority:** P4
**Depends on:** None

## logview

Deliberate omissions from the first version, recorded so they read as decisions
rather than oversights. The renderer and the view are complete without them.

### ANSI escape sequences render as literal text

**What:** A log line carrying SGR sequences (`\x1b[31m`) shows them as `[31m`
garbage instead of as color.

**Why:** Real streams carry them constantly — CI output, `docker logs`, anything
using a colorizing logger. They are noise in every line they appear in, and the
one thing they encode, severity, is exactly what the view already colors.

**Context:** Two shapes: strip SGR sequences in `Config.Line` before parsing, or
map them to spans with their own classes. Stripping is a few lines and fixes the
noise; mapping is more work and would need a decision about how ANSI's 16 colors
relate to a `Theme`. Start with stripping, behind a `Config` field, because a
caller who pipes raw terminal output usually wants it gone rather than rendered.

**Effort:** S
**Priority:** P3
**Depends on:** None

### The filter searches the DOM, not the backlog

**What:** The filter matches the lines currently in the view. Lines the
`MaxLines` cap already dropped, and everything behind the stream, are not
searched.

**Why:** A reader who types `panic`, sees nothing and concludes there was no
panic has been misled. The control is labelled "Filter" rather than "Search" for
that reason, but the honest fix is a real search.

**Context:** It belongs on the server, as a parameter of the stream URL, which
is a smaller change than any client-side alternative — the handler already
answers "everything after this id" and would answer "everything matching this"
the same way. Changing the filter would then have to reconnect the stream, since
`hx.SSEConnect` takes a static URL, so the view needs a way to swap its
connecting element.

**Effort:** M
**Priority:** P3
**Depends on:** None

### A carriage return is a line break, not a terminal overwrite

**What:** A progress-bar line that rewrites itself with `\r` renders as several
lines instead of one.

**Why:** It is a display-semantics choice, not corruption: `mx.SSEResponse`
already normalizes CR and CRLF to LF on the wire, so the bytes are intact.

**Context:** Collapsing `\r` runs — keeping only the text after the last one in
a segment — is a few lines in `Config.Line`, and would want to be a field rather
than the default, since a log that uses `\r` as a plain separator would lose
content.

**Effort:** S
**Priority:** P4
**Depends on:** None

### Expandable pretty-printed detail for a record

**What:** A collapsed line that expands to an indented, highlighted rendering of
the whole JSON record.

**Why:** A record with several nested objects gets long on one line, and the
compact `{k=v k=v}` form is harder to read the deeper it nests.

**Context:** `<details>` around the line, with the summary being what
`Config.Line` renders today and the body an indenting variant of the same value
renderer. The parse already produces an ordered tree, so only the rendering side
is new. Note the `<pre>` rule: a block with element children inside it is not
byte-safe under an indenting writer.

**Effort:** M
**Priority:** P4
**Depends on:** None

## Documentation site

### No docs-site page for `hx` or the streaming surface

**What:** The site at <https://ungerik.github.io/go-mx/> has tutorial, how-to
and reference pages for `html` and `shadcn` only. `hx` has none, and neither
does the streaming surface (`mx.SSEResponse`, `mx.KeyedID`, the `hx.SSE*`
attributes, `shadcn.StickToBottom`, `logview`).

**Why:** Streaming is documented — `README.md`, `hx/README.md`,
`logview/README.md` and two worked commands all cover it — but none of that is
on the site. A reader who starts at the docs site finds nothing and concludes
go-mx cannot stream, while the feature is one repo click away. Reference and
how-to coverage exist; the tutorial quadrant is empty for the whole surface.

**Context:** The gap predates the streaming work: `hx` never had a site page
either, so this is the site's structure trailing the repo rather than anything
this cycle removed. `docs/index.md` names `logview` and `mx.SSEResponse` in the
package table, which is the only site-level mention. The existing
`docs/html/{index,tutorial,how-to}.md` trio is the shape to copy; the gallery
under `docs/shadcn/gallery/` is generated and unaffected. `cmd/example-sse` and
`cmd/example-logstream` are ready-made tutorial material.

**Effort:** M
**Priority:** P3
**Depends on:** None

## web knowledge layer

Deliberate omissions from the alternates, Markdown alternates and llms.txt
work, recorded so they read as decisions rather than oversights.

### `Site.RegisterRoutes` serves neither llms.txt nor the Markdown alternates

**What:** `Site.RegisterRoutes` registers robots.txt and the sitemap, but not
`LLMsTxtPath` or the `.md` URLs that `MarkdownPath` derives.

**Why:** Both are opt-in publishing decisions — which pages an agent may read,
whether the alternate is served at its own URL, under the page URL by
`AcceptsMarkdown`, or both, and with which revision as ETag — so a site wires
them explicitly instead of getting them by default.

**Context:** A `Site.LLMsTxtOptions` field plus a catch-all `GET /{path...}.md`
route resolving the page by `MarkdownPath` would make it one line for the common
case. ServeMux cannot match a suffix, so the route needs a pattern per page or
a handler that strips `.md` itself.

**Effort:** S
**Priority:** P3
**Depends on:** None

### `list_pages` returns every public page at once

**What:** The MCP `list_pages` tool has no limit or cursor; it returns all
public pages in one response, which the SDK sends twice (structured content
plus its JSON text fallback).

**Why:** A site with thousands of pages answers with megabytes that overflow
an agent's context, while `search_pages` and `get_page` are bounded. The spec
asked for "all matching pages", so the first version keeps that.

**Context:** A `limit` with a `truncated` flag like `search_pages`, or a
`cursor`, in `mcp/pagetools.go` `listPages`. Decided in the /ship review of the
knowledge layer (2026-10-15) to keep the spec'd behavior for now.

**Effort:** S
**Priority:** P3
**Depends on:** None

### Page tools redo per-page work on every call

**What:** `Page.MarkdownSource` copies string content into a new `[]byte` on
every call, and `get_page` builds the full `PageInfo` (link validation, URL
building) for every public page while looking for one.

**Why:** With a cached source of 1000 pages of 16 KiB string content,
`list_pages` and `get_page` allocate about 22 MB per call, and `get_page` costs
about 4.6 ms where a path comparison per page would do.

**Context:** Give ok-only callers (`Site.LLMsTxt`, the filter in
`mcp.publicPages`) a check without the copy and keep string content as a
string in `publicPage`; in `getPage`, normalize the input to a page path once
(strip the BaseURL, map a `.md` path back) and build `PageInfo` only for the
match, while still draining the source for errors. Measured by the performance
specialist of the /ship review (2026-10-15).

**Effort:** S
**Priority:** P3
**Depends on:** None

### `GlobPageSource` drops front matter fields it does not know

**What:** Since the front matter is no longer part of a Markdown page's
`Content`, fields other than title, description, lang, date, publishDate,
draft and tags (an `image`, `layout`, `aliases`) are lost.

**Why:** A `ContentRenderer` that parsed the front matter itself for fields of
its own has nothing to read anymore. The CHANGELOG records the loss.

**Context:** Expose the parsed front matter, e.g. a `Page.FrontMatter
map[string]any` filled by `GlobPageSource` in `web/globpagesource.go`, or an
option keeping the raw file as content.

**Effort:** S
**Priority:** P4
**Depends on:** None

### No `og:locale:alternate` for translations

**What:** `Site.AddPageMetadata` emits hreflang links for translations but no
Open Graph `og:locale:alternate` tags.

**Why:** Link previews (Facebook, LinkedIn) read the Open Graph locales, not
hreflang, to offer a translated preview.

**Context:** `html.Document.MetaProperty` is a `map[string]string`, which holds
one value per property, while `og:locale:alternate` repeats once per locale.
It needs a document field that allows repeated properties, or the tags rendered
into `HeadCustom` next to the alternate links from the same list.

**Effort:** S
**Priority:** P4
**Depends on:** None

### `mcp.GuardHandler` sends no CORS headers

**What:** `GuardHandler` lets requests from `AllowedOrigins` through but neither
answers a CORS preflight nor sets `Access-Control-Allow-Origin`, so a page on
another origin cannot call the endpoint even when its origin is allowed.

**Why:** Which headers, methods and credentials an endpoint allows is a policy
of the site, and same-origin pages and non-browser clients — what a public
page endpoint serves — need no CORS at all. The guard only rejects; the doc
comment and the mcp README say so.

**Context:** CORS for the Streamable HTTP transport means answering `OPTIONS`
for allowed origins with `GET, POST, DELETE` and the `Content-Type`,
`Mcp-Session-Id`, `Mcp-Protocol-Version` and `Last-Event-ID` request headers,
and exposing `Mcp-Session-Id` on responses. A `GuardOptions.CORS` flag reusing
`AllowedOrigins` would cover it.

**Effort:** S
**Priority:** P4
**Depends on:** None

## mx reflected forms

### Date and time inputs drop the stored time zone

**What:** Bind a submitted date or time into a `time.Time` (or a type that
wraps it) in the location of the value it replaces, instead of UTC.

**Why:** A `time.Time` stored as 2026-07-06T09:00+02:00 shows 09:00 in the
input, which has no zone, and an unchanged save stores 09:00 UTC: the time
moves by the offset on every save. Types that wrap `time.Time` behave the same
way since they render and bind like it.

**Context:** `setTime` in `html/formdecider.go` parses with `time.Parse`, which
gives UTC; `time.ParseInLocation` with the loaded value's `Location()` keeps the
wall clock in its zone. It changes what existing `time.Time` fields store, so it
was left out of the string-kind date work. Fields holding text are not
affected: their text has no zone, and text with one falls back to a text input.

**Effort:** S
**Priority:** P2
**Depends on:** None

### Fractional seconds in date and time inputs

**What:** Accept and keep fractional seconds in datetime and time inputs, so
a `form:"step=…"` under one second works.

**Why:** A sub-second step makes the browser submit values such as
`2026-07-06T09:05:30.123`. A field holding text rejects them, since its parser
reads only whole seconds, and stored text with a fraction renders as an
invalid text input. A `time.Time` accepts them, because `time.Parse` tolerates a
fraction, but renders without it, so an unchanged save truncates. For now the
README and `FormTag.Step` document that the finest step is one second.

**Context:** `timeText.parse` and `shortest` in `formdatetime.go` read and
write the input forms; a `.999999999` layout variant would keep the fraction,
and `dateTimeInput` would format a `time.Time` with it. The `step="1"` rule in
`DateTimeInput` would then count sub-second remainders from `min` too.

**Effort:** S
**Priority:** P3
**Depends on:** None

### A time-of-day type whose zero is midnight cannot hold 00:00

**What:** Let a time-of-day type such as `civil.Time`, whose `IsZero` is true
at 00:00:00, show and store midnight in a time input.

**Why:** `DateTimeInput` renders a zero value as no value, so a stored
midnight shows an empty input, and the handler's required check reads the
bound zero value as missing, so a required time field rejects a posted 00:00.
An optional field saves midnight but shows it empty again.

**Context:** `marshaledText` in `formdatetime.go` honours the type's own
`IsZero`; `isEffectivelyEmpty` in `formhandler.go` decides required. A fix would
show a time-of-day value's text whenever it is not null and decide required
from the raw submission for date and time fields. The type cannot tell an
unset value from midnight itself, so either way it needs a pointer or a
nullable wrapper to mean "no time".

**Effort:** S
**Priority:** P4
**Depends on:** None

## mx dead code

### Remove or finish `reflectroutes.go`, `writereflect.go` and `TypedRoute.Path`

**What:** Three pieces of the root package are dead: `reflectroutes.go` (and
its `reflectroutes_test.go`) and `writereflect.go` are entirely commented out,
and `TypedRoute.Path` at `route.go:141` is a stub returning the raw pattern,
with its implementation commented out under a `TODO`.

**Why:** Dead code reads as API to a newcomer and to an agent searching the
repo, and `Path` is worse than dead: it is part of the `Route` interface, so a
caller building a URL with it silently gets `"/blog/{slug}"` back, placeholders
and all.

**Context:** Each has a replacement already. `ReflectRoutes` (route struct
fields registered by reflection) gave way to explicit `mx.Route`/`TypedRoute`
registration on `http.ServeMux` patterns and `web.Site.RegisterRoutes`.
`writereflect.go` (struct fields written as attributes) gave way to the typed
attribute constructors in `html`/`svg` and the `Attrib` interface; decide the
README "To Do" item `ReflectMarkup()` together with it, since that is the same
idea. For `Path`, either implement it with the existing `JoinAbsPath` and
`FormatPathValue` in `paths.go` (the commented-out body is most of it), or drop
`Path` from the `Route` interface, since sites build URLs with
`web.Site.URL`/`PageURL`.

**Effort:** S
**Priority:** P3
**Depends on:** None

## Completed

### Server-side `min`/`max` for `time.Time` date fields

**What:** Check `form:"min=…,max=…"` on the server when binding a `time.Time`
(or other non-text) date, datetime or time field, as `mx.ParseDateTimeInput`
already does for fields that hold the value as text.

**Why:** For a `time.Time` field the range only reaches the browser as the
input's `min`/`max`, so a tampered or scripted POST stores a date outside it.

**Context:** `ParseDateTimeInput` (`formdatetime.go`) parses the bounds with the
input type's layouts and compares times; `setTime` in `html/formdecider.go`
parses `time.Time` with its own layout list. Reusing the same bound check there
keeps both paths consistent. Left out of the string-kind date work because the
text path needed it (its fallback cannot enforce the range in the browser) and
the `time.Time` path was to stay unchanged.

**Effort:** S
**Priority:** P3
**Depends on:** None
**Completed:** 2026-10-19 — `setTime` checks the bounds with
`mx.CheckDateTimeRange` for `time.Time`, types defined as it, wrappers and date
structs.


### `datetime` and `time` widgets on string-kind fields

**What:** Make `form:"widget=datetime"` and `form:"widget=time"` prefill and
bind a field that holds its value as text (`string`, `*string`, a named string
type), the way `form:"widget=date"` does.

**Why:** A string field tagged `widget=datetime` or `widget=time` rendered an
empty input, and every non-empty submit failed with
`cannot assign time to <type>`, so such a field could not be edited at all.

**Context:** Settled formats: no zone, because these inputs are local times
and attaching or dropping a zone would change what the stored text means; and
seconds only when they are not zero, the form a `datetime-local` input uses
itself (`2006-01-02T15:04[:05]`, `15:04[:05]`). `mx.ParseDateTimeInput`
normalizes on bind, so a round trip does not depend on whether the browser
submits zero seconds. A type with its own spelling converts with `Format` and
`UnmarshalText`. Text the input cannot show, a time with a zone included,
falls back to a text input with the format as placeholder and hint.

**Effort:** S
**Priority:** P3
**Depends on:** None
**Completed:** 2026-10-19

### A tag-free opt-in for date types

**What:** Let a string-kind date type declare itself a date once, so its
fields render a date input without a `form:"widget=date"` tag on each of
them.

**Why:** A date type is a date wherever it appears, and a field that forgets
the tag renders as a plain text input, which accepts any spelling.

**Context:** `mx.FormWidgetHint` was the opt-in already but did not work for
dates: both renderers and the parser read only `FormTag.Widget`, so a type
returning `"date"` rendered `type="datetime-local"` and failed to bind.
`mx.DetectField` now returns the hint's widget as the tag's `Widget` when the
hint decides the kind, so every renderer and the parser read one value. That
also fixed a `"checkbox"` hint rendering a shadcn switch, an explicit
`widget=text` tag being overridden by an `"email"` hint, and a panic on a nil
pointer to a type with a value-receiver `FormWidget`.

**Effort:** S
**Priority:** P3
**Depends on:** None
**Completed:** 2026-10-19

### Enum-set checkboxes drop members missing from the option list

**What:** An enum-set field silently loses the checked members that are absent
from the per-request option list.

**Why:** Missing members are not rendered at all, and an unchecked box submits
nothing — so the next save removes them from the set. This is silent data loss
on an ordinary round-trip: the user never sees the values, never touches them,
and loses them by pressing Save.

**Context:** The per-request list is the trust boundary for the set. Members
it does not offer are not rendered and not echoed, and a save does not keep
them: the submitted checkboxes replace the set, so a filtered-out id cannot
be written back and an uncheck cannot be undone by the list changing before
POST. A value outside the list is still rejected. A static enum list is the
whole domain and replaces the set the same way.

**Effort:** M
**Priority:** P1
**Depends on:** None
**Completed:** 2026-10-16

### An empty provider list makes a select unsubmittable or clears it

**What:** An authoritative context provider that returns an empty list, combined
with a non-empty stored value, renders a select containing nothing but the
disabled placeholder.

**Why:** A required field can then never be submitted — the form is a dead end
the user cannot escape. A non-required field silently clears on save instead.
Either way the failure is invisible until someone tries to submit.

**Context:** An empty context-dependent list still renders. A stored select
value shows the generic disabled placeholder, a blank required select shows
"Choose…", and an enum set with members the list does not offer shows the
note and no checkboxes. A provider that failed returns its own error, as
before. The page stays usable; the field does not invent a value.

**Effort:** M
**Priority:** P1
**Depends on:** None
**Completed:** 2026-10-16

### A required select with an empty value submits its first option

**What:** A select whose current value is empty and whose option list has no
empty-valued option still lets the browser display and submit the first option.

**Why:** This is the new-record case, so it is the common one. The user submits
a value they never chose, and `required` is inert client-side because the
control is never actually empty — the browser has nothing to complain about.

**Context:** A select whose list has no empty option gets a leading empty
option: disabled and selected "Choose…" when the field is required and empty,
selected "(none)" when it is optional and empty, and an unselected "(none)"
when it is optional and already has a listed value so it can be cleared.
Required fields that already have a listed value get no extra option. The
out-of-list placeholder is unchanged. The same leading option applies to
static enum lists.

**Effort:** S
**Priority:** P1
**Depends on:** None
**Completed:** 2026-10-16

### `GlobPageSource.Dir` does not scope the glob

**What:** `Dir` is the base directory page paths are derived from, but the glob
itself still runs as a raw `filepath.Glob(Pattern)` against the process working
directory (`web/globpagesource.go`), so `Pattern` has to repeat the directory.

**Why:** Two spellings of the same directory that disagree yield either no
pages or pages whose `Page.Path` is relative to the wrong root, which lands in
the sitemap as wrong URLs. A single source of truth for the directory removes a
class of silent misconfiguration.

**Context:** Making `Dir` scope the match means resolving `Pattern` relative to
it (and deciding whether an absolute `Pattern` stays legal). The related
`content.IsDir()` branch in the same loop is still an empty `// TODO`: a
directory that matches the glob is silently skipped instead of being walked.

**Effort:** M
**Priority:** P3
**Depends on:** None
**Completed:** #31 (2026-08-25) — `Pattern` is matched below `Dir` and an
absolute `Pattern` with a `Dir` set is an error rather than a silent choice
between the two. A matching directory is now skipped explicitly; walking it
recursively was not added, `filepath.Match` has no `**`.

### SSE response type with a per-event flush loop

**What:** A `text/event-stream` response type that renders a `Component` per
event, flushes, and keeps the connection open. No flushing primitive exists in
go-mx today: `http.Flusher` and `Flush(` appear nowhere in the module.

**Why:** Without it there is no token streaming and no server-pushed message
append. `Component.Render(ctx, Writer) error` is pull-based and the context
already carries cancellation for a disconnected client, so nothing else about
the render path blocks streaming — this is the one load-bearing piece.

**Context:** The type owns the flush loop so callers cannot half-use it:
`NewSSEResponse(w, factory)` failing at handler entry if `w` cannot flush, then
`Send`/`SendError`/`Keepalive`/`Close`. Four details that are easy to get wrong:
`data:` is line-delimited, so rendered markup containing a newline must be
re-split into one `data:` line per line (it collides with `CheckedWriter.Newline`
and `WithIndent`); the headers must include `X-Accel-Buffering: no` because
on-premise deployments sit behind the customer's own reverse proxy; `Keepalive`
writes an SSE comment to defeat proxy idle timeouts; and `SendError` is the
in-band error contract — a reserved event name the client binds to, to be
documented together with `hx.EventSSEError` (`hx/events.go`), which is how htmx's
`sse` extension surfaces transport failures.

**Effort:** L
**Priority:** P0
**Depends on:** None
**Completed:** (2026-09-10) — `mx.SSEResponse` in `sse.go`. `NewSSEResponse`
probes the flush capability through the `Unwrap` chain *before* writing any
header, so a non-flushable writer is still answerable with a 500. `Send` renders
into a buffer and returns a render error before a byte of the frame is written,
and re-splits the output on CRLF/CR/LF into one `data:` line each. Frames are
written under a mutex so a `Keepalive` ticker cannot interleave with the
producer. `SendError` follows `RespondNonContextError`: generic message unless
`RevealInternalServerErrors`, silent on a context error, but with the ctx
cancellation stripped so the report still reaches a client that is reading.
Worked example in `cmd/example-sse`, verified in Chrome end to end.

### Content-derived stable element ids

**What:** A deterministic `id` `Attrib` derived from caller-supplied key parts,
alongside the existing counter-based `UniqueID`.

**Why:** `UniqueID()` (`uniqueid.go`) draws from a process-lifetime atomic
counter formatted in base 36. Two renders of the same entity produce different
ids, and a process restart restarts the sequence. Out-of-band swaps target by
id, so appending a token to "message N, part M" needs an id that is the same in
the initial render and in every later event. With only `UniqueID` available, a
streaming append cannot find its own target.

**Context:** Prefer a sanitising join over a hash — `_msg-<uuid>-3` is readable
in devtools where a hash is not, and debuggability is most of the value. Keep the
`_` prefix convention so the result is a valid HTML id that does not start with a
digit. Signature along the lines of `func KeyedID(parts ...any) Attrib`,
documented as "stable across renders and processes" in explicit contrast to
`UniqueID`.

**Effort:** S
**Priority:** P1
**Depends on:** None
**Completed:** (2026-09-10) — `mx.KeyedID` / `mx.KeyedIDValue` in `uniqueid.go`.
A sanitising join rather than a hash, so the id stays readable in devtools:
`KeyedID("msg", id, 3)` renders `_msg-<uuid>-3`. The plan's pure join was
amended on one point: reducing a character away makes distinct keys collide
(`"a.b"` and `"a/b"` become one id, and a collision is invisible — an
out-of-band swap just lands in the first match), so an FNV-1a digest of the
exact parts is appended *only when* a rune was reduced. Keys built from usable
characters alone keep the fully readable form the plan asked for.
`KeyedIDValue` was added beyond the sketch because the stated use case needs
the selector string (`"#"+KeyedIDValue(...)` as an `hx-swap-oob` target), and getting it
out of the `Attrib` otherwise means calling `AttribValue` with a context for a
value that has no context dependency.

### Typed `sse-connect` / `sse-swap` / `sse-close` attributes

**What:** Three typed attribute helpers for the htmx SSE extension.

**Why:** `hx/attributes.go` carries 30+ typed `hx-*` helpers and these are
absent, so every call site hand-writes `mx.NewAttrib("sse-connect", url)` — no
naming, no doc comment, no discoverability. The package already knows about the
extension: `hx/events.go` defines `EventSSEError` with a comment that htmx 2.0
moved SSE out of core.

**Context:** `sse-swap` takes one or more event names, so it should accept a
variadic and join on comma, mirroring how `SwapOOB` takes variadic selectors
(`attributes.go`). `hx.Ext("sse")` already loads the extension, so no new
machinery is needed beyond the attribs.

**Effort:** S
**Priority:** P1
**Depends on:** None
**Completed:** (2026-09-10) — `hx/sse.go`, documented in `hx/README.md`
alongside the two distinct error paths (`hx.EventSSEError` for a transport
failure, `mx.SSEEventError` for one the server reports in-band). `SSESwap` with
no event names defers an error instead of emitting a subscription to nothing.

### Scroll anchoring for `ScrollArea`

**What:** Stick-to-bottom behaviour for a scroll container whose content grows:
follow new content unless the user has scrolled up.

**Why:** It is the single most-noticed behaviour in a chat UI, and its absence is
noticed as a bug rather than a missing feature. `shadcn/scrollarea.go` is pure
classes (`scrollAreaClasses`) with no script.

**Context:** This needs **no new go-mx API**. `hx.OnHTMX` plus the existing
`hx.EventOOBAfterSwap` / `hx.EventAfterSettle` constants (`hx/events.go`) can
drive it, and `ScrollArea(attribsChildren ...any)` composes an extra attrib
naturally. The precedent for a small inline script shipped with a component is
`tabsSelectScript` (`shadcn/tabs.go`). Open question worth deciding before
writing it: library component or consumer-side recipe. A recipe is honest if it
stays ten lines; a `StickToBottom` attrib is better if the near-bottom threshold
needs tuning, because then every consumer would otherwise copy the same tuning.

**Effort:** S
**Priority:** P2
**Depends on:** SSE response type (there is nothing to anchor until content streams)
**Completed:** (2026-09-10) — `StickToBottom` / `StickToBottomThreshold` in
`shadcn/scrollarea.go`. The open question resolved to a library attrib: the
threshold does need tuning, so every consumer would otherwise copy the same
script, and it travels in the attribute value (`data-stick-to-bottom="96"`) so
one shared script serves every instance. It watches DOM mutations rather than
the htmx events the sketch suggested, which makes it work for an SSE stream, an
ordinary swap or any other script, and on a page with no htmx at all. Whether to
follow is recorded on scroll, not measured at mutation time — after a large
append every scroll position looks far from the bottom. Verified in Chrome
against a live `mx.SSEResponse` stream: pinned while content grew (gap 0 as
scrollHeight went 383→533), stayed put after scrolling up mid-stream (gap grew
255→405 instead of snapping back), and resumed following on scrolling back down.
