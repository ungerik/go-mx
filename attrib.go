package mx

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"
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

	idCounter atomic.Uint64
)

type Attrib interface {
	Attrib(context.Context) (name, value string)
}

func NewAttrib(name, value string) Attrib {
	return Attribute{Name: name, Value: value}
}

func AppendAttrib(attribs []Attrib, name, value string) []Attrib {
	return append(attribs, Attribute{Name: name, Value: value})
}

func PrependAttrib(name, value string, attribs []Attrib) []Attrib {
	return append([]Attrib{Attribute{Name: name, Value: value}}, attribs...)
}

func UniqueID() Attrib {
	return uniqueID(idCounter.Add(1))
}

type uniqueID uint64

func (id uniqueID) Attrib(context.Context) (name, value string) {
	return "id", "_" + strconv.FormatUint(uint64(id), 36)
}

// Attribute implements the Attrib interface.
type Attribute struct {
	Name  string
	Value string
}

func (a Attribute) Attrib(context.Context) (name, value string) {
	return a.Name, a.Value
}

func (a Attribute) String() string {
	return fmt.Sprintf("%s='%s'", a.Name, singleQuoteAttribEscaper.Replace(a.Value))
}

func (a Attribute) Validate() error {
	// TODO regex for valid attribute name
	if a.Name == "" {
		return errors.New("Attrib.Name is empty")
	}
	return nil
}

func (a Attribute) Valid() bool {
	return a.Validate() == nil
}

func AsAttrib(x any) (a Attrib, ok bool) {
	switch x := x.(type) {
	case Attrib:
		return x, true
	case func() Attrib:
		return x(), true
	case func() Attribute:
		return x(), true
	default:
		return nil, false
	}
}

func AsAttribs(x any) (a []Attrib, ok bool) {
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
