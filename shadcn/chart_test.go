package shadcn

import (
	"context"
	"math"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/ungerik/go-mx"
)

// chartSample is upstream's example data: desktop and mobile visitors per
// month.
var chartSample = ChartData{
	Categories: []string{"January", "February", "March"},
	Series: []ChartSeries{
		{Key: "desktop", Label: "Desktop", Values: []float64{186, 305, 237}},
		{Key: "mobile", Label: "Mobile", Values: []float64{80, 200, 120}},
	},
}

func TestChartBar(t *testing.T) {
	out := render(t, Chart(ChartBar, chartSample, ChartOptions{Label: "Visitors"}))
	for _, want := range []string{
		`data-slot="chart"`,
		`data-chart="bar"`,
		// The series colors default to the theme's chart variables.
		`style="--color-desktop: var(--chart-1); --color-mobile: var(--chart-2);"`,
		`fill="var(--color-desktop)"`,
		`<title>February: Desktop 305</title>`,
		`<title>March: Mobile 120</title>`,
		`aria-hidden="true"`,
		// The axes: categories and round value ticks from 0 to 350.
		`>January</text>`,
		`>350</text>`,
		// The data table read instead of the SVG.
		`<table class="sr-only"><caption>Visitors</caption>`,
		`<th scope="col">Desktop</th>`,
		`<th scope="row">February</th><td>305</td><td>200</td>`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %s in %s", want, out)
		}
	}
	if c := strings.Count(out, "<rect "); c != 6 {
		t.Errorf("want a bar per series and category, got %d", c)
	}
}

func TestChartOptions(t *testing.T) {
	out := render(t, Chart(ChartBar, chartSample, ChartOptions{
		HideGrid: true, HideYAxis: true, Legend: true, Stacked: true,
		FormatCategory: func(c string) string { return c[:3] },
		FormatValue:    func(v float64) string { return "≈" + strconv.FormatFloat(v, 'f', -1, 64) },
	}))
	if strings.Contains(out, "<line ") {
		t.Errorf("HideGrid draws no grid lines: %s", out)
	}
	if !strings.Contains(out, `>Jan</text>`) || strings.Contains(out, `>January</text>`) {
		t.Errorf("FormatCategory shortens the axis labels: %s", out)
	}
	// The tooltips and the table keep the full category.
	if !strings.Contains(out, `<title>January: Desktop ≈186</title>`) {
		t.Errorf("tooltips use the full category and FormatValue: %s", out)
	}
	if !strings.Contains(out, `data-slot="chart-legend"`) || !strings.Contains(out, `background-color: var(--color-mobile)`) {
		t.Errorf("Legend shows the series with their colors: %s", out)
	}
	// Stacked bars round only the top segment.
	if c := strings.Count(out, ` rx="`); c != 3 {
		t.Errorf("want the 3 top segments rounded, got %d", c)
	}
}

func TestChartLineAndArea(t *testing.T) {
	line := render(t, Chart(ChartLine, chartSample, ChartOptions{Curve: ChartCurveMonotone, Dots: true}))
	if !strings.Contains(line, ` C`) || !strings.Contains(line, `fill="none"`) {
		t.Errorf("a monotone line is a cubic path: %s", line)
	}
	if c := strings.Count(line, `r="4"`); c != 6 {
		t.Errorf("Dots draws a dot per point, got %d", c)
	}
	// Without Dots each point still carries its tooltip on a transparent
	// target.
	plain := render(t, Chart(ChartLine, chartSample, ChartOptions{}))
	if c := strings.Count(plain, `fill="transparent"`); c != 6 || !strings.Contains(plain, `<title>January: Mobile 80</title>`) {
		t.Errorf("tooltip targets: %s", plain)
	}
	area := render(t, Chart(ChartArea, chartSample, ChartOptions{Stacked: true, Curve: ChartCurveStep}))
	if c := strings.Count(area, `fill-opacity="0.4"`); c != 2 {
		t.Errorf("want an area per series, got %d", c)
	}
	if !strings.Contains(area, " H") || !strings.Contains(area, " Z") {
		t.Errorf("a step area is a closed path of steps: %s", area)
	}
	// Stacked, the largest stack (505) sets the range.
	if !strings.Contains(area, `>600</text>`) {
		t.Errorf("a stacked area reaches the largest sum: %s", area)
	}
}

func TestChartInvalidData(t *testing.T) {
	for name, data := range map[string]ChartData{
		"value count": {Categories: []string{"a", "b"}, Series: []ChartSeries{{Values: []float64{1}}}},
		"NaN":         {Categories: []string{"a"}, Series: []ChartSeries{{Values: []float64{math.NaN()}}}},
	} {
		var b strings.Builder
		if err := Chart(ChartBar, data, ChartOptions{}).Render(context.Background(), mx.NewCheckedWriter(&b)); err == nil {
			t.Errorf("%s: want a render error", name)
		}
	}
	// A key is an id, which panics when invalid, as other ids do (see
	// PanicOnInvalidID).
	func() {
		defer func() {
			if recover() == nil {
				t.Error("an invalid key must panic")
			}
		}()
		Chart(ChartBar, ChartData{Categories: []string{"a"}, Series: []ChartSeries{{Key: "a b", Values: []float64{1}}}}, ChartOptions{})
	}()
	// A key left empty and a missing label and color get defaults.
	out := render(t, Chart(ChartBar, ChartData{Categories: []string{"a"}, Series: []ChartSeries{{Values: []float64{-3}}}}, ChartOptions{}))
	if !strings.Contains(out, `--color-series-1: var(--chart-1)`) || !strings.Contains(out, `<title>a: series-1 -3</title>`) {
		t.Errorf("defaults: %s", out)
	}
}

func TestChartTicks(t *testing.T) {
	for _, tt := range []struct {
		lo, hi float64
		want   []float64
	}{
		{0, 305, []float64{0, 50, 100, 150, 200, 250, 300, 350}},
		{0, 1, []float64{0, 0.2, 0.4, 0.6, 0.8, 1}},
		{-30, 80, []float64{-40, -20, 0, 20, 40, 60, 80}},
		{0, 0, []float64{0, 0.2, 0.4, 0.6, 0.8, 1}},
	} {
		if got := chartTicks(tt.lo, tt.hi, 5); !slices.Equal(got, tt.want) {
			t.Errorf("chartTicks(%v, %v) = %v, want %v", tt.lo, tt.hi, got, tt.want)
		}
	}
}

func TestChartTangentIsMonotone(t *testing.T) {
	points := [][2]float64{{0, 0}, {1, 10}, {2, 10}, {3, 0}}
	if got := chartTangent(points, 1); got != 0 {
		t.Errorf("a flat neighbor gives a zero tangent, got %v", got)
	}
	if got := chartTangent([][2]float64{{0, 0}, {1, 1}, {2, 3}}, 1); got <= 1 || got >= 2 {
		t.Errorf("the tangent lies between the neighboring slopes, got %v", got)
	}
}

// browsers is upstream's pie data: visitors per browser, one series.
var browsers = ChartData{
	Categories: []string{"Chrome", "Safari", "Firefox"},
	Series:     []ChartSeries{{Key: "visitors", Label: "Visitors", Values: []float64{275, 200, 0}}},
}

func TestChartPolar(t *testing.T) {
	pie := render(t, Chart(ChartPie, browsers, ChartOptions{Legend: true}))
	// A slice per category with a value, colored per category.
	if c := strings.Count(pie, "<path "); c != 2 {
		t.Errorf("want 2 slices, got %d: %s", c, pie)
	}
	for _, want := range []string{`fill="var(--chart-1)"`, `fill="var(--chart-2)"`, `<title>Safari: 200</title>`, `data-chart="pie"`,
		`background-color: var(--chart-3)`, ">Firefox</div>"} {
		if !strings.Contains(pie, want) {
			t.Errorf("pie lacks %s: %s", want, pie)
		}
	}
	donut := render(t, Chart(ChartPie, ChartData{Categories: []string{"a"}, Series: []ChartSeries{{Values: []float64{1}}}, CategoryColors: []string{"red"}}, ChartOptions{Donut: 0.6}))
	if !strings.Contains(donut, `fill="red"`) || strings.Count(donut, " A") != 2 {
		t.Errorf("a donut slice has an outer and an inner arc: %s", donut)
	}
	radial := render(t, Chart(ChartRadial, browsers, ChartOptions{}))
	if c := strings.Count(radial, "<circle "); c != 3 {
		t.Errorf("want a track per category, got %d: %s", c, radial)
	}
	if !strings.Contains(radial, `stroke-linecap="round"`) || !strings.Contains(radial, `<title>Chrome: 275</title>`) {
		t.Errorf("radial: %s", radial)
	}
	radar := render(t, Chart(ChartRadar, chartSample, ChartOptions{Dots: true}))
	if c := strings.Count(radar, "<polygon "); c < 3 {
		t.Errorf("want grid and series polygons, got %d: %s", c, radar)
	}
	for _, want := range []string{`>January</text>`, `fill-opacity="0.6"`, `<title>March: Mobile 120</title>`, `r="4"`} {
		if !strings.Contains(radar, want) {
			t.Errorf("radar lacks %s: %s", want, radar)
		}
	}
}

// The styled tooltip replaces the native ones with a card per category that
// shows on hover with CSS alone.
func TestChartTooltip(t *testing.T) {
	out := render(t, Chart(ChartBar, chartSample, ChartOptions{Tooltip: true}))
	if strings.Contains(out, "<title>") {
		t.Errorf("no native tooltips with the styled one: %s", out)
	}
	if c := strings.Count(out, `class="group/chart-band"`); c != 3 {
		t.Errorf("want a band per category, got %d", c)
	}
	for _, want := range []string{"group-hover/chart-band:opacity-100", `>February</text>`, `>305</text>`, `>Mobile</text>`} {
		if !strings.Contains(out, want) {
			t.Errorf("tooltip lacks %s: %s", want, out)
		}
	}
	if line := render(t, Chart(ChartLine, chartSample, ChartOptions{Tooltip: true})); !strings.Contains(line, `class="stroke-border" stroke="currentColor"`) {
		t.Errorf("a line chart's tooltip has a cursor line: %s", line)
	}
}
