package mx

import (
	"database/sql/driver"
	"reflect"
)

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

func canBeNil(k reflect.Kind) bool {
	switch k {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		return true
	}
	return false
}

func IsZero(value any) bool {
	return value == nil || reflect.ValueOf(value).IsZero()
}
