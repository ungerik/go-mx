package mx

import (
	"context"
	"errors"
	"fmt"
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
	AttribName() string
	AttribValue(context.Context) string
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

// Attribute implements the Attrib interface.
type Attribute struct {
	Name  string
	Value string
}

func (a Attribute) AttribName() string {
	return a.Name
}

func (a Attribute) AttribValue(context.Context) string {
	return a.Value
}

func (a Attribute) String() string {
	return AttribString(a)
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

func AttribString(a Attrib) string {
	return fmt.Sprintf("%s='%s'", a.AttribName(), singleQuoteAttribEscaper.Replace(a.AttribValue(context.Background())))
}
