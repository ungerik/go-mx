package mx

import "testing"

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
