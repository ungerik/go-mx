package shadcn

import (
	"strconv"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// Progress renders a shadcn/ui progress bar as a track <div role="progressbar">
// containing an indicator <div> sized via a CSS transform. value is the percent
// filled, clamped to [0, 100]; pass 0 for an empty bar.
//
// shadcn/ui's Progress is Radix-driven: the indicator's translateX is updated
// from the React value prop. The value is known at render time in Go, so this
// port emits the same transform inline. role, aria-valuemin/max/now default
// here and are overridable.
func Progress(value float64, attribsChildren ...any) *mx.Element {
	if value < 0 {
		value = 0
	}
	if value > 100 {
		value = 100
	}
	valueStr := strconv.FormatFloat(value, 'f', -1, 64)
	offset := strconv.FormatFloat(100-value, 'f', -1, 64)

	e := html.Div(attribsChildren...)
	if e.AttribIndex("role") < 0 {
		e.Attribs = append(e.Attribs, html.Role("progressbar"))
	}
	if e.AttribIndex("aria-valuemin") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("aria-valuemin", "0"))
	}
	if e.AttribIndex("aria-valuemax") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("aria-valuemax", "100"))
	}
	if e.AttribIndex("aria-valuenow") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("aria-valuenow", valueStr))
	}

	indicator := finish(
		html.Div(html.Style("transform: translateX(-"+offset+"%)")),
		"progress-indicator",
		"bg-primary h-full w-full flex-1 transition-all",
	)
	e.Children = append(e.Children, indicator)

	return finish(e, "progress", "bg-primary/20 relative h-2 w-full overflow-hidden rounded-full")
}
