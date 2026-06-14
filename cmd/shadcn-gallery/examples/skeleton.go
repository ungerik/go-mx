package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

func SkeletonDemo() mx.Component {
	return html.DivClass("flex items-center space-x-4",
		shadcn.Skeleton(html.Class("size-12 rounded-full")),
		html.DivClass("space-y-2",
			shadcn.Skeleton(html.Class("h-4 w-[250px]")),
			shadcn.Skeleton(html.Class("h-4 w-[200px]")),
		),
	)
}
