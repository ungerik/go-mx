# shadcn port — TODOS

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
  minus Radix-only positioning/animation classes. **new-york-v4 is frozen
  upstream since July 2026** — for new ports see
  [Upstream restructure (July 2026)](#upstream-restructure-july-2026--findings--porting-guide)
  below for where class strings now live and how to extract them.

## Status legend

Open ports are H3 items under [Open ports](#open-ports) carrying **What /
Why / Context / Effort / Priority**, matching the root `TODOS.md` format.
Finished ports keep the `[x]` checklist form under
[Completed](#completed), grouped by the phase they landed in so the
topological build order stays readable.

**Cx** = complexity 1–6, as rated above. It maps to the **Effort** field as
Cx 1–2 = S, Cx 3–4 = M, Cx 5 = L, Cx 6 = XL.

Priorities: **P0** blocking · **P1** critical, this cycle · **P2** important ·
**P3** nice-to-have · **P4** someday.

---

## Open ports

None: Chart landed in Phase 8 and Questionnaire in Phase 9, so every
component of the shadcn/ui registry is ported. New upstream components are
triaged here as they appear (see "Upstream sync").

---

## Calendar and DatePicker follow-ups

Select mode, `MonthURL` navigation, the dropdown caption and the reflected
`widget=datepicker` landed in Phase 8, two months side by side, year keys,
`ExcludeDisabled`, text days and the remaining recipes in Phase 9. Nothing
from react-day-picker is open; new gaps go here, most useful first.

## Chart follow-ups

Phase 9 added the pie, radar and radial kinds and the styled tooltip.
Nothing is open; new gaps go here, most useful first.

## Chat/AI follow-ups

Phase 9 added upstream's shimmer and scroll-fade utilities (opt-in), the
scroll-to-start button and the hidden button leaving the tab order. Nothing
is open; new gaps go here, most useful first.

## Hard dependency edges (must respect)

```
ButtonClasses ─► Pagination, Calendar, Carousel, Combobox, DatePicker
ToggleClasses ─► ToggleGroup
Collapsible ───► Accordion, Trace
StickToBottom ─► MessageScroller
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

---

## Upstream sync

What each port was last compared against. To catch up, diff upstream from
the recorded ref to its current head and port what changed; then update the
row.

| Port | Upstream | Last synced | How to diff |
|------|----------|-------------|-------------|
| `clsx/` | [lukeed/clsx](https://github.com/lukeed/clsx) | v2.1.1 (`e2dc836`) | `gh api repos/lukeed/clsx/releases` (unchanged since 2024) |
| `twmerge/` | [dcastil/tailwind-merge](https://github.com/dcastil/tailwind-merge) | tailwind-merge@3.7.0 (`511d68a`) plus the unreleased fix #721 (merged 2026-10-11) | `git diff tailwind-merge@3.7.0 <new> -- packages/tailwind-merge/src/lib packages/tailwind-merge/tests` |
| `cva/` `New` | [joe-bell/cva](https://github.com/joe-bell/cva) class-variance-authority | v0.7.1 (`45462dd`) | npm `class-variance-authority`; v0.7.1 is still the latest |
| `cva/` `Config.Composes` | joe-bell/cva `cva` beta | v1.0.0-beta.12 (`1d93698`) | `packages/cva/src/config.ts` and the "cva — composes" tests in `index.test.ts` |
| components from `new-york-v4` | shadcn-ui/ui `apps/v4/registry/new-york-v4/ui` | frozen; checked at `c0d63c3` (2026-10-31) | `gh api "repos/shadcn-ui/ui/commits?path=apps/v4/registry/new-york-v4/ui&since=<date>"` |
| components reconstructed from `style-vega.css` (Field, Kbd, Spinner, Empty, Item, ButtonGroup, InputGroup, NativeSelect, Marker, Message, Bubble, Attachment, Trace, CodeBlock, MessageScroller) | shadcn-ui/ui `apps/v4/registry/styles/style-vega.css` + `bases/base/ui` | `c0d63c3` (2026-10-31) | fetch the CSS at the old and new ref and `diff`; check the `.cn-<component>` rules of the ported components |

Policy for a sync:

- A changed `style-vega.css` rule of a component pinned to `new-york-v4`
  (Checkbox, RadioGroup, Switch, …) is **not** ported: those stay on their
  frozen class strings. The exception is a rule that only takes effect
  inside a reconstructed component, like the
  `group-has-[:focus-visible]/field-label:` classes the 2026-11-01 sync
  added to Checkbox, RadioGroupItem and Switch so a `FieldLabel` choice
  card draws its focus ring once.
- **Known upstream bugs the port fixes.** These classes diverge from
  `style-vega.css` at `c0d63c3` because upstream's do not work (details on
  `fieldLabelClasses` in `field.go`). On each sync, check whether upstream
  fixed them and return to its classes if it did:
  - the controls' `group-has-[:focus-visible]/field-label:` overrides, which
    the port writes `group-has-[>[data-slot=field]_:focus-visible]/field-label:`
    (upstream's never beat `focus-visible:ring-[3px]` and also match an
    inline `FieldLabel`);
  - the `FieldLabel` card hover, where the port adds `not-has-checked:`
    (upstream's hover tint hides a checked card's `has-checked:` tint).
- The `*-aria` rules belong to upstream's third primitive base, React Aria
  (shadcn-ui/ui#11208, July 2026). This port uses no primitive library, so
  they are ignored.
- **`cn` is not adopted.** Since 2026-09-04 (shadcn-ui/ui#11758) upstream
  imports `cn` from its own [`cn`](https://github.com/shadcn-ui/cn) package,
  a new merge engine that replaces clsx + tailwind-merge with the same API.
  `twmerge/` keeps porting tailwind-merge: it is still released, the Go port
  transcribes its config line by line, and `cn` is a different
  implementation at 0.x. `shadcn.Cn` has the same contract either way.
  Revisit if tailwind-merge stops being maintained or `cn` and
  tailwind-merge start resolving a class differently.

## Upstream restructure (July 2026) — findings & porting guide

Verified against the `shadcn-ui/ui` repo and the
[July 2026 changelog](https://ui.shadcn.com/docs/changelog/2026-07-base-ui-default)
on 2026-07-05. Everything below was confirmed by fetching actual upstream
sources (paths and commands included), so a future port can start from here
without re-analyzing upstream.

### What changed upstream

1. **Base UI is the default primitive library.** New shadcn projects and the
   docs default to [Base UI](https://base-ui.com); Radix remains fully
   supported (`shadcn init -b radix`), and "every update and new component
   will ship for both libraries (unless a component only exists in Base UI)".
   React's `asChild` prop became `render` — irrelevant to this port.
2. **Two parallel registries.** Canonical component sources moved to
   `apps/v4/registry/bases/base/ui/*.tsx` (Base UI) and
   `apps/v4/registry/bases/radix/ui/*.tsx` (Radix). Upstream keeps them in
   parity (`apps/v4/registry/bases/README.md`); verified: `switch.tsx` is
   byte-identical between the two apart from imports.
3. **Utility classes moved out of the components** (the bigger change for
   us). Component `.tsx` files now carry only a `cn-<component>-<part>`
   class plus skeleton positioning classes; the styling lives in eight named
   style sheets `apps/v4/registry/styles/style-{vega,nova,maia,lyra,mira,
   luma,rhea,sera}.css` as `@apply` rules scoped under `.style-<name>`,
   distributed via a new `shadcn/tailwind.css` package. `shadcn eject`
   inlines them back into the components. `vega` is listed first
   ("Clean, neutral, and familiar" — closest to the classic look);
   `nova` is "Reduced padding and margins".
4. **Our baseline is legacy but still published.** The registry this port
   transcribed its class strings from still exists, frozen, at
   `apps/v4/registry/new-york-v4/ui/*.tsx` (see `_legacy-styles.ts`).
5. **The shared style CSS is written against Base UI's data attributes** —
   bare boolean attributes (`data-open:`, `data-checked:`) instead of
   Radix's `data-[state=open]:` — even for the Radix registry.
6. **`data-slot` part names are unchanged.** E.g. Base UI's
   `DialogPrimitive.Backdrop` still gets `data-slot="dialog-overlay"`. Our
   `finish(e, "<slot>", …)` conventions stay aligned with upstream.

### Impact on this port

**Nothing breaks and no existing component needs changes.** This port never
used Radix or Base UI JS — behavior is native web platform, and Radix-only
selectors were already rewritten. The impact is entirely on *future* work:

- Existing components stay pinned to the frozen `new-york-v4` class strings
  (self-contained, no external CSS artifact). This is a deliberate pin.
- New components must be sourced from the new registry (some exist only
  there), which requires the class-string reconstruction below.
- Upstream `Form` exists only in the Radix registry; Base UI replaces it
  with `Field`.

### Porting guide for new components

**Step 1 — fetch the sources.** For component `<name>`:

```bash
# component skeleton (prefer the Base UI variant; parity with radix anyway)
gh api repos/shadcn-ui/ui/contents/apps/v4/registry/bases/base/ui/<name>.tsx \
  --jq .content | base64 -d

# the style layer (vega = reference style for this port)
gh api repos/shadcn-ui/ui/contents/apps/v4/registry/styles/style-vega.css \
  --jq .content | base64 -d | grep -A8 'cn-<name>'
```

**Step 2 — reconstruct the full class string per part.** Concatenate the
skeleton classes from the `.tsx` with the `@apply` list of the matching
`.cn-<component>-<part>` rule (drop the `cn-*` class name itself). This is
exactly what `shadcn eject` produces, i.e. the equivalent of the old
new-york-v4 one-string-per-part. Example (`switch`):

- tsx: `cn-switch peer group/switch relative inline-flex items-center …`
- css: `.cn-switch { @apply data-checked:bg-primary data-unchecked:bg-input
  … data-[size=sm]:h-[14px] …; }`
- port: skeleton + @apply classes merged into one string, then rewritten
  per the table below.

**Step 3 — rewrite Base UI data-attribute selectors to native ones.** Same
policy as the existing Radix `data-[state=*]` rewrites; the Base UI names
are just the bare-boolean spellings:

| Base UI selector      | Radix equivalent        | Native rewrite in this port           |
|-----------------------|-------------------------|---------------------------------------|
| `data-open:`          | `data-[state=open]:`    | `[open]`/`group-open:` on `<details>`/`<dialog>`; `aria-expanded:` on triggers; `:popover-open` on popovers |
| `data-closed:`        | `data-[state=closed]:`  | usually the unprefixed base state; often paired with enter/exit animations → drop |
| `data-checked:`       | `data-[state=checked]:` | `checked:` on native inputs           |
| `data-unchecked:`     | `data-[state=unchecked]:` | unprefixed base classes             |
| `data-highlighted:`   | `data-[highlighted]:`   | `focus:` (menu items receive focus)   |
| `data-disabled:`      | `data-[disabled]:`      | `disabled:` (native controls) or keep if we set the attribute ourselves |
| `data-starting-style:` / `data-ending-style:` | (Radix: `animate-in/out` pairs) | transition hooks for JS-driven mount/unmount → drop (same policy as Radix enter/exit animations) |
| `data-[side=…]:` etc. | same                    | keep — author-set attributes we render ourselves (like existing `data-size`, `data-variant`, `data-inset`) |

Keep the existing test convention: each `_test.go` asserts the
library-only selectors were rewritten/dropped (grep for `data-open:` /
`data-closed:` the way current tests grep for `data-[state=`).

**Step 4 — map Base UI structural parts to the native infra** (all already
proven in this port):

| Base UI part          | Native equivalent here                     |
|-----------------------|--------------------------------------------|
| `Portal`              | none needed — top layer via `popover` / `<dialog>` |
| `Backdrop`            | `::backdrop` (dialogs) or overlay `<div>`; keeps `data-slot="…-overlay"` |
| `Positioner` + `Popup`| `popover` + CSS anchor positioning helpers in `popover.go` |
| `Trigger render={…}`  | N/A — Go composition instead of React Slot |
| `IconPlaceholder` (multi icon lib) | inline lucide SVG from `icons.go` |

**Step 5 — everything else is unchanged**: signature conventions,
`finish(e, "<slot>", base)`, `data-slot` names (upstream kept them),
`…Classes` helpers, gallery example + regeneration
(`go run ./cmd/shadcn-gallery …`, see CLAUDE.md).

Note: the new style sheets use a few `cn-` utility tokens beyond
component parts (e.g. `cn-font-heading` on titles) — resolve them from the
style CSS / `shadcn/tailwind.css` the same way, or substitute the concrete
classes they expand to.

---

## Completed

Ported components, grouped by the phase they landed in. The phase order is
the topological build order — every dependency lands before its dependents
— so it is preserved here rather than flattened into one list. See [Hard
dependency edges](#hard-dependency-edges-must-respect) for the edges these
phases satisfy.

### Phase 0 — Foundations

- [x] **clsx** — `clsx` npm-package port; `Join` (`clsx/` subpackage)
- [x] **twmerge** — `tailwind-merge` v3.7.0 port; `Merge` (`twmerge/` subpackage)
- [x] **cva** — `class-variance-authority` port (`cva/` subpackage)
- [x] **Cn** — thin `cn` helper composing clsx + twmerge (`cn.go`)
- [x] **finish** — shared class/attrib merge (`component.go`)
- [x] **Button** · Cx 2 — exports `ButtonClasses`
- [x] **Alert** · Cx 2 — `Alert` / `AlertTitle` / `AlertDescription`
- [x] **AlertDialog** · Cx 4 — native `<dialog>`, full sub-part set

### Phase 1 — Tier 1 structural (no deps, pure markup)

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

### Phase 2 — Tier 2 simple interactive

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

### Phase 3 — Tier 3 floating

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

### Phase 4 — Tier 4 modal / overlay (native `<dialog>`)

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

### Phase 5 — Tier 5 composite / heavy

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
- [x] **Combobox** · Cx 5 · deps: **Popover** + **Command** + Button
  Composition recipe (Popover trigger + filterable Command in PopoverContent) —
  shipped as a gallery example, matching how shadcn ships Combobox (copy-paste,
  not a primitive).
- [x] **Calendar** · Cx 5 · deps: Button
  Single-month grid generated server-side with Go's `time` package; selected
  day marked with `data-selected` (plus `aria-pressed` on a static calendar,
  `aria-selected` on the gridcell in select mode). `Calendar(month, selected, …)`. Prev/Next are
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

### Phase 6 — new upstream components (post-restructure)

Components in `apps/v4/registry/bases/base/ui/` with no counterpart in
this port when the restructure was surveyed (2026-07-05). All were triaged
on 2026-07-07 with the guide above; the four that remain unported moved to
[Open ports](#open-ports).

- [x] **Kbd** · Cx 1 — `Kbd` / `KbdGroup` (both `<kbd>` like upstream).
- [x] **Spinner** · Cx 1 — inline lucide loader-circle SVG (`icons.go`
  pattern) with `role="status"` + `aria-label`; `animate-spin` is the whole
  behavior.
- [x] **Empty** · Cx 2 — `Empty` / `EmptyHeader` / `EmptyMedia(variant)` /
  `EmptyTitle` / `EmptyDescription` / `EmptyContent`. Media slot is
  `data-slot="empty-icon"` (upstream quirk, kept). `cn-font-heading`
  dropped (theme font var not in the frozen baseline).
- [x] **Item** · Cx 2 — `Item(variant, size)` + Group/Separator/Media
  (variant)/Content/Title/Description/Actions/Header/Footer. The React
  `render` prop is Go composition (always a `<div>`).
- [x] **ButtonGroup** · Cx 2 — `ButtonGroup(orientation)` + Text/Separator;
  exports `ButtonGroupClasses` (upstream exports buttonGroupVariants).
  First real Base UI selector rewrite: bare `data-horizontal:`/
  `data-vertical:` → `data-[orientation=…]:` on the separator.
- [x] **InputGroup** · Cx 3 — `InputGroup` / `InputGroupAddon(align)` /
  `InputGroupButton(variant, size)` (own xs/sm/icon-xs/icon-sm scale,
  ghost default) / `InputGroupText` / `InputGroupInput` /
  `InputGroupTextarea` (borderless [Input]/[Textarea] variants, slot
  `input-group-control`). Addon keeps upstream's click-to-focus as a
  default inline `onclick`. `InputGroupText` emits a data-slot upstream
  lacks (consistency divergence, documented in the func comment).
- [x] **Field** · Cx 3 · deps: Label, Separator — **ported 2026-07-07 as
  triaged** (`field.go`): full part set below, plus a `FieldLabelFor`
  shortcut. `FormItem`/`FormLabel`/`FormLabelFor`/`FormDescription`/
  `FormMessage` removed (no aliases); the bare `Form` (`<form>` wrapper)
  stays. Rewrites applied: `has-data-checked:` → `has-checked:`,
  `[role=checkbox],[role=radio]` → our `data-slot` attributes,
  `group-has-data-horizontal/field:` →
  `group-data-[orientation=horizontal]/field:`. Original triage:
  Parts: `FieldSet` (`<fieldset>`) / `FieldLegend(variant legend|label)` /
  `FieldGroup` / `Field(orientation vertical|horizontal|responsive)` /
  `FieldContent` / `FieldLabel` / `FieldTitle` / `FieldDescription` /
  `FieldSeparator` / `FieldError` (`role="alert"`; take children — the
  React `errors` dedup prop is caller-side in Go). Mapping: `FormItem` →
  `Field`, `FormLabel` → `FieldLabel`, `FormDescription` →
  `FieldDescription`, `FormMessage` → `FieldError`; error state moves from
  `data-error` on the label to author-set `data-[invalid=true]` on the
  `Field` root (keep, author-set). When Field lands, **remove**
  `FormItem`/`FormLabel`/`FormLabelFor`/`FormDescription`/`FormMessage`
  (pre-release, no deprecation aliases; only the gallery `form.go` example
  uses them) and decide whether the bare `Form` (`<form>` wrapper) earns
  its keep vs `html.Form`. Rewrites: `has-data-checked:` (Base UI checkbox
  attr) → `has-checked:` (our native inputs); `group-data-[disabled=true]/
  field:` stays (author-set). `responsive` orientation uses container
  queries (`@container/field-group`, `@md/field-group:`) — plain Tailwind
  v4, keep. Classes at `style-vega.css` `.cn-field*`.
- [x] **Combobox** — **decided 2026-07-07: keep the Popover+Command gallery
  recipe, no primitive port.** Upstream's combobox.tsx (~15 parts:
  Value/Trigger/Clear/Input/Content/List/Item/Group/Label/Collection/Empty/
  Separator/…) is a pure styling wrapper over Base UI's Combobox JS — the
  entire behavior (filtering, selection model, chips) lives in the Base UI
  runtime this port doesn't use. Porting the part set would mean writing a
  large custom script that duplicates what our Command filter + Popover
  already provide. Revisit only if a concrete need (e.g. multi-select chips)
  outgrows the recipe; a native `<input list>`/`<datalist>` variant is the
  zero-JS alternative to explore first.
- [x] **Direction** — **resolved 2026-07-07: N/A, nothing to port.** A
  React context provider that only informs Base UI JS positioning; in
  server-rendered HTML the native mechanism is the `dir` attribute
  (`html.DirRTL` on any ancestor).

### Phase 7 — chat/AI set and NativeSelect (2026-11-01)

Reconstructed from `bases/base/ui/*.tsx` + `style-vega.css` at `c0d63c3`.
What none of them reproduces: upstream's `shimmer` and `scroll-fade-*`
utilities, which live in shadcn's own `tailwind.css` (see "Shimmer and
scroll-fade come from shadcn's own stylesheet"). Where upstream's React
`render` prop picks the element (Marker, BubbleContent,
AttachmentTrigger), an `href` attribute renders an `<a>` and a `type`
attribute a `<button>` (`renderTag` in `component.go`).

- [x] **NativeSelect** · Cx 2 — `NativeSelect(size)` /
  `NativeSelectOption(value)` / `NativeSelectOptGroup(label)`. As upstream
  does with className, a caller class goes to the wrapper, every other
  attribute to the `<select>`.
- [x] **Marker** · Cx 1 — `Marker(variant default|separator|border)` /
  `MarkerIcon` / `MarkerContent`.
- [x] **Message** · Cx 2 — `MessageGroup` / `Message(align)` /
  `MessageAvatar` / `MessageContent` / `MessageHeader` / `MessageFooter`.
  One `ChatAlign` type serves Message, Bubble and BubbleReactions. The
  header/footer ghost inset is scoped to a ghost bubble, as this port's
  Button always emits `data-variant` (the Item lesson).
- [x] **Bubble** · Cx 2 — `BubbleGroup` / `Bubble(variant ×7, align)` /
  `BubbleContent` / `BubbleReactions(side, align)`.
- [x] **Attachment** · Cx 2 — `Attachment(state, size, orientation)` +
  Group/Media(variant)/Content/Title/Description/Actions/Action/Trigger.
  Shimmer → `animate-pulse`; `scrollbar-none` → arbitrary scrollbar
  classes.
- [x] **Trace** · Cx 2 — `Trace` / `TraceLine(status)` / `TraceIcon` /
  `TraceTitle` / `TraceDescription` / `TraceCollapsible` (our `<details>`
  Collapsible) / `TraceCollapsibleTrigger` / `TraceCollapsibleContent`.
  The chevron's `group-aria-expanded/…:rotate-90` became
  `[details[open]>summary>&]:rotate-90`, so a nested collapsible does not
  turn it; the row's `in-[a,button]` hover selectors gained `summary`.
- [x] **CodeBlock** · Cx 3 — `CodeBlock(variant, size)` + Header/Title/
  Actions/CopyButton(value)/Content/Line(variant)/Footer, plus the
  `CodeBlockWrap`, `CodeBlockLineNumbers` and `CodeBlockStartLine`
  attributes for upstream's props. Decided: CodeBlock takes highlighted
  output as children and does not import `highlight`. The copy button
  emits one shared script (`mxCodeBlockCopy`); its value is
  percent-encoded because go-mx's attribute escaping turns line breaks
  into spaces. The `<code>` is rendered unindented, so an indenting
  writer cannot add whitespace inside the `<pre>`, and loose children
  (highlighted tokens) are wrapped in one span so the line grid does not
  give each token a row.
- [x] **MessageScroller** · Cx 3 (rated 4 at triage, before
  `StickToBottom` existed) — `MessageScroller` / `MessageScrollerViewport`
  / `MessageScrollerContent` / `MessageScrollerItem` /
  `MessageScrollerButton`. The viewport always carries `StickToBottom`,
  and the button shows from its `data-stuck` with CSS, so the port adds
  no script. Porting it found a StickToBottom bug: a scroll event caused
  by layout (the stylesheet arriving) unstuck the area before it was
  read; such an event now re-pins.

### Phase 8 — Chart and calendar navigation (2026-11-01)

- [x] **Chart** · Cx 6 — `Chart(kind bar|line|area, ChartData, ChartOptions)`,
  option A of the triage: the SVG is drawn in Go (scales, d3-style round
  ticks, grid, grouped and stacked bars, linear/monotone/step lines and
  stacked areas), colored through upstream's `--color-<key>` variables from
  the theme's `--chart-1…5`. Decided: tooltips are native `<title>`s, the
  SVG is `aria-hidden` and a visually hidden table carries the data, and the
  SVG scales through its viewBox in an `aspect-video` container. Pie, radar,
  radial and the styled tooltip are follow-ups.
- [x] **Calendar month navigation** — `CalendarOptions.MonthURL`: the month
  buttons, PageUp/PageDown and an arrow key past the grid fetch the month's
  URL and swap in the first calendar of the response, without htmx. The
  picked days live in the hidden inputs (with their `DayLabel` in
  `data-label`) and are marked again in the new month, so the handler need
  not know them; the focus returns to the button or the day the keys moved
  to.
- [x] **Dropdown caption** — `CalendarOptions.CaptionLayout` with
  `MonthName` and `Labels.CalendarMonth`/`CalendarYear`: transparent native
  selects over the shown label, as react-day-picker draws them, each option
  carrying its month's URL.
- [x] **Range and multiple selection** — `CalendarOptions.Mode`,
  `RangeEnd`, `EndName`, `Selected`, with upstream's range classes on the
  day buttons and cells (logical sides, so RTL joins up), and
  `aria-multiselectable`. `CalendarText` joins a range with " – " and
  several days with ", ", and restores its placeholder when emptied.
- [x] **Day names** — today's name is wrapped in `Labels.CalendarToday`, a
  selected day's in select mode in `Labels.CalendarSelected`, and the script
  renames the days as the selection moves.
- [x] **`form:"widget=datepicker"`** — the core reads it as `date`; shadcn's
  `FieldDecider` renders the Popover + Calendar recipe, configured per field
  with `ContextWithDatePickerOptions` and answered by `DatePickerCalendar`.
- [x] **Browser tests for the inline scripts** — `internal/headless` runs a
  page in the system's Chrome with `--dump-dom` (no Go dependency); the page
  records its findings into the DOM. Covers the calendar (pick, keys,
  navigation, RTL, range, multiple), the CodeBlock copy button and
  MessageScroller. CI sets `MX_REQUIRE_CHROME` so a missing Chrome fails
  instead of skipping.
- [x] **Highlighted code with line numbers** — `highlight.Highlighter.Lines`
  cuts the tokens at the line breaks, a multi-line token into one span per
  line, for one `CodeBlockLine` per line.

### Phase 9 — Questionnaire and the P4 follow-ups (2026-11-01)

- [x] **Questionnaire** · Cx 5 — `Questionnaire` / `QuestionnaireProgress` /
  `QuestionnaireItem(name)` / `QuestionnaireTitle` /
  `QuestionnaireDescription` / `QuestionnaireChoices(shortcuts)` /
  `QuestionnaireChoice(type, name, value)` /
  `QuestionnaireChoiceDescription` / `QuestionnaireInput` /
  `QuestionnaireError` / `QuestionnaireActions` / `QuestionnairePrevious` /
  `QuestionnaireSkip` / `QuestionnaireNext` / `QuestionnaireSubmit`, from
  `bases/base/ui/questionnaire.tsx` at `c257f68` and `style-vega.css` at
  `4233ed1`. Decided: shape (a), every step a round trip posting
  `questionnaire-item` and `questionnaire-action`; the server validates and
  answers with the next item. One shared script (`mxQuestionnaireKey`) does
  Enter→Next and the letter/number shortcuts. Base UI's
  `data-checked`/`data-disabled` on the card → `has-[>input:checked|disabled]:`,
  `data-invalid` → `group-data-[invalid=true]/questionnaire-item:`. The
  indicator's dot or check is picked in Go from the type. Not ported:
  arrow-key navigation between items, and clearing a choice when the user
  types a freeform answer (both are posted). Captions are children.
- [x] **Two months side by side** — `CalendarOptions.NumberOfMonths`: the
  grids share the buttons, at the outer ends, the selection and the keys;
  outside days are left out so no day is shown twice. `calendarKey` now moves
  by date rather than by grid index, which makes that work.
- [x] **Year keys** — Shift+PageUp/PageDown through the month buttons'
  `data-year-url`, clamped to StartMonth/EndMonth.
- [x] **`ExcludeDisabled`** — a range's second click that would span a
  disabled day of the shown grid starts a new range.
- [x] **Class merging per calendar day** — `calendarDayClasses` caches
  `Cn` per distinct class string: a month went from ~290 µs to ~82 µs.
- [x] **Static days are text** — as react-day-picker without a mode: the
  number, and the DayLabel (with the today and selected labels) as
  screen-reader text; the table is named by its month. No Tab stops, no
  `aria-pressed`.
- [x] **DatePicker recipes** — Basic, Time picker, Input (a typed date
  through the new `calendarPick` and an `<input data-calendar-text>`), With
  presets (`calendarPick` from a select) and RTL (Arabic); the Range recipe
  shows two months.
- [x] **Pie, radar and radial charts** — `ChartPie` (with `Donut`),
  `ChartRadar` and `ChartRadial`, colored per category through
  `ChartData.CategoryColors` or the theme, with category legends and a
  square default viewBox.
- [x] **Styled chart tooltip** — `ChartOptions.Tooltip`: a hover band per
  category with upstream's tooltip card, shown by `group-hover` alone.
- [x] **Shimmer and scroll-fade** — `UtilitiesCSS` embeds upstream's
  `@utility` rules (`packages/shadcn/src/tailwind.css` at `a9c1da4`), and
  `UseUtilities` makes TraceLine, Attachment, AttachmentGroup,
  CodeBlockContent and MessageScrollerViewport emit them instead of
  `animate-pulse` and nothing.
- [x] **Scroll-to-start button** — `MessageScrollerButton(…,
  MessageScrollerStart)`, shown while StickToBottom's new `data-at-start` is
  missing (`Labels.MessageScrollerStartButton`).
- [x] **Hidden buttons leave the tab order** — both scroller buttons add
  `invisible` to their hidden state and transition `visibility`
  (`transition-discrete`), so they fade out before leaving the tab order.

### Decided — stay on inlined utility strings (2026-07-07)

- [x] **Decided 2026-07-07: status quo — inlined utility strings**, not
  upstream's `cn-*` classes + shipped CSS file. Rationale: the zero-asset
  model is a core property of this port (components are self-contained Go
  source, no CSS artifact to distribute or version); `Cn`/twmerge cannot
  resolve conflicts hidden inside `cn-*` rules, which would break the
  caller-class-wins contract of `finish`; and `shadcn eject` proves the
  inlined form remains first-class upstream. New ports reconstruct the
  full class string per part (guide Step 2) — the equivalent of what
  `eject` produces. Cost accepted: pinned to one look (frozen new-york-v4
  ≈ vega-ish), no swappable nova/maia/… style variants. Revisit only if
  upstream stops publishing the style CSS in a reconstructable form.
  Consequence for tests: emitted class strings must never contain a
  `cn-` token — resolve/expand any `cn-*` utility (e.g. `cn-font-heading`)
  during reconstruction.

### Follow-ups from the Field port (2026-07-07)

- [x] **Migrate `FieldDecider` (`formdecider.go`) to the Field system**
  (done 2026-07-07). `shadcnField`/`shadcnBoolField` now wrap fields in a
  [Field] root, label with [FieldLabel], help with [FieldDescription] and
  each message with [FieldError] (`role="alert"`); the Field root gets
  `data-invalid="true"` on error while the control keeps its own
  `aria-invalid`. The `data-error=path` hook the html layer uses is
  preserved on each FieldError. Reflected and hand-written forms now render
  the same Field markup. (Surfaced by the /ship cross-model adversarial
  review.)

### Grid semantics for screen readers

**What:** Render the select-mode grid as react-day-picker does:
`role="grid"` on the table with the month as its label, and `aria-selected`
on a `role="gridcell"` instead of on the `<button>`.

**Why:** `aria-selected` is not an allowed attribute of a button, so a
screen reader may not announce the picked day, and the grid pattern tells it
that the arrow keys work.

**Context:** The static grid predates select mode and uses the same
markup. Moving `aria-selected` changes the `aria-selected:` class
selectors on the day button too.

**Effort:** S
**Priority:** P3
**Depends on:** None
**Completed:** 2026-10-19 — in select mode the calendar is a `role="grid"`
table named by its month with `aria-selected` on the gridcell; the button
keeps `data-selected` for the look in both modes, and the static calendar,
which has no keyboard handling, stays a plain table.
