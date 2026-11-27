package magicui

import (
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// BlurFadeOptions configures a [BlurFade]. The zero value is upstream's
// default: a 400ms entrance that moves 6px down out of a 6px blur.
type BlurFadeOptions struct {
	// Duration is how long the entrance takes; 0 means 400ms.
	Duration time.Duration
	// Delay is the time before the entrance starts, on top of the 40ms
	// upstream always waits.
	Delay time.Duration
	// Offset is how far, in pixels, the element moves in; 0 means 6.
	Offset float64
	// Direction is the direction the element moves in; "" means
	// [DirectionDown], so it comes from above.
	Direction Direction
	// InView ties the entrance to scrolling: the element fades in while it
	// scrolls into the viewport, with a scroll-driven animation
	// (animation-timeline: view()), and Duration and Delay do not apply.
	// Browsers without scroll-driven animations show it at once.
	InView bool
	// Blur is the CSS length of the blur the element fades in from; ""
	// means "6px".
	Blur string
}

// BlurFade fades its children in from a blur while they move a few pixels
// in opts.Direction. The element's own style is the final state, so it
// stays visible where the animation does not run.
//
// Upstream animates with the motion library and, with inView, plays the
// timed entrance once the element is 50px inside the viewport, which needs
// JavaScript to observe it; here InView links the entrance to the scroll
// position instead, so an element already in view when the page loads
// shows without an entrance. Not ported: the exit animation (there is no
// unmount), the custom variant prop, and motion props spread to the
// element.
//
// Needs [CSS] in the app's Tailwind input.
func BlurFade(opts BlurFadeOptions, attribsChildren ...any) *mx.Element {
	// The defaults are classes on the root, so a BlurFade left at them needs
	// no style attribute (see "Strict Content-Security-Policy" in the
	// README), and each root resets them, so a nested BlurFade does not
	// inherit its parent's. Only the options set go into style.
	var x, y string
	if opts.Offset != 0 || opts.Direction != "" {
		offset := or(opts.Offset, 6)
		x, y = "0px", "0px"
		switch opts.Direction.ValidOr(DirectionDown) {
		case DirectionDown:
			y = px(-offset)
		case DirectionUp:
			y = px(offset)
		case DirectionRight:
			x = px(-offset)
		case DirectionLeft:
			x = px(offset)
		}
	}
	var duration, delay string
	if opts.Duration > 0 {
		duration = seconds(opts.Duration, 0)
	}
	if opts.Delay != 0 {
		// A negative Delay shortens upstream's 40ms base delay; one that
		// cancels it falls back to the base.
		delay = seconds(40*time.Millisecond+opts.Delay, 40*time.Millisecond)
	}
	style, err := styleVars(
		"--blur-fade-duration", duration,
		"--blur-fade-delay", delay,
		"--blur-fade-blur", opts.Blur,
		"--blur-fade-x", x,
		"--blur-fade-y", y,
	)
	if err != nil {
		return mx.NewErrElement(err)
	}
	animation := "animate-blur-fade-in"
	if opts.InView {
		animation = "animate-blur-fade-in-view"
	}
	return finish(html.Div(attribsChildren...), "blur-fade",
		animation+" [--blur-fade-duration:0.4s] [--blur-fade-delay:0.04s] [--blur-fade-blur:6px] [--blur-fade-x:0px] [--blur-fade-y:-6px]",
		style)
}
