package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

func AspectRatioDemo() mx.Component {
	return html.DivClass("w-full max-w-md",
		shadcn.AspectRatio(16.0/9.0, html.Class("bg-muted rounded-md"),
			html.ImgSrc("https://images.unsplash.com/photo-1588345921523-c2dcdb7f1dcd?w=800&dpr=2&q=80",
				html.Alt("Photo by Drew Beamer"),
				html.Class("h-full w-full rounded-md object-cover"),
			),
		),
	)
}
