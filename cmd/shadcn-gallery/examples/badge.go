package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

func BadgeDefault() mx.Component {
	return shadcn.Badge(shadcn.BadgeDefault, "Badge")
}

func BadgeSecondary() mx.Component {
	return shadcn.Badge(shadcn.BadgeSecondary, "Secondary")
}

func BadgeDestructive() mx.Component {
	return shadcn.Badge(shadcn.BadgeDestructive, "Destructive")
}

func BadgeOutline() mx.Component {
	return shadcn.Badge(shadcn.BadgeOutline, "Outline")
}

func BadgeRow() mx.Component {
	return html.DivClass("flex flex-wrap items-center gap-2",
		shadcn.Badge(shadcn.BadgeDefault, "Default"),
		shadcn.Badge(shadcn.BadgeSecondary, "Secondary"),
		shadcn.Badge(shadcn.BadgeDestructive, "Destructive"),
		shadcn.Badge(shadcn.BadgeOutline, "Outline"),
	)
}
