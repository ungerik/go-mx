package shadcn

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestSpinner(t *testing.T) {
	out := render(t, Spinner())
	// role/aria-label make the spinner announce as a live loading indicator;
	// animate-spin is the entire behavior (no JS).
	for _, want := range []string{"<svg", `data-slot="spinner"`, `role="status"`, `aria-label="Loading"`, "animate-spin", "size-4", "lucide-loader-circle"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}

func TestSpinnerCallerOverrides(t *testing.T) {
	out := render(t, Spinner(html.Class("size-6")))
	// Caller classes must win the twmerge conflict so the spinner is sizable.
	if strings.Contains(out, "size-4") || !strings.Contains(out, "size-6") {
		t.Errorf("caller size class should override the default size-4: %s", out)
	}
}
