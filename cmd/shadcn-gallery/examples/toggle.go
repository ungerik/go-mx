package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// ToggleDemo renders a default toggle button.
func ToggleDemo() mx.Component {
	return shadcn.Toggle(shadcn.ToggleDefault, shadcn.ToggleSizeDefault,
		html.Attrib("aria-label", "Toggle bold"), "Bold")
}

// ToggleOutline renders an outline-variant toggle button.
func ToggleOutline() mx.Component {
	return shadcn.Toggle(shadcn.ToggleOutline, shadcn.ToggleSizeDefault,
		html.Attrib("aria-label", "Toggle italic"), "Italic")
}

// ToggleSizes renders outline toggle buttons in small, default, and large sizes.
func ToggleSizes() mx.Component {
	return html.DivClass("flex items-center gap-2",
		shadcn.Toggle(shadcn.ToggleOutline, shadcn.ToggleSizeSM, "Small"),
		shadcn.Toggle(shadcn.ToggleOutline, shadcn.ToggleSizeDefault, "Default"),
		shadcn.Toggle(shadcn.ToggleOutline, shadcn.ToggleSizeLG, "Large"),
	)
}

// ToggleDisabled renders a disabled toggle button.
func ToggleDisabled() mx.Component {
	return shadcn.Toggle(shadcn.ToggleDefault, shadcn.ToggleSizeDefault,
		html.Disabled, "Disabled")
}
