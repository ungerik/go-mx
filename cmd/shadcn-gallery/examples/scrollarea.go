package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

func ScrollAreaDemo() mx.Component {
	return shadcn.ScrollArea(html.Class("h-48 w-56 rounded-md border p-4 text-sm"),
		html.DivClass("mb-3 font-medium leading-none", "Tags"),
		mx.ForEach([]string{
			"v1.2.0-beta.50", "v1.2.0-beta.49", "v1.2.0-beta.48", "v1.2.0-beta.47",
			"v1.2.0-beta.46", "v1.2.0-beta.45", "v1.2.0-beta.44", "v1.2.0-beta.43",
			"v1.2.0-beta.42", "v1.2.0-beta.41", "v1.2.0-beta.40", "v1.2.0-beta.39",
		}, func(tag string) mx.Component {
			return html.DivClass("border-b py-1.5 last:border-b-0", tag)
		}),
	)
}
