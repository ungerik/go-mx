package html

import (
	"context"
	"io"
	"net/http"

	"github.com/ungerik/go-mx"
)

var _ mx.Component = Element{}

type Element struct {
	Name     string
	Attribs  Attribs
	Children mx.Component
}

func (e Element) RenderOpening(ctx context.Context, w io.Writer) error {
	renderer := RendererFromContext(ctx)
	err := renderer.OpenElement(w, e.Name)
	if err != nil {
		return err
	}
	for name, value := range e.Attribs.Iter() {
		err = renderer.Attribute(w, name, value)
		if err != nil {
			return err
		}
	}
	if e.Children == nil {
		return renderer.CloseVoidElement(w)
	}
	return renderer.CloseElement(w)
}

func (e Element) RenderChildren(ctx context.Context, w io.Writer) error {
	return mx.Render(ctx, w, e.Children)
}

func (e Element) RenderClosing(ctx context.Context, w io.Writer) error {
	if e.Children == nil {
		return nil
	}
	return RendererFromContext(ctx).EndElement(w, e.Name)
}

func (e Element) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mx.ServeHTTP(w, r, httpHeaderContentTypeHTML, e)
}
