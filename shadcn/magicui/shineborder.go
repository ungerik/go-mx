package magicui

import (
	"strings"
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// ShineBorderOptions configures a [ShineBorder]. The zero value is
// upstream's default: a 1px black shine that circles in 14s.
type ShineBorderOptions struct {
	// BorderWidth is the width of the border in pixels; 0 means 1.
	BorderWidth float64
	// Duration is how long the shine takes to travel around; 0 means 14s.
	Duration time.Duration
	// ShineColors are the colors of the shine, one or more CSS colors;
	// empty means "#000000".
	ShineColors []string
}

// ShineBorder draws an animated shine along the border of its positioned
// parent: an absolutely positioned, pointer-transparent layer filling the
// parent (rounded-[inherit]) whose radial gradient is masked down to a
// ring of opts.BorderWidth. Put it first inside a relative, overflow-hidden
// container such as a [shadcn.Card]. A color that could escape its CSS
// declaration makes it render an error.
//
// The layer is aria-hidden, which upstream does not set. It only moves
// under prefers-reduced-motion: no-preference.
//
// Needs [CSS] in the app's Tailwind input.
func ShineBorder(opts ShineBorderOptions, attribsChildren ...any) *mx.Element {
	colors := opts.ShineColors
	if len(colors) == 0 {
		colors = []string{"#000000"}
	}
	for _, c := range colors {
		if err := checkCSSValue("ShineColors", c); err != nil {
			return mx.NewErrElement(err)
		}
	}
	borderWidth := opts.BorderWidth
	if borderWidth <= 0 {
		borderWidth = 1
	}
	style, err := styleVars(
		"--border-width", px(borderWidth),
		"--duration", seconds(opts.Duration, 14*time.Second),
		"background-image", "radial-gradient(transparent,transparent, "+strings.Join(colors, ",")+",transparent,transparent)",
		"background-size", "300% 300%",
		"mask", "linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0)",
		"-webkit-mask", "linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0)",
		"-webkit-mask-composite", "xor",
		"mask-composite", "exclude",
		"padding", "var(--border-width)",
	)
	if err != nil {
		return mx.NewErrElement(err)
	}
	root := html.Div(append([]any{ariaHidden}, attribsChildren...)...)
	return finish(root, "shine-border",
		"motion-safe:animate-shine pointer-events-none absolute inset-0 size-full rounded-[inherit] will-change-[background-position]",
		style,
	)
}
