# shadcn port — TODOs

Build order for porting shadcn/ui components to go-mx, derived from
`RESEARCH.md`. Items are ordered so every dependency is built before its
dependents (topological), and by complexity within each phase.

Complexity is rated 1–6 **for this project** — a server-side Go HTML
generator with no client framework. It diverges from the React rating in
`RESEARCH.md` wherever a native web-platform primitive replaces a Radix
behavior (e.g. `<dialog>`, `<details>`, the Popover API). The established
project convention: port to native HTML/CSS first, accept a small inline
`onclick`/script only where the platform has no equivalent, and reach for
HTMX (`hx/`) for anything needing a server round-trip.

## Conventions every component follows

- Signature `Name(typedVariant…, attribsChildren ...any) *mx.Element`; a
  variant/size is a leading typed param, `""` selects the default.
- Build the element with `html.*`, then return `finish(e, "<slot>", base)`
  — `finish` merges caller `class`, sets one `data-slot`, dedupes attribs.
- Multi-part components are multiple exported functions linked by id or
  nesting, not one struct (see `AlertDialog*`).
- A `cva`-style variant component also exports a `…Classes(variant)` helper
  (the `buttonVariants` equivalent — see `ButtonClasses`).
- Class strings are transcribed verbatim from shadcn/ui new-york-v4 (TW v4),
  minus Radix-only positioning/animation classes.

## Status legend

`[x]` done · `[ ]` todo · **Cx** = complexity 1–6

---

## Phase 0 — Foundations (done)

- [x] **clsx** — `clsx` npm-package port; `Join` (`clsx/` subpackage)
- [x] **twmerge** — `tailwind-merge` v3.6.0 port; `Merge` (`twmerge/` subpackage)
- [x] **cva** — `class-variance-authority` port (`cva/` subpackage)
- [x] **Cn** — thin `cn` helper composing clsx + twmerge (`cn.go`)
- [x] **finish** — shared class/attrib merge (`component.go`)
- [x] **Button** · Cx 2 — exports `ButtonClasses`
- [x] **Alert** · Cx 2 — `Alert` / `AlertTitle` / `AlertDescription`
- [x] **AlertDialog** · Cx 4 — native `<dialog>`, full sub-part set

---

## Phase 1 — Tier 1 structural (no deps, pure markup) (done)

Styled HTML only; these port ~1:1 and are low-risk warm-ups.

- [x] **Separator** · Cx 1 · deps: none
  `<div role="separator">` with `orientation` (horizontal/vertical).
- [x] **Skeleton** · Cx 1 · deps: none
  One `<div>` with `animate-pulse rounded-md bg-accent`.
- [x] **Label** · Cx 1 · deps: none
  Styled `<label>`. Blocks **Form** later.
- [x] **Input** · Cx 1 · deps: none
  Styled `<input>`.
- [x] **Textarea** · Cx 1 · deps: none
  Styled `<textarea>`.
- [x] **AspectRatio** · Cx 1 · deps: none
  Wrapper `<div>` using the CSS `aspect-ratio` property; takes a `ratio`
  param (width/height).
- [x] **Badge** · Cx 2 · deps: cva
  `<span>` + variants; exports `BadgeClasses` (mirrors `ButtonClasses`).
- [x] **Card** · Cx 2 · deps: none
  `Card` / `CardHeader` / `CardTitle` / `CardDescription` / `CardAction` /
  `CardContent` / `CardFooter`.
- [x] **Table** · Cx 2 · deps: none
  `Table` (+ overflow wrapper) / `TableHeader` / `TableBody` /
  `TableFooter` / `TableRow` / `TableHead` / `TableCell` / `TableCaption`.
  Blocks **DataTable**.
- [x] **Breadcrumb** · Cx 2 · deps: none
  `Breadcrumb` / `BreadcrumbList` / `BreadcrumbItem` / `BreadcrumbLink` /
  `BreadcrumbPage` / `BreadcrumbSeparator` / `BreadcrumbEllipsis`. The
  separator/ellipsis default to inline lucide SVG icons (`icons.go`).
- [x] **Avatar** · Cx 2 · deps: none
  `Avatar` / `AvatarImage` / `AvatarFallback`. React detects image-load
  failure in JS; this port renders both, overlays the image (`absolute
  inset-0`) and hides it via `<img onerror>` to reveal the fallback.
- [x] **Pagination** · Cx 2 · deps: **ButtonClasses** (done)
  `Pagination` / `PaginationContent` / `PaginationItem` /
  `PaginationLink` / `PaginationPrevious` / `PaginationNext` /
  `PaginationEllipsis`. Links styled via `ButtonClasses(ButtonGhost, …)`.

---

## Phase 2 — Tier 2 simple interactive (done)

Native form controls and disclosure elements carry most of the behavior.
Stateful components (Toggle, Tabs, ToggleGroup) keep shadcn's ARIA markup +
a tiny inline JS handler, with an HTMX opt-out (a caller-supplied `hx-*`
attribute skips the default `onclick`).

- [x] **Progress** · Cx 2 · `<div>` track + inner indicator with inline
  `transform: translateX(-N%)`.
- [x] **Switch** · Cx 2 · Native `<input type="checkbox" role="switch">`;
  thumb via `before:` pseudo-element.
- [x] **Toggle** · Cx 2 · deps: cva · `<button aria-pressed>`; exports
  `ToggleClasses`. Default `onclick` flips `aria-pressed`; HTMX opt-out.
- [x] **RadioGroup** · Cx 3 · `<div role="radiogroup">` + `RadioGroupItem`
  styled void `<input type="radio">`; dot via `before:` pseudo-element.
- [x] **Checkbox** · Cx 3 · Styled void `<input type="checkbox">`; check mark
  via `background-image` data URL. Indeterminate via
  `CheckboxIndeterminateScript(id)`.
- [x] **Collapsible** · Cx 3 · Native `<details>`/`<summary>`/`<div>`; root
  carries `group` so a chevron can rotate via `group-open:`.
- [x] **Accordion** · Cx 3 · Multiple `<details>`; single-mode via the
  native `<details name="">` exclusive group, multiple-mode via `""`.
- [x] **Tabs** · Cx 3 · Faithful ARIA `tablist`/`tab`/`tabpanel` +
  one shared `tabsSelect` inline script; HTMX opt-out on each trigger.
- [x] **ToggleGroup** · Cx 3 · deps: **ToggleClasses** · single & multiple
  modes via one shared `toggleGroupClick` script that reads the parent's
  `data-type` at click time; HTMX opt-out on each item.
- [x] **ScrollArea** · Cx 3 · Single overflow `<div>` + CSS scrollbars;
  `ScrollBar` intentionally not exported (the native scrollbar is a
  pseudo-element, not an element).
- [x] **Slider** · Cx 4 · Single-thumb = native `<input type="range">`;
  two-thumb range = two overlaid inputs + fill `<div>` + `sliderClamp`
  script. Selected by `len(values)` (1 or 2; other panics).
- [x] **InputOTP** · Cx 4 · N real `<input maxlength="1">` slots + hidden
  field + `otpAdvance`/`otpKey` focus-management script.

---

## Phase 3 — Tier 3 floating (done)

Built on the native **Popover API** (`popover` + `popovertarget`) plus **CSS
Anchor Positioning** (`anchor-name` / `position-anchor` / `position-area`),
the native replacement for Radix's Popper/Portal/DismissableLayer/FocusScope.
Chromium 125+ and Safari 26 anchor next to the trigger; Firefox (anchor
positioning still in progress) falls back to viewport-centered popovers —
functional, just not anchored.

- [x] **Popover** · Cx 3 · `Popover` / `PopoverTrigger(id)` / `PopoverContent(id, side)`.
  Private helpers (`popoverAnchorStyle`, `popoverContentStyle`, `mergeStyle`,
  `PopoverSide`) live in `popover.go` and are reused by the rest of Phase 3.
- [x] **Tooltip** · Cx 3 · `<span>` wrapper trigger + `tooltipShow`/`tooltipHide`
  script wired via `mouseover`/`mouseout`/`focusin`/`focusout`.
- [x] **HoverCard** · Cx 3 · Same shape as Tooltip + timer-based open/close
  delays (`hoverCardShow`/`hoverCardHide` script; defaults 700/300ms).
- [x] **Select** · Cx 4 · Native `<select>` + `<optgroup>` + `<option>` styled
  with `appearance: base-select` (Chrome 130+/Safari TP); native chrome
  fallback elsewhere. `SelectTrigger`/`SelectContent`/`SelectItem` etc.
  intentionally not ported.
- [x] **DropdownMenu** · Cx 4 · Full set: Trigger, Content, Item, Label,
  Separator, Group, CheckboxItem, RadioGroup, RadioItem, Shortcut, Sub*.
  Shared `menuKeyNav` script (ArrowUp/Down/Home/End/typeahead/Escape +
  Sub Arrow/Right/Left) and `menuOpen` script (focus first item on open,
  flip trigger `aria-expanded`).
- [x] **ContextMenu** · Cx 4 · Same item parts; trigger is a `<div>` with
  `oncontextmenu` calling `contextMenuOpen` which pixel-positions the
  popover at the cursor. Content carries no anchor-positioning.
- [x] **Menubar** · Cx 4 · `Menubar` / `MenubarMenu` / `MenubarTrigger` /
  `MenubarContent` + all item parts. `menubarHover` script implements the
  OS-menubar click-to-switch-without-clicking behavior.
- [x] **NavigationMenu** · Cx 4 · `NavigationMenu` / `List` / `Item` /
  `Trigger(id)` / `Content(id, side)` / `Link(active)`. Chevron rotates on
  `[button[aria-expanded=true]>&]`. `Viewport` and `Indicator` not ported.

---

## Phase 4 — Tier 4 modal / overlay (native `<dialog>`)

Reuse the `<dialog>` approach already proven in `alertdialog.go`
(top layer, `::backdrop`, focus trap, Escape-to-close — no framework).

- [x] **Dialog** · Cx 3 · deps: none (shares `<dialog>` infra with AlertDialog)
  `Dialog` / `DialogTrigger` / `DialogContent` / `DialogHeader` /
  `DialogFooter` / `DialogTitle` / `DialogDescription` / `DialogClose`.
  Native `<dialog>` like AlertDialog, plus light-dismiss (backdrop click) and a
  built-in corner close button. Blocks **Sheet**.
- [x] **Sheet** · Cx 4 · deps: **Dialog**
  Native `<dialog>` pinned to an edge via per-side inset classes (top/right/
  bottom/left); reuses Dialog's close-button helper. SheetSide type, `""` =
  right. Blocks **Sidebar**.
- [x] **Drawer** · Cx 5 · deps: **Sheet** / **Dialog** (Option A)
  Native bottom `<dialog>` (reuses the Dialog/Sheet modal infra — top layer,
  ::backdrop, Escape, light-dismiss) plus one shared `drawerStart` pointer-drag
  script: drag the grab handle down, past a ~40% threshold it closes, else
  snaps back. The most client JS of any port, in the Slider/Resizable inline-
  script pattern. Dropped vs Vaul: multi snap-points, momentum physics,
  background-scale, non-bottom directions.

---

## Phase 5 — Tier 5 composite / heavy

Each leans on a React-only library with no Go equivalent — behavior must be
reimplemented (server-side in Go, with HTMX, or with a small script).
Order within the phase is by dependency, then complexity.

- [x] **Form** · Cx 5 · deps: **Label** (+ Input/Checkbox/RadioGroup/etc.)
  `Form` (native `<form>`) / `FormItem` / `FormLabel` (error-aware via
  `data-error`) / `FormDescription` / `FormMessage`. FormField/FormControl are
  React context/Slot plumbing with no server equivalent, so not ported; the
  caller wires ids/aria directly and renders FormMessage with the error.
- [x] **Command** · Cx 5 · deps: none
  Filterable command list (cmdk). One shared `commandFilter` script substring-
  matches each item's text on input, hides non-matching items/empty groups and
  toggles CommandEmpty. Fuzzy ranking and arrow-key nav not reproduced. Blocks
  **Combobox**.
- [ ] **Combobox** · Cx 5 · deps: **Popover** + **Command** + Button
  Composition recipe, not a primitive.
- [x] **Calendar** · Cx 5 · deps: Button
  Single-month grid generated server-side with Go's `time` package; selected
  day marked via `aria-selected`. `Calendar(month, selected, …)`. Prev/Next are
  plain buttons; month nav is a caller-wired round-trip. Blocks **DatePicker**.
- [x] **DatePicker** · Cx 5 · deps: **Popover** + **Calendar** + Button
  Composition recipe (Popover trigger + Calendar in PopoverContent) — shipped as
  a gallery example, matching how shadcn ships DatePicker (copy-paste, not a
  primitive).
- [x] **Carousel** · Cx 5 · deps: Button
  Native CSS `scroll-snap` track (drag/swipe/keyboard free); Previous/Next
  scroll the track one viewport via a tiny inline onclick. Autoplay/loop/
  orientation not ported.
- [x] **Resizable** · Cx 5 · deps: none
  Flex panels + a draggable handle; one shared `resizeStart` pointer-drag
  script adjusts the adjacent panels' flex-basis (the Slider-style tradeoff).
  `ResizeDirection` horizontal/vertical.
- [x] **Toast** (Sonner) · Cx 5 · deps: none
  `Toaster` (fixed bottom-right region) + one shared script defining a global
  `toast(msg, {description, duration})`: appends a styled toast and auto-
  dismisses it. Triggered from any onclick. Swipe-to-dismiss and stacking
  offsets not reproduced; HTMX out-of-band swap is the server-pushed alternative.
- [ ] **Chart** · Cx 6 · deps: none · **DEFERRED** (design decision pending)
  shadcn's Chart only adds a CSS-variable theming layer over Recharts' SVG;
  porting means generating the chart SVG ourselves. The divergence is in
  *generation*, not runtime — every option below can stay server-side. Options:
  - **A (recommended) — hand-rolled Go SVG generator** (bar/line/area): compute
    linear scales, axis ticks, gridlines and shapes (`<rect>`/`<polyline>`/
    `<path>`) in Go, colored from the theme's `--chart-1…5`. Pure server-side,
    zero client runtime — the most native-first option and a real go-mx
    strength. Responsive via SVG `viewBox` + `width:100%` (no resize-observer,
    the native replacement for Recharts' ResponsiveContainer). Tooltips: native
    `<title>` per data point by default (zero JS); optional styled JS tooltip
    later. Cost: reimplementing a *scoped* charting lib (not Recharts parity) —
    one of the two largest remaining builds (with DataTable).
  - **B — a Go charting dependency** (e.g. `wcharczuk/go-chart` → SVG): less
    code, but a dep whose styling/theming won't match shadcn.
  - **C — a JS charting lib via CDN** (like Tailwind/Shiki): fits the gallery's
    CDN demo pattern but reintroduces a client runtime for a package component —
    against the native-first model.
  Open decisions before building (A): chart types (bar/line/area, or also
  pie/radial?), tooltip strategy (native `<title>` vs styled JS), and the
  `ChartData{Categories []string; Series []ChartSeries}` + `ChartContainer`
  theming-wrapper API shape.
- [x] **DataTable** · Cx 6 · deps: **Table** + **DropdownMenu** + Input +
  Checkbox + **Pagination**
  Composition recipe (toolbar filter Input + Columns DropdownMenu + Table with
  select checkboxes, sortable headers and per-row action menus + selection
  count + Previous/Next) shipped as a gallery example, matching how shadcn ships
  DataTable (copy-paste, not a primitive). Live email filter via one small
  script; sorting/pagination would be server-side (HTMX/query params) in a real
  app and are rendered but inert here.
- [x] **Sidebar** · Cx 6 · deps: **Sheet** + Button + Separator
  ~18 sub-parts (Provider/Sidebar/Trigger/Inset/Header/Content/Footer/Group*/
  Menu*/Sub*/Separator). Expand↔icon collapse via a `data-state` on the
  group/sidebar-wrapper; one shared `sidebarToggle` script persists it to the
  `sidebar_state` cookie, restores on load, and binds Cmd/Ctrl+B. Floating/
  inset variants and the mobile-becomes-Sheet behavior are not reproduced.

---

## Hard dependency edges (must respect)

```
ButtonClasses ─► Pagination, Calendar, Carousel, Combobox, DatePicker
ToggleClasses ─► ToggleGroup
Collapsible ───► Accordion
Popover infra ─► Tooltip, HoverCard, Popover, Select, NavigationMenu
Popover ───────► DropdownMenu ─► ContextMenu, Menubar
Command ───────► Combobox
Calendar ──────► DatePicker
Dialog ────────► Sheet ─► Sidebar
Table ─────────► DataTable
Label ─────────► Form
```

`Sidebar` and `DataTable` are the convergence points — everything else
should land before them.
