package doc

import "github.com/ungerik/go-mx"

// Paragraph is a block of document content identified by ID and holding its
// content as a single child Component.
type Paragraph struct {
	ID       string
	Children mx.Component
}

// func (p *Paragraph) RenderOpening(ctx context.Context, w io.Writer) error {
// 	return RendererFromContext(ctx).RenderParagraphOpening(ctx, w, p)
// }

// func (p *Paragraph) GetChildren(ctx context.Context) ([]mx.Component, error) {
// 	return mx.ComponentSlice(p.Children), ctx.Err()
// }

// func (p *Paragraph) RenderClosing(ctx context.Context, w io.Writer) error {
// 	return RendererFromContext(ctx).RenderParagraphClosing(ctx, w, p)
// }

// func (p *Paragraph) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
// 	fmt.Fprintf(w, "%#v", p)
// }
