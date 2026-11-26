package magicui

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestInteractiveHoverButton(t *testing.T) {
	out := render(t, InteractiveHoverButton(html.Class("w-48"), "Hover Me"))
	mustContain(t, out,
		`data-slot="interactive-hover-button"`, `type="button"`, "group-hover:scale-[100.8]",
		"group-hover:-translate-x-5", "lucide-arrow-right", "w-48",
	)
	if strings.Contains(out, "w-auto") {
		t.Errorf("caller w-48 should replace w-auto: %s", out)
	}
	if n := strings.Count(out, "Hover Me"); n != 2 {
		t.Errorf("label rendered %d times, want 2", n)
	}
	if n := strings.Count(out, `aria-hidden="true"`); n != 1 {
		t.Errorf("%d aria-hidden elements, want only the sliding copy", n)
	}
}
