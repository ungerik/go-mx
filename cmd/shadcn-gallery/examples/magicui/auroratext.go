package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn/magicui"
)

// AuroraTextDemo is a headline with one word in aurora colors.
func AuroraTextDemo() mx.Component {
	return html.H1(html.Class("text-4xl font-bold tracking-tighter md:text-5xl lg:text-7xl"),
		"Ship ",
		magicui.AuroraText(magicui.AuroraTextOptions{}, "beautiful"),
	)
}
