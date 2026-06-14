package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

func DialogDemo() mx.Component {
	return shadcn.Dialog(
		shadcn.DialogTrigger("demo-dialog",
			html.Class(shadcn.ButtonClasses(shadcn.ButtonOutline, shadcn.SizeDefault)),
			"Edit profile"),
		shadcn.DialogContent("demo-dialog",
			shadcn.DialogHeader(
				shadcn.DialogTitle("Edit profile"),
				shadcn.DialogDescription("Make changes to your profile here. Click save when you're done."),
			),
			html.Div(html.Class("grid gap-4"),
				html.Div(html.Class("grid gap-2"),
					shadcn.Label(html.For("dialog-name"), "Name"),
					shadcn.Input(html.ID("dialog-name"), html.Value("Ada Lovelace")),
				),
				html.Div(html.Class("grid gap-2"),
					shadcn.Label(html.For("dialog-username"), "Username"),
					shadcn.Input(html.ID("dialog-username"), html.Value("@ada")),
				),
			),
			shadcn.DialogFooter(
				shadcn.DialogClose(html.Class(shadcn.ButtonClasses(shadcn.ButtonOutline, shadcn.SizeDefault)), "Cancel"),
				shadcn.Button(shadcn.ButtonDefault, shadcn.SizeDefault, "Save changes"),
			),
		),
	)
}
