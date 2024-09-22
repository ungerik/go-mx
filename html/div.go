package html

import (
	"context"
	"io"
	"net/http"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/xml"
)

var _ mx.Component = Div{}

type Div struct {
	ID       string `attr:"id"`
	Class    string `attr:"class"`
	Style    string `attr:"style"`
	Attribs  Attribs
	Children mx.Component
}

func DivText(text string) Div {
	return Div{Children: Text(text)}
}

func (div Div) Render(ctx context.Context, w io.Writer) error {
	renderer := RendererFromContext(ctx)
	err := xml.WriteStructAsElementStart(w, renderer, "div", div)
	if err != nil {
		return err
	}
	if div.Children != nil {
		err = div.Children.Render(ctx, w)
		if err != nil {
			return err
		}
	}
	return renderer.ElementEnd(w, "div")
}

func (div Div) GetChildren(ctx context.Context) ([]mx.Component, error) {
	return mx.ComponentSlice(div.Children), ctx.Err()
}

func (div Div) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mx.ServeComponent(w, r, contentTypeHTML, div)
}
