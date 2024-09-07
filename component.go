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

	RenderOpening(context.Context, io.Writer) error
	GetChildren(context.Context) ([]Component, error)
	RenderClosing(context.Context, io.Writer) error
}

func Render(ctx context.Context, w io.Writer, c Component) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if c == nil {
		return nil
	}
	err := c.RenderOpening(ctx, w)
	if err != nil {
		return err
	}
	children, err := c.GetChildren(ctx)
	if err != nil {
		return err
	}
	for _, child := range children {
		err = Render(ctx, w, child)
		if err != nil {
			return err
		}
	}
	return c.RenderClosing(ctx, w)
}

func ServeHTTP(w http.ResponseWriter, r *http.Request, h http.Header, c Component) {
	defer func() {
		if r := recover(); r != nil {
			RespondError(w, errs.AsErrorWithDebugStack(r))
		}
	}()
	var buf bytes.Buffer
	err := Render(r.Context(), &buf, c)
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
