package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn/magicui"
)

// RippleDemo is a word with ripples pulsing out behind it.
func RippleDemo() mx.Component {
	return html.DivClass("bg-background relative flex h-[500px] w-full flex-col items-center justify-center overflow-hidden rounded-lg border",
		html.PClass("z-10 text-center text-5xl font-medium tracking-tighter whitespace-pre-wrap text-white", "Ripple"),
		magicui.Ripple(magicui.RippleOptions{}),
	)
}
