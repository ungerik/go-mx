package shadcn

import (
	"context"

	"github.com/ungerik/go-mx"
)

// finish completes a shadcn component element. It merges the component's base
// classes with any caller-supplied class attributes via [Cn], guarantees a
// single data-slot attribute, and rebuilds the element's attributes in a
// stable order.
//
// The go-mx CheckedWriter rejects a duplicate attribute name on one element,
// so a component that carries base classes and also accepts a caller-supplied
// class must merge them into one class attribute itself. finish is that merge
// step, shared by every component in this package.
//
// Caller attributes are deduplicated by name (last value wins) so an
// accidental duplicate never causes a render error. A caller-supplied
// data-slot is dropped: the slot identifies the component and is not
// caller-configurable.
func finish(e *mx.Element, slot, baseClasses string) *mx.Element {
	var callerClasses []string
	other := make([]mx.Attrib, 0, len(e.Attribs))
	index := make(map[string]int, len(e.Attribs))

	for _, a := range e.Attribs {
		switch name := a.AttribName(); name {
		case "class":
			callerClasses = append(callerClasses, a.AttribValue(context.Background()))
		case "data-slot":
			// Component identity, not caller-configurable.
		default:
			if i, ok := index[name]; ok {
				other[i] = a // last occurrence wins
			} else {
				index[name] = len(other)
				other = append(other, a)
			}
		}
	}

	merged := Cn(baseClasses, callerClasses)

	attribs := make([]mx.Attrib, 0, len(other)+2)
	attribs = append(attribs, mx.NewAttrib("data-slot", slot))
	attribs = append(attribs, other...)
	if merged != "" {
		attribs = append(attribs, mx.NewAttrib("class", merged))
	}
	e.Attribs = attribs
	return e
}
