package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

func CardDemo() mx.Component {
	return shadcn.Card(html.Class("w-full max-w-sm"),
		shadcn.CardHeader(
			shadcn.CardTitle("Create project"),
			shadcn.CardDescription("Deploy your new project in one click."),
		),
		shadcn.CardContent(
			html.Div(html.Class("flex flex-col gap-2"),
				shadcn.Label(html.For("name"), "Name"),
				shadcn.Input(html.ID("name"), html.Placeholder("Name of your project")),
			),
		),
		shadcn.CardFooter(html.Class("flex justify-between"),
			shadcn.Button(shadcn.ButtonOutline, shadcn.SizeDefault, "Cancel"),
			shadcn.Button(shadcn.ButtonDefault, shadcn.SizeDefault, "Deploy"),
		),
	)
}
