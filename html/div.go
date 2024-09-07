package html

import (
	"context"
	"io"
	"net/http"

	"github.com/ungerik/go-mx"
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

func (div Div) RenderOpening(ctx context.Context, w io.Writer) error {
	return WriteStructAsStartTagWithAttribs(ctx, w, "div", div)
}

func (div Div) GetChildren(ctx context.Context) ([]mx.Component, error) {
	return mx.ComponentSlice(div.Children), nil
}

func (Div) RenderClosing(ctx context.Context, w io.Writer) error {
	return RendererFromContext(ctx).EndElement(w, "div")
}

func (div Div) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mx.ServeHTTP(w, r, httpHeaderContentTypeHTML, div)
}
