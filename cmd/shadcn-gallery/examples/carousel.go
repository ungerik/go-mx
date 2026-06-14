package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

func CarouselDemo() mx.Component {
	return shadcn.Carousel(html.Class("mx-auto w-full max-w-xs"),
		shadcn.CarouselContent(
			mx.ForEach([]string{"1", "2", "3", "4", "5"}, func(n string) mx.Component {
				return shadcn.CarouselItem(
					html.Div(html.Class("p-1"),
						shadcn.Card(
							shadcn.CardContent(html.Class("flex aspect-square items-center justify-center p-6"),
								html.Span(html.Class("text-4xl font-semibold"), n),
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
