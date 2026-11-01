package mx

import (
	"fmt"
	"math/big"
	"reflect"
	"regexp"
	"strconv"
)

// CheckNumberRange returns a validation error when value, a number just
// bound from the posted text raw, lies outside the tag's form:"min=…,max=…",
// the min and max of its number input, so a scripted POST is held to them
// too. The bound
// value is compared, never raw itself, so whatever text the parser read is
// checked as it is stored, including the 0 a blank stores in a number that
// is no pointer (a nil pointer is no number to check). An integer is
// compared exactly, so an int64 bound above 2^53 holds. A bound that is no
// valid HTML floating-point number is ignored, as a browser ignores it.
func CheckNumberRange(value reflect.Value, raw string, tag FormTag) error {
	if tag.Min == "" && tag.Max == "" {
		return nil
	}
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil
		}
		value = value.Elem()
	}
	if raw == "" {
		raw = FormValueText(value) // the 0 the re-render shows
	}
	n := new(big.Rat)
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n.SetInt64(value.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n.SetUint64(value.Uint())
	case reflect.Float32, reflect.Float64:
		return checkFloatRange(value.Float(), value.Type().Bits(), raw, tag)
	default:
		return nil
	}
	if lowest, ok := numberBound(tag.Min); ok && n.Cmp(lowest) < 0 {
		return rangeError(raw, tag, true)
	}
	if highest, ok := numberBound(tag.Max); ok && n.Cmp(highest) > 0 {
		return rangeError(raw, tag, false)
	}
	return nil
}

// rangeError is the error of a number below (or above) its form:"min=…" (or
// "max=…"). A form:"sensitive" field's names only the bound, as an
// [FormMessage], which the masking of its errors lets through.
func rangeError(raw string, tag FormTag, below bool) error {
	switch {
	case tag.Sensitive && below:
		return FormMessage("must be at least " + tag.Min)
	case tag.Sensitive:
		return FormMessage("must be at most " + tag.Max)
	case below:
		return fmt.Errorf("%q is less than %s", raw, tag.Min)
	}
	return fmt.Errorf("%q is greater than %s", raw, tag.Max)
}

// checkFloatRange compares a float at its own precision, as a browser
// compares doubles: compared exactly, the binary value of 99.9 is above
// the decimal 99.9, so a value on its bound would be rejected. f is finite,
// as the parsers reject the rest.
func checkFloatRange(f float64, bits int, raw string, tag FormTag) error {
	if lowest, ok := floatBound(tag.Min, bits); ok && f < lowest {
		return rangeError(raw, tag, true)
	}
	if highest, ok := floatBound(tag.Max, bits); ok && f > highest {
		return rangeError(raw, tag, false)
	}
	return nil
}

// floatBound reads a min or max as a browser does, rounded as the field's
// own values are, to a float of the given bits.
func floatBound(bound string, bits int) (float64, bool) {
	if !htmlNumber.MatchString(bound) {
		return 0, false
	}
	// Like a browser, a bound past the range of a double is no bound.
	if _, err := strconv.ParseFloat(bound, 64); err != nil {
		return 0, false
	}
	f, _ := strconv.ParseFloat(bound, bits) // past a float32's range: an infinity
	return f, true
}

// htmlNumber matches a valid HTML floating-point number, the only form a
// browser reads a min or max in.
var htmlNumber = regexp.MustCompile(`^-?(\d+|\d*\.\d+)([eE][-+]?\d+)?$`)

// numberBound reads a min or max as a browser does.
func numberBound(bound string) (*big.Rat, bool) {
	if !htmlNumber.MatchString(bound) {
		return nil, false
	}
	return new(big.Rat).SetString(bound)
}
