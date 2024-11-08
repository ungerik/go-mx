package mx

import (
	"iter"
	"reflect"
)

// FlatExportedStructFields returns an iterator over flattened struct fields,
// meaning that the fields of anonoymous embedded fields are yielded
// as if they were at the top level of the struct.
//
// The argument t must be the reflect.Type of a struct or a pointer to a struct.
func FlatExportedStructFields(t reflect.Type) iter.Seq[reflect.StructField] {
	structType := t
	for structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}
	if structType.Kind() != reflect.Struct {
		panic("need struct or pointer to struct, but got: " + t.String())
	}
	return func(yield func(reflect.StructField) bool) {
		for i := range structType.NumField() {
			field := structType.Field(i)
			switch {
			case field.Anonymous:
				for fieldA := range FlatExportedStructFields(field.Type) {
					if !yield(fieldA) {
						return
					}
				}
			case field.IsExported():
				if !yield(field) {
					return
				}
			}
		}
	}
}

// FlatExportedStructFieldsAndValues returns an iterator over
// flattened struct fields with their values,
// meaning that the fields of anonoymous embedded fields are yielded
// as if they were at the top level of the struct.
//
// The argument s must be the reflect.Value of a struct or a pointer to a struct.
func FlatExportedStructFieldsAndValues(s reflect.Value) iter.Seq2[reflect.StructField, reflect.Value] {
	structValue := s
	for structValue.Kind() == reflect.Ptr {
		if s.IsNil() {
			panic("nil pointer to " + structValue.Type().String())
		}
		structValue = structValue.Elem()
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
				for fieldA, valA := range FlatExportedStructFieldsAndValues(val) {
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
