package mx

import (
	"encoding"
	"reflect"
	"time"
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
//  8. Type implementing [FormWidgetHint] → kind from the hint's name,
//     which is also returned as the tag's Widget
//  9. Type implementing the Enums()/EnumStrings() convention →
//     FieldKindEnum
// 10. map[T]struct{} where T satisfies the enum convention →
//     FieldKindEnumSet
// 11. []T where T satisfies the enum convention → FieldKindEnumSet
// 12. bool / *bool → FieldKindBool
// 13. int*/uint*/float* and pointer variants → FieldKindNumber
// 14. time.Time, a type defined as it, or a struct embedding it as its only
//     field whose text is RFC 3339 (as it is without text methods of its
//     own) → FieldKindDateTime (other single-value structs → catch-all)
// 15. string / *string / named string types → FieldKindString
// 16. []string / named-slice-of-string → FieldKindTextarea
// 17. []byte / json.RawMessage-shaped types → FieldKindTextarea
// 18. struct field, no tag, has TextMarshaler/
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
//     FieldKindRepeatable, and on a slice of times FieldKindSkip; on any
//     other shape the tag is ignored.
//   - Between rules 7 and 8, a form:"options=name" tag yields
//     FieldKindEnum (set-shaped fields: FieldKindEnumSet); the name
//     must be registered via [RegisterNamedOptions] before rendering.
//     On []byte and non-set map shapes the tag is ignored.
//   - Rules 9-11 accept the options-provider interfaces
//     ([OptionsProvider], [NamedOptionsProvider],
//     [NamedOptionsContextProvider]) as equivalent to the
//     Enums()/EnumStrings() convention.
//
// A field whose widget is "password", by its tag or by [FormWidgetHint],
// is form:"sensitive" too: a password input never shows its value, so the
// stored one stays out of the page and a blank submission keeps it.
func DetectField(path FieldPath, field reflect.StructField, value reflect.Value) (FieldKind, FormTag) {
	kind, tag := detectField(path, field, value)
	if tag.Widget == "password" {
		tag.Sensitive = true
	}
	return kind, tag
}

func detectField(path FieldPath, field reflect.StructField, value reflect.Value) (FieldKind, FormTag) {
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
	// single-value type. A slice of times (time.Time, a type defined as it
	// or a struct embedding only one) is skipped, as no rule renders it;
	// on any other shape the tag is ignored and detection falls through to
	// the structural rules below. We dereference at most one pointer level
	// on purpose — not via derefType, which would also strip an interface
	// element (panicking on Elem()) or multiple pointer levels that the row
	// render/parse paths cannot bind.
	if tag.Repeatable && field.Type.Kind() == reflect.Slice {
		et := field.Type.Elem()
		if et.Kind() == reflect.Pointer {
			et = et.Elem()
		}
		if et.Kind() == reflect.Struct && (isTimeLike(et) || embedsOnlyTime(et)) {
			// A time's fields are unexported, so its rows would have no
			// cells and every save would zero the stored times, and no
			// input binds a slice of times: the field is skipped, never
			// rendered or bound, and its stored value is left alone.
			tag.Skip = true
			return FieldKindSkip, tag
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
		if ft.Kind() == reflect.Struct && !SingleValueTypes.Has(ft) && !isTimeLike(ft) {
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

	// 7.5. options tag names a provider registered with
	// [RegisterNamedOptions] → option-driven widget. Set-shaped fields
	// (map[T]struct{} or []T) render as an enum-set of checkboxes,
	// other shapes as a single select. On shapes that cannot render or
	// parse as options — []byte and non-set maps — the tag is ignored
	// and detection falls through to the structural rules, mirroring
	// how form:"repeatable" handles unsupported shapes.
	if tag.Options != "" {
		ft := derefType(field.Type)
		switch ft.Kind() {
		case reflect.Map:
			if ft.Elem().Kind() == reflect.Struct && ft.Elem().NumField() == 0 {
				return FieldKindEnumSet, tag
			}
		case reflect.Slice:
			if ft.Elem().Kind() != reflect.Uint8 {
				return FieldKindEnumSet, tag
			}
		default:
			return FieldKindEnum, tag
		}
	}

	// 8. self-declared widget hint
	if t := field.Type; implementsFormWidgetHint(t) {
		// The hint's widget goes into the returned tag, so renderers and
		// parsers read the widget from one place, whether a tag or the
		// type declared it.
		tag.Widget = hintWidget(value, t)
		switch tag.Widget {
		case "skip":
			tag.Skip = true
			return FieldKindSkip, tag
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

	// 9. Enums / EnumStrings convention or an options-provider interface
	if hasEnumMethods(field.Type) || hasOptionsProviderInterfaces(field.Type) {
		return FieldKindEnum, tag
	}

	// 10-11. set / slice of enum
	ft := derefType(field.Type)
	switch ft.Kind() {
	case reflect.Map:
		if ft.Elem().Kind() == reflect.Struct && ft.Elem().NumField() == 0 &&
			(hasEnumMethods(ft.Key()) || hasOptionsProviderInterfaces(ft.Key())) {
			return FieldKindEnumSet, tag
		}
	case reflect.Slice:
		// []byte / json.RawMessage-shape gets textarea (rule 17 below);
		// other slice elements that satisfy the enum convention get
		// the enum-set kind.
		if ft.Elem().Kind() != reflect.Uint8 &&
			(hasEnumMethods(ft.Elem()) || hasOptionsProviderInterfaces(ft.Elem())) {
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

	// 14. time-like struct types: time.Time, a type defined as it, or a
	// struct embedding a time.Time as its only field whose text is RFC 3339
	// (as it is without text methods of its own). Another single-value
	// struct, such as a registered money type, is a text input through its
	// MarshalText and UnmarshalText (the catch-all below).
	if ft.Kind() == reflect.Struct && isTimeLike(ft) {
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

// isTimeLike reports whether the struct type t holds a time the way a
// time.Time does: it is time.Time, a type defined as time.Time, or a struct
// whose only field is an embedded time.Time and whose text is that time's
// RFC 3339. A struct that writes its own text, such as a date-only one,
// stays a text input through its MarshalText and UnmarshalText.
func isTimeLike(t reflect.Type) bool {
	timeType := reflect.TypeFor[time.Time]()
	if t.ConvertibleTo(timeType) {
		return true
	}
	return embedsOnlyTime(t) && textIsRFC3339(t)
}

// embedsOnlyTime reports whether the struct type t holds nothing but an
// embedded time.Time, whatever text methods of its own it has.
func embedsOnlyTime(t reflect.Type) bool {
	return t.NumField() == 1 && t.Field(0).Anonymous && t.Field(0).Type == reflect.TypeFor[time.Time]()
}

// textIsRFC3339 reports whether the zero value of t writes its text as
// RFC 3339, as a time.Time does and so does a type embedding one without
// text methods of its own.
func textIsRFC3339(t reflect.Type) (ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	m, isMarshaler := reflect.New(t).Interface().(encoding.TextMarshaler)
	if !isMarshaler {
		return false
	}
	text, err := m.MarshalText()
	if err != nil {
		return false
	}
	_, err = time.Parse(time.RFC3339Nano, string(text))
	return err == nil
}

// derefType returns the element type of a pointer chain, stopping at
// the first non-pointer type. Useful for "what is this value really"
// checks where pointer-ness is incidental. Interface types are
// returned as-is: reflect.Type.Elem panics on interface kinds (a type
// has no element, only a value does), so callers' structural switches
// simply fall through for interface-typed fields.
func derefType(t reflect.Type) reflect.Type {
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}

// hasEnumMethods reports whether t (or its element type, if pointer)
// implements either Enums()/EnumStrings() convention, like the methods
// go-enum generates. The methods are detected by name and signature
// shape — we accept any method named "Enums" without parameters that
// returns a slice of any element type, and any method named
// "EnumStrings" without parameters that returns []string. A method with
// parameters could not be called to collect the options.
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
			// only the receiver, returning exactly one slice
			if ft.NumIn() != 1 || ft.NumOut() != 1 {
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
		// value receivers). A pointer type gets a pointer to a zero
		// element instead of a nil pointer, which would panic in a
		// value-receiver method.
		zero := reflect.New(t).Elem()
		if t.Kind() == reflect.Pointer {
			zero = reflect.New(t.Elem())
		}
		if h, ok := zero.Interface().(FormWidgetHint); ok {
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
