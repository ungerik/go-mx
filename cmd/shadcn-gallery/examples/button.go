// Package examples holds one function per labeled component preview in the
// gallery. Each returns the live component; its source is also shown in the
// page's "Code" tab, extracted from these files at startup. Keep every body a
// single self-contained expression so the displayed snippet reads cleanly.
package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// ButtonDefault renders a default-variant button.
func ButtonDefault() mx.Component {
	return shadcn.Button(shadcn.ButtonDefault, shadcn.SizeDefault, "Button")
}

// ButtonSecondary renders a secondary-variant button.
func ButtonSecondary() mx.Component {
	return shadcn.Button(shadcn.ButtonSecondary, shadcn.SizeDefault, "Secondary")
}

// ButtonDestructive renders a destructive-variant button.
func ButtonDestructive() mx.Component {
	return shadcn.Button(shadcn.ButtonDestructive, shadcn.SizeDefault, "Destructive")
}

// ButtonOutline renders an outline-variant button.
func ButtonOutline() mx.Component {
	return shadcn.Button(shadcn.ButtonOutline, shadcn.SizeDefault, "Outline")
}

// ButtonGhost renders a ghost-variant button.
func ButtonGhost() mx.Component {
	return shadcn.Button(shadcn.ButtonGhost, shadcn.SizeDefault, "Ghost")
}

// ButtonLink renders a link-variant button.
func ButtonLink() mx.Component {
	return shadcn.Button(shadcn.ButtonLink, shadcn.SizeDefault, "Link")
}

// ButtonSizes renders outline buttons in small, default, and large sizes.
func ButtonSizes() mx.Component {
	return html.DivClass("flex items-center gap-2",
		shadcn.Button(shadcn.ButtonOutline, shadcn.SizeSM, "Small"),
		shadcn.Button(shadcn.ButtonOutline, shadcn.SizeDefault, "Default"),
		shadcn.Button(shadcn.ButtonOutline, shadcn.SizeLG, "Large"),
	)
}

// ButtonDisabled renders a disabled button.
func ButtonDisabled() mx.Component {
	return shadcn.Button(shadcn.ButtonDefault, shadcn.SizeDefault, html.Disabled, "Disabled")
}
