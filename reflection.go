package mx

import (
	"iter"
	"reflect"
)

// FlatExportedStructFields returns an iterator over flattened struct fields,
// meaning that the fields of anonoymous embedded fields are yielded
// as if they were at the top level of the struct.
//
// The argument s must be a struct or a pointer to a struct.
func FlatExportedStructFields(s reflect.Value) iter.Seq2[reflect.StructField, reflect.Value] {
	structValue := s
	for s.Kind() == reflect.Ptr {
		if s.IsNil() {
			panic("nil pointer to " + s.Type().String())
		}
		structValue = s.Elem()
	}
	structType := structValue.Type()
	if structType.Kind() != reflect.Struct {
		panic("need struct or pointer to struct, but got: " + s.Type().String())
	}
	return func(yield func(reflect.StructField, reflect.Value) bool) {
		for i := range structType.NumField() {
			field, val := structType.Field(i), structValue.Field(i)
			switch {
			case field.Anonymous:
				for fieldA, valA := range FlatExportedStructFields(val) {
					if !yield(fieldA, valA) {
						return
					}
				}
			case field.IsExported():
				if !yield(field, val) {
					return
				}
			}
		}
	}
}
