package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// DrawerDemo renders an outline trigger that opens a drawer with a header, body, and footer actions.
func DrawerDemo() mx.Component {
	return shadcn.Drawer(
		shadcn.DrawerTrigger("demo-drawer",
			html.Class(shadcn.ButtonClasses(shadcn.ButtonOutline, shadcn.SizeDefault)),
			"Open Drawer"),
		shadcn.DrawerContent("demo-drawer",
			html.DivClass("mx-auto w-full max-w-sm",
				shadcn.DrawerHeader(
					shadcn.DrawerTitle("Move Goal"),
					shadcn.DrawerDescription("Set your daily activity goal. Drag the handle down to dismiss."),
				),
				html.DivClass("p-4 pb-0",
					html.DivClass("flex items-center justify-center text-6xl font-bold tracking-tighter", "350"),
					html.DivClass("text-muted-foreground mt-1 text-center text-xs uppercase", "Calories/day"),
				),
				shadcn.DrawerFooter(
					shadcn.Button(shadcn.ButtonDefault, shadcn.SizeDefault, "Submit"),
					shadcn.DrawerClose(html.Class(shadcn.ButtonClasses(shadcn.ButtonOutline, shadcn.SizeDefault)), "Cancel"),
				),
			),
		),
	)
}
