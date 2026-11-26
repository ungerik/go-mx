package magicui

import (
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// LightRaysOptions configures a [LightRays]. The zero value is upstream's
// default: seven pale blue rays, 70vh long, blurred by 36 pixels, in a 14
// second cycle.
type LightRaysOptions struct {
	// Count is the number of rays; 0 means 7, a negative count none.
	Count int
	// Color is the CSS color of the rays and the ambient glow; empty means
	// "rgba(160, 210, 255, 0.2)".
	Color string
	// Blur is the blur radius of the rays in pixels; 0 means 36.
	Blur float64
	// Speed is the base duration of a ray's cycle; each ray takes 0.75 to
	// 1.25 times as long. 0 means 14s; anything below 100ms means 100ms.
	Speed time.Duration
	// Length is the CSS length of the rays; empty means "70vh".
	Length string
	// Seed seeds the random positions, angles, widths and timings of the
	// rays; the same seed renders the same markup.
	Seed uint64
}

// LightRays fills its positioned parent with soft rays of light shining
// from above: blurred, screen-blended beams that fade in and out while
// swinging gently, over two faint radial glows. Give the content in front
// of it a z-index, as in html.Class("relative z-10"). The rays are
// aria-hidden; children are placed after them.
//
// The rays' motion-library animation is re-expressed as a CSS animation, with
// upstream's pause between cycles (a tenth of the duration) folded into
// the keyframes. The ray's resting style is its brightest, centered state
// instead of upstream's opacity-0, so the rays stay visible when
// prefers-reduced-motion switches the animation off.
//
// Needs [CSS] in the app's Tailwind input.
func LightRays(opts LightRaysOptions, attribsChildren ...any) *mx.Element {
	style, err := styleVars(
		"--light-rays-color", or(opts.Color, "rgba(160, 210, 255, 0.2)"),
		"--light-rays-blur", px(or(opts.Blur, 36)),
		"--light-rays-length", or(opts.Length, "70vh"),
	)
	if err != nil {
		return mx.NewErrElement(err)
	}
	cycle := or(opts.Speed, 14*time.Second).Seconds()
	cycle = max(cycle, 0.1)
	rnd := newRand(opts.Seed)

	layer := html.Div(ariaHidden, html.Class("absolute inset-0 overflow-hidden"),
		html.Div(html.Class("absolute inset-0 opacity-60"),
			html.Style("background:radial-gradient(circle at 20% 15%, color-mix(in srgb, var(--light-rays-color) 45%, transparent), transparent 70%)")),
		html.Div(html.Class("absolute inset-0 opacity-60"),
			html.Style("background:radial-gradient(circle at 80% 10%, color-mix(in srgb, var(--light-rays-color) 35%, transparent), transparent 75%)")),
	)
	for range or(opts.Count, 7) {
		left := 8 + rnd.Float64()*84
		rotate := numRounded(-28+rnd.Float64()*56) + "deg"
		width := 160 + rnd.Float64()*160
		swing := 0.8 + rnd.Float64()*1.8
		delay := rnd.Float64() * cycle
		duration := cycle * (0.75 + rnd.Float64()*0.5)
		intensity := numRounded(0.6 + rnd.Float64()*0.5)
		layer.Children = append(layer.Children, html.Div(
			html.Class("animate-light-rays-ray pointer-events-none absolute -top-[12%] left-[var(--ray-left)] h-[var(--light-rays-length)] w-[var(--ray-width)] origin-top -translate-x-1/2 rounded-full bg-linear-to-b from-[color-mix(in_srgb,var(--light-rays-color)_70%,transparent)] to-transparent mix-blend-screen blur-[var(--light-rays-blur)]"),
			html.Style("--ray-left:"+numRounded(left)+"%"+
				";--ray-width:"+numRounded(width)+"px"+
				";--ray-rotate:"+rotate+
				";--ray-swing:"+numRounded(swing)+"deg"+
				";--ray-intensity:"+intensity+
				";opacity:"+intensity+
				";rotate:"+rotate+
				// The repeat delay of a tenth of the duration is part of
				// the keyframes, so the animation runs 1.1 times as long.
				";animation-duration:"+secondsRounded(duration*1.1)+
				";animation-delay:"+secondsRounded(delay)),
		))
	}
	root := html.Div(attribsChildren...)
	root.Children = append([]mx.Component{layer}, root.Children...)
	return finish(root, "light-rays",
		"pointer-events-none absolute inset-0 isolate overflow-hidden rounded-[inherit]", style)
}
