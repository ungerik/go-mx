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

func (raw Raw) Render(_ context.Context, w io.Writer) error {
	_, err := w.Write([]byte(raw))
	return err
}

func (Raw) GetChildren(ctx context.Context) ([]mx.Component, error) { return nil, nil }

func (raw Raw) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mx.ServeComponent(w, r, contentTypeHTML, raw)
}
