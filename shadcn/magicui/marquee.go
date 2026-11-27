package magicui

import (
	"fmt"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// MarqueeOptions configures a [Marquee]. The zero value is upstream's
// default: a horizontal marquee repeating its content four times.
type MarqueeOptions struct {
	// Reverse runs the animation in the opposite direction.
	Reverse bool
	// PauseOnHover pauses the animation while the pointer is over it.
	PauseOnHover bool
	// Vertical scrolls the content vertically instead of horizontally.
	Vertical bool
	// Repeat is how many copies of the content scroll by; 0 means 4, and a
	// negative value is an error.
	Repeat int
}

// Marquee scrolls its children endlessly, horizontally or vertically.
// The content is repeated opts.Repeat times; every copy after the first is
// aria-hidden, so assistive technology reads the content once (upstream
// exposes every copy). The speed and the gap come from the CSS variables
// --duration (40s) and --gap (1rem), which a caller class such as
// html.Class("[--duration:20s]") overrides.
//
// Needs [CSS] in the app's Tailwind input.
func Marquee(opts MarqueeOptions, attribsChildren ...any) *mx.Element {
	if opts.Repeat < 0 {
		return mx.NewErrElement(fmt.Errorf("magicui: Marquee Repeat %d is negative", opts.Repeat))
	}
	root := html.Div(attribsChildren...)
	children := root.Children
	repeat := or(opts.Repeat, 4)
	copyClasses := shadcn.Cn(
		"flex shrink-0 justify-around gap-(--gap)",
		map[string]bool{
			"animate-marquee flex-row":                  !opts.Vertical,
			"animate-marquee-vertical flex-col":         opts.Vertical,
			"group-hover:[animation-play-state:paused]": opts.PauseOnHover,
			"[animation-direction:reverse]":             opts.Reverse,
		},
	)
	root.Children = make([]mx.Component, 0, repeat)
	for i := range repeat {
		copyAttribs := []any{html.Class(copyClasses)}
		if i > 0 {
			copyAttribs = append(copyAttribs, ariaHidden)
		}
		root.Children = append(root.Children, html.Div(append(copyAttribs, anys(children)...)...))
	}
	return finish(root, "marquee", shadcn.Cn(
		"group flex gap-(--gap) overflow-hidden p-2 [--duration:40s] [--gap:1rem]",
		map[string]bool{"flex-row": !opts.Vertical, "flex-col": opts.Vertical},
	), "")
}
