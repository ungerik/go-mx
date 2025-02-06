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

func NewAttribf(name, valueFmt string, a ...any) Attrib {
	return Attribute{Name: name, Value: fmt.Sprintf(valueFmt, a...)}
}

func AppendAttrib(attribs []Attrib, name, value string) []Attrib {
	return append(attribs, Attribute{Name: name, Value: value})
}

func PrependAttrib(name, value string, attribs []Attrib) []Attrib {
	return append([]Attrib{Attribute{Name: name, Value: value}}, attribs...)
}

// ConstAttrib implements the Attrib interface
// and holds the name and value
// as a string with the format "name=value".
type ConstAttrib string

var _ Attrib = ConstAttrib("")

func (a ConstAttrib) AttribName() string {
	return string(a)[:strings.IndexByte(string(a), '=')]
}

func (a ConstAttrib) AttribValue(context.Context) string {
	return string(a)[strings.IndexByte(string(a), '=')+1:]
}

// Attribute implements the Attrib interface.
type Attribute struct {
	Name  string
	Value string
}

var _ Attrib = Attribute{}

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
