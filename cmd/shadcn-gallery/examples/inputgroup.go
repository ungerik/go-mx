package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// InputGroupDemo renders an input with text addons on both sides.
func InputGroupDemo() mx.Component {
	return shadcn.InputGroup(html.Class("max-w-sm"),
		shadcn.InputGroupAddon("", shadcn.InputGroupText("@")),
		shadcn.InputGroupInput(html.Placeholder("Username")),
		shadcn.InputGroupAddon(shadcn.InputGroupAddonInlineEnd,
			shadcn.InputGroupText(".example.com")),
	)
}

// InputGroupWithButton renders an input with a button addon.
func InputGroupWithButton() mx.Component {
	return shadcn.InputGroup(html.Class("max-w-sm"),
		shadcn.InputGroupInput(html.Placeholder("https://x.com/@shadcn")),
		shadcn.InputGroupAddon(shadcn.InputGroupAddonInlineEnd,
			shadcn.InputGroupButton("", "", "Copy"),
		),
	)
}

// InputGroupWithKbd renders a search input with a keyboard-shortcut hint.
func InputGroupWithKbd() mx.Component {
	return shadcn.InputGroup(html.Class("max-w-sm"),
		shadcn.InputGroupInput(html.Placeholder("Search…")),
		shadcn.InputGroupAddon(shadcn.InputGroupAddonInlineEnd, shadcn.Kbd("⌘K")),
	)
}

// InputGroupWithTextarea renders a textarea with a full-width footer addon row.
func InputGroupWithTextarea() mx.Component {
	return shadcn.InputGroup(html.Class("max-w-sm"),
		shadcn.InputGroupTextarea(html.Placeholder("Ask, search or chat…"), html.Rows(4)),
		shadcn.InputGroupAddon(shadcn.InputGroupAddonBlockEnd,
			shadcn.InputGroupText("52 characters left"),
			shadcn.InputGroupButton("", shadcn.InputGroupButtonSM,
				html.Class("ml-auto"), "Send"),
		),
	)
}
