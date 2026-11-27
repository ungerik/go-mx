package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// ShinyButton renders a bordered button whose label and border are swept by
// a diagonal shine every few seconds. It pops in from 80% scale with a
// springy overshoot when it first renders and shrinks to 95% while pressed.
//
// Upstream drives all of this with the motion library. Here it is CSS: the
// shine position is the registered custom property --shiny-button-x
// (upstream animates an unregistered --x, which CSS cannot interpolate, and
// registering the generic name would retype every other --x on the page),
// and the spring curves of the shine and the pop are sampled into linear()
// easings. The press is an active: scale with a CSS transition instead of
// a spring. Upstream's gradients write a Tailwind color opacity inside
// plain CSS (var(--primary)/10%) and calc(var(--x)+20%) without spaces,
// which browsers reject; they are written with color-mix() and a spaced
// calc() here so the shine and the dark-mode glow are visible. Like
// [shadcn.Button] it defaults to type="button"; pass html.Type to override.
//
// Needs [CSS] in the app's Tailwind input.
func ShinyButton(attribsChildren ...any) *mx.Element {
	root := html.Button(attribsChildren...)
	if root.AttribIndex("type") < 0 {
		root.Attribs = append(root.Attribs, html.Type("button"))
	}
	children := anys(root.Children)
	root.Children = []mx.Component{
		html.Span(append([]any{
			html.Class("shiny-button-label relative block size-full text-sm tracking-wide text-[rgb(0,0,0,65%)] uppercase dark:font-light dark:text-[rgb(255,255,255,90%)]"),
		}, children...)...),
		html.Span(
			html.Class("shiny-button-shine absolute inset-0 z-10 block rounded-[inherit] p-px"),
			ariaHidden,
		),
	}
	return finish(root, "shiny-button",
		"animate-shiny-button relative cursor-pointer rounded-lg border px-6 py-2 font-medium backdrop-blur-xl transition-[box-shadow,scale] duration-300 ease-in-out active:scale-95 hover:shadow dark:bg-[radial-gradient(circle_at_50%_0%,color-mix(in_oklab,var(--primary)_10%,transparent)_0%,transparent_60%)] dark:hover:shadow-[0_0_20px_color-mix(in_oklab,var(--primary)_10%,transparent)]",
		"",
	)
}
