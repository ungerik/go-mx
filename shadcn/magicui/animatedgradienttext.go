package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// AnimatedGradientTextOptions configures an [AnimatedGradientText]. The
// zero value is upstream's default: orange to purple at speed 1.
type AnimatedGradientTextOptions struct {
	// Speed scales the gradient's width, so the colors pass by faster:
	// the background is Speed × 300% wide; 0 means 1.
	Speed float64
	// ColorFrom is the CSS color at both ends of the gradient; "" means
	// "#ffaa40".
	ColorFrom string
	// ColorTo is the CSS color in the middle of the gradient; "" means
	// "#9c40ff".
	ColorTo string
}

// AnimatedGradientText renders its children as text filled with a
// gradient that keeps flowing from left to right. The options become the
// CSS variables --bg-size, --color-from and --color-to; a color that could
// escape its CSS declaration makes it render an error.
//
// Needs [CSS] in the app's Tailwind input.
func AnimatedGradientText(opts AnimatedGradientTextOptions, attribsChildren ...any) *mx.Element {
	speed := opts.Speed
	if speed <= 0 {
		speed = 1
	}
	style, err := styleVars(
		"--bg-size", num(speed*300)+"%",
		"--color-from", or(opts.ColorFrom, "#ffaa40"),
		"--color-to", or(opts.ColorTo, "#9c40ff"),
	)
	if err != nil {
		return mx.NewErrElement(err)
	}
	return finish(html.Span(attribsChildren...), "animated-gradient-text",
		"animate-gradient inline bg-linear-to-r from-(--color-from) via-(--color-to) to-(--color-from) bg-size-[var(--bg-size)_100%] bg-clip-text text-transparent",
		style,
	)
}
