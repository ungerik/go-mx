package mx

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
// yields [OptionPrefaceChoose] (required) or [OptionPrefaceEmpty]
// (optional). An optional select that already has a listed value yields
// [OptionPrefaceClear]. A list that already contains an empty option is
// left alone.
func SelectPreface(selected string, tag FormTag, contextList bool, options []NamedOption) OptionPreface {
	hasEmpty, inList := optionListFacts(selected, options)
	// A select with no selected option submits its first option. Any stored
	// value the list does not offer, including a sensitive one and a value
	// whose list came back empty, gets the generic disabled placeholder so
	// that submit does not invent a choice. The placeholder does not echo
	// the value.
	if selected != "" && !inList && (contextList || tag.Required) {
		return OptionPrefaceUnavailable
	}
	if hasEmpty {
		return OptionPrefaceNone
	}
	if selected == "" {
		if tag.Required {
			return OptionPrefaceChoose
		}
		return OptionPrefaceEmpty
	}
	if !tag.Required && (inList || !contextList) {
		return OptionPrefaceClear
	}
	return OptionPrefaceNone
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
