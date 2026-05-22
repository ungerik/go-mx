# Research on shadcn/ui components

The `shadcn/` package here currently only has the `cn()` class-merging utility — no actual components yet. So I'll rate complexity through the lens that matters for go-mx: **porting shadcn/ui components to server-side Go HTML generation**. That changes the picture significantly from the React originals, so I'll cover both.

## The fundamental split

shadcn components divide into two camps that dominate the complexity rating:

- **Structural** — styled HTML, no runtime behavior. These port ~1:1 to go-mx: emit elements + Tailwind classes, let `cn()` merge them. Genuinely easy.
- **Behavioral** — wrap Radix Primitives for focus traps, Floating-UI positioning, portals, roving keyboard nav, scroll lock, drag. A Go server can emit the markup + ARIA, but **none of the behavior runs without client JS**. For these, "implementing" in go-mx means also picking a JS strategy (HTMX round-trips, Alpine.js, or a small custom runtime).

## Complexity tiers (least → most complex)

### Tier 1 — Structural, pure port (1/5)

| Component    | Cplx | Depends on     | Notes                            |
| ------------ | ---- | -------------- | -------------------------------- |
| Separator    | 1    | —              | `<hr>`/div; Radix only adds ARIA |
| Skeleton     | 1    | —              | One animated div                 |
| Label        | 1    | —              | Styled `<label>`                 |
| Input        | 1    | —              | Styled `<input>`                 |
| Textarea     | 1    | —              | Styled `<textarea>`              |
| Aspect Ratio | 1    | —              | Padding-trick wrapper            |
| Badge        | 2    | cva            | Needs a variant helper           |
| Button       | 2    | cva            | Exports `buttonVariants`, reused |
| Alert        | 2    | cva            | Composition + variants           |
| Card         | 2    | —              | 6 sub-parts (Header/Title/etc.)  |
| Table        | 2    | —              | ~8 styled table sub-parts        |
| Breadcrumb   | 2    | —              | Composition of styled parts      |
| Avatar       | 2    | —              | Image-load fallback needs JS     |
| Pagination   | 2    | buttonVariants | Reuses Button styling            |

### Tier 2 — Simple interactive, single primitive (2/5)

| Component    | Cplx | Depends on        | Notes                            |
| ------------ | ---- | ----------------- | -------------------------------- |
| Toggle       | 2    | cva               | Exports `toggleVariants`         |
| Switch       | 2    | —                 | Controlled on/off                |
| Checkbox     | 3    | —                 | Indeterminate state              |
| Radio Group  | 3    | —                 | Roving focus across group        |
| Progress     | 2    | —                 | Value-driven width               |
| Collapsible  | 3    | —                 | Open/close + height animation    |
| Toggle Group | 3    | toggleVariants    | Roving focus + Toggle styling    |
| Accordion    | 3    | Collapsible behav | Single/multiple mode             |
| Tabs         | 3    | —                 | Roving focus, panel switching    |
| Scroll Area  | 3    | —                 | Custom scrollbar rendering       |
| Slider       | 4    | —                 | Drag; range mode adds complexity |
| Input OTP    | 4    | input-otp lib     | Segmented input, focus mgmt      |

### Tier 3 — Floating / portal-based (3/5)

All share one infrastructure layer: **Popper (Floating UI) + Portal + DismissableLayer + FocusScope**. Build that once.

| Component       | Cplx | Depends on          | Notes                            |
| --------------- | ---- | ------------------- | -------------------------------- |
| Tooltip         | 3    | Popper infra        | Hover/focus + delay              |
| Hover Card      | 3    | Popper infra        | Hover + delay + portal           |
| Popover         | 3    | Popper infra        | Click + focus + dismiss          |
| Dropdown Menu   | 4    | Popover infra       | Menu roving, typeahead, submenus |
| Context Menu    | 4    | Dropdown Menu behav | Right-click triggered            |
| Select          | 4    | Popover infra       | Listbox + scroll-to-selected     |
| Menubar         | 4    | Dropdown Menu behav | Coordinated multi-menu           |
| Navigation Menu | 4    | Popper infra        | Viewport, indicator, motion      |

### Tier 4 — Modal / overlay (4/5)

Shared layer: **focus trap + scroll lock + portal + overlay lifecycle**.

| Component    | Cplx | Depends on  | Notes                           |
| ------------ | ---- | ----------- | ------------------------------- |
| Dialog       | 4    | Modal infra | Base for the next three         |
| Alert Dialog | 4    | Dialog      | Forced action, focus on Cancel  |
| Sheet        | 4    | Dialog      | Dialog + slide-in side variants |
| Drawer       | 5    | vaul lib    | Drag-to-dismiss, snap points    |

### Tier 5 — Composite / external-lib-heavy (5/5)

These are mostly **recipes** (compositions), not single primitives, and each leans on a React-only library with no Go equivalent — so in go-mx you'd reimplement the behavior yourself.

| Component    | Cplx | Depends on                                                                 | Notes                                                                                    |
| ------------ | ---- | -------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------- |
| Command      | 5    | cmdk lib                                                                   | Fuzzy filter + keyboard nav                                                              |
| Combobox     | 5    | Popover + Command + Button                                                 | A recipe, not a primitive                                                                |
| Calendar     | 5    | react-day-picker                                                           | Date grid, ranges, many edge cases                                                       |
| Date Picker  | 5    | Popover + Calendar + Button                                                | Recipe                                                                                   |
| Carousel     | 5    | embla-carousel + Button                                                    | Slides, drag, autoplay                                                                   |
| Resizable    | 5    | react-resizable-panels                                                     | Drag handles, panel persistence                                                          |
| Sonner/Toast | 5    | sonner lib                                                                 | Queue, timers, swipe, portal                                                             |
| Form         | 5    | react-hook-form + zod + Label                                              | Field context + validation wiring                                                        |
| Chart        | 6    | recharts                                                                   | Config, theming, tooltip/legend                                                          |
| Data Table   | 6    | tanstack-table + Table + Dropdown + Input + Checkbox + Select + Pagination | Recipe of ~7 components                                                                  |
| Sidebar      | 6    | Sheet + Button + Input + Separator + Skeleton + Tooltip                    | Most complex; ~20 sub-parts, cookie persistence, keyboard shortcut, mobile/desktop modes |

## The interdependency graph

Foundations (build first — not "components"):

```
cn()  ──────────────► everything
cva (variant helper) ─► Button, Badge, Alert, Toggle, Sheet …
buttonVariants ───────► Pagination, Calendar, Carousel, AlertDialog, Combobox, DatePicker
toggleVariants ───────► Toggle Group
```

Component → component edges (source actually imports source):

```
Popper/Portal infra ─► Tooltip, HoverCard, Popover, Select, NavigationMenu
Popover ─────────────► DropdownMenu ─► ContextMenu, Menubar
Dialog ──────────────► AlertDialog, Sheet
Sheet ───────────────► Sidebar (mobile mode)
Command ─────────────► Combobox, CommandDialog (Command + Dialog)
Calendar ────────────► DatePicker (+ Popover + Button)
Table ───────────────► DataTable (+ tanstack + Dropdown + Input + Checkbox + Select + Pagination)
Label ───────────────► Form (FormLabel)
```

`Sidebar` is the convergence point — it pulls in the most other components.

## Recommended build order for go-mx

Topologically sorted by the graph above:

1. **`cva`-equivalent variant helper** — a Go struct/func mapping variant keys → class strings, paired with the existing `cn()`. Unblocks all Tier 1–2.
2. **Tier 1 structural** — Label, Input, Textarea, Button (+`buttonVariants`), Badge, Card, Alert, Separator, Skeleton, Table, Avatar, Breadcrumb, Aspect Ratio, Pagination. Pure markup; immediate, low-risk wins.
3. **Decide the JS interactivity strategy** — this is the real fork in the road. HTMX (your `hx/` package) handles server round-trips well (open a dialog by swapping server-rendered content), but client-only state (accordion toggle, dropdown open) wants Alpine.js or a tiny custom runtime. Everything below depends on this decision.
4. **Tier 2** — Toggle/Toggle Group, Checkbox, Radio Group, Switch, Progress, Collapsible → Accordion, Tabs, Slider, Scroll Area.
5. **Shared floating layer, then Tier 3** — Tooltip, Hover Card, Popover → Dropdown Menu → Context Menu, Menubar, Select, Navigation Menu.
6. **Shared modal layer, then Tier 4** — Dialog → Alert Dialog, Sheet, Drawer.
7. **Tier 5 last** — Command → Combobox; Calendar → Date Picker; Form, Toast, Carousel, Resizable, Data Table, Chart; **Sidebar dead last**.

One honest caveat: Tier 3–5 components depend on React libraries (cmdk, embla, react-day-picker, tanstack-table, recharts, vaul, react-hook-form) that have no Go counterpart. For those, go-mx can only generate the static markup + ARIA — the actual behavior must be reimplemented in JS or accepted as a client-side dependency. That ceiling, more than raw complexity, is what should drive how far you take the port.