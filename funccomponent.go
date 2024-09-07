package mx

import (
	"context"
	"io"
	"net/http"
)

var _ Component = FuncComponent(nil)

// FuncComponent is a functional component that
// returns a component to be rendered as children.
//
// Multiple children can be returned as Components slice.
type FuncComponent func(ctx context.Context) (Component, error)

func (FuncComponent) RenderOpening(context.Context, io.Writer) error { return nil }

func (f FuncComponent) GetChildren(ctx context.Context) ([]Component, error) {
	c, err := f(ctx)
	if err != nil {
		return nil, err
	}
	return ComponentSlice(c), nil
}

func (FuncComponent) RenderClosing(context.Context, io.Writer) error { return nil }

func (f FuncComponent) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ServeHTTP(w, r, nil, f)
}
