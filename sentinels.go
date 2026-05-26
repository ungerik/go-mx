package mx

import "strings"

// SentinelPresent is the hidden-input name prefix that marks a field as
// having been rendered into the form. The full input name is
// SentinelPresent + string(path). The parser uses this to distinguish
// "field was in the form, user cleared it" from "field was not in the
// form at all" — the latter is an allowlist miss that must never write
// back to the loaded struct (D11/D12 in the design).
const SentinelPresent = "__present__"

// SentinelClear is the hidden-input name prefix that marks a nullable
// field as having been explicitly cleared by the user. The full input
// name is SentinelClear + string(path). The parser calls SetNull on the
// addressable field value when the corresponding form value is non-empty.
const SentinelClear = "__clear__"

// PresentSentinelName returns the form-input name used to mark path as
// rendered. The name is prefixed with [SentinelPresent].
func PresentSentinelName(path FieldPath) string {
	return SentinelPresent + string(path)
}

// ClearSentinelName returns the form-input name used to mark path as
// explicitly cleared by the user. The name is prefixed with [SentinelClear].
func ClearSentinelName(path FieldPath) string {
	return SentinelClear + string(path)
}

// ParsePresentSentinel returns the field path encoded in name and true
// when name uses the [SentinelPresent] prefix; otherwise it returns
// "", false.
func ParsePresentSentinel(name string) (path FieldPath, ok bool) {
	suffix, ok := strings.CutPrefix(name, SentinelPresent)
	if !ok {
		return "", false
	}
	return FieldPath(suffix), true
}

// ParseClearSentinel returns the field path encoded in name and true
// when name uses the [SentinelClear] prefix; otherwise it returns
// "", false.
func ParseClearSentinel(name string) (path FieldPath, ok bool) {
	suffix, ok := strings.CutPrefix(name, SentinelClear)
	if !ok {
		return "", false
	}
	return FieldPath(suffix), true
}
