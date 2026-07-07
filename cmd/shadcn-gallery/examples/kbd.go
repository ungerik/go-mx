package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// KbdDemo renders a few keyboard keys.
func KbdDemo() mx.Component {
	return html.DivClass("flex items-center gap-2",
		shadcn.Kbd("⌘"),
		shadcn.Kbd("⇧"),
		shadcn.Kbd("⌥"),
		shadcn.Kbd("Ctrl"),
	)
}

// KbdGroupDemo renders keyboard shortcuts as grouped keys.
func KbdGroupDemo() mx.Component {
	return html.DivClass("flex flex-col items-center gap-4",
		html.DivClass("text-muted-foreground text-sm",
			"Use ",
			shadcn.KbdGroup(shadcn.Kbd("Ctrl"), " + ", shadcn.Kbd("B")),
			" to toggle the sidebar.",
		),
		shadcn.KbdGroup(shadcn.Kbd("⌘"), shadcn.Kbd("K")),
	)
}
