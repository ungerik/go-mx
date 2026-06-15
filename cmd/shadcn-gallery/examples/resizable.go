package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// ResizableDemo renders two horizontally resizable panels separated by a drag handle.
func ResizableDemo() mx.Component {
	return shadcn.ResizablePanelGroup(shadcn.ResizeHorizontal,
		html.Class("h-[200px] max-w-md rounded-lg border md:min-w-[450px]"),
		shadcn.ResizablePanel(
			html.DivClass("flex h-full items-center justify-center p-6",
				html.SpanClass("font-semibold", "One")),
		),
		shadcn.ResizableHandle(),
		shadcn.ResizablePanel(
			html.DivClass("flex h-full items-center justify-center p-6",
				html.SpanClass("font-semibold", "Two")),
		),
	)
}
