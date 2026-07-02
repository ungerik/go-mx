package pdf

import "context"

// Page starts a new page and renders children onto it. It is the structural
// unit of a document, loosely analogous to an html section: content is grouped
// per page and a new Page begins a fresh one. The very first Page (or the first
// drawing primitive) opens page one, so wrapping a one-page document in Page is
// optional.
func Page(children ...any) Component {
	comps := AsComponents(children...)
	return ComponentFunc(func(ctx context.Context, r *Renderer) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		r.AddPage()
		if err := r.Error(); err != nil {
			return err
		}
		return comps.Render(ctx, r)
	})
}

// PageFormat is [Page] with a per-page orientation and size override, for
// documents that mix, say, portrait and landscape pages.
func PageFormat(orientation Orientation, size PageSize, children ...any) Component {
	comps := AsComponents(children...)
	return ComponentFunc(func(ctx context.Context, r *Renderer) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		r.AddPageFormat(string(orientation), r.GetPageSizeStr(string(size)))
		if err := r.Error(); err != nil {
			return err
		}
		return comps.Render(ctx, r)
	})
}
