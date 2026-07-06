package mx

import (
	"encoding"
	"reflect"
)

// DetectField inspects field and value and returns the [FieldKind]
// plus the parsed [FormTag]. Every layered FieldDecider (html, hx,
// shadcn, custom) calls this so the rules for "what KIND of widget
// does this field need" live in one place — the layers only differ in
// how they RENDER each kind.
//
// Detection priority (top-down, first match wins):
//
//  1. form:"-" or form:"widget=skip" tag → FieldKindSkip
//  2. form:"hidden" tag → FieldKindHidden
//  3. form:"section=..." or form:"nested" tag → FieldKindSection /
//     FieldKindNested
//  4. Anonymous embedded struct field, no section/nested tag →
//     FieldKindInline (the walker visits its visible fields instead)
//  5. form:"widget=textarea" tag → FieldKindTextarea
//  6. form:"widget=file" tag → FieldKindFile
//  7. form:"widget=<other>" tag (text/email/url/tel/date/datetime/
//     time/password/select/radio/switch/checkbox) → FieldKindString,
//     FieldKindDateTime, or FieldKindBool depending on the widget name
//  8. Type implementing [FormWidgetHint] → kind from the hint's name
//  9. Type implementing the Enums()/EnumStrings() convention →
//     FieldKindEnum
// 10. map[T]struct{} where T satisfies the enum convention →
//     FieldKindEnumSet
// 11. []T where T satisfies the enum convention → FieldKindEnumSet
// 12. bool / *bool → FieldKindBool
// 13. int*/uint*/float* and pointer variants → FieldKindNumber
// 14. time.Time (and any single-value-registered struct type) →
//     FieldKindDateTime
// 15. string / *string / named string types → FieldKindString
// 16. []string / named-slice-of-string → FieldKindTextarea
// 17. []byte / json.RawMessage-shaped types → FieldKindTextarea
// 18. struct field, no tag, not single-value, has TextMarshaler/
//     TextUnmarshaler → FieldKindCatchAll (renderer falls back to a
//     text input)
// 19. fallthrough → FieldKindCatchAll
//
// The first rule that matches wins. Notes:
//   - "satisfies the enum convention" means the type has either
//     Enums() []T or EnumStrings() []string as a method.
//   - Pointer types are normalized to their element type before the
//     structural checks (12-17). The Nullable interface, when present,
//     is orthogonal: detection still picks the same kind, and the
//     decider's Render checks IsNull at render time.
//   - Before rule 3, a form:"repeatable" tag on a direct []struct /
//     []*struct field (element not a single-value type) yields
//     FieldKindRepeatable; on any other shape the tag is ignored.
func DetectField(path FieldPath, field reflect.StructField, value reflect.Value) (FieldKind, FormTag) {
	tag := ParseFormTag(field)

	// 1. explicit skip
	if tag.Skip {
		return FieldKindSkip, tag
	}

	// 2. explicit hidden
	if tag.Hidden {
		return FieldKindHidden, tag
	}

	// 2.5. repeatable slice-of-struct. The tag only takes effect on a
	// direct slice field whose element is a plain struct or a single
	// pointer to one ([]T or []*T) that is not registered as a
	// single-value type (so []time.Time is not turned into rows). We
	// dereference at most one pointer level on purpose — not via
	// derefType, which would also strip an interface element (panicking
	// on Elem()) or multiple pointer levels that the row render/parse
	// paths cannot bind. On any other shape the tag is ignored and
	// detection falls through to the structural rules below.
	if tag.Repeatable && field.Type.Kind() == reflect.Slice {
		et := field.Type.Elem()
		if et.Kind() == reflect.Pointer {
			et = et.Elem()
		}
		if et.Kind() == reflect.Struct && !SingleValueTypes.Has(et) {
			return FieldKindRepeatable, tag
		}
	}

	// 3. explicit section / nested recursion
	if tag.Nested || tag.Section != "" {
		if tag.Section != "" {
			return FieldKindSection, tag
		}
		return FieldKindNested, tag
	}

	// 4. anonymous embed with no tag → inline
	if field.Anonymous {
		ft := derefType(field.Type)
		if ft.Kind() == reflect.Struct && !SingleValueTypes.Has(ft) {
			return FieldKindInline, tag
		}
	}

	// 5-7. widget tag override
	if tag.Widget != "" {
		switch tag.Widget {
		case "textarea":
			return FieldKindTextarea, tag
		case "file":
			return FieldKindFile, tag
		case "date", "datetime", "datetime-local", "time":
			return FieldKindDateTime, tag
		case "switch", "checkbox":
			return FieldKindBool, tag
		case "select", "radio":
			return FieldKindEnum, tag
		case "text", "email", "url", "tel", "password":
			return FieldKindString, tag
		case "hidden":
			return FieldKindHidden, tag
		}
		// Unrecognized widget — pass through as string and let the
		// renderer interpret tag.Widget directly.
		return FieldKindString, tag
	}

	// 8. self-declared widget hint
	if t := field.Type; implementsFormWidgetHint(t) {
		// We don't call the hint here (would require an addressable
		// value), but its presence flips us into the appropriate kind.
		// Renderers re-query the hint to get the actual widget name.
		switch hintWidget(value, t) {
		case "textarea":
			return FieldKindTextarea, tag
		case "file":
			return FieldKindFile, tag
		case "date", "datetime", "datetime-local", "time":
			return FieldKindDateTime, tag
		case "switch", "checkbox":
			return FieldKindBool, tag
		case "select", "radio":
			return FieldKindEnum, tag
		case "hidden":
			return FieldKindHidden, tag
		default:
			return FieldKindString, tag
		}
	}

	// 9. Enums / EnumStrings convention
	if hasEnumMethods(field.Type) {
		return FieldKindEnum, tag
	}

	// 10-11. set / slice of enum
	ft := derefType(field.Type)
	switch ft.Kind() {
	case reflect.Map:
		if ft.Elem().Kind() == reflect.Struct && ft.Elem().NumField() == 0 &&
			hasEnumMethods(ft.Key()) {
			return FieldKindEnumSet, tag
		}
	case reflect.Slice:
		// []byte / json.RawMessage-shape gets textarea (rule 17 below);
		// other slice elements that satisfy the enum convention get
		// the enum-set kind.
		if ft.Elem().Kind() != reflect.Uint8 && hasEnumMethods(ft.Elem()) {
			return FieldKindEnumSet, tag
		}
	}

	// 12. bool / *bool
	if ft.Kind() == reflect.Bool {
		return FieldKindBool, tag
	}

	// 13. numeric
	switch ft.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return FieldKindNumber, tag
	}

	// 14. single-value struct types (time.Time, FormWidgetHint impls)
	if ft.Kind() == reflect.Struct && SingleValueTypes.Has(ft) {
		return FieldKindDateTime, tag
	}

	// 15. string-shaped
	if ft.Kind() == reflect.String {
		return FieldKindString, tag
	}

	// 16. []string and []byte / json.RawMessage
	if ft.Kind() == reflect.Slice {
		switch ft.Elem().Kind() {
		case reflect.Uint8:
			return FieldKindTextarea, tag
		case reflect.String:
			return FieldKindTextarea, tag
		}
	}

	// 17. struct field falling through with TextMarshaler /
	// TextUnmarshaler → catch-all (parse-aware), otherwise still
	// catch-all (renderer must surface the registration error).
	return FieldKindCatchAll, tag
}

// derefType returns the element type of a pointer/interface chain,
// stopping at the first non-pointer/non-interface type. Useful for
// "what is this value really" checks where pointer-ness is incidental.
func derefType(t reflect.Type) reflect.Type {
	for t != nil && (t.Kind() == reflect.Pointer || t.Kind() == reflect.Interface) {
		t = t.Elem()
	}
	return t
}

// hasEnumMethods reports whether t (or its element type, if pointer)
// implements either Enums()/EnumStrings() convention. The methods are
// detected by name and signature shape — we accept any method named
// "Enums" that returns a slice of any element type, and any method
// named "EnumStrings" that returns []string.
func hasEnumMethods(t reflect.Type) bool {
	if t == nil {
		return false
	}
	check := func(tt reflect.Type) bool {
		if tt == nil {
			return false
		}
		for _, name := range []string{"EnumStrings", "Enums"} {
			m, ok := tt.MethodByName(name)
			if !ok {
				continue
			}
			ft := m.Func.Type()
			// receiver + zero or more args, returning exactly one slice
			if ft.NumOut() != 1 {
				continue
			}
			if ft.Out(0).Kind() != reflect.Slice {
				continue
			}
			if name == "EnumStrings" && ft.Out(0).Elem().Kind() != reflect.String {
				continue
			}
			return true
		}
		return false
	}
	if check(t) {
		return true
	}
	// Try pointer receiver if t is not already pointer.
	if t.Kind() != reflect.Pointer {
		if check(reflect.PointerTo(t)) {
			return true
		}
	}
	return false
}

// implementsFormWidgetHint reports whether t (or its pointer form)
// implements [FormWidgetHint].
func implementsFormWidgetHint(t reflect.Type) bool {
	if t == nil {
		return false
	}
	hint := reflect.TypeFor[FormWidgetHint]()
	if t.Implements(hint) {
		return true
	}
	if t.Kind() != reflect.Pointer {
		if reflect.PointerTo(t).Implements(hint) {
			return true
		}
	}
	return false
}

// hintWidget calls FormWidget() on v's value or addressable value.
// Returns "" when no hint method can be safely invoked.
func hintWidget(v reflect.Value, t reflect.Type) string {
	hint := reflect.TypeFor[FormWidgetHint]()
	if t.Implements(hint) {
		if v.IsValid() && (t.Kind() != reflect.Pointer || !v.IsNil()) {
			if h, ok := v.Interface().(FormWidgetHint); ok {
				return h.FormWidget()
			}
		}
		// Fallback: use a zero value to call the method (works for
		// value receivers).
		if h, ok := reflect.New(t).Elem().Interface().(FormWidgetHint); ok {
			return h.FormWidget()
		}
	}
	if t.Kind() != reflect.Pointer && reflect.PointerTo(t).Implements(hint) {
		if v.IsValid() && v.CanAddr() {
			if h, ok := v.Addr().Interface().(FormWidgetHint); ok {
				return h.FormWidget()
			}
		}
		if h, ok := reflect.New(t).Interface().(FormWidgetHint); ok {
			return h.FormWidget()
		}
	}
	return ""
}

// ImplementsTextUnmarshaler reports whether t (or its pointer form)
// implements [encoding.TextUnmarshaler]. Renderers use it on the
// catch-all path to decide whether a field is parseable from a string.
func ImplementsTextUnmarshaler(t reflect.Type) bool {
	if t == nil {
		return false
	}
	tu := reflect.TypeFor[encoding.TextUnmarshaler]()
	if t.Implements(tu) {
		return true
	}
	if t.Kind() != reflect.Pointer {
		return reflect.PointerTo(t).Implements(tu)
	}
	return false
}

// ImplementsTextMarshaler mirrors [ImplementsTextUnmarshaler] for the
// render path: when a catch-all field's type implements TextMarshaler,
// the renderer can use MarshalText to produce the displayed value
// instead of falling back to fmt.Sprint.
func ImplementsTextMarshaler(t reflect.Type) bool {
	if t == nil {
		return false
	}
	tm := reflect.TypeFor[encoding.TextMarshaler]()
	if t.Implements(tm) {
		return true
	}
	if t.Kind() != reflect.Pointer {
		return reflect.PointerTo(t).Implements(tm)
	}
	return false
}
