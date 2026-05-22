package shadcn

import "testing"

func TestCn(t *testing.T) {
	// TODO fix test
	examples := []struct {
		input    []any
		expected string
	}{
		{
			input:    []any{"px-2 py-1", "bg-red-500"},
			expected: "px-2 py-1 bg-red-500",
		},
		{
			input:    []any{"px-2", "p-4"},
			expected: "p-4",
		},
		{
			input:    []any{"px-2", "!px-4"},
			expected: "!px-4",
		},
		{
			input:    []any{"px-2", "", false, "py-1"},
			expected: "px-2 py-1",
		},
	}

	for _, example := range examples {
		result := Cn(example.input...)
		if result != example.expected {
			t.Fatal("Test failed: got " + result + ", expected " + example.expected)
		}
	}
}
