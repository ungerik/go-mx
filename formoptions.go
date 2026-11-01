package mx

import (
	"context"
	"reflect"
	"slices"
)

// OptionPreface is the extra control a reflected option field renders
// ahead of its option list. The zero value, [OptionPrefaceNone], renders
// nothing.
type OptionPreface int

const (
	// OptionPrefaceNone renders nothing before the options.
	OptionPrefaceNone OptionPreface = iota

	// OptionPrefaceUnavailable is a disabled, selected, empty option for a
	// select whose current value is absent from a context-dependent list.
	// Callers label it with a generic caption; the stored value must not
	// be echoed.
	OptionPrefaceUnavailable

	// OptionPrefaceChoose is a disabled, selected, empty option for a
	// required select with no current value. Placed first, it keeps the
	// browser from submitting the first real option and lets the required
	// attribute fire.
	OptionPrefaceChoose

	// OptionPrefaceEmpty is a selected empty option for an optional select
	// with no current value, so a submit leaves the field empty.
	OptionPrefaceEmpty

	// OptionPrefaceClear is an unselected empty option on an optional
	// select that already has a value the list offers, so that value can
	// be cleared.
	OptionPrefaceClear

	// OptionPrefaceValuesUnavailable is a note, not an option. An enum set
	// holds members a context-dependent list does not show. Those members
	// are not echoed, and a save does not keep them: the list is the trust
	// boundary, the same way a select clears a value the list no longer offers.
	OptionPrefaceValuesUnavailable
)

// RemovesValue reports whether saving the field drops the stored value
// that p says the option list does not offer, which renderers tell the
// user: an optional select's [OptionPrefaceUnavailable] submits nothing,
// so the value is cleared, and the checkboxes of
// [OptionPrefaceValuesUnavailable] are the whole set saved. A required
// select rejects the blank instead, and a readonly field is never saved.
func (p OptionPreface) RemovesValue(tag FormTag) bool {
	switch p {
	case OptionPrefaceUnavailable:
		return !tag.Required && !tag.Readonly
	case OptionPrefaceValuesUnavailable:
		return !tag.Readonly
	}
	return false
}

// SelectPreface decides what a single-value select renders before its
// options.
//
// contextList is the [OptionsNeedContext] answer for the field. An empty
// answer from a provider is a legitimate "nothing to offer"
// ([NamedOptionsContextProvider]); the select still renders, with the
// placeholder or empty option below, so the rest of the form stays usable.
// A provider that failed returns its own error from [CollectOptions].
//
// A stored value the list does not offer yields [OptionPrefaceUnavailable]
// for a context-dependent list and for a required field, including a
// sensitive one. The placeholder is generic and does not echo the value,
// and it is selected so the browser does not submit the first real option.
// Saving an optional one clears the value, which renderers say below the
// select ([OptionPreface.RemovesValue]); a required one rejects the blank
// instead. An optional static list stays lenient and offers
// [OptionPrefaceClear].
//
// When the list has no empty-valued option, a select with no current value
// yields [OptionPrefaceChoose] (required, unless readonly) or
// [OptionPrefaceEmpty] (optional or readonly). An optional select that already has a listed value yields
// [OptionPrefaceClear]. A list that already contains an empty option is
// left alone.
//
// selected is the text of value, the field's stored value. shown is the
// option value the select marks selected: selected itself, or "" for the
// zero value when the list offers no option for it. An enum whose zero is
// none of its options, such as an iota+1 constant or an int64 key from a
// list, holds no choice at its zero: it shows the leading "(none)" or
// "Choose…" option, as an empty select stores it, not an unavailable value
// that saving removes.
func SelectPreface(value reflect.Value, selected string, tag FormTag, contextList bool, options []NamedOption) (preface OptionPreface, shown string) {
	hasEmpty, inList := optionListFacts(selected, options)
	if !inList && value.IsValid() && value.IsZero() {
		selected = ""
	}
	// A select with no selected option submits its first option. Any stored
	// value the list does not offer, including a sensitive one and a value
	// whose list came back empty, gets the generic disabled placeholder so
	// that submit does not invent a choice. The placeholder does not echo
	// the value.
	if selected != "" && !inList && (contextList || tag.Required) {
		return OptionPrefaceUnavailable, selected
	}
	if hasEmpty {
		return OptionPrefaceNone, selected
	}
	if selected == "" {
		// A readonly select is never submitted, so it prompts no choice.
		if tag.Required && !tag.Readonly {
			return OptionPrefaceChoose, selected
		}
		return OptionPrefaceEmpty, selected
	}
	if !tag.Required && (inList || !contextList) {
		return OptionPrefaceClear, selected
	}
	return OptionPrefaceNone, selected
}

// EnumSetPreface decides the note an enum-set checkbox group renders
// before its options.
//
// Members of a context-dependent list that are missing from it, including
// when the list is empty, yield [OptionPrefaceValuesUnavailable], unless
// the field is sensitive. The note does not echo the members. Saving
// removes them ([OptionPreface.RemovesValue]): parse does not keep them, as
// the submitted checkboxes replace the set and a value outside the list is
// rejected; a readonly set is never saved and keeps them. Static lists
// yield [OptionPrefaceNone]: they are the whole domain, and parse replaces
// the set from the submission.
func EnumSetPreface(members []string, tag FormTag, contextList bool, options []NamedOption) OptionPreface {
	if !contextList {
		return OptionPrefaceNone
	}
	allowed := optionValueSet(options)
	for _, member := range members {
		if _, ok := allowed[member]; !ok && !tag.Sensitive {
			return OptionPrefaceValuesUnavailable
		}
	}
	return OptionPrefaceNone
}

func optionListFacts(selected string, options []NamedOption) (hasEmpty, inList bool) {
	for _, opt := range options {
		if opt.Value == "" {
			hasEmpty = true
		}
		if selected != "" && opt.Value == selected {
			inList = true
		}
	}
	return hasEmpty, inList
}

func optionValueSet(options []NamedOption) map[string]struct{} {
	allowed := make(map[string]struct{}, len(options))
	for _, opt := range options {
		allowed[opt.Value] = struct{}{}
	}
	return allowed
}

// SetMemberValues returns the members of a set-shaped value (see
// [SetElemType]) in [FormatOptionValue], the form option values use,
// sorted so a renderer looks an option up with [slices.BinarySearch]. A
// nil value or one of another shape has none.
func SetMemberValues(value reflect.Value) []string {
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil
		}
		value = value.Elem()
	}
	var members []string
	switch value.Kind() {
	case reflect.Map:
		members = make([]string, 0, value.Len())
		iter := value.MapRange()
		for iter.Next() {
			members = append(members, FormatOptionValue(iter.Key()))
		}
	case reflect.Slice:
		members = make([]string, 0, value.Len())
		for i := range value.Len() {
			members = append(members, FormatOptionValue(value.Index(i)))
		}
	}
	slices.Sort(members)
	return members
}

// OptionComponents resolves the option list of a field via
// [CollectOptions] and renders each entry through render. A static option
// list is resolved immediately, so build-time tree enrichment (the hx
// layer's attribute injection) still sees the elements; a list that needs
// the request context (form:"options=…" registry entries,
// [NamedOptionsContextProvider] types, see [OptionsNeedContext]) is
// collected when the component renders, where the context passed to
// Render is the real request context. A collection error surfaces at
// render time (deferred-error pattern).
//
// prepend, when non-nil, is called with the resolved option list and its
// result, if non-nil, is rendered before the options. Selects use it for
// the leading empty option ([SelectPreface]); enum sets use it for the
// note on hidden members ([EnumSetPreface]).
func OptionComponents(value reflect.Value, tag FormTag, t reflect.Type, render func(opt NamedOption) Component, prepend func(options []NamedOption) Component) Component {
	assemble := func(options []NamedOption) Components {
		children := make(Components, 0, len(options)+1)
		if prepend != nil {
			if first := prepend(options); first != nil {
				children = append(children, first)
			}
		}
		for _, opt := range options {
			children = append(children, render(opt))
		}
		return children
	}
	if !OptionsNeedContext(tag, t) {
		options, err := CollectOptions(context.Background(), value, tag, t)
		if err != nil {
			return NewErrElement(err)
		}
		return assemble(options)
	}
	return ComponentFunc(func(ctx context.Context, w Writer) error {
		options, err := CollectOptions(ctx, value, tag, t)
		if err != nil {
			return err
		}
		return assemble(options).Render(ctx, w)
	})
}
