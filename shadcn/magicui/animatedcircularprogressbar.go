package magicui

import (
	"fmt"
	"math"
	"strconv"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/svg"
)

// AnimatedCircularProgressBarOptions configures an
// [AnimatedCircularProgressBar]. The zero value is a gauge from 0 to 100
// showing 0.
type AnimatedCircularProgressBarOptions struct {
	// Max is the value of a full gauge; nil means 100.
	Max *float64
	// Min is the value of an empty gauge.
	Min float64
	// Value is the current value, shown as a rounded percentage of Min..Max.
	Value float64
	// GaugePrimaryColor is the CSS color of the filled arc; empty means
	// var(--primary). Upstream requires it.
	GaugePrimaryColor string
	// GaugeSecondaryColor is the CSS color of the remaining arc; empty means
	// var(--muted). Upstream requires it.
	GaugeSecondaryColor string
}

// AnimatedCircularProgressBar renders a circular gauge: an arc for the
// percentage of opts.Value between opts.Min and opts.Max, the rest of the
// circle as a second arc separated by gaps, and the percentage as text in
// the center. As upstream, everything is driven by CSS variables on the
// root (--circumference, --percent-to-px, --gap-percent, --transition-length,
// --delay, …) and --stroke-percent on each arc, and the arcs transition
// between values: the transition plays when a value is updated in place,
// e.g. by a script or a morphing swap setting the arcs' --stroke-percent,
// while a freshly rendered gauge shows its value at once. The number fades
// in once on render.
//
// Needs [CSS] in the app's Tailwind input.
func AnimatedCircularProgressBar(opts AnimatedCircularProgressBarOptions, attribsChildren ...any) *mx.Element {
	maximum := derefOr(opts.Max, 100)
	if maximum == opts.Min {
		return mx.NewErrElement(fmt.Errorf("magicui: AnimatedCircularProgressBar Max and Min are both %v", maximum))
	}
	circumference := 2 * math.Pi * 45
	percentPx := circumference / 100
	// Math.round rounds halves up, towards +Inf.
	currentPercent := int(math.Floor((opts.Value-opts.Min)/(maximum-opts.Min)*100 + 0.5))

	rootStyle, err := styleVars(
		"--circle-size", "100px",
		"--circumference", num(circumference),
		"--percent-to-px", px(percentPx),
		"--gap-percent", "5",
		"--offset-factor", "0",
		"--transition-length", "1s",
		"--transition-step", "200ms",
		"--delay", "0s",
		"--percent-to-deg", "3.6deg",
		"transform", "translateZ(0)",
	)
	if err != nil {
		return mx.NewErrElement(err)
	}
	gauge := svg.SVG(
		svg.Fill("none"),
		svg.Class("size-full"),
		svg.StrokeWidth(2),
		svg.ViewBox(0, 0, 100, 100),
	)
	if currentPercent <= 90 && currentPercent >= 0 {
		style, err := styleVars(
			"stroke", or(opts.GaugeSecondaryColor, "var(--muted)"),
			"--stroke-percent", strconv.Itoa(90-currentPercent),
			"--offset-factor-secondary", "calc(1 - var(--offset-factor))",
			"stroke-dasharray", "calc(var(--stroke-percent) * var(--percent-to-px)) var(--circumference)",
			"transform", "rotate(calc(1turn - 90deg - (var(--gap-percent) * var(--percent-to-deg) * var(--offset-factor-secondary)))) scaleY(-1)",
			"transition", "all var(--transition-length) ease var(--delay)",
			"transform-origin", "calc(var(--circle-size) / 2) calc(var(--circle-size) / 2)",
		)
		if err != nil {
			return mx.NewErrElement(err)
		}
		gauge.Children = append(gauge.Children, animatedCircularProgressBarArc(style))
	}
	style, err := styleVars(
		"stroke", or(opts.GaugePrimaryColor, "var(--primary)"),
		"--stroke-percent", strconv.Itoa(currentPercent),
		"stroke-dasharray", "calc(var(--stroke-percent) * var(--percent-to-px)) var(--circumference)",
		"transition", "var(--transition-length) ease var(--delay),stroke var(--transition-length) ease var(--delay)",
		"transition-property", "stroke-dasharray,transform",
		"transform", "rotate(calc(-90deg + var(--gap-percent) * var(--offset-factor) * var(--percent-to-deg)))",
		"transform-origin", "calc(var(--circle-size) / 2) calc(var(--circle-size) / 2)",
	)
	if err != nil {
		return mx.NewErrElement(err)
	}
	gauge.Children = append(gauge.Children, animatedCircularProgressBarArc(style))

	root := html.Div(attribsChildren...)
	root.Children = append(root.Children,
		gauge,
		html.Span(
			html.DataAttr("current-value", strconv.Itoa(currentPercent)),
			html.Class("animate-animated-circular-progress-bar-fade absolute inset-0 m-auto size-fit delay-(--delay) duration-(--transition-length) ease-linear"),
			strconv.Itoa(currentPercent),
		),
	)
	return finish(root, "animated-circular-progress-bar", "relative size-40 text-2xl font-semibold", rootStyle)
}

// animatedCircularProgressBarArc is one arc of the gauge, a circle drawn as
// a dash of the length its style computes.
func animatedCircularProgressBarArc(style string) *mx.Element {
	return svg.Circle(
		svg.CX(50),
		svg.CY(50),
		svg.R(45),
		svg.StrokeWidth(10),
		svg.StrokeDashOffset(0),
		svg.StrokeLineCapRound,
		svg.StrokeLineJoinRound,
		svg.Class("opacity-100"),
		svg.Style(style),
	)
}
