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
	for _, c := range opts.ShineColors {
		if err := checkCSSValue("ShineColors", c); err != nil {
			return mx.NewErrElement(err)
		}
	}
	// The defaults and the fixed declarations are classes, so a ShineBorder
	// left at its defaults needs no style attribute (see "Strict
	// Content-Security-Policy" in the README); only the options set go into
	// style.
	var borderWidth, duration, gradient string
	if opts.BorderWidth > 0 {
		borderWidth = px(opts.BorderWidth)
	}
	if opts.Duration > 0 {
		duration = seconds(opts.Duration, 0)
	}
	if len(opts.ShineColors) > 0 {
		gradient = "radial-gradient(transparent,transparent, " + strings.Join(opts.ShineColors, ",") + ",transparent,transparent)"
	}
	style, err := styleVars(
		"--border-width", borderWidth,
		"--duration", duration,
		"background-image", gradient,
	)
	if err != nil {
		return mx.NewErrElement(err)
	}
	root := html.Div(append([]any{ariaHidden}, attribsChildren...)...)
	return finish(root, "shine-border",
		"[--border-width:1px] [--duration:14s] bg-[radial-gradient(transparent,transparent,#000000,transparent,transparent)] bg-size-[300%_300%] "+
			"shine-border-mask "+
			"motion-safe:animate-shine pointer-events-none absolute inset-0 size-full rounded-[inherit] will-change-[background-position]",
		style,
	)
}
