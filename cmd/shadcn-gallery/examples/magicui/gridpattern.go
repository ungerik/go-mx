package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn/magicui"
)

// GridPatternDemo is a skewed grid with some cells filled, faded out
// towards the edges.
func GridPatternDemo() mx.Component {
	return html.DivClass("bg-background relative flex h-[500px] w-full flex-col items-center justify-center overflow-hidden rounded-lg border",
		magicui.GridPattern(magicui.GridPatternOptions{
			ID: "grid-pattern-demo",
			Squares: [][2]int{
				{4, 4}, {5, 1}, {8, 2}, {5, 3}, {5, 5}, {10, 10},
				{12, 15}, {15, 10}, {10, 15}, {15, 10}, {10, 15}, {15, 10},
			},
		}, html.Class(
			"[mask-image:radial-gradient(400px_circle_at_center,white,transparent)]",
			"inset-x-0 inset-y-[-30%] h-[200%] skew-y-12",
		)),
	)
}

// GridPatternLinearGradient is a fine grid faded out diagonally.
func GridPatternLinearGradient() mx.Component {
	return html.DivClass("bg-background relative flex size-full items-center justify-center overflow-hidden rounded-lg border p-20",
		magicui.GridPattern(magicui.GridPatternOptions{ID: "grid-pattern-linear-gradient", Width: 20, Height: 20},
			html.Class("[mask-image:linear-gradient(to_bottom_right,white,transparent,transparent)]")),
	)
}

// GridPatternDashed is a grid of dashed lines faded out towards the edges.
func GridPatternDashed() mx.Component {
	return html.DivClass("bg-background relative flex size-full items-center justify-center overflow-hidden rounded-lg border p-20",
		magicui.GridPattern(magicui.GridPatternOptions{ID: "grid-pattern-dashed", Width: 30, Height: 30, StrokeDasharray: "4 2"},
			html.Class("[mask-image:radial-gradient(300px_circle_at_center,white,transparent)]")),
	)
}
