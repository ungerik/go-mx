# shadcn

A Go port of parts of [shadcn/ui](https://ui.shadcn.com), built on the go-mx
`html` primitives. It contains a growing set of components, the thin `cn`
class-merging helper, and faithful Go ports of the three npm packages
shadcn/ui's class handling depends on — each in its own subpackage:

- [`clsx`](clsx/) — flattens class-value arguments into one class string
- [`twmerge`](twmerge/) — resolves Tailwind utility-class conflicts
- [`cva`](cva/) — class-variance-authority, the variant builder

shadcn/ui ships React components. go-mx renders HTML on the server with no
client runtime, so this is a *port*, not a wrapper: the markup and Tailwind
classes are reproduced in Go, and behavior that React delegates to libraries is
re-expressed with web-platform primitives.

## `Cn` — the class-merging helper

shadcn's `cn` helper is `clsx` piped through `tailwind-merge`. `Cn` is the Go
equivalent — a one-line composition: it flattens its arguments with
[`clsx.Join`](clsx/) and resolves Tailwind conflicts with
[`twmerge.Merge`](twmerge/) so a later class overrides an earlier one:

```go
shadcn.Cn("px-2 py-1", "p-4")        // "p-4"
shadcn.Cn("text-sm", "text-lg")      // "text-lg"
```

`clsx.Join` accepts strings, `[]string`, nested `[]any` and `map[string]bool`
conditional classes; see its doc comment for the contract. `twmerge.Merge` is a
faithful port of `tailwind-merge` v3.6.0 (Tailwind CSS v4) — the merge
algorithm and full config live in the [`twmerge`](twmerge/) subpackage
(`merge.go`, `classmap.go`, `validators.go`, `parse.go`, `defaultconfig.go`).

## `cva` — the variant builder

shadcn/ui declares each component's style variants with
[class-variance-authority](https://cva.style) (`cva`): a base class plus a
table of variants, compound variants and defaults. The [`cva`](cva/)
subpackage is a faithful Go port of cva v0.7.1.

```go
buttonVariants := cva.New(cva.Config{
	Base:            "inline-flex ...",
	Variants:        map[string]map[string]string{"variant": {/* ... */}},
	DefaultVariants: map[string]string{"variant": "default"},
})
buttonVariants(map[string]string{"variant": "destructive"})
```

Like the npm package, `cva` only concatenates classes — compose it with `Cn`
to resolve Tailwind conflicts, just as shadcn/ui composes `cva` with `cn`.
`Button` and `Alert` declare their variants this way.

## Component model — how a React component maps to Go

| shadcn/ui (React)            | this package (Go)                        |
|------------------------------|------------------------------------------|
| `<Alert variant="..." />`    | `Alert(variant, attribsChildren...)`     |
| `className` prop             | `html.Class("...")` in the variadic args |
| `cn(variants(), className)`  | the shared `finish` helper               |
| `data-slot="..."`            | always emitted by the component          |
| children / `...props`        | the rest of the variadic `...any`        |

Every component function takes go-mx's variadic `...any` of mixed attributes
and children, with any typed variant as a leading parameter. A caller supplies
extra classes the normal go-mx way, by passing `html.Class("...")`.

**Why `finish` exists.** go-mx's `CheckedWriter` rejects two attributes with
the same name on one element. A component carries base classes *and* may
receive a caller `class`, so it must merge them into a single `class` itself.
`finish` (in `component.go`) does that: it pulls every caller `class` value,
merges `Cn(baseClasses, callerClasses...)` so caller classes win conflicts,
guarantees one `data-slot`, deduplicates any other repeated attribute
(last wins), and rebuilds the element's attribute list.

Conditional classes: go-mx attributes have no `clsx`-style object form. Compute
them with `Cn(...)` yourself and pass the result as `html.Class(...)`.

## Button

`Button(variant ButtonVariant, size ButtonSize, attribsChildren ...any)`. Pass
`""` for either to get the default. It renders `<button data-slot="button"
data-variant data-size class>` and defaults to `type="button"` (a server-
rendered `<button>` would otherwise submit an enclosing form); pass `html.Type`
to override.

| Variant       | Size                                      |
|---------------|-------------------------------------------|
| `ButtonDefault`     | `SizeDefault`, `SizeXS`, `SizeSM`    |
| `ButtonDestructive` | `SizeLG`, `SizeIcon`                 |
| `ButtonOutline`     | `SizeIconXS`, `SizeIconSM`           |
| `ButtonSecondary`   | `SizeIconLG`                         |
| `ButtonGhost`, `ButtonLink` |                              |

`ButtonClasses(variant, size)` returns just the merged class string (the
equivalent of shadcn's exported `buttonVariants`) — use it to give the button
look to a non-button element such as an `AlertDialogTrigger`.

## Alert

`Alert(variant AlertVariant, attribsChildren ...any)` with `AlertTitle` and
`AlertDescription`. Variants: `AlertDefault`, `AlertDestructive`. `Alert` adds
`role="alert"` unless the caller supplies a role.

```go
shadcn.Alert(shadcn.AlertDefault,
    shadcn.AlertTitle("Heads up!"),
    shadcn.AlertDescription("You can add components to your app."),
)
```

## AlertDialog: native `<dialog>` instead of Radix

shadcn's Alert Dialog is built on Radix UI, which supplies the modal behavior
(top layer, backdrop, focus trap, Escape-to-close) in the browser via React.
That has no equivalent in server-rendered Go HTML, so this port replaces Radix
with the **native HTML `<dialog>` element**, which the browser provides for
free — no JavaScript framework, no client runtime.

What this means concretely:

- **`AlertDialogPortal` and `AlertDialogOverlay` are not ported.** A modal
  `<dialog>` already renders in the browser's top layer, and its backdrop is
  the `::backdrop` pseudo-element. The overlay's `bg-black/50` is reproduced
  with the Tailwind v4 `backdrop:` variant (`backdrop:bg-black/50`) on the
  `<dialog>`.
- **Opening.** `AlertDialogTrigger(dialogID, ...)` renders a button whose
  `onclick` calls `document.getElementById(dialogID).showModal()`.
- **Closing.** `AlertDialogContent` wraps its children in a
  `<form method="dialog">`. Any submit button inside it closes the dialog.
  `AlertDialogAction` and `AlertDialogCancel` are such buttons; they also set
  `dialog.returnValue` (default `"confirm"` / `"cancel"`) so the caller can
  tell which was used.
- **Dropped classes.** shadcn's content `class` includes Radix-only pieces:
  `fixed top-[50%] left-[50%] translate-x/y` (replaced by `m-auto` — see
  below), `z-50` (the top layer is above everything), and
  `data-[state=open|closed]:animate-*` (a native `<dialog>` has no
  `data-state`). These are dropped; the box, sizing and layout classes are
  kept, plus `backdrop:bg-black/50`.
- **Centering (`m-auto`).** A native modal `<dialog>` is centered in the top
  layer by the user agent's `margin: auto`. Tailwind v4 Preflight (which this
  package requires) resets every element's margin to `0`, defeating that and
  pinning the open dialog to the top-left — so the content keeps an explicit
  `m-auto` to restore the centering.
- **Display (`open:grid`, not `grid`).** shadcn mounts the content only while
  open; a native `<dialog>` is always in the DOM and stays hidden via the UA
  rule `dialog:not([open]){display:none}`. An unconditional `grid` is an
  author-origin style that overrides that rule and leaks the closed dialog onto
  the page, so the display utility is scoped to `open:` (the `[open]` attribute
  `showModal()` sets).

`AlertDialogTrigger` renders a `<button>`, so pass it content and styling, not
a nested `Button` (a `<button>` inside a `<button>` is invalid HTML). Use
`html.Class(shadcn.ButtonClasses(...))` for the button look:

```go
shadcn.AlertDialog(
    shadcn.AlertDialogTrigger("confirm-remove",
        html.Class(shadcn.ButtonClasses(shadcn.ButtonDestructive, shadcn.SizeDefault)),
        "Remove"),
    shadcn.AlertDialogContent("confirm-remove",
        shadcn.AlertDialogHeader(
            shadcn.AlertDialogTitle("Remove this item?"),
            shadcn.AlertDialogDescription(
                "It will be moved to the archive and can be restored later."),
        ),
        shadcn.AlertDialogFooter(
            shadcn.AlertDialogCancel("Cancel"),
            shadcn.AlertDialogAction("Continue"),
        ),
    ),
)
```

`AlertDialogTrigger` and `AlertDialogContent` are linked by the `dialogID`
string, not by DOM nesting. The `AlertDialog` wrapper is purely structural.

### `dialogID` is not sanitized

`dialogID` is interpolated into an `onclick` handler and an `id` attribute.
`AlertDialogTrigger` and `AlertDialogContent` call `validateID`, which **panics**
unless the id is a non-empty string of letters, digits, `-` and `_`. Pass a
constant, developer-chosen id — never user input.

### Known limitations

- Open/close is instant. A native `<dialog>` has no `data-state`, so the Radix
  enter/exit animations are not reproduced. They can be added in plain CSS with
  `@starting-style` and `transition-behavior: allow-discrete`; this package
  does not emit that CSS.
- The `<dialog>` does not get `aria-labelledby` / `aria-describedby`
  automatically (Radix wires these). Pass them yourself with `html.Attrib`.
- `AlertDialogAction` only closes the dialog by default. For a server-side
  action, add `html.Attrib("formaction", "/url")` plus `html.FormMethodPOST`,
  or attach `html.OnClick` for a client-side action.

## Tier 1 components

A set of structural and styled-markup components that port ~1:1 from
shadcn/ui. Each follows the model above — a leading typed variant where one
exists, then `attribsChildren ...any`, returning a single `data-slot`-tagged
element — and transcribes the new-york-v4 Tailwind classes verbatim.

- **Separator** — `Separator(orientation, …)`. A `<div role="separator">`; a
  `data-orientation` drives the sizing, a vertical separator also gets
  `aria-orientation`. Pass `""` for the default horizontal orientation.
- **Skeleton** — `Skeleton(…)`. A pulsing placeholder `<div>`; give it
  sizing classes to match the content it stands in for.
- **Label** — `Label(…)`. A styled `<label>`; link it with `html.For(id)`.
- **Input** — `Input(…)`. A styled void `<input>`; set the type the normal
  way with `html.Type("…")`.
- **Textarea** — `Textarea(…)`. A styled `<textarea>`.
- **AspectRatio** — `AspectRatio(ratio, …)`. A `<div>` holding a CSS
  `aspect-ratio` — the native replacement for Radix's padding-bottom hack.
  `ratio` is width/height (e.g. `16.0/9.0`); `0` selects a square.
- **Badge** — `Badge(variant, …)`. A `<span>` plus variants
  (`BadgeDefault`, `BadgeSecondary`, `BadgeDestructive`, `BadgeOutline`).
  `BadgeClasses(variant)` returns just the class string, mirroring
  `ButtonClasses`.
- **Card** — `Card` / `CardHeader` / `CardTitle` / `CardDescription` /
  `CardAction` / `CardContent` / `CardFooter`, all `<div>`s.
- **Table** — `Table` / `TableHeader` / `TableBody` / `TableFooter` /
  `TableRow` / `TableHead` / `TableCell` / `TableCaption`. `Table` renders a
  `<table>` inside an overflow-scrolling `data-slot="table-container"` `<div>`;
  caller attributes land on the inner `<table>`.
- **Breadcrumb** — `Breadcrumb` / `BreadcrumbList` / `BreadcrumbItem` /
  `BreadcrumbLink` / `BreadcrumbPage` / `BreadcrumbSeparator` /
  `BreadcrumbEllipsis`. The separator and ellipsis default to inline lucide
  SVG icons; pass children to override them.
- **Avatar** — `Avatar` / `AvatarImage` / `AvatarFallback`. See the note
  below.
- **Pagination** — `Pagination` / `PaginationContent` / `PaginationItem` /
  `PaginationLink(active, size, …)` / `PaginationPrevious` /
  `PaginationNext` / `PaginationEllipsis`. Links are styled with
  `ButtonClasses` — outline for the active page, ghost otherwise.

shadcn/ui draws icons from `lucide-react`. go-mx has no icon dependency, so the
few icons these components need by default (chevrons, ellipsis) are inlined as
SVG path data in `icons.go`.

### Avatar: revealing the fallback without JS

shadcn/ui's Avatar mounts either the image or the fallback based on a
JavaScript load check. This port renders both: `AvatarImage` is positioned
`absolute inset-0` so it overlays `AvatarFallback`, and it carries a default
`onerror` handler that hides the image when its `src` fails to load — revealing
the fallback beneath it. The `absolute inset-0` is a deliberate divergence from
shadcn's verbatim `aspect-square size-full`, the native replacement for Radix's
mount/unmount. Pass your own `onerror` attribute to override the default.

```go
shadcn.Avatar(
    shadcn.AvatarImage(html.Src("/avatar.png"), html.Alt("Jane Doe")),
    shadcn.AvatarFallback("JD"),
)
```

## Tier 2 components

Tier 2 covers interactive components whose behavior is carried by native form
controls and disclosure elements. Each follows the same model as Tier 1 — one
leading typed parameter where there is one, then `attribsChildren ...any`,
returning a single `data-slot`-tagged element — and transcribes the shadcn/ui
new-york-v4 Tailwind classes verbatim, with Radix's `data-[state=*]` selectors
rewritten to native equivalents:

| Radix selector                  | Native rewrite                         |
|---------------------------------|----------------------------------------|
| `data-[state=checked]:` (input) | `checked:`                             |
| `data-[state=on]:` (button)     | `aria-pressed:`                        |
| `data-[state=active]:` (button) | `aria-selected:`                       |
| `data-[state=open]:` (details)  | `group-open:` (with `group` on parent) |

- **Progress** — `Progress(value, …)`. Track + indicator with an inline
  `transform: translateX(-N%)`; the value is known server-side. Clamped to
  `[0, 100]`; sets `role="progressbar"` and `aria-valuemin/max/now`.
- **Switch** — `Switch(…)`. A single styled void `<input type="checkbox"
  role="switch">`; the thumb is drawn with a `before:` pseudo-element on the
  `appearance-none` input (the native equivalent of Radix's child `Thumb`).
- **Toggle** — `Toggle(variant, size, …)` plus `ToggleClasses(variant, size)`
  (the `toggleVariants` equivalent). Variants `ToggleDefault`, `ToggleOutline`;
  sizes `ToggleSizeDefault`, `ToggleSizeSM`, `ToggleSizeLG`. A `<button
  aria-pressed>`; a default inline `onclick` flips `aria-pressed` (see
  [HTMX integration](#htmx-integration) below).
- **RadioGroup** — `RadioGroup(name, …)` + `RadioGroupItem(name, value, …)`.
  Items are styled void `<input type="radio">`s with the dot drawn via
  `before:`. Roving focus and exclusive selection are native to radios with a
  shared `name`.
- **Checkbox** — `Checkbox(…)`. Styled void `<input type="checkbox">`; the
  check mark is a `background-image` data URL drawn when `:checked`.
  Indeterminate is a JavaScript-only DOM property — use
  `CheckboxIndeterminateScript(id)` to flip it after the page loads.
- **Collapsible** — `Collapsible` / `CollapsibleTrigger` / `CollapsibleContent`.
  Native `<details>`/`<summary>`; the root carries `group` so a caller chevron
  can rotate via `group-open:rotate-180`.
- **Accordion** — `Accordion` / `AccordionItem(groupName, …)` /
  `AccordionTrigger` / `AccordionContent`. Each item is a native `<details>`;
  **single mode** uses the same non-empty `groupName` for every item (the
  native `<details name="…">` exclusive group, one open at a time), **multiple
  mode** passes `""` (independent items). `AccordionTrigger` appends a default
  chevron-down icon that rotates `group-open:rotate-180`.
- **Tabs** — `Tabs(id, …)` / `TabsList(…)` / `TabsTrigger(tabsID, value,
  active, …)` / `TabsContent(tabsID, value, active, …)`. Faithful
  `role="tablist"`/`role="tab"`/`role="tabpanel"` markup; one short
  `tabsSelect` script is emitted once per `Tabs` instance (guarded with
  `if(!window.tabsSelect)`).
- **ToggleGroup** — `ToggleGroup(groupType, variant, size, id, …)` +
  `ToggleGroupItem(groupID, value, variant, size, …)`. `groupType` is
  `ToggleGroupSingle` or `ToggleGroupMultiple`. Items are styled with
  `ToggleClasses` plus join classes; one `toggleGroupClick` script reads the
  parent's `data-type` at click time to handle both modes.
- **ScrollArea** — `ScrollArea(…)`. A single overflow `<div>` with CSS-styled
  scrollbars. shadcn's `ScrollBar` is intentionally **not exported** — in the
  native port the scrollbar is the `::-webkit-scrollbar` pseudo-element (plus
  Firefox's `scrollbar-width`/`scrollbar-color`), not an element, so a Go
  `ScrollBar` would have nothing to render. (Same precedent as
  `AlertDialogOverlay`.)
- **Slider** — `Slider(min, max, step, values, id, …)`. `len(values)==1`
  renders a single-thumb styled void `<input type="range">`; `len(values)==2`
  renders a two-thumb range (two overlaid inputs + a fill `<div>` + one
  `sliderClamp` script that keeps the fill in sync). Any other length panics.
- **InputOTP** — `InputOTP(id, name, length, …)` + `InputOTPSeparator(…)`.
  Renders `length` real `<input maxlength="1">` slots plus a hidden field
  named `name` that the shared `otpAdvance` / `otpKey` script keeps in sync.
  This is a divergence from shadcn's `input-otp` library (one hidden input +
  fake slot `<div>`s): real inputs give a real per-slot caret and simpler
  per-slot styling, at the cost of N inputs in the form.

### CSS pseudo-element indicators on void inputs

A native `<input type=checkbox|radio>` is a void element and cannot hold the
child indicator shadcn renders for the check / dot / switch thumb. Switch,
Checkbox and RadioGroupItem instead draw the indicator with CSS on the
`appearance-none` input — a `before:` pseudo-element (Switch thumb,
RadioGroup dot) or a `background-image` data URL (Checkbox check) that shows
on `:checked`. Pseudo-elements render on checkbox/radio inputs once `appearance`
is removed. This is a deliberate divergence from shadcn's two-element
structure, mirroring the Avatar `absolute inset-0` note.

### Inline scripts for interactive state

Tabs, ToggleGroup, Slider (range mode) and InputOTP each emit one short
`<script>` once per component instance, guarded with `if(!window.fn)`. The
script is the native replacement for the React reducer / Radix context that
shadcn relies on. Inline `<script>` bodies are emitted with `html.Script(mx.Raw(...))`
— the same pattern `html.StyleElem` uses for CSS. The `Toggle` default
`onclick` is one inline expression (no `<script>` element); same for
`TabsTrigger` and `ToggleGroupItem`.

### HTMX integration

Toggle, TabsTrigger and ToggleGroupItem all check whether the caller passed
any `hx-*` attribute (the `hasHX` helper in `component.go`). When one is
present, the component **skips** its default `onclick` — htmx is in charge.
The signature is unchanged; opt in by passing `hx.Post(…)` / `hx.Get(…)` /
`hx.Swap(…)` / `hx.Target(…)` from `github.com/ungerik/go-mx/hx` alongside the
other attribs:

```go
// Toggle posting its state to the server. The server returns the re-rendered
// button with aria-pressed flipped.
shadcn.Toggle("", "",
    hx.Post("/toggle-bold"), hx.Swap("outerHTML"),
    "Bold",
)

// Tabs whose triggers swap server-rendered panels in.
shadcn.Tabs("settings",
    shadcn.TabsList(
        shadcn.TabsTrigger("settings", "account", true,
            hx.Get("/tabs/account"), hx.Target("#settings-panel"), "Account"),
        shadcn.TabsTrigger("settings", "billing", false,
            hx.Get("/tabs/billing"), hx.Target("#settings-panel"), "Billing"),
    ),
    html.Div(html.ID("settings-panel"), html.Class("flex-1 outline-none"), "Account panel"),
)
```

The shared `tabsSelect` / `toggleGroupClick` scripts are still emitted by the
root (they are tiny and harmless if no trigger calls them).

### Open / close animations

shadcn's Radix-driven open/close height animations on `CollapsibleContent` and
`AccordionContent` are not reproduced: a native `<details>` snaps open and
closed. They can be brought back in plain CSS with `@starting-style` and
`transition-behavior: allow-discrete`; this package does not emit that CSS.

## Tier 3 components

Tier 3 covers the floating components — content that opens above the page in
the top layer and anchors to a trigger. shadcn's React components wrap Radix's
Popper + Floating UI + Portal + DismissableLayer + FocusScope; the native port
replaces all of that with the HTML **Popover API** (`popover` attribute +
`popovertarget` / `popovertargetaction`) and **CSS Anchor Positioning**
(`anchor-name` / `position-anchor` / `position-area`).

| Native primitive               | What it replaces                        |
|--------------------------------|-----------------------------------------|
| `popover="auto"` + `popovertarget` | Radix Root / Portal / Trigger lifecycle |
| `:popover-open` pseudo-class   | Radix `data-[state=open]:`              |
| `position-anchor` + `position-area` | Radix Popper / Floating UI positioning |
| `<details name="…">` (Phase 2) | (in nav scenarios where one-open-at-a-time fits without JS) |

Radix `data-[state=*]` selectors are rewritten:

| Native element           | Radix selector             | Native rewrite             |
|--------------------------|----------------------------|----------------------------|
| `[popover]`              | `data-[state=open]:`       | `[&:popover-open]:` or `:popover-open:` (Tailwind v4 variant) |
| popover trigger button   | `data-[state=open]:`       | `aria-expanded:` (`menuOpen` script flips it on toggle) |

- **Popover** — `Popover` / `PopoverTrigger(id, …)` / `PopoverContent(id, side, …)`.
  The shared building block; everything else in Tier 3 reuses its anchor /
  position style helpers. Sides: `PopoverTop`, `PopoverRight`, `PopoverBottom`
  (default), `PopoverLeft`. Empty `""` selects the default.
- **Tooltip** — `Tooltip` / `TooltipTrigger(id, …)` / `TooltipContent(id, side, …)`.
  Default side is `PopoverTop`. The trigger is a `<span>` so any element can
  sit inside without nesting buttons; a shared `tooltipShow` / `tooltipHide`
  script wires `mouseover`/`mouseout`/`focusin`/`focusout` (the Popover API
  has no declarative hover-to-open).
- **HoverCard** — `HoverCard` / `HoverCardTrigger(id, openMs, closeMs, …)` /
  `HoverCardContent(id, side, openMs, closeMs, …)`. Same shape as Tooltip
  with timer-based open/close delays (defaults 700ms open, 300ms close to
  match shadcn). Trigger and content both fire show/hide so quick
  trigger-to-content travel doesn't close the card.
- **Select** — `Select` / `SelectGroup(label, …)` / `SelectOption(value, …)`.
  A native `<select>` + `<optgroup>` + `<option>` with `appearance: base-select`
  so Chromium 130+ and Safari Tech Preview render the dropdown with the full
  shadcn look; older browsers (and Firefox as of mid-2026) keep native chrome
  — visually different, fully functional, a real form control. shadcn's
  `SelectTrigger` / `SelectValue` / `SelectContent` / `SelectItem` /
  `SelectScrollUpButton` / `SelectScrollDownButton` are Radix-only abstractions
  that collapse to nothing here and are **not ported**.
- **DropdownMenu** — full set: `DropdownMenu` / `Trigger(id)` / `Content(id, side)`
  / `Item` / `Label` / `Separator` / `Group` / `CheckboxItem(checked, …)` /
  `RadioGroup(name)` / `RadioItem(name, value, selected, …)` / `Shortcut` /
  `Sub` / `SubTrigger(subID)` / `SubContent(subID)`. A shared `menuKeyNav`
  inline script handles ArrowUp/Down/Home/End/typeahead/Escape and
  ArrowRight/Left to open/close sub-menus. `menuOpen` (also shared) auto-focuses
  the first item on open and flips the trigger's `aria-expanded` for the
  open-state look.
- **ContextMenu** — same item parts as DropdownMenu, but the Trigger is a
  `<div>` wrapper with `oncontextmenu="contextMenuOpen(event, '{id}')"` and a
  shared `contextMenuOpen` script that prevents the browser's native context
  menu, pixel-positions the popover at the cursor (`top` / `left` in clientX /
  clientY), and shows it. ContextMenuContent carries **no** anchor-positioning
  — see "Cursor-positioned popovers" below.
- **Menubar** — `Menubar` / `MenubarMenu` / `MenubarTrigger(id)` /
  `MenubarContent(id, side)` plus all the item parts. A shared `menubarHover`
  script implements the OS-menubar click-to-switch-without-clicking idiom:
  hovering a trigger while a sibling menu is open closes it and opens this one.
- **NavigationMenu** — `NavigationMenu` / `List` / `Item` / `Trigger(id)` /
  `Content(id, side)` / `Link(active, …)`. The trigger appends a chevron-down
  icon that rotates when the popover opens. shadcn's `NavigationMenuViewport`
  (a shared content area shared across items) and `NavigationMenuIndicator`
  (the arrow that tracks the active trigger) are **not ported**: each item's
  content is its own popover, and active styling lives on each `Link` via the
  `active` bool (which emits `data-active="true"` + `aria-current="page"`).

### Anchor positioning

CSS Anchor Positioning is shipped in **Chromium 125+** and **Safari 26**;
**Firefox is in progress** as of mid-2026. In Chromium and Safari the floating
content is positioned correctly relative to its trigger via
`position-anchor: --{id}` + `position-area: {top|right|bottom|left}`. In
Firefox without anchor positioning the popover still opens in the top layer
and can be dismissed — it just renders centered in the viewport instead of
next to the trigger. This is a deliberate tradeoff: zero JavaScript for
positioning, perfect UX in modern browsers, functional degradation elsewhere.

### Cursor-positioned popovers (`ContextMenu`)

CSS Anchor Positioning is element-relative, not cursor-relative. ContextMenu
opts out of anchor positioning entirely (`ContextMenuContent` carries no
`position-anchor` / `position-area` style); the shared `contextMenuOpen`
script sets `position: fixed; top: {clientY}px; left: {clientX}px; margin: 0;
position-anchor: none` on the popover before calling `showPopover()`. This
works in every browser that ships the Popover API.

### Inline-script footprint

Each component emits one short `<script>` once per instance, guarded with
`if(!window.fn){…}`. The full list:

| Script               | Components                              | Size  | Wired via |
|----------------------|-----------------------------------------|-------|-----------|
| `tooltipShow/Hide`   | Tooltip                                 | ~2 lines | onmouseover/onmouseout/onfocusin/onfocusout |
| `hoverCardShow/Hide` | HoverCard                               | ~4 lines | same + setTimeout for delays |
| `contextMenuOpen`    | ContextMenu                             | ~2 lines | oncontextmenu |
| `menuKeyNav`         | DropdownMenu / ContextMenu / Menubar / NavigationMenu | ~25 lines | onkeydown on Content |
| `menuOpen`           | same set                                | ~3 lines | ontoggle on Content |
| `menubarHover`       | Menubar                                 | ~8 lines | onmouseenter on Trigger |

All scripts are bundled together with their content via `html.Script(mx.Raw(…))`,
the same pattern `html.StyleElem` uses for raw CSS. None of them are loaded
from a CDN; everything ships inline with the rendered HTML.

### HTMX integration

`popovertarget` is declarative and doesn't conflict with `hx-*` attributes,
so Popover / Tooltip / HoverCard / DropdownMenu / Menubar / NavigationMenu
triggers don't need an HTMX opt-out — pass `hx.*` attribs as normal and they
work alongside the popover-open behavior. For menu *items* (which fire
actions on click), the items are plain `<div>` / `<a>` / `<button>` — pass
`html.OnClick`, `hx.Post`, etc. however you like; no defaults to suppress.

## Tailwind CSS v4 is required

These components emit Tailwind v4 utility classes and nothing else. The
consuming application **must run Tailwind v4** over its templates and Go source
(so the class names are discovered) or the components render unstyled. This is
the most common "it doesn't work" cause.
