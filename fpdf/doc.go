// Package fpdf provides composable PDF rendering primitives structured like
// the go-mx html package, but targeting the external [codeberg.org/go-pdf/fpdf]
// renderer instead of an HTML markup writer.
//
// This is the legacy wrapper around the external fpdf module. It is kept only
// as the byte-for-byte parity baseline while the native
// github.com/ungerik/go-mx/pdf package (which inlines the fpdf engine) is
// brought up, and will be deleted afterwards. New code should import
// github.com/ungerik/go-mx/pdf.
//
// The correspondence to the html package is deliberate:
//
//   - [Component] is the html Component: anything that can draw implements
//     Render(context.Context, *Renderer) error.
//   - [Renderer] replaces the markup writer. It embeds [*fpdf.Fpdf], so every
//     fpdf method is available directly and any primitive can be expressed in
//     raw fpdf when needed.
//   - [Components], [If], [Iff], [ForEach] and [ForEachIter] mirror their mx
//     counterparts for composing and conditionally rendering content.
//   - [Document] is html.Document: it carries metadata, page setup and a body,
//     and renders to a [Renderer], an io.Writer, a file or an http.ResponseWriter.
//
// Unlike HTML, PDF has no element tree. fpdf is an imperative, stateful drawing
// API: the current page, cursor, font and colors persist across calls. The
// primitives here are therefore thin wrappers around that state machine rather
// than a retained tree. They divide into:
//
//   - text:    [Text], [Textf], [Cell], [CellFormat], [MultiCell],
//     [Paragraph], [TextAt], [Ln], [NewLine]
//   - vector:  [Line], [Rect], [RoundedRect], [Circle], [Ellipse], [Polygon]
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
// See README.md for the limitations of the fpdf renderer relative to the PDF
// specification.
package fpdf
