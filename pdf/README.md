# pdf

[![Go Reference](https://pkg.go.dev/badge/github.com/ungerik/go-mx/pdf.svg)](https://pkg.go.dev/github.com/ungerik/go-mx/pdf)

Composable PDF rendering primitives structured like the go-mx
[`html`](../html) package, but rendering to PDF through a **natively included
engine** instead of an HTML markup writer — with **no external PDF
dependency**. The PDF engine is inlined from
[`codeberg.org/go-pdf/fpdf`](https://codeberg.org/go-pdf/fpdf) v0.12.0 (MIT, see
[`THIRD-PARTY-LICENSES.md`](../THIRD-PARTY-LICENSES.md)) and adapted to the
component model, so `pdf` is a normal package of the root module — no separate
`go.mod`.

```go
import "github.com/ungerik/go-mx/pdf"

doc := pdf.NewDocument("Invoice",
    pdf.Font(pdf.Helvetica, pdf.StyleBold, 24),
    pdf.Paragraph("Hello, PDF!"),
    pdf.MoveDown(4),
    pdf.Save( // scoped style — restored after these children
        pdf.TextColor(pdf.Gray50),
        pdf.Font(pdf.Helvetica, pdf.StyleRegular, 10),
        pdf.Paragraph("Generated with go-mx/pdf."),
    ),
)
err := doc.OutputFile(ctx, "hello.pdf")
```

> The [`fpdf`](../fpdf) module is the previous wrapper around the *external*
> fpdf dependency. It is kept only as the byte-for-byte parity baseline while
> this native package is brought up, and will be deleted afterwards. New code
> should import `github.com/ungerik/go-mx/pdf`.

## Relationship to the `html` package

The mapping to `html` is deliberate, so the two packages feel the same:

| `html` / `mx`                       | `pdf`                                  |
| ----------------------------------- | -------------------------------------- |
| `Component` (`Render(ctx, Writer)`) | `Component` (`Render(ctx, *Renderer)`) |
| markup `Writer`                     | `Renderer` (the inlined PDF engine)    |
| `Components`, `If`, `ForEach`       | `Components`, `If`, `ForEach`          |
| `html.Document`                     | `pdf.Document`                         |
| elements (`Div`, `P`, …)            | primitives (`Paragraph`, `Rect`, …)    |
| attributes (`Class`, `Style`, …)    | state components (`Font`, `Color`, …)  |

The crucial difference is that **PDF has no element tree**. HTML is a retained
tree of nested elements with attributes; the PDF engine is an *imperative,
stateful* drawing API where the current page, cursor, font and colors persist
from call to call. So the pdf primitives are thin wrappers around that state
machine, not a tree, and "attributes" become **state components** (`Font`,
`TextColor`, …) that mutate the renderer until changed. Use
[`Save`](#scoping-state-with-save) to scope a state change to a group of
children.

## The `Renderer`

`Renderer` **is** the inlined PDF engine, so **every engine method is available
directly** and any primitive can be expressed in raw engine calls when the
typed helpers do not cover it:

```go
r := pdf.NewRendererA4Portrait()
r.AddPage()
r.SetDrawColor(255, 0, 0) // raw engine call
pdf.Line(10, 10, 100, 10).Render(ctx, r) // typed primitive
```

Component-model constructors build a renderer from typed values:
`NewRenderer(orientation, unit, size)`, `NewRendererA4Portrait`,
`NewRendererA4Landscape`, `NewRendererLetterPortrait`. The low-level engine
constructors `New(orientationStr, unitStr, sizeStr, fontDirStr)` and
`NewCustom(*InitType)` are also exported for full control. The renderer also
adds in-memory asset helpers (`LoadUTF8FontBytes`, `LoadUTF8FontReader`) — see
[In-memory assets](#in-memory-assets).

## Primitives

- **Text** — `Text`, `Textf`, `Cell`, `CellFormat`, `MultiCell`, `Paragraph`,
  `TextAt`, `Ln`, `NewLine`. A bare `string` child becomes flowing `Text`.
- **Vector** — `Line`, `Rect`, `RoundedRect`, `Circle`, `Ellipse`, `Polygon`.
- **Images** — `Image` (file path), `ImageReader` / `ImageBytes` (in memory).
- **State** — `Font`, `FontSize`, `TextColor`, `FillColor`, `DrawColor`,
  `LineWidth`, `LineCap`, `LineJoin`, `X`, `Y`, `XY`, `MoveDown`, `MoveRight`.
- **Layout** — `Page`, `PageFormat`, `Document`.

The first drawing primitive (or the first `Page`) automatically opens page one,
so a one-page document does not need an explicit `Page`.

### Shortcuts and constants

Closed-set values are **typed enums** with generated `Valid` / `Validate` /
`Enums` / `EnumStrings` methods (via `go generate ./...`, the go-enum tool
pinned in the `tools` module), so the compiler guides you and values can be
validated:

- `Orientation` (`Portrait`, `Landscape`)
- `Unit` (`UnitPoint`, `UnitMillimeter`, `UnitCentimeter`, `UnitInch`)
- `PageSize` (`A1`–`A8`, `Letter`, `Legal`, `Tabloid`)
- `FontStyle` (`StyleRegular`, `StyleBold`, `StyleItalic`, `StyleUnderline`,
  `StyleStrikeOut`, and every combination, e.g. `StyleBoldItalic`) — the named
  combos double as concatenations: `StyleBold + StyleItalic == StyleBoldItalic`
- `Border` (`BorderNone`, `BorderFull`, the four edges, and every combination,
  e.g. `BorderLeftTop`) — the named combos double as concatenations:
  `BorderLeft + BorderTop == BorderLeftTop`
- `HAlign` (`AlignLeft`, `AlignCenter`, `AlignRight`, `AlignJustify`) and
  `VAlign` (`AlignTop`, `AlignMiddle`, `AlignBottom`, `AlignBaseline`) — two
  single-choice axes for cell text alignment
- `DrawOp` (`Stroke`, `FillShape`, `FillStroke`)
- `LnPos` (`LnRight`, `LnNewline`, `LnBelow`)
- `LineCapStyle` (`CapButt`, `CapRound`, `CapSquare`)
- `LineJoinStyle` (`JoinMiter`, `JoinRound`, `JoinBevel`)
- `ImageType` (`ImagePNG`, `ImageJPEG`, `ImageGIF`) for in-memory images

The only genuinely open value is the **font family** name, a plain `string`
because any registered font is valid: `Helvetica`, `Arial`, `Times`, `Courier`,
`Symbol`, `ZapfDingbats`, or any family added with `LoadUTF8Font…`.

Plus the RGB `Color` type with named colors (`Black`, `White`, `Red`, …), `RGB`,
`Gray`, and a CSS-style `Hex` / `MustHex` parser, and `Point` (alias of
`PointType`) with the `Pt(x, y)` helper.

The shortest path to a one-page document is `Paragraph`, which is a full-width,
auto-line-height, left-aligned `MultiCell` — the PDF analog of `<p>`.

### Scoping state with `Save`

```go
pdf.Save(
    pdf.TextColor(pdf.Red),
    pdf.Font(pdf.Times, pdf.StyleItalic, 14),
    pdf.Paragraph("only this text is red and italic"),
)
```

`Save` captures and restores the font (family, style and size), the
text/fill/draw colors, the line width, the line cap/join styles and the cursor
position, using the engine's getters — so it works whether the state was set
through this package or the raw renderer. The **dash pattern** and the
**alpha/blend mode** are not restored (the engine has no dash-pattern getter,
and its zero alpha value is indistinguishable from a deliberate fully-transparent
setting); reset those explicitly if you change them inside the scope. The cursor
restore assumes children stay on the same page — if they trigger an automatic
page break, the restored position lands on the new page, so `Save` is for
scoping style, not page flow.

## In-memory assets

No asset needs to live on disk. Images draw from memory with `ImageReader` /
`ImageBytes`:

```go
pdf.ImageBytes("logo", pdf.ImagePNG, pngBytes, 20, 40, 30, 30)
```

The `name` argument is the engine's image cache key — reuse it to draw the same
image without re-decoding, and give distinct images distinct names.

Fonts load from memory through the renderer. `LoadUTF8FontBytes` /
`LoadUTF8FontReader` register an embedded TrueType font and switch the
translator to identity so non-Latin text works without any `.ttf` file:

```go
r := doc.NewRenderer()
r.LoadUTF8FontBytes("DejaVu", pdf.StyleRegular, ttfBytes)
// then select it: pdf.Font("DejaVu", pdf.StyleRegular, 12)
```

The remaining assets are already in-memory-capable on the `Renderer`:
metrics-format fonts via `AddFontFromBytes` / `AddFontFromReader`, file
attachments via `SetAttachments` (the `Attachment` carries its bytes), and XMP
metadata via `SetXmpMetadata`.

## `Document`

`Document` carries metadata, page setup, a default font, optional per-page
`Header`/`Footer`, and the body. It renders to a `Renderer`, an `io.Writer`
(`Output`), a `[]byte` (`Bytes`), a file (`OutputFile`), or an
`http.ResponseWriter` (`ServeHTTP`, served as `application/pdf` with a generic
500 on error). All page-setup fields default to A4 / portrait / millimeters
with a Helvetica 12pt font.

## Raw-engine surface

Because the engine is inlined, its lower-level API is exported alongside the
component layer for callers that need it: the constructors `New` / `NewCustom`
with `InitType` and `SizeType`, image placement via `ImageOptions`, file
attachments via `Attachment`, PDF/A output intents via `OutputIntentType`, page
boxes via `PageBox` / `SetPageBox`, and standalone helpers `ComparePDFs` (test
diffing), `TtfParse` / `UTF8CutFont` (font tooling), `Tickmarks` (axis labels),
and the package-level `SetDefaultCompression`. These mirror the upstream fpdf
API; reach for them only when a typed component does not cover what you need.

## Coordinate system and units

The engine puts the origin at the **top-left** and Y grows downward (the PDF
imaging model itself uses a bottom-left origin). All coordinates, widths and
heights are in the document `Unit`; **font sizes are always in points**
regardless of the document unit.

## Errors

The engine accumulates the first error internally and silently turns subsequent
calls into no-ops; you must check `Renderer.Error()` (or `Ok` / `Err`). Every
primitive in this package does that for you and returns it from `Render`, and
components honor `context` cancellation.

## Concurrency

A `Renderer` is stateful and **not safe for concurrent use**. Build each
document from a single goroutine.

---

## Limitations of the engine vs. the PDF specification

The engine is a pragmatic generator, not a full PDF implementation. It emits
**PDF 1.3** (bumped to 1.4 for alpha/blend modes and 1.5 for layers). The
following parts of the PDF spec are unsupported or only partially supported.
Where a primitive in this package can't help, drop down to the raw `Renderer`
or post-process the output with a more complete library.

### Text and fonts

- **Standard ("core") fonts are cp1252 only.** Helvetica/Arial, Times, Courier,
  Symbol and ZapfDingbats use the Windows Western-Europe encoding. Characters
  outside cp1252 (most non-Latin scripts, many typographic symbols) require
  embedding a TrueType/UTF-8 font — from a file via `Renderer.AddUTF8Font`, or
  from memory via `Renderer.LoadUTF8FontBytes` / `LoadUTF8FontReader`, which
  also switch the translator so UTF-8 strings pass through unchanged.
- **No complex text shaping.** No ligatures, contextual forms, Indic
  reordering, mark positioning or kerning beyond raw font metrics. `RTL()`
  merely reverses direction; there is no Unicode bidi algorithm and no
  Arabic/Hebrew shaping. No vertical writing modes.
- **No hyphenation.** `MultiCell` wraps on spaces only; a single word wider than
  the box is broken crudely by character.
- **Limited font formats.** TrueType and Type1; non-UTF8 core-font use needs
  metrics generated by the `makefont` tool. No reliable OpenType-CFF (`.otf`
  with PostScript outlines), no variable fonts, no automatic system-font
  discovery.

### Color and graphics

- **Color spaces:** DeviceRGB and DeviceGray only, plus Separation **spot
  colors** (`AddSpotColor`). No ICC-based color, no CIE Lab, and **CMYK only via
  spot colors** — there is no direct DeviceCMYK fill/stroke.
- **Shadings/patterns:** only **axial (linear)** and **radial** gradients. No
  tiling patterns, no function-based, free-form/lattice (Gouraud) or Coons/
  tensor mesh shadings (PDF shading types 4–7).
- **Transparency:** constant alpha and blend modes via `SetAlpha`, and an image
  soft mask, but no general soft masks, transparency groups, or
  isolated/knockout groups.

### Images

- **JPEG, PNG and GIF only** (PNG alpha supported via soft mask). No TIFF, BMP,
  WebP, JPEG 2000, or inline images. CMYK JPEG support is limited.

### Interactivity and structure

- **No interactive form fields (AcroForm).** No text fields, checkboxes, radio
  buttons, choice lists or push buttons. (Link annotations, file-attachment
  annotations, document-level JavaScript and a basic outline/bookmark tree are
  available.)
- **Annotations are limited** to links and file attachments — no highlight,
  text/sticky-note, stamp, ink, redaction or widget annotations.
- **No Tagged PDF / accessibility (PDF/UA).** No structure tree, logical reading
  order, alt text, or `/Tagged` marking, so output is not accessible.
- **No full PDF/A or PDF/X conformance.** `AddOutputIntent` supplies an ICC
  output intent and `SetXmpMetadata` supplies XMP — the building blocks — but
  nothing enforces or validates archival/print conformance, and font embedding,
  color and metadata still have to be made conformant by hand.
- **Optional content (layers)** is supported only as simple show/hide groups,
  not nested membership dictionaries or complex configurations.

### Security

- **Encryption is RC4 (40/128-bit) only** via `SetProtection` — the legacy
  standard security handler. **No AES-128 or AES-256**, so protection is weak by
  modern standards.

### Other

- **No SVG import.** The upstream `SVGBasicWrite` / `SVGBasicDraw` helpers were
  dropped when the engine was inlined, so this package has no SVG support at
  all; rasterize or convert SVG before placing it.
- **No page transitions, multimedia, 3D, embedded-file UI, or digital
  signatures.**
