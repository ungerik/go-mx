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
- **`shadcn.Calendar` can be picked from, which makes the DatePicker recipe work.** `CalendarOptions.Name` turns on select mode, react-day-picker's single mode re-expressed with one small shared script: a hidden input of that name carries the picked day as `2006-01-02`, a click or Enter picks a day and fires `input` and `change` there, the arrow keys and Home/End move the focus between enabled days with a single Tab stop, and a surrounding popover closes on the pick. The new `shadcn.CalendarText(name, …)` shows the picked day's `DayLabel`, as the trigger of a date picker does; the gallery's DatePicker uses it and starts empty with "Pick a date" like upstream. `CalendarOptions.WeekdayLabel` gives each weekday heading its full name as `aria-label` (English by default, but none when only `WeekdayName` localizes the headings), so a screen reader moving through the grid announces "Sunday" rather than "Su". `CalendarOptions.Today` marks today with the accent background and `aria-current="date"`, and `StartMonth`/`EndMonth` disable the month buttons at the limits; like `Disabled`, they only shape the grid, so a handler must check a submitted day itself. In select mode the calendar is a `role="grid"` table named by its month whose selected gridcell carries `aria-selected` (a button may not), and the month buttons are labelled from the new `Labels.CalendarPrevious`/`CalendarNext`. Month navigation is still a server round-trip; `shadcn/TODOS.md` lists what react-day-picker does beyond this.

- **Errors the user can fix are recognized by their method set, and the
  errors go-mx hides reach the application.**
  - `mx.UserMessageError` is any error with a `UserMessage() string` method.
    Wherever go-mx hides the text of other errors it shows the `UserMessage` of
    one found anywhere in the chain (as `errors.As` finds it, also through
    `errors.Join`), never the `Error()` text around it: the form-level message
    of a reflected form for an `onSubmit` error or an error no input shows, a
    `form:"sensitive"` field's validation and parse errors (also of a date or
    time field, which otherwise shows the format to enter), and a context
    option list that fails while a POST is read. A field of the `html` and
    `shadcn` deciders and a repeatable field's own box show an error's
    `UserMessage` too, and each message of a joined error, also a wrapped one,
    as an error of its own (a `form:"sensitive"` field shows the first). A
    field's error quotes at most 256 bytes of the posted text, so a forged
    value of any size reaches neither the page nor the hook whole. A
    domain package marks its user-fixable errors without importing go-mx;
    `mx.FormMessage` gains the method and works as before. An empty
    `UserMessage` counts as none. `mx.UserMessage(err)` is the test,
    `mx.FormErrorText(ctx, err)` the text a form shows for an error it cannot
    show at a field, and `mx.FieldErrorTexts(err)` the messages a field shows, for
    a decider of the application.
  - `mx.ReportError`, a `func(ctx context.Context, err error)` that is nil by
    default, gets every error go-mx answers with a generic text instead: the
    500s of `mx.RespondNonContextError` (so also the render errors of
    `mx.ComponentHTTPHandler`, `mx.ComponentFuncHandler`, the reflected forms
    and the `html`, `xml` and `web` handlers, and the panics the first two
    recover), a reflected
    form's load error, `SSEResponse.SendError`, an `onSubmit` error or an error
    no input shows that the form answers with `mx.FormErrorMessage`, a
    sensitive field's "not a valid value", a failing context option list of
    `html.FieldDecider`, and the 500 of `pdf.Document.ServeHTTP`. Each such
    error arrives once, also two that read alike and share one message above
    the form, whether or not `mx.RevealInternalServerErrors` is set, with the
    request's context stripped of its cancellation; a recovered panic carries
    its debug stack. An error the user sees is not reported, nor is a
    `context.Canceled` or `context.DeadlineExceeded`. A sensitive field's error
    may quote the value the form keeps off the page, so a hook that logs must
    keep it as secret as that value. `mx.ReportHiddenError(ctx, err)` reports
    through the same rules, for a decider or handler of the application that
    hides an error.

### Changed

- **`mx.RespondNonContextError` takes the request: `RespondNonContextError(w, r,
  err)`.** It now passes the error to `mx.ReportError`, which needs the
  request's context, and the former `(w, err)` had none to give. A sibling
  function taking the request would have left two ways to answer the same 500,
  one of them reporting nothing, and nothing is tagged yet; a caller in a
  handler has the request at hand. Pass it as the second argument.
- **A reflected form reads a multipart body up to 32 MB.** Every reflected form accepted a multipart body of any size, also one without a file field, and the parser writes file parts past `MaxMemory` to disk, so a single request could fill it. A larger body now answers 413. The new `ReflectFormConfig.MaxUploadSize` sets the bound per handler, and `mx.DefaultMaxUploadSize` sets it for all; a negative value removes it, for a handler behind a limit of its own (a larger `http.MaxBytesReader` limit is otherwise cut to it). Urlencoded bodies keep net/http's 10 MB limit, or that of an `http.MaxBytesReader`.
- **The static `shadcn.Calendar` marks the selected day differently.** The selected day's button no longer carries `aria-selected`, which ARIA does not allow on a button: it is `aria-pressed`, and the look moved from the `aria-selected:` class variants to `data-selected="true"` with `data-[selected=true]:` variants, so styles or tests keyed to `aria-selected` on the button need updating. Day buttons show a focus ring, disabled days no longer light up on hover, and days of the adjacent months are muted but no longer faded like disabled days.
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
- **A reflected form posts urlencoded unless it has a file field.** It was
  always `multipart/form-data`, whose limit is 1,000 parts, and every field
  takes two (its value and `__present` sentinel), so 100 rows of 5 cells
  were rejected with 413 and the record could no longer be edited. A form
  without a file input now posts `application/x-www-form-urlencoded`, which
  allows 10,000 values. A decider that renders its own upload widget for a
  field that is no file field sets the new `FieldBehavior.Multipart` to get
  a multipart form, and a file field posted urlencoded with a file name is
  a field error instead of a save that silently drops the upload.
- **A reflected form no longer shows the text of an internal `onSubmit`
  error.** An error that is neither `mx.FieldErrors` nor the new
  `mx.FormMessage` was shown above the form as its `Error()` text, which
  could carry SQL, addresses or a go-errs call stack. The form now shows
  `mx.FormErrorMessage` ("The form could not be saved.", reassignable)
  unless `mx.RevealInternalServerErrors` is set. Return a `mx.FormMessage`, or
  any error with a `UserMessage() string` method, also wrapped, for a message
  meant for the user.

### Removed

- **`hx.EventNoSSESourceError`.** htmx's `sse` extension never raises
  `htmx:noSSESourceError` — it does not check nesting at all, so an `sse-swap`
  element with no `sse-connect` above it subscribes to nothing silently. The
  constant named an event that cannot occur.

### Fixed

- **The date, datetime and time widgets work on fields that hold the value as text.** A `string`, `*string` or named string type (a date type with an underlying string) tagged `form:"widget=date"`, `widget=datetime` or `widget=time` rendered an empty input, so editing a saved record showed no value, and every non-empty submit failed with `cannot assign time to <type>`. Now, in both the `html` and the `shadcn` renderer, the input shows the stored value: a `Format(string) string` method is asked for it, and otherwise text in the input's own form is shown. That form has no zone and carries seconds only when they are not zero: `2006-01-02`, `2006-01-02T15:04[:05]` and `15:04[:05]`. A submitted value is stored in that same form, through `UnmarshalText` when the type has it, so zero seconds are dropped and other seconds are kept. An empty submission stores the zero value, which a required field still rejects, and `form:"min=…,max=…"` is checked on the server. Stored text the input cannot show, such as another spelling (`06.07.2026`), an invalid value or a time with a zone, renders in a `type="text"` input with the same name instead of a blank input that would clear it on the next save. The text input is marked `aria-invalid` from the start, in `shadcn` together with its `Field` (unless the field is readonly, so never saved), and has the format to enter (`YYYY-MM-DD`, `YYYY-MM-DDTHH:MM` or `HH:MM`) as its placeholder (unless a `form:"placeholder=…"` sets one) and in a hint below it, linked with `aria-describedby`, which reads "Not in the expected format" when the stored text is invalid (until a submit's field error says so), so not only the color says so; shadcn takes the hint's captions from the new `Labels.FormExpectedFormat` and `Labels.FormInvalidFormat`. Posting such text back fails with a field error naming the format, so nothing is saved, and the form keeps showing the posted text, also in a repeatable row after an add-row or remove-row click. A `Format` method that answers invalid text with `""` gets the text input too, and a string type without `Format` shows its `MarshalText`. A value that is not a whole number of minutes after the field's `min` gets `step="1"`, so the browser accepts it unchanged; a `form:"step=…"` tag sets the step itself. A GET followed by an unchanged POST stores exactly the value the GET showed. The new `mx.DateTimeInput` and `mx.ParseDateTimeInput` make these decisions for both renderers (with `mx.DateTimeInputLayout` and `mx.DateTimeFormatError` for a custom one), and `time.Time` fields render and bind as before, apart from the sensitive masking, server-side range, `step` and pointer changes below.
- **Date structs and types that wrap `time.Time` keep their value in a date or time input.** A type such as `civil.Date` that declares the date widget rendered a blank date input, so an unchanged save stored the zero value. Its `MarshalText` now fills the input under the same rules as text, and its zero or null value renders none. Types that wrap `time.Time`, a type defined as `time.Time` or a struct embedding it (whose `MarshalText` is RFC 3339), render and bind like a `time.Time`: the defined type takes the parsed time converted, and the embedded time is set through its `UnmarshalText` with RFC 3339 when the type does not read the input's own form, and a result that reports itself null afterwards (`IsNull`, or a `driver.Valuer` answering nil) is refused, so a struct holding a validity flag next to an embedded time is never half set, while a nullable time with its own RFC 3339 `UnmarshalText`, such as `null.Time`, binds. A time-of-day or date-and-time struct whose text needs seconds, such as `civil.Time`, gets the submitted value with seconds, since the inputs leave zero seconds out. A nil embedded time pointer renders an empty input instead of panicking and takes a submitted date. Untagged, a struct embedding `time.Time` with text methods of its own, such as a date-only one, stays a text input through them, as before; with a date widget it is a date struct. A null result of any bind step is refused, so a posted RFC 3339 value cannot half set a flagged struct either, and a wrapped time only takes what the input sends, not its own RFC 3339 text. A string type whose `UnmarshalText` needs seconds, as a `civil.Time` held as text does, gets them when the input leaves zero seconds out. A date struct reads its own text back first, so a struct whose text the picker cannot show renders it in a text input that is not marked invalid, and an unchanged save keeps the value, also in a repeatable row after an add-row or remove-row click. A `min` or `max` on such a field is checked on the value as the input shows it; text the input can only show as text cannot be checked, so with a bound set it is rejected with the format to enter and its input is marked invalid. A date type whose `UnmarshalText` reads only its own spelling, such as `02.01.2006`, cannot take what the picker sends: it renders as a text input of its own text without a format hint (a `form:"placeholder=…"` names its format), also when null or sensitive, and a `min` or `max` on it fails the form when it is shown, with the new `mx.DateTimeTagError`, instead of rejecting every save. A date type that is no struct, such as an integer day count with text methods, binds through its `UnmarshalText` too, instead of failing with `cannot assign time`. A date or time field bound through a pointer gets a new pointee, so a rejected value never reaches a record a shallow-copying `load` shares with a store. A registered single-value struct that holds no time, such as a money type, is now a text input through `MarshalText`/`UnmarshalText` instead of a date and time field, and an untagged type defined as `time.Time` is detected as one. `shadcn` renders every such catch-all value as a styled text input in a `Field`, like a string, instead of the unstyled `html` fallback.
- **A `form:"sensitive"` date or time field no longer shows its value.** The date, datetime and time inputs echoed a sensitive `time.Time`, and the new text path would have echoed the stored text. Like a sensitive text input, the field now renders empty, and its validation errors no longer quote the posted value.
- **A `FormWidgetHint` now decides the widget the way a tag does.** `mx.DetectField` returns the hint's widget as the tag's `Widget`, so renderers and the parser read one value. A type whose `FormWidget()` returns `"date"` rendered `type="datetime-local"` and failed to bind; it now renders and binds like a field tagged `form:"widget=date"`, which lets a date type declare itself once instead of every field carrying the tag. A `"skip"` hint now skips the field like `form:"widget=skip"` instead of rendering and binding it, a `"checkbox"` hint renders a checkbox in `shadcn` instead of a switch, and an explicit tag such as `form:"widget=text"` is no longer overridden by an `"email"` hint. A nil pointer field whose element type has a value-receiver `FormWidget` no longer panics `DetectField`.
- **The `shadcn` date and time inputs render `min` and `max`.** A `form:"min=…,max=…"` tag reached the `html` date input but was dropped by `shadcn`; both now render it on a date, datetime or time input. The text fallback leaves it out, since a text input cannot enforce it, and the server now checks the range for every date and time field, `time.Time` included, so a crafted POST cannot store a date outside it (`mx.CheckDateTimeRange`). The check compares the wall clock, as the bounds have no zone, so an RFC 3339 offset cannot move a date across a bound (an offset of 24 hours or more, which RFC 3339 does not allow, is rejected), a value whose text cannot be read back is rejected when a bound is set, and it allows a time range that wraps past midnight (`min=22:00,max=06:00`) as the browser does. A value outside the range is still bound, so the re-render and a repeatable row keep it after an add-row click, but the form is not saved. A value that is not whole minutes after the `min` gets `step="1"`.
- **A tampered POST can no longer null a `form:"readonly"` field.** The `__clear__` sentinel was handled before the readonly skip, so a POST carrying a readonly nullable field's `__present__` and `__clear__` sentinels set it to null, an approval timestamp for instance. Readonly is now checked first, at the top level and in repeatable rows.
- **Repeatable rows keep their stored values.** Rows were rebuilt from the submission alone, so a readonly cell, never parsed, was saved as its zero value on every save, and a typo in a date followed by an add-row click blanked the row and the next save stored zero. Each row's `__row__` marker now carries the index of the loaded row it came from, also through the re-renders of add-row and remove-row clicks and of a rejected save, and a rebuilt row starts as a copy of that loaded row before the submission is applied: a readonly cell keeps its value, a rejected cell behaves as the same top-level field does (it keeps the stored value, or the posted text where the field keeps that, as for an out-of-range number or date and rejected date text), and a cell the client did not send stays as loaded; an added row starts at the zero value. Rows are matched by position, so an `onSubmit` that returns an error must not reorder or remove them; rows it appends have no origin. The marker reads `o<index>`, so the bare `1` of a page rendered before this change claims no row. The marker is client-controlled, so each loaded row goes to at most one submitted row and a marker naming no loaded row claims none.
- **`form:"readonly"` on a section, nested struct or repeatable field protects what is in it.** It only worked on leaf fields: a readonly repeatable field still rendered add and remove buttons and rebuilt its rows from a POST, and the fields of a readonly section were editable. Now every field inside is readonly, keeping its own options (a field tagged `form:"-"` stays skipped), a readonly repeatable field keeps its loaded rows and takes no add or remove command, and a readonly file input is disabled instead of required.
- **A nil pointer section no longer hides the fields after it.** The form walk stopped at a nil `*T` tagged `form:"nested"` or `form:"section=…"`, so every later field at that level was neither rendered nor bound, and in a repeatable row it was saved as zero. The nil section is still skipped; the fields after it are not.
- **Numbers are checked against `form:"min=…,max=…"` on the server.** Only the browser enforced the range, although a code comment said the parsers did. The bound value of an integer is compared exactly, so an `int64` bound above 2^53 holds, while a float is compared at its own precision, as a browser compares doubles, so a value on a decimal bound such as `max=99.9` is in range. `NaN` and infinities, which no bound can hold and no input submits, are rejected for any float field, as is a number text longer than 64 bytes. A bound that is no valid HTML number, or lies past the range of a double, is ignored, as a browser ignores it. The number stays bound, so the re-render shows it, and the form is not saved.
- **A rejected text value keeps the stored one.** A type whose `UnmarshalText` zeroes itself on an error, as `netip.Addr`'s does, lost its stored value on a typo, and in a repeatable row an add-row click then saved the zero value; the value is now unmarshaled into a copy.
- **Every field bound through a pointer gets a new pointee.** Text, number, bool and checkbox-set fields wrote a submission through the existing pointer, so with a shallow-copying `load` a rejected form could still change a record shared with a store; now they bind into a new pointee that starts as a copy, as date and time fields do. The fields of a pointer section are still written in place, so `load` must return a deep copy.
- **A slice of times tagged `form:"repeatable"` is skipped.** A `time.Time`, a type defined as it or a struct embedding one has no visible fields, so its rows had no cells and every save zeroed the stored times, and no input binds a slice of times; the field is now never rendered or bound, and its stored value is left alone.
- **An apostrophe in a form tag value no longer drops the options after it.** The tag parser opened a quote at any `'` or `"`, so `form:"label=Owner's email,readonly"` lost `readonly`, and `required` or `sensitive` the same way. Only a quote that starts a value opens one now.
- **The form tag of an embedded struct applies to the fields it promotes.** The form walk skipped embeds before reading their tag, so `form:"-"` on an embed still rendered and bound its fields, a mass-assignment hole, and `form:"readonly"` did not protect them. Now `-` omits them and `readonly` locks them.
- **Repeatable rows validate like top-level fields.** A blank optional enum in a row could not be saved, nor the "(none)" of an optional int enum anywhere, and a checkbox set in a row accepted members that are not enum values.
- **A required field cannot be skipped by leaving out its `__present__` sentinel.** The sentinel gate came before the required check, so a POST without it saved a blank required field, the zero of an int enum included, of a submit-only form or an added repeatable row. A field the decider does not render, or renders as nothing, is still left alone. A number that is no pointer is never blank there, so its `required` and range are not checked without the sentinel yet (`TODOS.md`).
- **A hidden `time.Time` round-trips.** It rendered without its zone in a format its parser could not read, so a form with a hidden `UpdatedAt` failed every save without a message; it now renders and binds as RFC 3339, whatever its widget or range tag.
- **A sensitive number no longer shows its value.** The number inputs of both renderers echoed it in `value=`.
- **A blank `form:"sensitive"` field keeps its stored value.** A sensitive text, number, date or time input renders empty, so an unchanged save cleared the stored value (or failed a required field), also in a repeatable row. A blank submission now means "unchanged", as with a password field, when a value is stored; with nothing stored a blank is parsed as before, so `required` and a number's range still reject it. A new value replaces the stored one. Every optional sensitive input now has the clear control, also when it is no pointer or nullable, as it could otherwise never be emptied: it is ticked while the value is empty, so a clear survives an add-row click or a rejected save, and text typed into the input wins over the tick. While a value is stored, a required sensitive input no longer carries the HTML `required` attribute, which made the browser block an unchanged save until the secret was typed again; the server still rejects a blank when nothing is stored. A sensitive field's validation errors read "not a valid value", as its parse errors do, since a type's own message may quote the kept value; the `UserMessage` of an `mx.UserMessageError`, such as an `mx.FormMessage` for a password rule, shows as it is (without what it was wrapped or joined with), and a number outside its range says "must be at least …". A tick on a value without a null clears it by parsing the blank, so its range, `required` and validation apply. The clear checkbox is named after its field ("clear Token"; shadcn takes the wording from the new `Labels.FormClearField`, such as "%s leeren"), and in shadcn only the box and its caption are clickable. `FieldKind.KeepsStoredWhenBlank` is the rule a custom decider can use to tell which kinds keep a blank's stored value, and `mx.KeepsStoredValue` tells whether a field holds one; the form handler and both renderers share it, so a value that is null without being zero, such as `sql.NullString{String: "x"}`, counts as nothing stored everywhere. Checkboxes, checkbox sets, hidden and file fields are unaffected, and so are selects, which show their stored option, so their "(none)" still clears.
- **A `time.Time` shown as text round-trips.** A time field tagged `form:"widget=text"`, also of a type defined as `time.Time`, rendered Go's `String()` form (`2026-07-06 09:05:07 +0200 CEST`), which its own parser rejects, so saving the form unchanged failed. It renders as RFC 3339 now (`2026-07-06T09:05:07+02:00`, fraction included), and the zero time as an empty input.
- **A form body that cannot be parsed is rejected.** A truncated multipart body, or one past a size limit, read as a submission of no fields, so `onSubmit` saved the unchanged record and the user's edits were lost without a message. The handlers now answer 400, or 413 for a body past the multipart limits, net/http's 10 MB for a urlencoded body or an `http.MaxBytesReader` limit (a larger one a caller sets is honored), without calling `onSubmit`. A URL query that cannot be parsed, such as one with a semicolon, fails nothing on its own: the body, parsed apart from it, is the submission. The handler removes a multipart body's temporary files when it returns, also behind a middleware that passes a copy of the request on, which the server's own cleanup does not see, and keeps the uploads of a form a middleware parsed itself. A body a middleware parsed first and failed to read is answered with 400 too, not saved as what was left; a middleware that reads a multipart form must parse all of it (`r.ParseMultipartForm` or `r.FormValue`, not only `r.ParseForm`). A body that is no form, such as JSON or one without a content type, is answered with 415.
- **A forged clear no longer sets a required field to null.** A POST carrying the `__clear__` sentinel of a required nullable or pointer field, which the forms never offer, saved it as null past the required check, also in a repeatable row. It is now rejected as "required".
- **A blank required number is rejected.** A number that is no pointer stored 0 for a blank submission, which a required field accepted, and the server-side range skipped a blank, so `form:"min=1,required"` saved 0. A blank required number is now "required", as the browser already checks, and the range is checked on the 0 a blank optional one stores, so an optional `min=1` int that the user empties shows an error rather than storing an out-of-range 0; a number that may be empty is a pointer.
- **A time with a zone keeps its instant.** The input shows a stored time's wall clock in its own location without the zone, and the submission was read as UTC, so 10:30 in Berlin was saved as 10:30 UTC and moved by the offset on every save, a date at midnight in New York by a day. An unchanged submission now keeps the stored value as it is, also in the hour repeated when the clocks go back, for `time.Time`, types defined as it and wrappers. A changed one is read in the zone the input showed the stored time in (the new `mx.DateTimeInputSpec.Time`, a wrapper's own named zone where its text shows the same time): a date at the first instant of its day, which is not midnight where the clocks jump over it, a time on the stored date, and a time posted with an offset keeps its instant in that zone. The range is checked on the time as the form will show it. A new record still reads UTC.
- **Pressing Enter saves a form with repeatable rows.** Enter in an input clicks the form's first submit button, which was the first row's "Remove" (or "Add row"), so it removed a row instead of saving. Such a form now starts with an out-of-sight default submit button that saves.
- **A hidden number or bool binds.** A `form:"hidden"` int, float or bool went through the text binder, which cannot parse it, so every save of a form with such a hidden field, such as a numeric per-row key, failed. It binds as a number or checkbox value now; a nil `*bool` stays nil.
- **A password widget no longer shows the stored password.** A field tagged `form:"widget=password"` (or hinted so) without `sensitive` rendered its stored value into the page. A password widget is now sensitive: it never shows its value, and a blank submission keeps the stored one.
- **An error no input shows is reported.** A hidden field, a hidden cell in a repeatable row such as a per-row key, a field the decider does not render and a path that names no field have no place for a message, so a parse error of a tampered hidden value, or an `onSubmit` `mx.FieldErrors` entry for such a path (a typo, say), failed the save with no visible reason. Such errors are now shown above the form, once each, written as an `onSubmit` error is: the `UserMessage` of an `mx.UserMessageError` such as an `mx.FormMessage`, otherwise `mx.FormErrorMessage`. An error of a repeatable field itself, such as one asking for at least one row, shows inside its box after the legend. A nil entry in `mx.FieldErrors` is no error rather than a panic.
- **A field the decider renders as nothing is left alone.** It still got a `__present` sentinel, so every save blanked it; it gets none now.
- **An upload is read into memory only up to `MaxMemory`.** A file field read the whole upload into a buffer of its size, whatever `ReflectFormConfig.MaxMemory` said, so one large POST could exhaust memory. A larger file is now a field error; the new `mx.FormMaxMemory(ctx)` tells a decider the limit.
- **The note on unlisted checkbox-set members says they are removed.** It read "(some current values are not available)", although saving drops them; it now says "(… and are removed when you save)", in `html` and in `shadcn.Labels.FormValuesNotAvailable`. A readonly set, which is never saved and keeps them, still reads "(some current values are not available)" (`shadcn.Labels.FormValuesNotAvailableReadonly`).
- **A select whose stored value the list no longer offers says that saving removes it.** Its disabled placeholder "(current value not available)" submits nothing, so a save clears an optional select's value without a word. An optional, editable select now has the note "The current value is removed when you save." below it, as the checkbox set has; a required one, which rejects the blank, and a readonly one, which is never saved, have none. The new `OptionPreface.RemovesValue` tells a renderer when to say so for a select or a checkbox set, and shadcn takes the note from the new `Labels.FormValueRemovedOnSave`.
- **A hidden or catch-all `database/sql` type such as `sql.NullInt64` can be saved.** It rendered as a Go struct (`{42 true}`), which nothing could read back, so every save of the form failed. Both renderers now show the value it holds (a time in RFC 3339, a null as nothing), and the `html` parser reads it back with the type's `Scan`, as text, as bytes for a JSON column whose `Scan` takes only those, or as a time; an unchanged value is kept as it is, so a valid empty string stays valid and a time keeps its zone. A number follows the rules of a number field: it must be finite, at most 64 characters long and within `form:"min=…,max=…"`. Text that does not scan says what to enter ("not a whole number", "not true or false", an RFC 3339 example) instead of quoting the driver, and the clear control empties the value with `Scan(nil)`. Only a type a driver reads and writes this way is shown so: one with its own `MarshalText`, `UnmarshalText` or `String`, such as a `null.String` or a money type that shows major units but stores cents, keeps using those, and bytes an input cannot hold unchanged (invalid UTF-8, a NUL) are not shown at all, so such a field fails visibly instead of being changed by a save. The new `mx.DriverValueText` gives a custom renderer the same text.
- **The labels of a shadcn checkbox set name their boxes.** Each label pointed at an id no checkbox had, so the boxes had no accessible name and a click on an option's caption did nothing. Each box now has an id derived from the field and its value with `mx.KeyedIDValue`, valid also for a value with spaces.
- **A checkbox set's field label names its group.** In both renderers the label pointed with `for` at an id no element had, and the group of boxes had no accessible name. The label now has an id that the group's `aria-labelledby` names.
- **A blank required `*sql.NullString` no longer crashes the request.** The required check called `Value` through the nil pointer and panicked; `mx.IsNull` now reads a nil pointer as null first.
- **A field a decider renders as nothing is not written by a POST.** A decider that hides a field from a user by returning a nil `Render`, or a `Render` that returns nil, still had the field bound when a crafted POST added its `__present` sentinel and a value. The handler now asks `Render` with a render probe and leaves such a field alone, so a per-request decider can hide a field without also clearing its `Parse`; every decider in a chain is asked, nested hiding wrappers included. The probe holds no errors, as a GET's render, and `mx.IsRenderProbe` lets a `Render` answer it without building markup, as the built-in deciders do, so binding a POST costs no markup. A `form:"repeatable"` field is hidden the same way, by a nil `Render` for its path: its rows are then neither rendered nor rebuilt, and a forged row command does nothing.
- **A required enum whose zero is an option saves it.** A required select could not save a constant whose value is the zero, such as `Low = iota` or an option `"0"`, from any list (`Enums`, `EnumStrings`, `NamedOptions`, `Options`, `form:"options=…"`); a blank choice is now told from the submitted text. A non-empty value that binds the zero of an enum without such a constant is validated, not taken as "no choice".
- **An enum with nothing chosen shows "(none)" or "Choose…".** An enum whose zero is none of its options, such as an `iota+1` enum or an `int64` key from a list, showed "(current value not available)" and a note that saving removes it on a new record; its zero now selects the list's own empty option or the leading "(none)" ("Choose…" when required; a readonly select prompts no choice). The new `mx.SelectedOption` decides it for a renderer.
- **A required file field that holds a file does not demand a new upload.** Its input rendered HTML `required`, so the browser blocked every save of the edit form until the user picked the file again, although a blank file input keeps the stored file.
- **An option list that fails while a POST is read shows no internal error.** A context option list (`form:"options=…"` with a provider) failing during the parse put its error text, a go-errs call stack included, next to the field; it now shows the `UserMessage` of an `mx.UserMessageError` it returned, such as an `mx.FormMessage`, or `mx.FormErrorMessage` unless `mx.RevealInternalServerErrors` is set, also in a `form:"sensitive"` field.
- **Rejected date text is not echoed back without a bound.** Date or time text longer than any input sends (64 bytes) is rejected without being bound into the re-render or quoted in the error, as a `database/sql` number's error does not quote long text, so one large POST cannot grow into a larger page.
- **A date on its `max` day passes in zones whose clocks skip midnight.** A date read in such a zone starts at 01:00, which the range check compared as later than the `max` day; a date input's bounds now compare dates only.
- **A value whose promoted methods panic no longer crashes the required check.** A struct embedding a nil pointer whose `IsNull` or `IsZero` is promoted panicked when the field was checked; it now counts as empty.
- **The shadcn clear checkbox is named in the caption's language, and the removal note is read out.** A translation that set only `Labels.FormClear` left the box's accessible name in English ("clear Token" next to "leeren"); `FormClearField` now follows the caption. The note that saving removes a select's value, or a checkbox set's members, is the control's `aria-describedby` in both renderers.
- **Form errors no longer carry a call stack.** A few parser errors (`cannot parse value into …`, unsupported kinds, an unsettable field or set) were go-errs values whose text includes the stack, and the form showed it; they are plain errors now.
- **Text, number and option fields guard the type's own methods.** An `UnmarshalText` that panics, as the promoted method of a nil embedded pointer does, failed the whole request; it is now a field error, and a `MarshalText` that panics renders no value instead of failing the page. A `form:"sensitive"` text, number, catch-all, select or checkbox-set field's parse error reads `not a valid value` instead of quoting the posted value through an option list, `UnmarshalText` or `strconv` error; only the `UserMessage` of an `mx.UserMessageError` in it, such as a number's range error ("must be at least …"), shows, without what it was wrapped or joined with.
- **A readonly nullable field no longer offers a clear control.** The parser ignores `__clear__` on readonly fields, so the checkbox did nothing; neither renderer shows it now.
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
