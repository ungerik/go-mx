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

// SentinelRow is the hidden-input name prefix that marks one rendered
// row of a form:"repeatable" field. The full input name is
// SentinelRow + string(rowPath), where rowPath is the repeatable
// field's path with the row index appended (e.g. "Lines-0"). The
// parser scans these markers to discover which rows the client
// submitted, so removing a row's markup on the client (or via the
// server-side remove button) drops the row from the bound slice.
const SentinelRow = "__row__"

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

// RowSentinelName returns the form-input name used to mark rowPath as a
// rendered row of a repeatable field. rowPath is the repeatable field's
// path with the row index appended. The name is prefixed with
// [SentinelRow].
func RowSentinelName(rowPath FieldPath) string {
	return SentinelRow + string(rowPath)
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
