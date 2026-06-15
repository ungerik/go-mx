package mx

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/domonda/go-errs"
	pretty "github.com/domonda/go-pretty"
)

// Component is the fundamental abstraction of the package: anything that can
// render itself as markup. Render writes the component to w using the context,
// for example to read request-scoped data or to be canceled, and returns the
// first build or write error encountered.
type Component interface {
	Render(context.Context, Writer) error
}

// ComponentFunc adapts a render function to the [Component] interface.
type ComponentFunc func(context.Context, Writer) error

// Render calls the function, satisfying the [Component] interface.
func (f ComponentFunc) Render(ctx context.Context, w Writer) error {
	return f(ctx, w)
}

// DefaultAsComponent converts an arbitrary value into a [Component].
//
// The markup API accepts children as ...any, so this conversion happens
// dynamically at render-build time, not at compile time. The recognized
// types are:
//
//   - nil                        -> nil (renders nothing)
//   - Component                  -> returned unchanged
//   - string                     -> Text (escaped on render)
//   - the func(...) signatures
//     in the switch below         -> wrapped as ComponentFunc
//   - error                      -> Text of error.Error()
//   - fmt.Stringer               -> Text of String()
//
// Any other value falls back to Text(pretty.Sprint(c)) using
// github.com/domonda/go-pretty, giving primitives such as int or bool their
// plain textual form and any other value a compact, single-line
// representation: structs and pointers are tagged with their type name (for
// example "Item{Name:`x`;Count:3}" rather than fmt's anonymous "{x 3}"), while
// slices, maps and named scalars keep their literal form (a slice as [1,2], a
// named string left quoted). go-pretty is preferred over fmt.Sprint here
// because it dereferences pointers, collapses control characters to escapes so
// the text stays on one line, and bounds the length — which makes an unexpected
// value easy to spot. In every case the result is a [Text] node, so
// the value is escaped by the [Writer] when it is rendered and can never inject
// markup into the output — escaping is the Writer's job, independent of the
// target syntax (HTML, XHTML, SVG, XML share the same text-node escaping; see
// [TextEscaper]). Escaping the data content is therefore still mandatory; the
// pretty representation only makes the scaffolding readable, it is not a reason
// to skip escaping.
//
// The flip side is that a value the caller intended as markup but that does not
// implement Component (a struct, a *T whose Component methods are on T, …) is
// rendered as its escaped pretty representation rather than as markup, with no
// compile-time error; convert such a child to a Component explicitly.
//
// DefaultAsComponent is the default implementation of the package-level
// [AsComponent] variable. Assign a different func to [AsComponent] to change
// this behavior for the whole program — for example to panic on unexpected
// types during development, to log them, or to render values with fmt.Sprint or
// a custom go-pretty [pretty.Printer] instead. See the [AsComponent] docs.
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

// ComponentHTTPHandler returns an http.HandlerFunc that renders comp to a
// buffer using a Writer from writerFactory, then writes the buffered markup as
// the response after adding the given response header values. The component is
// rendered with the request's context. A render error or a recovered panic is
// reported via [RespondNonContextError], which returns a 500 (and reveals the
// error only if [RevealInternalServerErrors] is set) while staying silent on a
// canceled request.
func ComponentHTTPHandler(comp Component, writerFactory WriterFactory, header http.Header) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		defer func() {
			if p := recover(); p != nil {
				RespondNonContextError(response, errs.AsErrorWithDebugStack(p))
			}
		}()
		var buf bytes.Buffer
		writer := writerFactory.NewWriter(&buf)
		err := comp.Render(request.Context(), writer)
		if err != nil {
			RespondNonContextError(response, err)
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

// ComponentModifier transforms a [Component] into another one, for example to
// wrap or decorate it.
type ComponentModifier interface {
	ModifyComponent(Component) Component
}

// ComponentModifierFunc adapts a function to the [ComponentModifier] interface.
type ComponentModifierFunc func(Component) Component

// ModifyComponent calls the function, satisfying the [ComponentModifier] interface.
func (f ComponentModifierFunc) ModifyComponent(component Component) Component {
	return f(component)
}
