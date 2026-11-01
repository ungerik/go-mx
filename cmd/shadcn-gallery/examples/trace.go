package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// TraceDemo renders an agent's steps: one running, one linking to its
// details, and one disclosing a command's output.
func TraceDemo() mx.Component {
	return shadcn.Trace(html.Class("max-w-md"),
		shadcn.TraceLine(shadcn.TraceRunning,
			shadcn.TraceIcon(iconSearch()),
			shadcn.TraceTitle("Scanning 52 files"),
		),
		shadcn.TraceCollapsible(
			shadcn.TraceCollapsibleTrigger(
				shadcn.TraceLine("",
					shadcn.TraceIcon(iconTerminal()),
					shadcn.TraceTitle("Ran go version"),
					shadcn.TraceDescription("0.4s"),
				),
			),
			shadcn.TraceCollapsibleContent(
				shadcn.CodeBlock("", shadcn.CodeBlockSizeSM,
					shadcn.CodeBlockContent(shadcn.CodeBlockWrap, "go version go1.27.0 darwin/arm64"),
				),
			),
		),
		html.A(html.HRef("#trace"),
			shadcn.TraceLine("",
				shadcn.TraceIcon(iconFileText()),
				shadcn.TraceTitle("Read shadcn/trace.go"),
				shadcn.TraceIcon(iconChevronRight()),
			),
		),
		shadcn.TraceLine(shadcn.TracePending,
			shadcn.TraceIcon(iconGlobe()),
			shadcn.TraceTitle("Search the web for release notes"),
		),
	)
}

// TraceToolCalls renders tool calls that disclose their input and output,
// including a failed one.
func TraceToolCalls() mx.Component {
	return shadcn.Trace(html.Class("max-w-md"),
		shadcn.TraceCollapsible(
			shadcn.TraceCollapsibleTrigger(
				shadcn.TraceLine("",
					shadcn.TraceIcon(iconGlobe()),
					shadcn.TraceTitle("Searched the web for htmx sse extension"),
				),
			),
			shadcn.TraceCollapsibleContent(html.Class("flex flex-wrap gap-1.5"),
				shadcn.Badge(shadcn.BadgeSecondary, "htmx.org/extensions/sse"),
				shadcn.Badge(shadcn.BadgeSecondary, "github.com/bigskysoftware/htmx-extensions"),
			),
		),
		shadcn.TraceCollapsible(html.Open,
			shadcn.TraceCollapsibleTrigger(
				shadcn.TraceLine("",
					shadcn.TraceIcon(iconTerminal()),
					shadcn.TraceTitle("Ran go test ./shadcn/"),
				),
			),
			shadcn.TraceCollapsibleContent(
				shadcn.CodeBlock("", shadcn.CodeBlockSizeSM,
					shadcn.CodeBlockContent(shadcn.CodeBlockWrap, "$ go test ./shadcn/\nok  \tgithub.com/ungerik/go-mx/shadcn\t0.35s"),
				),
			),
		),
		shadcn.TraceCollapsible(
			shadcn.TraceCollapsibleTrigger(
				shadcn.TraceLine(shadcn.TraceError,
					shadcn.TraceIcon(iconTerminal()),
					shadcn.TraceTitle("Failed to run pnpm --version"),
				),
			),
			shadcn.TraceCollapsibleContent(
				shadcn.CodeBlock(shadcn.CodeBlockDestructive, shadcn.CodeBlockSizeSM,
					shadcn.CodeBlockContent(shadcn.CodeBlockWrap, "command not found: pnpm"),
				),
			),
		),
	)
}
