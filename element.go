package mx

import (
	"context"
	"strings"
)

type Element struct {
	Name     string
	Attribs  []Attrib
	Children Components // nil for void element, empty slice for no children
	Err      error      // if non-nil, Render returns it instead of rendering; see NewErrElement
}

func NewElement(name string, attribsChildren ...any) *Element {
	e := &Element{Name: name, Children: []Component{}}
	for _, ac := range attribsChildren {
		if ac == nil {
			continue
		}
		if attribs, ok := AsAttribs(ac); ok {
			e.Attribs = append(e.Attribs, attribs...)
			continue
		}
		e.Children = append(e.Children, AsComponent(ac))
	}
	return e
}

func NewVoidElement(name string, attribs ...Attrib) *Element {
	return &Element{Name: name, Attribs: attribs}
}

// NewErrElement returns an Element that renders nothing and whose Render method
// returns err. It defers an element-construction error to render time — the
// element-level counterpart of ErrAttrib — so a constructor that cannot build a
// valid element can report the failure when the element is rendered instead of
// panicking or returning nil.
func NewErrElement(err error) *Element {
	return &Element{Err: err}
}

func (e *Element) Render(ctx context.Context, w Writer) error {
	if e.Err != nil {
		return e.Err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	err := w.BeginElement(e.Name)
	if err != nil {
		return err
	}

	for _, a := range e.Attribs {
		value, valErr := a.AttribValue(ctx)
		if valErr != nil {
			return valErr
		}
		err = w.Attribute(a.AttribName(), value)
		if err != nil {
			return err
		}
	}

	if e.Children == nil {
		return w.EndElement()
	}
	err = w.CloseElementStartTag()
	if err != nil {
		return err
	}

	err = e.Children.Render(ctx, w)
	if err != nil {
		return err
	}

	return w.EndElement()
}

func (e *Element) String() string {
	var b strings.Builder
	err := e.Render(context.Background(), NewCheckedWriter(&b).WithSingleQuoteAttribs())
	if err != nil {
		return "mx.Element.String: " + err.Error()
	}
	return b.String()
}

func (e *Element) AttribIndex(name string) int {
	for i, a := range e.Attribs {
		if a.AttribName() == name {
			return i
		}
	}
	return -1
}
