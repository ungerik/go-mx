package fpdf_test

// The parity suite renders the same document through the legacy fpdf wrapper
// (this module, wrapping codeberg.org/go-pdf/fpdf) and through the native
// github.com/ungerik/go-mx/pdf package (which inlines that engine), and
// requires the outputs to be byte-identical. Every scenario is written twice
// — once per package — because the two APIs are deliberately identical in
// shape but distinct in types.
//
// This suite is deleted together with the fpdf module once the native
// package is adopted.

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"testing"
	"time"

	legacy "github.com/ungerik/go-mx/fpdf"
	native "github.com/ungerik/go-mx/pdf"
)

var fixedDate = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

// renderer is the determinism-setup surface shared by the legacy and native
// Renderer types, so both pipelines are guaranteed to render with identical
// settings.
type renderer interface {
	SetCreationDate(time.Time)
	SetModificationDate(time.Time)
	SetCompression(bool)
	SetCatalogSort(bool)
	Output(io.Writer) error
}

// renderDeterministic pins the determinism knobs on r, runs render, and
// returns the produced PDF. side names the pipeline in failure messages.
func renderDeterministic(t *testing.T, side string, r renderer, compress bool, render func() error) []byte {
	t.Helper()
	r.SetCreationDate(fixedDate)
	r.SetModificationDate(fixedDate)
	r.SetCompression(compress)
	r.SetCatalogSort(true)
	if err := render(); err != nil {
		t.Fatalf("%s render: %v", side, err)
	}
	var buf bytes.Buffer
	if err := r.Output(&buf); err != nil {
		t.Fatalf("%s output: %v", side, err)
	}
	return buf.Bytes()
}

func renderLegacy(t *testing.T, doc *legacy.Document, compress bool) []byte {
	t.Helper()
	r := doc.NewRenderer()
	return renderDeterministic(t, "legacy", r, compress, func() error {
		return doc.Render(context.Background(), r)
	})
}

func renderNative(t *testing.T, doc *native.Document, compress bool) []byte {
	t.Helper()
	r := doc.NewRenderer()
	return renderDeterministic(t, "native", r, compress, func() error {
		return doc.Render(context.Background(), r)
	})
}

// assertSamePDF fails the test when the two renders differ, dumping a
// normalized byte diff for diagnosis.
func assertSamePDF(t *testing.T, compress bool, legacyPDF, nativePDF []byte) {
	t.Helper()
	if !bytes.Equal(legacyPDF, nativePDF) {
		t.Errorf("compress=%t: legacy (%d bytes) and native (%d bytes) output differ",
			compress, len(legacyPDF), len(nativePDF))
		if err := native.CompareBytes(legacyPDF, nativePDF, true); err != nil {
			t.Log(err)
		}
	}
}

func assertParity(t *testing.T, doc1 *legacy.Document, doc2 *native.Document) {
	t.Helper()
	for _, compress := range []bool{false, true} {
		assertSamePDF(t, compress, renderLegacy(t, doc1, compress), renderNative(t, doc2, compress))
	}
}

// Deterministic in-memory test images shared by both document builders.

func pngImageBytes(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for y := range 8 {
		for x := range 8 {
			img.Set(x, y, color.RGBA{R: uint8(x * 32), G: uint8(y * 32), B: 128, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func jpegImageBytes(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 9, 8))
	for y := range 8 {
		for x := range 9 {
			img.Set(x, y, color.RGBA{R: 200, G: uint8(x * 25), B: uint8(y * 30), A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func gifImageBytes(t *testing.T) []byte {
	t.Helper()
	img := image.NewPaletted(image.Rect(0, 0, 10, 8), color.Palette{
		color.RGBA{R: 255, A: 255}, color.RGBA{G: 255, A: 255},
		color.RGBA{B: 255, A: 255}, color.RGBA{R: 255, G: 255, B: 255, A: 255},
	})
	for i := range img.Pix {
		img.Pix[i] = uint8(i % 4)
	}
	var buf bytes.Buffer
	if err := gif.Encode(&buf, img, nil); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestParityGolden(t *testing.T) {
	assertParity(t,
		&legacy.Document{
			Title:   "go-mx pdf golden",
			Author:  "go-mx",
			Subject: "golden reference",
			Body: legacy.Components{
				legacy.Font(legacy.Helvetica, legacy.StyleBold, 20),
				legacy.Paragraph("go-mx/pdf golden reference"),
				legacy.MoveDown(4),
				legacy.Font(legacy.Helvetica, legacy.StyleRegular, 11),
				legacy.MultiCell(0, 6, "This document is rendered deterministically so it can be byte-compared against a committed reference PDF.", legacy.BorderNone, legacy.AlignJustify, false),
				legacy.MoveDown(4),
				legacy.Cell(50, 8, "a plain cell"),
				legacy.NewLine(),
				legacy.CellFormat(50, 8, "bordered center", legacy.BorderFull, legacy.LnNewline, legacy.AlignCenter, legacy.AlignMiddle, false),
				legacy.Save(
					legacy.DrawColor(legacy.Blue),
					legacy.FillColor(legacy.MustHex("#ffcc00")),
					legacy.LineWidth(0.8),
					legacy.Rect(20, 90, 60, 25, legacy.FillStroke),
					legacy.Circle(120, 102, 12, legacy.FillStroke),
				),
				legacy.Line(20, 125, 180, 125),
				legacy.TextAt(20, 140, "positioned label"),
				legacy.Polygon(legacy.Stroke, legacy.Pt(20, 150), legacy.Pt(60, 150), legacy.Pt(40, 180)),
			},
		},
		&native.Document{
			Title:   "go-mx pdf golden",
			Author:  "go-mx",
			Subject: "golden reference",
			Body: native.Components{
				native.Font(native.Helvetica, native.StyleBold, 20),
				native.Paragraph("go-mx/pdf golden reference"),
				native.MoveDown(4),
				native.Font(native.Helvetica, native.StyleRegular, 11),
				native.MultiCell(0, 6, "This document is rendered deterministically so it can be byte-compared against a committed reference PDF.", native.BorderNone, native.AlignJustify, false),
				native.MoveDown(4),
				native.Cell(50, 8, "a plain cell"),
				native.NewLine(),
				native.CellFormat(50, 8, "bordered center", native.BorderFull, native.LnNewline, native.AlignCenter, native.AlignMiddle, false),
				native.Save(
					native.DrawColor(native.Blue),
					native.FillColor(native.MustHex("#ffcc00")),
					native.LineWidth(0.8),
					native.Rect(20, 90, 60, 25, native.FillStroke),
					native.Circle(120, 102, 12, native.FillStroke),
				),
				native.Line(20, 125, 180, 125),
				native.TextAt(20, 140, "positioned label"),
				native.Polygon(native.Stroke, native.Pt(20, 150), native.Pt(60, 150), native.Pt(40, 180)),
			},
		})
}

func TestParityTextFlow(t *testing.T) {
	const long = "The quick brown fox jumps over the lazy dog. Pack my box with five dozen liquor jugs. Sphinx of black quartz, judge my vow. How vexingly quick daft zebras jump!"
	assertParity(t,
		&legacy.Document{
			Title: "text flow",
			Body: legacy.Components{
				legacy.Text("flowing text with automatic wrapping: " + long),
				legacy.NewLine(),
				legacy.Textf("formatted: %d, %.2f, %s", 42, 3.14159, "str"),
				legacy.Ln(8),
				legacy.Paragraph(long),
				legacy.Paragraph("cp1252 characters: café €10 – naïve ½"),
				legacy.MultiCell(80, 5, "justified narrow column with embedded\nline breaks. "+long, legacy.BorderNone, legacy.AlignJustify, false),
				legacy.TextAt(150, 250, "absolutely positioned"),
			},
		},
		&native.Document{
			Title: "text flow",
			Body: native.Components{
				native.Text("flowing text with automatic wrapping: " + long),
				native.NewLine(),
				native.Textf("formatted: %d, %.2f, %s", 42, 3.14159, "str"),
				native.Ln(8),
				native.Paragraph(long),
				native.Paragraph("cp1252 characters: café €10 – naïve ½"),
				native.MultiCell(80, 5, "justified narrow column with embedded\nline breaks. "+long, native.BorderNone, native.AlignJustify, false),
				native.TextAt(150, 250, "absolutely positioned"),
			},
		})
}

func TestParityCells(t *testing.T) {
	assertParity(t,
		&legacy.Document{
			Title: "cells",
			Body: legacy.Components{
				legacy.Cell(40, 8, "plain"),
				legacy.CellFormat(40, 8, "right of it", legacy.BorderFull, legacy.LnNewline, legacy.AlignLeft, legacy.AlignMiddle, false),
				legacy.CellFormat(60, 10, "LT border", legacy.BorderLeftTop, legacy.LnRight, legacy.AlignCenter, legacy.AlignTop, false),
				legacy.CellFormat(60, 10, "RB border", legacy.BorderRightBottom, legacy.LnBelow, legacy.AlignRight, legacy.AlignBottom, false),
				legacy.FillColor(legacy.RGB(220, 235, 255)),
				legacy.CellFormat(80, 12, "filled", legacy.BorderLeftTopRightBottom, legacy.LnNewline, legacy.AlignCenter, legacy.AlignMiddle, true),
				legacy.TextColor(legacy.Red),
				legacy.CellFormat(80, 8, "red baseline", legacy.BorderNone, legacy.LnNewline, legacy.AlignLeft, legacy.AlignBaseline, false),
			},
		},
		&native.Document{
			Title: "cells",
			Body: native.Components{
				native.Cell(40, 8, "plain"),
				native.CellFormat(40, 8, "right of it", native.BorderFull, native.LnNewline, native.AlignLeft, native.AlignMiddle, false),
				native.CellFormat(60, 10, "LT border", native.BorderLeftTop, native.LnRight, native.AlignCenter, native.AlignTop, false),
				native.CellFormat(60, 10, "RB border", native.BorderRightBottom, native.LnBelow, native.AlignRight, native.AlignBottom, false),
				native.FillColor(native.RGB(220, 235, 255)),
				native.CellFormat(80, 12, "filled", native.BorderLeftTopRightBottom, native.LnNewline, native.AlignCenter, native.AlignMiddle, true),
				native.TextColor(native.Red),
				native.CellFormat(80, 8, "red baseline", native.BorderNone, native.LnNewline, native.AlignLeft, native.AlignBaseline, false),
			},
		})
}

func TestParityShapes(t *testing.T) {
	assertParity(t,
		&legacy.Document{
			Title: "shapes",
			Body: legacy.Components{
				legacy.DrawColor(legacy.RGB(30, 60, 120)),
				legacy.FillColor(legacy.RGB(250, 240, 200)),
				legacy.LineWidth(0.7),
				legacy.LineCap(legacy.CapRound),
				legacy.LineJoin(legacy.JoinBevel),
				legacy.Line(10, 10, 200, 10),
				legacy.Rect(10, 20, 50, 30, legacy.Stroke),
				legacy.Rect(70, 20, 50, 30, legacy.FillShape),
				legacy.Rect(130, 20, 50, 30, legacy.FillStroke),
				legacy.RoundedRect(10, 60, 50, 30, 5, legacy.FillStroke),
				legacy.Circle(95, 75, 15, legacy.Stroke),
				legacy.Ellipse(155, 75, 25, 12, 30, legacy.FillStroke),
				legacy.Polygon(legacy.FillStroke, legacy.Pt(30, 100), legacy.Pt(60, 110), legacy.Pt(50, 140), legacy.Pt(15, 130)),
			},
		},
		&native.Document{
			Title: "shapes",
			Body: native.Components{
				native.DrawColor(native.RGB(30, 60, 120)),
				native.FillColor(native.RGB(250, 240, 200)),
				native.LineWidth(0.7),
				native.LineCap(native.CapRound),
				native.LineJoin(native.JoinBevel),
				native.Line(10, 10, 200, 10),
				native.Rect(10, 20, 50, 30, native.Stroke),
				native.Rect(70, 20, 50, 30, native.FillShape),
				native.Rect(130, 20, 50, 30, native.FillStroke),
				native.RoundedRect(10, 60, 50, 30, 5, native.FillStroke),
				native.Circle(95, 75, 15, native.Stroke),
				native.Ellipse(155, 75, 25, 12, 30, native.FillStroke),
				native.Polygon(native.FillStroke, native.Pt(30, 100), native.Pt(60, 110), native.Pt(50, 140), native.Pt(15, 130)),
			},
		})
}

func TestParityStateAndSave(t *testing.T) {
	assertParity(t,
		&legacy.Document{
			Title: "state",
			Body: legacy.Components{
				legacy.XY(30, 40),
				legacy.Text("at 30,40"),
				legacy.MoveDown(10),
				legacy.MoveRight(20),
				legacy.Text("moved"),
				legacy.Save(
					legacy.Font(legacy.Times, legacy.StyleItalic, 14),
					legacy.TextColor(legacy.Red),
					legacy.LineWidth(1.5),
					legacy.Paragraph("scoped: red italic Times"),
					legacy.Save(
						legacy.Font(legacy.Courier, legacy.StyleBold, 9),
						legacy.Paragraph("nested: bold Courier 9"),
					),
					legacy.Paragraph("back to red italic Times"),
				),
				legacy.Paragraph("back to defaults"),
				legacy.X(100),
				legacy.Y(200),
				legacy.Text("via X and Y"),
			},
		},
		&native.Document{
			Title: "state",
			Body: native.Components{
				native.XY(30, 40),
				native.Text("at 30,40"),
				native.MoveDown(10),
				native.MoveRight(20),
				native.Text("moved"),
				native.Save(
					native.Font(native.Times, native.StyleItalic, 14),
					native.TextColor(native.Red),
					native.LineWidth(1.5),
					native.Paragraph("scoped: red italic Times"),
					native.Save(
						native.Font(native.Courier, native.StyleBold, 9),
						native.Paragraph("nested: bold Courier 9"),
					),
					native.Paragraph("back to red italic Times"),
				),
				native.Paragraph("back to defaults"),
				native.X(100),
				native.Y(200),
				native.Text("via X and Y"),
			},
		})
}

func TestParityPagesFontsAndBreaks(t *testing.T) {
	longParagraphs := func() []string {
		var out []string
		for range 30 {
			out = append(out, "Repeated body text that eventually triggers the automatic page break at the bottom margin of every page in this document.")
		}
		return out
	}()

	legacyBody := legacy.Components{
		legacy.Font(legacy.Helvetica, legacy.StyleBoldItalic, 16),
		legacy.Paragraph("Helvetica bold italic"),
		legacy.Font(legacy.Times, legacy.StyleUnderline, 12),
		legacy.Paragraph("Times underlined"),
		legacy.Font(legacy.Courier, legacy.StyleStrikeOut, 11),
		legacy.Paragraph("Courier struck out"),
		legacy.Font(legacy.Symbol, legacy.StyleRegular, 12),
		legacy.Paragraph("abgd"),
		legacy.Font(legacy.ZapfDingbats, legacy.StyleRegular, 12),
		legacy.Paragraph("34"),
		legacy.Font(legacy.Helvetica, legacy.StyleRegular, 12),
		legacy.FontSize(9),
		legacy.Paragraph("9pt via FontSize"),
	}
	nativeBody := native.Components{
		native.Font(native.Helvetica, native.StyleBoldItalic, 16),
		native.Paragraph("Helvetica bold italic"),
		native.Font(native.Times, native.StyleUnderline, 12),
		native.Paragraph("Times underlined"),
		native.Font(native.Courier, native.StyleStrikeOut, 11),
		native.Paragraph("Courier struck out"),
		native.Font(native.Symbol, native.StyleRegular, 12),
		native.Paragraph("abgd"),
		native.Font(native.ZapfDingbats, native.StyleRegular, 12),
		native.Paragraph("34"),
		native.Font(native.Helvetica, native.StyleRegular, 12),
		native.FontSize(9),
		native.Paragraph("9pt via FontSize"),
	}
	for _, p := range longParagraphs {
		legacyBody = append(legacyBody, legacy.Paragraph(p), legacy.NewLine())
		nativeBody = append(nativeBody, native.Paragraph(p), native.NewLine())
	}
	legacyBody = append(legacyBody,
		legacy.PageFormat(legacy.Landscape, legacy.A5, legacy.Paragraph("landscape A5 page")),
		legacy.Page(legacy.Paragraph("back on a default page")),
	)
	nativeBody = append(nativeBody,
		native.PageFormat(native.OrientationLandscape, native.PageSizeA5, native.Paragraph("landscape A5 page")),
		native.Page(native.Paragraph("back on a default page")),
	)

	assertParity(t,
		&legacy.Document{
			Title:   "pages and fonts",
			Margins: &legacy.Margins{Left: 25, Top: 30, Right: 20},
			Body:    legacyBody,
		},
		&native.Document{
			Title:   "pages and fonts",
			Margins: &native.Margins{Left: 25, Top: 30, Right: 20},
			Body:    nativeBody,
		})
}

func TestParityHeaderFooterMetadata(t *testing.T) {
	assertParity(t,
		&legacy.Document{
			Title:    "header and footer",
			Author:   "parity author",
			Subject:  "parity subject",
			Keywords: "pdf, parity, test",
			Creator:  "parity creator",
			Header: legacy.Components{
				legacy.Font(legacy.Helvetica, legacy.StyleBold, 9),
				legacy.CellFormat(0, 8, "running header", legacy.BorderBottom, legacy.LnNewline, legacy.AlignCenter, legacy.AlignMiddle, false),
				legacy.MoveDown(2),
			},
			Footer: legacy.ComponentFunc(func(ctx context.Context, r *legacy.Renderer) error {
				r.SetY(-15)
				r.SetFont(legacy.Helvetica, "I", 8)
				return legacy.Textf("page %d", r.PageNo()).Render(ctx, r)
			}),
			Body: legacy.Components{
				legacy.Paragraph("first page body"),
				legacy.Page(legacy.Paragraph("second page body")),
				legacy.Page(legacy.Paragraph("third page body")),
			},
		},
		&native.Document{
			Title:    "header and footer",
			Author:   "parity author",
			Subject:  "parity subject",
			Keywords: "pdf, parity, test",
			Creator:  "parity creator",
			Header: native.Components{
				native.Font(native.Helvetica, native.StyleBold, 9),
				native.CellFormat(0, 8, "running header", native.BorderBottom, native.LnNewline, native.AlignCenter, native.AlignMiddle, false),
				native.MoveDown(2),
			},
			Footer: native.ComponentFunc(func(ctx context.Context, r *native.Renderer) error {
				r.SetY(-15)
				r.SetFont(native.Helvetica, "I", 8)
				return native.Textf("page %d", r.PageNo()).Render(ctx, r)
			}),
			Body: native.Components{
				native.Paragraph("first page body"),
				native.Page(native.Paragraph("second page body")),
				native.Page(native.Paragraph("third page body")),
			},
		})
}

func TestParityImages(t *testing.T) {
	pngData := pngImageBytes(t)
	jpegData := jpegImageBytes(t)
	gifData := gifImageBytes(t)
	// The legacy ImageBytes consumes its reader on first render, so fresh
	// documents are built per compression variant instead of re-rendering one
	// document as assertParity does. (The native ImageBytes is re-renderable —
	// it creates a fresh reader per render.)
	for _, compress := range []bool{false, true} {
		legacyDoc := &legacy.Document{
			Title: "images",
			Body: legacy.Components{
				legacy.ImageBytes("png8", legacy.ImagePNG, pngData, 20, 20, 30, 0),
				legacy.ImageBytes("jpeg8", legacy.ImageJPEG, jpegData, 60, 20, 30, 30),
				legacy.ImageBytes("gif8", legacy.ImageGIF, gifData, 100, 20, 0, 30),
				// draw the cached PNG a second time under the same name
				legacy.ImageBytes("png8", legacy.ImagePNG, pngData, 140, 20, 20, 20),
			},
		}
		nativeDoc := &native.Document{
			Title: "images",
			Body: native.Components{
				native.ImageBytes("png8", native.ImagePNG, pngData, 20, 20, 30, 0),
				native.ImageBytes("jpeg8", native.ImageJPEG, jpegData, 60, 20, 30, 30),
				native.ImageBytes("gif8", native.ImageGIF, gifData, 100, 20, 0, 30),
				native.ImageBytes("png8", native.ImagePNG, pngData, 140, 20, 20, 20),
			},
		}
		assertSamePDF(t, compress, renderLegacy(t, legacyDoc, compress), renderNative(t, nativeDoc, compress))
	}
}

func TestParityConditionals(t *testing.T) {
	items := []string{"alpha", "beta", "gamma"}
	assertParity(t,
		&legacy.Document{
			Title: "conditionals",
			Body: legacy.Components{
				legacy.If(true, legacy.Paragraph("rendered")).Else(legacy.Paragraph("not rendered")),
				legacy.If(false, legacy.Paragraph("not rendered")).Else(legacy.Paragraph("fallback rendered")),
				legacy.ForEach(items, func(s string) legacy.Component {
					return legacy.Paragraph("item " + s)
				}),
			},
		},
		&native.Document{
			Title: "conditionals",
			Body: native.Components{
				native.If(true, native.Paragraph("rendered")).Else(native.Paragraph("not rendered")),
				native.If(false, native.Paragraph("not rendered")).Else(native.Paragraph("fallback rendered")),
				native.ForEach(items, func(s string) native.Component {
					return native.Paragraph("item " + s)
				}),
			},
		})
}
