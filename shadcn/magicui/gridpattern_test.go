package magicui

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestGridPatternDefault(t *testing.T) {
	out := render(t, GridPattern(GridPatternOptions{ID: "grid"}))
	mustContain(t, out, `data-slot="grid-pattern"`, `aria-hidden="true"`,
		"fill-gray-400/30 stroke-gray-400/30",
		`id="grid"`, `width="40"`, `height="40"`, `patternUnits="userSpaceOnUse"`, `x="-1"`, `y="-1"`,
		`d="M.5 40V.5H40"`, `stroke-dasharray="0"`, `fill="url(#grid)"`)
	if strings.Contains(out, "overflow-visible") {
		t.Errorf("no squares expected: %s", out)
	}
}

func TestGridPatternOptions(t *testing.T) {
	out := render(t, GridPattern(GridPatternOptions{
		ID: "g2", Width: 30, Height: 20, X: new(0.0), Y: new(2.0), StrokeDasharray: "4 2",
		Squares: [][2]int{{4, 4}, {5, 1}},
	}, html.Class("stroke-sky-500")))
	mustContain(t, out, `d="M.5 20V.5H30"`, `stroke-dasharray="4 2"`, `x="0"`, `y="2"`,
		`class="overflow-visible"`, `width="29"`, `height="19"`, `x="121"`, `y="81"`, `x="151"`, `y="21"`,
		"stroke-sky-500")
	if strings.Contains(out, "stroke-gray-400/30") {
		t.Errorf("caller stroke should replace the default: %s", out)
	}
}

func TestGridPatternInvalid(t *testing.T) {
	if renderErr(t, GridPattern(GridPatternOptions{ID: "a)b"})) == nil {
		t.Error("invalid id accepted")
	}
	if renderErr(t, GridPattern(GridPatternOptions{StrokeDasharray: "1;x:y"})) == nil {
		t.Error("invalid stroke-dasharray accepted")
	}
}

func TestGridPatternUniqueID(t *testing.T) {
	a := render(t, GridPattern(GridPatternOptions{}))
	b := render(t, GridPattern(GridPatternOptions{}))
	if a == b {
		t.Error("two patterns without ID share an id")
	}
}
