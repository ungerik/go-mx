package shadcn

import (
	"strconv"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// AspectRatio wraps content in a <div> that holds a fixed width-to-height
// ratio via the CSS aspect-ratio property. ratio is width/height (pass e.g.
// 16.0/9.0); a ratio <= 0 selects the default square ratio of 1.
//
// shadcn/ui's AspectRatio is a Radix component that emulates the ratio with a
// padding-bottom hack. This port uses the native CSS aspect-ratio property
// instead. A caller-supplied style attribute is left untouched.
func AspectRatio(ratio float64, attribsChildren ...any) *mx.Element {
	if ratio <= 0 {
		ratio = 1
	}
	e := html.Div(attribsChildren...)
	if e.AttribIndex("style") < 0 {
		e.Attribs = append(e.Attribs,
			html.Style("aspect-ratio: "+strconv.FormatFloat(ratio, 'f', -1, 64)))
	}
	return finish(e, "aspect-ratio", "")
}
