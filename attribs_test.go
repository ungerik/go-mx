package mx

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAttribs_Attribs(t *testing.T) {
	tests := []struct {
		name string
		a    Attribs
		want []Attrib
	}{
		{name: "nil", a: nil, want: nil},
		{name: "empty", a: Attribs{}, want: nil},
		{name: "single", a: Attribs{"a": "val_a"}, want: []Attrib{Attribute{"a", "val_a"}}},
		{name: "sort 3", a: Attribs{"c": "val_c", "b": "val_b", "a": "val_a"}, want: []Attrib{Attribute{"a", "val_a"}, Attribute{"b", "val_b"}, Attribute{"c", "val_c"}}},
		{name: "sort id", a: Attribs{"c": "val_c", "id": "val_id", "a": "val_a"}, want: []Attrib{Attribute{"id", "val_id"}, Attribute{"a", "val_a"}, Attribute{"c", "val_c"}}},
		{name: "sort id class", a: Attribs{"class": "val_class", "id": "val_id", "a": "val_a", "b": "val_b"}, want: []Attrib{Attribute{"id", "val_id"}, Attribute{"class", "val_class"}, Attribute{"a", "val_a"}, Attribute{"b", "val_b"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.a.Attribs()
			require.Equal(t, tt.want, got, "Attribs.Attribs()")
		})
	}
}
