package mx

import (
	"iter"
	"reflect"
)

// MaxNestingDepth is the maximum recursion depth honored by
// [ReflectFormFields]. Depth 0 is the top-level struct; depth 1 is
// inside one form:"section=..." or form:"nested" tag. Anything deeper
// panics at walk time so the misconfiguration surfaces immediately
// rather than rendering a half-baked form.
const MaxNestingDepth = 1

// ReflectFormFields returns an iterator over the form-relevant fields
// of target (a struct or pointer-to-struct). Each yielded [FieldVisit]
// carries the hyphen-separated path, the resolved [reflect.StructField],
// the addressable value, the [FieldKind] picked by [DetectField], and
// the enclosing section name (empty at the root).
//
// Walk rules:
//
//   - Anonymous embedded struct fields inline transparently — their
//     visible fields appear at the enclosing level (Go's
//     [reflect.VisibleFields] behavior).
//   - Named struct fields recurse only when tagged form:"section=..."
//     or form:"nested" (FieldKindSection / FieldKindNested). Max
//     recursion depth is [MaxNestingDepth]; deeper nesting panics.
//   - Struct fields registered with [SingleValueTypes] (or implementing
//     [FormWidgetHint]) render as a single widget and are NEVER
//     recursed into — the guard against turning time.Time into a
//     five-field section.
//   - Fields tagged form:"-" or form:"widget=skip" are omitted.
//
// The optional decider parameter is accepted for forward compatibility
// (a future API may want to consult the decider during the walk) but
// is ignored in v1 — recursion and detection are driven entirely by
// [FormTag] and [DetectField].
func ReflectFormFields(target any, decider ...FieldDecider) iter.Seq[FieldVisit] {
	_ = decider // reserved
	return func(yield func(FieldVisit) bool) {
		v := reflect.ValueOf(target)
		for v.Kind() == reflect.Pointer {
			if v.IsNil() {
				panic("ReflectFormFields: nil pointer to " + v.Type().String())
			}
			v = v.Elem()
		}
		if v.Kind() != reflect.Struct {
			panic("ReflectFormFields: need struct or pointer to struct, but got: " + reflect.TypeOf(target).String())
		}
		walkFormFields(v, "", "", 0, yield)
	}
}

// walkFormFields recurses through v yielding leaf field visits. It
// returns false when yield asks the iteration to stop, propagating up.
func walkFormFields(v reflect.Value, prefix FieldPath, section string, depth int, yield func(FieldVisit) bool) bool {
	for _, field := range reflect.VisibleFields(v.Type()) {
		if !field.IsExported() {
			continue
		}
		// Skip the anonymous embed field itself; VisibleFields
		// promotes its visible fields to the enclosing level so they
		// are visited separately at the right path.
		if field.Anonymous {
			continue
		}
		fv, err := v.FieldByIndexErr(field.Index)
		if err != nil {
			// nil embedded pointer-to-struct; subtree unreachable.
			continue
		}
		path := prefix.Append(field.Name)
		kind, tag := DetectField(path, field, fv)

		switch kind {
		case FieldKindSkip:
			continue

		case FieldKindSection, FieldKindNested:
			if depth >= MaxNestingDepth {
				panic("ReflectFormFields: nesting depth exceeds MaxNestingDepth at " + string(path))
			}
			sub := fv
			for sub.Kind() == reflect.Pointer {
				if sub.IsNil() {
					// Skip nil pointer-to-struct sections silently.
					return true
				}
				sub = sub.Elem()
			}
			if sub.Kind() != reflect.Struct {
				continue
			}
			next := section
			if tag.Section != "" {
				next = tag.Section
			} else if tag.Nested {
				if tag.Label != "" {
					next = tag.Label
				} else {
					next = field.Name
				}
			}
			if !walkFormFields(sub, path, next, depth+1, yield) {
				return false
			}

		case FieldKindInline:
			// Should not occur in practice because we skip f.Anonymous
			// above, but guard against direct DetectField use.
			continue

		default:
			if !yield(FieldVisit{
				Path:    path,
				Field:   field,
				Value:   fv,
				Kind:    kind,
				Tag:     tag,
				Section: section,
			}) {
				return false
			}
		}
	}
	return true
}
