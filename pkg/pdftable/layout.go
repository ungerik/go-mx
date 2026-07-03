package pdftable

import (
	"context"
	"strings"

	"github.com/domonda/go-errs"

	"github.com/ungerik/go-mx/pdf"
)

// columnDemand is one column's width requirement for resolveColumnWidths:
// exactly one sizing mode applies — fixed if fixed > 0, weighted if
// weight > 0, otherwise auto with the measured content width capped by max.
type columnDemand struct {
	fixed    float64
	weight   float64
	measured float64
	max      float64
}

// resolveColumnWidths distributes tableWidth over the columns: fixed columns
// take their width, auto columns their (capped) measured content width, and
// weighted columns share the remaining space proportionally. When the auto
// columns demand more than the space left by the fixed ones, they are scaled
// down proportionally to exactly fit — unless weighted columns compete for
// the same space, which is reported as an error instead of collapsing them
// to zero width. The resolved widths may sum to less than tableWidth when
// there are no weighted columns.
func resolveColumnWidths(tableWidth float64, cols []columnDemand) ([]float64, error) {
	if tableWidth <= 0 {
		return nil, errs.Errorf("table width %g must be positive", tableWidth)
	}
	widths := make([]float64, len(cols))
	var fixedSum, autoSum, weightSum float64
	for i, col := range cols {
		switch {
		case col.fixed > 0:
			widths[i] = col.fixed
			fixedSum += col.fixed
		case col.weight > 0:
			weightSum += col.weight
		default:
			measured := col.measured
			if col.max > 0 && measured > col.max {
				measured = col.max
			}
			widths[i] = measured
			autoSum += measured
		}
	}
	avail := tableWidth - fixedSum
	if avail < 0 {
		return nil, errs.Errorf("fixed column widths (%g) exceed the table width (%g)", fixedSum, tableWidth)
	}
	if autoSum > avail {
		if weightSum > 0 {
			return nil, errs.Errorf("auto column contents (%g) leave no room for weighted columns in the table width (%g)", autoSum, tableWidth)
		}
		scale := avail / autoSum
		for i, col := range cols {
			if col.fixed <= 0 && col.weight <= 0 {
				widths[i] *= scale
			}
		}
		return widths, nil
	}
	if weightSum > 0 {
		leftover := avail - autoSum
		for i, col := range cols {
			if col.fixed <= 0 && col.weight > 0 {
				widths[i] = leftover * col.weight / weightSum
			}
		}
	}
	return widths, nil
}

// cellLayout is one measured cell, ready to draw: the resolved effective
// style, the translated and wrapped text lines, and the cell height.
type cellLayout struct {
	style  Style
	pad    float64
	lineH  float64
	lines  []string // translated and wrapped; nil for empty or Draw cells
	draw   func(ctx context.Context, r *pdf.Renderer, x, y, w, h float64) error
	drawH  float64
	height float64 // content height + 2*pad
}

// rowLayout is one measured row: its cells (one per column) and the row
// height, the maximum of the cell heights and the row's MinHeight.
type rowLayout struct {
	cells  []cellLayout
	height float64
}

// tableLayout is the fully measured table: the column positions and widths
// and every row wrapped and sized, so the draw pass only places content and
// decides page breaks.
type tableLayout struct {
	x0        float64   // left edge
	colWidths []float64 // resolved column widths
	colX      []float64 // len(colWidths)+1 column boundaries, colX[0] == x0
	width     float64   // sum of colWidths (may be less than the requested width)
	header    *rowLayout
	rows      []rowLayout
}

// layout measures the table against the renderer: it resolves the effective
// style of every cell, measures auto columns with the cell fonts, resolves
// the column widths, and wraps every cell to compute the row heights. It
// mutates the renderer's font state; the caller restores it.
func (t *Table) layout(r *pdf.Renderer, x0, tableWidth float64, bodyBase, headerBase Style) (*tableLayout, error) {
	numCols := len(t.Columns)

	// Resolve the effective style of every cell: cell over row over column
	// over table base.
	colBody := make([]Style, numCols)
	colHeader := make([]Style, numCols)
	for j, col := range t.Columns {
		colBody[j] = overPtr(col.Style, bodyBase)
		colHeader[j] = overPtr(col.HeaderStyle, headerBase)
	}
	var header *rowLayout
	if t.hasHeader() {
		header = &rowLayout{cells: make([]cellLayout, numCols)}
		for j := range numCols {
			header.cells[j] = cellLayout{style: colHeader[j], pad: colHeader[j].pad()}
		}
	}
	rows := make([]rowLayout, len(t.Rows))
	for i, row := range t.Rows {
		rows[i] = rowLayout{cells: make([]cellLayout, numCols)}
		for j := range numCols {
			var cell Cell
			if j < len(row.Cells) {
				cell = row.Cells[j]
			}
			style := overPtr(cell.Style, overPtr(row.Style, colBody[j]))
			rows[i].cells[j] = cellLayout{
				style: style,
				pad:   style.pad(),
				draw:  cell.Draw,
				drawH: cell.Height,
			}
		}
	}

	// Measure auto columns: the widest cell content plus padding.
	demands := make([]columnDemand, numCols)
	for j, col := range t.Columns {
		demands[j] = columnDemand{fixed: col.Width, weight: col.Weight, max: col.MaxWidth}
		if col.Width > 0 || col.Weight > 0 {
			continue
		}
		if header != nil {
			cl := &header.cells[j]
			demands[j].measured = max(demands[j].measured, measureCellWidth(r, cl, t.Columns[j].Title))
		}
		for i := range rows {
			cl := &rows[i].cells[j]
			text := ""
			if j < len(t.Rows[i].Cells) {
				text = t.Rows[i].Cells[j].Text
			}
			demands[j].measured = max(demands[j].measured, measureCellWidth(r, cl, text))
		}
	}

	colWidths, err := resolveColumnWidths(tableWidth, demands)
	if err != nil {
		return nil, err
	}
	lay := &tableLayout{
		x0:        x0,
		colWidths: colWidths,
		colX:      make([]float64, numCols+1),
		header:    header,
		rows:      rows,
	}
	lay.colX[0] = x0
	for j, w := range colWidths {
		lay.width += w
		lay.colX[j+1] = lay.colX[j] + w
	}

	// Wrap every cell at its column width and derive the row heights.
	if header != nil {
		if err := measureRow(r, header, colWidths, headerTexts(t.Columns), 0); err != nil {
			return nil, err
		}
	}
	for i := range rows {
		texts := make([]string, numCols)
		for j := range min(len(t.Rows[i].Cells), numCols) {
			texts[j] = t.Rows[i].Cells[j].Text
		}
		if err := measureRow(r, &rows[i], colWidths, texts, t.Rows[i].MinHeight); err != nil {
			return nil, err
		}
	}
	return lay, r.Error()
}

// headerTexts returns the column titles as a row of cell texts.
func headerTexts(columns []Column) []string {
	texts := make([]string, len(columns))
	for j, col := range columns {
		texts[j] = col.Title
	}
	return texts
}

// measureCellWidth returns the width demand of one cell: its widest unwrapped
// text line (or nothing for Draw cells, whose width is unknowable) plus the
// horizontal padding.
func measureCellWidth(r *pdf.Renderer, cell *cellLayout, text string) float64 {
	width := 2 * cell.pad
	if cell.draw != nil || text == "" {
		return width
	}
	cell.style.applyFont(r)
	var textWidth float64
	for line := range strings.SplitSeq(text, "\n") {
		textWidth = max(textWidth, r.GetStringWidth(r.Str(line)))
	}
	return width + textWidth
}

// measureRow wraps each cell's text at its column width and computes the cell
// and row heights. The wrapped lines are kept (already translated for the
// cell font), so the draw pass can never disagree with the measurement.
func measureRow(r *pdf.Renderer, row *rowLayout, colWidths []float64, texts []string, minHeight float64) error {
	row.height = minHeight
	for j := range row.cells {
		cell := &row.cells[j]
		contentWidth := colWidths[j] - 2*cell.pad
		if contentWidth <= 0 {
			return errs.Errorf("column %d width %g leaves no room for the cell padding %g", j, colWidths[j], cell.pad)
		}
		cell.style.applyFont(r)
		if cell.style.LineHeight > 0 {
			cell.lineH = cell.style.LineHeight
		} else {
			cell.lineH = r.LineHeight()
		}
		var contentHeight float64
		if cell.draw != nil {
			contentHeight = cell.drawH
		} else {
			if texts[j] != "" {
				cell.lines = r.SplitText(r.Str(texts[j]), contentWidth)
			}
			// An empty cell is still one text line tall.
			contentHeight = float64(max(len(cell.lines), 1)) * cell.lineH
		}
		cell.height = contentHeight + 2*cell.pad
		row.height = max(row.height, cell.height)
	}
	return r.Error()
}
