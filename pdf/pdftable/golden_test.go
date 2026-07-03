package pdftable

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ungerik/go-mx/pdf"
	"github.com/ungerik/go-mx/svg"
)

var updateGolden = flag.Bool("update", false, "update the golden reference PDF in testdata/")

// goldenDocument renders a gallery of table features for visual judgement:
// an invoice-style table that breaks across pages with a repeated header,
// zebra striping, mixed column sizing, alignment, multi-line and wrapped
// cells, a Draw cell with an SVG icon, and grid style variants.
func goldenDocument() *pdf.Document {
	invoice := New([]Column{
		{Title: "Pos", Width: 12, Style: &Style{HAlign: pdf.AlignRight}},
		{Title: "", Width: 14}, // icon column
		{Title: "Article", Weight: 1},
		{Title: "Description", Weight: 2, Style: &Style{
			FontSize:  9,
			TextColor: &pdf.Gray50,
		}},
		{Title: "Qty", Style: &Style{HAlign: pdf.AlignRight}}, // auto width
		{Title: "Price", Width: 22, Style: &Style{HAlign: pdf.AlignRight}},
	})
	invoice.HeaderStyle = Style{FillColor: &pdf.Silver}

	icon := svg.SVG(
		svg.ViewBox(0, 0, 24, 24),
		svg.Circle(svg.CX(12), svg.CY(12), svg.R(10), svg.Fill("tomato")),
		svg.Path(svg.D("M7 12 L11 16 L17 8"), svg.Fill("none"),
			svg.Stroke("white"), svg.StrokeWidth(2.5)),
	)
	descriptions := []string{
		"Standard configuration",
		"With reinforced frame,\nweather-proof coating and a very long remark that has to wrap within its column to prove the row grows with it",
		"Bulk package",
	}
	for i := range 40 {
		row := Row{Cells: []Cell{
			{Text: fmt.Sprint(i + 1)},
			{
				Draw: func(ctx context.Context, r *pdf.Renderer, x, y, w, h float64) error {
					return pdf.SVG(icon, x, y, min(w, h), min(w, h)).Render(ctx, r)
				},
				Height: 6,
			},
			{Text: fmt.Sprintf("Article %c-%d", 'A'+i%26, i+1)},
			{Text: descriptions[i%len(descriptions)]},
			{Text: fmt.Sprint(i%12 + 1)},
			{Text: fmt.Sprintf("%d.%02d", (i*37)%90+10, (i*53)%100)},
		}}
		if i%2 == 1 {
			row.Style = &Style{FillColor: &pdf.Color{R: 235, G: 240, B: 248}}
		}
		invoice.Rows = append(invoice.Rows, row)
	}
	invoice.Rows = append(invoice.Rows, Row{
		Style: &Style{FontStyle: new(pdf.StyleBold)},
		Cells: []Cell{{}, {}, {Text: "Total"}, {}, {}, {Text: "1234.56"}},
	})

	gridDemo := func(grid Grid, title string) pdf.Components {
		table := New([]Column{
			{Title: "Grid", Weight: 1},
			{Title: title, Weight: 2},
		},
			[]string{"top", "middle line\nsecond line"},
			[]string{"bottom", "last row"},
		)
		table.Grid = grid
		table.Width = 80
		return pdf.Components{table, pdf.MoveDown(6)}
	}

	return &pdf.Document{
		Title: "pdftable reference",
		Body: pdf.Components{
			// Not pdf.Save: it would restore the cursor position and undo
			// the paragraph's advance, so reset the font explicitly.
			pdf.Font(pdf.Helvetica, pdf.StyleBold, 16),
			pdf.Paragraph("pdftable visual reference"),
			pdf.Font(pdf.Helvetica, pdf.StyleRegular, 12),
			pdf.MoveDown(4),
			invoice,
			pdf.MoveDown(8),
			pdf.Font(pdf.Helvetica, pdf.StyleBold, 12),
			pdf.Paragraph("Grid styles"),
			pdf.Font(pdf.Helvetica, pdf.StyleRegular, 12),
			pdf.MoveDown(2),
			gridDemo(GridAll, "GridAll"),
			gridDemo(GridHeaderRows, "GridHeaderRows"),
			gridDemo(GridOuterHeader, "GridOuterHeader"),
			gridDemo(GridNone, "GridNone"),
		},
	}
}

// renderGolden renders goldenDocument byte-for-byte reproducibly, mirroring
// the golden tests of the pdf package.
func renderGolden() ([]byte, error) {
	doc := goldenDocument()
	r := doc.NewRenderer()
	fixed := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	r.SetCreationDate(fixed)
	r.SetModificationDate(fixed)
	r.SetCompression(false)
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

// TestGoldenDeterministic guards the golden comparison itself: two renders
// must be byte-identical, otherwise TestGoldenPDF would be flaky.
func TestGoldenDeterministic(t *testing.T) {
	a, err := renderGolden()
	if err != nil {
		t.Fatal(err)
	}
	b, err := renderGolden()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(a, b) {
		t.Error("two golden renders differ; the golden comparison would be flaky")
	}
}

// TestGoldenPDF compares the rendered reference document byte-for-byte with
// the committed testdata/table_reference.pdf. Regenerate after intentional
// changes with:
//
//	go test ./pdf/pdftable -run TestGoldenPDF -update
func TestGoldenPDF(t *testing.T) {
	got, err := renderGolden()
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	path := filepath.Join("testdata", "table_reference.pdf")

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
		t.Fatalf("read golden reference (create it with `go test ./pdf/pdftable -run TestGoldenPDF -update`): %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("rendered PDF differs from %s (%d vs %d bytes); if the change is intended, regenerate with `go test ./pdf/pdftable -run TestGoldenPDF -update`",
			path, len(got), len(want))
	}
}
