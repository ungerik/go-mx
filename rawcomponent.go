package mx

import (
	"context"
	"io"
	"net/http"
)

var _ Component = RawComponent("")

type RawComponent string

func (raw RawComponent) Render(ctx context.Context, w io.Writer) error {
	_, err := w.Write([]byte(raw))
	return err
}

func (raw RawComponent) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(raw))
}
