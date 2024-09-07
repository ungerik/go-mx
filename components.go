package mx

import (
	"context"
	"io"
	"net/http"
)

var _ Component = Components{}

type Components []Component

func (cs Components) Render(ctx context.Context, w io.Writer) error {
	for _, c := range cs {
		err := c.Render(ctx, w)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cs Components) GetChildren(ctx context.Context) ([]Component, error) {
	return cs, nil
}

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
