package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// shadcn/ui draws component icons from lucide-react. go-mx has no icon
// dependency, so the handful of icons the ported components need by default
// are inlined here as their lucide path data.

// icon renders an inline SVG with lucide-react's default attributes, so the
// [&_svg]:size-* utility classes carried by the button, breadcrumb and
// pagination components can size it. name becomes the lucide-<name> class;
// extraClass is merged in (the equivalent of lucide-react's className prop).
func icon(name, extraClass string, shapes ...any) *mx.Element {
	cls := "lucide lucide-" + name
	if extraClass != "" {
		cls += " " + extraClass
	}
	attribs := []any{
		html.Attrib("xmlns", "http://www.w3.org/2000/svg"),
		html.Width("24"),
		html.Height("24"),
		html.Attrib("viewBox", "0 0 24 24"),
		html.Attrib("fill", "none"),
		html.Attrib("stroke", "currentColor"),
		html.Attrib("stroke-width", "2"),
		html.Attrib("stroke-linecap", "round"),
		html.Attrib("stroke-linejoin", "round"),
		html.Class(cls),
	}
	return html.Svg(append(attribs, shapes...)...)
}

// svgPath is one <path> shape of an inline [icon].
func svgPath(d string) *mx.Element {
	return html.VoidElement("path", html.Attrib("d", d))
}

// svgCircle is one <circle> shape of an inline [icon].
func svgCircle(cx, cy, r string) *mx.Element {
	return html.VoidElement("circle", html.Attrib("cx", cx), html.Attrib("cy", cy), html.Attrib("r", r))
}

// iconChevronLeft is the lucide chevron-left icon.
func iconChevronLeft() *mx.Element {
	return icon("chevron-left", "", svgPath("m15 18-6-6 6-6"))
}

// iconChevronRight is the lucide chevron-right icon.
func iconChevronRight() *mx.Element {
	return icon("chevron-right", "", svgPath("m9 18 6-6-6-6"))
}

// iconEllipsis is the lucide ellipsis (more-horizontal) icon.
func iconEllipsis(extraClass string) *mx.Element {
	return icon("ellipsis", extraClass,
		svgCircle("12", "12", "1"),
		svgCircle("19", "12", "1"),
		svgCircle("5", "12", "1"),
	)
}

// iconMinus is the lucide minus icon.
func iconMinus() *mx.Element {
	return icon("minus", "", svgPath("M5 12h14"))
}

// iconCheck is the lucide check icon.
func iconCheck() *mx.Element {
	return icon("check", "", svgPath("M20 6 9 17l-5-5"))
}

// iconX is the lucide x (close) icon, used by the Dialog and Sheet close button.
func iconX() *mx.Element {
	return icon("x", "", svgPath("M18 6 6 18"), svgPath("m6 6 12 12"))
}

// iconPanelLeft is the lucide panel-left icon, the default SidebarTrigger glyph.
func iconPanelLeft() *mx.Element {
	return icon("panel-left", "",
		html.VoidElement("rect", html.Width("18"), html.Height("18"), html.Attrib("x", "3"), html.Attrib("y", "3"), html.Attrib("rx", "2")),
		svgPath("M9 3v18"))
}

// iconCircle is a small filled lucide-style dot, used as the
// DropdownMenuRadioItem indicator (drawn as a tiny solid circle rather
// than the lucide stroke-only circle since the indicator should be solid).
func iconCircle() *mx.Element {
	return html.Svg(
		html.Attrib("xmlns", "http://www.w3.org/2000/svg"),
		html.Width("24"), html.Height("24"),
		html.Attrib("viewBox", "0 0 24 24"),
		html.Attrib("fill", "currentColor"),
		html.Class("lucide lucide-circle"),
		html.VoidElement("circle", html.Attrib("cx", "12"), html.Attrib("cy", "12"), html.Attrib("r", "3")),
	)
}
