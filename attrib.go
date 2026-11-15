package mx

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync/atomic"
	"unicode"
)

// The attribute escapers write line breaks and tabs as character references,
// so a value read back with getAttribute or dataset is the value written and
// the start tag stays on one line. A literal carriage return would be
// normalized to a line feed by the HTML parser, hence its reference too.
var (
	doubleQuoteAttribEscaper = strings.NewReplacer(
		`&`, "&amp;",
		`<`, "&lt;",
		`"`, "&quot;",
		"\n", "&#10;",
		"\r", "&#13;",
		"\t", "&#9;",
	)

	singleQuoteAttribEscaper = strings.NewReplacer(
		`&`, "&amp;",
		`<`, "&lt;",
		`'`, "&apos;",
		"\n", "&#10;",
		"\r", "&#13;",
		"\t", "&#9;",
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
// A string without '=' is both name and value,
// as for a boolean attribute: ConstAttrib("hidden") is hidden="hidden".
type ConstAttrib string

var _ Attrib = ConstAttrib("")

// AttribName returns the part of the "name=value" string before the first '=',
// or the whole string if it has no '='.
func (a ConstAttrib) AttribName() string {
	name, _, _ := strings.Cut(string(a), "=")
	return name
}

// AttribValue returns the part of the "name=value" string after the first '=',
// or the whole string if it has no '=', and a nil error;
// the value is static and does not depend on the context.
func (a ConstAttrib) AttribValue(context.Context) (string, error) {
	name, value, found := strings.Cut(string(a), "=")
	if !found {
		return name, nil
	}
	return value, nil
}

// String returns the attribute formatted as name='value' with the value
// single-quote escaped, as produced by AttribString, like [Attribute.String]:
// ConstAttrib("type=text").String() is "type='text'".
func (a ConstAttrib) String() string {
	return AttribString(a)
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

// Validate returns the error of [ValidateAttribName] for the attribute Name.
func (a Attribute) Validate() error {
	return ValidateAttribName(a.Name)
}

// ValidateAttribName returns an error if name cannot be written as an
// attribute name: if it is empty or holds a control character, whitespace or
// one of the characters " ' > / = that the HTML syntax forbids in a name
// because they would end the name or the start tag, or <, which the syntax
// allows but no real attribute uses. XML names forbid all of these too, so the
// check holds for every markup target. A name passing it cannot break out of
// the start tag it is written into; [CheckedWriter.Attribute] refuses every
// other name with this error.
func ValidateAttribName(name string) error {
	if name == "" {
		return errors.New("empty attribute name")
	}
	for _, r := range name {
		if unicode.IsControl(r) || unicode.IsSpace(r) || strings.ContainsRune(`"'>/=<`, r) {
			return fmt.Errorf("attribute name %q contains the invalid character %q", name, r)
		}
	}
	return nil
}

// ValidAttribName reports whether [ValidateAttribName] returns no error.
func ValidAttribName(name string) bool {
	return ValidateAttribName(name) == nil
}

// IsNilAttrib reports whether a is nil or holds a nil pointer or func. Such a
// typed nil is a non-nil interface, so a == nil misses it, but calling its
// AttribName or AttribValue usually panics. Code that skips or refuses a nil
// Attrib checks with IsNilAttrib to treat a typed nil the same way.
func IsNilAttrib(a Attrib) bool {
	if a == nil {
		return true
	}
	switch v := reflect.ValueOf(a); v.Kind() {
	case reflect.Pointer, reflect.Func:
		return v.IsNil()
	}
	return false
}

// Valid reports whether Validate returns no error.
func (a Attribute) Valid() bool {
	return a.Validate() == nil
}

// AttribString formats an Attrib as name='value' with name and value
// single-quote escaped, evaluating AttribValue with a background context. If
// AttribValue returns an error, the value is rendered as "!ERROR: " followed
// by the error message. It is intended for debugging and String methods, not
// markup output.
func AttribString(a Attrib) string {
	value, err := a.AttribValue(context.Background())
	if err != nil {
		value = "!ERROR: " + err.Error()
	}
	return formatEscapedAttrib(a.AttribName(), value)
}

// AttribStringContext formats an Attrib as name='value' like [AttribString],
// but evaluates AttribValue with ctx and returns its error instead of
// formatting it into the value. Name and value are escaped for a
// single-quoted attribute (&, <, ' and line breaks become character
// references), so the result is also safe as text content, whatever the
// name holds, and must not be escaped again.
func AttribStringContext(ctx context.Context, a Attrib) (string, error) {
	value, err := a.AttribValue(ctx)
	if err != nil {
		return "", err
	}
	return formatEscapedAttrib(a.AttribName(), value), nil
}

// formatEscapedAttrib returns name='value' with name and value escaped by
// singleQuoteAttribEscaper. The name is escaped too because it is not
// validated here: an invalid name must not inject markup either.
func formatEscapedAttrib(name, value string) string {
	return singleQuoteAttribEscaper.Replace(name) + "='" + singleQuoteAttribEscaper.Replace(value) + "'"
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
	// A deferred invalid-attribute-value error is a validation result
	// surfaced at render time, so it carries no callstack.
	return ErrAttrib{Name: name, Err: fmt.Errorf(format, args...)}
}
