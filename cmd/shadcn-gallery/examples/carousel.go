package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// CarouselDemo renders a basic card carousel with previous and next navigation arrows.
func CarouselDemo() mx.Component {
	return shadcn.Carousel(html.Class("mx-auto w-full max-w-xs"),
		shadcn.CarouselContent(
			mx.ForEach([]string{"1", "2", "3", "4", "5"}, func(n string) mx.Component {
				return shadcn.CarouselItem(
					html.DivClass("p-1",
						shadcn.Card(
							shadcn.CardContent(html.Class("flex aspect-square items-center justify-center p-6"),
								html.SpanClass("text-4xl font-semibold", n),
							),
						),
					),
				)
			}),
		),
		shadcn.CarouselPrevious(),
		shadcn.CarouselNext(),
	)
}
