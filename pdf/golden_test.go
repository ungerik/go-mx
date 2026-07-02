package pdf

import (
	"bytes"
	"context"
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var updateGolden = flag.Bool("update", false, "update the golden reference PDF in testdata/")

// goldenDocument builds a representative document exercising text, cells,
// shapes, colors and a Save scope. It is deterministic apart from the
// creation/modification dates, which renderGolden pins.
func goldenDocument() *Document {
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
}

// renderGolden renders goldenDocument with pinned dates and compression off so
// the output is byte-for-byte reproducible.
func renderGolden() ([]byte, error) {
	doc := goldenDocument()
	r := doc.NewRenderer()
	fixed := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	r.SetCreationDate(fixed)
	r.SetModificationDate(fixed)
	r.SetCompression(false)
	// fpdf assigns PDF object numbers to fonts and images by iterating Go maps,
	// whose order is randomized per process; SetCatalogSort sorts those keys so
	// the byte output is stable across runs (the pinned dates alone are not
	// enough — a document with more than one font is otherwise flaky).
	r.SetCatalogSort(true)
	if err := doc.Render(context.Background(), r); err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := r.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// TestGoldenDeterministic guards the golden comparison itself: two renders must
// be byte-identical, otherwise TestGoldenPDF would be flaky.
func TestGoldenDeterministic(t *testing.T) {
	a, err := renderGolden()
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	b, err := renderGolden()
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if !bytes.Equal(a, b) {
		t.Fatal("golden render is not deterministic across runs")
	}
}

// TestGoldenPDF renders the document and compares it against the committed
// reference PDF. Regenerate the reference after an intended change with:
//
//	go test ./pdf -run TestGoldenPDF -update
func TestGoldenPDF(t *testing.T) {
	got, err := renderGolden()
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	path := filepath.Join("testdata", "reference.pdf")

	if *updateGolden {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, got, 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote %s (%d bytes)", path, len(got))
		return
	}

	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden reference (create it with `go test ./pdf -run TestGoldenPDF -update`): %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("rendered PDF differs from %s (%d vs %d bytes); if the change is intended, regenerate with `go test ./pdf -run TestGoldenPDF -update`",
			path, len(got), len(want))
	}
}
