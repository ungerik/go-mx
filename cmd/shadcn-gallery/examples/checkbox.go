package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

func CheckboxDemo() mx.Component {
	return html.Div(html.Class("flex items-start gap-2"),
		shadcn.Checkbox(html.ID("terms-cb")),
		html.Div(html.Class("grid gap-1.5"),
			shadcn.Label(html.For("terms-cb"), "Accept terms and conditions"),
			html.P(html.Class("text-sm text-muted-foreground"),
				"You agree to our Terms of Service and Privacy Policy."),
		),
	)
}

func CheckboxChecked() mx.Component {
	return html.Div(html.Class("flex items-center gap-2"),
		shadcn.Checkbox(html.ID("checked-cb"), html.Attrib("checked", "")),
		shadcn.Label(html.For("checked-cb"), "Subscribe to the newsletter"),
	)
}

func CheckboxDisabled() mx.Component {
	return html.Div(html.Class("flex items-center gap-2"),
		shadcn.Checkbox(html.ID("disabled-cb"), html.Attrib("disabled", "")),
		shadcn.Label(html.For("disabled-cb"), "Disabled"),
	)
}
