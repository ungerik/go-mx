package pdf

import (
	"context"
	"fmt"
)

// Component is the fundamental interface of the pdf package, mirroring the
// Component interface of the html package: everything that can draw onto a
// page implements Render. The Renderer replaces the markup writer used for
// HTML, but the shape — Render(context.Context, target) error — is identical,
// so the same composition patterns (Components, If, ForEach, …) apply.
type Component interface {
	Render(ctx context.Context, r *Renderer) error
}

// ComponentFunc adapts a function to the [Component] interface.
type ComponentFunc func(ctx context.Context, r *Renderer) error

func (f ComponentFunc) Render(ctx context.Context, r *Renderer) error {
	return f(ctx, r)
}

// AsComponent converts an arbitrary value into a [Component], the PDF
// counterpart of mx.DefaultAsComponent. Children are accepted as ...any and
// converted at render-build time:
//
//   - nil                                  -> nil (renders nothing)
//   - Component                            -> returned unchanged
//   - string                               -> Text (flowing text)
//   - the func signatures in the switch    -> wrapped as ComponentFunc
//
// Any other value falls back to Text(fmt.Sprint(c)). As in the html package a
// value passed by mistake is silently stringified rather than causing a
// compile error, so convert non-obvious children to a Component explicitly.
func AsComponent(c any) Component {
	switch c := c.(type) {
	case nil:
		return nil
	case Component:
		return c
	case string:
		return Text(c)
	case func(context.Context, *Renderer) error:
		return ComponentFunc(c)
	case func(*Renderer) error:
		return ComponentFunc(func(_ context.Context, r *Renderer) error {
			return c(r)
		})
	case func() Component:
		return ComponentFunc(func(ctx context.Context, r *Renderer) error {
			return c().Render(ctx, r)
		})
	case func(context.Context) Component:
		return ComponentFunc(func(ctx context.Context, r *Renderer) error {
			return c(ctx).Render(ctx, r)
		})
	default:
		return Text(fmt.Sprint(c))
	}
}
