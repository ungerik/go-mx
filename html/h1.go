package html

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/xml"
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

func (h1 H1) Render(ctx context.Context, w io.Writer) error {
	renderer := RendererFromContext(ctx)
	err := xml.WriteStructAsElementStart(w, renderer, "h1", h1)
	if err != nil {
		return err
	}
	if h1.Children != nil {
		err = h1.Children.Render(ctx, w)
		if err != nil {
			return err
		}
	}
	return renderer.ElementEnd(w, "h1")
}

func (h1 H1) GetChildren(ctx context.Context) ([]mx.Component, error) {
	return mx.ComponentSlice(h1.Children), ctx.Err()
}

func (h1 H1) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mx.ServeComponent(w, r, contentTypeHTML, h1)
}
