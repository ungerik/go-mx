package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn/magicui"
)

// StripedPatternDemo is a field of diagonal stripes faded out towards the
// edges.
func StripedPatternDemo() mx.Component {
	return html.DivClass("relative flex h-[500px] w-full flex-col items-center justify-center overflow-hidden rounded-lg border",
		magicui.StripedPattern(magicui.StripedPatternOptions{ID: "striped-pattern-demo"},
			html.Class("[mask-image:radial-gradient(300px_circle_at_center,white,transparent)]")),
	)
}

// StripedPatternDashed is a field of thin, dashed stripes.
func StripedPatternDashed() mx.Component {
	return html.DivClass("relative flex h-[300px] w-full flex-col items-center justify-center overflow-hidden rounded-lg border",
		magicui.StripedPattern(magicui.StripedPatternOptions{ID: "striped-pattern-dashed"},
			html.Class("stroke-[0.3] [stroke-dasharray:8,4]")),
	)
}

// StripedPatternRight is a field of stripes leaning right.
func StripedPatternRight() mx.Component {
	return html.DivClass("relative flex h-[300px] w-full flex-col items-center justify-center overflow-hidden rounded-lg border",
		magicui.StripedPattern(magicui.StripedPatternOptions{ID: "striped-pattern-right", Direction: magicui.StripedPatternRight}),
	)
}
