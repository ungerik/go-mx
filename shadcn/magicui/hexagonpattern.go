//go:generate go -C ../../tools tool go-enum ../shadcn/magicui/$GOFILE

package magicui

import (
	"fmt"
	"math"
	"strings"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/svg"
)

// HexagonPatternDirection is the orientation of the hexagons of a
// [HexagonPattern].
type HexagonPatternDirection string //#enum

const (
	// HexagonPatternHorizontal tiles flat-top hexagons in a horizontal
	// honeycomb grid.
	HexagonPatternHorizontal HexagonPatternDirection = "horizontal"
	// HexagonPatternVertical tiles pointy-top hexagons in a vertical
	// honeycomb grid.
	HexagonPatternVertical HexagonPatternDirection = "vertical"
)

// Valid indicates if h is any of the valid values for HexagonPatternDirection
func (h HexagonPatternDirection) Valid() bool {
	switch h {
	case
		HexagonPatternHorizontal,
		HexagonPatternVertical:
		return true
	}
	return false
}

// Validate returns an error if h is none of the valid values for HexagonPatternDirection
func (h HexagonPatternDirection) Validate() error {
	if !h.Valid() {
		return fmt.Errorf("invalid value %#v for type magicui.HexagonPatternDirection", h)
	}
	return nil
}

// ValidOr returns h if it is any of the valid values for HexagonPatternDirection, else defaultVal
func (h HexagonPatternDirection) ValidOr(defaultVal HexagonPatternDirection) HexagonPatternDirection {
	if h.Valid() {
		return h
	}
	return defaultVal
}

// Enums returns all valid values for HexagonPatternDirection
func (HexagonPatternDirection) Enums() []HexagonPatternDirection {
	return []HexagonPatternDirection{
		HexagonPatternHorizontal,
		HexagonPatternVertical,
	}
}

// EnumStrings returns all valid values for HexagonPatternDirection as strings
func (HexagonPatternDirection) EnumStrings() []string {
	return []string{
		"horizontal",
		"vertical",
	}
}

// String implements the fmt.Stringer interface for HexagonPatternDirection
func (h HexagonPatternDirection) String() string {
	return string(h)
}

// HexagonPatternOptions configures a [HexagonPattern]. The zero value is
// upstream's default: flat-top hexagons of radius 40 without gaps, offset by
// -1 pixel, with solid outlines.
type HexagonPatternOptions struct {
	// ID is the id of the SVG <pattern>, which the background rect references
	// as url(#ID). Empty means a process-unique id from [mx.UniqueID]; pass a
	// fixed one where the markup has to be deterministic. It may only contain
	// letters, digits, '_' and '-'.
	ID string
	// Radius is the distance from a hexagon's center to its vertices in
	// pixels; 0 means 40.
	Radius float64
	// Gap is the spacing between adjacent hexagons in pixels. The tile grows
	// by it while the radius stays, so the gap is even on all sides.
	Gap float64
	// X is the horizontal offset of the pattern in pixels; nil means -1.
	X *float64
	// Y is the vertical offset of the pattern in pixels; nil means -1.
	Y *float64
	// Direction is the orientation of the hexagons; empty or unknown means
	// [HexagonPatternHorizontal].
	Direction HexagonPatternDirection
	// StrokeDasharray is the stroke-dasharray of the outlines, such as "4 2";
	// empty means "0", a solid line. A dashed outline is drawn edge by edge,
	// each shared edge once, so the dashes of neighbors do not overlap.
	StrokeDasharray string
	// Hexagons are the hexagons, as [column, row] pairs, that are filled.
	Hexagons [][2]int
}

// HexagonPattern draws a honeycomb of hexagon outlines as an absolutely
// positioned SVG background that fills its positioned parent, optionally
// with some hexagons filled. The outlines and filled hexagons take their
// colors from the SVG's stroke and fill, which a caller class such as
// html.Class("stroke-sky-500/30") overrides; a mask class fades the pattern
// out. Children are appended inside the <svg>, after the pattern. The SVG is
// aria-hidden.
func HexagonPattern(opts HexagonPatternOptions, attribsChildren ...any) *mx.Element {
	id, err := svgRefID(opts.ID)
	if err != nil {
		return mx.NewErrElement(err)
	}
	dash := or(opts.StrokeDasharray, "0")
	if err := checkCSSValue("stroke-dasharray", dash); err != nil {
		return mx.NewErrElement(err)
	}
	r := or(opts.Radius, 40)
	dir := opts.Direction.ValidOr(HexagonPatternHorizontal)
	x := derefOr(opts.X, -1)
	y := derefOr(opts.Y, -1)

	tileW, tileH, centers := hexagonPatternTile(r, dir, opts.Gap)
	pattern := svg.Pattern(
		svg.ID(id),
		svg.Width(tileW),
		svg.Height(tileH),
		svg.PatternUnitsUserSpaceOnUse,
		svg.X(x),
		svg.Y(y),
	)
	if t := strings.TrimSpace(dash); t == "" || t == "none" || t == "0" {
		for _, c := range centers {
			pattern.Children = append(pattern.Children, svg.Polygon(
				svg.Class("fill-none"),
				hexagonPatternPoints(c[0], c[1], r, dir),
				svg.StrokeDashArray(dash),
			))
		}
	} else {
		for _, edge := range hexagonPatternEdges(centers, r, dir) {
			a, b := edge[0], edge[1]
			pattern.Children = append(pattern.Children, svg.Line(
				svg.Class("fill-none"),
				svg.X1(a[0]), svg.X2(b[0]), svg.Y1(a[1]), svg.Y2(b[1]),
				svg.StrokeDashArray(dash),
			))
		}
	}
	parts := []mx.Component{
		svg.Defs(pattern),
		svg.Rect(svg.Width("100%"), svg.Height("100%"), svg.Fill("url(#"+id+")"), svg.Stroke("none")),
	}
	if len(opts.Hexagons) > 0 {
		filled := svg.SVG(ariaHidden, svg.Class("overflow-visible"), svg.X(x), svg.Y(y))
		for _, h := range opts.Hexagons {
			cx, cy := hexagonPatternCenter(h[0], h[1], r, dir, opts.Gap)
			filled.Children = append(filled.Children, svg.Polygon(
				hexagonPatternPoints(cx, cy, r-1, dir),
				svg.StrokeWidth(0),
			))
		}
		parts = append(parts, filled)
	}
	return patternSVG("hexagon-pattern",
		"pointer-events-none absolute inset-0 h-full w-full fill-gray-400/30 stroke-gray-400/30",
		parts, attribsChildren)
}

// hexagonPatternVertices returns the six vertices of the hexagon centered at
// cx, cy with radius r.
func hexagonPatternVertices(cx, cy, r float64, dir HexagonPatternDirection) [6][2]float64 {
	start := 0.0
	if dir == HexagonPatternVertical {
		start = 30
	}
	var verts [6][2]float64
	for i := range verts {
		angle := (start + float64(i)*60) * math.Pi / 180
		verts[i] = [2]float64{
			hexagonPatternRound(cx + r*math.Cos(angle)),
			hexagonPatternRound(cy + r*math.Sin(angle)),
		}
	}
	return verts
}

// hexagonPatternRound rounds f to a millionth of a pixel, so a vertex on an
// axis renders as 0 rather than as a float error such as 1.2e-15 or -0.
func hexagonPatternRound(f float64) float64 {
	return math.Round(f*1e6)/1e6 + 0 // + 0 turns -0 into 0
}

// hexagonPatternPoints returns the points attribute of a hexagon.
func hexagonPatternPoints(cx, cy, r float64, dir HexagonPatternDirection) mx.Attrib {
	verts := hexagonPatternVertices(cx, cy, r, dir)
	points := make([]string, len(verts))
	for i, v := range verts {
		points[i] = num(v[0]) + "," + num(v[1])
	}
	return svg.Attrib("points", strings.Join(points, " "))
}

// hexagonPatternEdges returns the edges of the hexagons centered at centers,
// each shared edge once.
func hexagonPatternEdges(centers [][2]float64, r float64, dir HexagonPatternDirection) [][2][2]float64 {
	seen := map[string]bool{}
	var edges [][2][2]float64
	for _, c := range centers {
		verts := hexagonPatternVertices(c[0], c[1], r, dir)
		for i := range verts {
			a, b := verts[i], verts[(i+1)%6]
			p, q := a, b
			if !(a[0] < b[0] || (a[0] == b[0] && a[1] <= b[1])) {
				p, q = b, a
			}
			key := fmt.Sprintf("%.6f,%.6f|%.6f,%.6f", p[0], p[1], q[0], q[1])
			if !seen[key] {
				seen[key] = true
				edges = append(edges, [2][2]float64{a, b})
			}
		}
	}
	return edges
}

// hexagonPatternSpacing returns the distance between the centers of
// adjacent columns and rows. The gap is added along the shared edge's
// normal, so it matches the visible edge-to-edge spacing.
func hexagonPatternSpacing(r float64, dir HexagonPatternDirection, gap float64) (colStep, rowStep float64) {
	sqrt3 := math.Sqrt(3)
	if dir == HexagonPatternHorizontal {
		return 3*r/2 + sqrt3*gap/2, sqrt3*r + gap
	}
	return sqrt3*r + gap, 3*r/2 + sqrt3*gap/2
}

// hexagonPatternTile returns the size of the pattern tile and the centers of
// the hexagons drawn in it: two canonical ones, plus the copies of those
// that cross an edge of the tile, so they continue in the next tile.
func hexagonPatternTile(r float64, dir HexagonPatternDirection, gap float64) (tileW, tileH float64, centers [][2]float64) {
	colStep, rowStep := hexagonPatternSpacing(r, dir, gap)
	var canonical [2][2]float64
	if dir == HexagonPatternHorizontal {
		tileW, tileH = colStep*2, rowStep
		canonical = [2][2]float64{{colStep / 2, rowStep / 2}, {colStep * 3 / 2, rowStep}}
	} else {
		tileW, tileH = colStep, rowStep*2
		canonical = [2][2]float64{{colStep / 2, rowStep / 2}, {colStep, rowStep * 3 / 2}}
	}
	for _, c := range canonical {
		cx, cy := c[0], c[1]
		top, bottom, left, right := cy-r < 0, cy+r > tileH, cx-r < 0, cx+r > tileW
		centers = append(centers, [2]float64{cx, cy})
		if top {
			centers = append(centers, [2]float64{cx, cy + tileH})
		}
		if bottom {
			centers = append(centers, [2]float64{cx, cy - tileH})
		}
		if left {
			centers = append(centers, [2]float64{cx + tileW, cy})
		}
		if right {
			centers = append(centers, [2]float64{cx - tileW, cy})
		}
		if top && left {
			centers = append(centers, [2]float64{cx + tileW, cy + tileH})
		}
		if top && right {
			centers = append(centers, [2]float64{cx - tileW, cy + tileH})
		}
		if bottom && left {
			centers = append(centers, [2]float64{cx + tileW, cy - tileH})
		}
		if bottom && right {
			centers = append(centers, [2]float64{cx - tileW, cy - tileH})
		}
	}
	return tileW, tileH, centers
}

// hexagonPatternCenter returns the center of the hexagon in column col and
// row row.
func hexagonPatternCenter(col, row int, r float64, dir HexagonPatternDirection, gap float64) (cx, cy float64) {
	colStep, rowStep := hexagonPatternSpacing(r, dir, gap)
	if dir == HexagonPatternHorizontal {
		cx = float64(col)*colStep + colStep/2
		cy = float64(row)*rowStep + rowStep/2
		if col%2 != 0 {
			cy += rowStep / 2
		}
		return cx, cy
	}
	cx = float64(col)*colStep + colStep/2
	if row%2 != 0 {
		cx += colStep / 2
	}
	cy = float64(row)*rowStep + rowStep/2
	return cx, cy
}
