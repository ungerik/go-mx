// Package pdf provides composable PDF rendering primitives structured like
// the go-mx html package, rendering to PDF through a natively included
// engine instead of an HTML markup writer — with no external PDF dependency.
//
// The correspondence to the html package is deliberate:
//
//   - [Component] is the html Component: anything that can draw implements
//     Render(context.Context, *Renderer) error.
//   - [Renderer] replaces the markup writer. It is the PDF engine itself, so
//     every rendering method (Cell, Rect, Image, Transform, …) is available
//     directly and any primitive can be expressed in raw engine calls when
//     the typed components do not cover it.
//   - [Components], [If], [Iff], [ForEach] and [ForEachIter] mirror their mx
//     counterparts for composing and conditionally rendering content.
//   - [Document] is html.Document: it carries metadata, page setup and a body,
//     and renders to a [Renderer], an io.Writer, a file or an http.ResponseWriter.
//
// Unlike HTML, PDF has no element tree. The renderer is an imperative,
// stateful drawing engine: the current page, cursor, font and colors persist
// across calls. The components here are therefore thin wrappers around that
// state machine rather than a retained tree. They divide into:
//
//   - text:    [Text], [Textf], [Cell], [CellFormat], [MultiCell],
//     [Paragraph], [TextAt], [Ln], [NewLine]
//   - vector:  [Line], [Rect], [RoundedRect], [Circle], [Ellipse], [Polygon]
//   - svg:     [SVG] draws an svg-package element tree as vector graphics
//   - images:  [Image] (file), [ImageReader] / [ImageBytes] (in memory)
//   - state:   [Font], [FontSize], [TextColor], [FillColor], [DrawColor],
//     [LineWidth], [LineCap], [LineJoin], [X], [Y], [XY], [MoveDown],
//     [MoveRight], and [Save] to scope state changes to a group of children
//   - layout:  [Page], [PageFormat], [Document]
//
// Closed-set values are typed enums with generated Valid, Validate, Enums and
// EnumStrings methods ([Orientation], [Unit], [PageSize], [FontStyle], [HAlign],
// [VAlign], [Border], [DrawOp], [LnPos], [LineCapStyle], [LineJoinStyle],
// [ImageType]); [FontStyle] enumerates all sixteen bold/italic/underline/
// strike-out combinations and [Border] all sixteen left/top/right/bottom edge
// combinations plus the "1" full-border shorthand, while horizontal and vertical
// cell alignment are two independent single-choice axes ([HAlign], [VAlign]).
// The only genuinely open value is the font family name ([Helvetica], …, or any
// registered font), which stays a plain string constant. There is also an RGB
// [Color] type with named colors and a [Hex] parser.
//
// A minimal document:
//
//	doc := pdf.NewDocument("Hello",
//		pdf.Font(pdf.Helvetica, pdf.StyleBold, 24),
//		pdf.Paragraph("Hello, PDF!"),
//	)
//	err := doc.OutputFile(ctx, "hello.pdf")
//
// The PDF-generation engine in this package is inlined from
// [codeberg.org/go-pdf/fpdf] v0.12.0 (MIT license, see THIRD-PARTY-LICENSES.md
// in the repository root): the [Renderer] type is fpdf's Fpdf adapted to the
// go-mx component model, with the contrib bridges (gofpdi templates), the
// makefont tooling, and the HTML/SVG/grid extras removed. See README.md for
// the limitations of the engine relative to the PDF specification.
package pdf
