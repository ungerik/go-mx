package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/svg"
)

// dotPatternGlowCells is the number of dots per side of the pattern tile of
// a glowing [DotPattern]: each of its 8×8 dots gets its own timing, and the
// tile repeats every 8 dots.
const dotPatternGlowCells = 8

// DotPatternOptions configures a [DotPattern]. The zero value is upstream's
// default: dots of radius 1 every 16 pixels, without glow.
type DotPatternOptions struct {
	// ID is the id of the SVG <pattern>; the dots' radial gradient gets ID
	// plus "-gradient". Empty means a process-unique id from [mx.UniqueID];
	// pass a fixed one where the markup has to be deterministic. It may only
	// contain letters, digits, '_' and '-'.
	ID string
	// Width is the horizontal spacing between dots in pixels; 0 means 16.
	Width float64
	// Height is the vertical spacing between dots in pixels; 0 means 16.
	Height float64
	// X is the horizontal offset of the whole pattern in pixels.
	X float64
	// Y is the vertical offset of the whole pattern in pixels.
	Y float64
	// CX is the horizontal offset of each dot within its cell; 0 means 1.
	CX float64
	// CY is the vertical offset of each dot within its cell; 0 means 1.
	CY float64
	// CR is the radius of each dot; 0 means 1.
	CR float64
	// Glow fills the dots with a radial gradient and lets each of them pulse
	// at its own random delay and duration.
	Glow bool
	// Seed seeds the random timings of the glowing dots; the same seed
	// renders the same markup.
	Seed uint64
}

// DotPattern draws a grid of dots as an absolutely positioned SVG
// background that fills its positioned parent. The dots are filled with
// currentColor, so a text color class such as html.Class("text-sky-500/50")
// recolors them; a mask class fades the pattern out. Children are appended
// inside the <svg>, after the pattern. The SVG is aria-hidden.
//
// Upstream measures its container and renders one circle per dot; this port
// repeats an SVG <pattern> tile instead, so it needs no JavaScript and
// fills any size. With opts.Glow the tile holds 8×8 dots, each pulsing with
// its own seeded delay and duration, so the timings repeat every 8 dots
// where upstream's are random for every dot.
//
// Needs [CSS] in the app's Tailwind input when opts.Glow is set.
func DotPattern(opts DotPatternOptions, attribsChildren ...any) *mx.Element {
	id, err := svgRefID(opts.ID)
	if err != nil {
		return mx.NewErrElement(err)
	}
	width := or(opts.Width, 16)
	height := or(opts.Height, 16)
	cx := or(opts.CX, 1)
	cy := or(opts.CY, 1)
	r := or(opts.CR, 1)
	gradientID := id + "-gradient"

	cells := 1
	if opts.Glow {
		cells = dotPatternGlowCells
	}
	tileW, tileH := width*float64(cells), height*float64(cells)
	rnd := newRand(opts.Seed)
	pattern := svg.Pattern(
		svg.ID(id),
		svg.Width(tileW),
		svg.Height(tileH),
		svg.PatternUnitsUserSpaceOnUse,
		svg.X(opts.X),
		svg.Y(opts.Y),
	)
	for row := range cells {
		for col := range cells {
			attribs := []any{svg.R(r)}
			if opts.Glow {
				delay := rnd.Float64() * 5
				duration := rnd.Float64()*3 + 2
				attribs = append(attribs,
					svg.Fill("url(#"+gradientID+")"),
					svg.Class("animate-dot-pattern-glow origin-center [transform-box:fill-box]"),
					svg.Style("animation-duration:"+secondsRounded(duration)+";animation-delay:"+secondsRounded(delay)),
				)
			} else {
				attribs = append(attribs, svg.Fill("currentColor"))
			}
			x := float64(col)*width + cx
			y := float64(row)*height + cy
			// A dot crossing the tile's edge would be cut off there, so it is
			// drawn again on the opposite side, where the next tile shows it.
			for _, dx := range dotPatternWraps(x, r, tileW) {
				for _, dy := range dotPatternWraps(y, r, tileH) {
					pattern.Children = append(pattern.Children,
						svg.Circle(append([]any{svg.CX(x + dx), svg.CY(y + dy)}, attribs...)...))
				}
			}
		}
	}
	var rect mx.Attrib
	if opts.Glow {
		// Chrome repaints a pattern for an animation of its content only
		// when something else on the page animates, so the rect it fills
		// animates imperceptibly.
		rect = svg.Class("animate-dot-pattern-repaint")
	}
	parts := []mx.Component{
		svg.Defs(
			svg.RadialGradient(
				svg.ID(gradientID),
				svg.Stop(svg.Offset("0%"), svg.StopColor("currentColor"), svg.StopOpacity(1)),
				svg.Stop(svg.Offset("100%"), svg.StopColor("currentColor"), svg.StopOpacity(0)),
			),
			pattern,
		),
		svg.Rect(svg.Width("100%"), svg.Height("100%"), svg.Fill("url(#"+id+")"), rect),
	}
	return patternSVG("dot-pattern",
		"pointer-events-none absolute inset-0 h-full w-full text-neutral-400/80",
		parts, attribsChildren)
}

// dotPatternWraps returns the offsets at which a dot at pos with radius r is
// drawn in a tile of the given size: 0, plus ±size if it crosses an edge.
func dotPatternWraps(pos, r, size float64) []float64 {
	wraps := []float64{0}
	if pos-r < 0 {
		wraps = append(wraps, size)
	}
	if pos+r > size {
		wraps = append(wraps, -size)
	}
	return wraps
}
