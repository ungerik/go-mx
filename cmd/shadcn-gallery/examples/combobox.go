package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// comboItem is one Combobox option: a check (visible only when selected) plus
// the label.
func comboItem(label string, selected bool) mx.Component {
	cls := "mr-2"
	if !selected {
		cls += " opacity-0"
	}
	return shadcn.CommandItem(
		html.Span(html.Class(cls), navIcon("M20 6 9 17l-5-5")),
		html.Span(label),
	)
}

// ComboboxDemo is shadcn's Combobox recipe: a Popover whose trigger shows the
// selected value and whose content is a filterable Command. shadcn ships
// Combobox as a copy-paste composition, not an exported component, so this is a
// composition example rather than a new shadcn primitive.
func ComboboxDemo() mx.Component {
	return shadcn.Popover(
		shadcn.PopoverTrigger("demo-combobox",
			html.Class(shadcn.ButtonClasses(shadcn.ButtonOutline, shadcn.SizeDefault)+" w-[200px] justify-between"),
			html.Span("Next.js"),
			html.Span(html.Class("opacity-50"), navIcon("m7 15 5 5 5-5", "m7 9 5-5 5 5")),
		),
		shadcn.PopoverContent("demo-combobox", "",
			html.Class("w-[200px] p-0"),
			shadcn.Command(
				shadcn.CommandInput(html.Attrib("placeholder", "Search framework...")),
				shadcn.CommandList(
					shadcn.CommandEmpty("No framework found."),
					shadcn.CommandGroup("",
						comboItem("Next.js", true),
						comboItem("SvelteKit", false),
						comboItem("Nuxt.js", false),
						comboItem("Remix", false),
						comboItem("Astro", false),
					),
				),
			),
		),
	)
}
