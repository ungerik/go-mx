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
// back with Scan. ok is false for any other value, when Value fails, for a
// driver value of another type and for text an input cannot hold
// unchanged: invalid UTF-8, as binary bytes are, or a NUL.
func DriverValueText(value reflect.Value) (text string, ok bool) {
	if value.Kind() != reflect.Struct {
		return "", false
	}
	ptr := reflect.PointerTo(value.Type())
	if !ptr.Implements(reflect.TypeFor[sqlScanner]()) || ptr.Implements(reflect.TypeFor[encoding.TextMarshaler]()) ||
		ptr.Implements(reflect.TypeFor[encoding.TextUnmarshaler]()) || ptr.Implements(reflect.TypeFor[fmt.Stringer]()) {
		return "", false
	}
	valuer, isValuer := safeInterface(value).(driver.Valuer)
	if !isValuer && value.CanAddr() {
		valuer, isValuer = safeInterface(value.Addr()).(driver.Valuer)
	}
	if !isValuer {
		return "", false
	}
	dv, err := valuer.Value()
	if err != nil {
		return "", false
	}
	switch x := dv.(type) {
	case nil:
		return "", true
	case string:
		if inputText(x) {
			return x, true
		}
	case []byte:
		if inputText(string(x)) {
			return string(x), true
		}
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

// inputText reports whether an input holds s unchanged: a browser replaces
// invalid UTF-8 and a NUL.
func inputText(s string) bool {
	return utf8.ValidString(s) && !strings.ContainsRune(s, 0)
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
