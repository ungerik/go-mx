package mx

import (
	"iter"
	"reflect"
)

// ReflectStructFields returns an iterator over the exported, flattened struct
// fields of s with their values. Fields of anonymous embedded structs are
// yielded as if declared at the top level, in declaration order; anonymous
// fields themselves and unexported fields are skipped, and fields shadowed by
// outer fields with the same name are not yielded (Go's normal visibility
// rules, courtesy of [reflect.VisibleFields]).
//
// The argument s must be the [reflect.Value] of a struct or a pointer to a
// struct; otherwise ReflectStructFields panics. A nil pointer argument also
// panics. A nil embedded pointer-to-struct *field* is skipped silently — its
// promoted fields are unreachable, not an error — via [reflect.Value.FieldByIndexErr].
func ReflectStructFields(s reflect.Value) iter.Seq2[reflect.StructField, reflect.Value] {
	sv := s
	for sv.Kind() == reflect.Pointer {
		if sv.IsNil() {
			panic("nil pointer to " + sv.Type().String())
		}
		sv = sv.Elem()
	}
	if sv.Kind() != reflect.Struct {
		panic("need struct or pointer to struct, but got: " + s.Type().String())
	}
	return func(yield func(reflect.StructField, reflect.Value) bool) {
		for _, f := range reflect.VisibleFields(sv.Type()) {
			// VisibleFields includes the anonymous embed itself plus every
			// promoted field, exported or not. Keep only the exported,
			// non-anonymous ones to match the package's existing contract.
			if f.Anonymous || !f.IsExported() {
				continue
			}
			val, err := sv.FieldByIndexErr(f.Index)
			if err != nil {
				// Path crosses a nil embedded pointer-to-struct; skip the
				// whole subtree silently rather than panicking.
				continue
			}
			if !yield(f, val) {
				return
			}
		}
	}
}
