package mx

import (
	"context"
)

var _ Component = Components{}

// Components is a slice of [Component] that is itself a [Component],
// rendering its elements in order. It is the type used for an element's
// children. Nil elements are skipped.
type Components []Component

// Render renders each non-nil component in order, returning the first error.
func (cs Components) Render(ctx context.Context, w Writer) error {
	for _, c := range cs {
		if c != nil {
			err := c.Render(ctx, w)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// AsComponents converts the variadic arguments into a [Components] slice using
// [AsComponent] for each value, dropping any that convert to nil (such as a nil
// argument).
func AsComponents(cs ...any) Components {
	components := make(Components, 0, len(cs))
	for _, c := range cs {
		component := AsComponent(c)
		if component != nil {
			components = append(components, component)
		}
	}
	return components
}

// func ComponentsFromComponent(c Component) Components {
// 	switch x := c.(type) {
// 	case nil:
// 		return nil
// 	case Components:
// 		return x
// 	default:
// 		return Components{x}
// 	}
// }

// ComponentsModifier transforms a [Components] slice into another one, for
// example to filter, reorder or wrap children.
type ComponentsModifier interface {
	ModifyComponents(Components) Components
}

// ComponentsModifierFunc adapts a function to the [ComponentsModifier] interface.
type ComponentsModifierFunc func(Components) Components

// ModifyComponents calls the function, satisfying the [ComponentsModifier] interface.
func (f ComponentsModifierFunc) ModifyComponents(components Components) Components {
	return f(components)
}

// ModifyOnRender returns a [Component] that, when rendered, converts cs to
// [Components] (via [AsComponents]), passes them through modify together with
// the render context, and renders the result. The modification is deferred to
// render time so it can use request-scoped context data.
func ModifyOnRender(modify func(context.Context, Components) Components, cs ...any) Component {
	return ComponentFunc(func(ctx context.Context, w Writer) error {
		return modify(ctx, AsComponents(cs...)).Render(ctx, w)
	})
}
