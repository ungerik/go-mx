# hx

[![Go Reference](https://pkg.go.dev/badge/github.com/ungerik/go-mx/hx.svg)](https://pkg.go.dev/github.com/ungerik/go-mx/hx)

[htmx](https://htmx.org) integration for go-mx: the `hx-*` attributes as
`mx.Attrib` constructors, htmx event and CSS-class name constants, and a server
side for HTTP handlers (typed request readers and response-header setters).
Attribute values are typed wherever htmx constrains them, so an invalid swap
style or a misspelled boolean is caught in Go instead of silently emitting a
broken attribute.

```go
import (
    "github.com/ungerik/go-mx/html"
    "github.com/ungerik/go-mx/hx"
)

html.Button(
    hx.Post("/clicked"),
    hx.Target("#result"),
    hx.Swap(hx.SwapOuterHTML),
    "Click me",
)
// <button hx-post="/clicked" hx-target="#result" hx-swap="outerHTML">Click me</button>
```

## Request attributes

Every attribute in the [htmx attribute reference](https://htmx.org/reference/#attributes)
has a constructor returning an `mx.Attrib`. The URL / CSS-selector / string-valued
ones take a `string`:

`hx.Get`, `hx.Post`, `hx.Put`, `hx.Patch`, `hx.Delete`, `hx.Target`,
`hx.Trigger`, `hx.Select`, `hx.SelectOOB`, `hx.Vals`, `hx.Confirm`, `hx.Prompt`,
`hx.Include`, `hx.Params`, `hx.Indicator`, `hx.Sync`, `hx.Headers`, `hx.Request`,
`hx.Ext`, `hx.Inherit`, `hx.Disinherit`, `hx.DisabledElt`, `hx.PushURL`,
`hx.ReplaceURL`.

The remaining attributes have typed values (below). `mx.NewAttrib("hx-…", value)`
is the escape hatch for anything without a dedicated constructor.

## Typed attribute values

htmx constrains some attribute values to a closed set of keywords, a boolean, or
a single fixed string. Those constructors are typed so the constraint is enforced
in Go rather than left to a raw string:

- **`hx.Swap(style hx.SwapStyle, modifiers ...string)`** — the swap style is a
  typed [`SwapStyle`](enums.go) keyword enum (`hx.SwapInnerHTML`,
  `hx.SwapOuterHTML`, `hx.SwapTextContent`, `hx.SwapBeforeBegin`,
  `hx.SwapAfterBegin`, `hx.SwapBeforeEnd`, `hx.SwapAfterEnd`, `hx.SwapDelete`,
  `hx.SwapNone`). Modifiers (`"swap:1s"`, `"settle:1s"`, `"scroll:bottom"`, …)
  are passed through verbatim. An out-of-set style is not silently emitted: it
  returns an `mx.ErrAttrib`, so rendering the enclosing element fails with a
  descriptive error. The bare conversion `hx.SwapStyle("innerHTML")` also works
  for dynamic values, and the generated `Valid`/`Validate`/`Enums`/`EnumStrings`/
  `String` methods report whether a value is in the set.
- **`hx.Boost(enable bool)`, `hx.History(enable bool)`,
  `hx.Validate(enable bool)`** — htmx accepts only `"true"` / `"false"` here, so
  these take a `bool`.
- **`hx.Disable`, `hx.Preserve`, `hx.HistoryElt`** — htmx ignores the value, so
  these are bare boolean attributes (`html.BoolAttrib`) used directly as
  constants, not called.
- **`hx.EncodingMultipart`** — an `mx.ConstAttrib` for the only value htmx
  accepts on `hx-encoding` (`multipart/form-data`, for file uploads).

### Event handlers (`hx-on`)

- **`hx.On(event, handler string)`** renders `hx-on:<event>` for a DOM event:
  `hx.On("click", "alert('hi')")` → `hx-on:click="alert('hi')"`.
- **`hx.OnHTMX(event, handler string)`** renders `hx-on::<event>` (shorthand for
  `hx-on:htmx:<event>`) for an htmx event:
  `hx.OnHTMX("after-request", "doStuff()")` → `hx-on::after-request="doStuff()"`.

### Out-of-band swaps

- **`hx.SwapOOB(style hx.SwapStyle, selector ...string)`** marks an element to
  swap out of band with a typed style and optional target CSS selector(s).
- **`hx.SwapOOBTrue`** is the plain `hx-swap-oob="true"` form (swap the element
  in by its own id).
- **`hx.SelectOOB(value string)`** selects OOB content to pull from a response.

## Event and class name constants

- **Event names** ([events.go](events.go)) — every htmx event as a constant
  holding its full `htmx:`-prefixed name: `hx.EventLoad`, `hx.EventConfigRequest`,
  `hx.EventBeforeRequest`, `hx.EventAfterRequest`, `hx.EventBeforeSwap`,
  `hx.EventAfterSwap`, `hx.EventAfterSettle`, … (request, swap, node-processing,
  history, validation, XHR-progress and SSE groups). Use them with `hx.Trigger`,
  `hx.OnHTMX`, or the response-header trigger setters: `hx.Trigger(hx.EventLoad)`.
- **CSS classes** ([classes.go](classes.go)) — the classes htmx applies during
  the request lifecycle: `hx.ClassAdded`, `hx.ClassIndicator`, `hx.ClassRequest`,
  `hx.ClassSettling`, `hx.ClassSwapping`. Use them with `html.Class(...)` to
  style indicators and in-flight content.

## Server side: HTTP headers

[headers.go](headers.go) implements the htmx request / response header protocol
for `net/http` handlers.

**Header name constants** cover the full request and response header sets from
the [htmx reference](https://htmx.org/reference/#request_headers):
`hx.HeaderRequest`, `hx.HeaderBoosted`, `hx.HeaderTrigger`, `hx.HeaderTarget`,
`hx.HeaderCurrentURL`, `hx.HeaderPrompt`, `hx.HeaderLocation`, `hx.HeaderPushURL`,
`hx.HeaderRedirect`, `hx.HeaderRefresh`, `hx.HeaderReplaceURL`, `hx.HeaderReswap`,
`hx.HeaderRetarget`, `hx.HeaderReselect`, and the `HX-Trigger-After-*` names.

**Request readers** take an `*http.Request` and return a `bool`:

| Function                        | Reports                          |
|---------------------------------|----------------------------------|
| `hx.IsRequest(r)`               | request was made by htmx         |
| `hx.IsBoosted(r)`               | request came via `hx-boost`      |
| `hx.IsHistoryRestoreRequest(r)` | history restore after cache miss |

**Response setters** take an `http.ResponseWriter` and set the
correspondingly-named `HX-*` header:

- Redirects and history: `hx.SetLocation`, `hx.SetPushURL`, `hx.SetRedirect`,
  `hx.SetRefresh`, `hx.SetReplaceURL`.
- Swap control: `hx.SetReswap`, `hx.SetRetarget`, `hx.SetReselect`.
- Client-side triggers: `hx.SetTrigger`, `hx.SetTriggerAfterSettle`,
  `hx.SetTriggerAfterSwap` join their `events ...string` with `, `. The
  `…JSON` variants (`hx.SetTriggerJSON`, `hx.SetTriggerAfterSettleJSON`,
  `hx.SetTriggerAfterSwapJSON`) take a `map[string]any` event-to-detail payload
  and return an error if it cannot be marshaled.

```go
func handler(w http.ResponseWriter, r *http.Request) {
    if !hx.IsRequest(r) {
        http.Redirect(w, r, "/", http.StatusFound)
        return
    }
    hx.SetTrigger(w, hx.EventLoad, "myEvent")
    w.WriteHeader(http.StatusNoContent) // 204, no body
}
```

## Reflected forms: `FieldDecider`

`hx.FieldDecider` is the HTMX layer of go-mx's layered form-rendering chain. It
delegates rendering, parsing, and validation to `html.FieldDecider`, then adds
`hx-trigger="change"` to the rendered live inputs (`input` / `textarea` /
`select`, excluding buttons, hidden, and clear-sentinel fields), so a form that
wires `hx-post` / `hx-target` submits live as the user edits. Install it once per
route subtree with `mx.Middleware(hx.FieldDecider)`, or pass it to a single
`mx.ReflectFormHandler`. (The hx layer is intentionally minimal in v1; deeper
HTMX form semantics are deferred.)

## Loading htmx

`hx.ScriptFromCDN` and `hx.ScriptDebugFromCDN` are ready-made `<script>` elements
that load htmx (currently 2.0.10) from unpkg with a Subresource Integrity hash.

## Online sources

- [htmx attribute reference](https://htmx.org/reference/#attributes)
- [htmx request & response headers](https://htmx.org/reference/#headers)
- [htmx events](https://htmx.org/events/)
- [htmx CSS class reference](https://htmx.org/reference/#classes)
- [hx-on attribute](https://htmx.org/attributes/hx-on/)
