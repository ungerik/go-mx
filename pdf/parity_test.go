package pdf

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestEngineGoldenParity replays, as raw engine calls, the exact call sequence
// that the golden test of the legacy fpdf wrapper (golden_test.go in
// github.com/ungerik/go-mx-fpdf) issues through its component layer, and
// requires the output to be byte-identical to the reference PDF rendered by
// the legacy stack. This proves the inlined engine produces the same bytes as
// codeberg.org/go-pdf/fpdf v0.12.0 before the component layer is ported on top
// of it.
//
// testdata/legacy_reference.pdf is a copy of that repository's
// testdata/reference.pdf. It is kept apart from testdata/reference.pdf, which
// TestGoldenPDF -update rewrites from this engine's own output.
func TestEngineGoldenParity(t *testing.T) {
	// Document.NewRenderer: New + default font.
	f := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	f.SetFont("Helvetica", "", 12)

	// renderGolden determinism setup.
	fixed := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	f.SetCreationDate(fixed)
	f.SetModificationDate(fixed)
	f.SetCompression(false)
	f.SetCatalogSort(true)

	// Document.applySetup: metadata, then the default font again.
	f.SetTitle("go-mx pdf golden", true)
	f.SetAuthor("go-mx", true)
	f.SetSubject("golden reference", true)
	f.SetFont("Helvetica", "", 12)

	// The wrapper's automatic line height: 1.15 × current font size in units.
	lineHt := func() float64 {
		_, unitSize := f.GetFontSize()
		return unitSize * 1.15
	}

	// Body. State components (Font, MoveDown) do not open a page; the first
	// drawing primitive (Paragraph) adds page one before drawing.
	f.SetFont("Helvetica", "B", 20) // Font(Helvetica, StyleBold, 20)
	f.AddPage()
	f.MultiCell(0, lineHt(), "go-mx/pdf golden reference", "", "L", false) // Paragraph
	f.SetY(f.GetY() + 4)                                                   // MoveDown(4)
	f.SetFont("Helvetica", "", 11)                                         // Font(Helvetica, StyleRegular, 11)
	f.MultiCell(0, 6, "This document is rendered deterministically so it can be byte-compared against a committed reference PDF.", "", "J", false)
	f.SetY(f.GetY() + 4) // MoveDown(4)
	f.Cell(50, 8, "a plain cell")
	f.Ln(lineHt()) // NewLine()
	f.CellFormat(50, 8, "bordered center", "1", 1, "CM", false, 0, "")

	// Save(...): capture via getters, render children, restore via setters.
	family := f.GetFontFamily()
	style := f.GetFontStyle()
	sizePt, _ := f.GetFontSize()
	tr, tg, tb := f.GetTextColor()
	fr, fg, fb := f.GetFillColor()
	dr, dg, db := f.GetDrawColor()
	lineWidth := f.GetLineWidth()
	capStyle := f.GetLineCapStyle()
	joinStyle := f.GetLineJoinStyle()
	x, y := f.GetXY()

	f.SetDrawColor(0, 0, 255)    // DrawColor(Blue)
	f.SetFillColor(255, 204, 0)  // FillColor(MustHex("#ffcc00"))
	f.SetLineWidth(0.8)          // LineWidth(0.8)
	f.Rect(20, 90, 60, 25, "FD") // Rect(…, FillStroke)
	f.Circle(120, 102, 12, "FD") // Circle(…, FillStroke)

	f.SetFont(family, style, sizePt)
	f.SetTextColor(tr, tg, tb)
	f.SetFillColor(fr, fg, fb)
	f.SetDrawColor(dr, dg, db)
	f.SetLineWidth(lineWidth)
	f.SetLineCapStyle(capStyle)
	f.SetLineJoinStyle(joinStyle)
	f.SetXY(x, y)

	f.Line(20, 125, 180, 125)
	f.Text(20, 140, "positioned label") // TextAt
	f.Polygon([]Point{{X: 20, Y: 150}, {X: 60, Y: 150}, {X: 40, Y: 180}}, "D")

	var buf bytes.Buffer
	if err := f.Output(&buf); err != nil {
		t.Fatalf("output: %v", err)
	}

	want, err := os.ReadFile("testdata/legacy_reference.pdf")
	if err != nil {
		t.Fatalf("read legacy reference: %v", err)
	}
	if !bytes.Equal(buf.Bytes(), want) {
		t.Errorf("engine output differs from legacy_reference.pdf (%d vs %d bytes)", buf.Len(), len(want))
		if err := CompareBytes(buf.Bytes(), want, true); err != nil {
			t.Log(err)
		}
	}
}

// TestLegacyParity renders the scenarios of the parity suite of the legacy fpdf
// wrapper (github.com/ungerik/go-mx-fpdf, around codeberg.org/go-pdf/fpdf
// v0.12.0) through this package and requires output byte-identical to what the
// legacy stack rendered for the same documents. It keeps that suite's coverage
// (text flow, cells, shapes, saved state, page breaks and fonts, header and
// footer, images, conditionals) on every change to this package, now that the
// legacy wrapper lives in its own repository.
//
// The references in testdata/parity were rendered once by the legacy stack,
// with the dates pinned and compression off; the golden scenario uses
// testdata/legacy_reference.pdf. Compression stays off because compressed
// streams depend on the zlib encoder of the Go release; go-mx-fpdf's CI still
// compares the compressed output of both stacks. The images scenario embeds the
// committed image files rather than encoding images at test time, for the same
// reason. Never regenerate a reference from this package, which would compare
// the engine with itself: render it again with the legacy stack, or list a
// deliberate divergence in README.md and change the test.
func TestLegacyParity(t *testing.T) {
	const long = "The quick brown fox jumps over the lazy dog. Pack my box with five dozen liquor jugs. Sphinx of black quartz, judge my vow. How vexingly quick daft zebras jump!"
	readFixture := func(t *testing.T, name string) []byte {
		t.Helper()
		data, err := os.ReadFile(filepath.Join("testdata", "parity", name))
		if err != nil {
			t.Fatal(err)
		}
		return data
	}

	for _, tc := range []struct {
		name      string
		reference string // below testdata
		document  func(t *testing.T) *Document
	}{
		{"golden", "legacy_reference.pdf", func(*testing.T) *Document {
			return &Document{
				Title:   "go-mx pdf golden",
				Author:  "go-mx",
				Subject: "golden reference",
				Body: Components{
					Font(Helvetica, StyleBold, 20),
					Paragraph("go-mx/pdf golden reference"),
					MoveDown(4),
					Font(Helvetica, StyleRegular, 11),
					MultiCell(0, 6, "This document is rendered deterministically so it can be byte-compared against a committed reference PDF.", BorderNone, AlignJustify, false),
					MoveDown(4),
					Cell(50, 8, "a plain cell"),
					NewLine(),
					CellFormat(50, 8, "bordered center", BorderFull, LnNewline, AlignCenter, AlignMiddle, false),
					Save(
						DrawColor(Blue),
						FillColor(MustHex("#ffcc00")),
						LineWidth(0.8),
						Rect(20, 90, 60, 25, FillStroke),
						Circle(120, 102, 12, FillStroke),
					),
					Line(20, 125, 180, 125),
					TextAt(20, 140, "positioned label"),
					Polygon(Stroke, Pt(20, 150), Pt(60, 150), Pt(40, 180)),
				},
			}
		}},
		{"text_flow", "parity/legacy_text_flow.pdf", func(*testing.T) *Document {
			return &Document{
				Title: "text flow",
				Body: Components{
					Text("flowing text with automatic wrapping: " + long),
					NewLine(),
					Textf("formatted: %d, %.2f, %s", 42, 3.14159, "str"),
					Ln(8),
					Paragraph(long),
					Paragraph("cp1252 characters: café €10 – naïve ½"),
					MultiCell(80, 5, "justified narrow column with embedded\nline breaks. "+long, BorderNone, AlignJustify, false),
					TextAt(150, 250, "absolutely positioned"),
				},
			}
		}},
		{"cells", "parity/legacy_cells.pdf", func(*testing.T) *Document {
			return &Document{
				Title: "cells",
				Body: Components{
					Cell(40, 8, "plain"),
					CellFormat(40, 8, "right of it", BorderFull, LnNewline, AlignLeft, AlignMiddle, false),
					CellFormat(60, 10, "LT border", BorderLeftTop, LnRight, AlignCenter, AlignTop, false),
					CellFormat(60, 10, "RB border", BorderRightBottom, LnBelow, AlignRight, AlignBottom, false),
					FillColor(RGB(220, 235, 255)),
					CellFormat(80, 12, "filled", BorderLeftTopRightBottom, LnNewline, AlignCenter, AlignMiddle, true),
					TextColor(Red),
					CellFormat(80, 8, "red baseline", BorderNone, LnNewline, AlignLeft, AlignBaseline, false),
				},
			}
		}},
		{"shapes", "parity/legacy_shapes.pdf", func(*testing.T) *Document {
			return &Document{
				Title: "shapes",
				Body: Components{
					DrawColor(RGB(30, 60, 120)),
					FillColor(RGB(250, 240, 200)),
					LineWidth(0.7),
					LineCap(CapRound),
					LineJoin(JoinBevel),
					Line(10, 10, 200, 10),
					Rect(10, 20, 50, 30, Stroke),
					Rect(70, 20, 50, 30, FillShape),
					Rect(130, 20, 50, 30, FillStroke),
					RoundedRect(10, 60, 50, 30, 5, FillStroke),
					Circle(95, 75, 15, Stroke),
					Ellipse(155, 75, 25, 12, 30, FillStroke),
					Polygon(FillStroke, Pt(30, 100), Pt(60, 110), Pt(50, 140), Pt(15, 130)),
				},
			}
		}},
		{"state_and_save", "parity/legacy_state_and_save.pdf", func(*testing.T) *Document {
			return &Document{
				Title: "state",
				Body: Components{
					XY(30, 40),
					Text("at 30,40"),
					MoveDown(10),
					MoveRight(20),
					Text("moved"),
					Save(
						Font(Times, StyleItalic, 14),
						TextColor(Red),
						LineWidth(1.5),
						Paragraph("scoped: red italic Times"),
						Save(
							Font(Courier, StyleBold, 9),
							Paragraph("nested: bold Courier 9"),
						),
						Paragraph("back to red italic Times"),
					),
					Paragraph("back to defaults"),
					X(100),
					Y(200),
					Text("via X and Y"),
				},
			}
		}},
		{"pages_fonts_and_breaks", "parity/legacy_pages_fonts_and_breaks.pdf", func(*testing.T) *Document {
			body := Components{
				Font(Helvetica, StyleBoldItalic, 16),
				Paragraph("Helvetica bold italic"),
				Font(Times, StyleUnderline, 12),
				Paragraph("Times underlined"),
				Font(Courier, StyleStrikeOut, 11),
				Paragraph("Courier struck out"),
				Font(Symbol, StyleRegular, 12),
				Paragraph("abgd"),
				Font(ZapfDingbats, StyleRegular, 12),
				Paragraph("34"),
				Font(Helvetica, StyleRegular, 12),
				FontSize(9),
				Paragraph("9pt via FontSize"),
			}
			// Enough text to trigger automatic page breaks at the bottom margin.
			for range 30 {
				body = append(body,
					Paragraph("Repeated body text that eventually triggers the automatic page break at the bottom margin of every page in this document."),
					NewLine(),
				)
			}
			body = append(body,
				PageFormat(OrientationLandscape, PageSizeA5, Paragraph("landscape A5 page")),
				Page(Paragraph("back on a default page")),
			)
			return &Document{
				Title:   "pages and fonts",
				Margins: &Margins{Left: 25, Top: 30, Right: 20},
				Body:    body,
			}
		}},
		{"header_footer_metadata", "parity/legacy_header_footer_metadata.pdf", func(*testing.T) *Document {
			return &Document{
				Title:    "header and footer",
				Author:   "parity author",
				Subject:  "parity subject",
				Keywords: "pdf, parity, test",
				Creator:  "parity creator",
				Header: Components{
					Font(Helvetica, StyleBold, 9),
					CellFormat(0, 8, "running header", BorderBottom, LnNewline, AlignCenter, AlignMiddle, false),
					MoveDown(2),
				},
				Footer: ComponentFunc(func(ctx context.Context, r *Renderer) error {
					r.SetY(-15)
					r.SetFont(Helvetica, "I", 8)
					return Textf("page %d", r.PageNo()).Render(ctx, r)
				}),
				Body: Components{
					Paragraph("first page body"),
					Page(Paragraph("second page body")),
					Page(Paragraph("third page body")),
				},
			}
		}},
		{"images", "parity/legacy_images.pdf", func(t *testing.T) *Document {
			pngData := readFixture(t, "image.png")
			return &Document{
				Title: "images",
				Body: Components{
					ImageBytes("png8", ImagePNG, pngData, 20, 20, 30, 0),
					ImageBytes("jpeg8", ImageJPEG, readFixture(t, "image.jpg"), 60, 20, 30, 30),
					ImageBytes("gif8", ImageGIF, readFixture(t, "image.gif"), 100, 20, 0, 30),
					// The cached PNG drawn a second time under the same name.
					ImageBytes("png8", ImagePNG, pngData, 140, 20, 20, 20),
				},
			}
		}},
		{"conditionals", "parity/legacy_conditionals.pdf", func(*testing.T) *Document {
			return &Document{
				Title: "conditionals",
				Body: Components{
					If(true, Paragraph("rendered")).Else(Paragraph("not rendered")),
					If(false, Paragraph("not rendered")).Else(Paragraph("fallback rendered")),
					ForEach([]string{"alpha", "beta", "gamma"}, func(s string) Component {
						return Paragraph("item " + s)
					}),
				},
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			doc := tc.document(t)
			r := doc.NewRenderer()
			parityDate := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
			r.SetCreationDate(parityDate)
			r.SetModificationDate(parityDate)
			r.SetCompression(false)
			r.SetCatalogSort(true)
			err := doc.Render(t.Context(), r)
			if err != nil {
				t.Fatalf("render: %v", err)
			}
			var got bytes.Buffer
			err = r.Output(&got)
			if err != nil {
				t.Fatalf("output: %v", err)
			}

			want := readReference(t, tc.reference)
			if !bytes.Equal(got.Bytes(), want) {
				i := firstDifference(got.Bytes(), want)
				t.Errorf("output (%d bytes) differs from testdata/%s (%d bytes) at byte %d: got %q, want %q",
					got.Len(), tc.reference, len(want), i, excerpt(got.Bytes(), i), excerpt(want, i))
			}
		})
	}
}

// readReference reads the reference PDF at path below testdata.
func readReference(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", path))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// firstDifference returns the index of the first byte where a and b differ,
// or the length of the shorter one when it is a prefix of the other.
func firstDifference(a, b []byte) int {
	n := min(len(a), len(b))
	for i := range n {
		if a[i] != b[i] {
			return i
		}
	}
	return n
}

// excerpt returns up to 40 bytes of data starting a few bytes before i, to
// show where two PDFs diverge.
func excerpt(data []byte, i int) []byte {
	start := max(i-8, 0)
	return data[start:min(i+32, len(data))]
}
