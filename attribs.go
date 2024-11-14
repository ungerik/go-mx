package mx

import (
	"fmt"
	"iter"
	"maps"
	"slices"
	"strings"
)

func AsAttribs(x any) (a []Attrib, ok bool) {
	switch x := x.(type) {
	case []Attrib:
		return x, true
	case []Attribute:
		a = make([]Attrib, len(x))
		for i, attribute := range x {
			a[i] = attribute
		}
		return a, true
	case func() []Attrib:
		return x(), true
	case func() []Attribute:
		return AsAttribs(x())
	case iter.Seq[Attrib]:
		return slices.Collect(x), true
	case iter.Seq[Attribute]:
		attributes := slices.Collect(x)
		a = make([]Attrib, len(attributes))
		for i, attribute := range attributes {
			a[i] = attribute
		}
		return a, true
	case AttribProvider:
		return x.Attribs(), true
	case map[string]any:
		return Attribs(x).Attribs(), false
	case map[string]string:
		a := make(Attribs, len(x))
		for name, value := range x {
			a[name] = value
		}
		return a.Attribs(), false
	default:
		return nil, false
	}
}

type AttribProvider interface {
	Attribs() []Attrib
}

type Attribs map[string]any

// Attribs returns the Attribs map as a slice of Attribs
// sorted by name except that "id" is always first and "class" is always second
// if there is an "id" attribute, otherwise first.
func (a Attribs) Attribs() []Attrib {
	if len(a) == 0 {
		return nil
	}
	names := slices.Collect(maps.Keys(a))
	slices.SortFunc(names, func(a, b string) int {
		// "id" should always be first
		// "class" should always be second if ther is an "id", otherwise first
		switch {
		case a == "id":
			return -1
		case b == "id":
			return +1
		case a == "class":
			return -1
		case b == "class":
			return +1
		}
		return strings.Compare(a, b)
	})
	attribs := make([]Attrib, len(a))
	for i, name := range names {
		attribs[i] = NewAttrib(name, fmt.Sprint(a[name]))
	}
	return attribs
}

func (a Attribs) String() string {
	var b strings.Builder
	for i, attrib := range a.Attribs() {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(AttribString(attrib))
	}
	return b.String()
}
