package mx

import (
	"context"
)

var _ Component = Components{}

type Components []Component

func (cs Components) Render(ctx context.Context, w Writer) error {
	for _, c := range cs {
		if c == nil {
			continue
		}
		err := c.Render(ctx, w)
		if err != nil {
			return err
		}
	}
	return nil
}

func AsComponents(obj ...any) Components {
	comps := make(Components, 0, len(obj))
	for _, o := range obj {
		comp := AsComponent(o)
		if comp != nil {
			comps = append(comps, comp)
		}
	}
	return comps
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
