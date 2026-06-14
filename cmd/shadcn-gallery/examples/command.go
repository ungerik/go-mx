package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

func CommandDemo() mx.Component {
	return shadcn.Command(html.Class("max-w-md rounded-lg border shadow-md"),
		shadcn.CommandInput(html.Placeholder("Type a command or search...")),
		shadcn.CommandList(
			shadcn.CommandEmpty("No results found."),
			shadcn.CommandGroup("Suggestions",
				shadcn.CommandItem("Calendar"),
				shadcn.CommandItem("Search Emoji"),
				shadcn.CommandItem("Calculator"),
			),
			shadcn.CommandSeparator(),
			shadcn.CommandGroup("Settings",
				shadcn.CommandItem("Profile", shadcn.CommandShortcut("⌘P")),
				shadcn.CommandItem("Billing", shadcn.CommandShortcut("⌘B")),
				shadcn.CommandItem("Settings", shadcn.CommandShortcut("⌘S")),
			),
		),
	)
}
