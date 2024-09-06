package html

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/ungerik/go-mx"
)

var _ mx.Component = H1{}

type H1 struct {
	ID       string `attr:"id"`
	Class    string `attr:"class"`
	Style    string `attr:"style"`
	Args     Attribs
	Children mx.Component
}

func H1Text(text string) H1 {
	return H1{Children: Text(text)}
}

func H1Textf(textFmt string, args ...any) H1 {
	return H1{Children: Text(fmt.Sprintf(textFmt, args...))}
}

func (h1 H1) RenderOpening(ctx context.Context, w io.Writer) error {
	return WriteStructAsStartTagWithAttribs(ctx, w, "h1", h1)
}

func (h1 H1) RenderChildren(ctx context.Context, w io.Writer) error {
	return mx.Render(ctx, w, h1.Children)
}

func (H1) RenderClosing(ctx context.Context, w io.Writer) error {
	return RendererFromContext(ctx).EndElement(w, "h1")
}

func (h1 H1) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mx.ServeHTTP(w, r, httpHeaderContentTypeHTML, h1)
}
