package magicui

import (
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// BorderBeamOptions configures a [BorderBeam]. The zero value is upstream's
// default: a 50px beam from #ffaa40 to #9c40ff that circles a 1px border
// every 6 seconds.
type BorderBeamOptions struct {
	// Size is the beam's width and height in pixels, and the corner radius
	// of the path it follows; 0 means 50.
	Size float64
	// Duration is the time of one full circle; 0 means 6s.
	Duration time.Duration
	// Delay starts the beam this far into its circle: upstream passes it to
	// motion negated, so it becomes a negative animation-delay.
	Delay time.Duration
	// ColorFrom is the CSS color at the head of the beam; empty means
	// #ffaa40.
	ColorFrom string
	// ColorTo is the CSS color in the middle of the beam; empty means
	// #9c40ff.
	ColorTo string
	// Reverse circles the border counter-clockwise.
	Reverse bool
	// InitialOffset is where on the border the beam starts, as a percentage
	// (0–100) of the path.
	InitialOffset float64
	// BorderWidth is the width of the border the beam is visible in, in
	// pixels; 0 means 1.
	BorderWidth float64
}

// BorderBeam is an animated beam of light travelling along the border of its
// container. Place it as the last child of an element with position relative
// (and overflow hidden); it fills the element, takes its border radius, and
// shows the beam only within a border of opts.BorderWidth.
//
// As upstream's className and style, the caller's attributes go to the beam
// itself, so a gradient class such as
// html.Class("from-transparent via-red-500 to-transparent") replaces the
// beam's colors; the outer border mask carries data-slot="border-beam" and
// is aria-hidden. The beam moves along offset-path with a CSS animation of
// offset-distance instead of the motion library, so upstream's transition
// prop (a custom motion easing or spring) is not ported: the beam always
// moves linearly. With reduced motion it rests at its starting point.
//
// Needs [CSS] in the app's Tailwind input.
func BorderBeam(opts BorderBeamOptions, attribsChildren ...any) *mx.Element {
	size := or(opts.Size, 50)
	from, to := opts.InitialOffset, 100+opts.InitialOffset
	if opts.Reverse {
		from, to = 100-opts.InitialOffset, -opts.InitialOffset
	}
	beamStyle, err := styleVars(
		"width", px(size),
		"offset-path", "rect(0 auto auto 0 round "+px(size)+")",
		"--color-from", or(opts.ColorFrom, "#ffaa40"),
		"--color-to", or(opts.ColorTo, "#9c40ff"),
		"offset-distance", num(opts.InitialOffset)+"%",
		"--border-beam-from", num(from)+"%",
		"--border-beam-to", num(to)+"%",
		"--border-beam-duration", seconds(opts.Duration, 6*time.Second),
		"--border-beam-delay", num((-opts.Delay).Seconds())+"s",
	)
	if err != nil {
		return mx.NewErrElement(err)
	}
	beam := finish(html.Div(attribsChildren...), "border-beam-light",
		"absolute aspect-square bg-linear-to-l from-(--color-from) via-(--color-to) to-transparent animate-border-beam",
		beamStyle)
	if beam.Err != nil {
		return beam
	}
	widthStyle, err := styleVars("--border-beam-width", px(or(opts.BorderWidth, 1)))
	if err != nil {
		return mx.NewErrElement(err)
	}
	return finish(html.Div(ariaHidden, beam), "border-beam",
		"pointer-events-none absolute inset-0 rounded-[inherit] border-(length:--border-beam-width) border-transparent mask-[linear-gradient(transparent,transparent),linear-gradient(#000,#000)] mask-intersect [mask-clip:padding-box,border-box]",
		widthStyle)
}
