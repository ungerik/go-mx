package mx

import (
	"errors"
	"fmt"
	"iter"
	"slices"
	"strings"
)

var (
	doubleQuoteAttribEscaper = strings.NewReplacer(
		`&`, "&amp;",
		`<`, "&lt;",
		`"`, "&quot;",
		"\n", " ",
		"\t", "  ",
	)
	singleQuoteAttribEscaper = strings.NewReplacer(
		`&`, "&amp;",
		`<`, "&lt;",
		`'`, "&apos;",
		"\n", " ",
		"\t", "  ",
	)
)

type Attrib struct {
	Name  string
	Value string
}

func Attribute(name, value string) Attrib {
	return Attrib{Name: name, Value: value}
}

func AppendAttrib(attribs []Attrib, name, value string) []Attrib {
	return append(attribs, Attrib{Name: name, Value: value})
}

func PrependAttrib(name, value string, attribs []Attrib) []Attrib {
	return append([]Attrib{{Name: name, Value: value}}, attribs...)
}

func (a Attrib) String() string {
	return fmt.Sprintf("%s='%s'", a.Name, singleQuoteAttribEscaper.Replace(a.Value))
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
