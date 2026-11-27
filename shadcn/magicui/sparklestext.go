package magicui

import (
	"fmt"
	"math"
	"strings"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/svg"
)

// SparklesTextColors are the two colors a [SparklesText]'s sparkles are
// drawn in, CSS colors.
type SparklesTextColors struct {
	// First is the color of about half of the sparkles; "" means "#9E7AFF".
	First string
	// Second is the color of the others; "" means "#FE8BBB".
	Second string
}

// SparklesTextOptions configures a [SparklesText]. The zero value is
// upstream's default: ten purple and pink sparkles.
type SparklesTextOptions struct {
	// SparklesCount is the number of sparkles; 0 means 10, and a negative
	// value is an error.
	SparklesCount int
	// Colors of the sparkles.
	Colors SparklesTextColors
	// Seed draws the sparkles' positions, colors, sizes and delays; the
	// same seed renders the same markup.
	Seed uint64
}

// sparklesTextPath is upstream's four-pointed star in a 21x21 viewBox.
const sparklesTextPath = "M9.82531 0.843845C10.0553 0.215178 10.9446 0.215178 11.1746 0.843845L11.8618 2.72026C12.4006 4.19229 12.3916 6.39157 13.5 7.5C14.6084 8.60843 16.8077 8.59935 18.2797 9.13822L20.1561 9.82534C20.7858 10.0553 20.7858 10.9447 20.1561 11.1747L18.2797 11.8618C16.8077 12.4007 14.6084 12.3916 13.5 13.5C12.3916 14.6084 12.4006 16.8077 11.8618 18.2798L11.1746 20.1562C10.9446 20.7858 10.0553 20.7858 9.82531 20.1562L9.13819 18.2798C8.59932 16.8077 8.60843 14.6084 7.5 13.5C6.39157 12.3916 4.19225 12.4007 2.72023 11.8618L0.843814 11.1747C0.215148 10.9447 0.215148 10.0553 0.843814 9.82534L2.72023 9.13822C4.19225 8.59935 6.39157 8.60843 7.5 7.5C8.60843 6.39157 8.59932 4.19229 9.13819 2.72026L9.82531 0.843845Z"

// SparklesText renders its children in bold with sparkles twinkling over
// them: stars at positions drawn from opts.Seed, each growing, turning and
// fading out again every 0.8s after a delay of its own. The colors are the
// CSS variables --sparkles-first-color and --sparkles-second-color. The
// sparkles are aria-hidden; under prefers-reduced-motion they do not show.
//
// Upstream animates the sparkles with the motion library and replaces
// every sparkle with a new random one after a lifespan of 5 to 15 seconds;
// here the sparkles keep their places, so there is no script. Not ported:
// the as prop, which upstream declares but does not use.
//
// Needs [CSS] in the app's Tailwind input.
func SparklesText(opts SparklesTextOptions, attribsChildren ...any) *mx.Element {
	if opts.SparklesCount < 0 {
		return mx.NewErrElement(fmt.Errorf("magicui: SparklesText SparklesCount %d is negative", opts.SparklesCount))
	}
	style, err := styleVars(
		"--sparkles-first-color", or(opts.Colors.First, "#9E7AFF"),
		"--sparkles-second-color", or(opts.Colors.Second, "#FE8BBB"),
	)
	if err != nil {
		return mx.NewErrElement(err)
	}
	root := html.Div(attribsChildren...)
	if root.Err != nil {
		return root
	}
	rng := newRand(opts.Seed)
	count := or(opts.SparklesCount, 10)
	content := make([]any, 0, count+2)
	content = append(content, html.Class("relative inline-block"))
	for range count {
		x, y := rng.Float64()*100, rng.Float64()*100
		color := "--sparkles-second-color"
		if rng.Float64() > 0.5 {
			color = "--sparkles-first-color"
		}
		delay, scale := rng.Float64()*2, rng.Float64()+0.3
		content = append(content, svg.SVG(
			svg.Class("pointer-events-none absolute z-20 opacity-0 animate-sparkles-text"),
			ariaHidden,
			svg.Style(strings.Join([]string{
				"left:" + sparklesTextRound(x) + "%",
				"top:" + sparklesTextRound(y) + "%",
				"color:var(" + color + ")",
				"--sparkles-text-scale:" + sparklesTextRound(scale),
				"--sparkles-text-delay:" + sparklesTextRound(delay) + "s",
			}, ";")),
			svg.Width(21),
			svg.Height(21),
			svg.ViewBox(0, 0, 21, 21),
			svg.Path(svg.D(sparklesTextPath), svg.Fill("currentColor")),
		))
	}
	content = append(content, html.Strong(anys(root.Children)...))
	root.Children = []mx.Component{html.Span(content...)}
	return finish(root, "sparkles-text", "text-6xl font-bold", style)
}

// sparklesTextRound rounds a drawn value to three decimals for the markup.
func sparklesTextRound(f float64) string {
	return num(math.Round(f*1000) / 1000)
}
