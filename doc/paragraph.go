package doc

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/ungerik/go-mx"
)

type Paragraph struct {
	ID       string
	Children mx.Component
}

func (p *Paragraph) RenderOpening(ctx context.Context, w io.Writer) error {
	return RendererFromContext(ctx).RenderParagraphOpening(ctx, w, p)
}

func (p *Paragraph) RenderChildren(ctx context.Context, w io.Writer) error {
	return RendererFromContext(ctx).RenderParagraphChildren(ctx, w, p)
}

func (p *Paragraph) RenderClosing(ctx context.Context, w io.Writer) error {
	return RendererFromContext(ctx).RenderParagraphClosing(ctx, w, p)
}

func (p *Paragraph) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintf(w, "%#v", p)
}
