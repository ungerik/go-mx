package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn/magicui"
)

// ShimmerButtonDemo is a black pill button with a white spark circling its
// border.
func ShimmerButtonDemo() mx.Component {
	return magicui.ShimmerButton(magicui.ShimmerButtonOptions{}, html.Class("shadow-2xl"),
		html.Span(
			html.Class("text-center text-sm leading-none font-medium tracking-tight whitespace-pre-wrap text-white lg:text-lg dark:from-white dark:to-slate-900/10"),
			"Shimmer Button",
		),
	)
}
