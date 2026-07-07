package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// ItemDemo renders a basic outline item with icon, text and an action.
func ItemDemo() mx.Component {
	return shadcn.Item(shadcn.ItemOutline, "", html.Class("max-w-md"),
		shadcn.ItemMedia(shadcn.ItemMediaIcon,
			// lucide shield
			navIcon("M20 13c0 5-3.5 7.5-7.66 8.95a1 1 0 0 1-.67-.01C7.5 20.5 4 18 4 13V6a1 1 0 0 1 1-1c2 0 4.5-1.2 6.24-2.72a1 1 0 0 1 1.52 0C14.51 3.81 17 5 19 5a1 1 0 0 1 1 1z"),
		),
		shadcn.ItemContent(
			shadcn.ItemTitle("Basic Item"),
			shadcn.ItemDescription("A simple item with title and description."),
		),
		shadcn.ItemActions(
			shadcn.Button(shadcn.ButtonOutline, shadcn.SizeSM, "Action"),
		),
	)
}

// ItemGroupDemo renders a list of items separated by ItemSeparators.
func ItemGroupDemo() mx.Component {
	item := func(title, description string) mx.Component {
		return shadcn.Item("", "",
			shadcn.ItemContent(
				shadcn.ItemTitle(title),
				shadcn.ItemDescription(description),
			),
			shadcn.ItemActions(
				shadcn.Button(shadcn.ButtonGhost, shadcn.SizeIconSM, navIcon("M5 12h14", "m12 5 7 7-7 7")),
			),
		)
	}
	return shadcn.ItemGroup(html.Class("max-w-md"),
		item("Two-factor authentication", "Protect your account with an extra step."),
		shadcn.ItemSeparator(),
		item("Backup codes", "Generate one-time codes for account recovery."),
		shadcn.ItemSeparator(),
		item("Passkeys", "Sign in without a password."),
	)
}

// ItemMutedDemo renders muted-variant items in the small and extra-small sizes.
func ItemMutedDemo() mx.Component {
	return html.DivClass("flex w-full max-w-md flex-col gap-4",
		shadcn.Item(shadcn.ItemMuted, shadcn.ItemSizeSM,
			shadcn.ItemContent(shadcn.ItemTitle("Small muted item")),
		),
		shadcn.Item(shadcn.ItemMuted, shadcn.ItemSizeXS,
			shadcn.ItemContent(shadcn.ItemTitle("Extra-small muted item")),
		),
	)
}
