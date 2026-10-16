package mx

import "github.com/domonda/go-errs"

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
	// stay on save ([PreserveUnlistedSetMembers]) and must not be echoed
	// into the note.
	OptionPrefaceValuesUnavailable
)

// SelectPreface decides what a single-value select renders before its
// options.
//
// contextList is the [OptionsNeedContext] answer for the field. An empty
// context-dependent list is a render error when the field has a current
// value or is required: the select would otherwise be a disabled
// placeholder the user cannot submit, or a save that clears the stored
// value. An optional field with no current value still renders; an empty
// answer from a provider is a legitimate "nothing to offer" for that case,
// while a provider that failed returns its own error from [CollectOptions].
//
// A stored value the list does not offer yields [OptionPrefaceUnavailable]
// for a context-dependent list and for a required field, including a
// sensitive one. The placeholder is generic and does not echo the value,
// and it is selected so the browser does not submit the first real option.
// An optional static list stays lenient and offers [OptionPrefaceClear].
//
// When the list has no empty-valued option, a select with no current value
// yields [OptionPrefaceChoose] (required) or [OptionPrefaceEmpty]
// (optional). An optional select that already has a listed value yields
// [OptionPrefaceClear]. A list that already contains an empty option is
// left alone.
func SelectPreface(selected string, tag FormTag, contextList bool, options []NamedOption) (OptionPreface, error) {
	hasEmpty, inList := optionListFacts(selected, options)
	if contextList && len(options) == 0 && (selected != "" || tag.Required) {
		return OptionPrefaceNone, errs.New("option list is empty")
	}
	// A select with no selected option submits its first option. Any stored
	// value the list does not offer, including a sensitive one, gets the
	// generic disabled placeholder so that submit does not invent a choice.
	// The placeholder does not echo the value.
	if selected != "" && !inList && (contextList || tag.Required) {
		return OptionPrefaceUnavailable, nil
	}
	if hasEmpty {
		return OptionPrefaceNone, nil
	}
	if selected == "" {
		if tag.Required {
			return OptionPrefaceChoose, nil
		}
		return OptionPrefaceEmpty, nil
	}
	if !tag.Required && (inList || !contextList) {
		return OptionPrefaceClear, nil
	}
	return OptionPrefaceNone, nil
}

// EnumSetPreface decides the note an enum-set checkbox group renders
// before its options.
//
// A context-dependent list that comes back empty is a render error when
// the set already has members or the field is required — the same dead
// form [SelectPreface] refuses. Members of a non-empty context-dependent
// list that are missing from it yield [OptionPrefaceValuesUnavailable],
// unless the field is sensitive. Static lists yield [OptionPrefaceNone]:
// they are the whole domain, and parse replaces the set from the
// submission.
func EnumSetPreface(members []string, tag FormTag, contextList bool, options []NamedOption) (OptionPreface, error) {
	if !contextList {
		return OptionPrefaceNone, nil
	}
	allowed := optionValueSet(options)
	missing := false
	for _, member := range members {
		if _, ok := allowed[member]; !ok {
			missing = true
			break
		}
	}
	if len(options) == 0 && (missing || tag.Required) {
		return OptionPrefaceNone, errs.New("option list is empty")
	}
	if missing && !tag.Sensitive {
		return OptionPrefaceValuesUnavailable, nil
	}
	return OptionPrefaceNone, nil
}

// PreserveUnlistedSetMembers builds the enum set to store after a submit.
// submitted is the checkbox values the browser sent; callers reject any of
// them that are outside options before calling this, so the list stays
// authoritative for new values. Members already stored and absent from
// options were not rendered, so they are appended — a save must not drop a
// value the user could not see or uncheck. Stored members that are in
// options are kept only when submitted. Unlisted stored members are
// appended in stored order, once each.
func PreserveUnlistedSetMembers(stored, submitted []string, options []NamedOption) []string {
	allowed := optionValueSet(options)
	out := append([]string{}, submitted...)
	seen := make(map[string]struct{}, len(out)+len(stored))
	for _, v := range out {
		seen[v] = struct{}{}
	}
	for _, member := range stored {
		if _, ok := allowed[member]; ok {
			continue
		}
		if _, ok := seen[member]; ok {
			continue
		}
		out = append(out, member)
		seen[member] = struct{}{}
	}
	return out
}

// EmptyOptionListError is the render error for a context-dependent option
// list that [SelectPreface] or [EnumSetPreface] refused. provider is the
// form:"options=…" name, or empty when the list came from a
// [NamedOptionsContextProvider] on the field type.
func EmptyOptionListError(path FieldPath, provider string) error {
	if provider != "" {
		return errs.Errorf("field %s: options provider %q returned no options", path, provider)
	}
	return errs.Errorf("field %s: option list is empty", path)
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
