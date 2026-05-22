package shadcn

import "testing"

// TestCn exercises the full cn helper — the clsx flatten layer composed with
// twmerge conflict resolution. Packages clsx and twmerge carry their own
// exhaustive suites; this verifies the composition.
//
// Case 3's expectation matches real tailwind-merge: an important and a
// non-important class do not conflict, so both are kept.
func TestCn(t *testing.T) {
	examples := []struct {
		input    []any
		expected string
	}{
		{[]any{"px-2 py-1", "bg-red-500"}, "px-2 py-1 bg-red-500"},
		{[]any{"px-2", "p-4"}, "p-4"},
		{[]any{"px-2", "!px-4"}, "px-2 !px-4"},
		{[]any{"px-2", "", false, "py-1"}, "px-2 py-1"},
		// flatten feeds twmerge: a conflict across a nested []string is resolved.
		{[]any{[]string{"p-2", "", "p-4"}}, "p-4"},
		// multiple args are joined before merging.
		{[]any{"  block  px-2", " ", "     py-4  "}, "block px-2 py-4"},
	}
	for _, ex := range examples {
		if got := Cn(ex.input...); got != ex.expected {
			t.Fatalf("Cn(%v) = %q, want %q", ex.input, got, ex.expected)
		}
	}
}
