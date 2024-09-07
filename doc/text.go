package doc

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/ungerik/go-mx"
)

var _ mx.Component = Text("")

type Text string

func Textf(format string, args ...any) Text {
	return Text(fmt.Sprintf(format, args...))
}

func (text Text) Render(ctx context.Context, w io.Writer) error {
	return RendererFromContext(ctx).RenderText(ctx, w, text)
}

func (text Text) GetChildren(ctx context.Context) ([]mx.Component, error) {
	return nil, nil
}

func (text Text) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte(text))
}
