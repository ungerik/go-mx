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

	// Step is the step attribute; "" renders none. It is the tag's step, or
	// "1" when the value is not a whole number of minutes after the min: the
	// default step of 60 seconds, counted from min, would reject it.
	Step string

	// Invalid reports that the text input holds stored text the parser
	// rejects, so saving it unchanged fails. Renderers mark the input
	// invalid. A string-kind value's text is always checked by
	// [ParseDateTimeInput]. A non-string type with MarshalText, such as a
	// date struct, has its UnmarshalText read its own text back, so it is
	// invalid only when the tag sets a min or max, which cannot be checked
	// on text the input type cannot show.
	Invalid bool
}

// TextFallback reports whether the spec is the text input for stored text
// that the widget's input type cannot show. Renderers name the format to
// enter there and leave out the date range, which a text input cannot
// enforce (the parser checks it instead).
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
// input; a blank submission keeps the stored value.
//
// A string-kind value holds the date or time as text. That covers string,
// *string and named types with an underlying string, such as a date type.
// Through any pointers:
//
//  1. A Format(string) string method is asked for the text in the input
//     type's layout ([time.DateOnly], 2006-01-02T15:04:05 or
//     [time.TimeOnly]); without one, a MarshalText method gives the text.
//  2. Otherwise text written in one of the input type's own forms is
//     shown. The forms have no zone and carry seconds only when they are
//     not zero: 2006-01-02, 2006-01-02T15:04[:05] and 15:04[:05]. Zero
//     seconds are dropped, as a browser does.
//  3. Empty text or a null value renders no value, and so does a Format
//     answer of "" for stored text that is a zero date. For stored text
//     that is any other valid value, such as posted text kept after a
//     rejection, the input shows that text.
//  4. Any other text, such as an invalid value, another spelling of a
//     valid one or a time with a zone, cannot be shown by the input, and
//     a blank input would clear the stored value on the next save. It
//     renders as type="text" with the stored text, so the user sees it
//     and the parser reports it.
//
// Any other type with a MarshalText method, such as a date struct, is shown
// by its text under the same rules; its zero or null value renders no value,
// and so does a value whose methods panic, such as a nil embedded pointer.
// Every other value renders no value.
func DateTimeInput(value reflect.Value, tag FormTag) DateTimeInputSpec {
	spec := dateTimeInput(value, tag)
	if tag.Step != "" {
		spec.Step = tag.Step
	} else if !spec.TextFallback() {
		// The browser counts the default step of 60 seconds from min, or
		// from the value attribute when there is no valid min.
		tt := timeTexts[spec.Type]
		t, hasValue := tt.parse(spec.Value)
		lowest, hasMin := tt.parse(tag.Min)
		if hasValue && hasMin && t.Sub(lowest)%time.Minute != 0 {
			spec.Step = "1"
		}
	}
	return spec
}

// dateTimeInput is [DateTimeInput] without the step.
func dateTimeInput(value reflect.Value, tag FormTag) DateTimeInputSpec {
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
		spec.Invalid = spec.TextFallback()
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
		spec.Invalid = spec.TextFallback() && (tag.Min != "" || tag.Max != "")
	}
	return spec
}

// DateTimeInputLayout returns the layout, with seconds, of the input type
// the tag's widget renders: [time.DateOnly], 2006-01-02T15:04:05 or
// [time.TimeOnly]. A parser can hand a type's UnmarshalText a submitted
// value in it, since the inputs leave out zero seconds.
func DateTimeInputLayout(tag FormTag) string {
	return timeTexts[dateTimeInputType(tag)].layouts[0]
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
// validation error naming the expected format ([DateTimeFormatError]), and
// so is a value outside form:"min=…,max=…" ([CheckDateTimeRange]).
func ParseDateTimeInput(tag FormTag, raw string) (string, error) {
	if raw == "" {
		return "", nil
	}
	tt := timeTexts[dateTimeInputType(tag)]
	t, ok := tt.parse(raw)
	if !ok {
		return "", DateTimeFormatError(tag, raw)
	}
	if err := CheckDateTimeRange(tag, raw, t); err != nil {
		return "", err
	}
	return tt.shortest(t), nil
}

// DateTimeFormatError is the validation error for raw, a submission that is
// no value of the tag's input type: it names the expected format. A
// form:"sensitive" field's error leaves raw out, so the value is not echoed
// into the page.
func DateTimeFormatError(tag FormTag, raw string) error {
	tt := timeTexts[dateTimeInputType(tag)]
	if tag.Sensitive {
		return fmt.Errorf("not %s in the format %s", tt.noun, tt.format)
	}
	return fmt.Errorf("%q is not %s in the format %s", raw, tt.noun, tt.format)
}

// CheckDateTimeRange returns a validation error when t, parsed from the
// submitted raw text, lies outside the tag's form:"min=…,max=…". The bounds
// are read in the layouts of the tag's input type, as a browser reads the
// input's min and max, and a bound that is no value of that type is
// ignored, as a browser ignores it. The bounds have no zone, so t's wall
// clock is compared, whatever its zone. A time input's range may wrap past
// midnight, such as min=22:00,max=06:00: then a time from min on or up to
// max is in range. A form:"sensitive" field's error leaves raw out.
func CheckDateTimeRange(tag FormTag, raw string, t time.Time) error {
	inputType := dateTimeInputType(tag)
	tt := timeTexts[inputType]
	subject := fmt.Sprintf("%q is ", raw)
	if tag.Sensitive {
		subject = ""
	}
	t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.UTC)
	lowest, hasMin := tt.parse(tag.Min)
	highest, hasMax := tt.parse(tag.Max)
	if inputType == "time" && hasMin && hasMax && highest.Before(lowest) {
		if t.Before(lowest) && t.After(highest) {
			return fmt.Errorf("%slater than %s and earlier than %s", subject, tag.Max, tag.Min)
		}
		return nil
	}
	if hasMin && t.Before(lowest) {
		return fmt.Errorf("%searlier than %s", subject, tag.Min)
	}
	if hasMax && t.After(highest) {
		return fmt.Errorf("%slater than %s", subject, tag.Max)
	}
	return nil
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

// input decides the input for formatted, the text in this type's layout
// that stands for the stored text. Stored text that is itself a value of
// the type shows even when formatted is not.
func (tt timeText) input(inputType, formatted, stored string) (typ, val string) {
	if formatted == "" {
		// Format answers "" for no value. Stored text that is no value
		// itself would otherwise blank the input and clear the stored text
		// on the next save, and a real value, such as posted text kept
		// after a rejection, still shows.
		t, ok := tt.parse(stored)
		switch {
		case !ok && stored != "":
			return "text", stored
		case !ok || t.IsZero():
			return inputType, ""
		}
		return inputType, tt.shortest(t)
	}
	if t, ok := tt.parse(formatted); ok {
		return inputType, tt.shortest(t)
	}
	if t, ok := tt.parse(stored); ok {
		return inputType, tt.shortest(t)
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
	if f, ok := formatterOf(v); ok {
		formatted = f.Format(tt.layouts[0])
	} else if text, ok := marshalText(v); ok {
		formatted = text
	}
	return tt.input(inputType, formatted, stored)
}

// formatterOf returns the Format method of v, or of its address.
func formatterOf(v reflect.Value) (dateFormatter, bool) {
	if f, ok := safeInterface(v).(dateFormatter); ok {
		return f, true
	}
	if v.CanAddr() {
		f, ok := v.Addr().Interface().(dateFormatter)
		return f, ok
	}
	return nil, false
}

// marshaledText returns the text of a dereferenced value that is neither
// text nor a time.Time but has a MarshalText method, such as a date struct.
// A null or zero value has none, and neither has a value whose methods
// panic, such as a nil embedded pointer.
func marshaledText(v reflect.Value) (text string, ok bool) {
	defer func() {
		if recover() != nil {
			text, ok = "", false
		}
	}()
	iface := safeInterface(v)
	if iface == nil || IsNull(iface) {
		return "", false
	}
	if z, isZeroable := iface.(Zeroable); isZeroable {
		if z.IsZero() {
			return "", false
		}
	} else if v.IsZero() {
		return "", false
	}
	return marshalText(v)
}

// marshalText calls the MarshalText method of v, or of its address.
func marshalText(v reflect.Value) (string, bool) {
	m, ok := safeInterface(v).(encoding.TextMarshaler)
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
