package magicui

import (
	"strconv"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
	"github.com/ungerik/go-mx/shadcn/magicui"
)

// ProgressiveBlurDemo is a scrolling list whose bottom edge blurs out.
//
// Upstream puts the blur into the Radix ScrollArea root, outside its
// scrolling viewport. shadcn.ScrollArea is the scroll container itself, so
// the blur is its sibling in a positioned wrapper here, to stay in place
// while the list scrolls.
func ProgressiveBlurDemo() mx.Component {
	items := make([]any, 20)
	for i := range items {
		items[i] = html.DivClass("bg-card flex h-20 w-full items-center justify-center rounded-xl border", strconv.Itoa(i))
	}
	return html.DivClass("relative w-full rounded-xl border",
		html.DivClass("relative h-[400px] overflow-hidden",
			shadcn.ScrollArea(html.Class("h-full"),
				html.DivClass("flex flex-col gap-2 p-4", items...),
			),
			magicui.ProgressiveBlur(magicui.ProgressiveBlurOptions{Position: magicui.ProgressiveBlurBottom, Height: "40%"}),
		),
	)
}
