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

// Attrib is an HTML or SVG attribute rendered as name="value".
//
// AttribName returns the static attribute name. AttribValue returns the value,
// which may be produced dynamically: an implementation can derive it from the
// passed context (request-scoped data, a lookup the context can cancel, and so
// on), so the call may fail and therefore returns an error. A non-nil error
// aborts rendering of the whole element; otherwise the value is rendered escaped
// and an empty value with a nil error is a valid (empty) attribute. Simple
// static attributes ignore the context and return their value with a nil error.
// A constructor that detects an invalid value up front may return an ErrAttrib,
// deferring the error to render time.
type Attrib interface {
	AttribName() string
	AttribValue(context.Context) (string, error)
}

// NewAttrib returns an Attrib with the given name and static value.
func NewAttrib(name, value string) Attrib {
	return Attribute{Name: name, Value: value}
}

// NewAttribf returns an Attrib with the given name and a value formatted
// from valueFmt and a using fmt.Sprintf.
func NewAttribf(name, valueFmt string, a ...any) Attrib {
	return Attribute{Name: name, Value: fmt.Sprintf(valueFmt, a...)}
}

// AppendAttrib appends an Attrib with the given name and value to attribs
// and returns the extended slice.
func AppendAttrib(attribs []Attrib, name, value string) []Attrib {
	return append(attribs, Attribute{Name: name, Value: value})
}

// PrependAttrib returns a new slice with an Attrib of the given name and
// value placed before the elements of attribs.
func PrependAttrib(name, value string, attribs []Attrib) []Attrib {
	return append([]Attrib{Attribute{Name: name, Value: value}}, attribs...)
}

// ConstAttrib implements the Attrib interface
// and holds the name and value
// as a string with the format "name=value".
type ConstAttrib string

var _ Attrib = ConstAttrib("")

// AttribName returns the part of the "name=value" string before the first '='.
func (a ConstAttrib) AttribName() string {
	return string(a)[:strings.IndexByte(string(a), '=')]
}

// AttribValue returns the part of the "name=value" string after the first '='
// and a nil error; the value is static and does not depend on the context.
func (a ConstAttrib) AttribValue(context.Context) (string, error) {
	return string(a)[strings.IndexByte(string(a), '=')+1:], nil
}

// Attribute is an Attrib holding a static name and value pair.
type Attribute struct {
	Name  string
	Value string
}

var _ Attrib = Attribute{}

// AttribName returns the attribute Name.
func (a Attribute) AttribName() string {
	return a.Name
}

// AttribValue returns the static attribute Value and a nil error,
// ignoring the context.
func (a Attribute) AttribValue(context.Context) (string, error) {
	return a.Value, nil
}

// String returns the attribute formatted as name='value' with the value
// single-quote escaped, as produced by AttribString.
func (a Attribute) String() string {
	return AttribString(a)
}

// Validate returns an error if the attribute Name is empty.
func (a Attribute) Validate() error {
	// TODO regex for valid attribute name
	if a.Name == "" {
		return errors.New("Attrib.Name is empty")
	}
	return nil
}

// Valid reports whether Validate returns no error.
func (a Attribute) Valid() bool {
	return a.Validate() == nil
}

// AttribString formats an Attrib as name='value' with the value single-quote
// escaped, evaluating AttribValue with a background context. If AttribValue
// returns an error, the value is rendered as "!ERROR: " followed by the error
// message. It is intended for debugging and String methods, not markup output.
func AttribString(a Attrib) string {
	value, err := a.AttribValue(context.Background())
	if err != nil {
		value = "!ERROR: " + err.Error()
	}
	return fmt.Sprintf("%s='%s'", a.AttribName(), singleQuoteAttribEscaper.Replace(value))
}

// ErrAttrib is an Attrib whose AttribValue always returns Err, deferring an
// attribute-construction error to render time. A constructor that detects an
// invalid value returns an ErrAttrib instead of panicking or emitting broken
// markup, so the error surfaces — and aborts rendering — when the element
// holding it is rendered.
type ErrAttrib struct {
	Name string
	Err  error
}

var _ Attrib = ErrAttrib{}

// AttribName returns the name of the attribute this ErrAttrib stands in for, so
// it occupies the correct slot in attribute deduplication and lookups. The error
// is reported by AttribValue, which aborts rendering of the enclosing element.
func (a ErrAttrib) AttribName() string { return a.Name }

// AttribValue always returns Err, aborting rendering of the enclosing element.
func (a ErrAttrib) AttribValue(context.Context) (string, error) { return "", a.Err }

// ErrAttribf returns an ErrAttrib wrapping a formatted error.
func ErrAttribf(name, format string, args ...any) ErrAttrib {
	return ErrAttrib{Name: name, Err: fmt.Errorf(format, args...)}
}
