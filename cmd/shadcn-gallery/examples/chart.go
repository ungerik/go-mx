package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// visitors is upstream's chart data: desktop and mobile visitors for the
// first half of a year.
var visitors = shadcn.ChartData{
	Categories: []string{"January", "February", "March", "April", "May", "June"},
	Series: []shadcn.ChartSeries{
		{Key: "desktop", Label: "Desktop", Values: []float64{186, 305, 237, 73, 209, 214}},
		{Key: "mobile", Label: "Mobile", Values: []float64{80, 200, 120, 190, 130, 140}},
	},
}

// month shortens a month name to its first three letters for an axis, as
// upstream's examples do with tickFormatter.
func month(name string) string { return name[:3] }

// ChartBarDemo is upstream's Bar Chart - Multiple: grouped bars per month in
// a Card. Hovering a bar shows its value.
func ChartBarDemo() mx.Component {
	return shadcn.Card(html.Class("w-full max-w-lg"),
		shadcn.CardHeader(
			shadcn.CardTitle("Bar Chart - Multiple"),
			shadcn.CardDescription("January - June 2024"),
		),
		shadcn.CardContent(
			shadcn.Chart(shadcn.ChartBar, visitors, shadcn.ChartOptions{
				HideYAxis:      true,
				FormatCategory: month,
				Label:          "Visitors per month",
			}),
		),
	)
}

// ChartBarStacked stacks the bars and adds a legend.
func ChartBarStacked() mx.Component {
	return shadcn.Card(html.Class("w-full max-w-lg"),
		shadcn.CardHeader(
			shadcn.CardTitle("Bar Chart - Stacked + Legend"),
			shadcn.CardDescription("January - June 2024"),
		),
		shadcn.CardContent(
			shadcn.Chart(shadcn.ChartBar, visitors, shadcn.ChartOptions{
				Stacked:        true,
				Legend:         true,
				HideYAxis:      true,
				FormatCategory: month,
				Label:          "Visitors per month",
			}),
		),
	)
}

// ChartLineDemo is upstream's Line Chart - Dots: smooth monotone lines with a
// dot on each value.
func ChartLineDemo() mx.Component {
	return shadcn.Card(html.Class("w-full max-w-lg"),
		shadcn.CardHeader(
			shadcn.CardTitle("Line Chart - Dots"),
			shadcn.CardDescription("January - June 2024"),
		),
		shadcn.CardContent(
			shadcn.Chart(shadcn.ChartLine, visitors, shadcn.ChartOptions{
				Curve:          shadcn.ChartCurveMonotone,
				Dots:           true,
				HideYAxis:      true,
				FormatCategory: month,
				Label:          "Visitors per month",
			}),
		),
	)
}

// ChartAreaStacked is upstream's Area Chart - Stacked, with the value axis.
func ChartAreaStacked() mx.Component {
	return shadcn.Card(html.Class("w-full max-w-lg"),
		shadcn.CardHeader(
			shadcn.CardTitle("Area Chart - Stacked"),
			shadcn.CardDescription("Showing total visitors for the last 6 months"),
		),
		shadcn.CardContent(
			shadcn.Chart(shadcn.ChartArea, visitors, shadcn.ChartOptions{
				Stacked:        true,
				Curve:          shadcn.ChartCurveMonotone,
				Legend:         true,
				FormatCategory: month,
				Label:          "Visitors per month",
			}),
		),
	)
}

// browsers is upstream's pie data: visitors per browser.
var browsers = shadcn.ChartData{
	Categories: []string{"Chrome", "Safari", "Firefox", "Edge", "Other"},
	Series:     []shadcn.ChartSeries{{Key: "visitors", Label: "Visitors", Values: []float64{275, 200, 187, 173, 90}}},
}

// ChartBarTooltip shows upstream's styled tooltip: hovering a month shows a
// card with both series, drawn and shown with CSS alone.
func ChartBarTooltip() mx.Component {
	return shadcn.Card(html.Class("w-full max-w-lg"),
		shadcn.CardHeader(
			shadcn.CardTitle("Tooltip"),
			shadcn.CardDescription("Hover a month"),
		),
		shadcn.CardContent(
			shadcn.Chart(shadcn.ChartBar, visitors, shadcn.ChartOptions{
				Tooltip:        true,
				HideYAxis:      true,
				FormatCategory: month,
				Label:          "Visitors per month",
			}),
		),
	)
}

// ChartPieDonut is upstream's Pie Chart - Donut, a slice per browser with a
// legend.
func ChartPieDonut() mx.Component {
	return shadcn.Card(html.Class("w-full max-w-sm"),
		shadcn.CardHeader(
			shadcn.CardTitle("Pie Chart - Donut"),
			shadcn.CardDescription("January - June 2024"),
		),
		shadcn.CardContent(
			shadcn.Chart(shadcn.ChartPie, browsers, shadcn.ChartOptions{
				Donut:  0.6,
				Legend: true,
				Label:  "Visitors per browser",
			}, html.Class("aspect-square")),
		),
	)
}

// ChartRadarDemo is upstream's Radar Chart, a polygon per series over the
// months.
func ChartRadarDemo() mx.Component {
	return shadcn.Card(html.Class("w-full max-w-sm"),
		shadcn.CardHeader(
			shadcn.CardTitle("Radar Chart - Multiple"),
			shadcn.CardDescription("Showing total visitors for the last 6 months"),
		),
		shadcn.CardContent(
			shadcn.Chart(shadcn.ChartRadar, visitors, shadcn.ChartOptions{
				Legend:         true,
				FormatCategory: month,
				Label:          "Visitors per month",
			}, html.Class("aspect-square")),
		),
	)
}

// ChartRadialDemo is upstream's Radial Chart, a ring per browser.
func ChartRadialDemo() mx.Component {
	return shadcn.Card(html.Class("w-full max-w-sm"),
		shadcn.CardHeader(
			shadcn.CardTitle("Radial Chart"),
			shadcn.CardDescription("January - June 2024"),
		),
		shadcn.CardContent(
			shadcn.Chart(shadcn.ChartRadial, browsers, shadcn.ChartOptions{
				Legend: true,
				Label:  "Visitors per browser",
			}, html.Class("aspect-square")),
		),
	)
}
