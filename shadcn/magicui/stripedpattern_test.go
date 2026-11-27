package magicui

import (
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestStripedPatternDefault(t *testing.T) {
	out := render(t, StripedPattern(StripedPatternOptions{ID: "stripes"}))
	mustContain(t, out, `data-slot="striped-pattern"`, `aria-hidden="true"`,
		"pointer-events-none absolute inset-0 z-10 h-full w-full stroke-[0.5]",
		`<pattern id="stripes" width="10" height="10" patternUnits="userSpaceOnUse">`,
		`<line x1="0" y1="10" x2="10" y2="0" stroke="currentColor">`,
		`<line x1="-10" y1="10" x2="0" y2="0"`, `<line x1="10" y1="10" x2="20" y2="0"`,
		`fill="url(#stripes)"`)
}

func TestStripedPatternOptions(t *testing.T) {
	out := render(t, StripedPattern(StripedPatternOptions{ID: "s", Direction: StripedPatternRight, Width: 20, Height: 8},
		html.Class("stroke-[0.3]")))
	mustContain(t, out, `width="20" height="8"`, `<line x1="0" y1="0" x2="20" y2="8"`,
		`<line x1="-20" y1="0" x2="0" y2="8"`, `<line x1="20" y1="0" x2="40" y2="8"`, "stroke-[0.3]")
	if renderErr(t, StripedPattern(StripedPatternOptions{ID: "a'b"})) == nil {
		t.Error("invalid id accepted")
	}
}
