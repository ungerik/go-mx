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

func DefaultAsComponent(obj any) Component {
	switch x := obj.(type) {
	case nil:
		return nil
	case ComponentFunc:
		return x
	case Component:
		return x
	case string:
		return Text(x)
	case func() Component:
		return x()
	case func() Components:
		return x()
	case func(context.Context, Writer) error:
		return ComponentFunc(x)
	case func(Writer) error:
		return ComponentFunc(func(_ context.Context, w Writer) error {
			return x(w)
		})
	default:
		return Text(fmt.Sprint(x))
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
