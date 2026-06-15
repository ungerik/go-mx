package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// SheetDemo renders an outline trigger that opens a side sheet with a profile-editing form and footer actions.
func SheetDemo() mx.Component {
	return shadcn.Sheet(
		shadcn.SheetTrigger("demo-sheet",
			html.Class(shadcn.ButtonClasses(shadcn.ButtonOutline, shadcn.SizeDefault)),
			"Open sheet"),
		shadcn.SheetContent("demo-sheet", "",
			shadcn.SheetHeader(
				shadcn.SheetTitle("Edit profile"),
				shadcn.SheetDescription("Make changes to your profile here. Click save when you're done."),
			),
			html.DivClass("grid flex-1 auto-rows-min gap-4 px-4",
				html.DivClass("grid gap-2",
					shadcn.LabelFor("sheet-name", "Name"),
					shadcn.InputID("sheet-name", html.Value("Ada Lovelace")),
				),
				html.DivClass("grid gap-2",
					shadcn.LabelFor("sheet-username", "Username"),
					shadcn.InputID("sheet-username", html.Value("@ada")),
				),
			),
			shadcn.SheetFooter(
				shadcn.Button(shadcn.ButtonDefault, shadcn.SizeDefault, "Save changes"),
				shadcn.SheetClose(html.Class(shadcn.ButtonClasses(shadcn.ButtonOutline, shadcn.SizeDefault)), "Close"),
			),
		),
	)
}
