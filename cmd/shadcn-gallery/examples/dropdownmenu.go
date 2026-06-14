package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

func DropdownMenuDemo() mx.Component {
	return shadcn.DropdownMenu(
		shadcn.DropdownMenuTrigger("demo-dropdown",
			html.Class(shadcn.ButtonClasses(shadcn.ButtonOutline, shadcn.SizeDefault)),
			"Open"),
		shadcn.DropdownMenuContent("demo-dropdown", "",
			shadcn.DropdownMenuLabel("My Account"),
			shadcn.DropdownMenuSeparator(),
			shadcn.DropdownMenuGroup(
				shadcn.DropdownMenuItem("Profile", shadcn.DropdownMenuShortcut("⇧⌘P")),
				shadcn.DropdownMenuItem("Billing", shadcn.DropdownMenuShortcut("⌘B")),
				shadcn.DropdownMenuItem("Settings", shadcn.DropdownMenuShortcut("⌘S")),
			),
			shadcn.DropdownMenuSeparator(),
			shadcn.DropdownMenuItem("Log out", shadcn.DropdownMenuShortcut("⇧⌘Q")),
		),
	)
}
