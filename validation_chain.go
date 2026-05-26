package mx

import (
	"errors"
	"reflect"
)

// RunValidationChain probes value for the first method present in the
// 4-rung validation chain and returns its errors. The chain order
// (richest first):
//
//  1. Normalize() []error — preferred; mutates value in place AND
//     returns every error found.
//  2. Normalize() error — same canonicalization, single error.
//  3. Validate() error — pure validator, no normalization.
//  4. Valid() bool — last resort; false produces a generic
//     "invalid value" error.
//
// Only one rung runs per call — the first match wins. The Normalize
// variants and any other pointer-receiver methods are invoked via
// Addr; the caller therefore typically passes a value that satisfies
// CanAddr. When value is not addressable, the chain still probes the
// value-receiver variants but skips the pointer-receiver ones.
func RunValidationChain(value reflect.Value) []error {
	if !value.IsValid() {
		return nil
	}

	// Step 1+2: Normalize variants. These are typically pointer
	// receivers because they mutate.
	if value.CanAddr() {
		ptr := value.Addr().Interface()
		if n, ok := ptr.(Normalizer); ok {
			return n.Normalize()
		}
		if n, ok := ptr.(SingleErrNormalizer); ok {
			if err := n.Normalize(); err != nil {
				return []error{err}
			}
			return nil
		}
	}
	// Value receivers — included for completeness; we accept them on
	// either the addressable form or the bare value.
	if iface := safeInterface(value); iface != nil {
		if n, ok := iface.(Normalizer); ok {
			return n.Normalize()
		}
		if n, ok := iface.(SingleErrNormalizer); ok {
			if err := n.Normalize(); err != nil {
				return []error{err}
			}
			return nil
		}
	}

	// Step 3: Validate() error
	if iface := safeInterface(value); iface != nil {
		if v, ok := iface.(SingleErrValidator); ok {
			if err := v.Validate(); err != nil {
				return []error{err}
			}
			return nil
		}
	}
	if value.CanAddr() {
		ptr := value.Addr().Interface()
		if v, ok := ptr.(SingleErrValidator); ok {
			if err := v.Validate(); err != nil {
				return []error{err}
			}
			return nil
		}
	}

	// Step 4: Valid() bool
	if iface := safeInterface(value); iface != nil {
		if v, ok := iface.(Validator); ok {
			if !v.Valid() {
				return []error{errors.New("invalid value")}
			}
			return nil
		}
	}
	if value.CanAddr() {
		ptr := value.Addr().Interface()
		if v, ok := ptr.(Validator); ok {
			if !v.Valid() {
				return []error{errors.New("invalid value")}
			}
			return nil
		}
	}

	return nil
}

// safeInterface returns value.Interface() unless reflect would panic
// (unexported or zero value.Value). Returns nil in that case.
func safeInterface(value reflect.Value) any {
	if !value.IsValid() || !value.CanInterface() {
		return nil
	}
	return value.Interface()
}
