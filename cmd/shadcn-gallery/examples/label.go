package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// LabelDemo renders a label associated with a checkbox.
func LabelDemo() mx.Component {
	return html.DivClass("flex items-center gap-2",
		shadcn.CheckboxID("terms"),
		shadcn.LabelFor("terms", "Accept terms and conditions"),
	)
}
