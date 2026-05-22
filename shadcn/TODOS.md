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

## Phase 1 — Tier 1 structural (no deps, pure markup)

Styled HTML only; these port ~1:1 and are low-risk warm-ups.

- [ ] **Separator** · Cx 1 · deps: none
  `<div role="separator">` with `orientation` (horizontal/vertical).
- [ ] **Skeleton** · Cx 1 · deps: none
  One `<div>` with `animate-pulse rounded-md bg-accent`.
- [ ] **Label** · Cx 1 · deps: none
  Styled `<label>`. Blocks **Form** later.
- [ ] **Input** · Cx 1 · deps: none
  Styled `<input>`.
- [ ] **Textarea** · Cx 1 · deps: none
  Styled `<textarea>`.
- [ ] **AspectRatio** · Cx 1 · deps: none
  Wrapper `<div>` using the Tailwind `aspect-*` / CSS `aspect-ratio`.
- [ ] **Badge** · Cx 2 · deps: cva
  `<span>` + variants; export `BadgeClasses` (mirror `ButtonClasses`).
- [ ] **Card** · Cx 2 · deps: none
  `Card` / `CardHeader` / `CardTitle` / `CardDescription` / `CardAction` /
  `CardContent` / `CardFooter`.
- [ ] **Table** · Cx 2 · deps: none
  `Table` (+ overflow wrapper) / `TableHeader` / `TableBody` /
  `TableFooter` / `TableRow` / `TableHead` / `TableCell` / `TableCaption`.
  Blocks **DataTable**.
- [ ] **Breadcrumb** · Cx 2 · deps: none
  `Breadcrumb` / `BreadcrumbList` / `BreadcrumbItem` / `BreadcrumbLink` /
  `BreadcrumbPage` / `BreadcrumbSeparator` / `BreadcrumbEllipsis`.
- [ ] **Avatar** · Cx 2 · deps: none
  `Avatar` / `AvatarImage` / `AvatarFallback`. React detects image-load
  failure in JS; use `<img onerror>` to reveal the fallback, or accept the
  broken-image degradation.
- [ ] **Pagination** · Cx 2 · deps: **ButtonClasses** (done)
  `Pagination` / `PaginationContent` / `PaginationItem` /
  `PaginationLink` / `PaginationPrevious` / `PaginationNext` /
  `PaginationEllipsis`. Links styled via `ButtonClasses(ButtonGhost, …)`.

---

## Phase 2 — Tier 2 simple interactive

Native form controls and disclosure elements carry most of the behavior.

- [ ] **Progress** · Cx 2 · deps: none
  `<div>` track + inner bar width from a value param (or native
  `<progress>`).
- [ ] **Switch** · Cx 2 · deps: none
  Native `<input type="checkbox" role="switch">`, visual via CSS.
- [ ] **Toggle** · Cx 2 · deps: cva
  `<button aria-pressed>`; export `ToggleClasses` (the `toggleVariants`
  equivalent). Pressed state flips via inline `onclick` or a checkbox.
  Blocks **ToggleGroup**.
- [ ] **RadioGroup** · Cx 3 · deps: none
  `RadioGroup` (`<fieldset>`/`role=radiogroup`) + `RadioGroupItem`
  (`<input type="radio">`); roving focus is native to radios.
- [ ] **Checkbox** · Cx 3 · deps: none
  Styled `<input type="checkbox">`. Note: `indeterminate` is a JS-only
  property — needs a one-line script when used.
- [ ] **Collapsible** · Cx 3 · deps: none
  Native `<details>`/`<summary>`: `Collapsible` / `CollapsibleTrigger` /
  `CollapsibleContent`. Blocks **Accordion**.
- [ ] **Accordion** · Cx 3 · deps: **Collapsible** pattern
  Multiple `<details>`; single-mode uses the native `<details name="">`
  exclusive-group attribute.
- [ ] **Tabs** · Cx 3 · deps: none
  `Tabs` / `TabsList` / `TabsTrigger` / `TabsContent`. No-JS option:
  radio-input + CSS sibling selectors; or HTMX panel swap.
- [ ] **ToggleGroup** · Cx 3 · deps: **Toggle** / `ToggleClasses`
  `ToggleGroup` / `ToggleGroupItem`, single & multiple modes.
- [ ] **ScrollArea** · Cx 3 · deps: none
  `<div overflow-auto>` + CSS-styled scrollbars (drop Radix's custom
  scrollbar rendering).
- [ ] **Slider** · Cx 4 · deps: none
  Styled `<input type="range">`. Two-thumb range = two inputs + small
  script.
- [ ] **InputOTP** · Cx 4 · deps: none
  Segmented input; needs JS for per-segment focus management. Lowest
  priority of the phase.

---

## Phase 3 — Tier 3 floating (Popover API + CSS anchor positioning)

**Decision first:** build a shared floating layer on the native **Popover
API** (`popover` attribute + `popovertarget`) plus **CSS Anchor
Positioning** (`anchor-name` / `position-anchor`) — the native replacement
for Radix's Popper/Portal/DismissableLayer/FocusScope. Anchor positioning
is Chromium-shipped and progressing elsewhere; gate the rollout on that.

- [ ] **Popover infra** · Cx 3 · deps: none
  Shared trigger/content helpers: `popover`, `popovertarget`, anchor wiring,
  validated ids (reuse the `validateID` pattern from `alertdialog.go`).
- [ ] **Tooltip** · Cx 3 · deps: Popover infra
  Hover/focus-triggered; CSS hover + `popover="hint"`.
- [ ] **HoverCard** · Cx 3 · deps: Popover infra
  Hover-triggered popover with open/close delay.
- [ ] **Popover** · Cx 3 · deps: Popover infra
  Click-triggered: `Popover` / `PopoverTrigger` / `PopoverContent`.
  Blocks **DropdownMenu**, **Combobox**, **DatePicker**.
- [ ] **Select** · Cx 4 · deps: Popover infra
  Prefer the CSS-customizable native `<select>`; fall back to a
  Popover-based listbox. Blocks **DataTable**.
- [ ] **DropdownMenu** · Cx 4 · deps: **Popover**
  Menu markup + roving focus (arrow-key script); submenus.
  Blocks **ContextMenu**, **Menubar**, **DataTable**.
- [ ] **ContextMenu** · Cx 4 · deps: **DropdownMenu**
  Same menu behavior, opened at the cursor on `contextmenu`.
- [ ] **Menubar** · Cx 4 · deps: **DropdownMenu**
  Coordinated row of menus.
- [ ] **NavigationMenu** · Cx 4 · deps: Popover infra
  Viewport + active indicator; CSS-driven transitions.

---

## Phase 4 — Tier 4 modal / overlay (native `<dialog>`)

Reuse the `<dialog>` approach already proven in `alertdialog.go`
(top layer, `::backdrop`, focus trap, Escape-to-close — no framework).

- [ ] **Dialog** · Cx 3 · deps: none (shares `<dialog>` infra with AlertDialog)
  `Dialog` / `DialogTrigger` / `DialogContent` / `DialogHeader` /
  `DialogFooter` / `DialogTitle` / `DialogDescription` / `DialogClose`.
  Factor the shared `<dialog>` helpers out of `alertdialog.go`.
  Blocks **Sheet**.
- [ ] **Sheet** · Cx 4 · deps: **Dialog**
  `<dialog>` + slide-in `side` variants (top/right/bottom/left) via CSS.
  Blocks **Sidebar**.
- [ ] **Drawer** · Cx 5 · deps: **Dialog**
  `<dialog>` + drag-to-dismiss and snap points — requires JS; lowest
  priority of the phase.

---

## Phase 5 — Tier 5 composite / heavy

Each leans on a React-only library with no Go equivalent — behavior must be
reimplemented (server-side in Go, with HTMX, or with a small script).
Order within the phase is by dependency, then complexity.

- [ ] **Form** · Cx 5 · deps: **Label** (+ Input/Checkbox/RadioGroup/etc.)
  `Form` / `FormItem` / `FormLabel` / `FormControl` / `FormDescription` /
  `FormMessage`. Server-side validation display is natural here; consider
  wiring into the existing `html.ReflectFormComponents`.
- [ ] **Command** · Cx 5 · deps: none
  Filterable command list (cmdk). Filter server-side via HTMX or in a
  script. Blocks **Combobox**.
- [ ] **Combobox** · Cx 5 · deps: **Popover** + **Command** + Button
  Composition recipe, not a primitive.
- [ ] **Calendar** · Cx 5 · deps: Button
  Date grid — generate server-side with Go's `time` package; month nav via
  HTMX round-trip. Blocks **DatePicker**.
- [ ] **DatePicker** · Cx 5 · deps: **Popover** + **Calendar** + Button
  Composition recipe.
- [ ] **Carousel** · Cx 5 · deps: Button
  CSS `scroll-snap` covers most of it; drag/autoplay need a script.
- [ ] **Resizable** · Cx 5 · deps: none
  Drag handles — JS; CSS `resize` only covers trivial cases.
- [ ] **Toast** (Sonner) · Cx 5 · deps: none
  Queue/timers/swipe — JS, or HTMX out-of-band swaps for server-pushed
  toasts.
- [ ] **Chart** · Cx 6 · deps: none
  recharts wrapper — needs a Go SVG chart generator or a JS charting lib.
- [ ] **DataTable** · Cx 6 · deps: **Table** + **DropdownMenu** + Input +
  Checkbox + **Select** + **Pagination**
  Sorting/filtering/pagination done server-side via HTMX fits go-mx well.
- [ ] **Sidebar** · Cx 6 · deps: **Sheet** + Button + Input + Separator +
  Skeleton + **Tooltip**
  Most complex: ~20 sub-parts, collapse state with cookie persistence,
  keyboard shortcut, mobile (Sheet) vs desktop modes. Build last.

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
