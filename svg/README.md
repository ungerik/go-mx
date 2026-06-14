# svg

Maps SVG elements and attributes to go-mx the same way the [`html`](../html)
package maps HTML: every element is a function returning a `*mx.Element`, and
every attribute is a function returning an `mx.Attrib`.

```go
import "github.com/ungerik/go-mx/svg"

doc := svg.Root(
    svg.ViewBox(0, 0, 100, 100),
    svg.Circle(svg.CX(50), svg.CY(50), svg.R(40), svg.Fill("tomato")),
)
// <svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 100 100'>...
```

## Relationship to the `html` package

The [`html`](../html) package has no `<svg>` constructor of its own — it does not
provide the SVG element/attribute vocabulary, `xmlns` namespace handling, or
numeric attribute values. Build all SVG content with this `svg` package: use
`svg.Root` for a standalone document (it prepends the `xmlns` namespace) or
`svg.SVG` for an inline `<svg>` embedded in an HTML page.

## Conventions

- **No void elements.** SVG has none — every element may contain children such
  as `<title>`, `<desc>`, `<metadata>` or animation elements, so all element
  constructors take `attribsChildren ...any`. Childless elements therefore
  render with an explicit close tag (`<rect ...></rect>`). Use `VoidElement`
  for the compact self-closing form when needed.
- **camelCase preserved.** SVG element and attribute names are case-sensitive;
  `LinearGradient` → `linearGradient`, `ViewBox` → `viewBox`,
  `FeGaussianBlur` → `feGaussianBlur`.
- **Numbers or strings.** Attribute constructors are generic over `Value`, so a
  number literal or a string both work: `svg.CX(50)`, `svg.StrokeWidth(1.5)`,
  `svg.Width("100%")`. (`Class` stays `...string` for multiple class names.)
  The exception is `ViewBox`, whose value is always the four numbers min-x,
  min-y, width and height, so it takes them as `float64` args:
  `svg.ViewBox(0, 0, 24, 24)`.
- **Element vs. attribute name collisions** are resolved by keeping the clean
  name for the element and suffixing the attribute with `Attr`:
  `Filter`/`FilterAttr`, `Mask`/`MaskAttr`, `ClipPath`/`ClipPathAttr`,
  `Path`/`PathAttr`.
- **`Root`** prepends the `xmlns` namespace for standalone documents; **`SVG`**
  is the plain `<svg>` element for inline embedding in HTML.
- **`StyleElem(css)`** renders a `<style>` element with raw (unescaped) CSS;
  **`Style(value)`** is the presentation attribute.

## Online Sources

### SVG Elements

- [MDN Web Docs - SVG element reference](https://developer.mozilla.org/en-US/docs/Web/SVG/Element)
- [SVG 2 Specification - Element index](https://www.w3.org/TR/SVG2/eltindex.html)

### SVG Attributes

- [MDN Web Docs - SVG attribute reference](https://developer.mozilla.org/en-US/docs/Web/SVG/Attribute)
- [SVG 2 Specification - Attribute index](https://www.w3.org/TR/SVG2/attindex.html)
- [MDN Web Docs - SVG presentation attributes](https://developer.mozilla.org/en-US/docs/Web/SVG/Attribute/Presentation)
