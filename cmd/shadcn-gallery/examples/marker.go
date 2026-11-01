package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// MarkerDemo renders default markers: plain, with an icon, reporting progress
// with a spinner, and as a link.
func MarkerDemo() mx.Component {
	return html.DivClass("flex w-full max-w-md flex-col gap-6",
		shadcn.Marker("", shadcn.MarkerContent("A default marker")),
		shadcn.Marker("",
			shadcn.MarkerIcon(iconFileText()),
			shadcn.MarkerContent("Marker with icon"),
		),
		shadcn.Marker("", html.Role("status"),
			shadcn.MarkerIcon(shadcn.Spinner()),
			shadcn.MarkerContent(html.Class("animate-pulse"), "Thinking…"),
		),
		shadcn.Marker("", html.HRef("#marker"),
			shadcn.MarkerIcon(iconClock()),
			shadcn.MarkerContent("Marker as a link"),
		),
		shadcn.Marker("", html.Class("justify-center"),
			shadcn.MarkerContent(html.Strong(html.Class("font-medium"), "Olivia Rose"), " left the chat"),
		),
	)
}

// MarkerSeparatorDemo renders markers between rules, as dividers in a
// transcript.
func MarkerSeparatorDemo() mx.Component {
	return html.DivClass("flex w-full max-w-md flex-col gap-6",
		shadcn.Marker(shadcn.MarkerSeparator, shadcn.MarkerContent("Worked for 42s")),
		shadcn.Marker(shadcn.MarkerSeparator, html.Role("status"),
			shadcn.MarkerIcon(shadcn.Spinner()),
			shadcn.MarkerContent("Compacting conversation"),
		),
		shadcn.Marker(shadcn.MarkerSeparator,
			shadcn.MarkerIcon(iconCheck()),
			shadcn.MarkerContent("Conversation compacted"),
		),
		shadcn.Marker(shadcn.MarkerSeparator,
			shadcn.MarkerContent("With a ", html.A(html.HRef("#marker"), "link to learn more")),
		),
	)
}

// MarkerBorderDemo renders markers with a rule below, as a list of steps.
func MarkerBorderDemo() mx.Component {
	return html.DivClass("flex w-full max-w-md flex-col gap-3",
		shadcn.Marker(shadcn.MarkerBorder,
			shadcn.MarkerIcon(iconTerminal()),
			shadcn.MarkerContent("Switched to release-candidate"),
		),
		shadcn.Marker(shadcn.MarkerBorder,
			shadcn.MarkerIcon(iconSearch()),
			shadcn.MarkerContent("Reviewed 8 related files"),
		),
		shadcn.Marker(shadcn.MarkerBorder,
			shadcn.MarkerIcon(iconFileText()),
			shadcn.MarkerContent("Opened implementation notes"),
		),
	)
}
