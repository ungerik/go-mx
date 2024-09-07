package mx

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/domonda/go-errs"
)

type Component interface {
	http.Handler

	Render(context.Context, io.Writer) error
}

type ComponentWithChildren interface {
	Component

	GetChildren(ctx context.Context) ([]Component, error)
}

func ServeComponent(w http.ResponseWriter, r *http.Request, h http.Header, c Component) {
	defer func() {
		if r := recover(); r != nil {
			RespondError(w, errs.AsErrorWithDebugStack(r))
		}
	}()
	var buf bytes.Buffer
	err := c.Render(r.Context(), &buf)
	if err != nil {
		RespondError(w, err)
		return
	}
	for key, vals := range h {
		for _, val := range vals {
			w.Header().Add(key, val)
		}
	}
	w.Write(buf.Bytes())
}
