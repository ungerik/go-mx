package mx

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

// isoDate mimics a date type with an underlying string: Format returns a
// valid date normalized, "" for an empty or zero date, and an invalid
// value unchanged.
type isoDate string

func (d isoDate) Format(layout string) string {
	for _, l := range []string{time.DateOnly, "02.01.2006"} {
		if t, err := time.Parse(l, string(d)); err == nil {
			if t.IsZero() {
				return ""
			}
			return t.Format(layout)
		}
	}
	return string(d)
}

// ptrDate has its Format method on the pointer receiver.
type ptrDate string

func (d *ptrDate) Format(layout string) string { return isoDate(*d).Format(layout) }

// markerDate's Format answers unparseable text with a marker instead of
// returning it unchanged, which tells the stored text and Format's answer
// apart.
type markerDate string

func (d markerDate) Format(layout string) string {
	if t, err := time.Parse(time.DateOnly, string(d)); err == nil {
		return t.Format(layout)
	}
	return "not a date"
}

// blankDate's Format answers unparseable text with "", as a Format that
// treats anything invalid as no date might.
type blankDate string

func (d blankDate) Format(layout string) string {
	if t, err := time.Parse(time.DateOnly, string(d)); err == nil && !t.IsZero() {
		return t.Format(layout)
	}
	return ""
}

// textDate is a date struct with a text form, like civil.Date.
type textDate struct {
	Year, Month, Day int
}

func (d textDate) MarshalText() ([]byte, error) {
	return fmt.Appendf(nil, "%04d-%02d-%02d", d.Year, d.Month, d.Day), nil
}

// ptrTextDate is a date struct with MarshalText on the pointer receiver.
type ptrTextDate struct {
	Year, Month, Day int
}

func (d *ptrTextDate) MarshalText() ([]byte, error) {
	return fmt.Appendf(nil, "%04d-%02d-%02d", d.Year, d.Month, d.Day), nil
}

// zeroableDate decides its zero value itself: 0001-01-01, like time.Time,
// not the struct's reflect zero.
type zeroableDate struct {
	Year, Month, Day int
}

func (d zeroableDate) IsZero() bool { return d == zeroableDate{1, 1, 1} }

func (d zeroableDate) MarshalText() ([]byte, error) {
	return fmt.Appendf(nil, "%04d-%02d-%02d", d.Year, d.Month, d.Day), nil
}

// wrappedTime embeds time.Time and so has its RFC 3339 MarshalText.
type wrappedTime struct {
	time.Time
}

// definedTime is defined as time.Time, without its methods.
type definedTime time.Time

// sentinelDate is null when it holds the text "NULL".
type sentinelDate string

func (d sentinelDate) IsNull() bool { return d == "NULL" }

// clockTime mimics a time-of-day type with an underlying string whose
// Format re-spells a stored "9.05" as 09:05.
type clockTime string

func (c clockTime) Format(layout string) string {
	t, err := time.Parse("15.04", string(c))
	if err != nil {
		return string(c)
	}
	return t.Format(layout)
}

func TestDateTimeInput(t *testing.T) {
	when := time.Date(2026, 7, 6, 9, 5, 7, 0, time.UTC)
	date := FormTag{Widget: "date"}
	datetime := FormTag{Widget: "datetime"}
	clock := FormTag{Widget: "time"}
	tests := []struct {
		name  string
		value reflect.Value
		tag   FormTag
		want  DateTimeInputSpec
	}{
		{"no value", reflect.Value{}, FormTag{}, DateTimeInputSpec{"datetime-local", "", "YYYY-MM-DDTHH:MM"}},
		{"time default widget", reflect.ValueOf(when), FormTag{}, DateTimeInputSpec{"datetime-local", "2026-07-06T09:05:07", "YYYY-MM-DDTHH:MM"}},
		{"time date widget", reflect.ValueOf(when), date, DateTimeInputSpec{"date", "2026-07-06", "YYYY-MM-DD"}},
		{"time time widget", reflect.ValueOf(when), clock, DateTimeInputSpec{"time", "09:05:07", "HH:MM"}},
		{"time keeps zero seconds", reflect.ValueOf(when.Truncate(time.Minute)), clock, DateTimeInputSpec{"time", "09:05:00", "HH:MM"}},
		{"zero time", reflect.ValueOf(time.Time{}), date, DateTimeInputSpec{"date", "", "YYYY-MM-DD"}},
		{"nil time pointer", reflect.ValueOf((*time.Time)(nil)), date, DateTimeInputSpec{"date", "", "YYYY-MM-DD"}},

		// Dates held as text.
		{"Format", reflect.ValueOf(isoDate("06.07.2026")), date, DateTimeInputSpec{"date", "2026-07-06", "YYYY-MM-DD"}},
		{"Format zero date", reflect.ValueOf(isoDate("0001-01-01")), date, DateTimeInputSpec{"date", "", "YYYY-MM-DD"}},
		{"Format returns invalid unchanged", reflect.ValueOf(isoDate("garbage")), date, DateTimeInputSpec{"text", "garbage", "YYYY-MM-DD"}},
		// The text input shows what is stored, not what Format made of it.
		{"Format answers invalid otherwise", reflect.ValueOf(markerDate("garbage")), date, DateTimeInputSpec{"text", "garbage", "YYYY-MM-DD"}},
		// A Format that answers invalid text with "" must not blank it: a
		// blank picker would clear the stored text on the next save. A zero
		// date it answers with "" is still no value.
		{"Format answers invalid with empty", reflect.ValueOf(blankDate("garbage")), date, DateTimeInputSpec{"text", "garbage", "YYYY-MM-DD"}},
		{"Format answers zero date with empty", reflect.ValueOf(blankDate("0001-01-01")), date, DateTimeInputSpec{"date", "", "YYYY-MM-DD"}},
		{"pointer receiver Format", reflect.ValueOf(new(ptrDate("06.07.2026"))), date, DateTimeInputSpec{"date", "2026-07-06", "YYYY-MM-DD"}},
		{"string", reflect.ValueOf("2026-07-06"), date, DateTimeInputSpec{"date", "2026-07-06", "YYYY-MM-DD"}},
		{"string through two pointers", reflect.ValueOf(new(new("2026-07-06"))), date, DateTimeInputSpec{"date", "2026-07-06", "YYYY-MM-DD"}},
		{"empty string", reflect.ValueOf(""), date, DateTimeInputSpec{"date", "", "YYYY-MM-DD"}},
		{"null", reflect.ValueOf(sentinelDate("NULL")), date, DateTimeInputSpec{"date", "", "YYYY-MM-DD"}},
		{"non-null sentinel type", reflect.ValueOf(sentinelDate("2026-07-06")), date, DateTimeInputSpec{"date", "2026-07-06", "YYYY-MM-DD"}},
		{"other spelling", reflect.ValueOf("06.07.2026"), date, DateTimeInputSpec{"text", "06.07.2026", "YYYY-MM-DD"}},
		{"impossible date", reflect.ValueOf("2026-02-30"), date, DateTimeInputSpec{"text", "2026-02-30", "YYYY-MM-DD"}},

		// Dates and times held as text: no zone, seconds only when not zero.
		{"datetime minutes", reflect.ValueOf("2026-07-06T09:05"), datetime, DateTimeInputSpec{"datetime-local", "2026-07-06T09:05", "YYYY-MM-DDTHH:MM"}},
		{"datetime drops zero seconds", reflect.ValueOf("2026-07-06T09:05:00"), datetime, DateTimeInputSpec{"datetime-local", "2026-07-06T09:05", "YYYY-MM-DDTHH:MM"}},
		{"datetime seconds", reflect.ValueOf("2026-07-06T09:05:07"), datetime, DateTimeInputSpec{"datetime-local", "2026-07-06T09:05:07", "YYYY-MM-DDTHH:MM"}},
		{"datetime-local widget", reflect.ValueOf("2026-07-06T09:05"), FormTag{Widget: "datetime-local"}, DateTimeInputSpec{"datetime-local", "2026-07-06T09:05", "YYYY-MM-DDTHH:MM"}},
		{"datetime with zone", reflect.ValueOf("2026-07-06T09:05:07Z"), datetime, DateTimeInputSpec{"text", "2026-07-06T09:05:07Z", "YYYY-MM-DDTHH:MM"}},
		{"datetime with fraction", reflect.ValueOf("2026-07-06T09:05:07.5"), datetime, DateTimeInputSpec{"text", "2026-07-06T09:05:07.5", "YYYY-MM-DDTHH:MM"}},
		{"datetime with space", reflect.ValueOf("2026-07-06 09:05"), datetime, DateTimeInputSpec{"text", "2026-07-06 09:05", "YYYY-MM-DDTHH:MM"}},
		{"time minutes", reflect.ValueOf("09:05"), clock, DateTimeInputSpec{"time", "09:05", "HH:MM"}},
		{"time drops zero seconds", reflect.ValueOf("09:05:00"), clock, DateTimeInputSpec{"time", "09:05", "HH:MM"}},
		{"time seconds", reflect.ValueOf("09:05:07"), clock, DateTimeInputSpec{"time", "09:05:07", "HH:MM"}},
		{"time Format", reflect.ValueOf(clockTime("9.05")), clock, DateTimeInputSpec{"time", "09:05", "HH:MM"}},
		{"time one-digit hour", reflect.ValueOf("9:05"), clock, DateTimeInputSpec{"text", "9:05", "HH:MM"}},
		{"time out of range", reflect.ValueOf("25:00"), clock, DateTimeInputSpec{"text", "25:00", "HH:MM"}},

		// A sensitive field never shows its value, not even as text.
		{"sensitive text", reflect.ValueOf("2026-07-06"), FormTag{Widget: "date", Sensitive: true}, DateTimeInputSpec{"date", "", "YYYY-MM-DD"}},
		{"sensitive other spelling", reflect.ValueOf("06.07.2026"), FormTag{Widget: "date", Sensitive: true}, DateTimeInputSpec{"date", "", "YYYY-MM-DD"}},
		{"sensitive time", reflect.ValueOf(when), FormTag{Widget: "date", Sensitive: true}, DateTimeInputSpec{"date", "", "YYYY-MM-DD"}},

		// A type with MarshalText, such as a date struct, shows its text.
		{"MarshalText", reflect.ValueOf(textDate{2026, 7, 6}), date, DateTimeInputSpec{"date", "2026-07-06", "YYYY-MM-DD"}},
		{"MarshalText through pointer", reflect.ValueOf(&textDate{2026, 7, 6}), date, DateTimeInputSpec{"date", "2026-07-06", "YYYY-MM-DD"}},
		{"MarshalText zero value", reflect.ValueOf(textDate{}), date, DateTimeInputSpec{"date", "", "YYYY-MM-DD"}},
		{"MarshalText other form", reflect.ValueOf(textDate{2026, 7, 6}), datetime, DateTimeInputSpec{"text", "2026-07-06", "YYYY-MM-DDTHH:MM"}},
		{"no MarshalText", reflect.ValueOf(struct{ N int }{1}), date, DateTimeInputSpec{"date", "", "YYYY-MM-DD"}},
		{"MarshalText on the pointer receiver", reflect.ValueOf(&ptrTextDate{2026, 7, 6}), date, DateTimeInputSpec{"date", "2026-07-06", "YYYY-MM-DD"}},
		// The type's own IsZero decides, not the struct's reflect zero.
		{"IsZero date", reflect.ValueOf(zeroableDate{1, 1, 1}), date, DateTimeInputSpec{"date", "", "YYYY-MM-DD"}},
		{"non-zero IsZero date", reflect.ValueOf(zeroableDate{2026, 7, 6}), date, DateTimeInputSpec{"date", "2026-07-06", "YYYY-MM-DD"}},

		// Types that wrap time.Time render like it.
		{"embedded time date", reflect.ValueOf(wrappedTime{when}), date, DateTimeInputSpec{"date", "2026-07-06", "YYYY-MM-DD"}},
		{"embedded time datetime", reflect.ValueOf(wrappedTime{when}), datetime, DateTimeInputSpec{"datetime-local", "2026-07-06T09:05:07", "YYYY-MM-DDTHH:MM"}},
		{"embedded time time", reflect.ValueOf(&wrappedTime{when}), clock, DateTimeInputSpec{"time", "09:05:07", "HH:MM"}},
		{"embedded zero time", reflect.ValueOf(wrappedTime{}), date, DateTimeInputSpec{"date", "", "YYYY-MM-DD"}},
		{"defined time", reflect.ValueOf(definedTime(when)), date, DateTimeInputSpec{"date", "2026-07-06", "YYYY-MM-DD"}},
		{"defined zero time", reflect.ValueOf(definedTime{}), datetime, DateTimeInputSpec{"datetime-local", "", "YYYY-MM-DDTHH:MM"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DateTimeInput(tt.value, tt.tag); got != tt.want {
				t.Errorf("DateTimeInput = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParseDateTimeInput(t *testing.T) {
	date := FormTag{Widget: "date"}
	datetime := FormTag{Widget: "datetime"}
	clock := FormTag{Widget: "time"}
	tests := []struct {
		name    string
		tag     FormTag
		raw     string
		want    string
		wantErr string
	}{
		{name: "empty", tag: date, raw: "", want: ""},
		{name: "date", tag: date, raw: "2026-07-06", want: "2026-07-06"},
		{name: "date other spelling", tag: date, raw: "06.07.2026",
			wantErr: `"06.07.2026" is not a date in the format YYYY-MM-DD`},
		{name: "datetime minutes", tag: datetime, raw: "2026-07-06T09:05", want: "2026-07-06T09:05"},
		{name: "datetime zero seconds", tag: datetime, raw: "2026-07-06T09:05:00", want: "2026-07-06T09:05"},
		{name: "datetime seconds", tag: datetime, raw: "2026-07-06T09:05:07", want: "2026-07-06T09:05:07"},
		{name: "datetime zone", tag: datetime, raw: "2026-07-06T09:05:07Z",
			wantErr: `"2026-07-06T09:05:07Z" is not a date and time in the format YYYY-MM-DDTHH:MM`},
		{name: "datetime fraction", tag: datetime, raw: "2026-07-06T09:05:07.123",
			wantErr: `"2026-07-06T09:05:07.123" is not a date and time in the format YYYY-MM-DDTHH:MM`},
		{name: "time minutes", tag: clock, raw: "09:05", want: "09:05"},
		{name: "time zero seconds", tag: clock, raw: "09:05:00", want: "09:05"},
		{name: "time seconds", tag: clock, raw: "09:05:07", want: "09:05:07"},
		{name: "time one-digit hour", tag: clock, raw: "9:05",
			wantErr: `"9:05" is not a time in the format HH:MM`},

		// The range applies on the server: a text input cannot enforce it.
		{name: "date in range", tag: FormTag{Widget: "date", Min: "2026-01-01", Max: "2026-12-31"}, raw: "2026-07-06", want: "2026-07-06"},
		{name: "date on the bounds", tag: FormTag{Widget: "date", Min: "2026-07-06", Max: "2026-07-06"}, raw: "2026-07-06", want: "2026-07-06"},
		{name: "date before min", tag: FormTag{Widget: "date", Min: "2026-01-01"}, raw: "1900-01-01",
			wantErr: `"1900-01-01" is earlier than 2026-01-01`},
		{name: "date after max", tag: FormTag{Widget: "date", Max: "2026-12-31"}, raw: "2027-01-01",
			wantErr: `"2027-01-01" is later than 2026-12-31`},
		{name: "datetime seconds past max", tag: FormTag{Widget: "datetime", Max: "2026-07-06T09:05"}, raw: "2026-07-06T09:05:01",
			wantErr: `"2026-07-06T09:05:01" is later than 2026-07-06T09:05`},
		{name: "time before min", tag: FormTag{Widget: "time", Min: "08:00"}, raw: "07:59", wantErr: `"07:59" is earlier than 08:00`},
		{name: "invalid bound ignored", tag: FormTag{Widget: "date", Min: "soon"}, raw: "1900-01-01", want: "1900-01-01"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDateTimeInput(tt.tag, tt.raw)
			switch {
			case tt.wantErr == "" && err != nil:
				t.Fatalf("ParseDateTimeInput(%q): %v", tt.raw, err)
			case tt.wantErr != "" && (err == nil || err.Error() != tt.wantErr):
				t.Fatalf("ParseDateTimeInput(%q): error %v, want %s", tt.raw, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ParseDateTimeInput(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestDateTimeInputSpecTextFallback(t *testing.T) {
	if !(DateTimeInputSpec{Type: "text"}).TextFallback() {
		t.Error(`Type "text" is the text fallback`)
	}
	for _, typ := range []string{"date", "datetime-local", "time"} {
		if (DateTimeInputSpec{Type: typ}).TextFallback() {
			t.Errorf("Type %q is no text fallback", typ)
		}
	}
}
