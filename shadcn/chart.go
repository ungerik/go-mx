//go:generate go -C ../tools tool go-enum ../shadcn/$GOFILE

package shadcn

import (
	"cmp"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/svg"
)

// Chart is a Go port of shadcn/ui's Chart. Upstream's Chart is a theming
// layer over Recharts: ChartContainer sets a --color-<key> CSS variable per
// series from a ChartConfig and styles the SVG Recharts draws in the
// browser. Here the SVG is drawn in Go, so a chart is plain server-rendered
// markup with no client runtime: scales, axis ticks, grid lines, bars, lines
// and areas are computed from [ChartData] and colored through the same
// --color-<key> variables, by default the theme's --chart-1 … --chart-5.
//
// The SVG scales to its container through its viewBox, the replacement for
// Recharts' ResponsiveContainer, and the container keeps upstream's
// aspect-video default; pass a class such as "h-[250px] w-full" to size it
// otherwise. Each data point carries a native <title>, so hovering shows its
// value without a script, or with [ChartOptions].Tooltip upstream's styled
// ChartTooltip, shown with CSS alone. The SVG is hidden from assistive
// technology, which reads a visually hidden table of the same data instead.
//
// Bar, line and area charts (grouped or stacked), pie and donut, radar and
// radial charts are ported.

// ChartKind selects how a [Chart] draws its series.
type ChartKind string //#enum

const (
	// ChartBar draws a bar per series and category, grouped side by side,
	// or stacked with [ChartOptions].Stacked.
	ChartBar ChartKind = "bar"
	// ChartLine draws a line per series through its values.
	ChartLine ChartKind = "line"
	// ChartArea draws a filled area per series under its line, overlapping,
	// or stacked with [ChartOptions].Stacked.
	ChartArea ChartKind = "area"
	// ChartPie draws the first series as a pie, a slice per category, or a
	// donut with [ChartOptions].Donut.
	ChartPie ChartKind = "pie"
	// ChartRadar draws a polygon per series over a spoke per category.
	ChartRadar ChartKind = "radar"
	// ChartRadial draws the first series as concentric rings, one per
	// category, each filled by its share of the largest value.
	ChartRadial ChartKind = "radial"
)

// Valid indicates if c is any of the valid values for ChartKind
func (c ChartKind) Valid() bool {
	switch c {
	case
		ChartBar,
		ChartLine,
		ChartArea,
		ChartPie,
		ChartRadar,
		ChartRadial:
		return true
	}
	return false
}

// Validate returns an error if c is none of the valid values for ChartKind
func (c ChartKind) Validate() error {
	if !c.Valid() {
		return fmt.Errorf("invalid value %#v for type shadcn.ChartKind", c)
	}
	return nil
}

// ValidOr returns c if it is any of the valid values for ChartKind, else defaultVal
func (c ChartKind) ValidOr(defaultVal ChartKind) ChartKind {
	if c.Valid() {
		return c
	}
	return defaultVal
}

// Enums returns all valid values for ChartKind
func (ChartKind) Enums() []ChartKind {
	return []ChartKind{
		ChartBar,
		ChartLine,
		ChartArea,
		ChartPie,
		ChartRadar,
		ChartRadial,
	}
}

// EnumStrings returns all valid values for ChartKind as strings
func (ChartKind) EnumStrings() []string {
	return []string{
		"bar",
		"line",
		"area",
		"pie",
		"radar",
		"radial",
	}
}

// String implements the fmt.Stringer interface for ChartKind
func (c ChartKind) String() string {
	return string(c)
}

// ChartCurve selects how a line or area [Chart] connects its points.
type ChartCurve string //#enum

const (
	// ChartCurveLinear connects the points with straight lines, the
	// default.
	ChartCurveLinear ChartCurve = "linear"
	// ChartCurveMonotone connects them with a smooth curve that does not
	// overshoot the values, Recharts' "monotone".
	ChartCurveMonotone ChartCurve = "monotone"
	// ChartCurveStep connects them with steps, Recharts' "step".
	ChartCurveStep ChartCurve = "step"
)

// Valid indicates if c is any of the valid values for ChartCurve
func (c ChartCurve) Valid() bool {
	switch c {
	case
		ChartCurveLinear,
		ChartCurveMonotone,
		ChartCurveStep:
		return true
	}
	return false
}

// Validate returns an error if c is none of the valid values for ChartCurve
func (c ChartCurve) Validate() error {
	if !c.Valid() {
		return fmt.Errorf("invalid value %#v for type shadcn.ChartCurve", c)
	}
	return nil
}

// ValidOr returns c if it is any of the valid values for ChartCurve, else defaultVal
func (c ChartCurve) ValidOr(defaultVal ChartCurve) ChartCurve {
	if c.Valid() {
		return c
	}
	return defaultVal
}

// Enums returns all valid values for ChartCurve
func (ChartCurve) Enums() []ChartCurve {
	return []ChartCurve{
		ChartCurveLinear,
		ChartCurveMonotone,
		ChartCurveStep,
	}
}

// EnumStrings returns all valid values for ChartCurve as strings
func (ChartCurve) EnumStrings() []string {
	return []string{
		"linear",
		"monotone",
		"step",
	}
}

// String implements the fmt.Stringer interface for ChartCurve
func (c ChartCurve) String() string {
	return string(c)
}

// ChartSeries is one series of a [Chart]: a value per category, upstream's
// data key together with its ChartConfig entry.
type ChartSeries struct {
	// Key names the series' CSS variable, --color-<Key>, and so must be a
	// valid id (letters, digits, "-", "_"). "" uses "series-<n>", counted
	// from 1.
	Key string
	// Label names the series in the legend, the tooltips and the data
	// table. "" uses Key.
	Label string
	// Color is the series' CSS color. "" uses the theme's --chart-<n>
	// variable, cycling through --chart-1 … --chart-5. It is written into a
	// style attribute, so it must be a single CSS value: one holding ";",
	// "{", "}", a quote, a backslash, "/*" or unbalanced parentheses is
	// invalid.
	Color string
	// Values holds one value per category of the [ChartData]. NaN and
	// infinities are invalid.
	Values []float64
}

// ChartData is what a [Chart] draws: categories along the x axis, such as
// months, and the series that have a value for each.
type ChartData struct {
	Categories []string
	Series     []ChartSeries
	// CategoryColors colors the categories of a pie or radial chart, which
	// draw a slice or ring per category of their one series. "" or a
	// missing entry uses the theme's --chart-<n> variable, cycling. A color
	// is invalid as a series' Color is.
	CategoryColors []string
}

// ChartOptions configures a [Chart]. The zero value draws grouped bars or
// overlapping areas with grid lines, both axes and five value ticks in a
// 16:9 SVG.
type ChartOptions struct {
	// Width and Height are the SVG's viewBox in pixels, which also sets the
	// size of its text relative to the chart. 0 uses 480 and 270, which
	// keeps 12px text readable in a card, and 300 and 300 for a pie, radar
	// or radial chart, which is round.
	Width, Height int
	// Stacked stacks the bars or areas of the series on each other instead
	// of grouping or overlapping them.
	Stacked bool
	// Curve connects the points of a line or area chart; "" is
	// ChartCurveLinear.
	Curve ChartCurve
	// Dots draws a dot on each point of a line chart.
	Dots bool
	// HideGrid, HideXAxis and HideYAxis leave out the horizontal grid lines,
	// the category labels and the value labels.
	HideGrid, HideXAxis, HideYAxis bool
	// Legend renders the series' labels with their colors below the chart,
	// upstream's ChartLegend.
	Legend bool
	// Ticks is the number of value ticks wanted; the axis picks the nearest
	// round steps. 0 uses 5.
	Ticks int
	// FormatValue formats a value for the value axis, the tooltips and the
	// data table. nil uses the shortest decimal form.
	FormatValue func(float64) string
	// FormatCategory formats a category for the category axis, such as
	// upstream's examples shortening months to three letters. The tooltips
	// and the data table show the full category. nil shows it as it is.
	FormatCategory func(string) string
	// Label is the caption of the data table that assistive technology
	// reads instead of the SVG. "" leaves the table without a caption.
	Label string

	// Tooltip shows upstream's styled ChartTooltip on a bar, line or area
	// chart instead of the native per-point tooltips: hovering a category
	// shows a card with it and every series' color, label and value, with
	// CSS alone.
	Tooltip bool

	// Donut is the inner radius of a pie chart as a fraction of its radius,
	// making it a donut: 0.6 is upstream's donut. 0 draws a full pie.
	Donut float64
}

// chartClasses are the styles of the chart container, upstream's
// ChartContainer classes minus the selectors of Recharts' own markup, which
// this port's SVG does not have.
const chartClasses = "flex aspect-video flex-col justify-center text-xs"

// Chart renders data as a kind of chart. It returns an [mx.NewErrElement] for
// data it cannot draw: a series whose value count differs from the category
// count, a value that is NaN, infinite or beyond ±1e300, values too close
// together to scale, or a color that is no single CSS value. A Key that is no valid id is
// reported like any other invalid id of this package (see
// [PanicOnInvalidID]).
//
// The chart is a [Card]-free container like upstream's ChartContainer: put it
// in a Card with a header and footer as upstream's examples do.
func Chart(kind ChartKind, data ChartData, opts ChartOptions, attribsChildren ...any) *mx.Element {
	kind = kind.ValidOr(ChartBar)
	series, err := chartSeries(data)
	if err != nil {
		return mx.NewErrElement(err)
	}
	polar := kind == ChartPie || kind == ChartRadar || kind == ChartRadial
	width, height := float64(opts.Width), float64(opts.Height)
	switch {
	case width <= 0 && polar:
		width = 300
	case width <= 0:
		width = 480
	}
	switch {
	case height <= 0 && polar:
		height = 300
	case height <= 0:
		height = 270
	}
	format := opts.FormatValue
	if format == nil {
		format = func(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }
	}
	formatCategory := opts.FormatCategory
	if formatCategory == nil {
		formatCategory = func(c string) string { return c }
	}
	var chart *mx.Element
	if polar {
		chart = chartPolarSVG(kind, data, series, opts, width, height, format, formatCategory)
	} else {
		chart = chartCartesianSVG(kind, data, series, opts, width, height, format, formatCategory)
	}

	// The variables of the series, as upstream's ChartStyle sets them.
	var style strings.Builder
	for _, s := range series {
		fmt.Fprintf(&style, "--color-%s: %s; ", s.key, s.color)
	}
	structure := []any{chart}
	if opts.Legend {
		legend := html.DivClass("flex flex-wrap items-center justify-center gap-4 pt-3")
		item := func(color, label string) {
			legend.Children = append(legend.Children, html.DivClass("flex items-center gap-1.5",
				html.Div(html.Class("size-2 shrink-0 rounded-[2px]"), html.Style("background-color: "+color)),
				label,
			))
		}
		if kind == ChartPie || kind == ChartRadial {
			// A slice or ring per category: the legend names the categories.
			for i, c := range data.Categories {
				item(chartCategoryColor(data, i), c)
			}
		} else {
			for _, s := range series {
				item("var(--color-"+s.key+")", s.label)
			}
		}
		structure = append(structure, finish(legend, "chart-legend", ""))
	}
	structure = append(structure, chartTable(data, series, opts.Label, format))
	e := html.Div(append(structure, attribsChildren...)...)
	mergeStyle(e, strings.TrimSpace(style.String()))
	e.Attribs = append(e.Attribs, html.DataAttr("chart", string(kind)))
	return finish(e, "chart", chartClasses)
}

// chartCartesianSVG draws a bar, line or area chart of series over data's
// categories in a width × height viewBox.
func chartCartesianSVG(kind ChartKind, data ChartData, series []chartSeriesInfo, opts ChartOptions, width, height float64, format func(float64) string, formatCategory func(string) string) *mx.Element {
	stacked := opts.Stacked && kind != ChartLine
	n := len(data.Categories)

	// The value range: from 0, or below for negative values, to the
	// largest value or stack.
	lo, hi := 0.0, 0.0
	for i := range n {
		pos, neg := 0.0, 0.0
		for _, s := range series {
			v := s.Values[i]
			if stacked {
				if v >= 0 {
					pos += v
				} else {
					neg += v
				}
			} else {
				pos, neg = max(pos, v), min(neg, v)
			}
		}
		hi, lo = max(hi, pos), min(lo, neg)
	}
	ticks := chartTicks(lo, hi, cmp.Or(opts.Ticks, 5))
	if len(ticks) < 2 {
		return mx.NewErrElement(fmt.Errorf("shadcn.Chart: values from %g to %g have no scale", lo, hi))
	}
	lo, hi = ticks[0], ticks[len(ticks)-1]

	// The plot area inside the axis labels.
	const fontSize = 12
	left, right, top, bottom := 12.0, 12.0, 12.0, 12.0
	if !opts.HideYAxis {
		longest := 0
		for _, t := range ticks {
			longest = max(longest, len([]rune(format(t))))
		}
		left = float64(longest)*fontSize*0.6 + 16
	}
	if !opts.HideXAxis {
		bottom = fontSize + 18
	}
	plotW, plotH := width-left-right, height-top-bottom
	y := func(v float64) float64 { return top + plotH - (v-lo)/(hi-lo)*plotH }
	// Bars take a band per category; lines and areas a point per category,
	// spanning the plot area from edge to edge.
	band := plotW / float64(max(n, 1))
	x := func(i int) float64 {
		if kind == ChartBar {
			return left + band*(float64(i)+0.5)
		}
		if n == 1 {
			return left + plotW/2
		}
		return left + plotW*float64(i)/float64(n-1)
	}

	chart := svg.SVG(
		svg.XMLNS,
		svg.ViewBox(0, 0, width, height),
		mx.NewAttrib("preserveAspectRatio", "xMidYMid meet"),
		html.Class("size-full overflow-visible"),
		mx.NewAttrib("aria-hidden", "true"),
		mx.NewAttrib("font-size", chartNum(fontSize)),
	)
	if !opts.HideGrid {
		grid := svg.G(html.Class("stroke-border"), mx.NewAttrib("stroke", "currentColor"), mx.NewAttrib("stroke-opacity", "0.5"))
		for _, t := range ticks {
			grid.Children = append(grid.Children, svg.Line(
				chartAttr("x1", left), chartAttr("x2", left+plotW),
				chartAttr("y1", y(t)), chartAttr("y2", y(t)),
			))
		}
		chart.Children = append(chart.Children, grid)
	}
	if !opts.HideYAxis {
		axis := svg.G(html.Class("fill-muted-foreground"), mx.NewAttrib("text-anchor", "end"))
		for _, t := range ticks {
			axis.Children = append(axis.Children, svg.Text(
				chartAttr("x", left-8), chartAttr("y", y(t)), mx.NewAttrib("dy", "0.32em"),
				format(t),
			))
		}
		chart.Children = append(chart.Children, axis)
	}
	if !opts.HideXAxis {
		axis := svg.G(html.Class("fill-muted-foreground"), mx.NewAttrib("text-anchor", "middle"))
		for i, c := range data.Categories {
			axis.Children = append(axis.Children, svg.Text(
				chartAttr("x", x(i)), chartAttr("y", top+plotH+fontSize+8),
				formatCategory(c),
			))
		}
		chart.Children = append(chart.Children, axis)
	}

	// With the styled tooltip a point carries no native one as well.
	title := func(i int, s chartSeriesInfo) mx.Component {
		if opts.Tooltip {
			return nil
		}
		return svg.Title(data.Categories[i] + ": " + s.label + " " + format(s.Values[i]))
	}
	switch kind {
	case ChartBar:
		// Recharts' defaults: 10% of the band between categories, 4px
		// between the bars of a group, and upstream's radius of 4.
		group := band * 0.8
		bars := len(series)
		barW := group
		if !stacked {
			barW = (group - 4*float64(bars-1)) / float64(bars)
		}
		for si, s := range series {
			g := svg.G(chartFill(s))
			for i := range n {
				v := s.Values[i]
				base := 0.0
				if stacked {
					for _, below := range series[:si] {
						if w := below.Values[i]; (w >= 0) == (v >= 0) {
							base += w
						}
					}
				}
				x0 := x(i) - group/2
				if !stacked {
					x0 += float64(si) * (barW + 4)
				}
				y0, y1 := y(base+v), y(base)
				if y0 > y1 {
					y0, y1 = y1, y0
				}
				rect := svg.Rect(
					chartAttr("x", x0), chartAttr("y", y0),
					chartAttr("width", max(barW, 0)), chartAttr("height", y1-y0),
					title(i, s),
				)
				// A stack rounds only the bar that ends it on its side of
				// zero: the last one of its sign with a height, so neither a
				// zero-height bar nor one below zero takes the corners.
				rounded := !stacked || v != 0
				if stacked {
					for _, above := range series[si+1:] {
						if w := above.Values[i]; w != 0 && (w >= 0) == (v >= 0) {
							rounded = false
						}
					}
				}
				if rounded {
					rect.Attribs = append(rect.Attribs, chartAttr("rx", min(4, barW/2)))
				}
				g.Children = append(g.Children, rect)
			}
			chart.Children = append(chart.Children, g)
		}
	case ChartLine, ChartArea:
		curve := opts.Curve.ValidOr(ChartCurveLinear)
		baseline := make([]float64, n)
		for _, s := range series {
			top := make([]float64, n)
			for i := range n {
				top[i] = s.Values[i]
				if stacked {
					top[i] += baseline[i]
				}
			}
			points := make([][2]float64, n)
			for i := range n {
				points[i] = [2]float64{x(i), y(top[i])}
			}
			line := chartPath(points, curve)
			if kind == ChartArea {
				// The area runs along the line and back along its base, the
				// series below for a stack, else zero.
				below := make([][2]float64, n)
				for i := range n {
					b := 0.0
					if stacked {
						b = baseline[i]
					}
					below[n-1-i] = [2]float64{x(i), y(b)}
				}
				back := chartPath(below, curve)
				area := line + " L" + strings.TrimPrefix(back, "M") + " Z"
				chart.Children = append(chart.Children, svg.Path(svg.D(area), chartFill(s), mx.NewAttrib("fill-opacity", "0.4")))
			}
			g := svg.G(chartStroke(s))
			g.Children = append(g.Children, svg.Path(svg.D(line), mx.NewAttrib("fill", "none"), mx.NewAttrib("stroke-width", "2")))
			// A dot per point, invisible without Dots, carries the tooltip.
			for i, p := range points {
				dot := svg.Circle(chartAttr("cx", p[0]), chartAttr("cy", p[1]), title(i, s))
				if opts.Dots {
					dot.Attribs = append(dot.Attribs, chartAttr("r", 4), chartFill(s))
				} else {
					dot.Attribs = append(dot.Attribs, chartAttr("r", 8), mx.NewAttrib("fill", "transparent"), mx.NewAttrib("stroke", "none"))
				}
				g.Children = append(g.Children, dot)
			}
			chart.Children = append(chart.Children, g)
			if stacked {
				copy(baseline, top)
			}
		}
	}

	if opts.Tooltip {
		// The bands over each category come last, so they take the pointer:
		// hovering one shows its tooltip with CSS alone.
		for i := range n {
			lo, hi := x(i)-band/2, x(i)+band/2
			if kind != ChartBar {
				lo, hi = left, left+plotW
				if i > 0 {
					lo = (x(i-1) + x(i)) / 2
				}
				if i < n-1 {
					hi = (x(i) + x(i+1)) / 2
				}
			}
			chart.Children = append(chart.Children, chartTooltip(data, series, i, x(i), lo, hi, top, plotH, width, kind != ChartBar, format))
		}
	}
	return chart
}

// chartCategoryColor returns the color of category i of a pie or radial
// chart.
func chartCategoryColor(data ChartData, i int) string {
	if i < len(data.CategoryColors) && data.CategoryColors[i] != "" {
		return data.CategoryColors[i]
	}
	return "var(--chart-" + strconv.Itoa(i%5+1) + ")"
}

// chartPolarSVG draws a pie, radar or radial chart in a width × height
// viewBox: a pie and a radial chart the first series, a slice or ring per
// category, a radar chart every series over a spoke per category.
func chartPolarSVG(kind ChartKind, data ChartData, series []chartSeriesInfo, opts ChartOptions, width, height float64, format func(float64) string, formatCategory func(string) string) *mx.Element {
	const fontSize = 12
	chart := svg.SVG(
		svg.XMLNS,
		svg.ViewBox(0, 0, width, height),
		mx.NewAttrib("preserveAspectRatio", "xMidYMid meet"),
		html.Class("size-full overflow-visible"),
		mx.NewAttrib("aria-hidden", "true"),
		mx.NewAttrib("font-size", chartNum(fontSize)),
	)
	cx, cy := width/2, height/2
	n := len(data.Categories)
	// The point at radius r and angle a, in degrees clockwise from the top.
	at := func(r, a float64) (float64, float64) {
		rad := (a - 90) * math.Pi / 180
		return cx + r*math.Cos(rad), cy + r*math.Sin(rad)
	}
	// arc returns the path of the arc at radius r from angle a0 to a1,
	// drawn clockwise, as a continuation of a path (an "A" command) after
	// a move to its start.
	arc := func(r, a0, a1 float64) string {
		x1, y1 := at(r, a1)
		large := 0
		if a1-a0 > 180 {
			large = 1
		}
		return fmt.Sprintf("A%s,%s 0 %d 1 %s,%s", chartNum(r), chartNum(r), large, chartNum(x1), chartNum(y1))
	}
	if len(series) == 0 || n == 0 {
		return chart
	}
	switch kind {
	case ChartPie:
		values := series[0].Values
		total := 0.0
		for _, v := range values {
			total += max(v, 0)
		}
		outer := min(width, height)/2 - 8
		inner := outer * min(max(opts.Donut, 0), 0.95)
		g := svg.G(html.Class("stroke-background"), mx.NewAttrib("stroke", "currentColor"), mx.NewAttrib("stroke-width", "1"))
		a := 0.0
		for i, v := range values {
			if v <= 0 || total == 0 {
				continue
			}
			sweep := 360 * v / total
			// A full circle cannot be one arc, whose ends would coincide: a
			// slice of the whole pie stops a hundredth of a degree short.
			a1 := a + min(sweep, 359.99)
			x0, y0 := at(outer, a)
			d := fmt.Sprintf("M%s,%s %s", chartNum(x0), chartNum(y0), arc(outer, a, a1))
			if inner > 0 {
				xi, yi := at(inner, a1)
				xs, ys := at(inner, a)
				large := 0
				if a1-a > 180 {
					large = 1
				}
				d += fmt.Sprintf(" L%s,%s A%s,%s 0 %d 0 %s,%s Z", chartNum(xi), chartNum(yi), chartNum(inner), chartNum(inner), large, chartNum(xs), chartNum(ys))
			} else {
				d += fmt.Sprintf(" L%s,%s Z", chartNum(cx), chartNum(cy))
			}
			g.Children = append(g.Children, svg.Path(svg.D(d), mx.NewAttrib("fill", chartCategoryColor(data, i)),
				svg.Title(data.Categories[i]+": "+format(v))))
			a += sweep
		}
		chart.Children = append(chart.Children, g)
	case ChartRadial:
		values := series[0].Values
		hi := 0.0
		for _, v := range values {
			hi = max(hi, v)
		}
		outer := min(width, height)/2 - 8
		inner := outer * 0.25
		band := (outer - inner) / float64(n)
		thick := band * 0.7
		for i, v := range values {
			// The first category is the outer ring, as in upstream's
			// examples.
			r := outer - band*(float64(i)+0.5)
			g := svg.G(mx.NewAttrib("fill", "none"), mx.NewAttrib("stroke-width", chartNum(thick)), mx.NewAttrib("stroke-linecap", "round"))
			g.Children = append(g.Children, svg.Circle(chartAttr("cx", cx), chartAttr("cy", cy), chartAttr("r", r),
				html.Class("stroke-muted"), mx.NewAttrib("stroke", "currentColor")))
			if v > 0 && hi > 0 {
				sweep := min(360*v/hi, 359.99)
				x0, y0 := at(r, 0)
				g.Children = append(g.Children, svg.Path(
					svg.D(fmt.Sprintf("M%s,%s %s", chartNum(x0), chartNum(y0), arc(r, 0, sweep))),
					mx.NewAttrib("stroke", chartCategoryColor(data, i)),
					svg.Title(data.Categories[i]+": "+format(v))))
			}
			chart.Children = append(chart.Children, g)
		}
	case ChartRadar:
		hi := 0.0
		for _, s := range series {
			for _, v := range s.Values {
				hi = max(hi, v)
			}
		}
		ticks := chartTicks(0, hi, cmp.Or(opts.Ticks, 4))
		if len(ticks) < 2 {
			return mx.NewErrElement(fmt.Errorf("shadcn.Chart: values up to %g have no scale", hi))
		}
		hi = ticks[len(ticks)-1]
		outer := min(width, height)/2 - fontSize*2.5
		angle := func(i int) float64 { return 360 * float64(i) / float64(n) }
		polygon := func(radius func(i int) float64) string {
			var b strings.Builder
			for i := range n {
				x, y := at(radius(i), angle(i))
				if i > 0 {
					b.WriteByte(' ')
				}
				b.WriteString(chartNum(x) + "," + chartNum(y))
			}
			return b.String()
		}
		if !opts.HideGrid {
			grid := svg.G(html.Class("stroke-border"), mx.NewAttrib("stroke", "currentColor"), mx.NewAttrib("fill", "none"))
			for _, t := range ticks[1:] {
				r := outer * t / hi
				grid.Children = append(grid.Children, svg.Polygon(mx.NewAttrib("points", polygon(func(int) float64 { return r }))))
			}
			for i := range n {
				x, y := at(outer, angle(i))
				grid.Children = append(grid.Children, svg.Line(chartAttr("x1", cx), chartAttr("y1", cy), chartAttr("x2", x), chartAttr("y2", y)))
			}
			chart.Children = append(chart.Children, grid)
		}
		if !opts.HideXAxis {
			labels := svg.G(html.Class("fill-muted-foreground"))
			for i, c := range data.Categories {
				a := angle(i)
				x, y := at(outer+fontSize, a)
				anchor := "middle"
				switch {
				case a > 10 && a < 170:
					anchor = "start"
				case a > 190 && a < 350:
					anchor = "end"
				}
				labels.Children = append(labels.Children, svg.Text(chartAttr("x", x), chartAttr("y", y),
					mx.NewAttrib("text-anchor", anchor), mx.NewAttrib("dy", "0.32em"), formatCategory(c)))
			}
			chart.Children = append(chart.Children, labels)
		}
		for _, s := range series {
			g := svg.G(chartFill(s), chartStroke(s))
			g.Children = append(g.Children, svg.Polygon(
				mx.NewAttrib("points", polygon(func(i int) float64 { return outer * max(s.Values[i], 0) / hi })),
				mx.NewAttrib("fill-opacity", "0.6"), mx.NewAttrib("stroke-width", "2")))
			for i := range n {
				x, y := at(outer*max(s.Values[i], 0)/hi, angle(i))
				dot := svg.Circle(chartAttr("cx", x), chartAttr("cy", y), svg.Title(data.Categories[i]+": "+s.label+" "+format(s.Values[i])))
				if opts.Dots {
					dot.Attribs = append(dot.Attribs, chartAttr("r", 4), mx.NewAttrib("fill-opacity", "1"))
				} else {
					dot.Attribs = append(dot.Attribs, chartAttr("r", 8), mx.NewAttrib("fill", "transparent"), mx.NewAttrib("stroke", "none"))
				}
				g.Children = append(g.Children, dot)
			}
			chart.Children = append(chart.Children, g)
		}
	}
	return chart
}

// chartTooltip renders the hover band of category i, from lo to hi across
// the plot area, with upstream's ChartTooltip card: the category, and each
// series' color, label and value. The card sits beside the category's x, on
// the side with room, and a line chart also gets a cursor line at x. It
// shows while the band is hovered (group-hover), so it needs no script.
func chartTooltip(data ChartData, series []chartSeriesInfo, i int, x, lo, hi, top, plotH, width float64, cursor bool, format func(float64) string) *mx.Element {
	const lineH, pad, fontSize = 18.0, 8.0, 12.0
	// The card's width is estimated from the longest label and value, as
	// the SVG text cannot be measured server-side.
	labelW, valueW := float64(len([]rune(data.Categories[i]))), 0.0
	for _, s := range series {
		labelW = max(labelW, float64(len([]rune(s.label)))+2)
		valueW = max(valueW, float64(len([]rune(format(s.Values[i])))))
	}
	w := (labelW+valueW+2)*fontSize*0.6 + 2*pad
	h := lineH*float64(len(series)+1) + pad
	cardX := x + 12
	if cardX+w > width {
		cardX = x - 12 - w
	}
	card := svg.G(html.Class("pointer-events-none opacity-0 transition-opacity group-hover/chart-band:opacity-100"))
	if cursor {
		card.Children = append(card.Children, svg.Line(chartAttr("x1", x), chartAttr("x2", x), chartAttr("y1", top), chartAttr("y2", top+plotH),
			html.Class("stroke-border"), mx.NewAttrib("stroke", "currentColor")))
	}
	card.Children = append(card.Children,
		svg.Rect(chartAttr("x", cardX), chartAttr("y", top), chartAttr("width", w), chartAttr("height", h), chartAttr("rx", 6),
			html.Class("fill-background stroke-border/50 drop-shadow-xl"), mx.NewAttrib("stroke", "currentColor")),
		svg.Text(chartAttr("x", cardX+pad), chartAttr("y", top+pad+fontSize*0.8), html.Class("fill-foreground font-medium"), data.Categories[i]),
	)
	for si, s := range series {
		y := top + pad + lineH*float64(si+1)
		card.Children = append(card.Children,
			svg.Rect(chartAttr("x", cardX+pad), chartAttr("y", y+2), chartAttr("width", 8), chartAttr("height", 8), chartAttr("rx", 2), chartFill(s)),
			svg.Text(chartAttr("x", cardX+pad+14), chartAttr("y", y+fontSize*0.8), html.Class("fill-muted-foreground"), s.label),
			svg.Text(chartAttr("x", cardX+w-pad), chartAttr("y", y+fontSize*0.8), mx.NewAttrib("text-anchor", "end"),
				html.Class("fill-foreground font-mono font-medium tabular-nums"), format(s.Values[i])),
		)
	}
	return svg.G(html.Class("group/chart-band"),
		svg.Rect(chartAttr("x", lo), chartAttr("y", top), chartAttr("width", hi-lo), chartAttr("height", plotH), mx.NewAttrib("fill", "transparent")),
		card,
	)
}

// chartSeriesInfo is a [ChartSeries] with its defaults resolved.
type chartSeriesInfo struct {
	ChartSeries
	key, label, color string
}

// chartMaxValue bounds the values of a [Chart], so that the sums, scales and
// coordinates computed from them stay finite.
const chartMaxValue = 1e300

// chartSeries resolves the defaults of data's series and checks that they
// can be drawn.
func chartSeries(data ChartData) ([]chartSeriesInfo, error) {
	series := make([]chartSeriesInfo, len(data.Series))
	for i, s := range data.Series {
		info := chartSeriesInfo{ChartSeries: s, key: s.Key, label: s.Label, color: s.Color}
		if info.key == "" {
			info.key = "series-" + strconv.Itoa(i+1)
		}
		if err := validateID(info.key); err != nil {
			return nil, fmt.Errorf("shadcn.Chart: series key %q: %w", info.key, err)
		}
		if info.label == "" {
			info.label = info.key
		}
		if err := validateChartColor(info.color); err != nil {
			return nil, fmt.Errorf("shadcn.Chart: series %q: %w", info.key, err)
		}
		if info.color == "" {
			info.color = "var(--chart-" + strconv.Itoa(i%5+1) + ")"
		}
		if len(s.Values) != len(data.Categories) {
			return nil, fmt.Errorf("shadcn.Chart: series %q has %d values for %d categories", info.key, len(s.Values), len(data.Categories))
		}
		for j, v := range s.Values {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return nil, fmt.Errorf("shadcn.Chart: series %q has no number for %q", info.key, data.Categories[j])
			}
			if math.Abs(v) > chartMaxValue {
				return nil, fmt.Errorf("shadcn.Chart: series %q has a value too large to draw for %q", info.key, data.Categories[j])
			}
		}
		series[i] = info
	}
	for _, c := range data.CategoryColors {
		if err := validateChartColor(c); err != nil {
			return nil, fmt.Errorf("shadcn.Chart: category color: %w", err)
		}
	}
	return series, nil
}

// chartTable renders data as the visually hidden table assistive technology
// reads instead of the chart's SVG.
func chartTable(data ChartData, series []chartSeriesInfo, caption string, format func(float64) string) *mx.Element {
	head := html.Element("tr", html.Element("td"))
	for _, s := range series {
		head.Children = append(head.Children, html.Element("th", mx.NewAttrib("scope", "col"), s.label))
	}
	body := html.Element("tbody")
	for i, c := range data.Categories {
		row := html.Element("tr", html.Element("th", mx.NewAttrib("scope", "row"), c))
		for _, s := range series {
			row.Children = append(row.Children, html.Element("td", format(s.Values[i])))
		}
		body.Children = append(body.Children, row)
	}
	table := html.Element("table", html.Class("sr-only"))
	if caption != "" {
		table.Children = append(table.Children, html.Element("caption", caption))
	}
	table.Children = append(table.Children, html.Element("thead", head), body)
	return table
}

// chartTicks returns the value ticks of an axis from lo to hi: about count
// round steps (1, 2 or 5 times a power of ten) that include both, picked as
// d3's ticks pick them. An empty range gets the ticks of 0 to 1.
func chartTicks(lo, hi float64, count int) []float64 {
	if hi <= lo {
		lo, hi = min(lo, 0), max(hi, 1)
	}
	raw := (hi - lo) / float64(max(count, 1))
	// A range too wide for a float64 or too narrow for a step it can hold
	// has no ticks.
	if math.IsInf(raw, 0) || math.IsNaN(raw) || raw < 1e-300 {
		return nil
	}
	power := math.Pow(10, math.Floor(math.Log10(raw)))
	step := power
	switch e := raw / power; {
	case e >= math.Sqrt(50):
		step *= 10
	case e >= math.Sqrt(10):
		step *= 5
	case e >= math.Sqrt(2):
		step *= 2
	}
	first := math.Floor(lo / step)
	last := math.Ceil(hi / step)
	// A tick is rounded to the decimals of the step, so 3 × 0.2 is 0.6,
	// not 0.6000000000000001.
	scale := math.Pow(10, max(0, -math.Floor(math.Log10(step))))
	ticks := make([]float64, 0, int(last-first)+1)
	for i := first; i <= last; i++ {
		ticks = append(ticks, math.Round(i*step*scale)/scale)
	}
	return ticks
}

// chartPath returns the SVG path data through points, which are ordered by x
// (or all reversed), connected as curve says.
func chartPath(points [][2]float64, curve ChartCurve) string {
	var b strings.Builder
	for i, p := range points {
		switch {
		case i == 0:
			fmt.Fprintf(&b, "M%s,%s", chartNum(p[0]), chartNum(p[1]))
		case curve == ChartCurveStep:
			mid := (points[i-1][0] + p[0]) / 2
			fmt.Fprintf(&b, " H%s V%s H%s", chartNum(mid), chartNum(p[1]), chartNum(p[0]))
		case curve == ChartCurveMonotone:
			t0, t1 := chartTangent(points, i-1), chartTangent(points, i)
			h := (p[0] - points[i-1][0]) / 3
			fmt.Fprintf(&b, " C%s,%s %s,%s %s,%s",
				chartNum(points[i-1][0]+h), chartNum(points[i-1][1]+h*t0),
				chartNum(p[0]-h), chartNum(p[1]-h*t1),
				chartNum(p[0]), chartNum(p[1]))
		default:
			fmt.Fprintf(&b, " L%s,%s", chartNum(p[0]), chartNum(p[1]))
		}
	}
	return b.String()
}

// chartTangent returns the slope of a monotone cubic curve at points[i], by
// the Fritsch–Carlson method d3's curveMonotoneX uses: zero at a local
// extremum, so the curve never overshoots a value.
func chartTangent(points [][2]float64, i int) float64 {
	slope := func(a, b [2]float64) float64 {
		if b[0] == a[0] {
			return 0
		}
		return (b[1] - a[1]) / (b[0] - a[0])
	}
	n := len(points)
	switch {
	case n < 2:
		return 0
	case i == 0:
		return slope(points[0], points[1])
	case i == n-1:
		return slope(points[n-2], points[n-1])
	}
	s0, s1 := slope(points[i-1], points[i]), slope(points[i], points[i+1])
	if s0*s1 <= 0 {
		return 0
	}
	// The harmonic mean of the neighboring slopes stays within both.
	return 2 * s0 * s1 / (s0 + s1)
}

// chartFill is the fill attribute of a series' shapes.
func chartFill(s chartSeriesInfo) mx.Attrib {
	return mx.NewAttrib("fill", "var(--color-"+s.key+")")
}

// chartStroke is the stroke attribute of a series' lines.
func chartStroke(s chartSeriesInfo) mx.Attrib {
	return mx.NewAttrib("stroke", "var(--color-"+s.key+")")
}

// chartAttr is an SVG attribute with a coordinate value.
func chartAttr(name string, v float64) mx.Attrib {
	return mx.NewAttrib(name, chartNum(v))
}

// chartNum formats a coordinate with at most two decimals, which is finer
// than a pixel at any size the viewBox is shown at.
func chartNum(v float64) string {
	return strconv.FormatFloat(math.Round(v*100)/100, 'f', -1, 64)
}

// validateChartColor rejects a color that would not stay one value of the
// style attribute it is written into: one that ends its declaration and adds
// its own, or opens a string, comment or function that swallows the
// declarations after it.
func validateChartColor(color string) error {
	depth := 0
	for _, r := range color {
		switch r {
		case '(':
			depth++
		case ')':
			depth--
		}
		if depth < 0 {
			break
		}
	}
	if depth != 0 || strings.ContainsAny(color, `;{}"'\`) || strings.Contains(color, "/*") {
		return fmt.Errorf("color %q is no single CSS value", color)
	}
	return nil
}
