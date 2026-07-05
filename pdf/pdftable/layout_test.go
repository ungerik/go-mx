package pdftable

import (
	"math"
	"testing"

	"github.com/ungerik/go-mx/pdf"
)

const widthEps = 1e-9

func widthsEqual(got, want []float64) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if math.Abs(got[i]-want[i]) > widthEps {
			return false
		}
	}
	return true
}

func TestResolveColumnWidths(t *testing.T) {
	tests := []struct {
		name       string
		tableWidth float64
		cols       []columnDemand
		want       []float64
		wantErr    bool
	}{
		{
			name:       "fixed only",
			tableWidth: 100,
			cols:       []columnDemand{{fixed: 30}, {fixed: 20}},
			want:       []float64{30, 20},
		},
		{
			name:       "weights share leftover",
			tableWidth: 100,
			cols:       []columnDemand{{fixed: 40}, {weight: 1}, {weight: 2}},
			want:       []float64{40, 20, 40},
		},
		{
			name:       "auto gets measured width",
			tableWidth: 100,
			cols:       []columnDemand{{measured: 25}, {weight: 1}},
			want:       []float64{25, 75},
		},
		{
			name:       "auto capped by max",
			tableWidth: 100,
			cols:       []columnDemand{{measured: 60, max: 30}, {weight: 1}},
			want:       []float64{30, 70},
		},
		{
			name:       "auto columns scaled down to fit",
			tableWidth: 100,
			cols:       []columnDemand{{measured: 150}, {measured: 50}},
			want:       []float64{75, 25},
		},
		{
			name:       "auto overflow with fixed column",
			tableWidth: 100,
			cols:       []columnDemand{{fixed: 20}, {measured: 100}, {measured: 60}},
			want:       []float64{20, 50, 30},
		},
		{
			name:       "auto narrower than table leaves width unused",
			tableWidth: 100,
			cols:       []columnDemand{{measured: 10}, {measured: 20}},
			want:       []float64{10, 20},
		},
		{
			name:       "fixed exceeding table width errors",
			tableWidth: 100,
			cols:       []columnDemand{{fixed: 60}, {fixed: 50}},
			wantErr:    true,
		},
		{
			name:       "auto overflow starves weighted column",
			tableWidth: 100,
			cols:       []columnDemand{{measured: 120}, {weight: 1}},
			wantErr:    true,
		},
		{
			name:       "non-positive table width errors",
			tableWidth: 0,
			cols:       []columnDemand{{weight: 1}},
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveColumnWidths(tt.tableWidth, tt.cols)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got widths %v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !widthsEqual(got, tt.want) {
				t.Errorf("got widths %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStyleOver(t *testing.T) {
	base := Style{
		FontFamily: pdf.Helvetica,
		FontStyle:  new(pdf.StyleRegular),
		FontSize:   12,
		TextColor:  &pdf.Black,
		HAlign:     pdf.AlignLeft,
		VAlign:     pdf.AlignMiddle,
		Padding:    1,
	}
	override := Style{
		FontStyle: new(pdf.StyleBold),
		HAlign:    pdf.AlignRight,
		Padding:   -1, // explicit "no padding"
	}
	got := override.over(base)
	if got.FontFamily != pdf.Helvetica || got.FontSize != 12 {
		t.Errorf("font family/size not inherited: %+v", got)
	}
	if *got.FontStyle != pdf.StyleBold {
		t.Errorf("font style not overridden: %v", *got.FontStyle)
	}
	if got.HAlign != pdf.AlignRight || got.VAlign != pdf.AlignMiddle {
		t.Errorf("alignment merge wrong: %+v", got)
	}
	if got.Padding != -1 || got.pad() != 0 {
		t.Errorf("negative padding must resolve to zero, got Padding %g pad %g", got.Padding, got.pad())
	}
	if got.FillColor != nil {
		t.Errorf("fill color must stay nil, got %v", got.FillColor)
	}
	if overPtr(nil, base) != base {
		t.Error("overPtr(nil) must return the base unchanged")
	}
}

func TestMeasureRowHeights(t *testing.T) {
	r := pdf.NewRendererA4Portrait()
	r.AddPage()
	r.SetCellMargin(0)

	style := Style{
		FontFamily: pdf.Helvetica,
		FontStyle:  new(pdf.StyleRegular),
		FontSize:   12,
		TextColor:  &pdf.Black,
		Padding:    1,
		LineHeight: 5,
	}
	row := rowLayout{cells: []cellLayout{
		{style: style, pad: 1},
		{style: style, pad: 1},
	}}
	// Column 0 is wide enough for one line, column 1 forces a wrap.
	long := "some words that will need to wrap"
	err := measureRow(r, &row, []float64{100, 20}, []string{"short", long}, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(row.cells[0].lines) != 1 {
		t.Errorf("cell 0: got %d lines, want 1", len(row.cells[0].lines))
	}
	if len(row.cells[1].lines) < 2 {
		t.Errorf("cell 1: got %d lines, want a wrap", len(row.cells[1].lines))
	}
	wantHeight := float64(len(row.cells[1].lines))*5 + 2
	if math.Abs(row.height-wantHeight) > widthEps {
		t.Errorf("row height %g, want %g", row.height, wantHeight)
	}

	// An empty row is still one line tall, and MinHeight wins over content.
	empty := rowLayout{cells: []cellLayout{{style: style, pad: 1}}}
	if err := measureRow(r, &empty, []float64{50}, []string{""}, 0); err != nil {
		t.Fatal(err)
	}
	if math.Abs(empty.height-7) > widthEps { // 1 line * 5 + 2*1 padding
		t.Errorf("empty row height %g, want 7", empty.height)
	}
	tall := rowLayout{cells: []cellLayout{{style: style, pad: 1}}}
	if err := measureRow(r, &tall, []float64{50}, []string{"x"}, 30); err != nil {
		t.Fatal(err)
	}
	if tall.height != 30 {
		t.Errorf("MinHeight row height %g, want 30", tall.height)
	}

	// A column narrower than its padding is a deferred error.
	narrow := rowLayout{cells: []cellLayout{{style: style, pad: 3}}}
	if err := measureRow(r, &narrow, []float64{5}, []string{"x"}, 0); err == nil {
		t.Error("expected error for column narrower than its padding")
	}
}
