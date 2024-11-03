package html

import (
	"context"

	"github.com/ungerik/go-mx"
)

type Element struct {
	Name     string
	Attribs  []mx.Attrib
	Children mx.Components // nil for void element, empty slice for no children
}

func NewElement(name string, attribsChildren ...any) *Element {
	e := &Element{Name: name, Children: []mx.Component{}}
	for _, x := range attribsChildren {
		if attrib, ok := x.(mx.Attrib); ok {
			e.Attribs = append(e.Attribs, attrib)
		} else if x != nil {
			e.Children = append(e.Children, mx.AsComponent(x))
		}
	}
	return e
}

func NewVoidElement(name string, attribs ...mx.Attrib) *Element {
	return &Element{Name: name, Attribs: attribs}
}

func (e *Element) Render(ctx context.Context, w mx.Writer) error {
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
		return w.CloseAndEndElement()
	}
	err = w.CloseElement()
	if err != nil {
		return err
	}

	err = e.Children.Render(ctx, w)
	if err != nil {
		return err
	}

	return w.EndElement()
}
