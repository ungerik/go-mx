package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// CheckboxDemo renders a checkbox with a label and supporting description text.
func CheckboxDemo() mx.Component {
	return html.DivClass("flex items-start gap-2",
		shadcn.CheckboxID("terms-cb"),
		html.DivClass("grid gap-1.5",
			shadcn.LabelFor("terms-cb", "Accept terms and conditions"),
			html.PClass("text-sm text-muted-foreground",
				"You agree to our Terms of Service and Privacy Policy."),
		),
	)
}

// CheckboxChecked renders a checked, labeled checkbox.
func CheckboxChecked() mx.Component {
	return html.DivClass("flex items-center gap-2",
		shadcn.CheckboxID("checked-cb", html.Checked),
		shadcn.LabelFor("checked-cb", "Subscribe to the newsletter"),
	)
}

// CheckboxDisabled renders a disabled, labeled checkbox.
func CheckboxDisabled() mx.Component {
	return html.DivClass("flex items-center gap-2",
		shadcn.CheckboxID("disabled-cb", html.Disabled),
		shadcn.LabelFor("disabled-cb", "Disabled"),
	)
}
