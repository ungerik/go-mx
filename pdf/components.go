package pdf

import "context"

var _ Component = Components{}

// Components is an ordered list of components rendered one after another, the
// PDF counterpart of mx.Components. A nil element renders nothing.
type Components []Component

func (cs Components) Render(ctx context.Context, r *Renderer) error {
	for _, c := range cs {
		if c == nil {
			continue
		}
		if err := c.Render(ctx, r); err != nil {
			return err
		}
	}
	return nil
}

// AsComponents converts a list of arbitrary values into [Components] using
// [AsComponent], dropping nil results.
func AsComponents(cs ...any) Components {
	components := make(Components, 0, len(cs))
	for _, c := range cs {
		if component := AsComponent(c); component != nil {
			components = append(components, component)
		}
	}
	return components
}
