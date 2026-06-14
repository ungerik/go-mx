---
title: shadcn how-to guides
---

# shadcn how-to guides

Task-oriented recipes. Each assumes you've done the
[tutorial](tutorial.html) or are already comfortable rendering a go-mx page.

- [Set up Tailwind v4 for production](#set-up-tailwind-v4-for-production)
- [Override or extend a component's classes](#override-or-extend-a-components-classes)
- [Wire a component to htmx](#wire-a-component-to-htmx)
- [Build a confirm dialog](#build-a-confirm-dialog)
- [Give a non-button the button look](#give-a-non-button-the-button-look)
- [Export a page to static HTML](#export-a-page-to-static-html)

---

## Set up Tailwind v4 for production

The CDN build from the tutorial recompiles in the browser on every load. For
production, compile a static CSS file once with the Tailwind v4 CLI. This is the
single most common "my components are unstyled" fix.

### Prerequisites

- Node.js (only to run the Tailwind CLI; your app stays pure Go)
- A `theme.css` with the shadcn tokens (see the tutorial)

### Steps

1. Create `input.css`:

   ```css
   @import "tailwindcss";
   @import "./theme.css";
   ```

2. Compile, scanning your project for the classes the Go components emit:

   ```bash
   npx @tailwindcss/cli@4 -i input.css -o public/app.css --minify
   ```

   Tailwind v4 auto-detects source files in the project (including your `.go`
   files), so the class names baked into the components are discovered.

3. Link the compiled stylesheet from your document instead of the CDN script:

   ```go
   page := html.NewDocument("My app", body)
   page.Stylesheets = []string{"/app.css"}
   ```

   Serve `public/` as static files (`http.FileServer`).

### Verification

Load a page with a `shadcn.Button` on it. It should be fully styled with **no**
flash of unstyled content (the CSS is precompiled, not compiled in the browser).

### Troubleshooting

- **Still unstyled:** the CLI didn't see your classes. Make sure you run it from
  the project root so it scans your `.go` files, and re-run it after adding new
  components (it's a build step, not a watcher, unless you pass `--watch`).
- **Colors look wrong / missing:** `theme.css` isn't imported. The components
  reference `--background`, `--primary`, etc.; without the tokens they fall back
  to nothing.

---

## Override or extend a component's classes

Every component carries its own base classes. To add or override, pass
`html.Class("…")` alongside the children — go-mx merges all `class` values into
one, and **your classes win** any Tailwind conflict.

```go
// Add layout classes:
shadcn.Button(shadcn.ButtonDefault, shadcn.SizeDefault, html.Class("w-full"), "Save")

// Override a base class — your rounded-none beats the component's rounded-md:
shadcn.Input(html.Class("rounded-none"), html.Type("text"))
```

For computed or conditional classes (go-mx attributes have no `clsx`-style object
form), build the string yourself with `Cn` and pass the result:

```go
cls := shadcn.Cn(
    "transition-colors",
    map[string]bool{"bg-destructive text-white": hasError},
)
shadcn.Card(html.Class(cls), /* … */)
```

`Cn` flattens its arguments and resolves Tailwind conflicts (later wins), exactly
like shadcn/ui's `cn`.

### Verification

Inspect the rendered element: there is exactly **one** `class` attribute, and
your override appears in place of the base class it conflicts with.

---

## Wire a component to htmx

`Toggle`, `TabsTrigger`, and `ToggleGroupItem` ship a default inline `onclick`
that flips their state client-side. If you pass **any** `hx-*` attribute, the
component drops that default and lets [htmx](https://htmx.org) drive instead — the
signature is unchanged.

### Steps

1. Load htmx in your document `<head>` (CDN or self-hosted).
2. Pass `hx.*` attributes from `github.com/ungerik/go-mx/hx`:

   ```go
   import "github.com/ungerik/go-mx/hx"

   // Server returns the re-rendered button with aria-pressed flipped.
   shadcn.Toggle("", "",
       hx.Post("/toggle-bold"), hx.Swap("outerHTML"),
       "Bold",
   )
   ```

3. On the server, handle the route and return the re-rendered component.

### Verification

Click the toggle with htmx loaded: a network request fires to your endpoint and
the swapped-in markup replaces the button. Without any `hx-*` attribute, the same
`Toggle` flips `aria-pressed` purely client-side.

### Troubleshooting

- **Both the default onclick *and* htmx fire:** you passed the `hx-*` attribute as
  a raw string the component doesn't recognize as `hx-*`. Use the `hx` package
  helpers so the opt-out detection (`hasHX`) sees them.

---

## Build a confirm dialog

`AlertDialog` uses the native HTML `<dialog>` element — no JavaScript framework.
The trigger and content are linked by a developer-chosen id string, not by DOM
nesting.

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

### Verification

Click the trigger: a centered modal opens in the browser's top layer with a dimmed
backdrop. `Cancel` and `Continue` both close it (they're submit buttons inside a
`<form method="dialog">`) and set `dialog.returnValue` to `"cancel"` / `"confirm"`.

### Troubleshooting

- **Panic at startup:** the `dialogID` must be a non-empty string of letters,
  digits, `-`, and `_`. Pass a constant id, never user input — it's interpolated
  into an `onclick` and an `id`.
- **`Continue` doesn't do anything server-side:** by default it only closes the
  dialog. Add `html.Attrib("formaction", "/url")` + `html.FormMethodPOST`, or
  attach `html.OnClick`.

---

## Give a non-button the button look

`ButtonClasses(variant, size)` returns just the merged class string — the
equivalent of shadcn/ui's exported `buttonVariants`. Use it to style any element
like a button (a link, an `AlertDialogTrigger`) without nesting a real
`<button>`.

```go
html.A(html.HRef("/dashboard"),
    html.Class(shadcn.ButtonClasses(shadcn.ButtonDefault, shadcn.SizeDefault)),
    "Open dashboard",
)
```

`BadgeClasses(variant)` and `ToggleClasses(variant, size)` do the same for badges
and toggles.

### Verification

The `<a>` renders with the full button styling but is still a link
(right-click → open in new tab works), and there's no invalid `<button>`-inside-a
control.

---

## Export a page to static HTML

Any `Component` or `*html.Document` renders to a string or file — useful for
static-site output, snapshots, or email.

```go
var buf strings.Builder
w := mx.NewCheckedWriter(&buf).WithIndent("", "  ")
if err := page.Render(context.Background(), w); err != nil {
    log.Fatal(err)
}
os.WriteFile("index.html", []byte(buf.String()), 0o644)
```

For a whole site, the `cmd/shadcn-gallery` command does exactly this with its
`-out` flag — see
[its source](https://github.com/ungerik/go-mx/tree/main/cmd/shadcn-gallery) for a
worked example, including a `-base` flag that prefixes links for hosting under a
URL sub-path (like GitHub Pages project pages).

### Verification

Open the written `index.html` directly in a browser (or serve the directory). The
markup matches what the server would send, because it's the same `Render` call.
