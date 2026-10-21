package mx

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestIsNull(t *testing.T) {
	var intPtr *int
	tests := []struct {
		name string
		val  any
		want bool
	}{
		{"nil", nil, true},
		{"nil intPtr", intPtr, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNull(tt.val); got != tt.want {
				t.Errorf("IsNull(%#v) = %v, want %v", tt.val, got, tt.want)
			}
		})
	}
}

// scans gives a test type the Scan method a database type has.
type scans struct{}

func (*scans) Scan(any) error { return nil }

// failingValuer is a driver.Valuer whose Value fails.
type failingValuer struct{ scans }

func (failingValuer) Value() (driver.Value, error) { return nil, errors.New("broken") }

// intValuer returns a driver value no driver accepts, an int.
type intValuer struct {
	scans
	n int
}

func (v intValuer) Value() (driver.Value, error) { return v.n, nil }

// ptrValuer has Value on the pointer, so only an addressable value has it.
type ptrValuer struct {
	scans
	s string
}

func (v *ptrValuer) Value() (driver.Value, error) { return v.s, nil }

// bytesValuer holds a []byte driver value.
type bytesValuer struct {
	scans
	b []byte
}

func (v bytesValuer) Value() (driver.Value, error) { return v.b, nil }

// writeOnly has Value but no Scan, so nothing reads its text back.
type writeOnly struct{ n int64 }

func (v writeOnly) Value() (driver.Value, error) { return v.n, nil }

// centsMoney stores cents but reads and shows major units with its own
// methods, so its driver value is not its text.
type centsMoney struct {
	scans
	cents int64
}

func (m centsMoney) Value() (driver.Value, error) { return m.cents, nil }
func (m centsMoney) String() string               { return fmt.Sprintf("%d.%02d", m.cents/100, m.cents%100) }

// writesText has MarshalText but no UnmarshalText.
type writesText struct {
	scans
	n int64
}

func (v writesText) Value() (driver.Value, error) { return v.n, nil }
func (writesText) MarshalText() ([]byte, error)   { return []byte("x"), nil }

// readsText has UnmarshalText but no MarshalText or String.
type readsText struct {
	scans
	n int64
}

func (v readsText) Value() (driver.Value, error) { return v.n, nil }
func (*readsText) UnmarshalText([]byte) error    { return nil }

func TestDriverValueText(t *testing.T) {
	berlin, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Fatal(err)
	}
	at := time.Date(2026, 3, 29, 3, 30, 0, 5, berlin)
	for _, tc := range []struct {
		name   string
		value  any
		text   string
		wantOK bool
	}{
		{"int64", sql.NullInt64{Int64: 42, Valid: true}, "42", true},
		{"null", sql.NullInt64{Int64: 42}, "", true},
		{"string", sql.NullString{String: "a b", Valid: true}, "a b", true},
		{"float", sql.NullFloat64{Float64: 1.5, Valid: true}, "1.5", true},
		// Plain decimal, as a user types it, never an exponent.
		{"large float", sql.NullFloat64{Float64: 2500000, Valid: true}, "2500000", true},
		{"bool", sql.NullBool{Bool: true, Valid: true}, "true", true},
		{"byte as int64", sql.NullByte{Byte: 7, Valid: true}, "7", true},
		{"time", sql.NullTime{Time: at, Valid: true}, "2026-03-29T03:30:00.000000005+02:00", true},
		{"generic", sql.Null[string]{V: "x", Valid: true}, "x", true},
		{"text bytes", bytesValuer{b: []byte(`{"a":1}`)}, `{"a":1}`, true},
		{"binary bytes", bytesValuer{b: []byte{0xff, 0x00, 'A'}}, "", false},
		{"NUL in a string", sql.NullString{String: "a\x00b", Valid: true}, "", false},
		{"failing Value", failingValuer{}, "", false},
		{"unsupported driver value", intValuer{n: 3}, "", false},
		{"no Scan", writeOnly{7}, "", false},
		{"own String", centsMoney{cents: 1250}, "", false},
		{"own UnmarshalText", readsText{n: 1250}, "", false},
		{"own MarshalText", writesText{n: 3}, "", false},
		{"no Valuer", struct{ N int }{1}, "", false},
		{"not a struct", 5, "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			text, ok := DriverValueText(reflect.ValueOf(tc.value))
			if text != tc.text || ok != tc.wantOK {
				t.Errorf("DriverValueText = %q, %v; want %q, %v", text, ok, tc.text, tc.wantOK)
			}
		})
	}
	addressable := reflect.ValueOf(&ptrValuer{s: "p"}).Elem()
	if text, ok := DriverValueText(addressable); text != "p" || !ok {
		t.Errorf("pointer-receiver Value: %q, %v; want \"p\", true", text, ok)
	}
	if _, ok := DriverValueText(reflect.ValueOf(ptrValuer{s: "p"})); ok {
		t.Error("an unaddressable value has no pointer-receiver Value")
	}
}

// A nil pointer is null without calling a value method through it.
func TestIsNullNilPointer(t *testing.T) {
	if !IsNull((*sql.NullString)(nil)) {
		t.Error("a nil *sql.NullString must be null")
	}
}
