package shadcn

import (
	"strings"
	"testing"
)

func TestAspectRatio(t *testing.T) {
	out := render(t, AspectRatio(16.0/9.0, "content"))
	for _, want := range []string{
		`data-slot="aspect-ratio"`,
		"aspect-ratio: 1.7777",
		">content<",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}

func TestAspectRatioDefault(t *testing.T) {
	out := render(t, AspectRatio(0))
	if !strings.Contains(out, `aspect-ratio: 1"`) {
		t.Errorf("ratio <= 0 should default to a square ratio of 1: %s", out)
	}
}
