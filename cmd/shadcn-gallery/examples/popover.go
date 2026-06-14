package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

func PopoverDemo() mx.Component {
	return shadcn.Popover(
		shadcn.PopoverTrigger("demo-popover",
			html.Class(shadcn.ButtonClasses(shadcn.ButtonOutline, shadcn.SizeDefault)),
			"Open popover"),
		shadcn.PopoverContent("demo-popover", "",
			html.Div(html.Class("grid gap-2"),
				html.Element("h4", html.Class("leading-none font-medium"), "Dimensions"),
				html.P(html.Class("text-sm text-muted-foreground"), "Set the dimensions for the layer."),
			),
		),
	)
}
