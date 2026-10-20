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

// ptrWrappedTime embeds a pointer to a time; its promoted methods panic on
// a nil one.
type ptrWrappedTime struct {
	*time.Time
}

// compactDate stores 20260706 but writes and reads ISO text, with
// MarshalText and no Format.
type compactDate string

func (d compactDate) MarshalText() ([]byte, error) {
	t, err := time.Parse("20060102", string(d))
	if err != nil {
		return []byte(d), nil
	}
	return []byte(t.Format(time.DateOnly)), nil
}

// dottedDate's Format reads only its own spelling 02.01.2006 and answers
// anything else with "", such as ISO text kept after a rejection.
type dottedDate string

func (d dottedDate) Format(layout string) string {
	if t, err := time.Parse("02.01.2006", string(d)); err == nil {
		return t.Format(layout)
	}
	return ""
}

// civDateTime is a date-and-time struct shaped like civil.DateTime.
type civDateTime struct {
	T time.Time
}

func (c civDateTime) MarshalText() ([]byte, error) {
	return []byte(c.T.Format("2006-01-02T15:04:05")), nil
}

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

// zonedText stores a date but writes it with a time and zone, text no
// date input reads.
type zonedText string

func (d zonedText) MarshalText() ([]byte, error) { return []byte(d + "T00:00:00Z"), nil }

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
		{"no value", reflect.Value{}, FormTag{}, DateTimeInputSpec{"datetime-local", "", "YYYY-MM-DDTHH:MM", "", false, time.Time{}}},
		{"time default widget", reflect.ValueOf(when), FormTag{}, DateTimeInputSpec{"datetime-local", "2026-07-06T09:05:07", "YYYY-MM-DDTHH:MM", "", false, time.Time{}}},
		{"time date widget", reflect.ValueOf(when), date, DateTimeInputSpec{"date", "2026-07-06", "YYYY-MM-DD", "", false, time.Time{}}},
		{"time time widget", reflect.ValueOf(when), clock, DateTimeInputSpec{"time", "09:05:07", "HH:MM", "", false, time.Time{}}},
		{"time keeps zero seconds", reflect.ValueOf(when.Truncate(time.Minute)), clock, DateTimeInputSpec{"time", "09:05:00", "HH:MM", "", false, time.Time{}}},
		{"zero time", reflect.ValueOf(time.Time{}), date, DateTimeInputSpec{"date", "", "YYYY-MM-DD", "", false, time.Time{}}},
		{"nil time pointer", reflect.ValueOf((*time.Time)(nil)), date, DateTimeInputSpec{"date", "", "YYYY-MM-DD", "", false, time.Time{}}},

		// Dates held as text.
		{"Format", reflect.ValueOf(isoDate("06.07.2026")), date, DateTimeInputSpec{"date", "2026-07-06", "YYYY-MM-DD", "", false, time.Time{}}},
		{"Format zero date", reflect.ValueOf(isoDate("0001-01-01")), date, DateTimeInputSpec{"date", "", "YYYY-MM-DD", "", false, time.Time{}}},
		{"Format returns invalid unchanged", reflect.ValueOf(isoDate("garbage")), date, DateTimeInputSpec{"text", "garbage", "YYYY-MM-DD", "", true, time.Time{}}},
		// The text input shows what is stored, not what Format made of it.
		{"Format answers invalid otherwise", reflect.ValueOf(markerDate("garbage")), date, DateTimeInputSpec{"text", "garbage", "YYYY-MM-DD", "", true, time.Time{}}},
		// A Format that answers invalid text with "" must not blank it: a
		// blank picker would clear the stored text on the next save. A zero
		// date it answers with "" is still no value.
		{"Format answers invalid with empty", reflect.ValueOf(blankDate("garbage")), date, DateTimeInputSpec{"text", "garbage", "YYYY-MM-DD", "", true, time.Time{}}},
		{"Format answers zero date with empty", reflect.ValueOf(blankDate("0001-01-01")), date, DateTimeInputSpec{"date", "", "YYYY-MM-DD", "", false, time.Time{}}},
		{"pointer receiver Format", reflect.ValueOf(new(ptrDate("06.07.2026"))), date, DateTimeInputSpec{"date", "2026-07-06", "YYYY-MM-DD", "", false, time.Time{}}},
		{"string", reflect.ValueOf("2026-07-06"), date, DateTimeInputSpec{"date", "2026-07-06", "YYYY-MM-DD", "", false, time.Time{}}},
		{"string through two pointers", reflect.ValueOf(new(new("2026-07-06"))), date, DateTimeInputSpec{"date", "2026-07-06", "YYYY-MM-DD", "", false, time.Time{}}},
		{"empty string", reflect.ValueOf(""), date, DateTimeInputSpec{"date", "", "YYYY-MM-DD", "", false, time.Time{}}},
		{"null", reflect.ValueOf(sentinelDate("NULL")), date, DateTimeInputSpec{"date", "", "YYYY-MM-DD", "", false, time.Time{}}},
		{"non-null sentinel type", reflect.ValueOf(sentinelDate("2026-07-06")), date, DateTimeInputSpec{"date", "2026-07-06", "YYYY-MM-DD", "", false, time.Time{}}},
		{"other spelling", reflect.ValueOf("06.07.2026"), date, DateTimeInputSpec{"text", "06.07.2026", "YYYY-MM-DD", "", true, time.Time{}}},
		{"impossible date", reflect.ValueOf("2026-02-30"), date, DateTimeInputSpec{"text", "2026-02-30", "YYYY-MM-DD", "", true, time.Time{}}},

		// Dates and times held as text: no zone, seconds only when not zero.
		{"datetime minutes", reflect.ValueOf("2026-07-06T09:05"), datetime, DateTimeInputSpec{"datetime-local", "2026-07-06T09:05", "YYYY-MM-DDTHH:MM", "", false, time.Time{}}},
		{"datetime drops zero seconds", reflect.ValueOf("2026-07-06T09:05:00"), datetime, DateTimeInputSpec{"datetime-local", "2026-07-06T09:05", "YYYY-MM-DDTHH:MM", "", false, time.Time{}}},
		{"datetime seconds", reflect.ValueOf("2026-07-06T09:05:07"), datetime, DateTimeInputSpec{"datetime-local", "2026-07-06T09:05:07", "YYYY-MM-DDTHH:MM", "", false, time.Time{}}},
		{"datetime-local widget", reflect.ValueOf("2026-07-06T09:05"), FormTag{Widget: "datetime-local"}, DateTimeInputSpec{"datetime-local", "2026-07-06T09:05", "YYYY-MM-DDTHH:MM", "", false, time.Time{}}},
		{"datetime with zone", reflect.ValueOf("2026-07-06T09:05:07Z"), datetime, DateTimeInputSpec{"text", "2026-07-06T09:05:07Z", "YYYY-MM-DDTHH:MM", "", true, time.Time{}}},
		{"datetime with fraction", reflect.ValueOf("2026-07-06T09:05:07.5"), datetime, DateTimeInputSpec{"text", "2026-07-06T09:05:07.5", "YYYY-MM-DDTHH:MM", "", true, time.Time{}}},
		{"datetime with space", reflect.ValueOf("2026-07-06 09:05"), datetime, DateTimeInputSpec{"text", "2026-07-06 09:05", "YYYY-MM-DDTHH:MM", "", true, time.Time{}}},
		{"time minutes", reflect.ValueOf("09:05"), clock, DateTimeInputSpec{"time", "09:05", "HH:MM", "", false, time.Time{}}},
		{"time drops zero seconds", reflect.ValueOf("09:05:00"), clock, DateTimeInputSpec{"time", "09:05", "HH:MM", "", false, time.Time{}}},
		{"time seconds", reflect.ValueOf("09:05:07"), clock, DateTimeInputSpec{"time", "09:05:07", "HH:MM", "", false, time.Time{}}},
		{"time Format", reflect.ValueOf(clockTime("9.05")), clock, DateTimeInputSpec{"time", "09:05", "HH:MM", "", false, time.Time{}}},
		{"time one-digit hour", reflect.ValueOf("9:05"), clock, DateTimeInputSpec{"text", "9:05", "HH:MM", "", true, time.Time{}}},
		{"time out of range", reflect.ValueOf("25:00"), clock, DateTimeInputSpec{"text", "25:00", "HH:MM", "", true, time.Time{}}},

		// A sensitive field never shows its value, not even as text.
		{"sensitive text", reflect.ValueOf("2026-07-06"), FormTag{Widget: "date", Sensitive: true}, DateTimeInputSpec{"date", "", "YYYY-MM-DD", "", false, time.Time{}}},
		{"sensitive other spelling", reflect.ValueOf("06.07.2026"), FormTag{Widget: "date", Sensitive: true}, DateTimeInputSpec{"date", "", "YYYY-MM-DD", "", false, time.Time{}}},
		{"sensitive time", reflect.ValueOf(when), FormTag{Widget: "date", Sensitive: true}, DateTimeInputSpec{"date", "", "YYYY-MM-DD", "", false, time.Time{}}},

		// A type with MarshalText, such as a date struct, shows its text.
		{"MarshalText", reflect.ValueOf(textDate{2026, 7, 6}), date, DateTimeInputSpec{"date", "2026-07-06", "YYYY-MM-DD", "", false, time.Time{}}},
		{"MarshalText through pointer", reflect.ValueOf(&textDate{2026, 7, 6}), date, DateTimeInputSpec{"date", "2026-07-06", "YYYY-MM-DD", "", false, time.Time{}}},
		{"MarshalText zero value", reflect.ValueOf(textDate{}), date, DateTimeInputSpec{"date", "", "YYYY-MM-DD", "", false, time.Time{}}},
		{"MarshalText other form", reflect.ValueOf(textDate{2026, 7, 6}), datetime, DateTimeInputSpec{"text", "2026-07-06", "YYYY-MM-DDTHH:MM", "", false, time.Time{}}},
		// A bound cannot be checked on text the input cannot show.
		{"MarshalText other form with min", reflect.ValueOf(textDate{2026, 7, 6}), FormTag{Widget: "datetime", Min: "2026-01-01T00:00"}, DateTimeInputSpec{"text", "2026-07-06", "YYYY-MM-DDTHH:MM", "", true, time.Time{}}},
		{"no MarshalText", reflect.ValueOf(struct{ N int }{1}), date, DateTimeInputSpec{"date", "", "YYYY-MM-DD", "", false, time.Time{}}},
		{"MarshalText on the pointer receiver", reflect.ValueOf(&ptrTextDate{2026, 7, 6}), date, DateTimeInputSpec{"date", "2026-07-06", "YYYY-MM-DD", "", false, time.Time{}}},
		// The type's own IsZero decides, not the struct's reflect zero.
		{"IsZero date", reflect.ValueOf(zeroableDate{1, 1, 1}), date, DateTimeInputSpec{"date", "", "YYYY-MM-DD", "", false, time.Time{}}},
		{"non-zero IsZero date", reflect.ValueOf(zeroableDate{2026, 7, 6}), date, DateTimeInputSpec{"date", "2026-07-06", "YYYY-MM-DD", "", false, time.Time{}}},

		// Types that wrap time.Time render like it.
		{"embedded time date", reflect.ValueOf(wrappedTime{when}), date, DateTimeInputSpec{"date", "2026-07-06", "YYYY-MM-DD", "", false, time.Time{}}},
		{"embedded time datetime", reflect.ValueOf(wrappedTime{when}), datetime, DateTimeInputSpec{"datetime-local", "2026-07-06T09:05:07", "YYYY-MM-DDTHH:MM", "", false, time.Time{}}},
		{"embedded time time", reflect.ValueOf(&wrappedTime{when}), clock, DateTimeInputSpec{"time", "09:05:07", "HH:MM", "", false, time.Time{}}},
		{"embedded zero time", reflect.ValueOf(wrappedTime{}), date, DateTimeInputSpec{"date", "", "YYYY-MM-DD", "", false, time.Time{}}},
		{"defined time", reflect.ValueOf(definedTime(when)), date, DateTimeInputSpec{"date", "2026-07-06", "YYYY-MM-DD", "", false, time.Time{}}},
		{"defined zero time", reflect.ValueOf(definedTime{}), datetime, DateTimeInputSpec{"datetime-local", "", "YYYY-MM-DDTHH:MM", "", false, time.Time{}}},
		{"defined time datetime", reflect.ValueOf(definedTime(when)), datetime, DateTimeInputSpec{"datetime-local", "2026-07-06T09:05:07", "YYYY-MM-DDTHH:MM", "", false, time.Time{}}},
		{"defined time time", reflect.ValueOf(definedTime(when)), clock, DateTimeInputSpec{"time", "09:05:07", "HH:MM", "", false, time.Time{}}},
		// The wall clock of the stored zone is shown, not UTC's.
		{"embedded time in another zone", reflect.ValueOf(wrappedTime{time.Date(2026, 7, 6, 0, 30, 0, 0, time.FixedZone("CEST", 2*60*60))}), date, DateTimeInputSpec{"date", "2026-07-06", "YYYY-MM-DD", "", false, time.Time{}}},
		// A nil embedded pointer's promoted methods panic; it is no value.
		{"nil embedded pointer", reflect.ValueOf(ptrWrappedTime{}), date, DateTimeInputSpec{"date", "", "YYYY-MM-DD", "", false, time.Time{}}},
		{"datetime struct", reflect.ValueOf(civDateTime{when.Truncate(time.Minute)}), datetime, DateTimeInputSpec{"datetime-local", "2026-07-06T09:05", "YYYY-MM-DDTHH:MM", "", false, time.Time{}}},

		// Without Format, a string type's MarshalText gives the text.
		{"string MarshalText", reflect.ValueOf(compactDate("20260706")), date, DateTimeInputSpec{"date", "2026-07-06", "YYYY-MM-DD", "", false, time.Time{}}},
		// Stored text that is a valid value shows even when MarshalText's
		// text is not one, since saving it unchanged works.
		{"string MarshalText unreadable, stored valid", reflect.ValueOf(zonedText("2024-01-02")), date, DateTimeInputSpec{"date", "2024-01-02", "YYYY-MM-DD", "", false, time.Time{}}},
		// A Format that answers "" still lets a valid stored text show, such
		// as posted text kept after a rejection; a zero date stays empty.
		{"Format answers kept text with empty", reflect.ValueOf(dottedDate("2026-08-05")), date, DateTimeInputSpec{"date", "2026-08-05", "YYYY-MM-DD", "", false, time.Time{}}},
		{"Format reads its own spelling", reflect.ValueOf(dottedDate("05.08.2026")), date, DateTimeInputSpec{"date", "2026-08-05", "YYYY-MM-DD", "", false, time.Time{}}},

		// Sensitive values of every kind stay hidden.
		{"sensitive Format", reflect.ValueOf(isoDate("06.07.2026")), FormTag{Widget: "date", Sensitive: true}, DateTimeInputSpec{"date", "", "YYYY-MM-DD", "", false, time.Time{}}},
		{"sensitive MarshalText", reflect.ValueOf(textDate{2026, 7, 6}), FormTag{Widget: "date", Sensitive: true}, DateTimeInputSpec{"date", "", "YYYY-MM-DD", "", false, time.Time{}}},
		{"sensitive defined time", reflect.ValueOf(definedTime(when)), FormTag{Widget: "date", Sensitive: true}, DateTimeInputSpec{"date", "", "YYYY-MM-DD", "", false, time.Time{}}},
		{"sensitive embedded time", reflect.ValueOf(wrappedTime{when}), FormTag{Widget: "date", Sensitive: true}, DateTimeInputSpec{"date", "", "YYYY-MM-DD", "", false, time.Time{}}},

		// The browser counts the default 60-second step from min: a value
		// that is not whole minutes after it needs step 1. The tag's step
		// wins.
		{"seconds with min", reflect.ValueOf("2026-07-06T09:05:07"), FormTag{Widget: "datetime", Min: "2026-01-01T00:00"}, DateTimeInputSpec{"datetime-local", "2026-07-06T09:05:07", "YYYY-MM-DDTHH:MM", "1", false, time.Time{}}},
		{"time with seconds and min", reflect.ValueOf(when), FormTag{Widget: "time", Min: "08:00"}, DateTimeInputSpec{"time", "09:05:07", "HH:MM", "1", false, time.Time{}}},
		{"whole minutes with min", reflect.ValueOf("09:05"), FormTag{Widget: "time", Min: "08:00"}, DateTimeInputSpec{"time", "09:05", "HH:MM", "", false, time.Time{}}},
		// A time.Time always shows its seconds, zero ones too.
		{"time zero seconds with min", reflect.ValueOf(when.Truncate(time.Minute)), FormTag{Widget: "time", Min: "08:00"}, DateTimeInputSpec{"time", "09:05:00", "HH:MM", "", false, time.Time{}}},
		{"whole minutes off a min with seconds", reflect.ValueOf("09:05"), FormTag{Widget: "time", Min: "08:00:30"}, DateTimeInputSpec{"time", "09:05", "HH:MM", "1", false, time.Time{}}},
		{"seconds on a min with seconds", reflect.ValueOf("09:05:30"), FormTag{Widget: "time", Min: "08:00:30"}, DateTimeInputSpec{"time", "09:05:30", "HH:MM", "", false, time.Time{}}},
		// Without a valid min the browser counts from the value itself.
		{"seconds without min", reflect.ValueOf("09:05:07"), clock, DateTimeInputSpec{"time", "09:05:07", "HH:MM", "", false, time.Time{}}},
		{"seconds with invalid min", reflect.ValueOf("09:05:07"), FormTag{Widget: "time", Min: "soon"}, DateTimeInputSpec{"time", "09:05:07", "HH:MM", "", false, time.Time{}}},
		{"text fallback with min", reflect.ValueOf("2026-07-06 09:05:07"), FormTag{Widget: "datetime", Min: "2026-01-01T00:00"}, DateTimeInputSpec{"text", "2026-07-06 09:05:07", "YYYY-MM-DDTHH:MM", "", true, time.Time{}}},
		{"tag step", reflect.ValueOf("2026-07-06"), FormTag{Widget: "date", Step: "7"}, DateTimeInputSpec{"date", "2026-07-06", "YYYY-MM-DD", "7", false, time.Time{}}},
		{"tag step over seconds and min", reflect.ValueOf("09:05:07"), FormTag{Widget: "time", Min: "08:00", Step: "30"}, DateTimeInputSpec{"time", "09:05:07", "HH:MM", "30", false, time.Time{}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The shown time is checked by TestDateTimeInputTime.
			got := DateTimeInput(tt.value, tt.tag)
			got.Time = time.Time{}
			if got != tt.want {
				t.Errorf("DateTimeInput = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// The spec carries the time the input shows, in the location it is shown
// in, for a time.Time and a type wrapping one, and no time for text, a
// zero time or a sensitive field.
func TestDateTimeInputTime(t *testing.T) {
	cest := time.FixedZone("CEST", 2*60*60)
	when := time.Date(2026, 7, 6, 9, 5, 7, 0, cest)
	type embedded struct{ time.Time }
	for _, c := range []struct {
		name  string
		value reflect.Value
		tag   FormTag
		want  time.Time
	}{
		{"time", reflect.ValueOf(when), FormTag{}, when},
		{"embedded time", reflect.ValueOf(embedded{when}), FormTag{Widget: "date"}, when},
		{"zero time", reflect.ValueOf(time.Time{}), FormTag{}, time.Time{}},
		{"text", reflect.ValueOf("2026-07-06"), FormTag{Widget: "date"}, time.Time{}},
		{"sensitive", reflect.ValueOf(when), FormTag{Sensitive: true}, time.Time{}},
	} {
		got := DateTimeInput(c.value, c.tag).Time
		if !got.Equal(c.want) {
			t.Errorf("%s: Time = %v, want %v", c.name, got, c.want)
		}
		if _, offset := got.Zone(); !c.want.IsZero() && offset != 2*60*60 {
			t.Errorf("%s: Time shown in offset %d, want the stored one", c.name, offset)
		}
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
		{name: "invalid max ignored", tag: FormTag{Widget: "date", Max: "later"}, raw: "2999-01-01", want: "2999-01-01"},

		// A time range may wrap past midnight; a date range may not.
		{name: "overnight after min", tag: FormTag{Widget: "time", Min: "22:00", Max: "06:00"}, raw: "23:30", want: "23:30"},
		{name: "overnight before max", tag: FormTag{Widget: "time", Min: "22:00", Max: "06:00"}, raw: "01:00", want: "01:00"},
		{name: "overnight on the bounds", tag: FormTag{Widget: "time", Min: "22:00", Max: "06:00"}, raw: "06:00", want: "06:00"},
		{name: "overnight midday", tag: FormTag{Widget: "time", Min: "22:00", Max: "06:00"}, raw: "12:00",
			wantErr: `"12:00" is later than 06:00 and earlier than 22:00`},
		{name: "reversed date range", tag: FormTag{Widget: "date", Min: "2026-12-31", Max: "2026-01-01"}, raw: "2026-07-06",
			wantErr: `"2026-07-06" is earlier than 2026-12-31`},

		// A sensitive field's errors leave the posted value out.
		{name: "sensitive format", tag: FormTag{Widget: "date", Sensitive: true}, raw: "02.01.1980",
			wantErr: "not a date in the format YYYY-MM-DD"},
		{name: "sensitive range", tag: FormTag{Widget: "date", Sensitive: true, Max: "2010-12-31"}, raw: "2015-03-04",
			wantErr: "later than 2010-12-31"},
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

// The bounds have no zone, so a zoned time is compared by its wall clock.
func TestCheckDateTimeRangeWallClock(t *testing.T) {
	tag := FormTag{Min: "2024-01-01T00:00", Max: "2024-12-31T23:59"}
	for _, tt := range []struct {
		t       time.Time
		wantErr string
	}{
		{time.Date(2023, 12, 31, 13, 0, 0, 0, time.FixedZone("", -12*60*60)), `"x" is earlier than 2024-01-01T00:00`},
		{time.Date(2025, 1, 1, 9, 0, 0, 0, time.FixedZone("", 14*60*60)), `"x" is later than 2024-12-31T23:59`},
		{time.Date(2024, 1, 1, 0, 30, 0, 0, time.FixedZone("", 2*60*60)), ""},
	} {
		err := CheckDateTimeRange(tag, "x", tt.t)
		if got := fmt.Sprint(err); (tt.wantErr == "" && err != nil) || (tt.wantErr != "" && got != tt.wantErr) {
			t.Errorf("CheckDateTimeRange(%v): error %v, want %q", tt.t, err, tt.wantErr)
		}
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

func TestDateTimeInputLayout(t *testing.T) {
	for widget, want := range map[string]string{
		"date":     "2006-01-02",
		"datetime": "2006-01-02T15:04:05",
		"time":     "15:04:05",
		"":         "2006-01-02T15:04:05",
	} {
		if got := DateTimeInputLayout(FormTag{Widget: widget}); got != want {
			t.Errorf("DateTimeInputLayout(%q) = %q, want %q", widget, got, want)
		}
	}
}

// panickyDay's Format panics on text it never accepted, which a rejected
// submission, bound as posted, is.
type panickyDay string

func (d panickyDay) Format(layout string) string {
	t, err := time.Parse(time.DateOnly, string(d))
	if err != nil {
		panic(err)
	}
	return t.Format(layout)
}

// A Format that panics shows the stored text as it is in the text input
// instead of failing the page.
func TestDateTimeInput_FormatPanicShowsStoredText(t *testing.T) {
	d := panickyDay("14.06.2026")
	in := DateTimeInput(reflect.ValueOf(&d).Elem(), FormTag{Widget: "date"})
	if !in.TextFallback() || in.Value != "14.06.2026" || !in.Invalid {
		t.Errorf("DateTimeInput = %+v", in)
	}
}
