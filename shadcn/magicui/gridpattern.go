package magicui

import (
	"fmt"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/svg"
)

// GridPatternOptions configures a [GridPattern]. The zero value is
// upstream's default: 40×40 pixel cells with solid lines, offset by -1 pixel.
type GridPatternOptions struct {
	// ID is the id of the SVG <pattern>, which the background rect references
	// as url(#ID). Empty means a process-unique id from [mx.UniqueID]; pass a
	// fixed one where the markup has to be deterministic. It may only contain
	// letters, digits, '_' and '-'.
	ID string
	// Width is the cell width in pixels; 0 means 40.
	Width float64
	// Height is the cell height in pixels; 0 means 40.
	Height float64
	// X is the horizontal offset of the grid in pixels; nil means -1.
	X *float64
	// Y is the vertical offset of the grid in pixels; nil means -1.
	Y *float64
	// StrokeDasharray is the stroke-dasharray of the grid lines, such as
	// "4 2"; empty means "0", a solid line.
	StrokeDasharray string
	// Squares are the cells, as [column, row] pairs, that are filled.
	Squares [][2]int
}

// GridPattern draws a grid of lines as an absolutely positioned SVG
// background that fills its positioned parent, optionally with some cells
// filled. The lines and filled cells take their colors from the SVG's stroke
// and fill, which a caller class such as html.Class("stroke-sky-500/30")
// overrides; a mask class fades the grid out. Children are appended inside
// the <svg>, after the pattern. The SVG is aria-hidden.
func GridPattern(opts GridPatternOptions, attribsChildren ...any) *mx.Element {
	id, err := svgRefID(opts.ID)
	if err != nil {
		return mx.NewErrElement(err)
	}
	dash := or(opts.StrokeDasharray, "0")
	if err := checkCSSValue("stroke-dasharray", dash); err != nil {
		return mx.NewErrElement(err)
	}
	width := or(opts.Width, 40)
	height := or(opts.Height, 40)
	x := derefOr(opts.X, -1)
	y := derefOr(opts.Y, -1)

	parts := []mx.Component{
		svg.Defs(
			svg.Pattern(
				svg.ID(id),
				svg.Width(width),
				svg.Height(height),
				svg.PatternUnitsUserSpaceOnUse,
				svg.X(x),
				svg.Y(y),
				svg.Path(
					svg.D(fmt.Sprintf("M.5 %sV.5H%s", num(height), num(width))),
					svg.Fill("none"),
					svg.StrokeDashArray(dash),
				),
			),
		),
		svg.Rect(svg.Width("100%"), svg.Height("100%"), svg.StrokeWidth(0), svg.Fill("url(#"+id+")")),
	}
	if len(opts.Squares) > 0 {
		squares := svg.SVG(svg.X(x), svg.Y(y), svg.Class("overflow-visible"))
		for _, sq := range opts.Squares {
			squares.Children = append(squares.Children, svg.Rect(
				svg.StrokeWidth(0),
				svg.Width(width-1),
				svg.Height(height-1),
				svg.X(float64(sq[0])*width+1),
				svg.Y(float64(sq[1])*height+1),
			))
		}
		parts = append(parts, squares)
	}
	return patternSVG("grid-pattern",
		"pointer-events-none absolute inset-0 h-full w-full fill-gray-400/30 stroke-gray-400/30",
		parts, attribsChildren)
}
