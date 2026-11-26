package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// ScrollProgress is a thin gradient bar fixed to the top of the viewport
// that grows from the left as the page scrolls down, showing how much of
// it has been read. A caller class such as html.Class("top-[65px]") moves
// it below a fixed navigation bar.
//
// Upstream scales the bar with the motion library's useScroll; here a
// scroll-driven animation (animation-timeline: scroll(root)) does it
// without JavaScript. Browsers without scroll-driven animations, and
// prefers-reduced-motion, hide the bar rather than show it at a width
// that does not match the scroll position.
//
// Needs [CSS] in the app's Tailwind input.
func ScrollProgress(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "scroll-progress",
		"fixed inset-x-0 top-0 z-50 h-px origin-left bg-linear-to-r from-[#A97CF8] via-[#F38CB8] to-[#FDCC92] animate-scroll-progress",
		"",
	)
}
