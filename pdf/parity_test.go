package pdf

import (
	"bytes"
	"os"
	"testing"
	"time"
)

// TestEngineGoldenParity replays, as raw engine calls, the exact call sequence
// that the legacy fpdf wrapper's golden test (fpdf/golden_test.go) issues
// through its component layer, and requires the output to be byte-identical to
// the committed reference PDF rendered by the legacy stack. This proves the
// inlined engine produces the same bytes as codeberg.org/go-pdf/fpdf v0.12.0
// before the component layer is ported on top of it.
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

	want, err := os.ReadFile("../fpdf/testdata/reference.pdf")
	if err != nil {
		t.Fatalf("read legacy reference: %v", err)
	}
	if !bytes.Equal(buf.Bytes(), want) {
		t.Errorf("engine output differs from legacy reference.pdf (%d vs %d bytes)", buf.Len(), len(want))
		if err := CompareBytes(buf.Bytes(), want, true); err != nil {
			t.Log(err)
		}
	}
}
