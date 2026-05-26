package mx

import (
	"reflect"
	"strings"
)

// FormTag is the parsed result of a form:"..." struct tag.
// All fields are zero when the corresponding key is absent.
type FormTag struct {
	// Widget overrides the type-based widget choice. Recognized
	// values: text, email, url, tel, date, datetime, time, password,
	// textarea, select, radio, switch, checkbox, file, hidden, skip.
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

	// Pattern is a regex constraint applied to string-shaped inputs.
	Pattern string

	// Options names an OptionsProvider entry — for types that do not
	// implement Enums()/EnumStrings() themselves.
	Options string

	// Min, Max, Step are numeric / range constraints. Empty when
	// absent.
	Min, Max, Step string

	// Required forces the value to be non-empty / not-null at submit
	// time. Type-aware: it is a no-op for bool and numeric kinds.
	Required bool

	// Sensitive suppresses value round-tripping (passwords, secrets):
	// the input never echoes its value on re-render.
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

	// Skip means do not render or parse this field at all
	// (form:"-" or form:"widget=skip").
	Skip bool

	// Repeatable is reserved for slice-of-struct support in a
	// follow-on phase; parsing accepts the key so that callers can
	// already mark fields, but v1 ignores it.
	Repeatable bool
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
		case "repeatable":
			tag.Repeatable = !hasEq || boolish(val, true)
		}
	}
	return tag
}

// splitFormTag splits raw at commas that are NOT inside a value-quoted
// substring. The grammar is intentionally simple: a value introduced
// by '=' may be wrapped in '...' or "..." to embed commas (useful for
// help text and labels). Quotes are stripped from the returned value.
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
		case c == '\'' || c == '"':
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
