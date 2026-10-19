package mx

import (
	"encoding"
	"fmt"
	"reflect"
	"time"
)

// DateTimeInputSpec is the <input> that [DateTimeInput] decides for a
// reflected [FieldKindDateTime] field.
type DateTimeInputSpec struct {
	// Type is the type attribute: "date", "datetime-local" or "time", or
	// "text" for stored text that the widget's input type cannot show.
	Type string

	// Value is the value attribute; "" renders none.
	Value string

	// Format is how a value of the widget's input type is written:
	// "YYYY-MM-DD", "YYYY-MM-DDTHH:MM" or "HH:MM". Renderers show it as
	// the format to enter on a "text" input.
	Format string
}

// TextFallback reports whether the spec is the text input for stored text
// that the widget's input type cannot show. Renderers name the format to
// enter there, mark the input invalid and leave out the date range, which a
// text input cannot enforce ([ParseDateTimeInput] checks it instead).
func (s DateTimeInputSpec) TextFallback() bool {
	return s.Type == "text"
}

// DateTimeInput decides the <input> that a reflected [FieldKindDateTime]
// field renders. The html and shadcn renderers both call it, so a stored
// value shows the same in either, and [ParseDateTimeInput] reads back
// what the input submits.
//
// form:"widget=date" renders type="date", form:"widget=time" type="time",
// and any other widget type="datetime-local". A [FormWidgetHint] counts
// as the widget, because [DetectField] puts it into the tag. A time.Time
// is formatted for that type, and the zero time renders no value. So is a
// type that wraps time.Time: a type defined as time.Time, or one whose
// MarshalText is RFC 3339, such as a struct embedding time.Time. A
// form:"sensitive" field never shows its value, like a sensitive text
// input, so an unchanged save clears it.
//
// A string-kind value holds the date or time as text. That covers string,
// *string and named types with an underlying string, such as a date type.
// Through any pointers:
//
//  1. A Format(string) string method is asked for the text in the input
//     type's layout ([time.DateOnly], 2006-01-02T15:04:05 or
//     [time.TimeOnly]).
//  2. Otherwise text written in one of the input type's own forms is
//     shown. The forms have no zone and carry seconds only when they are
//     not zero: 2006-01-02, 2006-01-02T15:04[:05] and 15:04[:05]. Zero
//     seconds are dropped, as a browser does.
//  3. Empty text or a null value renders no value, and so does a Format
//     answer of "" for stored text that is a valid value, such as a zero
//     date.
//  4. Any other text, such as an invalid value, another spelling of a
//     valid one or a time with a zone, cannot be shown by the input, and
//     a blank input would clear the stored value on the next save. It
//     renders as type="text" with the stored text, so the user sees it
//     and the parser reports it.
//
// Any other type with a MarshalText method, such as a date struct, is shown
// by its text under the same rules; its zero or null value renders no value.
// Every other value renders no value.
func DateTimeInput(value reflect.Value, tag FormTag) DateTimeInputSpec {
	inputType := dateTimeInputType(tag)
	tt := timeTexts[inputType]
	spec := DateTimeInputSpec{Type: inputType, Format: tt.format}
	if tag.Sensitive || !value.IsValid() {
		return spec
	}
	v := value
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return spec
		}
		v = v.Elem()
	}
	if v.Kind() == reflect.String {
		spec.Type, spec.Value = timeTextInput(v, inputType)
		return spec
	}
	if t, ok := timeOf(v); ok {
		if !t.IsZero() {
			spec.Value = t.Format(tt.layouts[0])
		}
		return spec
	}
	if text, ok := marshaledText(v); ok {
		// RFC 3339 is the text of an embedded time.Time.
		if t, err := time.Parse(time.RFC3339Nano, text); err == nil {
			spec.Value = t.Format(tt.layouts[0])
			return spec
		}
		spec.Type, spec.Value = tt.input(inputType, text, text)
	}
	return spec
}

// timeOf returns the time a dereferenced value holds when it is a time.Time
// or of a type defined as time.Time.
func timeOf(v reflect.Value) (time.Time, bool) {
	if t, ok := safeInterface(v).(time.Time); ok {
		return t, true
	}
	timeType := reflect.TypeFor[time.Time]()
	if v.Kind() == reflect.Struct && v.Type().ConvertibleTo(timeType) {
		return v.Convert(timeType).Interface().(time.Time), true
	}
	return time.Time{}, false
}

// ParseDateTimeInput reads what a date, datetime-local or time input
// submitted for a field that holds its value as text, and returns the
// text to store: the input type's own form that [DateTimeInput] shows,
// with zero seconds dropped. Empty input returns "". Anything else is a
// validation error naming the expected format, and so is a value outside
// form:"min=…,max=…", because the text fallback cannot enforce the range in
// the browser. A bound that is no value of the input type is ignored, as a
// browser ignores it.
func ParseDateTimeInput(tag FormTag, raw string) (string, error) {
	if raw == "" {
		return "", nil
	}
	tt := timeTexts[dateTimeInputType(tag)]
	t, ok := tt.parse(raw)
	if !ok {
		return "", fmt.Errorf("%q is not %s in the format %s", raw, tt.noun, tt.format)
	}
	if lowest, ok := tt.parse(tag.Min); ok && t.Before(lowest) {
		return "", fmt.Errorf("%q is earlier than %s", raw, tag.Min)
	}
	if highest, ok := tt.parse(tag.Max); ok && t.After(highest) {
		return "", fmt.Errorf("%q is later than %s", raw, tag.Max)
	}
	return tt.shortest(t), nil
}

// dateTimeInputType maps a date/time widget to its input type.
func dateTimeInputType(tag FormTag) string {
	switch tag.Widget {
	case "date":
		return "date"
	case "time":
		return "time"
	}
	return "datetime-local"
}

// timeText is the text form of one input type: the layouts it is written
// in, longest first, plus the format and the noun for messages.
type timeText struct {
	layouts []string
	format  string
	noun    string
}

var timeTexts = map[string]timeText{
	"date":           {[]string{time.DateOnly}, "YYYY-MM-DD", "a date"},
	"datetime-local": {[]string{"2006-01-02T15:04:05", "2006-01-02T15:04"}, "YYYY-MM-DDTHH:MM", "a date and time"},
	"time":           {[]string{time.TimeOnly, "15:04"}, "HH:MM", "a time"},
}

// parse returns the time text stands for, when text is written exactly in
// one of the layouts. Exactly means formatting the parsed time gives text
// back, which rules out fractional seconds that time.Parse would otherwise
// accept.
func (tt timeText) parse(text string) (time.Time, bool) {
	for _, layout := range tt.layouts {
		t, err := time.Parse(layout, text)
		if err == nil && t.Format(layout) == text {
			return t, true
		}
	}
	return time.Time{}, false
}

// shortest writes t in the shortest layout that holds it, without seconds
// when they are zero.
func (tt timeText) shortest(t time.Time) string {
	if t.Second() != 0 {
		return t.Format(tt.layouts[0])
	}
	return t.Format(tt.layouts[len(tt.layouts)-1])
}

// normalize returns text in the shortest layout that holds it, when text
// is written exactly in one of the layouts.
func (tt timeText) normalize(text string) (string, bool) {
	t, ok := tt.parse(text)
	if !ok {
		return "", false
	}
	return tt.shortest(t), true
}

// input decides the input for formatted, the text in this type's layout
// that stands for the stored text.
func (tt timeText) input(inputType, formatted, stored string) (typ, val string) {
	if formatted == "" {
		// An empty answer for stored text that is no value itself would
		// blank the input and clear the stored text on the next save.
		if _, ok := tt.normalize(stored); stored != "" && !ok {
			return "text", stored
		}
		return inputType, ""
	}
	if normalized, ok := tt.normalize(formatted); ok {
		return inputType, normalized
	}
	return "text", stored
}

// dateFormatter is the Format method of time.Time, which date and time
// types with an underlying string mirror.
type dateFormatter interface {
	Format(layout string) string
}

// timeTextInput is [DateTimeInput] for a value held as text in v, a
// dereferenced string-kind value.
func timeTextInput(v reflect.Value, inputType string) (typ, val string) {
	if IsNull(safeInterface(v)) {
		return inputType, ""
	}
	tt := timeTexts[inputType]
	stored := v.String()
	formatted := stored
	if f, ok := safeInterface(v).(dateFormatter); ok {
		formatted = f.Format(tt.layouts[0])
	} else if v.CanAddr() {
		if f, ok := v.Addr().Interface().(dateFormatter); ok {
			formatted = f.Format(tt.layouts[0])
		}
	}
	return tt.input(inputType, formatted, stored)
}

// marshaledText returns the text of a dereferenced value that is neither
// text nor a time.Time but has a MarshalText method, such as a date struct.
// A null or zero value has none.
func marshaledText(v reflect.Value) (string, bool) {
	iface := safeInterface(v)
	if iface == nil || IsNull(iface) {
		return "", false
	}
	if z, ok := iface.(Zeroable); ok {
		if z.IsZero() {
			return "", false
		}
	} else if v.IsZero() {
		return "", false
	}
	m, ok := iface.(encoding.TextMarshaler)
	if !ok && v.CanAddr() {
		m, ok = v.Addr().Interface().(encoding.TextMarshaler)
	}
	if !ok {
		return "", false
	}
	text, err := m.MarshalText()
	if err != nil {
		return "", false
	}
	return string(text), true
}
