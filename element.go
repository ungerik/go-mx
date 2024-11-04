package mx

import (
	"context"
)

type Element struct {
	Name     string
	Attribs  []Attrib
	Children Components // nil for void element, empty slice for no children
}

func NewElement(name string, attribsChildren ...any) *Element {
	e := &Element{Name: name, Children: []Component{}}
	for _, ac := range attribsChildren {
		if ac == nil {
			continue
		}
		if attrib, ok := AsAttrib(ac); ok {
			e.Attribs = append(e.Attribs, attrib)
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

func (e *Element) Render(ctx context.Context, w Writer) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	err := w.BeginElement(e.Name)
	if err != nil {
		return err
	}

	for _, attrib := range e.Attribs {
		err = w.Attribute(attrib.Name, attrib.Value)
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
