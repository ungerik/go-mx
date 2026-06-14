package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

func SliderDemo() mx.Component {
	return html.Div(html.Class("w-full max-w-sm"),
		shadcn.Slider(0, 100, 1, []float64{50}, "volume"),
	)
}

func SliderRange() mx.Component {
	return html.Div(html.Class("w-full max-w-sm"),
		shadcn.Slider(0, 100, 1, []float64{25, 75}, "price-range"),
	)
}
