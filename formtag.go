package mx

import (
	"reflect"
	"strconv"
	"strings"
)

// FormTag is the parsed result of a form:"..." struct tag.
// All fields are zero when the corresponding key is absent.
type FormTag struct {
	// Widget overrides the type-based widget choice. Recognized
	// values: text, email, url, tel, date, datepicker, datetime, time,
	// password, textarea, select, radio, switch, checkbox, file, hidden,
	// skip. datepicker is a date that a renderer with a date picker, such
	// as shadcn's, shows as one; to the others it is date.
	// "" means "use type-based default."
	Widget string

	// Label overrides the auto-generated label (default is the field
	// name).
	Label string

	// Placeholder is rendered into the input's placeholder attribute.
	Placeholder string

	// Help is rendered as small help text under the field.
	Help string

	// Section assigns the field to a named section, grouping every
	// field with the same Section name into one card.
	Section string

	// Pattern is a regex constraint applied to string-shaped inputs: the
	// input's pattern attribute, and checked by html.FieldDecider's parser
	// too, anchored to the whole text as the attribute is. An empty text is
	// not checked (see Required), nor is a pattern Go's RE2 cannot compile.
	Pattern string

	// Options names an options provider registered with
	// [RegisterNamedOptions] — for field types that do not (or cannot)
	// implement an options-provider interface or the
	// Enums()/EnumStrings() convention themselves (foreign ID types,
	// plain strings). A non-empty Options implies a select widget
	// (enum-set for set-shaped fields) unless Widget says otherwise;
	// the option list is resolved at render time with the request
	// context via [CollectOptions].
	Options string

	// Accept is a file input's accept attribute (form:"accept=…"), such as
	// "image/*", the file types the picker offers. A list holds commas, so
	// it is quoted: form:"accept='.pdf,.docx'". It limits the picker only;
	// check an upload's type on the server.
	Accept string

	// MinLength and MaxLength bound the length of a text, textarea or
	// catch-all field's text (form:"minlength=N,maxlength=N"): the input's
	// minlength and maxlength attributes, and checked by html.FieldDecider's
	// parser too ([CheckTextLength]), counted in UTF-16 code units as the
	// browser counts them. An empty text is not checked (see Required). 0
	// means no bound.
	MinLength, MaxLength int

	// Autocomplete is the input's autocomplete attribute
	// (form:"autocomplete=…"), such as "email", "new-password" or
	// "one-time-code", which tells a browser or password manager what to
	// fill in. Several tokens are separated by spaces.
	Autocomplete string

	// InputMode is the input's inputmode attribute (form:"inputmode=…"),
	// the virtual keyboard to show: none, text, tel, email, url, numeric,
	// decimal or search. A renderer defers an error for another value.
	InputMode string

	// Min, Max, Step are numeric / range constraints. Empty when
	// absent. On a date or time field the finest step is a millisecond,
	// the finest fraction an input's value holds.
	Min, Max, Step string

	// Required forces the value to be non-empty / not-null at submit
	// time. Type-aware: a number rejects a blank submission, a bool is
	// never empty.
	Required bool

	// Sensitive suppresses value round-tripping (passwords, secrets):
	// the input never echoes its value on re-render, and a blank
	// submission keeps the stored value, as with a password field. A
	// widget=password field is sensitive without the option.
	Sensitive bool

	// Nested asks the walker to recurse into a named struct field
	// (max depth 1 in v1). Inferred true when Section is non-empty.
	Nested bool

	// Readonly renders the field but never writes a submitted value
	// into it. Useful for visible identity columns / timestamps.
	Readonly bool

	// Hidden round-trips the GET value through a hidden input but
	// never displays it. Useful for IDs and version columns.
	Hidden bool

	// Key marks a record's identity, such as the ID column of a
	// repeatable row. Like Hidden it renders as a hidden input, which
	// form:"key" implies, but like Readonly a submitted value is never
	// written into it: in a repeatable row the submitted key names the
	// loaded row the submitted row edits, and a row whose key names none
	// is a new row with a zero key. See repeatable.go.
	Key bool

	// Skip means do not render or parse this field at all
	// (form:"-" or form:"widget=skip").
	Skip bool

	// Repeatable renders a slice-of-struct field ([]T or []*T) as a
	// dynamic list of rows and binds the submitted rows back into the
	// slice. On a slice of times it skips the field, whose rows would
	// have no cells; it is a no-op on any other field shape. See
	// repeatable.go.
	Repeatable bool

	// MaxRows caps the rows of a repeatable field (form:"maxrows=N"): the
	// add-row button is hidden at the cap, and a POST with more rows is
	// rejected before any row is bound. 0 means [DefaultMaxRows].
	MaxRows int
}

// ParseFormTag parses the form:"..." tag from a struct field. The
// special value "-" means skip. An empty tag returns the zero
// [FormTag].
func ParseFormTag(field reflect.StructField) FormTag {
	raw := field.Tag.Get("form")
	return ParseFormTagString(raw)
}

// ParseFormTagString parses a raw form:"..." tag value (without the
// surrounding form:"" wrapper). Useful in tests and for callers that
// already have the string. Values are comma-separated; keys with no
// equals sign are boolean flags.
func ParseFormTagString(raw string) FormTag {
	tag := FormTag{}
	if raw == "-" {
		tag.Skip = true
		return tag
	}
	if raw == "" {
		return tag
	}
	for _, part := range splitFormTag(raw) {
		key, val, hasEq := strings.Cut(part, "=")
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		switch key {
		case "":
			continue
		case "widget":
			tag.Widget = val
			if val == "skip" {
				tag.Skip = true
			}
		case "label":
			tag.Label = val
		case "placeholder":
			tag.Placeholder = val
		case "help":
			tag.Help = val
		case "section":
			tag.Section = val
			tag.Nested = true
		case "pattern":
			tag.Pattern = val
		case "options":
			tag.Options = val
		case "min":
			tag.Min = val
		case "max":
			tag.Max = val
		case "step":
			tag.Step = val
		case "minlength":
			if n, err := strconv.Atoi(val); err == nil && n > 0 {
				tag.MinLength = n
			}
		case "maxlength":
			if n, err := strconv.Atoi(val); err == nil && n > 0 {
				tag.MaxLength = n
			}
		case "accept":
			tag.Accept = val
		case "autocomplete":
			tag.Autocomplete = val
		case "inputmode":
			tag.InputMode = val
		case "required":
			tag.Required = !hasEq || boolish(val, true)
		case "sensitive":
			tag.Sensitive = !hasEq || boolish(val, true)
		case "nested":
			tag.Nested = !hasEq || boolish(val, true)
		case "readonly":
			tag.Readonly = !hasEq || boolish(val, true)
		case "hidden":
			tag.Hidden = !hasEq || boolish(val, true)
		case "key":
			tag.Key = !hasEq || boolish(val, true)
			tag.Hidden = tag.Hidden || tag.Key
		case "repeatable":
			tag.Repeatable = !hasEq || boolish(val, true)
		case "maxrows":
			if n, err := strconv.Atoi(val); err == nil && n > 0 {
				tag.MaxRows = n
			}
		}
	}
	return tag
}

// FieldLabel returns the label of a reflected form field: its
// form:"label=…", or the Go field name.
func FieldLabel(field reflect.StructField, tag FormTag) string {
	if tag.Label != "" {
		return tag.Label
	}
	return field.Name
}

// StringInputType returns the input type of a [FieldKindString] field:
// the tag's widget when it is email, url, tel or password, else text.
// [DetectField] has already put a [FormWidgetHint]'s widget into the tag,
// so the tag is the only place to look.
func StringInputType(tag FormTag) string {
	switch tag.Widget {
	case "email", "url", "tel", "password":
		return tag.Widget
	}
	return "text"
}

// splitFormTag splits raw at commas that are NOT inside a value-quoted
// substring. The grammar is intentionally simple: a value introduced
// by '=' may be wrapped in '...' or "..." to embed commas (useful for
// help text and labels). Quotes are stripped from the returned value.
// Only a quote that starts a value opens one: an apostrophe inside a
// value, as in label=Owner's email, is a literal and never swallows the
// options after it.
func splitFormTag(raw string) []string {
	var (
		parts   []string
		current strings.Builder
		quote   byte
	)
	for i := 0; i < len(raw); i++ {
		c := raw[i]
		switch {
		case quote != 0:
			if c == quote {
				quote = 0
				continue
			}
			current.WriteByte(c)
		case (c == '\'' || c == '"') && strings.HasSuffix(strings.TrimRight(current.String(), " "), "="):
			quote = c
		case c == ',':
			parts = append(parts, current.String())
			current.Reset()
		default:
			current.WriteByte(c)
		}
	}
	if current.Len() > 0 || len(parts) == 0 {
		parts = append(parts, current.String())
	}
	return parts
}

// boolish reports whether v reads as true. The empty string returns
// def (so a bare "required" maps to true). Anything else is false
// unless it equals one of "1", "true", "yes", "on" (case-insensitive).
func boolish(v string, def bool) bool {
	if v == "" {
		return def
	}
	switch strings.ToLower(v) {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}
