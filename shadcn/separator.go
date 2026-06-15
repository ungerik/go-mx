package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// SeparatorOrientation selects a [Separator]'s axis.
type SeparatorOrientation string // TODO use go-enum

const (
	// SeparatorHorizontal orients the separator as a horizontal rule (the default).
	SeparatorHorizontal SeparatorOrientation = "horizontal"
	// SeparatorVertical orients the separator as a vertical rule.
	SeparatorVertical SeparatorOrientation = "vertical"
)

// separatorClasses is shadcn/ui's separator class string. The data-orientation
// driven sizing utilities resolve against the data-orientation attribute that
// [Separator] always emits.
const separatorClasses = "bg-border shrink-0 data-[orientation=horizontal]:h-px data-[orientation=horizontal]:w-full data-[orientation=vertical]:h-full data-[orientation=vertical]:w-px"

// Separator renders a shadcn/ui separator as a <div role="separator">.
// orientation may be "" for the default (horizontal). A data-orientation
// attribute drives the sizing classes; a vertical separator additionally gets
// aria-orientation="vertical". A caller-supplied role, data-orientation or
// aria-orientation is left untouched.
func Separator(orientation SeparatorOrientation, attribsChildren ...any) *mx.Element {
	o := SeparatorHorizontal
	if orientation == SeparatorVertical {
		o = SeparatorVertical
	}
	e := html.Div(attribsChildren...)
	if e.AttribIndex("role") < 0 {
		e.Attribs = append(e.Attribs, html.Role("separator"))
	}
	if e.AttribIndex("data-orientation") < 0 {
		e.Attribs = append(e.Attribs, html.DataAttr("orientation", string(o)))
	}
	if o == SeparatorVertical && e.AttribIndex("aria-orientation") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("aria-orientation", "vertical"))
	}
	return finish(e, "separator", separatorClasses)
}
