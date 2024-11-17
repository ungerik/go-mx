package mx

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/domonda/go-errs"
)

type Component interface {
	Render(context.Context, Writer) error
}

type ComponentFunc func(context.Context, Writer) error

func (f ComponentFunc) Render(ctx context.Context, w Writer) error {
	return f(ctx, w)
}

func DefaultAsComponent(c any) Component {
	switch c := c.(type) {
	case nil:
		return nil
	case Component:
		return c
	case string:
		return Text(c)
	case func(context.Context, Writer) error:
		return ComponentFunc(c)
	case func(Writer) error:
		return ComponentFunc(func(_ context.Context, w Writer) error {
			return c(w)
		})
	case func() Component:
		return ComponentFunc(func(ctx context.Context, w Writer) error {
			return c().Render(ctx, w)
		})
	case func() Components:
		return ComponentFunc(func(ctx context.Context, w Writer) error {
			return c().Render(ctx, w)
		})
	case func(context.Context) Component:
		return ComponentFunc(func(ctx context.Context, w Writer) error {
			return c(ctx).Render(ctx, w)
		})
	case func(context.Context) Components:
		return ComponentFunc(func(ctx context.Context, w Writer) error {
			return c(ctx).Render(ctx, w)
		})
	default:
		return Text(fmt.Sprint(c))
	}
}

func ComponentHTTPHandler(comp Component, writerFactory WriterFactory, header http.Header) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		defer func() {
			if p := recover(); p != nil {
				RespondError(response, errs.AsErrorWithDebugStack(p))
			}
		}()
		var buf bytes.Buffer
		writer := writerFactory.NewWriter(&buf)
		err := comp.Render(request.Context(), writer)
		if err != nil {
			RespondError(response, err)
			return
		}
		for key, vals := range header {
			for _, val := range vals {
				response.Header().Add(key, val)
			}
		}
		response.Write(buf.Bytes())
	}
}

type ComponentModifier interface {
	ModifyComponent(Component) Component
}

type ComponentModifierFunc func(Component) Component

func (f ComponentModifierFunc) ModifyComponent(component Component) Component {
	return f(component)
}
