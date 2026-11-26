package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn/magicui"
)

// KineticTextDemo is a large word whose letters swell under the pointer.
func KineticTextDemo() mx.Component {
	return html.DivClass("relative justify-center",
		magicui.KineticText(magicui.KineticTextOptions{Text: "Nostalgia"},
			html.Class("text-[6rem] tracking-[-5%] [font-optical-sizing:auto]"),
		),
	)
}
