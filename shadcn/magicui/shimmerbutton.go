package magicui

import (
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// ShimmerButtonOptions configures a [ShimmerButton]. The zero value is
// upstream's default: a black pill with a white spark.
type ShimmerButtonOptions struct {
	// ShimmerColor is the color of the spark; "" means "#ffffff".
	ShimmerColor string
	// ShimmerSize is the width of the shimmering border, a CSS length;
	// "" means "0.05em".
	ShimmerSize string
	// BorderRadius is the CSS border radius; "" means "100px".
	BorderRadius string
	// ShimmerDuration is how long the spark takes to slide across the
	// button; 0 means 3s. One turn of the spark takes twice as long.
	ShimmerDuration time.Duration
	// Background is the CSS background of the button; "" means
	// "rgba(0, 0, 0, 1)".
	Background string
}

// ShimmerButton renders a button with a spark of light travelling around
// its border. The options become the CSS variables --shimmer-color, --cut,
// --radius, --speed and --bg on the button (plus --spread: 90deg, the
// spark's angle), so a caller style or class can still override them.
// A string option that could escape its CSS declaration makes the button
// render an error.
//
// Like [shadcn.Button], it defaults to type="button" (upstream leaves the
// type unset, which submits an enclosing form); pass html.Type to override.
//
// Needs [CSS] in the app's Tailwind input.
func ShimmerButton(opts ShimmerButtonOptions, attribsChildren ...any) *mx.Element {
	style, err := styleVars(
		"--spread", "90deg",
		"--shimmer-color", or(opts.ShimmerColor, "#ffffff"),
		"--radius", or(opts.BorderRadius, "100px"),
		"--speed", seconds(opts.ShimmerDuration, 3*time.Second),
		"--cut", or(opts.ShimmerSize, "0.05em"),
		"--bg", or(opts.Background, "rgba(0, 0, 0, 1)"),
	)
	if err != nil {
		return mx.NewErrElement(err)
	}
	root := html.Button(attribsChildren...)
	if root.AttribIndex("type") < 0 {
		root.Attribs = append(root.Attribs, html.Type("button"))
	}
	children := root.Children
	root.Children = make([]mx.Component, 0, len(children)+3)
	// spark container
	root.Children = append(root.Children, html.Div(
		html.Class("-z-30 blur-[2px] @container-[size] absolute inset-0 overflow-visible"),
		ariaHidden,
		// spark
		html.DivClass("animate-shimmer-slide absolute inset-0 aspect-[1] h-[100cqh] rounded-none [mask:none]",
			// spark before
			html.DivClass("animate-spin-around absolute -inset-full w-auto [translate:0_0] rotate-0 [background:conic-gradient(from_calc(270deg-(var(--spread)*0.5)),transparent_0,var(--shimmer-color)_var(--spread),transparent_var(--spread))]"),
		),
	))
	root.Children = append(root.Children, children...)
	root.Children = append(root.Children,
		// highlight
		html.Div(
			html.Class("absolute inset-0 size-full rounded-2xl px-4 py-1.5 text-sm font-medium shadow-[inset_0_-8px_10px_#ffffff1f] transform-gpu transition-all duration-300 ease-in-out group-hover:shadow-[inset_0_-6px_10px_#ffffff3f] group-active:shadow-[inset_0_-10px_10px_#ffffff3f]"),
			ariaHidden,
		),
		// backdrop
		html.Div(
			html.Class("absolute inset-(--cut) -z-20 [border-radius:var(--radius)] [background:var(--bg)]"),
			ariaHidden,
		),
	)
	return finish(root, "shimmer-button",
		"group relative z-0 flex cursor-pointer items-center justify-center overflow-hidden [border-radius:var(--radius)] border border-white/10 px-6 py-3 whitespace-nowrap text-white [background:var(--bg)] transform-gpu transition-transform duration-300 ease-in-out active:translate-y-px",
		style,
	)
}
