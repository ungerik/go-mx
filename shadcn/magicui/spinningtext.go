package magicui

import (
	"strconv"
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// SpinningTextOptions configures a [SpinningText]. The zero value is
// upstream's default: one clockwise turn every 10s on a radius of 5ch.
type SpinningTextOptions struct {
	// Duration is how long one turn takes; 0 means 10s.
	Duration time.Duration
	// Reverse turns counterclockwise.
	Reverse bool
	// Radius of the circle in ch; 0 means 5.
	Radius float64
}

// SpinningText lays its text out on a circle around the element's
// position and turns the circle endlessly. The children must be strings,
// as upstream's children: string | string[]; another child makes it render
// an error. Every character is an aria-hidden span rotated into place by
// the CSS variables --index, --total and --spinning-text-radius, followed by a space
// that closes the circle, and an sr-only copy holds the text for assistive
// technology. The element itself takes no space, so place it in a box of
// its own.
//
// Upstream turns the circle with the motion library; here it is a CSS
// keyframe animation on the rotate property, stopped under
// prefers-reduced-motion. Not ported: the transition and variants props.
//
// Needs [CSS] in the app's Tailwind input.
func SpinningText(opts SpinningTextOptions, attribsChildren ...any) *mx.Element {
	root := html.Div(attribsChildren...)
	if root.Err != nil {
		return root
	}
	text, err := childText("SpinningText", root)
	if err != nil {
		return mx.NewErrElement(err)
	}
	letters := append([]rune(text), ' ')
	style, err := styleVars(
		"--spinning-text-duration", seconds(opts.Duration, 10*time.Second),
		"--total", strconv.Itoa(len(letters)),
		"--spinning-text-radius", num(or(opts.Radius, 5)),
	)
	if err != nil {
		return mx.NewErrElement(err)
	}
	root.Children = make([]mx.Component, 0, len(letters)+1)
	for i, letter := range letters {
		root.Children = append(root.Children, html.Span(
			html.Class("absolute top-1/2 left-1/2 inline-block"),
			ariaHidden,
			html.Style("--index:"+strconv.Itoa(i)+";transform:translate(-50%, -50%) rotate(calc(360deg / var(--total) * var(--index))) translateY(calc(var(--spinning-text-radius, 5) * -1ch));transform-origin:center"),
			string(letter),
		))
	}
	root.Children = append(root.Children, html.Span(html.Class("sr-only"), text))
	return finish(root, "spinning-text", shadcn.Cn(
		"relative animate-spinning-text",
		map[string]bool{"[animation-direction:reverse]": opts.Reverse},
	), style)
}
