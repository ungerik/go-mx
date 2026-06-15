package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/shadcn"
)

// TooltipDemo renders an outline button that reveals a tooltip on hover.
func TooltipDemo() mx.Component {
	return shadcn.Tooltip(
		shadcn.TooltipTrigger("demo-tooltip",
			shadcn.Button(shadcn.ButtonOutline, shadcn.SizeDefault, "Hover")),
		shadcn.TooltipContent("demo-tooltip", "", "Add to library"),
	)
}
