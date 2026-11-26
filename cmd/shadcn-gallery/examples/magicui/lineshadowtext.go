package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn/magicui"
)

// LineShadowTextDemo is a headline with a moving hatched shadow on its
// second word. Upstream picks a black or white shadow from the theme in
// JavaScript; the CSS variable --foreground does that without it.
func LineShadowTextDemo() mx.Component {
	return html.H1(html.Class("text-5xl leading-none font-semibold tracking-tighter text-balance sm:text-6xl md:text-7xl lg:text-8xl"),
		"Ship ",
		magicui.LineShadowText(magicui.LineShadowTextOptions{ShadowColor: "var(--foreground)"}, html.Class("italic"),
			"Fast",
		),
	)
}
