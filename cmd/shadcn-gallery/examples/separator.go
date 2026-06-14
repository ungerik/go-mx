package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

func SeparatorDemo() mx.Component {
	return html.Div(
		html.Div(html.Class("space-y-1"),
			html.Element("h4", html.Class("text-sm leading-none font-medium"), "go-mx"),
			html.P(html.Class("text-muted-foreground text-sm"), "An HTML component library for Go."),
		),
		shadcn.Separator(shadcn.SeparatorHorizontal, html.Class("my-4")),
		html.Div(html.Class("flex h-5 items-center space-x-4 text-sm"),
			html.Div("Blog"),
			shadcn.Separator(shadcn.SeparatorVertical),
			html.Div("Docs"),
			shadcn.Separator(shadcn.SeparatorVertical),
			html.Div("Source"),
		),
	)
}
