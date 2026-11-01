package mx

import (
	"math"
	"reflect"
	"testing"
)

// The range is compared on the bound value: an integer exactly, also past
// 2^53, a float at its own precision, so a value on its bound passes, and a
// bound that is no HTML number is ignored, as a browser ignores it.
func TestCheckNumberRange(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value any
		tag   FormTag
		ok    bool
	}{
		{"no bounds", 5, FormTag{}, true},
		{"within", 5, FormTag{Min: "1", Max: "10"}, true},
		{"below", 0, FormTag{Min: "1"}, false},
		{"above", 11, FormTag{Max: "10"}, false},
		{"int64 past 2^53", int64(math.MaxInt64), FormTag{Max: "9223372036854775806"}, false},
		{"float on its bound", 99.9, FormTag{Max: "99.9"}, true},
		{"float32 on its bound", float32(0.1), FormTag{Min: "0.1"}, true},
		{"bound no HTML number", 0, FormTag{Min: "one"}, true},
		{"nil pointer", (*int)(nil), FormTag{Min: "1"}, true},
		{"pointer below", new(0), FormTag{Min: "1"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := CheckNumberRange(reflect.ValueOf(tc.value), "", tc.tag); (err == nil) != tc.ok {
				t.Errorf("CheckNumberRange = %v, want ok %t", err, tc.ok)
			}
		})
	}
	// A sensitive field's error names only the bound, as a user message.
	err := CheckNumberRange(reflect.ValueOf(0), "0", FormTag{Min: "1", Sensitive: true})
	if _, ok := UserMessage(err); !ok || err.Error() != "must be at least 1" {
		t.Errorf("sensitive: %v", err)
	}
}
