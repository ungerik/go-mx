# logview

Streamed log lines rendered as a readable, searchable surface.

Logs are mostly structured JSON today, and a raw JSON line is unreadable at
stream speed. `logview` parses each line and renders it as its fields — the
timestamp and level promoted, everything else as `key=value` with the value
colored by its JSON type — plus the surface those lines stream into: a filter, a
pause toggle, and a scroll area that follows the stream until you scroll up.

```
11:53:27.781  debug  request started   method=POST path=/api/invoices/42 request_id=c9e90fab
11:53:27.782  info   request finished  method=POST path=/api/invoices/42 status=201 cached=false
11:53:27.914  warn   acquired advisory lock  db={host=db2.internal port=5432 idle=6} last_error=null
[ERROR] failed to renew certificate: rate limited by ACME server
```

It sits above `mx`, `html`, `hx` and `shadcn`, and is fed by
[`mx.SSEResponse`](../sse.go). A worked end-to-end example — a simulated
production log, streamed indefinitely at irregular intervals — is in
[`cmd/example-logstream`](../cmd/example-logstream).

## Streaming a log

The page renders a `View` wired to the stream:

```go
html.Head(
    hx.ScriptFromCDN,
    hx.ScriptSSEFromCDN,
    logview.StyleElement(),
),
html.Body(
    logview.View(
        hx.Ext("sse"),
        hx.SSEConnect("/logs/stream"),
        hx.SSEClose("done"),
    ),
)
```

and the handler sends one `Line` per event:

```go
func stream(w http.ResponseWriter, r *http.Request) {
    sse, err := mx.NewSSEResponse(w, nil)
    if err != nil {
        mx.RespondNonContextError(w, err) // still a clean 500
        return
    }
    defer sse.Close()
    go sse.KeepaliveLoop(r.Context(), 20*time.Second)

    for line := range tail(r.Context(), mx.LastEventID(r)) {
        err := sse.SendEvent(r.Context(), mx.SSEEvent{
            Name: logview.Default.EventName(),
            ID:   line.Seq,
            Comp: logview.Line(line.Text),
        })
        if err != nil {
            return // the client went away
        }
    }
}
```

Two parts of that are not optional in practice:

- **`KeepaliveLoop`.** A log stream is idle most of the time, so it is the
  stream most likely to be dropped by whatever proxy sits in front of it. Pick
  an interval below the shortest idle timeout on the path.
- **`SSEEvent.ID` plus `mx.LastEventID`.** Lines are appended, so a reconnect
  that replays the backlog appends it a second time. An id the handler can
  answer "give me everything after this" for is what prevents that.

Send a burst as one event with `Lines(...)` rather than one event per line: htmx
inserts a fragment's element children individually, so a batch costs one flush
and one swap instead of forty.

## How a line is rendered

| Input | Rendered as |
| ------------------------------ | ------------------------------------- |
| `{"time":…}` as the first key | the value alone, no `time=` label |
| `{"time":…}` later in the record | `time=<value>`, still styled as a time |
| `{"level":"warn"}` | the value alone, colored by the theme |
| `{"msg":…}` with `MessageKey` set | the value alone |
| any other field | `key=value`, value colored by its JSON type |
| `"a string"` | unquoted by default; `QuoteText` adds quotes |
| `{"host":"db1","port":5432}` | `{host=db1 port=5432}`, formatted like a record |
| `["a","b"]` | `[a, b]` |
| a string containing newlines | a `<pre>` block, so a stack trace keeps its indentation |
| anything that is not one JSON object | a plain-text line |
| `WARN …`, `[WARN] …`, `warn: …` | the level word colored, the rest plain |

Field order is preserved, including duplicate keys — two `err` fields in one
record is a bug worth seeing, not one to hide.

The level is a fixed column: `Theme.CSS` gives it a `min-width` in `ch` — one
character in the view's monospace font — as wide as the longest level the theme
defines, so severity can be scanned straight down instead of being chased as it
shifts with the word before it. `LevelUndefined` is excluded from that
measurement, because its key is never displayed: an unknown level shows its raw
value, which has no bound, and such a value runs past the column rather than
being cut off — a log must not hide what it was told.

## Configuration

Everything is a field of `Config`, whose zero value is usable and equals
`logview.Default`. A `Config` is read-only once it is serving; configure it
during program setup.

| Field | Purpose |
| ------------ | ---------------------------------------------------------- |
| `Prefix` | CSS class prefix (default `log-`) |
| `TimeKey`, `LevelKey`, `MessageKey` | the field names a record is read with |
| `Levels` | level vocabulary for plain-text lines |
| `QuoteText` | quote JSON string values |
| `MaxDepth`, `MaxLineLen`, `MaxLines` | the limits below |
| `Height` | scroll area height, a preference — see below |
| `Event` | SSE event name; read it back with `EventName()` |
| `Theme` | colors, level styles, level icons |
| `Labels` | the strings the view renders on its own |

`Height` is what the scroll area *prefers*. The view is a flex column, so a page
that gives the view a height of its own — by making it a flex item, say — has
the scroll area fill whatever the toolbar and error sink leave:

```css
html, body { height: 100% }
body { margin: 0; display: flex; flex-direction: column }
.log-view { flex: 1; min-height: 0 }
```

Three limits exist because a log has no upper bound: `MaxLineLen` truncates one
runaway line, `MaxDepth` stops descending into nested JSON (unbounded recursion
on hostile nesting would overflow the goroutine stack, which is fatal and
unrecoverable), and `MaxLines` caps how many lines the browser keeps.

## Styling

The markup carries only classes, so the same lines can be styled by any theme.
`StyleElement()` emits the stylesheet; `DarkTheme` (the default) and
`LightTheme` are provided.

```go
cfg := &logview.Config{Theme: logview.LightTheme}
```

Every `Class` and every level takes a `Style` with foreground and background
color, bold, italic, underline and strikethrough. `ClassMatch` is the exception:
the filter's matches are a CSS custom highlight rather than a class, so it is
emitted as a `::highlight()` rule, where only color, background-color and
text-decoration apply. A level can additionally
render an image in place of its text:

```go
theme.Levels["fatal"] = logview.LevelStyle{
    Style: logview.Style{Color: "#fff", Background: "#8e1519", Bold: true},
    Image: "/icons/fatal.svg", // alt text is the level value
}
```

### The level class set is closed

A level value comes from whoever wrote the log line, so it is never used as a
class directly. `{"level":"foo hidden"}` would otherwise emit
`class="log-level log-level-foo hidden"` — the space ends the class token and
`hidden` is a Tailwind utility that makes the line disappear.

The rule is therefore not "validate the value" but **only ever emit a token the
theme defined**: `log-level-<v>` when the lowercased value is a key of
`Theme.Levels`, and `log-level-undefined` in every other case. `Theme.Levels`
keys are themselves restricted to `mx.ValidIDRune` characters, so a careless
theme cannot open the hole either. The level *text* still shows the raw value,
escaped, so nothing is lost.

## The view's behavior

One inline script, defined once per page behind a guard, drives all of it
through fixed `data-mx-log-*` attributes — independent of `Prefix`, so one
definition serves every view on the page while each keeps its own filter, pause
and cap. Several views can read one stream: give each its own `hx.Ext("sse")`
and `hx.SSEConnect`, and each opens its own connection and its own backlog, so
pausing or filtering one leaves the others streaming. Putting a single
`sse-connect` above them instead would share one connection, and then pausing
would be a property of the page rather than of a viewer. It watches DOM mutations rather than
htmx events, so it works for SSE, an ordinary swap, or any other script.

- **Pause** keeps arriving lines in the DOM and marks them `data-mx-log-pending`
  instead of detaching them, so order, `hx-swap-oob` targeting and htmx's settle
  step all keep working. Nothing is lost and the connection stays open; while
  paused, the held lines get a `MaxLines` cap of their own, so a view left paused
  under a firehose stays bounded — at up to twice `MaxLines`, since the lines
  already on screen are what the reader is reading and are not trimmed under
  them. The button swaps its caption to `Labels.Resume` and clears the marks in
  order when pressed again. The `N new` badge counts only the held lines the
  filter also shows, because resuming reveals nothing else.
- **Scrolling up pauses too**, and scrolling all the way back down resumes —
  scrolling back to read something and wanting the stream to hold still are one
  intent, not two. Which one paused is remembered, because it decides what
  resumes: a pause the scroll caused is undone by returning to the bottom, while
  a pause the button caused was asked for explicitly and only the button takes
  it back. The threshold is at most `shadcn.StickToBottomDefaultThresholdPx`, so
  scrolling can only pause once the scroll area has already stopped following.
  Either way the state reaches a screen reader through a `role="status"` region
  carrying `Labels.Paused`: a caption change on a button that does not have
  focus announces nothing, and the scroll never touches the button at all. The
  button's caption names the next action rather than carrying `aria-pressed`,
  which would say "Resume, pressed" — the opposite of what the caption says; the
  state a stylesheet needs is `data-mx-log-paused` on the root instead.
- **The filter highlights what it matched**, through the CSS custom highlight
  API rather than by wrapping matches in elements. A log line is a tree of spans
  that htmx swapped in and two MutationObservers are watching, and rewriting it
  on every keystroke would disturb all of that; ranges disturb nothing, and one
  range can span the several spans a field is split into — `status=200` is
  three. Style it through `ClassMatch`, which `Theme.CSS` emits as a
  `::highlight()` rule. A browser without the API filters without highlighting.
- **Filter** hides non-matching lines with `data-mx-log-nomatch`. Pausing and
  filtering are two independent reasons to hide one line, which is why each has
  its own attribute — resuming must not reveal a line the filter excludes — and
  why both are hidden by a stylesheet rule rather than the `hidden` property,
  whose UA-stylesheet rule any author `display` declaration beats.
- **The line cap** trims from the top, but only while the view is scrolled to
  the bottom. Trimming while the reader has scrolled up would drag the text
  they are reading upward by a line for every line that arrives. Scrolling up
  pauses, so that is mostly the paused case above; the scroll position is read
  again at trim time because a scroll event is dispatched after the fact and a
  line can arrive in between.

## Non-goals

- **ANSI escape sequences** are not interpreted; a stream carrying them shows
  them as text.
- **The filter searches the DOM, not the backlog.** A search over everything
  belongs in the handler, as a parameter of the stream URL.
- **A carriage return is a line break here, not a terminal overwrite,** so a
  progress-bar line renders as several lines.

## Requirements

Tailwind for `shadcn.ScrollArea`'s scrollbar styling — the view still scrolls
without it, because the height and overflow it needs are set inline. htmx plus
the SSE extension for the stream itself.
