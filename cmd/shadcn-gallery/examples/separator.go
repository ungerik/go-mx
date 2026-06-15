package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// SeparatorDemo renders horizontal and vertical separators between sections.
func SeparatorDemo() mx.Component {
	return html.Div(
		html.DivClass("space-y-1",
			html.Element("h4", html.Class("text-sm leading-none font-medium"), "go-mx"),
			html.PClass("text-muted-foreground text-sm", "An HTML component library for Go."),
		),
		shadcn.Separator(shadcn.SeparatorHorizontal, html.Class("my-4")),
		html.DivClass("flex h-5 items-center space-x-4 text-sm",
			html.Div("Blog"),
			shadcn.Separator(shadcn.SeparatorVertical),
			html.Div("Docs"),
			shadcn.Separator(shadcn.SeparatorVertical),
			html.Div("Source"),
		),
	)
}
