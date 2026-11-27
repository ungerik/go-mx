package magicui

import (
	"strconv"
	"strings"
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// GlareHoverOptions configures a [GlareHover]. The zero value is upstream's
// default: a white glare at half opacity sweeping diagonally in 650ms over
// a black background.
type GlareHoverOptions struct {
	// Width is the CSS width of the root, such as "100%" or "320px"; empty
	// leaves it to the size-fit class.
	Width string
	// Height is the CSS height of the root; empty leaves it to the size-fit
	// class.
	Height string
	// Background is the CSS background of the root; empty means "#000".
	Background string
	// Color is the glare color as "#rrggbb" or "#rgb", which is combined
	// with Opacity; any other CSS color is used as it is, without Opacity.
	// Empty means "#ffffff".
	Color string
	// Opacity is the opacity of a hex Color, from 0 to 1; 0 means 0.5.
	Opacity float64
	// Angle is the angle of the glare's gradient in degrees; 0 means -45.
	// Pass 360 for a gradient at 0 degrees.
	Angle float64
	// Size is the size of the glare tile as a percentage of the element;
	// 0 means 250.
	Size float64
	// Duration is how long the glare takes to sweep across; 0 means 650ms.
	Duration time.Duration
	// PlayOnce sweeps the glare only while the pointer enters, and lets it
	// jump back without a transition when the pointer leaves.
	PlayOnce bool
}

// GlareHover sweeps a diagonal glare of light across its children when the
// pointer moves over them, like light glancing off a glossy card. The glare
// is a ::before gradient whose background-position transitions on hover, so
// it needs no JavaScript. Round the root like its child, as in
// html.Class("rounded-xl"), so the glare does not overhang the corners.
func GlareHover(opts GlareHoverOptions, attribsChildren ...any) *mx.Element {
	duration := opts.Duration
	if duration <= 0 {
		duration = 650 * time.Millisecond
	}
	style, err := styleVars(
		"--gh-angle", num(or(opts.Angle, -45))+"deg",
		"--gh-duration", num(float64(duration)/float64(time.Millisecond))+"ms",
		"--gh-size", num(or(opts.Size, 250))+"%",
		"--gh-rgba", glareHoverRGBA(or(opts.Color, "#ffffff"), or(opts.Opacity, 0.5)),
		"background", or(opts.Background, "#000"),
		"width", opts.Width,
		"height", opts.Height,
	)
	if err != nil {
		return mx.NewErrElement(err)
	}
	return finish(html.Div(attribsChildren...), "glare-hover", shadcn.Cn(
		"relative grid size-fit cursor-pointer place-items-center overflow-hidden bg-transparent",
		"before:pointer-events-none before:absolute before:inset-0 before:z-10 before:bg-no-repeat before:content-['']",
		"before:[background-image:linear-gradient(var(--gh-angle),transparent_60%,var(--gh-rgba)_70%,transparent,transparent_100%)]",
		"before:[background-size:var(--gh-size)_var(--gh-size),100%_100%]",
		"before:[background-position:-100%_-100%,0_0]",
		map[string]bool{
			"before:transition-[background-position] before:duration-[var(--gh-duration)] before:ease-in-out":                 !opts.PlayOnce,
			"before:transition-none hover:before:transition-[background-position] hover:before:duration-[var(--gh-duration)]": opts.PlayOnce,
		},
		"hover:before:[background-position:100%_100%,0_0]",
	), style)
}

// glareHoverRGBA converts a "#rrggbb" or "#rgb" color to rgba() with the
// given opacity, and returns any other color unchanged.
func glareHoverRGBA(color string, opacity float64) string {
	hex := strings.Replace(color, "#", "", 1)
	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}
	if len(hex) != 6 {
		return color
	}
	channels := make([]string, 3)
	for i := range channels {
		v, err := strconv.ParseUint(hex[2*i:2*i+2], 16, 8)
		if err != nil {
			return color
		}
		channels[i] = strconv.FormatUint(v, 10)
	}
	return "rgba(" + strings.Join(channels, ",") + "," + num(opacity) + ")"
}
