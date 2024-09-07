package mx

import (
	"context"
	"io"
	"net/http"
)

var _ Component = Components{}

type Components []Component

func (Components) RenderOpening(context.Context, io.Writer) error { return nil }

func (cs Components) GetChildren(ctx context.Context) ([]Component, error) {
	return cs, nil
}

func (Components) RenderClosing(context.Context, io.Writer) error { return nil }

func (cs Components) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ServeHTTP(w, r, nil, cs)
}

func ComponentSlice(c Component) []Component {
	if c == nil {
		return nil
	}
	if comps, ok := c.(Components); ok {
		return comps
	}
	return []Component{c}
}
