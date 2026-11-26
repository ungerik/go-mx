package magicui

import (
	"strconv"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// RippleOptions configures a [Ripple]. The zero value is upstream's
// default: eight circles, the innermost 210 pixels wide at opacity 0.24.
type RippleOptions struct {
	// MainCircleSize is the diameter of the innermost circle in pixels;
	// every further circle is 70 pixels wider. 0 means 210.
	MainCircleSize float64
	// MainCircleOpacity is the opacity of the innermost circle; every
	// further circle is 0.03 more transparent. 0 means 0.24.
	MainCircleOpacity float64
	// NumCircles is the number of circles; 0 means 8.
	NumCircles int
}

// Ripple draws concentric circles that pulse outwards from the center of
// its positioned parent, like ripples on water, as an absolutely positioned
// background faded out towards the bottom. The circles are aria-hidden;
// children are placed before them. The pulse lasts --duration (2s), which a
// caller class such as html.Class("[--duration:3s]") overrides.
//
// Needs [CSS] in the app's Tailwind input.
func Ripple(opts RippleOptions, attribsChildren ...any) *mx.Element {
	root := html.Div(attribsChildren...)
	size := or(opts.MainCircleSize, 210)
	opacity := or(opts.MainCircleOpacity, 0.24)
	for i := range or(opts.NumCircles, 8) {
		circleSize := px(size + float64(i)*70)
		root.Children = append(root.Children, html.Div(
			ariaHidden,
			html.Class("animate-ripple bg-foreground/25 absolute rounded-full border shadow-xl"),
			html.Style("--i:"+strconv.Itoa(i)+
				";width:"+circleSize+
				";height:"+circleSize+
				";opacity:"+numRounded(opacity-float64(i)*0.03)+
				";animation-delay:"+numRounded(float64(i)*0.06)+"s"+
				";border-style:solid;border-width:1px;border-color:var(--foreground)"+
				";top:50%;left:50%;transform:translate(-50%, -50%) scale(1)"),
		))
	}
	return finish(root, "ripple",
		"pointer-events-none absolute inset-0 mask-[linear-gradient(to_bottom,white,transparent)] select-none", "")
}
