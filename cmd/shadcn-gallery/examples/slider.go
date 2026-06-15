package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// SliderDemo renders a single-thumb slider.
func SliderDemo() mx.Component {
	return html.DivClass("w-full max-w-sm",
		shadcn.Slider(0, 100, 1, []float64{50}, "volume"),
	)
}

// SliderRange renders a two-thumb range slider.
func SliderRange() mx.Component {
	return html.DivClass("w-full max-w-sm",
		shadcn.Slider(0, 100, 1, []float64{25, 75}, "price-range"),
	)
}
