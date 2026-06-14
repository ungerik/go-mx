package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

func ContextMenuDemo() mx.Component {
	return shadcn.ContextMenu(
		shadcn.ContextMenuTrigger("demo-contextmenu",
			html.Class("flex h-[150px] w-[300px] items-center justify-center rounded-md border border-dashed text-sm"),
			"Right-click here"),
		shadcn.ContextMenuContent("demo-contextmenu",
			shadcn.ContextMenuItem("Back", shadcn.ContextMenuShortcut("⌘[")),
			shadcn.ContextMenuItem("Forward", shadcn.ContextMenuShortcut("⌘]")),
			shadcn.ContextMenuItem("Reload", shadcn.ContextMenuShortcut("⌘R")),
			shadcn.ContextMenuSeparator(),
			shadcn.ContextMenuCheckboxItem(true, "Show Bookmarks"),
			shadcn.ContextMenuCheckboxItem(false, "Show Full URLs"),
		),
	)
}
