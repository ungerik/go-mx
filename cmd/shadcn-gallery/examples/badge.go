package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// BadgeDefault renders a default-variant badge.
func BadgeDefault() mx.Component {
	return shadcn.Badge(shadcn.BadgeDefault, "Badge")
}

// BadgeSecondary renders a secondary-variant badge.
func BadgeSecondary() mx.Component {
	return shadcn.Badge(shadcn.BadgeSecondary, "Secondary")
}

// BadgeDestructive renders a destructive-variant badge.
func BadgeDestructive() mx.Component {
	return shadcn.Badge(shadcn.BadgeDestructive, "Destructive")
}

// BadgeOutline renders an outline-variant badge.
func BadgeOutline() mx.Component {
	return shadcn.Badge(shadcn.BadgeOutline, "Outline")
}

// BadgeRow renders a row of badges in all four variants.
func BadgeRow() mx.Component {
	return html.DivClass("flex flex-wrap items-center gap-2",
		shadcn.Badge(shadcn.BadgeDefault, "Default"),
		shadcn.Badge(shadcn.BadgeSecondary, "Secondary"),
		shadcn.Badge(shadcn.BadgeDestructive, "Destructive"),
		shadcn.Badge(shadcn.BadgeOutline, "Outline"),
	)
}
