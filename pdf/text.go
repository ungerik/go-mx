package pdf

import (
	"context"
	"fmt"
)

// Text is flowing text printed at the current cursor with automatic line
// wrapping at the right margin, the PDF counterpart of mx.Text and the type a
// bare string child is converted to. It uses the renderer's default line
// height and advances the cursor; embedded "\n" force line breaks.
type Text string

func (t Text) Render(ctx context.Context, r *Renderer) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.ensurePage()
	r.Write(r.lineHt(0), r.tr(string(t)))
	return r.Error()
}

// Textf is [Text] with fmt.Sprintf formatting, mirroring html.Textf.
func Textf(format string, args ...any) Text {
	return Text(fmt.Sprintf(format, args...))
}

// Cell prints text in a single-line box of the given width and height and moves
// the cursor to its right. A width of 0 extends the cell to the right margin.
// For borders, alignment, fill or a different cursor move use [CellFormat].
func Cell(w, h float64, text string) Component {
	return drawing(func(r *Renderer) {
		r.Cell(w, h, r.tr(text))
	})
}

// CellFormat prints text in a single-line box with full control over the border,
// post-cell cursor move, horizontal and vertical alignment and background fill,
// mirroring fpdf.CellFormat with typed parameters. A width of 0 extends to the
// right margin. fill paints the box with the current fill color before the text.
func CellFormat(w, h float64, text string, border Border, ln LnPos, hAlign HAlign, vAlign VAlign, fill bool) Component {
	return drawing(func(r *Renderer) {
		r.CellFormat(w, h, r.tr(text), string(border), int(ln), string(hAlign)+string(vAlign), fill, 0, "")
	})
}

// MultiCell prints word-wrapped text in a box of the given width, breaking into
// as many lines of height h as needed and advancing the cursor below the block,
// mirroring fpdf.MultiCell. A width of 0 wraps at the right margin. Only the
// horizontal alignment applies; vertical alignment is meaningless for flowing,
// multi-line text.
func MultiCell(w, h float64, text string, border Border, hAlign HAlign, fill bool) Component {
	return drawing(func(r *Renderer) {
		r.MultiCell(w, h, r.tr(text), string(border), string(hAlign), fill)
	})
}

// Paragraph is the common-case shortcut for a block of wrapped, left-aligned
// body text: full content width, automatic line height, no border, no fill.
// It is the PDF analog of an html.P.
func Paragraph(text string) Component {
	return drawing(func(r *Renderer) {
		r.MultiCell(0, r.lineHt(0), r.tr(text), string(BorderNone), string(AlignLeft), false)
	})
}

// TextAt prints text once at the absolute coordinate (x, y) without wrapping or
// advancing the cursor, mirroring fpdf.Text. Useful for labels and annotations
// placed by coordinate rather than by flow.
func TextAt(x, y float64, text string) Component {
	return drawing(func(r *Renderer) {
		r.Text(x, y, r.tr(text))
	})
}

// Ln breaks to a new line, advancing the cursor down by h document units and
// back to the left margin. A value <= 0 uses the height of the last printed
// cell (fpdf's Ln(-1) behavior).
func Ln(h float64) Component {
	return drawing(func(r *Renderer) {
		if h <= 0 {
			r.Ln(-1)
		} else {
			r.Ln(h)
		}
	})
}

// NewLine breaks to a new line using the renderer's default line height.
func NewLine() Component {
	return drawing(func(r *Renderer) {
		r.Ln(r.lineHt(0))
	})
}
