package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn/magicui"
)

// ScrollProgressDemo shows the page's scroll progress as a bar below the
// top of the viewport.
func ScrollProgressDemo() mx.Component {
	return html.DivClass("z-10 rounded-lg p-4",
		magicui.ScrollProgress(html.Class("top-[65px]")),
		html.H2Class("pb-4 font-bold", "Note: The scroll progress is shown below the navbar of the page."),
	)
}
