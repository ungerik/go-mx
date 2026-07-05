package fpdf

import "context"

// drawing wraps fn as a [Component] that draws page content: it honors context
// cancellation, makes sure a page exists (so the first primitive does not need
// an explicit Page), runs fn, and returns any error accumulated by fpdf.
func drawing(fn func(r *Renderer)) Component {
	return ComponentFunc(func(ctx context.Context, r *Renderer) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		r.ensurePage()
		fn(r)
		return r.Error()
	})
}

// op wraps fn as a [Component] that changes renderer state (font, color, cursor,
// metadata) without drawing, so it does not force a page to start. It honors
// context cancellation and returns any error accumulated by fpdf.
func op(fn func(r *Renderer)) Component {
	return ComponentFunc(func(ctx context.Context, r *Renderer) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		fn(r)
		return r.Error()
	})
}
