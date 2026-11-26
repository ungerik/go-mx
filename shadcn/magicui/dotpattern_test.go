package magicui

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestDotPatternDefault(t *testing.T) {
	out := render(t, DotPattern(DotPatternOptions{ID: "dots"}))
	mustContain(t, out, `data-slot="dot-pattern"`, `aria-hidden="true"`, "text-neutral-400/80",
		`<radialGradient id="dots-gradient">`, `<pattern id="dots" width="16" height="16"`, `x="0" y="0"`,
		`<circle cx="1" cy="1" r="1" fill="currentColor">`, `fill="url(#dots)"`)
	if strings.Contains(out, "animate-") {
		t.Errorf("a dot pattern without glow should not animate: %s", out)
	}
}

func TestDotPatternOptions(t *testing.T) {
	out := render(t, DotPattern(DotPatternOptions{ID: "d", Width: 20, Height: 10, X: 2, Y: 3, CX: 5, CY: 4, CR: 2},
		html.Class("text-sky-500")))
	mustContain(t, out, `width="20" height="10"`, `x="2" y="3"`, `cx="5" cy="4" r="2"`, "text-sky-500")
	if strings.Contains(out, "text-neutral-400/80") {
		t.Errorf("caller color should replace the default: %s", out)
	}
	// A dot crossing the tile's edge is drawn again on the opposite side.
	out = render(t, DotPattern(DotPatternOptions{ID: "d", CX: 1, CR: 3}))
	mustContain(t, out, `cx="1" cy="1" r="3"`, `cx="17" cy="1" r="3"`, `cx="1" cy="17" r="3"`, `cx="17" cy="17" r="3"`)
}

func TestDotPatternGlow(t *testing.T) {
	out := render(t, DotPattern(DotPatternOptions{ID: "g", Glow: true, Seed: 1}))
	mustContain(t, out, `width="128" height="128"`, `fill="url(#g-gradient)"`,
		"animate-dot-pattern-glow", "[transform-box:fill-box]", "animation-delay:", "animation-duration:",
		`class="animate-dot-pattern-repaint"`)
	if n := strings.Count(out, "animate-dot-pattern-glow"); n != dotPatternGlowCells*dotPatternGlowCells {
		t.Errorf("%d glowing dots, want %d", n, dotPatternGlowCells*dotPatternGlowCells)
	}
	if again := render(t, DotPattern(DotPatternOptions{ID: "g", Glow: true, Seed: 1})); again != out {
		t.Error("the same seed rendered different markup")
	}
	if other := render(t, DotPattern(DotPatternOptions{ID: "g", Glow: true, Seed: 2})); other == out {
		t.Error("a different seed rendered the same markup")
	}
}

func TestDotPatternInvalidID(t *testing.T) {
	if renderErr(t, DotPattern(DotPatternOptions{ID: "x y"})) == nil {
		t.Error("invalid id accepted")
	}
}
