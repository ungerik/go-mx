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

func (e Element) Render(ctx context.Context, w io.Writer) error {
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
	err = renderer.CloseElement(w)
	if err != nil {
		return err
	}
	err = e.Children.Render(ctx, w)
	if err != nil {
		return err
	}
	return renderer.EndElement(w, e.Name)
}

func (e Element) GetChildren(ctx context.Context) ([]mx.Component, error) {
	return mx.ComponentSlice(e.Children), ctx.Err()
}

func (e Element) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mx.ServeComponent(w, r, contentTypeHTML, e)
}
