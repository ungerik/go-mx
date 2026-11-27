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
	// of the path it follows; 0 means 50. Both read --border-beam-size, so
	// under a strict CSP html.Class("[--border-beam-size:80px]") does the
	// same; a size-* or w-* class resizes the beam but not its path.
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
	// pixels; 0 means 1. The border is the outer element, which the
	// caller's classes do not reach, so under a strict CSP set
	// [--border-beam-width:2px] on the BorderBeam's parent instead: the
	// border falls back to 1px only where no ancestor sets it.
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
	// The defaults are classes, so a BorderBeam left at them needs no style
	// attribute (see "Strict Content-Security-Policy" in the README); only
	// the options set go into style.
	var size, offset, from, to, duration, delay string
	if opts.Size != 0 {
		size = px(opts.Size)
	}
	if opts.InitialOffset != 0 || opts.Reverse {
		start, end := opts.InitialOffset, 100+opts.InitialOffset
		if opts.Reverse {
			start, end = 100-opts.InitialOffset, -opts.InitialOffset
		}
		offset, from, to = num(opts.InitialOffset)+"%", num(start)+"%", num(end)+"%"
	}
	if opts.Duration > 0 {
		duration = seconds(opts.Duration, 0)
	}
	if opts.Delay != 0 {
		delay = num((-opts.Delay).Seconds()) + "s"
	}
	beamStyle, err := styleVars(
		"--border-beam-size", size,
		"--color-from", opts.ColorFrom,
		"--color-to", opts.ColorTo,
		"offset-distance", offset,
		"--border-beam-from", from,
		"--border-beam-to", to,
		"--border-beam-duration", duration,
		"--border-beam-delay", delay,
	)
	if err != nil {
		return mx.NewErrElement(err)
	}
	beam := finish(html.Div(attribsChildren...), "border-beam-light",
		"[--border-beam-size:50px] w-(--border-beam-size) [offset-path:rect(0_auto_auto_0_round_var(--border-beam-size))] [--color-from:#ffaa40] [--color-to:#9c40ff] [offset-distance:0%] "+
			"[--border-beam-from:0%] [--border-beam-to:100%] [--border-beam-duration:6s] [--border-beam-delay:0s] "+
			"absolute aspect-square bg-linear-to-l from-(--color-from) via-(--color-to) to-transparent animate-border-beam",
		beamStyle)
	if beam.Err != nil {
		return beam
	}
	var borderWidth string
	if opts.BorderWidth != 0 {
		borderWidth = px(opts.BorderWidth)
	}
	widthStyle, err := styleVars("--border-beam-width", borderWidth)
	if err != nil {
		return mx.NewErrElement(err)
	}
	// No default class for --border-beam-width: the caller's classes go to
	// the beam, so the border falls back to 1px where no ancestor sets it.
	return finish(html.Div(ariaHidden, beam), "border-beam",
		"pointer-events-none absolute inset-0 rounded-[inherit] border-[length:var(--border-beam-width,1px)] border-transparent mask-[linear-gradient(transparent,transparent),linear-gradient(#000,#000)] mask-intersect [mask-clip:padding-box,border-box]",
		widthStyle)
}
