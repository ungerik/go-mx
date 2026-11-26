package magicui

import (
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
	"github.com/ungerik/go-mx/svg"
)

// OrbitingCirclesOptions configures [OrbitingCircles]. The zero value is
// upstream's default: a 160px orbit drawn as a faint circle, travelled in
// 20 seconds by 30px items.
type OrbitingCirclesOptions struct {
	// Reverse orbits counter-clockwise instead of clockwise.
	Reverse bool
	// Duration is the time of one full orbit at speed 1; 0 means 20s.
	Duration time.Duration
	// Radius is the orbit's radius in pixels; 0 means 160.
	Radius float64
	// HidePath leaves out the circle that draws the orbit.
	HidePath bool
	// IconSize is the width and height of each orbiting item in pixels;
	// 0 means 30.
	IconSize float64
	// Speed divides Duration, so 2 orbits twice as fast; 0 means 1.
	Speed float64
}

// OrbitingCircles moves each of its children along a circular orbit around
// the center of the nearest positioned ancestor, spaced evenly by angle. The
// caller's attributes (classes, style, …) go to every orbiting item, as
// upstream spreads its props onto each of them, so do not pass an id. The
// items and the orbit path are wrapped in a display:contents element that
// carries data-slot="orbiting-circles", so the wrapper adds no box.
//
// Upstream declares a delay prop that it never applies; it is not ported.
//
// Needs [CSS] in the app's Tailwind input.
func OrbitingCircles(opts OrbitingCirclesOptions, attribsChildren ...any) *mx.Element {
	split := html.Div(attribsChildren...)
	if split.Err != nil {
		return split
	}
	radius := or(opts.Radius, 160)
	speed := or(opts.Speed, 1)
	duration := or(opts.Duration, 20*time.Second).Seconds() / speed
	itemClasses := shadcn.Cn(
		"animate-orbit absolute flex size-(--icon-size) transform-gpu items-center justify-center rounded-full",
		map[string]bool{"[animation-direction:reverse]": opts.Reverse},
	)
	callerAttribs := make([]any, len(split.Attribs))
	for i, a := range split.Attribs {
		callerAttribs[i] = a
	}

	wrapper := html.Div()
	if !opts.HidePath {
		wrapper.Children = append(wrapper.Children, svg.SVG(
			svg.Version("1.1"),
			svg.Class("pointer-events-none absolute inset-0 size-full"),
			ariaHidden,
			svg.Circle(
				svg.Class("stroke-black/10 stroke-1 dark:stroke-white/10"),
				svg.CX("50%"),
				svg.CY("50%"),
				svg.R(radius),
				svg.Fill("none"),
			),
		))
	}
	items := animatedListItems(split.Children, nil)
	for i, child := range items {
		angle := 360 / float64(len(items)) * float64(i)
		style, err := styleVars(
			"--duration", num(duration),
			"--radius", num(radius),
			"--angle", num(angle),
			"--icon-size", px(or(opts.IconSize, 30)),
		)
		if err != nil {
			return mx.NewErrElement(err)
		}
		item := finish(html.Div(append(callerAttribs, child)...), "orbiting-circles-item", itemClasses, style)
		if item.Err != nil {
			return item
		}
		wrapper.Children = append(wrapper.Children, item)
	}
	return finish(wrapper, "orbiting-circles", "contents", "")
}
