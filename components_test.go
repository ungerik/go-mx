package mx

import (
	"reflect"
	"testing"
)

func TestComponentSlice(t *testing.T) {
	c := RawComponent{Opening: "Test"}
	tests := []struct {
		name string
		c    Component
		want []Component
	}{
		{name: "nil", c: nil, want: nil},
		{name: "single", c: c, want: []Component{c}},
		{name: "Components empty", c: Components{}, want: []Component{}},
		{name: "Components single", c: Components{c}, want: []Component{c}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ComponentSlice(tt.c); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ComponentSlice(%#v) = %#v, want %#v", tt.c, got, tt.want)
			}
		})
	}
}
