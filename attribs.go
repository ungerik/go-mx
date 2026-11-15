package mx

import (
	"fmt"
	"iter"
	"maps"
	"reflect"
	"slices"
	"strings"
	"sync"
)

// DefaultAsAttribs tries to convert an arbitrary value into a slice of
// [Attrib], returning ok false if the value is not one of the recognized
// attribute forms. It is how element constructors decide whether a variadic
// argument is an attribute (versus a child component).
//
// Recognized are a single [Attrib] or [Attribute], slices, iterators and
// constructor functions of those, an [AttribProvider], a map[string]any
// ([Attribs]) or map[string]string, and a struct (or non-nil pointer to one)
// whose fields carry "attr" tags, which are read via [ReflectAttribs].
//
// DefaultAsAttribs is the default implementation of the package-level
// [AsAttribs] variable, which may be reassigned to customize this.
func DefaultAsAttribs(maybeAttrib any) (a []Attrib, ok bool) {
	if x, ok := maybeAttrib.(Attrib); ok {
		// Kept even as a typed nil, for Element.Render to refuse.
		return []Attrib{x}, true
	}
	switch maybeAttrib.(type) {
	case func() Attrib, func() Attribute, func() []Attrib, func() []Attribute,
		iter.Seq[Attrib], iter.Seq[Attribute], AttribProvider:
		if isNilPointerOrFunc(maybeAttrib) {
			// Calling it, or the Attribs method of a typed-nil
			// AttribProvider, would usually panic.
			return nil, true
		}
	}

	switch x := maybeAttrib.(type) {
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
	if v.Kind() == reflect.Struct && hasAttrTags(v.Type()) {
		if a := ReflectAttribs(v, "attr"); len(a) > 0 {
			return a, true
		}
	}

	return nil, false
}

// structTypeHasAttrTags caches hasAttrTags per struct type.
var structTypeHasAttrTags sync.Map // reflect.Type -> bool

// hasAttrTags reports whether ReflectAttribs can find an "attr" tagged field
// in the struct type t, caching the answer per type, so that DefaultAsAttribs
// skips ReflectAttribs, which allocates on every call, for the many structs
// without such tags: errors, fmt.Stringers and other children that
// DefaultAsComponent asks AsAttribs about first.
func hasAttrTags(t reflect.Type) bool {
	if has, ok := structTypeHasAttrTags.Load(t); ok {
		return has.(bool)
	}
	has := false
	for _, f := range reflect.VisibleFields(t) {
		// The fields ReflectStructFields yields, see there.
		if f.Anonymous || !f.IsExported() {
			continue
		}
		if name := f.Tag.Get("attr"); name != "" && name != "-" {
			has = true
			break
		}
	}
	structTypeHasAttrTags.Store(t, has)
	return has
}

func toAttribSlice[T Attrib](attribs []T) []Attrib {
	result := make([]Attrib, len(attribs))
	for i, a := range attribs {
		result[i] = a
	}
	return result
}

// AttribProvider is implemented by types that can supply their attributes
// as a slice of [Attrib]. [DefaultAsAttribs] recognizes such values.
type AttribProvider interface {
	Attribs() []Attrib
}

// Attribs maps attribute names to values of any type, which are stringified
// with fmt.Sprint when rendered. It implements [AttribProvider] via its
// Attribs method, so a literal map can be passed where attributes are expected.
type Attribs map[string]any

// AttribsFromStringMap returns an Attribs built from a map of string values.
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

// ReflectAttribs builds attributes from the fields of struct value s,
// iterating flattened fields (including embedded structs) via
// [ReflectStructFields]. If tag is non-empty, a field's attribute name is
// taken from that struct tag and the field is skipped when the tag is absent
// or "-"; if tag is empty, the lower-cased field name is used. Each value is
// stringified with fmt.Sprint.
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
