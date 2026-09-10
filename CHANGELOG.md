# Changelog

All notable changes to this project are documented in this file. The format
follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

Versions use the Go module scheme, `vMAJOR.MINOR.PATCH`, and a release tags
every module of the repo in lockstep: the root module as `vX.Y.Z` and each
submodule with its path as prefix (`fpdf/vX.Y.Z`, `wordpress/vX.Y.Z`).

This file starts with the entry below; the changes made before it are in the
commit history. Nothing is tagged yet, so everything here is unreleased and the
API is still free to change.

## [Unreleased]

### Added

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
  of the exact parts is appended so that distinct keys cannot collide silently.
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
- **`mx.ContentTypeEventStream`** for `text/event-stream`.
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
  independent of the class prefix. Pausing keeps arriving lines in the DOM and
  marks them rather than detaching them, so order, out-of-band targeting and
  htmx's settle step keep working; pause and filter get one attribute each,
  because resuming must not reveal a filtered-out line; and the line cap only
  trims while the view is at the bottom, since trimming under a reader who has
  scrolled up drags the text they are reading upward.
- **`logview.Theme`** gives every value type and every log level a foreground
  and background color, bold, italic, underline and strikethrough, plus an
  optional image rendered in place of a level's text, with the level value as
  its alt text. `DarkTheme` and `LightTheme` ship; `Theme.CSS(prefix)` is
  deterministic.
- **`cmd/example-logstream`** simulates a production log against all of the
  above: randomly composed lines covering every rendering path, streamed
  indefinitely at exponentially distributed intervals with bursts and quiet
  stretches, resumable from a bounded history, with `-rate` and `-fail`.

### Removed

- **`hx.EventNoSSESourceError`.** htmx's `sse` extension never raises
  `htmx:noSSESourceError` — it does not check nesting at all, so an `sse-swap`
  element with no `sse-connect` above it subscribes to nothing silently. The
  constant named an event that cannot occur.

### Changed

- **`mx.SSEEventError` is `"mx-error"`, not `"error"`.** A browser dispatches
  its own transport failures at the `EventSource` under the name `error`, so a
  subscriber to `error` would also fire on every dropped connection, with an
  event carrying no data for htmx to swap.

### Added

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

### Changed

- `web.GlobPageSource.Dir` now scopes the glob: `Pattern` is matched below it
  instead of against the working directory, so the directory is named once
  rather than repeated in both fields where the two could disagree. A `Pattern`
  that is absolute while `Dir` is set is an error, and a directory matching the
  pattern is skipped instead of being read as a page.

### Fixed

- `web.DefaultRenderPage` no longer panics when a page is marked `NoIndex`, and
  now marks drafts and scheduled pages `noindex, nofollow` too — previously
  only an explicit `NoIndex` was honored, so a draft served from a preview
  deployment could be indexed.

[Unreleased]: https://github.com/ungerik/go-mx/commits/main
