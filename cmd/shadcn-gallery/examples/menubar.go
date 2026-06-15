package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/shadcn"
)

// MenubarDemo renders a menubar with File, Edit, and View menus containing items, shortcuts, and checkbox items.
func MenubarDemo() mx.Component {
	return shadcn.Menubar(
		shadcn.MenubarMenu(
			shadcn.MenubarTrigger("mb-file", "File"),
			shadcn.MenubarContent("mb-file", "",
				shadcn.MenubarItem("New Tab", shadcn.MenubarShortcut("⌘T")),
				shadcn.MenubarItem("New Window", shadcn.MenubarShortcut("⌘N")),
				shadcn.MenubarSeparator(),
				shadcn.MenubarItem("Print…", shadcn.MenubarShortcut("⌘P")),
			),
		),
		shadcn.MenubarMenu(
			shadcn.MenubarTrigger("mb-edit", "Edit"),
			shadcn.MenubarContent("mb-edit", "",
				shadcn.MenubarItem("Undo", shadcn.MenubarShortcut("⌘Z")),
				shadcn.MenubarItem("Redo", shadcn.MenubarShortcut("⇧⌘Z")),
			),
		),
		shadcn.MenubarMenu(
			shadcn.MenubarTrigger("mb-view", "View"),
			shadcn.MenubarContent("mb-view", "",
				shadcn.MenubarCheckboxItem(true, "Always Show Bookmarks Bar"),
				shadcn.MenubarCheckboxItem(false, "Always Show Full URLs"),
			),
		),
	)
}
