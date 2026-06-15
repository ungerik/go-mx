package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// DialogDemo renders an outline trigger that opens a dialog with a profile-editing form and footer actions.
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
			html.DivClass("grid gap-4",
				html.DivClass("grid gap-2",
					shadcn.LabelFor("dialog-name", "Name"),
					shadcn.InputID("dialog-name", html.Value("Ada Lovelace")),
				),
				html.DivClass("grid gap-2",
					shadcn.LabelFor("dialog-username", "Username"),
					shadcn.InputID("dialog-username", html.Value("@ada")),
				),
			),
			shadcn.DialogFooter(
				shadcn.DialogClose(html.Class(shadcn.ButtonClasses(shadcn.ButtonOutline, shadcn.SizeDefault)), "Cancel"),
				shadcn.Button(shadcn.ButtonDefault, shadcn.SizeDefault, "Save changes"),
			),
		),
	)
}
