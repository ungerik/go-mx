package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/svg"
)

// Spinner renders a spinning loading indicator — the lucide loader-circle
// icon (inlined like the other icons in icons.go, since go-mx has no icon
// dependency) with role="status" and aria-label="Loading" for screen
// readers. Size it with a caller class, e.g. html.Class("size-6").
func Spinner(attribs ...mx.Attrib) *mx.Element {
	e := icon("loader-circle", "", svg.Path(svg.D("M21 12a9 9 0 1 1-6.219-8.56")))
	e.Attribs = append(e.Attribs, attribs...)
	if e.AttribIndex("role") < 0 {
		e.Attribs = append(e.Attribs, svg.Role("status"))
	}
	if e.AttribIndex("aria-label") < 0 {
		e.Attribs = append(e.Attribs, svg.Attrib("aria-label", "Loading"))
	}
	return finish(e, "spinner", "size-4 animate-spin")
}
