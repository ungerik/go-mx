package clsx

import "testing"

// TestJoin covers the flatten layer: slices, nested []any, conditional maps
// and falsy values. Join does not resolve Tailwind conflicts, so conflicting
// classes are both kept (see TestMerge in package twmerge for resolution).
func TestJoin(t *testing.T) {
	cases := []struct {
		in   []any
		want string
	}{
		{[]any{[]string{"px-2", "py-1"}}, "px-2 py-1"},
		{[]any{[]string{"p-2", "", "p-4"}}, "p-2 p-4"},
		{[]any{"flex", map[string]bool{"hidden": false, "block": false}}, "flex"},
		{[]any{map[string]bool{"font-bold": true}}, "font-bold"},
		{[]any{map[string]bool{"z-10": true, "a-1": true}}, "a-1 z-10"},
		{[]any{nil, false, "", "block"}, "block"},
		{[]any{[]any{"foo", []any{"bar", []any{"", []any{[]any{"baz"}}}}}}, "foo bar baz"},
		{[]any{[]any{}, []any{}}, ""},
	}
	for _, c := range cases {
		if got := Join(c.in...); got != c.want {
			t.Errorf("Join(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}
