# Changelog

All notable changes to this project are documented in this file. The format
follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

Versions use the Go module scheme, `vMAJOR.MINOR.PATCH`, and a release tags
every module of the repo in lockstep: the root module as `vX.Y.Z` and each
submodule with its path as prefix (`fpdf/vX.Y.Z`, `wordpress/vX.Y.Z`,
`mcp/vX.Y.Z`).

This file starts with the entry below; the changes made before it are in the
commit history. Nothing is tagged yet, so everything here is unreleased and the
API is still free to change.

## [Unreleased]

### Added

- **A public knowledge layer for crawlers and AI agents.** A site publishes the
  same pages as HTML, Markdown, an llms.txt index, schema.org JSON-LD and MCP
  tools, and every one of those surfaces lists exactly the pages
  `web.Page.Indexable` lets a search engine see — one predicate for "public",
  so a draft cannot leak through a surface that forgot to check. Only the parts
  that need no knowledge of what a site sells live here; prices, offerings and
  revision policy stay with the site.
- **`web.Alternate` and `web.Page.Alternates`: translations and other formats
  of a page, from one list.** `Site.AddPageMetadata` renders them as `<link>`
  elements after the canonical link and `Site.PageHeaders` returns the same
  links as `Link` headers, so the HTML head and the response headers cannot
  drift apart.
  - A page with translations also names itself, from the new `Page.Lang`
    (falling back to `Site.Lang`): search engines ignore an hreflang set whose
    pages do not each list themselves. A translated page without any language
    is an error, not a set crawlers will discard.
  - `Site.DefaultLang` adds the `hreflang="x-default"` link, pointing at the
    version in that language, and only when that version exists. An explicit
    `x-default` alternate (a language chooser) takes precedence.
  - What would give crawlers two answers to one question is an error:
    `Page.Validate` rejects a repeated rel, type and hreflang, a translation
    pointing back at the page itself and a `text/markdown` alternate anywhere
    but at the page's `MarkdownPath` (or its own URL, for content
    negotiation), which llms.txt and the MCP tools link; building the links
    rejects a translation in the page's own language.
- **Markdown alternates of pages.** A page with Markdown content gets a
  Markdown version for readers that have no use for layout and scripts.
  `Site.MarkdownAlternate` puts a header in front of the page's own source —
  title, description, canonical URL, language, last update and translations —
  so the document makes sense without the HTML page around it, and leaves room
  for `MarkdownDocument.Sections` a site adds itself. The last update is the
  new `Page.LastModified` (`LastUpdated`, else `Published`), the date the
  sitemap uses too. `MarkdownPath` maps `/en/faq` to `/en/faq.md`, and
  `Site.MarkdownAlternateURL` returns the absolute URL of a page's alternate,
  failing for a page whose alternate could not be built, so llms.txt and the
  MCP tools never list one that is not there. `Site.ServeMarkdown` serves it
  with a canonical `Link` header pointing at the HTML page, `Vary: Accept`, and
  conditional requests answered by `http.ServeContent` — under a tag of the
  caller's revision marked `-md`, which never equals the HTML page's under
  content negotiation — and a 404 for a page without a Markdown source.
  `AcceptsMarkdown` decides from the `Accept` header whether a client asked
  for Markdown over HTML — a browser's `*/*` never counts as asking, while an
  agent's `text/markdown, text/html, */*` does, because at equal quality the
  client's order decides.
- **`web.Site.ServeMarkdownDocument`** serves a `MarkdownDocument` the caller
  built with `Site.MarkdownAlternate` and then extended, typically with
  `Sections` of facts only the site knows. Without it a site adding sections
  could not use `ServeMarkdown` and had to copy its response headers, a copy
  that drifts: a second `Vary: Accept` after a content-negotiating handler set
  one, or a `-md` tag that mishandles weak entity tags. `ServeMarkdown` now
  builds the document and calls it, so both write the same headers. A nil
  document is a caller bug and gets a 500.
- **`web.LLMsTxt`** renders an [llms.txt](https://llmstxt.org) index and, like
  `web.Robots`, validates first: a line break in any field, a bracket,
  backtick, `<` or trailing backslash in a link title, a parenthesis or
  backslash in a URL or a context opening a heading, quote, code or HTML block
  would add a heading, a link or a summary nobody wrote, or swallow the
  sections after it, so each is an error instead of a file. `HandleHTTP` serves it with an ETag
  derived from the content. `Site.LLMsTxt` builds one from the public pages
  with a Markdown alternate, grouped by `Page.Type`.
- **`html.JSONLD`** embeds structured data as a
  `<script type="application/ld+json">`. Unlike `html.ScriptJS` it is safe for
  untrusted values: `encoding/json` escapes `<`, `>` and `&` inside strings, so
  no `</script` can end the element early. A value that fails to marshal, or
  marshals to `null`, is a deferred render error.
- **`html.ValidLangTag`** exposes the BCP 47 charset check `Document.Lang` is
  rendered with, which `web` now validates `Page.Lang`, `Site.Lang` and
  hreflang values with instead of a copy of it.
- **`web/schemaorg`: a typed subset of schema.org** — `Person`,
  `Organization`, `WebSite`, `WebPage` and its `AboutPage`/`ContactPage`/
  `CollectionPage`/`FAQPage` subtypes, `ItemList`, `Service`, `Offer`,
  `Event`, `Place`, `VirtualLocation`, `PostalAddress`, `Question`/`Answer` —
  as plain structs that marshal to JSON-LD. `Price` formats integer cents as a
  decimal string and there is no `float64` in the package, so a price is never
  whatever its nearest float rounds back to. `EventStatus`,
  `EventAttendanceMode` and `Availability` are typed enums carrying the
  schema.org URLs.
- **`mcp` module: a site's public pages as MCP tools.** `AddPageTools` adds
  `search_pages`, `list_pages` and `get_page` — read-only, returning Markdown
  alternates of public pages only — to a Model Context Protocol server of the
  official Go SDK that the caller owns, next to the caller's own tools.
  Nothing is exposed without the caller passing a source, `get_page` answers a
  missing page and a non-public one alike, and server-side failures reach the
  client only as "unavailable" — a panic on the page path included, which the
  SDK would otherwise let crash the whole process. A `search_pages` query is
  limited to `MaxQueryBytes` and `MaxQueryWords`, because one call costs its
  words times the text of every page, a `language` filter has to be a
  language tag, and a cancelled call stops reading the source.
  `GuardHandler` adds what a public endpoint needs: an `Origin` allowlist, an
  optional `Host` allowlist against DNS rebinding, a 413 for oversized bodies,
  a 400 for JSON-RPC batches (which would carry hundreds of tool calls past
  every per-call bound; `AllowBatches` opts back in), and
  `no-store, no-transform`, which replaces the `no-cache` the SDK sets on
  its own responses. It is a nested module, so the SDK stays out of the core
  module.
- **`mx.SSEResponse`: components streamed to a client as Server-Sent Events.**
  go-mx buffers whole responses on purpose, so that a deferred computation
  failing mid-render becomes a clean 500 instead of a truncated page.
  `SSEResponse` keeps that guarantee at *event* granularity: `Send` renders into
  a buffer and reports a render error — or a panic — before a byte of the frame
  is written. Only the response-level contract changes, and it changes in one
  documented place: after the first flush the 200 is committed, so
  `SendError` carries a later failure in band as the `mx.SSEEventError` event.
  - Rendered markup is re-split into one `data:` line per line of output, on all
    three SSE line terminators, because a client stops a `data:` field at the
    first newline and `Writer.Newline` puts newlines in ordinary markup.
  - `NewSSEResponse` fails at handler entry if the writer cannot flush, while a
    500 is still available to send, and sets `X-Accel-Buffering: no` so a reverse
    proxy the caller does not control cannot buffer the stream into uselessness.
  - `SendEvent` takes an `SSEEvent` with an `ID`, and `LastEventID` reads back
    what the client acknowledged. A browser silently reconnects to any stream
    that ends, so without these a dropped connection replays the whole stream or
    skips what it missed. `SetRetry` tunes the reconnect delay and
    `KeepaliveLoop` keeps an idle connection from being dropped at all.
  - `SetWriteTimeout` bounds a single frame. The response-wide
    `http.Server.WriteTimeout` has to be cleared for a stream that never ends,
    which would otherwise leave a write with no bound: a client that stops
    reading would pin the producer goroutine forever.
- **`mx.KeyedID` / `mx.KeyedIDValue`** derive an element id deterministically
  from key parts, so the id is the same in the initial render and in every later
  event — which `mx.UniqueID`'s process-lifetime counter cannot do, and which an
  out-of-band swap needs to find its own target. A sanitising join keeps the id
  readable (`_msg-<uuid>-3`); when a character has to be reduced away, a digest
  of the exact parts is appended, so keys that differ only in reduced characters
  stay distinct instead of colliding silently. Part boundaries are deliberately
  outside the digest — `KeyedIDValue("a-b")` and `KeyedIDValue("a", "b")` are
  still one id — because keeping a literal `-` readable is worth more than a
  case only a caller mixing both spellings for one entity can reach.
  `mx.ValidIDRune` is the shared definition of which characters an id may hold.
- **`hx.SSEConnect` / `hx.SSESwap` / `hx.SSEClose`** for the htmx SSE extension,
  with `hx.ScriptSSEFromCDN` to load it (htmx 2.0 moved SSE out of core) and
  `hx.EventSSEOpen` / `EventSSEClose` / `EventSSEBeforeMessage` /
  `EventSSEMessage` alongside the existing `EventSSEError`. Each attribute
  defers an error for an empty value, because the extension skips a falsy
  attribute and the result looks exactly like a server that never sends.
- **`shadcn.StickToBottom` / `StickToBottomThreshold`** make a `ScrollArea`
  follow content that grows after the page was delivered, and leave a user who
  has scrolled up where they are. Following survives layout-only growth (an
  image or webfont loading, a container resize, a class or style change), holds
  position when a scroll event has not been dispatched yet, and marks the
  element `data-stuck` so a "jump to latest" affordance is pure CSS.
- **`mx.ContentTypeEventStream`** for `text/event-stream`, the content type
  `SSEResponse` sends. It is the one `text/*` constant carrying no `charset`
  parameter: the SSE specification fixes the stream at UTF-8 and requires
  clients to ignore the parameter.
- **`cmd/example-sse`** streams a chat transcript against all of the above:
  `-drop` cuts the connection mid-reply so the browser reconnects and the
  handler resumes, and `-fail` reports an error in band.
- **`logview` package: streamed log lines rendered as a readable surface.**
  Logs are mostly structured JSON, and a raw JSON line is unreadable at stream
  speed. `logview.Line` parses one and renders its fields in source order: the
  timestamp and level promoted without labels, everything else as `key=value`
  with the value colored by its JSON type, nested objects in braces, arrays in
  brackets, and a multi-line string value in a `<pre>` so a stack trace keeps
  its indentation. Anything that is not one JSON object renders as plain text
  with its leading level word colored, so a mixed stream still reads as one log.
  - **A level value can never reach a class attribute.** A space would end the
    class token and let the rest name arbitrary CSS classes — `hidden` alone
    would make the line vanish. Only tokens the `Theme` defines are emitted, and
    every other value renders as `logview.LevelUndefined`; `Theme.Levels` keys
    are restricted to `mx.ValidIDRune` characters so a careless theme cannot
    open the hole either. The level text still shows the raw value, escaped.
  - **Lines render through their own non-indenting writer,** so the markup is
    the same whether the caller's writer indents or not and never ends in a
    newline — which would otherwise reach the client as an extra `data:` line
    and become a stray text node between every pair of log lines.
  - `MaxLineLen`, `MaxDepth` and `MaxLines` bound what a log with no upper bound
    can do: a runaway line, hostile JSON nesting (unbounded recursion there
    would overflow the goroutine stack, which no recover can catch), and the
    number of lines the browser keeps.
- **`logview.View`** is the surface those lines stream into: a filter over the
  lines already received, a pause toggle, `shadcn.ScrollArea` with
  `StickToBottom` following the stream, and an `mx.SSEEventError` sink. One
  inline script drives all of it through fixed `data-mx-log-*` attributes,
  independent of the class prefix.
  - **Pausing keeps arriving lines in the DOM** and marks them rather than
    detaching them, so order, out-of-band targeting and htmx's settle step keep
    working. Pause and filter get one attribute each, because resuming must not
    reveal a filtered-out line — which is also why the "12 new" badge counts
    only the held lines the filter shows. The held lines get a `MaxLines` cap of
    their own, so a view left paused under a firehose stays bounded, at up to
    twice `MaxLines`: what is already on screen is what the reader is reading,
    and is not trimmed under them.
  - **Scrolling up pauses as well, and scrolling back down resumes.** Scrolling
    back to read something and wanting the stream to hold still are one intent,
    not two. Which one paused is remembered, because it decides what resumes: a
    pause the scroll caused is undone by returning to the bottom, while one the
    button caused stays until the button — whose caption swaps to
    `Labels.Resume` while paused — takes it back. The threshold is at most
    `shadcn.StickToBottomDefaultThresholdPx` rather than a number of its own, so
    scrolling can only pause once the scroll area has already stopped following.
  - **The paused state is announced, not just shown.** A caption change on a
    button that does not have focus reaches no screen reader, and the scroll
    entry point never touches the button, so the state also goes into a
    `role="status"` region carrying `Labels.Paused`. The caption names the next
    action instead of carrying `aria-pressed`, which paused would announce
    "Resume, pressed" — the opposite of what it reads; a stylesheet keys off
    `data-mx-log-paused` on the root instead, and the button is sized by the
    caption it swaps to rather than by an em value measured in English.
  - **The filter highlights what it matched**, as a CSS custom highlight styled
    by the `::highlight()` rule `Theme.CSS` emits for `ClassMatch`. Ranges
    rather than wrapper elements, because a line is a tree of spans that htmx
    swapped in and two MutationObservers are watching — and because one range
    can span the several spans a field is split into, which `status=200` is.
  - **The level is a fixed column,** `min-width` in `ch` as wide as the longest
    level the theme defines, so severity can be scanned straight down the log.
    `LevelUndefined` is excluded, since an unknown level displays its raw value
    rather than that key; such a value runs past the column rather than being
    cut off.
  - **The line cap only trims while the view is at the bottom,** since trimming
    under a reader who has scrolled up drags the text they are reading upward.
    The position is read live at trim time rather than remembered from the last
    scroll event: hiding or revealing lines moves the bottom without dispatching
    one, so clearing a filter would otherwise leave the view trimming and
    stick-to-bottom pinning as if the reader were still following.
  - **The view is a flex column,** so `Config.Height` is what the scroll area
    prefers rather than a fixed size: a page that gives the view a height of its
    own has the area fill whatever the toolbar and error sink leave.
- **`logview.Theme`** gives every value type and every log level a foreground
  and background color, bold, italic, underline and strikethrough, plus an
  optional image rendered in place of a level's text, with the level value as
  its alt text. `DarkTheme` and `LightTheme` ship; `Theme.CSS(prefix)` is
  deterministic.
- **`cmd/example-logstream`** simulates a production log against all of the
  above: randomly composed lines covering every rendering path, streamed
  indefinitely at exponentially distributed intervals with bursts and quiet
  stretches, resumable from a bounded history, with `-rate` and `-fail`. Its
  Split control puts a second viewer beside the first, each with its own
  connection to the same stream, so that filtering or pausing one while the
  other keeps streaming shows what "independent subscriber" means.

- **`web` package: robots.txt, sitemaps and page metadata for a whole site.**
  A `Site` holds what all pages share — the `BaseURL` every absolute URL is
  built from, the title, the language — and turns its `PageSource`s into the
  files search engines expect:
  - `Site.Sitemap` builds a sitemap of the pages that are meant to be found
    (published, not scheduled, not marked `NoIndex`), sorted by URL so a
    regenerated sitemap only differs where the site did.
  - `Site.RobotsTxt` serves a robots.txt pointing at that sitemap, and
    `Site.RegisterRoutes` registers both on an `http.ServeMux` at the paths the
    robots.txt advertises.
  - `Site.AddPageMetadata` gives a rendered page its canonical link,
    description, author and Open Graph tags, and marks drafts, scheduled pages
    and `NoIndex` pages `noindex, nofollow` so they cannot reach a search index.
- **`web.Robots`** builds a robots.txt from rule groups and sitemap URLs, with
  `AllowAllRobots` and `DisallowAllRobots` for the two cases most sites need.
  It writes through `encoding.TextMarshaler`, `io.WriterTo` or an
  `http.HandlerFunc`, and validates first: a rule that would not mean what it
  says once written — a smuggled line break, a path that is not anchored, a `#`
  that truncates the rule at a comment — is an error instead of a file.
- **`web.Sitemap` and `web.SitemapIndex`** render the sitemaps.org 0.9 protocol
  through the `xml` package. Both are `mx.Component`s, so the same value can be
  written to a file for a statically generated site or served over HTTP, and
  both enforce the protocol's limits (50000 URLs, 50 MiB, 2048-byte `<loc>`,
  the `changefreq` keyword set, the priority range) rather than only documenting
  them.
- **`web.Page`** gains `Path` (the URL a canonical link and a sitemap entry are
  built from), `Description` (the meta description), and `IsPublished`/
  `Indexable`, which put the definition of "should search engines see this" in
  one place.
- **`web.GlobPageSource`** derives `Page.Path` from the file path, so a
  directory of markdown files produces a working sitemap without any extra
  wiring, and reads a `description` from the front matter. Files without front
  matter (HTML, plain text) are dated by their modification time so they can be
  indexable at all.
- **`shadcn.Calendar` can be picked from, which makes the DatePicker recipe work.** `CalendarOptions.Name` turns on select mode, react-day-picker's single mode re-expressed with one small shared script: a hidden input of that name carries the picked day as `2006-01-02`, a click or Enter picks a day and fires `input` and `change` there, the arrow keys and Home/End move the focus between enabled days with a single Tab stop, and a surrounding popover closes on the pick. The new `shadcn.CalendarText(name, …)` shows the picked day's `DayLabel`, as the trigger of a date picker does; the gallery's DatePicker uses it and starts empty with "Pick a date" like upstream. `CalendarOptions.Today` marks today with the accent background and `aria-current="date"`, and `StartMonth`/`EndMonth` disable the month buttons at the limits; like `Disabled`, they only shape the grid, so a handler must check a submitted day itself. In select mode the calendar is a `role="grid"` table named by its month whose selected gridcell carries `aria-selected` (a button may not; the button's `data-selected` keeps the look, and the static calendar stays a plain table whose selected day's button is `aria-pressed`), the month buttons are labelled from the new `Labels.CalendarPrevious`/`CalendarNext`, day buttons show a focus ring, disabled days no longer light up on hover, and days of the adjacent months are muted but no longer faded like disabled days. Month navigation is still a server round-trip; `shadcn/TODOS.md` lists what react-day-picker does beyond this.

### Changed

- **go-mx requires Go 1.27 and no longer depends on
  `github.com/domonda/go-types`.** The `doc/dbdoc` record types use the
  standard library `uuid.UUID` for their ids and `*uuid.UUID` for the nullable
  `Text.FontID`. Unlike the go-types ids, `uuid.UUID` has no `Scan` and `Value`
  methods: the default converter of `database/sql` stores it as its string form
  and scans it from a string or 16 bytes, but a driver that converts arguments
  itself has to support the type.
  `Document.Name` is a `string`, and the nullable `DocVersion.Description` and
  `DocVersion.CreatedBy` are `*string`, so they are no longer trimmed the way
  the go-types string types trimmed them. Only a nil pointer is NULL now: an
  empty string or the Nil UUID is stored as such, where the go-types nullable
  types stored it as NULL.
- **`web.GlobPageSource` no longer includes the front matter in the content of
  a Markdown page.** It is parsed into the page fields, and passing it along as
  well put it into every rendering of the page — and into the Markdown
  alternate, where a reader would get the YAML block as text. The front matter
  is also read for a `lang`, the new `Page.Lang`. A locale like `en_US`, as
  Jekyll writes it, becomes the language tag `en-US`; anything else that is
  no language tag fails the source with the name of the file, as front matter
  that cannot be parsed now does too. A `ContentRenderer` that parsed the
  front matter itself for fields of its own no longer gets it: only `title`,
  `description`, `lang`, `date`, `publishDate`, `draft` and `tags` are read.
- `Site.AddPageMetadata` sets the html `lang` attribute from `Page.Lang` before
  `Site.Lang` (the new `Site.PageLang`), and fails for a page that does not
  pass `Page.Validate` even when the renderer declared the page URL itself, so
  a page with a Path needs a `Site.BaseURL` in every case. A
  renderer declaring it has to declare the canonical URL of the page, because
  `Site.PageHeaders` sends that one: a document whose `og:url` names another
  URL is an error rather than a response with two canonicals. `Site.Validate`
  checks `Lang` and `DefaultLang` too, and `Site.Sitemap` fails for a public
  page whose links cannot be built, instead of listing a URL that answers
  with an error.

- **`mx.SSEEventError` is `"mx-error"`, not `"error"`.** A browser dispatches
  its own transport failures at the `EventSource` under the name `error`, so a
  subscriber to `error` would also fire on every dropped connection, with an
  event carrying no data for htmx to swap. Both names are unreleased — the
  rename happened within this cycle, so there is nothing to migrate; it is
  recorded because anyone naming their own SSE events walks into the same
  collision.

- `web.GlobPageSource.Dir` now scopes the glob: `Pattern` is matched below it
  instead of against the working directory, so the directory is named once
  rather than repeated in both fields where the two could disagree. A `Pattern`
  that is absolute while `Dir` is set is an error, and a directory matching the
  pattern is skipped instead of being read as a page.

- **htmx is served from jsDelivr, not unpkg.** `hx.ScriptFromCDN`,
  `hx.ScriptDebugFromCDN` and `hx.ScriptSSEFromCDN` now point at
  `cdn.jsdelivr.net`. The Subresource Integrity hashes are unchanged and were
  verified against the jsDelivr bytes. A Content-Security-Policy `script-src`
  allowlist that names `unpkg.com` has to name `cdn.jsdelivr.net` instead.

- **Closed sets of values are go-enum types, checked with the generated
  methods instead of hand-written lists.** An unknown value now fails with the
  generated `Validate` error (`invalid value "x" for type pkg.Type`) or
  resolves to the documented default with `ValidOr`, the same way everywhere.
  The build-time generator is pinned to go-enum v1.3.0 and all enum methods
  have been regenerated:
  - `pdf.Renderer.SetLineCapStyle` and `SetLineJoinStyle` take, and their
    getters return, `pdf.LineCapStyle` and `pdf.LineJoinStyle` instead of a
    string. An unknown style is the renderer error rather than a silent butt
    cap or miter join, so `pdf.LineCap` and `pdf.LineJoin` fail the render
    for one. The errors for an invalid unit, page size, orientation, blend
    mode or attachment `AFRelationship` are the generated ones too.
  - `hx.SetReswap` takes a `hx.SwapStyle` plus modifiers like `hx.Swap`, and
    returns the `Validate` error of an invalid style instead of sending it.
  - `web.Alternate.Rel` is the new `web.AlternateRel` type.
  - `shadcn.Button` and `shadcn.InputGroupButton` emit the default in
    `data-variant` and `data-size` for an unknown value, like every other
    component, instead of echoing it next to the default classes.
    `shadcn.SheetSide` and `shadcn.ResizeDirection` get the go-enum methods.
  - `logview.Class` gets the go-enum methods, and `Theme.CSS` ignores a
    `Styles` key that is none of the classes, as no markup carries it.
  - `wordpress.PermalinkStyle` gets the go-enum methods. `WriteStatic`
    rejects an unknown `Options.Permalinks` instead of writing slug URLs, and
    `wordpress-import -permalinks` fails on one.

### Removed

- **`hx.EventNoSSESourceError`.** htmx's `sse` extension never raises
  `htmx:noSSESourceError` — it does not check nesting at all, so an `sse-swap`
  element with no `sse-connect` above it subscribes to nothing silently. The
  constant named an event that cannot occur.

### Fixed

- **The date, datetime and time widgets work on fields that hold the value as text.** A `string`, `*string` or named string type (a date type with an underlying string) tagged `form:"widget=date"`, `widget=datetime` or `widget=time` rendered an empty input, so editing a saved record showed no value, and every non-empty submit failed with `cannot assign time to <type>`. Now, in both the `html` and the `shadcn` renderer, the input shows the stored value: a `Format(string) string` method is asked for it, and otherwise text in the input's own form is shown. That form has no zone and carries seconds only when they are not zero: `2006-01-02`, `2006-01-02T15:04[:05]` and `15:04[:05]`. A submitted value is stored in that same form, through `UnmarshalText` when the type has it, so zero seconds are dropped and other seconds are kept. An empty submission stores the zero value, which a required field still rejects, and `form:"min=…,max=…"` is checked on the server. Stored text the input cannot show, such as another spelling (`06.07.2026`), an invalid value or a time with a zone, renders in a `type="text"` input with the same name instead of a blank input that would clear it on the next save. The text input is marked `aria-invalid` from the start (unless the field is readonly, so never saved) and has the format to enter (`YYYY-MM-DD`, `YYYY-MM-DDTHH:MM` or `HH:MM`) as its placeholder and as a hint below it, linked with `aria-describedby`; shadcn takes the hint's caption from the new `Labels.FormExpectedFormat`. Posting such text back fails with a field error naming the format, so nothing is saved, and the form keeps showing the posted text, also in a repeatable row after an add-row or remove-row click. A `Format` method that answers invalid text with `""` gets the text input too, and a string type without `Format` shows its `MarshalText`. A value that is not a whole number of minutes after the field's `min` gets `step="1"`, so the browser accepts it unchanged; a `form:"step=…"` tag sets the step itself. A GET followed by an unchanged POST stores exactly the value the GET showed. The new `mx.DateTimeInput` and `mx.ParseDateTimeInput` make these decisions for both renderers, and `time.Time` fields render and bind as before, apart from `form:"sensitive"` below.
- **Date structs and types that wrap `time.Time` keep their value in a date or time input.** A type such as `civil.Date` that declares the date widget rendered a blank date input, so an unchanged save stored the zero value. Its `MarshalText` now fills the input under the same rules as text, and its zero or null value renders none. Types that wrap `time.Time`, a type defined as `time.Time` or a struct embedding it (whose `MarshalText` is RFC 3339), render and bind like a `time.Time`: the defined type takes the parsed time converted, and the embedded time is set through its `UnmarshalText` with RFC 3339 when the type does not read the input's own form, but only when the time (or a pointer to it) is the struct's only field, so a struct holding a validity flag is never half set. A time-of-day or date-and-time struct whose text needs seconds, such as `civil.Time`, gets the submitted value with seconds, since the inputs leave zero seconds out. A nil embedded time pointer renders an empty input instead of panicking. A date struct reads its own text back first, so a struct whose text the picker cannot show renders it in a text input that is not marked invalid, and an unchanged save keeps the value, also in a repeatable row after an add-row or remove-row click. A `min` or `max` on such a field is checked on the value as the input shows it; text the input can only show as text cannot be checked, so with a bound set it is rejected with the format to enter and its input is marked invalid. A date type that is no struct, such as an integer day count with text methods, binds through its `UnmarshalText` too, instead of failing with `cannot assign time`. A date or time field bound through a pointer gets a new pointee, so a rejected value never reaches a record a shallow-copying `load` shares with a store. A registered single-value struct that holds no time, such as a money type, is now a text input through `MarshalText`/`UnmarshalText` instead of a date and time field, and an untagged type defined as `time.Time` is detected as one.
- **A `form:"sensitive"` date or time field no longer shows its value.** The date, datetime and time inputs echoed a sensitive `time.Time`, and the new text path would have echoed the stored text. Like a sensitive text input, the field now renders empty, so an unchanged save clears it, and its validation errors no longer quote the posted value.
- **A `FormWidgetHint` now decides the widget the way a tag does.** `mx.DetectField` returns the hint's widget as the tag's `Widget`, so renderers and the parser read one value. A type whose `FormWidget()` returns `"date"` rendered `type="datetime-local"` and failed to bind; it now renders and binds like a field tagged `form:"widget=date"`, which lets a date type declare itself once instead of every field carrying the tag. A `"skip"` hint now skips the field like `form:"widget=skip"` instead of rendering and binding it, a `"checkbox"` hint renders a checkbox in `shadcn` instead of a switch, and an explicit tag such as `form:"widget=text"` is no longer overridden by an `"email"` hint. A nil pointer field whose element type has a value-receiver `FormWidget` no longer panics `DetectField`.
- **The `shadcn` date and time inputs render `min` and `max`.** A `form:"min=…,max=…"` tag reached the `html` date input but was dropped by `shadcn`; both now render it on a date, datetime or time input. The text fallback leaves it out, since a text input cannot enforce it, and the server now checks the range for every date and time field, `time.Time` included, so a crafted POST cannot store a date outside it (`mx.CheckDateTimeRange`). The check compares the wall clock, as the bounds have no zone, so an RFC 3339 offset cannot move a date across a bound, and it allows a time range that wraps past midnight (`min=22:00,max=06:00`) as the browser does. A value outside the range is still bound, so the re-render and a repeatable row keep it after an add-row click, but the form is not saved. A value that is not whole minutes after the `min` gets `step="1"`.
- **A required reflected-form field no longer accepts a blank ID or date.** An empty submission stores the zero value of a `uuid.UUID`, any other fixed-size array ID, or a `time.Time`, and the required check now reads that zero value as missing. A hidden or text ID field tagged `form:"required"` rejects a blank instead of saving `uuid.Nil()`, and a required date rejects a blank instead of saving 0001-01-01. A zero ID also renders as no value, in both the `html` and the `shadcn` renderer: a select leads with "Choose…" or "(none)" instead of the "(current value not available)" placeholder, and an input stays empty instead of showing `00000000-0000-0000-0000-000000000000`.
- **Reflected selects and enum-sets no longer invent a value when the option list and the stored value disagree.** A required select with no current value leads with a disabled "Choose…" option, and an optional one with "(none)", so the browser does not submit the first real option. A stored value the list does not offer, including a sensitive one, uses that same generic placeholder. A context-dependent list that comes back empty still renders the form. An enum-set shows a note when stored members are absent from the list and does not keep those members on save; a submitted value outside the list is still rejected. A blank optional enum saves, and a decider `Validate` on that blank value still runs. A required select rejects an empty submission even when the zero value is not "empty" to reflection (a `[4]byte` id). Enum-set membership uses the same text form as the option list — `MarshalText`, otherwise `fmt.Sprint` — so a `Stringer` stays checked and a named string whose text form adds a prefix round-trips instead of growing that prefix on every save.

- PDF output intent subtypes escape spaces, delimiters, and other special
  bytes as PDF names, preserving custom extension values without malformed markup.
  The `OutputConditionIdentifier` and `Info` strings of an output intent are
  escaped too, so a parenthesis or backslash no longer ends the string early,
  and a protected document encrypts them like its other strings.

- **Every string of a protected PDF decrypts, not only the first of each
  object.** The engine, like upstream fpdf, kept one RC4 cipher per object
  number, so the second string of an object continued the first one's
  keystream, while a reader decrypts each string on its own: the `Title`,
  `Author` and other metadata after `Producer` came out as garbage, as did an
  embedded file stream after its `/ModDate`. Each string and stream now starts
  a new keystream from the object key.

- **A reflected form validates every element of an enum set.** For a
  `[]T` or `map[T]struct{}` field the validation chain only probed the slice
  or map, which has no `Validate`, so a tampered POST could bind any string
  as an element. Each element now runs through the chain of `T`, rejecting a
  value that is none of the enum's values like a single enum field does.
- `mx.DetectField` no longer takes a type whose `Enums` or `EnumStrings`
  method has parameters for an enum: the options could not be collected, so
  it rendered an empty select.
- The committed go-enum methods of `fpdf` were out of date. CI now runs
  `go-enum -validate` on every module.

- `web.DefaultRenderPage` no longer panics when a page is marked `NoIndex`, and
  now marks drafts and scheduled pages `noindex, nofollow` too — previously
  only an explicit `NoIndex` was honored, so a draft served from a preview
  deployment could be indexed.

[Unreleased]: https://github.com/ungerik/go-mx/commits/main
