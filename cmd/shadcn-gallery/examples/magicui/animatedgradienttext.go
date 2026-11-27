package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
	"github.com/ungerik/go-mx/shadcn/magicui"
	"github.com/ungerik/go-mx/svg"
)

// animatedGradientTextChevron is lucide's chevron-right icon with the
// given classes.
func animatedGradientTextChevron(class string) mx.Component {
	return svg.SVG(
		svg.XMLNS, svg.Width(24), svg.Height(24), svg.ViewBox(0, 0, 24, 24),
		svg.Fill("none"), svg.Stroke("currentColor"), svg.StrokeWidth(2),
		svg.StrokeLineCapRound, svg.StrokeLineJoinRound,
		svg.Class("lucide lucide-chevron-right", class),
		svg.Path(svg.D("m9 18 6-6-6-6")),
	)
}

// AnimatedGradientTextDemo is an announcement pill with a gradient border
// and gradient text.
func AnimatedGradientTextDemo() mx.Component {
	return html.DivClass("group relative mx-auto flex items-center justify-center rounded-full px-4 py-1.5 shadow-[inset_0_-8px_10px_#8fdfff1f] transition-shadow duration-500 ease-out hover:shadow-[inset_0_-5px_10px_#8fdfff3f]",
		html.Span(
			html.Class(shadcn.Cn("animate-gradient absolute inset-0 block h-full w-full rounded-[inherit] bg-gradient-to-r from-[#ffaa40]/50 via-[#9c40ff]/50 to-[#ffaa40]/50 bg-[length:300%_100%] p-[1px]")),
			html.Style("-webkit-mask:linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);-webkit-mask-composite:destination-out;mask:linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);mask-composite:subtract;-webkit-clip-path:padding-box"),
		),
		"🎉 ",
		html.HR(html.Class("mx-2 h-4 w-px shrink-0 bg-neutral-500")),
		magicui.AnimatedGradientText(magicui.AnimatedGradientTextOptions{}, html.Class("text-sm font-medium"),
			"Introducing Magic UI",
		),
		animatedGradientTextChevron("ml-1 size-4 stroke-neutral-500 transition-transform duration-300 ease-in-out group-hover:translate-x-0.5"),
	)
}

// AnimatedGradientTextFast is large gradient text in green and cyan,
// flowing at twice the speed.
func AnimatedGradientTextFast() mx.Component {
	return magicui.AnimatedGradientText(
		magicui.AnimatedGradientTextOptions{Speed: 2, ColorFrom: "#4ade80", ColorTo: "#06b6d4"},
		html.Class("text-4xl font-semibold tracking-tight"),
		"Fast Gradient",
	)
}
