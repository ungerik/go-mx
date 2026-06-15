package mx

import (
	"context"
	"strings"
)

// Element is a [Component] for a single markup element with a name, attributes
// and children. A nil Children means a void (self-closing) element that may not
// have content; an empty (non-nil) slice means a regular element with no
// children. A non-nil Err defers a construction error to render time (see
// [NewErrElement]).
type Element struct {
	Name     string
	Attribs  []Attrib
	Children Components // nil for void element, empty slice for no children
	Err      error      // if non-nil, Render returns it instead of rendering; see NewErrElement
}

// NewElement builds a regular (non-void) Element with the given name, sorting
// the variadic attribsChildren into attributes and children: each argument that
// [AsAttribs] recognizes as attributes is appended to Attribs, every other
// non-nil argument is converted with [AsComponent] and appended to Children. Nil
// arguments are ignored. The returned element always has a non-nil (possibly
// empty) Children slice, so it renders with a separate closing tag rather than
// as a void element.
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

// NewVoidElement builds a void (self-closing) Element with the given name and
// attributes. Its Children is nil, so it renders without a closing tag and
// cannot hold content (for example <img> or <br>).
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

// Render writes the element and its children to w. It first returns Err if set
// (the deferred-error pattern) and then the context error if the context is
// done. Each attribute's value is resolved via [Attrib.AttribValue] with ctx,
// aborting on the first error. A void element (nil Children) is closed without
// content; otherwise the children are rendered between the start and end tags.
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

// String renders the element to a string using a [CheckedWriter] with
// single-quoted attributes and a background context. It is meant for tests and
// debugging; a render error (including a deferred Err) is returned as
// "mx.Element.String: " followed by the error message.
func (e *Element) String() string {
	var b strings.Builder
	err := e.Render(context.Background(), NewCheckedWriter(&b).WithSingleQuoteAttribs())
	if err != nil {
		return "mx.Element.String: " + err.Error()
	}
	return b.String()
}

// AttribIndex returns the index of the first attribute with the given name in
// Attribs, or -1 if none has that name.
func (e *Element) AttribIndex(name string) int {
	for i, a := range e.Attribs {
		if a.AttribName() == name {
			return i
		}
	}
	return -1
}
