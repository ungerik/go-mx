package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

func ToggleGroupDemo() mx.Component {
	return shadcn.ToggleGroup(shadcn.ToggleGroupMultiple, shadcn.ToggleOutline, shadcn.ToggleSizeDefault, "text-style",
		shadcn.ToggleGroupItem("text-style", "bold", shadcn.ToggleOutline, shadcn.ToggleSizeDefault,
			html.Attrib("aria-label", "Bold"), "B"),
		shadcn.ToggleGroupItem("text-style", "italic", shadcn.ToggleOutline, shadcn.ToggleSizeDefault,
			html.Attrib("aria-label", "Italic"), "I"),
		shadcn.ToggleGroupItem("text-style", "underline", shadcn.ToggleOutline, shadcn.ToggleSizeDefault,
			html.Attrib("aria-label", "Underline"), "U"),
	)
}

func ToggleGroupSingleDemo() mx.Component {
	return shadcn.ToggleGroup(shadcn.ToggleGroupSingle, shadcn.ToggleDefault, shadcn.ToggleSizeDefault, "text-align",
		shadcn.ToggleGroupItem("text-align", "left", shadcn.ToggleDefault, shadcn.ToggleSizeDefault,
			html.Attrib("aria-label", "Left"), "Left"),
		shadcn.ToggleGroupItem("text-align", "center", shadcn.ToggleDefault, shadcn.ToggleSizeDefault,
			html.Attrib("aria-label", "Center"), "Center"),
		shadcn.ToggleGroupItem("text-align", "right", shadcn.ToggleDefault, shadcn.ToggleSizeDefault,
			html.Attrib("aria-label", "Right"), "Right"),
	)
}
