package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn/magicui"
)

// DotPatternDemo is a dot grid faded out towards the edges.
func DotPatternDemo() mx.Component {
	return html.DivClass("bg-background relative flex h-[500px] w-full flex-col items-center justify-center overflow-hidden rounded-lg border",
		magicui.DotPattern(magicui.DotPatternOptions{ID: "dot-pattern-demo"},
			html.Class("[mask-image:radial-gradient(300px_circle_at_center,white,transparent)]")),
	)
}

// DotPatternLinearGradient is a wider dot grid faded out diagonally.
func DotPatternLinearGradient() mx.Component {
	return html.DivClass("bg-background relative flex size-full items-center justify-center overflow-hidden rounded-lg border p-20",
		magicui.DotPattern(magicui.DotPatternOptions{ID: "dot-pattern-linear-gradient", Width: 20, Height: 20, CX: 1, CY: 1, CR: 1},
			html.Class("[mask-image:linear-gradient(to_bottom_right,white,transparent,transparent)]")),
	)
}

// DotPatternWithGlowEffect is a dot grid whose dots pulse at their own
// pace.
func DotPatternWithGlowEffect() mx.Component {
	return html.DivClass("relative flex h-[500px] w-full flex-col items-center justify-center overflow-hidden",
		magicui.DotPattern(magicui.DotPatternOptions{ID: "dot-pattern-glow", Glow: true, Seed: 1},
			html.Class("[mask-image:radial-gradient(300px_circle_at_center,white,transparent)]")),
	)
}
