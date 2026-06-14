package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

func SwitchDemo() mx.Component {
	return html.Div(html.Class("flex items-center gap-2"),
		shadcn.Switch(html.ID("airplane-mode")),
		shadcn.Label(html.For("airplane-mode"), "Airplane Mode"),
	)
}

func SwitchDisabled() mx.Component {
	return html.Div(html.Class("flex items-center gap-2"),
		shadcn.Switch(html.ID("disabled-switch"), html.Disabled),
		shadcn.Label(html.For("disabled-switch"), "Disabled"),
	)
}
