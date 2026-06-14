package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

func CollapsibleDemo() mx.Component {
	return shadcn.Collapsible(html.Class("w-full max-w-sm space-y-2"),
		shadcn.CollapsibleTrigger(
			html.Class("flex w-full items-center justify-between rounded-md border px-4 py-2 text-sm font-semibold"),
			"@peduarte starred 3 repositories",
			html.SpanClass("transition-transform group-open:rotate-180", "▾"),
		),
		shadcn.CollapsibleContent(html.Class("space-y-2"),
			html.DivClass("rounded-md border px-4 py-2 font-mono text-sm", "@radix-ui/primitives"),
			html.DivClass("rounded-md border px-4 py-2 font-mono text-sm", "@radix-ui/colors"),
			html.DivClass("rounded-md border px-4 py-2 font-mono text-sm", "@stitches/react"),
		),
	)
}
