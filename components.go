package mx

import (
	"context"
)

var _ Component = Components{}

type Components []Component

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

type ComponentsModifier interface {
	ModifyComponents(Components) Components
}

type ComponentsModifierFunc func(Components) Components

func (f ComponentsModifierFunc) ModifyComponents(components Components) Components {
	return f(components)
}

func ModifyOnRender(modify func(context.Context, Components) Components, cs ...any) Component {
	return ComponentFunc(func(ctx context.Context, w Writer) error {
		return modify(ctx, AsComponents(cs...)).Render(ctx, w)
	})
}
