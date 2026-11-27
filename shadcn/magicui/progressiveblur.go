//go:generate go -C ../../tools tool go-enum ../shadcn/magicui/$GOFILE

package magicui

import (
	"fmt"
	"strconv"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// ProgressiveBlurPosition is the edge of its parent a [ProgressiveBlur]
// covers.
type ProgressiveBlurPosition string //#enum

const (
	// ProgressiveBlurTop blurs the top edge, strongest at the very top.
	ProgressiveBlurTop ProgressiveBlurPosition = "top"
	// ProgressiveBlurBottom blurs the bottom edge, strongest at the very
	// bottom.
	ProgressiveBlurBottom ProgressiveBlurPosition = "bottom"
	// ProgressiveBlurBoth covers the whole parent, with the blur layers
	// faded in and out only at the top and bottom 5%.
	ProgressiveBlurBoth ProgressiveBlurPosition = "both"
)

// Valid indicates if p is any of the valid values for ProgressiveBlurPosition
func (p ProgressiveBlurPosition) Valid() bool {
	switch p {
	case
		ProgressiveBlurTop,
		ProgressiveBlurBottom,
		ProgressiveBlurBoth:
		return true
	}
	return false
}

// Validate returns an error if p is none of the valid values for ProgressiveBlurPosition
func (p ProgressiveBlurPosition) Validate() error {
	if !p.Valid() {
		return fmt.Errorf("invalid value %#v for type magicui.ProgressiveBlurPosition", p)
	}
	return nil
}

// ValidOr returns p if it is any of the valid values for ProgressiveBlurPosition, else defaultVal
func (p ProgressiveBlurPosition) ValidOr(defaultVal ProgressiveBlurPosition) ProgressiveBlurPosition {
	if p.Valid() {
		return p
	}
	return defaultVal
}

// Enums returns all valid values for ProgressiveBlurPosition
func (ProgressiveBlurPosition) Enums() []ProgressiveBlurPosition {
	return []ProgressiveBlurPosition{
		ProgressiveBlurTop,
		ProgressiveBlurBottom,
		ProgressiveBlurBoth,
	}
}

// EnumStrings returns all valid values for ProgressiveBlurPosition as strings
func (ProgressiveBlurPosition) EnumStrings() []string {
	return []string{
		"top",
		"bottom",
		"both",
	}
}

// String implements the fmt.Stringer interface for ProgressiveBlurPosition
func (p ProgressiveBlurPosition) String() string {
	return string(p)
}

// ProgressiveBlurOptions configures a [ProgressiveBlur]. The zero value is
// upstream's default: the bottom 30% of the parent, blurred in eight steps
// from 0.5 to 64 pixels.
type ProgressiveBlurOptions struct {
	// Height is the CSS height of the blurred area, such as "40%" or "8rem";
	// empty means "30%". It is ignored for [ProgressiveBlurBoth], which
	// covers the whole parent.
	Height string
	// Position is the edge the blur covers; empty or unknown means
	// [ProgressiveBlurBottom].
	Position ProgressiveBlurPosition
	// BlurLevels are the blur radii in pixels, from the weakest to the
	// strongest, one backdrop-filter layer each; nil means
	// 0.5, 1, 2, 4, 8, 16, 32, 64. It needs at least two levels.
	BlurLevels []float64
}

// progressiveBlurDefaultLevels are upstream's default blur levels.
var progressiveBlurDefaultLevels = []float64{0.5, 1, 2, 4, 8, 16, 32, 64}

// ProgressiveBlur blurs the content behind an edge of its positioned
// parent, more strongly towards the edge, by stacking backdrop-filter
// layers whose masks each let one band of the blur through, like a
// scrolling list fading out under a frosted edge. It is aria-hidden and
// lets the pointer through. Children are placed after the blur layers.
//
// The position class (top-0, bottom-0 or inset-y-0) is a base class here,
// so a caller class can override it; upstream applies it after the caller's
// className.
func ProgressiveBlur(opts ProgressiveBlurOptions, attribsChildren ...any) *mx.Element {
	levels := opts.BlurLevels
	if levels == nil {
		levels = progressiveBlurDefaultLevels
	}
	if len(levels) < 2 {
		return mx.NewErrElement(fmt.Errorf("magicui: ProgressiveBlur needs at least 2 blur levels, got %d", len(levels)))
	}
	position := opts.Position.ValidOr(ProgressiveBlurBottom)
	height := or(opts.Height, "30%")
	if position == ProgressiveBlurBoth {
		height = "100%"
	}
	style, err := styleVars("height", height)
	if err != nil {
		return mx.NewErrElement(err)
	}

	const both = "linear-gradient(rgba(0,0,0,0) 0%, rgba(0,0,0,1) 5%, rgba(0,0,0,1) 95%, rgba(0,0,0,0) 100%)"
	// mask returns the gradient that shows the band from start to end
	// percent of the blurred area, measured from its outer side; with
	// position both, every layer shows the whole area.
	mask := func(stops string) string {
		switch position {
		case ProgressiveBlurBottom:
			return "linear-gradient(to bottom, " + stops + ")"
		case ProgressiveBlurTop:
			return "linear-gradient(to top, " + stops + ")"
		}
		return both
	}
	layer := func(zIndex int, blur float64, mask string) mx.Component {
		filter := "blur(" + px(blur) + ")"
		return html.Div(html.Class("absolute inset-0"), html.Style(
			"z-index:"+strconv.Itoa(zIndex)+
				";backdrop-filter:"+filter+
				";-webkit-backdrop-filter:"+filter+
				";mask-image:"+mask+
				";-webkit-mask-image:"+mask))
	}

	layers := []mx.Component{
		layer(1, levels[0], mask("rgba(0,0,0,0) 0%, rgba(0,0,0,1) 12.5%, rgba(0,0,0,1) 25%, rgba(0,0,0,0) 37.5%")),
	}
	for index := range len(levels) - 2 {
		blurIndex := index + 1
		start := float64(blurIndex) * 12.5
		mid := float64(blurIndex+1) * 12.5
		end := float64(blurIndex+2) * 12.5
		layers = append(layers, layer(index+2, levels[blurIndex], mask(fmt.Sprintf(
			"rgba(0,0,0,0) %s%%, rgba(0,0,0,1) %s%%, rgba(0,0,0,1) %s%%, rgba(0,0,0,0) %s%%",
			num(start), num(mid), num(end), num(end+12.5)))))
	}
	layers = append(layers, layer(len(levels), levels[len(levels)-1], mask("rgba(0,0,0,0) 87.5%, rgba(0,0,0,1) 100%")))

	root := html.Div(append([]any{ariaHidden}, attribsChildren...)...)
	root.Children = append(layers, root.Children...)
	positionClass := map[ProgressiveBlurPosition]string{
		ProgressiveBlurTop:    "top-0",
		ProgressiveBlurBottom: "bottom-0",
		ProgressiveBlurBoth:   "inset-y-0",
	}[position]
	return finish(root, "progressive-blur",
		"gradient-blur pointer-events-none absolute inset-x-0 z-10 "+positionClass, style)
}
