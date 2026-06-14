# html

Maps HTML5 elements and attributes to go-mx: every element is a function
returning a `*mx.Element` (`Div`, `Span`, `Img`, …) and every attribute is a
function or constant returning an `mx.Attrib` (`Class`, `ID`, `HRef`, …). Markup
is ordinary Go, so it composes, refactors, and type-checks with the rest of your
code — no templates and no separate template language.

```go
import "github.com/ungerik/go-mx/html"

html.Div(html.Class("card"),
    html.H2("Hello, go-mx"),
    html.P("Rendered in Go, escaped and validated."),
)
// <div class="card"><h2>Hello, go-mx</h2><p>Rendered in Go, escaped and validated.</p></div>
```

This README is the **reference** for the package. For a guided introduction and
task recipes see:

- **[Tutorial](../docs/html/tutorial.md)** — build and serve your first page.
- **[How-to guides](../docs/html/how-to.md)** — focused recipes for common tasks.
- **[Full API on pkg.go.dev](https://pkg.go.dev/github.com/ungerik/go-mx/html)**

## Contents

- [Relationship to the `mx` package](#relationship-to-the-mx-package)
- [Elements](#elements)
- [Convenience constructors](#convenience-constructors)
- [Children](#children)
- [Attributes](#attributes)
  - [Free values](#free-values)
  - [Boolean attributes](#boolean-attributes)
  - [Fixed-value constants](#fixed-value-constants)
  - [Strict keyword enums](#strict-keyword-enums)
  - [Typed and numeric helpers](#typed-and-numeric-helpers)
  - [`data-*` attributes](#data--attributes)
  - [Event handlers](#event-handlers)
  - [Name collisions: the `Attr` / `Elem` suffixes](#name-collisions-the-attr--elem-suffixes)
- [Text, raw HTML, and escaping](#text-raw-html-and-escaping)
- [Conditional rendering and iteration](#conditional-rendering-and-iteration)
- [Documents and serving](#documents-and-serving)
- [Forms](#forms)
- [Templates and table views](#templates-and-table-views)
- [The deferred-error pattern](#the-deferred-error-pattern)
- [Online sources](#online-sources)
- [Tips](#tips)

## Relationship to the `mx` package

`html` is a thin, HTML-flavored vocabulary over the root
[`mx`](https://pkg.go.dev/github.com/ungerik/go-mx) package. `mx` owns the core
abstractions; `html` names the HTML5 elements and attributes that build on them.

| `mx` abstraction | What it is                                |
|------------------|-------------------------------------------|
| `Component`      | The one interface: `Render(ctx, Writer) error` |
| `Element`        | An element: name, `[]Attrib`, children    |
| `Attrib`         | `AttribName()` + `AttribValue(ctx)` pair  |
| `Writer`         | Output sink with element lifecycle methods |
| `CheckedWriter`  | Escaping, validating, optionally indenting writer |

`html` re-exports the types you reach for most, so you rarely import `mx`
directly for everyday markup:

```go
type (
    Text     = mx.Text     // escaped text node
    Raw      = mx.Raw      // pre-trusted, unescaped HTML
    RawBytes = mx.RawBytes // pre-trusted, unescaped bytes
    Attribs  = mx.Attribs  // []Attrib helper
)
```

## Elements

Every HTML5 element is a constructor function. **Regular elements** take
attributes and children together as variadic `any`; **void elements** (those
that cannot have children, like `<img>` or `<br>`) take only `mx.Attrib`:

```go
html.Div(html.Class("container"), html.P("Hello"))  // <div class="container"><p>Hello</p></div>
html.Img(html.Src("/logo.png"), html.Alt("Logo"))   // void: attributes only, self-closing
```

Constructors return `*mx.Element`. Arguments are sorted at build time: anything
that is an `mx.Attrib` (or a slice/map/struct of them) becomes an attribute,
everything else becomes a child. The order you write children in is preserved.

Helpers and escape hatches:

| Function                              | Use                              |
|---------------------------------------|----------------------------------|
| `Element(name, attribsChildren...)`   | Any element by name, incl. custom/web-component tags |
| `VoidElement(name, attribs...)`       | Any void element by name         |
| `Hyperlink(href, text, attribs...)`   | Shortcut for `A(HRef(href), …, text)` |
| `Textf(format, args...)`              | `fmt`-formatted escaped text node |

`InputType*` constructors prepend the matching `type` attribute, so you don't
repeat it: `html.InputTypeEmail(html.Name("email"))` renders
`<input type="email" name="email"/>`. The full set covers every HTML input type
(`InputTypeText`, `InputTypeCheckbox`, `InputTypeDate`, `InputTypeFile`,
`InputTypeNumber`, `InputTypePassword`, `InputTypeSubmit`, …).

The same prepend-the-`type` pattern gives typed `<button>` and `<ol>`
constructors. `SubmitButton`, `ResetButton`, and `ButtonButton` set
`type="submit"`, `"reset"`, and `"button"` — handy because a bare `Button`
defaults to submitting its form, a common gotcha. `OLDecimal`, `OLLowerAlpha`,
`OLUpperAlpha`, `OLLowerRoman`, and `OLUpperRoman` set the ordered-list marker
via `type="1"`, `"a"`, `"A"`, `"i"`, and `"I"`.

Elements deprecated in HTML5 (`<center>`, `<font>`, `<marquee>`, …) are
intentionally left out; they remain as commented-out source so the omission is
explicit.

## Convenience constructors

The `*Button`, `OL*`, and `InputType*` families above bake a common attribute
into an element. `shortcuts.go` extends that idea to the element-and-attribute
pairings you write in almost every document — `<meta>`, `<script>`, `<link>`,
and a safe new-tab `<a>` — so one call replaces the boilerplate:

| Function                          | Renders                              |
|-----------------------------------|--------------------------------------|
| `MetaCharset(charset)`            | `<meta charset="…">`                 |
| `MetaCharsetUTF8`                 | `<meta charset="UTF-8">` as a `Raw` constant |
| `MetaName(name, content)`         | `<meta name="…" content="…">`        |
| `MetaProperty(property, content)` | `<meta property="…" content="…">` (Open Graph) |
| `MetaViewport(content)`           | `<meta name="viewport" content="…">` |
| `ScriptSrc(url, …)`               | `<script src="…">` external classic script |
| `ScriptModule(…)`                 | `<script type="module">` ES module   |
| `StyleSheet(href, …)`             | `<link rel="stylesheet" href="…">`   |
| `Icon(href, …)`                   | `<link rel="icon" href="…">` favicon |
| `LinkPreload(href, as, …)`        | `<link rel="preload" href="…" as="…">` |
| `BlankLink(href, text, …)`        | `<a … target="_blank" rel="noopener noreferrer">` |

`LinkPreload` takes the typed `As` destination enum (`AsScript`, `AsStyle`,
`AsFont`, …); it is named `LinkPreload` rather than `Preload` because `Preload`
is the media-element preload attribute enum. `BlankLink` sets
`rel="noopener noreferrer"`, which blocks reverse tabnabbing: without
`noopener` the opened page could navigate yours through `window.opener`.

## Children

Children are passed as `...any` and converted at render-build time by `mx`'s
`AsComponent`:

| You pass        | Renders as                            |
|-----------------|---------------------------------------|
| `nil`           | nothing                               |
| `string`        | escaped `Text`                        |
| `mx.Component`  | itself (passes through)               |
| a function      | wrapped as a `ComponentFunc`          |
| anything else   | `fmt.Sprint`-ed, then escaped as text |

Because non-`Component` values fall through to "stringified and escaped," a value
passed by mistake produces no compile error — it silently renders as text.
Convert non-obvious children to a `Component` explicitly.

## Attributes

### Free values

Plain functions whose value is an arbitrary string: `Class(...string)`,
`HRef(string)`, `Alt(string)`, `ID(string)`, `Name(string)`, `Src(string)`, and
so on. Many have a `*f` printf variant:

```go
html.Class("btn", "btn-primary")          // class="btn btn-primary"
html.IDf("row-%d", i)                       // id="row-0"
```

Variadic constructors join their arguments the way the attribute expects:
`Class`/`Rel`/`ItemProp` join with spaces, `Accept`/`Sizes`/`SrcSet` with
commas, `Coords` formats `float64`s into a comma list.

### Boolean attributes

HTML boolean attributes are present-or-absent: include them to turn them on,
omit them to leave the attribute off. They are `BoolAttrib` constants, and each
renders as `name="name"` (which HTML treats as equivalent to the bare attribute):

```go
html.Input(html.Type("checkbox"), html.Checked, html.Disabled)
// <input type="checkbox" checked="checked" disabled="disabled"/>
```

Available: `Checked`, `Disabled`, `Required`, `Readonly`, `Selected`, `Multiple`,
`Hidden`, `Async`, `Defer`, `AutoFocus`, `AutoPlay`, `Controls`, `Loop`, `Muted`,
`Open`, `Reversed`, `Inert`, `IsMap`, `ItemScope`, `NoValidate`,
`FormNoValidate`, `PlaysInline`, `NoModule`, `Default`, `Alpha`, and more.

`BoolAttrib` is exported so you can declare your own: `const HxBoosted =
html.BoolAttrib("hx-boost")`.

### Fixed-value constants

A few attributes have a small fixed set of values exposed as
`mx.ConstAttrib` constants (rendering `name=value`):

```go
html.A(html.HRef("/"), html.TargetBlank)   // target="_blank"
html.Form(html.AutoCompleteOff)             // autocomplete="off"
```

Includes `TargetSelf` / `TargetBlank` / `TargetParent` / `TargetTop` /
`TargetUnfencedTop`, `AutoCompleteOn` / `AutoCompleteOff`, and
`HiddenUntilFound`.

### Strict keyword enums

Attributes whose value is a closed set of keywords are typed enums in
[`enums.go`](enums.go) rather than `string` constructors or loose `name=value`
constants. Each is a `string` type implementing `mx.Attrib`, so the typed
constants are used directly as attributes:

```go
html.Div(html.DirRTL, html.SpellCheckFalse)
html.Img(html.Src("/a.png"), html.LoadingLazy, html.DecodingAsync)
html.Form(html.MethodPOST, html.EncTypeMultipartFormData)
```

A conversion such as `html.Dir("rtl")` also works for dynamic values, and the
`Valid`/`Validate`/`Enums`/`EnumStrings`/`String` methods (generated by
[`ungerik/go-enum`](https://github.com/ungerik/go-enum)) report whether a value
is one of the defined keywords. An out-of-set value (such as `html.Dir("bogus")`)
is not silently emitted: the enum's `AttribValue` returns its `Validate` error,
so rendering the enclosing element fails with that error — the same deferred-error
pattern used by the [`svg`](../svg) package (see
[below](#the-deferred-error-pattern)). The generator is pinned in the nested
[`tools`](../tools) module (kept out of the shipped go-mx dependency tree) and
run with `go generate ./html/...`.

Covered attributes: `autocapitalize`, `autocorrect`, `contenteditable`,
`crossorigin`, `decoding`, `dir`, `enctype`, `enterkeyhint`, `fetchpriority`,
`formenctype`, `formmethod`, `http-equiv`, `inputmode`, `kind`, `loading`,
`method`, `referrerpolicy`, `shape`, `as`, `capture`, `preload`, `scope`,
`spellcheck`, `translate` and `wrap`. Attributes whose value can be a keyword
*or* a free value (`target`, `autocomplete`) stay loose, as do space-separated
token lists (`rel`, `class`).

### Typed and numeric helpers

Where a value has a natural Go type, the constructor takes it and formats the
string for you:

| Constructor                  | Signature                | Renders            |
|------------------------------|--------------------------|--------------------|
| `Cols` / `Rows`              | `(int)`                  | `cols="40"`        |
| `ColSpan` / `RowSpan`        | `(int)`                  | `colspan="2"`      |
| `MaxLength` / `MinLength`    | `(int)`                  | `maxlength="255"`  |
| `High` / `Low` / `Optimum`   | `(float64)`              | `high="0.8"`       |
| `Coords`                     | `(...float64)`           | `coords="0,0,10"`  |
| `Draggable`                  | `(bool)`                 | `draggable="true"` |
| `WidthPx` / `HeightPx`       | `(float64)`              | `width="64px"`     |
| `WidthEm` / `HeightEm`       | `(float64)`              | `width="4em"`      |

`Width`/`Height` themselves take a `string` (so `"100%"` works); the `*Px`/`*Em`
variants are convenience wrappers.

### `data-*` attributes

```go
html.Div(html.DataAttr("user-id", "42"))         // data-user-id="42"
html.Div(html.DataAttrf("count", "%d", n))        // data-count="3"
```

### Event handlers

Every standard HTML event handler attribute has an `On*` constructor taking the
script string to run: `OnClick`, `OnChange`, `OnInput`, `OnSubmit`, `OnLoad`,
`OnKeyDown`, `OnMouseOver`, … (and the window-level ones like `OnBeforeUnload`,
`OnPopState`). For interactivity without inline JavaScript, see the
[`hx`](../hx) package.

```go
html.Button(html.OnClick("save()"), "Save")
```

### Name collisions: the `Attr` / `Elem` suffixes

Some HTML element names and attribute names are identical (`<cite>` the element
vs the `cite` attribute). Where both exist, the element keeps the plain name and
the attribute gets an `Attr` suffix — or vice-versa where the attribute is more
common. The full set:

| HTML name  | Element constructor | Attribute constructor |
|------------|---------------------|-----------------------|
| `cite`     | `Cite`              | `CiteAttr`            |
| `content`  | `Content`           | `ContentAttr`         |
| `data`     | `Data`              | `DataAttr` (`data-*`) |
| `form`     | `Form`              | `FormAttr`            |
| `label`    | `Label`             | `LabelAttr`           |
| `slot`     | `Slot`              | `SlotAttr`            |
| `span`     | `Span`              | `SpanAttr`            |
| `style`    | `StyleElem`         | `Style`               |
| `title`    | `TitleElem`         | `Title`               |
| `template` | `TemplateElem`      | —                     |
| `textarea` | `TextArea`          | —                     |

(`Lang` and `Language` are both attributes — they map to the distinct `lang` and
`language` HTML attributes.)

## Text, raw HTML, and escaping

| Value / function       | Behavior                                  |
|------------------------|-------------------------------------------|
| `Text("a < b")`        | escaped text node → `a &lt; b`            |
| `Textf("%d", n)`       | `fmt`-formatted, then escaped             |
| `Raw("<b>x</b>")`      | emitted verbatim — you vouch for it       |
| `RawBytes([]byte…)`    | same as `Raw` for a byte slice            |
| `Escape(s)`            | returns the escaped string                |
| `WriteEscaped(w, s)`   | writes escaped to an `io.Writer`          |
| `WriteRaw(w, s)`       | writes verbatim to an `io.Writer`         |

`Raw` is the trusted-HTML escape hatch: never pass user input to it. A bare
`string` child is always escaped, so the safe path is the default.

Named HTML character references live in the [`html/entity`](entity) subpackage
as `mx.Raw` constants, kept out of `html` so their short names don't collide
with element and attribute constructors. It covers markup characters, spaces,
currency signs, legal marks (`Copyright` ©, `Registered` ®, `Trademark` ™),
punctuation, quotation marks, math and comparison signs, arrows, and card suits:

```go
html.P("Total: ", entity.Euro, "9.99 ", entity.Copyright, " 2026")
```

## Conditional rendering and iteration

Re-exported from `mx` so you can keep one import:

```go
html.If(loggedIn, html.Span("Welcome")).Else(html.A(html.HRef("/login"), "Log in"))

html.UL(
    html.ForEach(items, func(s string) *mx.Element {
        return html.LI(s)
    }),
)
```

- `If(cond, comps...)` returns an `mx.IfElse` with `.Else(...)`, `.ElseIf(...)`,
  and `.ElseIff(...)`.
- `Iff(condFunc, comps...)` is like `If` but takes the condition as a
  `func() bool`; the func is called immediately when the `IfElse` is built.
- `ForEach(slice, fn)` maps a slice to components; `ForEachIter(seq, fn)` does the
  same for an `iter.Seq`.

## Documents and serving

`Document` assembles a complete `<!DOCTYPE html>` page and is itself a
`Component`:

```go
type Document struct {
    Title        string
    Meta         map[string]string // name  -> content
    MetaProperty map[string]string // property -> content (Open Graph, …)
    Stylesheets  []string          // <link rel="stylesheet"> hrefs
    Style        string            // inline <style> after the stylesheets
    HeadCustom   mx.Component       // extra <head> content, last
    Body         mx.Component
}
```

`NewDocument(title, body...)` is the quick constructor; it converts the body
args with `mx.AsComponents`. It writes a `<head>` with a UTF-8 `<meta>`, the
title, sorted meta tags, your stylesheets, and any inline style, then your body.

Serving:

| Method / function                 | What it does                          |
|-----------------------------------|---------------------------------------|
| `(*Document).HandleHTTP`          | An `http.HandlerFunc`: renders the page with an indenting `CheckedWriter` |
| `(*Document).Serve(addr)`         | `ListenAndServe` serving just this document |
| `Serve(addr, component)`          | `ListenAndServe` for any `mx.Component` |
| `DOCTYPE`                         | The `<!DOCTYPE html>` `Raw` constant  |

Both serving paths set `Content-Type: text/html; charset=utf-8` and turn a
render error into a 500 via `mx.RespondNonContextError` (which never leaks the
error string to the client).

```go
page := html.NewDocument("Hello",
    html.H1("Hello, go-mx"),
    html.P("Type-safe HTML, rendered in Go — no templates."),
)
http.HandleFunc("/", page.HandleHTTP)
log.Fatal(http.ListenAndServe(":8080", nil))
```

## Forms

Two ways to turn a Go struct into a form, one current and one deprecated.

### `FieldDecider` — the current path

`html.FieldDecider` is the plain-HTML implementation of `mx.FieldDecider`. Paired
with `mx.ReflectFormHandler`, it renders, parses, and validates a form for a
struct type, picking an `<input>` / `<select>` / `<textarea>` per field from its
detected kind and `form:"…"` tag:

```go
type Signup struct {
    Email string `form:"required,placeholder=you@example.com"`
    Age   int    `form:"min=13"`
    Bio   string `form:"widget=textarea,help=Tell us about yourself"`
}

handler := mx.ReflectFormHandler(
    nil, // load: nil = submit-only form, seeded with new(Signup)
    func(ctx context.Context, s *Signup) error {
        // persist s; return an mx.FieldErrors to show per-field errors
        return nil
    },
    html.FieldDecider, // or install app-wide with mx.Middleware(html.FieldDecider)
)
http.Handle("/signup", handler)
```

The handler renders the `<form>` on GET, and on POST parses **only the fields it
rendered** (an allowlist that defends against mass-assignment), runs the
validation chain, and either re-renders with inline errors or 303-redirects on
success. `FieldDecider` handles the full `mx.FieldKind` dispatch — strings,
numbers, bools, dates, files, enums, enum sets — plus the `__clear` sentinel for
nullable fields. The layered renderers in [`hx`](../hx) and
[`shadcn`](../shadcn) wrap this decider, customizing the kinds they care about
and delegating the rest.

Tag keys (comma-separated inside `form:"…"`): `required`, `readonly`,
`sensitive` (never echo the value back), `hidden`, `-` (skip),
`placeholder=…`, `help=…`, `pattern=…`, `min=…`, `max=…`, `step=…`,
`label=…`, `widget=…` (`textarea`, `email`, `url`, `tel`, `password`, `date`,
`time`, `radio`, `file`, …).

### `ReflectFormComponents` — deprecated

`ReflectFormComponents` renders a struct as a flat list of inputs from `input:"…"`
struct tags. It predates the decider chain and lacks parsing, validation, and
sentinels.

```go
html.Form(html.Action("/submit"), html.MethodPOST,
    html.ReflectFormComponents(UserDetails{Name: "John Doe"}),
    html.InputTypeSubmit(html.Value("Submit")),
)
```

> **Deprecated:** use `mx.ReflectFormHandler` with `FieldDecider` instead.

## Templates and table views

- `Template{File, Data, Funcs}` is a `Component` that parses a `text/template`
  glob and executes it into the writer — an interop bridge when you already have
  template files.
- `TableView` is an interface (`Title`, `Columns`, `NumRows`, `Cell`) for types
  that expose tabular data in a uniform way.

## The deferred-error pattern

Element and attribute constructors return a single value and never a separate
error, so the nested, variadic form composes cleanly. A constructor that detects
an invalid input does not panic or drop it — it returns a value that reports the
error at render time:

- An invalid keyword enum (`html.Dir("bogus")`) renders the enclosing element
  with that error: `(*mx.Element).String()` returns
  `mx.Element.String: invalid value "bogus" for type html.Dir`, and a real
  render returns the error from `Render`.
- `mx.NewErrElement(err)` builds an element whose `Render` returns `err`;
  `mx.ErrAttrib` does the same for an attribute.

A build-time problem is therefore never lost — it surfaces the first time the
affected subtree is rendered. See the
[`mx` package docs](https://pkg.go.dev/github.com/ungerik/go-mx#hdr-Deferred_errors)
for the full rationale.

## Online sources

### HTML Elements

- [MDN Web Docs - HTML elements reference](https://developer.mozilla.org/en-US/docs/Web/HTML/Element)
- [WHATWG HTML Living Standard - Elements](https://html.spec.whatwg.org/multipage/indices.html#elements-3)
- [W3Schools HTML Element Reference](https://www.w3schools.com/tags/)

### HTML Attributes

- [MDN Web Docs - HTML attribute reference](https://developer.mozilla.org/en-US/docs/Web/HTML/Attributes)
- [WHATWG HTML Living Standard - Attributes](https://html.spec.whatwg.org/multipage/indices.html#attributes-3)
- [W3Schools HTML Attribute Reference](https://www.w3schools.com/tags/ref_attributes.asp)

### Void Elements (Self-Closing)

- [MDN Web Docs - Void elements](https://developer.mozilla.org/en-US/docs/Glossary/Void_element)
- [WHATWG HTML Living Standard - Void elements](https://html.spec.whatwg.org/multipage/syntax.html#void-elements)

### Boolean Attributes

- [MDN Web Docs - Boolean attributes](https://developer.mozilla.org/en-US/docs/Glossary/Boolean/HTML)
- [WHATWG HTML Living Standard - Boolean attributes](https://html.spec.whatwg.org/multipage/common-microsyntaxes.html#boolean-attributes)

### Additional Resources

- [All HTML elements and attributes](https://github.com/nickytonline/all-html-elements-and-attributes) - Comprehensive list
- [Can I use](https://caniuse.com/) - Browser compatibility tables

## Tips

### Use HTML without JavaScript

https://www.htmhell.dev/adventcalendar/2025/27/
