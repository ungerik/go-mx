package doc

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ungerik/go-mx"
)

var _ mx.Component = Text("")

// Text is a string of document text that renders itself through the
// Renderer found in the context and can also be served directly over HTTP.
type Text string

// Textf returns a Text formatted according to a format specifier,
// using fmt.Sprintf semantics.
func Textf(format string, args ...any) Text {
	return Text(fmt.Sprintf(format, args...))
}

// Render writes the text using the Renderer from the context's RenderText method.
func (text Text) Render(ctx context.Context, w mx.Writer) error {
	return RendererFromContext(ctx).RenderText(ctx, w, text)
}

func (text Text) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte(text))
}
