package magicui

import (
	"strings"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// AuroraTextOptions configures an [AuroraText]. The zero value is
// upstream's default: a pink, purple, blue and sky gradient at speed 1.
type AuroraTextOptions struct {
	// Colors are the CSS colors of the gradient, which returns to the first
	// one at its end; empty means "#FF0080", "#7928CA", "#0070F3",
	// "#38bdf8".
	Colors []string
	// Speed multiplies the speed of the animation, whose cycle takes
	// 10s / Speed; 0 means 1.
	Speed float64
}

// auroraTextClasses are the root classes of an [AuroraText]: upstream's
// default colors and its 10s cycle as the variables the gradient copy reads.
// They sit on the root, where the caller's classes go, so a caller class
// such as [--aurora-text-duration:5s] replaces them. One literal, so
// Tailwind finds every class in the source.
const auroraTextClasses = "relative inline-block " +
	"[--aurora-text-gradient:linear-gradient(135deg,#FF0080,#7928CA,#0070F3,#38bdf8,#FF0080)] [--aurora-text-duration:10s]"

// AuroraText renders its children as text filled with a slowly swirling,
// tilting color gradient. The children are rendered twice: once sr-only
// for assistive technology and once aria-hidden with the gradient, so they
// should be text or inline markup without ids. A color that could escape
// its CSS declaration makes it render an error.
//
// Upstream passes on only className; here every attribute goes to the
// root span.
//
// Needs [CSS] in the app's Tailwind input.
func AuroraText(opts AuroraTextOptions, attribsChildren ...any) *mx.Element {
	for _, color := range opts.Colors {
		// Checked one by one: a color like "red), radial-gradient(blue" is
		// balanced only together with another.
		if err := checkCSSValue("AuroraText color", color); err != nil {
			return mx.NewErrElement(err)
		}
	}
	// The defaults are in auroraTextClasses, so the component left at them
	// needs no style attribute (see "Strict Content-Security-Policy" in the
	// README); only the options set go into the root's style.
	var gradient, duration string
	if len(opts.Colors) > 0 {
		gradient = "linear-gradient(135deg, " + strings.Join(opts.Colors, ", ") + ", " + opts.Colors[0] + ")"
	}
	if opts.Speed > 0 {
		duration = num(10/opts.Speed) + "s"
	}
	style, err := styleVars(
		"--aurora-text-gradient", gradient,
		"--aurora-text-duration", duration,
	)
	if err != nil {
		return mx.NewErrElement(err)
	}
	root := html.Span(attribsChildren...)
	children := anys(root.Children)
	root.Children = []mx.Component{
		html.Span(append([]any{html.Class("sr-only")}, children...)...),
		html.Span(append([]any{
			html.Class("animate-aurora relative bg-size-[200%_auto] bg-(image:--aurora-text-gradient) bg-clip-text text-transparent [-webkit-text-fill-color:transparent]"),
			ariaHidden,
		}, children...)...),
	}
	return finish(root, "aurora-text", auroraTextClasses, style)
}
