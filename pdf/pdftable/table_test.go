package pdftable

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/ungerik/go-mx/pdf"
)

func TestNewAndAddRow(t *testing.T) {
	table := New(
		[]Column{{Title: "A"}, {Title: "B"}},
		[]string{"1", "2"},
	).AddRow("3", 4)
	if table.Grid != GridAll || !table.RepeatHeader {
		t.Errorf("New must default to GridAll and RepeatHeader, got %q %v", table.Grid, table.RepeatHeader)
	}
	if len(table.Rows) != 2 {
		t.Fatalf("got %d rows, want 2", len(table.Rows))
	}
	if table.Rows[0].Cells[1].Text != "2" || table.Rows[1].Cells[1].Text != "4" {
		t.Errorf("unexpected cell texts: %+v", table.Rows)
	}
}

func TestAsCell(t *testing.T) {
	if got := AsCell(nil); got.Text != "" || got.Style != nil || got.Draw != nil {
		t.Errorf("nil: got %+v", got)
	}
	cell := Cell{Text: "x", Style: &Style{HAlign: pdf.AlignRight}}
	if got := AsCell(cell); got.Text != "x" || got.Style != cell.Style {
		t.Errorf("Cell: got %+v", got)
	}
	if got := AsCell(&cell); got.Text != "x" || got.Style != cell.Style {
		t.Errorf("*Cell: got %+v", got)
	}
	if got := AsCell((*Cell)(nil)); got.Text != "" || got.Style != nil {
		t.Errorf("nil *Cell: got %+v", got)
	}
	if got := AsCell("text"); got.Text != "text" {
		t.Errorf("string: got %+v", got)
	}
	if got := AsCell(errors.New("boom")); got.Text != "boom" {
		t.Errorf("error: got %+v", got)
	}
	if got := AsCell(GridAll); got.Text != "OHRC" { // Grid is a fmt.Stringer
		t.Errorf("Stringer: got %+v", got)
	}
	if got := AsCell(42); got.Text != "42" {
		t.Errorf("int: got %+v", got)
	}
}

func testTable(numRows int) *Table {
	table := New([]Column{
		{Title: "Item", Weight: 1},
		{Title: "Qty", Style: &Style{HAlign: pdf.AlignRight}},
		{Title: "Price", Style: &Style{HAlign: pdf.AlignRight}},
	})
	for i := range numRows {
		table.AddRow(fmt.Sprintf("Item number %d", i+1), fmt.Sprint(i%9+1), "19.99")
	}
	return table
}

func TestTableRenderBasic(t *testing.T) {
	r := pdf.NewRendererA4Portrait()
	r.AddPage()
	startY := r.GetY()

	err := testTable(3).Render(context.Background(), r)
	if err != nil {
		t.Fatal(err)
	}
	if r.PageNo() != 1 {
		t.Errorf("a 3-row table must fit one page, got %d pages", r.PageNo())
	}
	if r.GetY() <= startY {
		t.Errorf("cursor must move below the table: startY %g, y %g", startY, r.GetY())
	}
	left, _, _, _ := r.GetMargins()
	if r.GetX() != left {
		t.Errorf("cursor x must return to the left margin %g, got %g", left, r.GetX())
	}
}

func TestTablePageBreak(t *testing.T) {
	r := pdf.NewRendererA4Portrait()
	err := testTable(100).Render(context.Background(), r)
	if err != nil {
		t.Fatal(err)
	}
	if r.PageNo() < 2 {
		t.Errorf("a 100-row table must break pages, got %d", r.PageNo())
	}
}

func TestTableRowsNeverSplit(t *testing.T) {
	// Rows must move to the next page whole: with rows of a known height,
	// every page holds exactly floor(contentHeight/rowHeight) rows, so the
	// page count is predictable.
	r := pdf.NewRendererA4Portrait()
	r.AddPage()
	const rowHeight = 20
	table := New([]Column{{Weight: 1}})
	table.Grid = GridNone
	for range 30 {
		table.Rows = append(table.Rows, Row{Cells: []Cell{{Text: "row"}}, MinHeight: rowHeight})
	}
	err := table.Render(context.Background(), r)
	if err != nil {
		t.Fatal(err)
	}
	rowsPerPage := math.Floor(r.ContentHeight() / rowHeight)
	wantPages := int(math.Ceil(30 / rowsPerPage))
	if r.PageNo() != wantPages {
		t.Errorf("got %d pages, want %d (%g rows per page)", r.PageNo(), wantPages, rowsPerPage)
	}
}

func TestTableSplitsRowTallerThanPage(t *testing.T) {
	r := pdf.NewRendererA4Portrait()
	// One cell with enough text lines to exceed a whole page.
	table := New([]Column{{Weight: 1}}).
		AddRow(strings.Repeat("The quick brown fox jumps over the lazy dog. ", 400))
	err := table.Render(context.Background(), r)
	if err != nil {
		t.Fatal(err)
	}
	if r.PageNo() < 2 {
		t.Errorf("a row taller than a page must split across pages, got %d", r.PageNo())
	}
}

func TestTableRestoresRendererState(t *testing.T) {
	r := pdf.NewRendererA4Portrait()
	r.AddPage()
	r.SetFont(pdf.Times, string(pdf.StyleItalic), 9)
	r.SetTextColor(1, 2, 3)
	r.SetFillColor(4, 5, 6)
	r.SetCellMargin(2.5)
	r.SetAutoPageBreak(true, 25)

	table := testTable(60) // breaks pages, restore must still hold
	table.Style.FillColor = &pdf.Silver
	if err := table.Render(context.Background(), r); err != nil {
		t.Fatal(err)
	}

	if family := r.GetFontFamily(); family != "times" {
		t.Errorf("font family not restored: %q", family)
	}
	if style := r.GetFontStyle(); style != "I" {
		t.Errorf("font style not restored: %q", style)
	}
	if pt, _ := r.GetFontSize(); pt != 9 {
		t.Errorf("font size not restored: %g", pt)
	}
	if cr, cg, cb := r.GetTextColor(); cr != 1 || cg != 2 || cb != 3 {
		t.Errorf("text color not restored: %d %d %d", cr, cg, cb)
	}
	if cr, cg, cb := r.GetFillColor(); cr != 4 || cg != 5 || cb != 6 {
		t.Errorf("fill color not restored: %d %d %d", cr, cg, cb)
	}
	if margin := r.GetCellMargin(); margin != 2.5 {
		t.Errorf("cell margin not restored: %g", margin)
	}
	if auto, margin := r.GetAutoPageBreak(); !auto || margin != 25 {
		t.Errorf("auto page break not restored: %v %g", auto, margin)
	}
}

func TestTableErrors(t *testing.T) {
	ctx := context.Background()

	err := (&Table{}).Render(ctx, pdf.NewRendererA4Portrait())
	if err == nil {
		t.Error("expected error for a table without columns")
	}

	tooMany := New([]Column{{Weight: 1}}).AddRow("a", "b")
	err = tooMany.Render(ctx, pdf.NewRendererA4Portrait())
	if err == nil {
		t.Error("expected error for a row with more cells than columns")
	}

	tooWide := New([]Column{{Width: 300}}, []string{"x"})
	err = tooWide.Render(ctx, pdf.NewRendererA4Portrait())
	if err == nil {
		t.Error("expected error for fixed widths exceeding the table width")
	}
}

func TestTableRaggedRowPadded(t *testing.T) {
	r := pdf.NewRendererA4Portrait()
	table := New([]Column{{Weight: 1}, {Weight: 1}}).AddRow("only one cell")
	if err := table.Render(context.Background(), r); err != nil {
		t.Fatal(err)
	}
}

func TestTableDrawCell(t *testing.T) {
	r := pdf.NewRendererA4Portrait()
	var box [4]float64
	called := 0
	table := New([]Column{{Width: 50}, {Weight: 1}})
	table.Rows = append(table.Rows, Row{Cells: []Cell{
		{Draw: func(_ context.Context, _ *pdf.Renderer, x, y, w, h float64) error {
			called++
			box = [4]float64{x, y, w, h}
			return nil
		}, Height: 30},
		{Text: "beside the drawing"},
	}})
	if err := table.Render(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	if called != 1 {
		t.Fatalf("Draw called %d times, want 1", called)
	}
	pad := r.GetCellMargin() // default padding inherits the cell margin
	if math.Abs(box[2]-(50-2*pad)) > widthEps {
		t.Errorf("Draw content width %g, want %g", box[2], 50-2*pad)
	}
	if math.Abs(box[3]-30) > widthEps {
		t.Errorf("Draw content height %g, want 30", box[3])
	}

	sentinel := errors.New("draw failed")
	failing := New([]Column{{Weight: 1}})
	failing.Rows = append(failing.Rows, Row{Cells: []Cell{
		{Draw: func(context.Context, *pdf.Renderer, float64, float64, float64, float64) error {
			return sentinel
		}},
	}})
	err := failing.Render(context.Background(), pdf.NewRendererA4Portrait())
	if !errors.Is(err, sentinel) {
		t.Errorf("Draw error not propagated: %v", err)
	}
}

func TestTableCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := testTable(3).Render(ctx, pdf.NewRendererA4Portrait())
	if !errors.Is(err, context.Canceled) {
		t.Errorf("got %v, want context.Canceled", err)
	}
}

func TestTableInDocument(t *testing.T) {
	doc := pdf.NewDocument("Table test",
		pdf.Paragraph("Before the table."),
		testTable(50),
		pdf.Paragraph("After the table."),
	)
	doc.Footer = pdf.ComponentFunc(func(_ context.Context, r *pdf.Renderer) error {
		r.SetY(-15)
		return pdf.Textf("Page %d", r.PageNo()).Render(context.Background(), r)
	})
	pdfBytes, err := doc.Bytes(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(pdfBytes) == 0 {
		t.Error("empty PDF output")
	}
}
