package mx

import (
	"fmt"
	"iter"
	"maps"
	"reflect"
	"slices"
	"strings"
)

func DefaultAsAttribs(maybeAttrib any) (a []Attrib, ok bool) {
	switch x := maybeAttrib.(type) {
	case Attrib:
		return []Attrib{x}, true

	case func() Attrib:
		return []Attrib{x()}, true

	case func() Attribute:
		return []Attrib{x()}, true

	case AttribProvider:
		return x.Attribs(), true

	case []Attrib:
		return x, true

	case []Attribute:
		return toAttribSlice(x), true

	case func() []Attrib:
		return x(), true

	case func() []Attribute:
		return toAttribSlice(x()), true

	case iter.Seq[Attrib]:
		return slices.Collect(x), true

	case iter.Seq[Attribute]:
		return toAttribSlice(slices.Collect(x)), true

	case map[string]any:
		return Attribs(x).Attribs(), true

	case map[string]string:
		return AttribsFromStringMap(x).Attribs(), true
	}

	v := reflect.ValueOf(maybeAttrib)
	for v.Kind() == reflect.Pointer && !v.IsNil() {
		v = v.Elem()
	}
	if v.Kind() == reflect.Struct {
		if a := ReflectAttribs(v, "attr"); len(a) > 0 {
			return a, true
		}
	}

	return nil, false
}

func toAttribSlice[T Attrib](attribs []T) []Attrib {
	result := make([]Attrib, len(attribs))
	for i, a := range attribs {
		result[i] = a
	}
	return result
}

type AttribProvider interface {
	Attribs() []Attrib
}

type Attribs map[string]any

func AttribsFromStringMap(m map[string]string) Attribs {
	attribs := make(Attribs, len(m))
	for name, value := range m {
		attribs[name] = value
	}
	return attribs
}

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

func ReflectAttribs(s reflect.Value, tag string) []Attrib {
	var attribs []Attrib
	for field, value := range ReflectStructFields(s) {
		var name string
		if tag != "" {
			name = field.Tag.Get(tag)
			if name == "" || name == "-" {
				continue
			}
		} else {
			name = strings.ToLower(field.Name)
		}
		attribs = AppendAttrib(attribs, name, fmt.Sprint(value))
	}
	return attribs
}
