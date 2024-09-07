package html

import (
	"context"
	"io"
	"net/http"

	"github.com/ungerik/go-mx"
)

const (
	N  Raw = "\n"
	BR Raw = "<br/>"
)

var _ mx.Component = Raw("")

type Raw string

func (raw Raw) RenderOpening(_ context.Context, w io.Writer) error {
	_, err := w.Write([]byte(raw))
	return err
}

func (Raw) GetChildren(ctx context.Context) ([]mx.Component, error) { return nil, nil }
func (Raw) RenderClosing(context.Context, io.Writer) error          { return nil }

func (raw Raw) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mx.ServeHTTP(w, r, httpHeaderContentTypeHTML, raw)
}
