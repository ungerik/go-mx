package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

func RadioGroupDemo() mx.Component {
	return shadcn.RadioGroup("plan",
		html.Div(html.Class("flex items-center gap-2"),
			shadcn.RadioGroupItem("plan", "default", html.ID("r1"), html.Checked),
			shadcn.Label(html.For("r1"), "Default"),
		),
		html.Div(html.Class("flex items-center gap-2"),
			shadcn.RadioGroupItem("plan", "comfortable", html.ID("r2")),
			shadcn.Label(html.For("r2"), "Comfortable"),
		),
		html.Div(html.Class("flex items-center gap-2"),
			shadcn.RadioGroupItem("plan", "compact", html.ID("r3")),
			shadcn.Label(html.For("r3"), "Compact"),
		),
	)
}
