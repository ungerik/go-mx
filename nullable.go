package mx

import (
	"database/sql/driver"
	"reflect"
)

type Nullable interface {
	IsNull() bool
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
