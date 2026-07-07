package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// ButtonGroupDemo renders a horizontal group of joined outline buttons.
func ButtonGroupDemo() mx.Component {
	return shadcn.ButtonGroup("",
		shadcn.Button(shadcn.ButtonOutline, "", "Archive"),
		shadcn.Button(shadcn.ButtonOutline, "", "Report"),
		shadcn.Button(shadcn.ButtonOutline, "", "Snooze"),
	)
}

// ButtonGroupVerticalDemo renders a vertical button group.
func ButtonGroupVerticalDemo() mx.Component {
	return shadcn.ButtonGroup(shadcn.ButtonGroupVertical,
		shadcn.Button(shadcn.ButtonOutline, shadcn.SizeIcon, navIcon("m18 15-6-6-6 6")),
		shadcn.Button(shadcn.ButtonOutline, shadcn.SizeIcon, navIcon("m6 9 6 6 6-6")),
	)
}

// ButtonGroupWithText renders a group mixing static text, an input and a button.
func ButtonGroupWithText() mx.Component {
	return shadcn.ButtonGroup("",
		shadcn.ButtonGroupText("https://"),
		shadcn.Input(html.Placeholder("example.com")),
		shadcn.Button(shadcn.ButtonOutline, "", "Go"),
	)
}

// ButtonGroupWithSeparator renders secondary buttons split by an explicit
// separator (secondary buttons have no border, so the automatic join line
// between them needs one).
func ButtonGroupWithSeparator() mx.Component {
	return shadcn.ButtonGroup("",
		shadcn.Button(shadcn.ButtonSecondary, "", "Copy"),
		shadcn.ButtonGroupSeparator(""),
		shadcn.Button(shadcn.ButtonSecondary, "", "Paste"),
	)
}
