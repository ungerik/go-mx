package html

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/xml"
)

var _ mx.Component = Style{}

type Style struct {
	Media string `attr:"media"`
	Nonce string `attr:"nonce"`
	Title string `attr:"title"`
	Style string
}

func Stylef(textFmt string, args ...any) Style {
	return Style{Style: fmt.Sprintf(textFmt, args...)}
}

func (style Style) Render(ctx context.Context, w io.Writer) error {
	renderer := RendererFromContext(ctx)
	return errors.Join(
		xml.WriteStructAsElementStart(w, renderer, "style", style),
		WriteRaw(w, style.Style),
		renderer.ElementEnd(w, "style"),
	)
}

func (style Style) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mx.ServeComponent(w, r, contentTypeHTML, style)
}
