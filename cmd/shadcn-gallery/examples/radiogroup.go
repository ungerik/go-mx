package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// RadioGroupDemo renders a labeled radio group with the first option preselected.
func RadioGroupDemo() mx.Component {
	return shadcn.RadioGroup("plan",
		html.DivClass("flex items-center gap-2",
			shadcn.RadioGroupItem("plan", "default", html.ID("r1"), html.Checked),
			shadcn.LabelFor("r1", "Default"),
		),
		html.DivClass("flex items-center gap-2",
			shadcn.RadioGroupItem("plan", "comfortable", html.ID("r2")),
			shadcn.LabelFor("r2", "Comfortable"),
		),
		html.DivClass("flex items-center gap-2",
			shadcn.RadioGroupItem("plan", "compact", html.ID("r3")),
			shadcn.LabelFor("r3", "Compact"),
		),
	)
}
