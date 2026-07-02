package pdf

import "context"

// Font selects the font family, style and size for subsequent text. It is the
// stateful PDF analog of setting CSS font properties: the selection persists
// until changed. Wrap it in [Save] to scope it to a group of children.
func Font(family string, style FontStyle, size float64) Component {
	return op(func(r *Renderer) {
		r.SetFont(family, string(style), size)
	})
}

// FontSize changes only the size of the current font, in points.
func FontSize(size float64) Component {
	return op(func(r *Renderer) {
		r.SetFontSize(size)
	})
}

// TextColor sets the fill color used to paint text.
func TextColor(c Color) Component {
	return op(func(r *Renderer) {
		r.SetTextColor(c.R, c.G, c.B)
	})
}

// FillColor sets the color used to fill shapes and cell backgrounds.
func FillColor(c Color) Component {
	return op(func(r *Renderer) {
		r.SetFillColor(c.R, c.G, c.B)
	})
}

// DrawColor sets the color used to stroke lines and shape outlines.
func DrawColor(c Color) Component {
	return op(func(r *Renderer) {
		r.SetDrawColor(c.R, c.G, c.B)
	})
}

// LineWidth sets the stroke width in document units.
func LineWidth(w float64) Component {
	return op(func(r *Renderer) {
		r.SetLineWidth(w)
	})
}

// LineCap sets the shape drawn at the ends of open stroked paths.
func LineCap(style LineCapStyle) Component {
	return op(func(r *Renderer) {
		r.SetLineCapStyle(string(style))
	})
}

// LineJoin sets the shape drawn where stroked path segments meet.
func LineJoin(style LineJoinStyle) Component {
	return op(func(r *Renderer) {
		r.SetLineJoinStyle(string(style))
	})
}

// X moves the cursor to the absolute horizontal position x, keeping y.
func X(x float64) Component {
	return op(func(r *Renderer) { r.SetX(x) })
}

// Y moves the cursor to the absolute vertical position y and resets x to the
// left margin (fpdf.SetY behavior).
func Y(y float64) Component {
	return op(func(r *Renderer) { r.SetY(y) })
}

// XY moves the cursor to the absolute position (x, y).
func XY(x, y float64) Component {
	return op(func(r *Renderer) { r.SetXY(x, y) })
}

// MoveDown moves the cursor down by dy document units, keeping x.
func MoveDown(dy float64) Component {
	return op(func(r *Renderer) { r.SetY(r.GetY() + dy) })
}

// MoveRight moves the cursor right by dx document units, keeping y.
func MoveRight(dx float64) Component {
	return op(func(r *Renderer) { r.SetX(r.GetX() + dx) })
}

// Save renders children with the current graphics state restored afterwards,
// the PDF analog of wrapping content in an element so style changes inside do
// not leak out. It captures and restores the font (family, style and size),
// the text, fill and draw colors, the line width, the line cap and join styles,
// and the cursor position — using fpdf's getters, so it works regardless of
// whether the state was set through this package or the raw embedded renderer.
//
// The dash pattern and the alpha/blend mode are not restored: fpdf exposes no
// getter for the dash pattern, and its zero alpha value is indistinguishable
// from a deliberate fully-transparent setting. Reset those explicitly if you
// change them inside a Save.
//
// The cursor restore assumes children stay on the same page: if they trigger an
// automatic page break or add a page, the restored x, y lands on the new page,
// where later content can overprint. Save is meant for scoping style, not page
// flow. State setters and drawing primitives may be freely mixed as children,
// e.g. Save(TextColor(Red), Text("warning")).
func Save(children ...any) Component {
	comps := AsComponents(children...)
	return ComponentFunc(func(ctx context.Context, r *Renderer) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		// Capture restorable state via fpdf's getters.
		family := r.GetFontFamily()
		style := r.GetFontStyle()
		sizePt, _ := r.GetFontSize() // SetFont takes points, the first return

		tr, tg, tb := r.GetTextColor()
		fr, fg, fb := r.GetFillColor()
		dr, dg, db := r.GetDrawColor()
		lineWidth := r.GetLineWidth()
		capStyle := r.GetLineCapStyle()
		joinStyle := r.GetLineJoinStyle()
		x, y := r.GetXY()

		err := comps.Render(ctx, r)

		// Restore, even on error, so the rest of the document is unaffected.
		r.SetFont(family, style, sizePt)
		r.SetTextColor(tr, tg, tb)
		r.SetFillColor(fr, fg, fb)
		r.SetDrawColor(dr, dg, db)
		r.SetLineWidth(lineWidth)
		r.SetLineCapStyle(capStyle)
		r.SetLineJoinStyle(joinStyle)
		r.SetXY(x, y)
		return err
	})
}
