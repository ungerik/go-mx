package magicui

import (
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn/magicui"
)

// Text3DFlipDemo rolls a quote's letters over towards the top on hover,
// starting at the first letter.
func Text3DFlipDemo() mx.Component {
	return magicui.Text3DFlip(magicui.Text3DFlipOptions{
		TextClass:       "bg-background text-foreground",
		FlipTextClass:   "bg-background text-foreground",
		RotateDirection: magicui.Text3DFlipTop,
		StaggerDuration: 30 * time.Millisecond,
		StaggerFrom:     magicui.Text3DFlipFromFirst,
		// upstream's spring with damping 25 and stiffness 160 settles in about 650ms
		Duration: 650 * time.Millisecond,
	},
		html.Class("bg-background font-serif text-2xl sm:text-5xl md:text-[56px]"),
		"Stay hungry, stay foolish",
	)
}

// Text3DFlipFromCenter rolls the letters over starting at the middle one.
func Text3DFlipFromCenter() mx.Component {
	return magicui.Text3DFlip(magicui.Text3DFlipOptions{
		TextClass:       "bg-background text-foreground",
		FlipTextClass:   "bg-background text-foreground",
		RotateDirection: magicui.Text3DFlipTop,
		StaggerDuration: 30 * time.Millisecond,
		StaggerFrom:     magicui.Text3DFlipFromCenter,
	},
		html.Class("bg-background font-serif text-2xl sm:text-5xl md:text-[56px]"),
		"Design for failure",
	)
}
