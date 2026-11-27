package magicui

import (
	"math"
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestHexagonPatternDefault(t *testing.T) {
	out := render(t, HexagonPattern(HexagonPatternOptions{ID: "hex"}))
	mustContain(t, out, `data-slot="hexagon-pattern"`, `aria-hidden="true"`,
		"fill-gray-400/30 stroke-gray-400/30",
		// Flat-top hexagons of radius 40: the tile is 3r wide and √3·r high.
		`<pattern id="hex" width="120" height="69.28203230275508"`, `x="-1" y="-1"`,
		`<polygon class="fill-none" points="70,34.641016 50,69.282032 10,69.282032 -10,34.641016 10,0 50,0"`,
		`fill="url(#hex)" stroke="none"`)
	if strings.Contains(out, "<line") || strings.Contains(out, "-0,") {
		t.Errorf("solid outlines are polygons with clean coordinates: %s", out)
	}
}

func TestHexagonPatternOptions(t *testing.T) {
	out := render(t, HexagonPattern(HexagonPatternOptions{
		ID: "h", Radius: 20, Gap: 4, X: new(0.0), Y: new(0.0),
		Direction: HexagonPatternVertical, StrokeDasharray: "4 2",
		Hexagons: [][2]int{{1, 1}},
	}, html.Class("stroke-sky-500")))
	mustContain(t, out, `x="0" y="0"`, `<line class="fill-none"`, `stroke-dasharray="4 2"`,
		`class="overflow-visible"`, `stroke-width="0"`, "stroke-sky-500")
	if strings.Contains(out, "<polygon class=\"fill-none\"") {
		t.Errorf("dashed outlines are drawn edge by edge: %s", out)
	}
	if strings.Contains(out, "stroke-gray-400/30") {
		t.Errorf("caller stroke should replace the default: %s", out)
	}
	// Without a gap, neighbors share edges, each drawn once, so there are
	// fewer lines than six per hexagon.
	out = render(t, HexagonPattern(HexagonPatternOptions{ID: "h", StrokeDasharray: "4 2"}))
	_, _, centers := hexagonPatternTile(40, HexagonPatternHorizontal, 0)
	if lines := strings.Count(out, "<line"); lines == 0 || lines >= 6*len(centers) {
		t.Errorf("%d edges for %d hexagons", lines, len(centers))
	}
}

func TestHexagonPatternCenter(t *testing.T) {
	// Odd columns of flat-top hexagons are shifted down by half a row.
	x, y := hexagonPatternCenter(1, 0, 40, HexagonPatternHorizontal, 0)
	if x != 90 || math.Abs(y-40*math.Sqrt(3)) > 1e-9 {
		t.Errorf("center = %v, %v", x, y)
	}
}

func TestHexagonPatternInvalid(t *testing.T) {
	if renderErr(t, HexagonPattern(HexagonPatternOptions{StrokeDasharray: "1;x:y"})) == nil {
		t.Error("invalid stroke-dasharray accepted")
	}
	if renderErr(t, HexagonPattern(HexagonPatternOptions{ID: "#h"})) == nil {
		t.Error("invalid id accepted")
	}
}
