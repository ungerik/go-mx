//go:generate go -C ../../tools tool go-enum ../shadcn/magicui/$GOFILE

package magicui

import (
	"fmt"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/svg"
)

// StripedPatternDirection is the direction the stripes of a
// [StripedPattern] lean towards.
type StripedPatternDirection string //#enum

const (
	// StripedPatternLeft draws stripes rising from the bottom left to the
	// top right.
	StripedPatternLeft StripedPatternDirection = "left"
	// StripedPatternRight draws stripes falling from the top left to the
	// bottom right.
	StripedPatternRight StripedPatternDirection = "right"
)

// Valid indicates if s is any of the valid values for StripedPatternDirection
func (s StripedPatternDirection) Valid() bool {
	switch s {
	case
		StripedPatternLeft,
		StripedPatternRight:
		return true
	}
	return false
}

// Validate returns an error if s is none of the valid values for StripedPatternDirection
func (s StripedPatternDirection) Validate() error {
	if !s.Valid() {
		return fmt.Errorf("invalid value %#v for type magicui.StripedPatternDirection", s)
	}
	return nil
}

// ValidOr returns s if it is any of the valid values for StripedPatternDirection, else defaultVal
func (s StripedPatternDirection) ValidOr(defaultVal StripedPatternDirection) StripedPatternDirection {
	if s.Valid() {
		return s
	}
	return defaultVal
}

// Enums returns all valid values for StripedPatternDirection
func (StripedPatternDirection) Enums() []StripedPatternDirection {
	return []StripedPatternDirection{
		StripedPatternLeft,
		StripedPatternRight,
	}
}

// EnumStrings returns all valid values for StripedPatternDirection as strings
func (StripedPatternDirection) EnumStrings() []string {
	return []string{
		"left",
		"right",
	}
}

// String implements the fmt.Stringer interface for StripedPatternDirection
func (s StripedPatternDirection) String() string {
	return string(s)
}

// StripedPatternOptions configures a [StripedPattern]. The zero value is
// upstream's default: stripes leaning left in 10×10 pixel tiles.
type StripedPatternOptions struct {
	// ID is the id of the SVG <pattern>, which the background rect references
	// as url(#ID). Empty means a process-unique id from [mx.UniqueID]; pass a
	// fixed one where the markup has to be deterministic. It may only contain
	// letters, digits, '_' and '-'.
	ID string
	// Direction is the direction the stripes lean towards; empty or unknown
	// means [StripedPatternLeft].
	Direction StripedPatternDirection
	// Width is the width of a pattern tile in pixels; 0 means 10.
	Width float64
	// Height is the height of a pattern tile in pixels; 0 means 10.
	Height float64
}

// StripedPattern draws diagonal stripes as an absolutely positioned SVG
// background that fills its positioned parent. The stripes are stroked with
// currentColor, so a text color class recolors them, and a stroke-width or
// stroke-dasharray class such as html.Class("stroke-[0.3]
// [stroke-dasharray:8,4]") thins or dashes them. Children are appended
// inside the <svg>, after the pattern. The SVG is aria-hidden.
func StripedPattern(opts StripedPatternOptions, attribsChildren ...any) *mx.Element {
	id, err := svgRefID(opts.ID)
	if err != nil {
		return mx.NewErrElement(err)
	}
	w := or(opts.Width, 10)
	h := or(opts.Height, 10)
	// Three lines per tile: the diagonal and its neighbors on either side,
	// whose ends reach into the tile's corners.
	var y1, y2 float64 = h, 0
	if opts.Direction.ValidOr(StripedPatternLeft) == StripedPatternRight {
		y1, y2 = 0, h
	}
	line := func(x1, x2 float64) mx.Component {
		return svg.Line(svg.X1(x1), svg.Y1(y1), svg.X2(x2), svg.Y2(y2), svg.Stroke("currentColor"))
	}
	parts := []mx.Component{
		svg.Defs(
			svg.Pattern(
				svg.ID(id),
				svg.Width(w),
				svg.Height(h),
				svg.PatternUnitsUserSpaceOnUse,
				line(0, w),
				line(-w, 0),
				line(w, w*2),
			),
		),
		svg.Rect(svg.Width("100%"), svg.Height("100%"), svg.Fill("url(#"+id+")")),
	}
	return patternSVG("striped-pattern",
		"pointer-events-none absolute inset-0 z-10 h-full w-full stroke-[0.5]",
		parts, attribsChildren)
}
