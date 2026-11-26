package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn/magicui"
)

// MeteorsDemo is a meteor shower behind a gradient headline.
func MeteorsDemo() mx.Component {
	return html.DivClass("relative flex h-[500px] w-full flex-col items-center justify-center overflow-hidden rounded-lg border",
		magicui.Meteors(magicui.MeteorsOptions{Number: 30, Seed: 1}),
		html.Span(html.Class("pointer-events-none bg-linear-to-b from-black to-gray-300/80 bg-clip-text text-center text-8xl leading-none font-semibold whitespace-pre-wrap text-transparent dark:from-white dark:to-slate-900/10"),
			"Meteors"),
	)
}
