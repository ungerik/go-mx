package mx

import (
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

// DateTimeInput decides the <input> that a reflected [FieldKindDateTime]
// field renders. The html and shadcn renderers both call it, so a stored
// value shows the same in either, and [ParseDateTimeInput] reads back
// what the input submits.
//
// form:"widget=date" renders type="date", form:"widget=time" type="time",
// and any other widget type="datetime-local". A [FormWidgetHint] counts
// as the widget, because [DetectField] puts it into the tag. A time.Time
// is formatted for that type, and the zero time renders no value.
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
//  3. Empty text or a null value renders no value.
//  4. Any other text, such as an invalid value, another spelling of a
//     valid one or a time with a zone, cannot be shown by the input, and
//     a blank input would clear the stored value on the next save. It
//     renders as type="text" with the stored text, so the user sees it
//     and the parser reports it.
//
// Every other value renders no value.
func DateTimeInput(value reflect.Value, tag FormTag) DateTimeInputSpec {
	inputType := dateTimeInputType(tag)
	spec := DateTimeInputSpec{Type: inputType, Format: timeTexts[inputType].format}
	if !value.IsValid() {
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
	t, ok := v.Interface().(time.Time)
	if !ok || t.IsZero() {
		return spec
	}
	switch inputType {
	case "date":
		spec.Value = t.Format("2006-01-02")
	case "time":
		spec.Value = t.Format("15:04:05")
	default:
		spec.Value = t.Format("2006-01-02T15:04:05")
	}
	return spec
}

// ParseDateTimeInput reads what a date, datetime-local or time input
// submitted for a field that holds its value as text, and returns the
// text to store: the input type's own form that [DateTimeInput] shows,
// with zero seconds dropped. Empty input returns "". Anything else is a
// validation error naming the expected format.
func ParseDateTimeInput(tag FormTag, raw string) (string, error) {
	if raw == "" {
		return "", nil
	}
	tt := timeTexts[dateTimeInputType(tag)]
	if text, ok := tt.normalize(raw); ok {
		return text, nil
	}
	return "", fmt.Errorf("%q is not %s in the format %s", raw, tt.noun, tt.format)
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

// normalize returns text in the shortest layout that holds it, when text
// is written exactly in one of the layouts. Exactly means formatting the
// parsed time gives text back, which rules out fractional seconds that
// time.Parse would otherwise accept.
func (tt timeText) normalize(text string) (string, bool) {
	for _, layout := range tt.layouts {
		t, err := time.Parse(layout, text)
		if err != nil || t.Format(layout) != text {
			continue
		}
		short := tt.layouts[len(tt.layouts)-1]
		if t.Second() != 0 {
			short = tt.layouts[0]
		}
		return t.Format(short), true
	}
	return "", false
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
	text := stored
	if f, ok := safeInterface(v).(dateFormatter); ok {
		text = f.Format(tt.layouts[0])
	} else if v.CanAddr() {
		if f, ok := v.Addr().Interface().(dateFormatter); ok {
			text = f.Format(tt.layouts[0])
		}
	}
	if text == "" {
		return inputType, ""
	}
	if normalized, ok := tt.normalize(text); ok {
		return inputType, normalized
	}
	return "text", stored
}
