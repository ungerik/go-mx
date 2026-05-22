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
  `fixed top-[50%] left-[50%] translate-x/y` (a native modal `<dialog>` is
  centered by the browser), `z-50` (the top layer is above everything), and
  `data-[state=open|closed]:animate-*` (a native `<dialog>` has no
  `data-state`). These are dropped; the box, sizing and layout classes are
  kept, plus `backdrop:bg-black/50`.

`AlertDialogTrigger` renders a `<button>`, so pass it content and styling, not
a nested `Button` (a `<button>` inside a `<button>` is invalid HTML). Use
`html.Class(shadcn.ButtonClasses(...))` for the button look:

```go
shadcn.AlertDialog(
    shadcn.AlertDialogTrigger("confirm-delete",
        html.Class(shadcn.ButtonClasses(shadcn.ButtonDestructive, shadcn.SizeDefault)),
        "Delete"),
    shadcn.AlertDialogContent("confirm-delete",
        shadcn.AlertDialogHeader(
            shadcn.AlertDialogTitle("Are you absolutely sure?"),
            shadcn.AlertDialogDescription(
                "This permanently deletes your account."),
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

## Tailwind CSS v4 is required

These components emit Tailwind v4 utility classes and nothing else. The
consuming application **must run Tailwind v4** over its templates and Go source
(so the class names are discovered) or the components render unstyled. This is
the most common "it doesn't work" cause.
