package mx

import (
	"database/sql/driver"
	"reflect"
	"strconv"
	"time"
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

// DriverValueText returns the text a reflected form writes for a struct
// implementing database/sql/driver.Valuer, such as sql.NullInt64, which has
// no text methods: the driver value it holds, a time in RFC 3339 as the
// forms write every time. The html form parser reads that text back with
// the type's Scan method. A null is "". ok is false for any other value,
// for a driver value of another type and when Value fails. Renderers
// prefer a type's own MarshalText.
func DriverValueText(value reflect.Value) (text string, ok bool) {
	if value.Kind() != reflect.Struct {
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
		return x, true
	case []byte:
		return string(x), true
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
