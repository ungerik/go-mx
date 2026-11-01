package mx

import (
	"database/sql"
	"reflect"
	"testing"
	"time"
	"uuid"
)

// textPtrReceiver writes its text only through a pointer receiver, which
// an addressable value reaches.
type textPtrReceiver struct{ s string }

func (t *textPtrReceiver) MarshalText() ([]byte, error) { return []byte("ptr:" + t.s), nil }

// promotedText promotes MarshalText through an embedded pointer, which
// panics while the pointer is nil.
type promotedText struct{ *textPtrReceiver }

// FormValueText is the one text both form renderers show and the html
// parser reads back, so a zero ID or time is no value and a time keeps its
// zone and fraction.
func TestFormValueText(t *testing.T) {
	at := time.Date(2026, 3, 29, 3, 30, 0, 5, time.FixedZone("CEST", 2*60*60))
	id := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	for _, tc := range []struct {
		name  string
		value reflect.Value
		want  string
	}{
		{"invalid", reflect.Value{}, ""},
		{"nil pointer", reflect.ValueOf((*int)(nil)), ""},
		{"pointer", reflect.ValueOf(new(7)), "7"},
		{"string", reflect.ValueOf("text"), "text"},
		{"bool", reflect.ValueOf(true), "true"},
		{"uint", reflect.ValueOf(uint8(200)), "200"},
		{"float", reflect.ValueOf(1.5), "1.5"},
		{"zero ID", reflect.ValueOf(uuid.UUID{}), ""},
		{"ID", reflect.ValueOf(id), id.String()},
		{"zero time", reflect.ValueOf(time.Time{}), ""},
		{"time", reflect.ValueOf(at), "2026-03-29T03:30:00.000000005+02:00"},
		{"defined time", reflect.ValueOf(definedTime(at)), "2026-03-29T03:30:00.000000005+02:00"},
		{"pointer receiver", reflect.ValueOf(&textPtrReceiver{s: "x"}).Elem(), "ptr:x"},
		{"panicking MarshalText", reflect.ValueOf(&promotedText{}).Elem(), ""},
		{"null", reflect.ValueOf(sql.NullInt64{Int64: 5}), ""},
		{"driver value", reflect.ValueOf(sql.NullInt64{Int64: 5, Valid: true}), "5"},
		{"lines", reflect.ValueOf([]string{"a", "b"}), "a\nb"},
		{"bytes", reflect.ValueOf([]byte("raw")), "raw"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := FormValueText(tc.value); got != tc.want {
				t.Errorf("FormValueText = %q, want %q", got, tc.want)
			}
		})
	}
}
