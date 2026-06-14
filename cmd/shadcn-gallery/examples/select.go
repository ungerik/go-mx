package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

func SelectDemo() mx.Component {
	return shadcn.Select(html.Name("fruit"), html.Class("w-[180px]"),
		shadcn.SelectGroup("Fruits",
			shadcn.SelectOption("apple", "Apple"),
			shadcn.SelectOption("banana", "Banana"),
			shadcn.SelectOption("blueberry", "Blueberry"),
			shadcn.SelectOption("grapes", "Grapes"),
			shadcn.SelectOption("pineapple", "Pineapple"),
		),
	)
}
