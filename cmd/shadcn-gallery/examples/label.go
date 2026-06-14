package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

func LabelDemo() mx.Component {
	return html.Div(html.Class("flex items-center gap-2"),
		shadcn.Checkbox(html.ID("terms")),
		shadcn.Label(html.For("terms"), "Accept terms and conditions"),
	)
}
