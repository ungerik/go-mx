package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

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
			html.Div(html.Class("grid flex-1 auto-rows-min gap-4 px-4"),
				html.Div(html.Class("grid gap-2"),
					shadcn.Label(html.For("sheet-name"), "Name"),
					shadcn.Input(html.ID("sheet-name"), html.Value("Ada Lovelace")),
				),
				html.Div(html.Class("grid gap-2"),
					shadcn.Label(html.For("sheet-username"), "Username"),
					shadcn.Input(html.ID("sheet-username"), html.Value("@ada")),
				),
			),
			shadcn.SheetFooter(
				shadcn.Button(shadcn.ButtonDefault, shadcn.SizeDefault, "Save changes"),
				shadcn.SheetClose(html.Class(shadcn.ButtonClasses(shadcn.ButtonOutline, shadcn.SizeDefault)), "Close"),
			),
		),
	)
}
