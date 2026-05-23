package shadcn

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestProgressDefault(t *testing.T) {
	out := render(t, Progress(40))
	for _, want := range []string{
		`data-slot="progress"`,
		`role="progressbar"`,
		`aria-valuemin="0"`,
		`aria-valuemax="100"`,
		`aria-valuenow="40"`,
		"bg-primary/20",
		`data-slot="progress-indicator"`,
		"bg-primary h-full",
		"transform: translateX(-60%)",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}

func TestProgressZero(t *testing.T) {
	out := render(t, Progress(0))
	if !strings.Contains(out, "translateX(-100%)") {
		t.Errorf("zero value should translate by -100%%: %s", out)
	}
	if !strings.Contains(out, `aria-valuenow="0"`) {
		t.Errorf("aria-valuenow=0 missing: %s", out)
	}
}

func TestProgressClamps(t *testing.T) {
	hi := render(t, Progress(150))
	if !strings.Contains(hi, "translateX(-0%)") || !strings.Contains(hi, `aria-valuenow="100"`) {
		t.Errorf("value >100 should clamp to 100: %s", hi)
	}
	lo := render(t, Progress(-10))
	if !strings.Contains(lo, "translateX(-100%)") || !strings.Contains(lo, `aria-valuenow="0"`) {
		t.Errorf("value <0 should clamp to 0: %s", lo)
	}
}

func TestProgressCallerClassMerges(t *testing.T) {
	out := render(t, Progress(50, html.Class("h-3")))
	for _, want := range []string{"h-3", "bg-primary/20", "rounded-full"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}
