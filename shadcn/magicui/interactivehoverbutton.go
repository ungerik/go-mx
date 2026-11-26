package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/svg"
)

// InteractiveHoverButton renders a pill button with a small dot before its
// label. On hover the dot grows to fill the button and the label slides out
// to the right while a copy of it with an arrow slides in. It needs no
// [CSS] of its own: the effect is Tailwind transitions on group-hover.
//
// The children are rendered twice, so they should not carry ids. Unlike
// upstream, the sliding copy is aria-hidden, so assistive technology reads
// the label once, and like [shadcn.Button] the button defaults to
// type="button"; pass html.Type to override.
func InteractiveHoverButton(attribsChildren ...any) *mx.Element {
	root := html.Button(attribsChildren...)
	if root.AttribIndex("type") < 0 {
		root.Attribs = append(root.Attribs, html.Type("button"))
	}
	children := anys(root.Children)
	root.Children = []mx.Component{
		html.DivClass("flex items-center justify-center gap-2",
			html.DivClass("bg-primary h-2 w-2 rounded-full transition-all duration-300 group-hover:scale-[100.8]"),
			html.Span(append([]any{html.Class("inline-block transition-all duration-300 group-hover:translate-x-12 group-hover:opacity-0")}, children...)...),
		),
		html.Div(
			html.Class("text-primary-foreground absolute top-0 z-10 flex h-full w-full translate-x-12 items-center justify-center gap-2 opacity-0 transition-all duration-300 group-hover:-translate-x-5 group-hover:opacity-100"),
			ariaHidden,
			html.Span(children...),
			interactiveHoverButtonArrow(),
		),
	}
	return finish(root, "interactive-hover-button",
		"group bg-background relative w-auto cursor-pointer overflow-hidden rounded-full border p-2 px-6 text-center font-semibold",
		"",
	)
}

// interactiveHoverButtonArrow is lucide's arrow-right icon with
// lucide-react's default attributes.
func interactiveHoverButtonArrow() *mx.Element {
	return svg.SVG(
		svg.XMLNS, svg.Width(24), svg.Height(24), svg.ViewBox(0, 0, 24, 24),
		svg.Fill("none"), svg.Stroke("currentColor"), svg.StrokeWidth(2),
		svg.StrokeLineCapRound, svg.StrokeLineJoinRound,
		svg.Class("lucide lucide-arrow-right"),
		svg.Path(svg.D("M5 12h14")),
		svg.Path(svg.D("m12 5 7 7-7 7")),
	)
}
