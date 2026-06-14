package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/svg"
)

// shadcn/ui draws component icons from lucide-react. go-mx has no icon
// dependency, so the handful of icons the ported components need by default are
// inlined here as their lucide path data, built with the svg package.

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
		svg.XMLNS,
		svg.Width(24),
		svg.Height(24),
		svg.ViewBox(0, 0, 24, 24),
		svg.Fill("none"),
		svg.Stroke("currentColor"),
		svg.StrokeWidth(2),
		svg.StrokeLineCap("round"),
		svg.StrokeLineJoin("round"),
		svg.Class(cls),
	}
	return svg.SVG(append(attribs, shapes...)...)
}

// iconChevronLeft is the lucide chevron-left icon.
func iconChevronLeft() *mx.Element {
	return icon("chevron-left", "", svg.Path(svg.D("m15 18-6-6 6-6")))
}

// iconChevronRight is the lucide chevron-right icon.
func iconChevronRight() *mx.Element {
	return icon("chevron-right", "", svg.Path(svg.D("m9 18 6-6-6-6")))
}

// iconEllipsis is the lucide ellipsis (more-horizontal) icon.
func iconEllipsis(extraClass string) *mx.Element {
	return icon("ellipsis", extraClass,
		svg.Circle(svg.CX(12), svg.CY(12), svg.R(1)),
		svg.Circle(svg.CX(19), svg.CY(12), svg.R(1)),
		svg.Circle(svg.CX(5), svg.CY(12), svg.R(1)),
	)
}

// iconMinus is the lucide minus icon.
func iconMinus() *mx.Element {
	return icon("minus", "", svg.Path(svg.D("M5 12h14")))
}

// iconCheck is the lucide check icon.
func iconCheck() *mx.Element {
	return icon("check", "", svg.Path(svg.D("M20 6 9 17l-5-5")))
}

// iconX is the lucide x (close) icon, used by the Dialog and Sheet close button.
func iconX() *mx.Element {
	return icon("x", "", svg.Path(svg.D("M18 6 6 18")), svg.Path(svg.D("m6 6 12 12")))
}

// iconPanelLeft is the lucide panel-left icon, the default SidebarTrigger glyph.
func iconPanelLeft() *mx.Element {
	return icon("panel-left", "",
		svg.Rect(svg.Width(18), svg.Height(18), svg.X(3), svg.Y(3), svg.RX(2)),
		svg.Path(svg.D("M9 3v18")),
	)
}

// iconCircle is a small filled lucide-style dot, used as the
// DropdownMenuRadioItem indicator (drawn as a tiny solid circle rather than the
// lucide stroke-only circle since the indicator should be solid).
func iconCircle() *mx.Element {
	return svg.SVG(
		svg.XMLNS,
		svg.Width(24),
		svg.Height(24),
		svg.ViewBox(0, 0, 24, 24),
		svg.Fill("currentColor"),
		svg.Class("lucide lucide-circle"),
		svg.Circle(svg.CX(12), svg.CY(12), svg.R(3)),
	)
}
