package html

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

func (text Text) RenderOpening(_ context.Context, w io.Writer) error {
	return WriteEscaped(w, text)
}

func (Text) GetChildren(context.Context) ([]mx.Component, error) { return nil, nil }
func (Text) RenderClosing(context.Context, io.Writer) error      { return nil }

func (text Text) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mx.ServeHTTP(w, r, httpHeaderContentTypeHTML, text)
}
