package mx

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// FormValueText returns the text a reflected form's input shows for value,
// the text the html form parser reads back: "" for an invalid value, a
// null ([IsNull]), a nil pointer, the zero of a fixed-size ID such as
// uuid.UUID and a zero time; a time, also of a type defined as one, in
// RFC 3339 with zone and fraction; else the text of an
// [encoding.TextMarshaler], the driver value of a database/sql type
// ([DriverValueText]), or the plain form of a string, bool or number. A
// []string is one line per element, a []byte its bytes as text. A method
// that panics, as the promoted MarshalText of a nil embedded pointer does,
// shows no value instead of failing the render.
//
// Option lists and set members use [FormatOptionValue] instead.
func FormValueText(value reflect.Value) (text string) {
	defer func() {
		if recover() != nil {
			text = ""
		}
	}()
	if !value.IsValid() {
		return ""
	}
	if value.CanInterface() && IsNull(value.Interface()) {
		return ""
	}
	v := value
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}
	// The zero value of a fixed-size ID is what an empty submission stores,
	// so it is no value, not 00000000-0000-0000-0000-000000000000.
	if v.Kind() == reflect.Array && v.IsZero() {
		return ""
	}
	// A time, also of a type defined as one, is written as RFC 3339, the
	// text its parser reads back, zone and fraction included.
	if t, ok := timeOf(v); ok {
		if t.IsZero() {
			return ""
		}
		return t.Format(time.RFC3339Nano)
	}
	if text, ok := callMarshalText(v); ok {
		return text
	}
	// A database/sql type such as sql.NullInt64 writes the value it holds,
	// which the html parser's Scan reads back.
	if text, ok := DriverValueText(v); ok {
		return text
	}
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, v.Type().Bits())
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.String {
			parts := make([]string, v.Len())
			for i := range v.Len() {
				parts[i] = v.Index(i).String()
			}
			return strings.Join(parts, "\n")
		}
		if v.Type().Elem().Kind() == reflect.Uint8 {
			return string(v.Bytes())
		}
	}
	return fmt.Sprint(v.Interface())
}

// FormValueChecked reports whether a reflected form's checkbox for value
// renders checked: value is a true bool, a pointer to one, or a
// database/sql type holding true, such as sql.NullBool. A bool is read by
// its kind, not by [FormValueText], because a named bool's MarshalText may
// spell true otherwise, as "yes".
func FormValueChecked(value reflect.Value) bool {
	v := value
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return false
		}
		v = v.Elem()
	}
	if v.Kind() == reflect.Bool {
		return v.Bool()
	}
	return FormValueText(value) == "true"
}
