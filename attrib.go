package mx

import (
	"errors"
	"iter"
	"slices"
)

type Attrib struct {
	Name  string
	Value string
}

func (a Attrib) Validate() error {
	// TODO regex for valid attribute name
	if a.Name == "" {
		return errors.New("Attrib.Name is empty")
	}
	return nil
}

func (a Attrib) Valid() bool {
	return a.Validate() == nil
}

func AsAttrib(x any) (a Attrib, ok bool) {
	switch x := x.(type) {
	case Attrib:
		return x, true
	case func() Attrib:
		return x(), true
	default:
		return Attrib{}, false
	}
}

func AsAttribs(x any) (a []Attrib, ok bool) {
	// if attrib, ok := AsAttrib(x); ok {
	// 	return []Attrib{attrib}, true
	// }
	switch x := x.(type) {
	case []Attrib:
		return x, true
	case func() []Attrib:
		return x(), true
	case iter.Seq[Attrib]:
		return slices.Collect(x), true
	default:
		return nil, false
	}
}
