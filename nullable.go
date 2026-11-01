package mx

import (
	"database/sql/driver"
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// Nullable is implemented by types that can report whether they hold a NULL
// (absent) value. The reflection-based [IsNull] helper recognizes it.
type Nullable interface {
	IsNull() bool
}

// Zeroable lets a type declare its own zero-value semantics. The
// reflection-based [IsZero] helper falls back to [reflect.Value.IsZero]
// for types that do not implement Zeroable.
type Zeroable interface {
	IsZero() bool
}

// NullSetter is implemented by nullable types whose value can be reset
// to NULL in place. The [ReflectFormHandler] parser calls SetNull on a
// field when the form submits the __clear sentinel for that field. The
// method is on the pointer receiver because it must mutate the value.
type NullSetter interface {
	SetNull()
}

// IsNull returns true if the passed value is nil
// or implements the Nullable interface and IsNull() returns true
// or implements the database/sql/driver.Valuer interface and Value() returns nil and no error.
func IsNull(value any) bool {
	// A nil pointer is null, and a value method called through it panics,
	// as the Value of a nil *sql.NullString does.
	if v := reflect.ValueOf(value); v.Kind() == reflect.Pointer && v.IsNil() {
		return true
	}
	switch x := value.(type) {
	case nil:
		return true
	case Nullable:
		return x.IsNull()
	case driver.Valuer:
		v, err := x.Value()
		return v == nil && err == nil
	}
	if v := reflect.ValueOf(value); canBeNil(v.Kind()) {
		return v.IsNil()
	}
	return false
}

// sqlScanner is database/sql.Scanner, without linking database/sql.
type sqlScanner interface{ Scan(src any) error }

// DriverValueText returns the text a reflected form writes for a struct a
// database driver reads and writes, such as sql.NullInt64: its Value
// (database/sql/driver.Valuer) gives the value it holds, its Scan
// (database/sql.Scanner) reads one back, and it has no MarshalText,
// UnmarshalText or String method, whose text a renderer or parser would
// use instead. The text is the driver value, a time in RFC 3339 as the
// forms write every time, and a null is "". The html form parser reads it
// back with Scan. ok is false for any other value, when Value fails, and
// for a driver value of another type.
//
// Text an input cannot hold unchanged, invalid UTF-8 (as binary bytes are)
// or a NUL, is written with each of them replaced by U+FFFD, so the input
// shows where they are; [DriverValueTextProblem] tells a renderer to say
// so, and the html parser keeps the stored text when that replacement comes
// back unchanged.
func DriverValueText(value reflect.Value) (text string, ok bool) {
	if !value.IsValid() || !DriverValueType(value.Type()) {
		return "", false
	}
	dv, ok := driverValue(value)
	if !ok {
		return "", false
	}
	switch x := dv.(type) {
	case nil:
		return "", true
	case string:
		return inputSafeText(x), true
	case []byte:
		return inputSafeText(string(x)), true
	case int64:
		return strconv.FormatInt(x, 10), true
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64), true
	case bool:
		return strconv.FormatBool(x), true
	case time.Time:
		return x.Format(time.RFC3339Nano), true
	}
	return "", false
}

// DriverValueType reports whether t is a struct a database driver reads
// and writes, which the forms bind with its Scan and show as
// [DriverValueText]: a pointer to it implements database/sql.Scanner, it
// or its pointer database/sql/driver.Valuer, and neither has a MarshalText,
// UnmarshalText or String method, whose text a renderer or parser would use
// instead. It is a property of the type, whatever a value holds.
func DriverValueType(t reflect.Type) bool {
	if t == nil || t.Kind() != reflect.Struct {
		return false
	}
	ptr := reflect.PointerTo(t)
	if !ptr.Implements(reflect.TypeFor[sqlScanner]()) || !ptr.Implements(reflect.TypeFor[driver.Valuer]()) ||
		ptr.Implements(reflect.TypeFor[encoding.TextMarshaler]()) || ptr.Implements(reflect.TypeFor[encoding.TextUnmarshaler]()) ||
		ptr.Implements(reflect.TypeFor[fmt.Stringer]()) {
		return false
	}
	return true
}

// TextProblem names why an input cannot hold a stored text unchanged
// ([DriverValueTextProblem]).
type TextProblem int

const (
	// TextProblemNUL is a NUL character, which a browser replaces.
	TextProblemNUL TextProblem = iota + 1
	// TextProblemInvalidUTF8 is text that is not valid UTF-8, such as
	// binary bytes, which a browser replaces.
	TextProblemInvalidUTF8
	// TextProblemLineBreak is a line break, which a browser strips from a
	// one-line text input.
	TextProblemLineBreak
)

// DriverValueTextProblem reports why an input cannot show the text a
// database/sql value ([DriverValueType]) holds unchanged, and whether it
// cannot: a NUL, invalid UTF-8 or, unless the input is multiline (a
// textarea), a line break. The input shows [DriverValueText] instead, and
// saving that unchanged keeps the stored text.
func DriverValueTextProblem(value reflect.Value, multiline bool) (TextProblem, bool) {
	value = reflect.Indirect(value)
	if !value.IsValid() || !DriverValueType(value.Type()) {
		return 0, false
	}
	dv, ok := driverValue(value)
	if !ok {
		return 0, false
	}
	var text string
	switch x := dv.(type) {
	case string:
		text = x
	case []byte:
		text = string(x)
	default:
		return 0, false
	}
	switch {
	case !utf8.ValidString(text):
		return TextProblemInvalidUTF8, true
	case strings.ContainsRune(text, 0):
		return TextProblemNUL, true
	case !multiline && strings.ContainsAny(text, "\r\n"):
		return TextProblemLineBreak, true
	}
	return 0, false
}

// driverValue returns the driver value of a [DriverValueType] value, and
// whether its Value method gave one.
func driverValue(value reflect.Value) (any, bool) {
	valuer, isValuer := safeInterface(value).(driver.Valuer)
	if !isValuer && value.CanAddr() {
		valuer, isValuer = safeInterface(value.Addr()).(driver.Valuer)
	}
	if !isValuer {
		return nil, false
	}
	dv, err := valuer.Value()
	if err != nil {
		return nil, false
	}
	return dv, true
}

// inputSafeText returns s with what a browser replaces in an input, invalid
// UTF-8 and a NUL, replaced by U+FFFD as the browser would.
func inputSafeText(s string) string {
	return strings.ReplaceAll(strings.ToValidUTF8(s, "\uFFFD"), "\x00", "\uFFFD")
}

func canBeNil(k reflect.Kind) bool {
	switch k {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		return true
	}
	return false
}

// IsZero reports whether value is nil or equal to the zero value of its type,
// using [reflect.Value.IsZero].
func IsZero(value any) bool {
	return value == nil || reflect.ValueOf(value).IsZero()
}
