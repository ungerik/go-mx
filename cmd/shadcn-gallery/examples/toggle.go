package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

func ToggleDemo() mx.Component {
	return shadcn.Toggle(shadcn.ToggleDefault, shadcn.ToggleSizeDefault,
		html.Attrib("aria-label", "Toggle bold"), "Bold")
}

func ToggleOutline() mx.Component {
	return shadcn.Toggle(shadcn.ToggleOutline, shadcn.ToggleSizeDefault,
		html.Attrib("aria-label", "Toggle italic"), "Italic")
}

func ToggleSizes() mx.Component {
	return html.Div(html.Class("flex items-center gap-2"),
		shadcn.Toggle(shadcn.ToggleOutline, shadcn.ToggleSizeSM, "Small"),
		shadcn.Toggle(shadcn.ToggleOutline, shadcn.ToggleSizeDefault, "Default"),
		shadcn.Toggle(shadcn.ToggleOutline, shadcn.ToggleSizeLG, "Large"),
	)
}

func ToggleDisabled() mx.Component {
	return shadcn.Toggle(shadcn.ToggleDefault, shadcn.ToggleSizeDefault,
		html.Disabled, "Disabled")
}
