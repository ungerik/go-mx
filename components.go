package mx

import (
	"context"
	"io"
	"net/http"
)

var _ Component = Components{}

type Components []Component

func (Components) RenderOpening(context.Context, io.Writer) error { return nil }

func (cs Components) RenderChildren(ctx context.Context, w io.Writer) error {
	for _, c := range cs {
		err := Render(ctx, w, c)
		if err != nil {
			return err
		}
	}
	return nil
}

func (Components) RenderClosing(context.Context, io.Writer) error { return nil }

func (cs Components) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ServeHTTP(w, r, nil, cs)
}
