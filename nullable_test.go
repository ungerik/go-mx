package mx

import (
	"database/sql"
	"database/sql/driver"
	"errors"
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

// failingValuer is a driver.Valuer whose Value fails.
type failingValuer struct{ n int }

func (failingValuer) Value() (driver.Value, error) { return nil, errors.New("broken") }

// intValuer returns a driver value no driver accepts, an int.
type intValuer struct{ n int }

func (v intValuer) Value() (driver.Value, error) { return v.n, nil }

// ptrValuer has Value on the pointer, so only an addressable value has it.
type ptrValuer struct{ s string }

func (v *ptrValuer) Value() (driver.Value, error) { return v.s, nil }

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
		{"bool", sql.NullBool{Bool: true, Valid: true}, "true", true},
		{"byte as int64", sql.NullByte{Byte: 7, Valid: true}, "7", true},
		{"time", sql.NullTime{Time: at, Valid: true}, "2026-03-29T03:30:00.000000005+02:00", true},
		{"generic", sql.Null[string]{V: "x", Valid: true}, "x", true},
		{"failing Value", failingValuer{}, "", false},
		{"unsupported driver value", intValuer{3}, "", false},
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
