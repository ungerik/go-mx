package pdftable

import (
	"context"
	"math"

	"github.com/domonda/go-errs"

	"github.com/ungerik/go-mx/pdf"
)

// fitEps absorbs float summation error when checking whether a row fits the
// remaining page height, so a row measured to exactly fill the space does not
// spuriously break to a new page.
const fitEps = 1e-4

// Render measures the table and draws it at the current cursor position,
// starting at the cursor x and extending to [Table.Width] or the right
// margin. Rows never break internally: a row that does not fit the remaining
// page height moves to a new page (only a row taller than a whole page is
// split by text lines). Page breaks run the document's Footer and Header as
// usual, and the table's header row is redrawn when RepeatHeader is set.
//
// fpdf's automatic page break is suspended while the table draws and restored
// afterwards, as are the font, the text and fill colors and the cell margin.
// Grid rules use the current draw color and line width. Afterwards the cursor
// sits below the table at the left margin, like after a MultiCell.
func (t *Table) Render(ctx context.Context, r *pdf.Renderer) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(t.Columns) == 0 {
		return errs.New("pdftable: table has no columns")
	}
	for i := range t.Rows {
		if len(t.Rows[i].Cells) > len(t.Columns) {
			return errs.Errorf("pdftable: row %d has %d cells but the table has %d columns", i, len(t.Rows[i].Cells), len(t.Columns))
		}
	}
	if r.PageNo() == 0 {
		r.AddPage()
	}
	if err := r.Error(); err != nil {
		return err
	}

	// Capture the renderer state the table mutates and restore it at the
	// end, like pdf.Save does — via the getters, so it works regardless of
	// how the state was set.
	family := r.GetFontFamily()
	fontStyle := r.GetFontStyle()
	fontSizePt, _ := r.GetFontSize()
	textR, textG, textB := r.GetTextColor()
	fillR, fillG, fillB := r.GetFillColor()
	cellMargin := r.GetCellMargin()
	autoBreak, breakMargin := r.GetAutoPageBreak()
	defer func() {
		r.SetFont(family, fontStyle, fontSizePt)
		r.SetTextColor(textR, textG, textB)
		r.SetFillColor(fillR, fillG, fillB)
		r.SetCellMargin(cellMargin)
		r.SetAutoPageBreak(autoBreak, breakMargin)
	}()
	// Padding is the table's own concern: with the cell margin zeroed,
	// SplitText wraps at exactly the content width and CellFormat places
	// text at the exact x. The table breaks pages itself, between rows.
	r.SetCellMargin(0)
	r.SetAutoPageBreak(false, breakMargin)

	// The style cascade roots at the renderer's current state; the header
	// base defaults to the body base in bold.
	base := Style{
		FontFamily: family,
		FontStyle:  new(pdf.FontStyle(fontStyle)),
		FontSize:   fontSizePt,
		TextColor:  &pdf.Color{R: textR, G: textG, B: textB},
		HAlign:     pdf.AlignLeft,
		VAlign:     pdf.AlignMiddle,
		Padding:    cellMargin,
	}
	bodyBase := t.Style.over(base)
	headerBase := t.HeaderStyle.over(Style{FontStyle: new(pdf.StyleBold)}.over(bodyBase))

	x0 := r.GetX()
	pageWidth, pageHeight := r.GetPageSize()
	_, topMargin, rightMargin, bottomMargin := r.GetMargins()
	breakY := pageHeight - bottomMargin
	tableWidth := t.Width
	if tableWidth <= 0 {
		tableWidth = pageWidth - rightMargin - x0
	}

	lay, err := t.layout(r, x0, tableWidth, bodyBase, headerBase)
	if err != nil {
		return err
	}

	y := r.GetY()
	segTop := y // top of the table's segment on the current page

	// closeSegment strokes the vertical rules and the outer outline for
	// the part of the table drawn on the current page.
	closeSegment := func(yEnd float64) {
		if yEnd <= segTop {
			return
		}
		if t.Grid.Has(GridCols) {
			for _, bx := range lay.colX[1 : len(lay.colX)-1] {
				r.Line(bx, segTop, bx, yEnd)
			}
		}
		if t.Grid.Has(GridOuter) {
			right := x0 + lay.width
			r.Line(x0, segTop, x0, yEnd)
			r.Line(right, segTop, right, yEnd)
			r.Line(x0, segTop, right, segTop)
			r.Line(x0, yEnd, right, yEnd)
		}
	}

	drawHeaderRow := func() error {
		if err := drawRowAt(ctx, r, lay, lay.header, y); err != nil {
			return err
		}
		y += lay.header.height
		if t.Grid.Has(GridHeader) || t.Grid.Has(GridRows) {
			r.Line(x0, y, x0+lay.width, y)
		}
		return r.Error()
	}

	// newPage closes the current segment, breaks to a new page (running
	// the document's footer and header callbacks) and redraws the table
	// header when configured.
	newPage := func(withHeader bool) error {
		closeSegment(y)
		r.AddPage()
		if err := r.Error(); err != nil {
			return err
		}
		y = r.GetY()
		segTop = y
		if withHeader && t.RepeatHeader && lay.header != nil {
			return drawHeaderRow()
		}
		return nil
	}

	if lay.header != nil {
		// Keep the header with the first body row: never leave an
		// orphaned header at the bottom of a page.
		needed := lay.header.height
		if len(lay.rows) > 0 {
			needed += lay.rows[0].height
		}
		if y+needed > breakY+fitEps {
			if err := newPage(false); err != nil {
				return err
			}
		}
		if err := drawHeaderRow(); err != nil {
			return err
		}
	}

	for i := range lay.rows {
		if err := ctx.Err(); err != nil {
			return err
		}
		row := &lay.rows[i]
		if y+row.height > breakY+fitEps && row.height <= breakY-topMargin+fitEps {
			// The row does not fit here but fits a fresh page whole.
			if err := newPage(true); err != nil {
				return err
			}
		}
		if y+row.height > breakY+fitEps {
			// Taller than a whole page: split by text lines.
			if err := drawSplitRow(ctx, r, lay, row, &y, breakY, newPage); err != nil {
				return err
			}
		} else {
			if err := drawRowAt(ctx, r, lay, row, y); err != nil {
				return err
			}
			y += row.height
		}
		if t.Grid.Has(GridRows) && i < len(lay.rows)-1 {
			r.Line(x0, y, x0+lay.width, y)
		}
	}

	closeSegment(y)
	r.SetY(y) // below the table, x back to the left margin like MultiCell
	return r.Error()
}

// drawRowAt draws one measured row at the vertical position y: per cell the
// background fill, then the custom Draw content or the wrapped text lines
// with their alignment.
func drawRowAt(ctx context.Context, r *pdf.Renderer, lay *tableLayout, row *rowLayout, y float64) error {
	for j := range row.cells {
		cell := &row.cells[j]
		x := lay.colX[j]
		w := lay.colWidths[j]
		if cell.style.FillColor != nil {
			fill := *cell.style.FillColor
			r.SetFillColor(fill.R, fill.G, fill.B)
			r.Rect(x, y, w, row.height, "F")
		}
		if cell.draw != nil {
			err := cell.draw(ctx, r, x+cell.pad, y+cell.pad, w-2*cell.pad, row.height-2*cell.pad)
			if err != nil {
				return err
			}
			continue
		}
		if err := drawCellLines(r, cell, x, y, w, row.height, cell.lines); err != nil {
			return err
		}
	}
	return r.Error()
}

// drawCellLines draws the given wrapped text lines into the cell box
// (x, y, w, h) honoring the cell's alignment.
func drawCellLines(r *pdf.Renderer, cell *cellLayout, x, y, w, h float64, lines []string) error {
	if len(lines) == 0 {
		return nil
	}
	cell.style.applyFont(r)
	text := *cell.style.TextColor
	r.SetTextColor(text.R, text.G, text.B)
	hAlign := cell.style.HAlign
	if hAlign == pdf.AlignJustify {
		hAlign = pdf.AlignLeft
	}
	textHeight := float64(len(lines)) * cell.lineH
	var top float64
	switch cell.style.VAlign {
	case pdf.AlignTop, pdf.AlignBaseline:
		top = y + cell.pad
	case pdf.AlignBottom:
		top = max(y+h-cell.pad-textHeight, y+cell.pad)
	default: // middle
		top = max(y+(h-textHeight)/2, y+cell.pad)
	}
	for k, line := range lines {
		if line == "" {
			continue
		}
		r.SetXY(x+cell.pad, top+float64(k)*cell.lineH)
		r.CellFormat(w-2*cell.pad, cell.lineH, line, "", 0, string(hAlign), false, 0, "")
	}
	return r.Error()
}

// drawSplitRow draws a row taller than a whole page as a sequence of
// top-aligned fragments, breaking between text lines. Fragment fills and the
// row's vertical rules are drawn per page segment; a Draw cell receives its
// box on the first fragment. y is advanced to below the last fragment.
func drawSplitRow(ctx context.Context, r *pdf.Renderer, lay *tableLayout, row *rowLayout, y *float64, breakY float64, newPage func(withHeader bool) error) error {
	remaining := make([][]string, len(row.cells))
	for j := range row.cells {
		remaining[j] = row.cells[j].lines
	}
	first := true
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		avail := breakY - *y
		// How many of each cell's remaining lines fit the space.
		counts := make([]int, len(row.cells))
		fragHeight := 0.0
		drawnTotal := 0
		remainingTotal := 0
		for j := range row.cells {
			cell := &row.cells[j]
			n := int(math.Floor((avail-2*cell.pad)/cell.lineH + fitEps))
			n = max(min(n, len(remaining[j])), 0)
			counts[j] = n
			drawnTotal += n
			remainingTotal += len(remaining[j])
			if n > 0 {
				fragHeight = max(fragHeight, float64(n)*cell.lineH+2*cell.pad)
			}
			if first && cell.draw != nil {
				fragHeight = max(fragHeight, min(cell.height, avail))
			}
		}
		if drawnTotal == 0 && remainingTotal > 0 && !first {
			return errs.Errorf("pdftable: a cell line height is taller than the page content height %g", avail)
		}
		last := drawnTotal == remainingTotal
		if last {
			// The final fragment keeps the row's measured proportions
			// (padding below the last lines, MinHeight leftovers).
			fragHeight = 0
			for j := range row.cells {
				cell := &row.cells[j]
				height := float64(len(remaining[j]))*cell.lineH + 2*cell.pad
				if first && cell.draw != nil {
					height = max(height, cell.height)
				}
				fragHeight = max(fragHeight, height)
			}
			fragHeight = min(max(fragHeight, 0), avail)
		}
		// Draw the fragment: fills, then content, top-aligned.
		for j := range row.cells {
			cell := &row.cells[j]
			x := lay.colX[j]
			w := lay.colWidths[j]
			if cell.style.FillColor != nil {
				fill := *cell.style.FillColor
				r.SetFillColor(fill.R, fill.G, fill.B)
				r.Rect(x, *y, w, fragHeight, "F")
			}
			if cell.draw != nil {
				if first {
					err := cell.draw(ctx, r, x+cell.pad, *y+cell.pad, w-2*cell.pad, cell.height-2*cell.pad)
					if err != nil {
						return err
					}
				}
				continue
			}
			if counts[j] > 0 {
				topAligned := *cell
				topAligned.style.VAlign = pdf.AlignTop
				if err := drawCellLines(r, &topAligned, x, *y, w, fragHeight, remaining[j][:counts[j]]); err != nil {
					return err
				}
				remaining[j] = remaining[j][counts[j]:]
			}
		}
		*y += fragHeight
		if err := r.Error(); err != nil {
			return err
		}
		if last {
			return nil
		}
		if err := newPage(true); err != nil {
			return err
		}
		first = false
	}
}
