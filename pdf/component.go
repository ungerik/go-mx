package pdf

import (
	"context"
	"fmt"

	pretty "github.com/domonda/go-pretty"
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

// Render calls the wrapped function.
func (f ComponentFunc) Render(ctx context.Context, r *Renderer) error {
	return f(ctx, r)
}

// DefaultAsComponent converts an arbitrary value into a [Component], the PDF
// counterpart of mx.DefaultAsComponent. Children are accepted as ...any and
// converted at render-build time:
//
//   - nil                                  -> nil (renders nothing)
//   - Component                            -> returned unchanged
//   - string                               -> Text (flowing text)
//   - the func signatures in the switch    -> wrapped as ComponentFunc
//   - error                                -> Text of error.Error()
//   - fmt.Stringer                         -> Text of String()
//
// Any other value falls back to Text(pretty.Sprint(c)) using
// github.com/domonda/go-pretty, mirroring the mx package: primitives get their
// plain textual form and any other value a compact, single-line representation
// where structs and pointers are tagged with their type name (for example
// "Item{Name:`x`;Count:3}" rather than fmt's anonymous "{x 3}") while slices
// and maps keep their literal form, with pointers dereferenced and the length
// bounded, which makes an unexpected value easy to spot. Unlike the markup
// packages there is no
// escaping step — a PDF is not markup, so a stringified value cannot inject
// anything; it is simply drawn as text. The flip side is the same as in mx: a
// value passed by mistake is silently drawn as text rather than causing a
// compile error, so convert non-obvious children to a Component explicitly.
//
// DefaultAsComponent is the default implementation of the package-level
// [AsComponent] variable, which may be reassigned to customize this.
func DefaultAsComponent(c any) Component {
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
	case error:
		// Honor textual intent: error and fmt.Stringer render as their own
		// text. fmt's verbs prefer error.Error over Stringer, so do the same.
		// (go-pretty would not call these and would dump the struct instead,
		// skipping unexported fields, so handle them before the default.)
		return Text(c.Error())
	case fmt.Stringer:
		return Text(c.String())
	default:
		return Text(pretty.Sprint(c))
	}
}
