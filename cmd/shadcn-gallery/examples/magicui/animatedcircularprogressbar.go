package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn/magicui"
)

// AnimatedCircularProgressBarDemo shows gauges in upstream's demo colors.
// Upstream's demo steps one gauge from 0 to 100 every two seconds with a
// React interval; without a script the page shows three values instead.
func AnimatedCircularProgressBarDemo() mx.Component {
	gauge := func(value float64) mx.Component {
		return magicui.AnimatedCircularProgressBar(magicui.AnimatedCircularProgressBarOptions{
			Value:               value,
			GaugePrimaryColor:   "rgb(79 70 229)",
			GaugeSecondaryColor: "rgba(0, 0, 0, 0.1)",
		})
	}
	return html.DivClass("flex flex-wrap items-center justify-center gap-8", gauge(30), gauge(60), gauge(100))
}
