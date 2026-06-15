package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// SwitchDemo renders a labeled toggle switch.
func SwitchDemo() mx.Component {
	return html.DivClass("flex items-center gap-2",
		shadcn.SwitchID("airplane-mode"),
		shadcn.LabelFor("airplane-mode", "Airplane Mode"),
	)
}

// SwitchDisabled renders a disabled labeled toggle switch.
func SwitchDisabled() mx.Component {
	return html.DivClass("flex items-center gap-2",
		shadcn.SwitchID("disabled-switch", html.Disabled),
		shadcn.LabelFor("disabled-switch", "Disabled"),
	)
}
