package mx

import (
	"context"
	"io"
	"net/http"
)

var _ Component = RawComponent{}

type RawComponent struct {
	Opening  string
	Children []Component
	Closing  string
}

func (raw RawComponent) RenderOpening(_ context.Context, w io.Writer) error {
	_, err := w.Write([]byte(raw.Opening))
	return err
}

func (raw RawComponent) GetChildren(context.Context) ([]Component, error) {
	return raw.Children, nil
}

func (raw RawComponent) RenderClosing(_ context.Context, w io.Writer) error {
	_, err := w.Write([]byte(raw.Closing))
	return err
}

func (raw RawComponent) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ServeHTTP(w, r, nil, raw)
}
