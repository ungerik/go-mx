package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// AnimatedShinyTextOptions configures an [AnimatedShinyText]. The zero
// value is upstream's default glare width of 100px.
type AnimatedShinyTextOptions struct {
	// ShimmerWidth is the width of the glare in pixels; 0 means 100.
	ShimmerWidth float64
}

// AnimatedShinyText renders its children as muted text with a glare that
// sweeps across it every few seconds. The glare width is the CSS variable
// --shiny-width.
//
// Under prefers-reduced-motion the glare stays outside the text, where
// upstream would leave it over the first letters.
//
// Needs [CSS] in the app's Tailwind input.
func AnimatedShinyText(opts AnimatedShinyTextOptions, attribsChildren ...any) *mx.Element {
	var style string
	if opts.ShimmerWidth > 0 {
		style = "--shiny-width:" + px(opts.ShimmerWidth)
	}
	return finish(html.Span(attribsChildren...), "animated-shiny-text",
		"[--shiny-width:100px] mx-auto max-w-md text-neutral-600/70 dark:text-neutral-400/70 "+
			"animate-shiny-text bg-size-[var(--shiny-width)_100%] bg-clip-text bg-position-[0_0] bg-no-repeat [transition:background-position_1s_cubic-bezier(.6,.6,0,1)_infinite] "+
			"bg-linear-to-r from-transparent via-black/80 via-50% to-transparent dark:via-white/80",
		style,
	)
}
