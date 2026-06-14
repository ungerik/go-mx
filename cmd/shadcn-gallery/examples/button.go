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

func ButtonDefault() mx.Component {
	return shadcn.Button(shadcn.ButtonDefault, shadcn.SizeDefault, "Button")
}

func ButtonSecondary() mx.Component {
	return shadcn.Button(shadcn.ButtonSecondary, shadcn.SizeDefault, "Secondary")
}

func ButtonDestructive() mx.Component {
	return shadcn.Button(shadcn.ButtonDestructive, shadcn.SizeDefault, "Destructive")
}

func ButtonOutline() mx.Component {
	return shadcn.Button(shadcn.ButtonOutline, shadcn.SizeDefault, "Outline")
}

func ButtonGhost() mx.Component {
	return shadcn.Button(shadcn.ButtonGhost, shadcn.SizeDefault, "Ghost")
}

func ButtonLink() mx.Component {
	return shadcn.Button(shadcn.ButtonLink, shadcn.SizeDefault, "Link")
}

func ButtonSizes() mx.Component {
	return html.DivClass("flex items-center gap-2",
		shadcn.Button(shadcn.ButtonOutline, shadcn.SizeSM, "Small"),
		shadcn.Button(shadcn.ButtonOutline, shadcn.SizeDefault, "Default"),
		shadcn.Button(shadcn.ButtonOutline, shadcn.SizeLG, "Large"),
	)
}

func ButtonDisabled() mx.Component {
	return shadcn.Button(shadcn.ButtonDefault, shadcn.SizeDefault, html.Disabled, "Disabled")
}
