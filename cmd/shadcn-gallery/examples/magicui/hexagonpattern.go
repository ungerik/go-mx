package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn/magicui"
)

// HexagonPatternDemo is a skewed honeycomb with some hexagons filled,
// faded out towards the edges.
func HexagonPatternDemo() mx.Component {
	return html.DivClass("bg-background relative flex h-[500px] w-full flex-col items-center justify-center overflow-hidden rounded-lg border",
		magicui.HexagonPattern(magicui.HexagonPatternOptions{
			ID:       "hexagon-pattern-demo",
			Hexagons: [][2]int{{1, 1}, {4, 4}, {2, 2}, {3, 4}, {5, 4}, {8, 2}, {6, 3}, {8, 5}, {10, 10}},
		}, html.Class(
			"mask-[radial-gradient(420px_circle_at_center,white,transparent)]",
			"inset-0 skew-y-6",
		)),
	)
}

// HexagonPatternDashed is a honeycomb of dashed outlines.
func HexagonPatternDashed() mx.Component {
	return html.DivClass("bg-background relative flex h-[500px] w-full items-center justify-center overflow-hidden rounded-lg border p-20",
		magicui.HexagonPattern(magicui.HexagonPatternOptions{ID: "hexagon-pattern-dashed", Radius: 40, StrokeDasharray: "4 2"}),
	)
}

// HexagonPatternSpacing is a honeycomb with gaps between the hexagons.
func HexagonPatternSpacing() mx.Component {
	return html.DivClass("bg-background relative flex h-[500px] w-full items-center justify-center overflow-hidden rounded-lg border p-20",
		magicui.HexagonPattern(magicui.HexagonPatternOptions{ID: "hexagon-pattern-spacing", Gap: 20, Radius: 40}),
	)
}
