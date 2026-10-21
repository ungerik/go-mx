package mx

import (
	"iter"
	"reflect"
	"slices"
	"strconv"
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
//   - A nil pointer section is skipped; the fields after it are not.
//   - form:"-" on an embedded struct omits its promoted fields, and
//     form:"readonly" on one makes them readonly, as for a section.
//   - form:"readonly" on a section or nested field makes every field in
//     it readonly: its visits carry the tag in both [FieldVisit.Tag] and
//     [FieldVisit.Field], so a decider detecting the field again sees it.
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
		walkFormFields(v, "", "", 0, false, yield)
	}
}

// walkFormFields recurses through v yielding leaf field visits. readonly
// marks every field of a readonly section. It returns false when yield
// asks the iteration to stop, propagating up.
func walkFormFields(v reflect.Value, prefix FieldPath, section string, depth int, readonly bool, yield func(FieldVisit) bool) bool {
	// The index paths of embedded structs whose form tag omits or locks
	// their promoted fields; VisibleFields lists an embed right before
	// the fields it promotes.
	var skippedEmbeds, readonlyEmbeds [][]int
fields:
	for _, field := range reflect.VisibleFields(v.Type()) {
		if indexUnder(field.Index, skippedEmbeds) {
			continue
		}
		// Skip the anonymous embed field itself; VisibleFields
		// promotes its visible fields to the enclosing level so they
		// are visited separately at the right path. Its form tag still
		// applies to them.
		if field.Anonymous {
			switch tag := ParseFormTagString(field.Tag.Get("form")); {
			case tag.Skip:
				skippedEmbeds = append(skippedEmbeds, field.Index)
			case tag.Readonly:
				readonlyEmbeds = append(readonlyEmbeds, field.Index)
			}
			continue
		}
		if !field.IsExported() {
			continue
		}
		fv, err := v.FieldByIndexErr(field.Index)
		if err != nil {
			// nil embedded pointer-to-struct; subtree unreachable.
			continue
		}
		path := prefix.Append(field.Name)
		// The kind comes from the field's own tag, so a skipped field
		// (form:"-") stays skipped in a readonly section, where its
		// rewritten tag would no longer read as skip.
		kind, tag := DetectField(path, field, fv)
		if readonly || indexUnder(field.Index, readonlyEmbeds) {
			// Deciders detect the field again, from its tag.
			field.Tag = withReadonlyTag(field.Tag)
			tag.Readonly = true
		}

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
					// Skip a nil pointer-to-struct section silently, but
					// not the fields after it.
					continue fields
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
			if !walkFormFields(sub, path, next, depth+1, readonly || tag.Readonly, yield) {
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

// indexUnder reports whether the field index path index lies under one of
// the embed index paths in embeds.
func indexUnder(index []int, embeds [][]int) bool {
	for _, embed := range embeds {
		if len(index) > len(embed) && slices.Equal(index[:len(embed)], embed) {
			return true
		}
	}
	return false
}

// withReadonlyTag returns tag with readonly added to its form tag, first,
// so an unterminated quote in the field's own options cannot swallow it,
// and last, so the field's own readonly=false cannot override it. The new
// form key goes first, where [reflect.StructTag.Lookup] finds it. A
// skipped field's tag (form:"-") stops reading as skip, so the walker
// decides the kind from the original tag.
func withReadonlyTag(tag reflect.StructTag) reflect.StructTag {
	value := "readonly"
	if form, ok := tag.Lookup("form"); ok && form != "" {
		value += "," + form + ",readonly"
	}
	return reflect.StructTag("form:" + strconv.Quote(value) + " " + string(tag))
}
