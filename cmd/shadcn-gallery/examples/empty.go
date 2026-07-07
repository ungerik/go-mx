package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// EmptyDemo renders an empty state with an icon, title, description and action.
func EmptyDemo() mx.Component {
	return shadcn.Empty(html.Class("border"),
		shadcn.EmptyHeader(
			shadcn.EmptyMedia(shadcn.EmptyMediaIcon,
				// lucide folder-open
				navIcon("m6 14 1.5-2.9A2 2 0 0 1 9.24 10H20a2 2 0 0 1 1.94 2.5l-1.54 6a2 2 0 0 1-1.95 1.5H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h3.9a2 2 0 0 1 1.69.9l.81 1.2a2 2 0 0 0 1.67.9H18a2 2 0 0 1 2 2v2"),
			),
			shadcn.EmptyTitle("No Projects Yet"),
			shadcn.EmptyDescription(
				"You haven't created any projects yet. Get started by creating your first project."),
		),
		shadcn.EmptyContent(
			html.DivClass("flex gap-2",
				shadcn.Button("", "", "Create Project"),
				shadcn.Button(shadcn.ButtonOutline, "", "Import Project"),
			),
		),
	)
}
