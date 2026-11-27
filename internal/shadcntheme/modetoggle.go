package shadcntheme

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
	"github.com/ungerik/go-mx/svg"
)

// ModeToggle is the light/dark switch of a demo page, shadcn/ui's mode
// toggle: a ghost icon button showing a sun in light mode and a moon in dark
// mode, which opens a [shadcn.DropdownMenu] to choose Light, Dark or System.
// The choice goes through the mxSetTheme function [InitScript] defines, so
// the page needs [Head]. menuID names the menu and must be unique in the
// page.
func ModeToggle(menuID string) mx.Component {
	item := func(choice, label string, icon *mx.Element) mx.Component {
		return shadcn.DropdownMenuItem(
			html.OnClick("mxSetTheme('"+choice+"');this.closest('[popover]').hidePopover()"),
			icon, label,
		)
	}
	return shadcn.DropdownMenu(
		shadcn.DropdownMenuTrigger(menuID,
			html.Class(shadcn.ButtonClasses(shadcn.ButtonGhost, shadcn.SizeIcon)),
			html.Attrib("aria-label", "Theme"),
			html.Title("Theme"),
			lucideIcon("sun", "dark:hidden", sunShapes()...),
			lucideIcon("moon", "hidden dark:block", svg.Path(svg.D("M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z"))),
		),
		// Aligned to the trigger's right edge, so the menu stays on screen
		// when the toggle sits at the right of a narrow header; important,
		// because the port's inline style centers it.
		shadcn.DropdownMenuContent(menuID, "", html.Class("min-w-32 [position-area:bottom_span-left]!"),
			item("light", "Light", lucideIcon("sun", "", sunShapes()...)),
			item("dark", "Dark", lucideIcon("moon", "", svg.Path(svg.D("M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z")))),
			item("system", "System", lucideIcon("monitor", "",
				svg.Rect(svg.Width(20), svg.Height(14), svg.X(2), svg.Y(3), svg.RX(2)),
				svg.Line(svg.X1(8), svg.X2(16), svg.Y1(21), svg.Y2(21)),
				svg.Line(svg.X1(12), svg.X2(12), svg.Y1(17), svg.Y2(21)),
			)),
		),
	)
}

// sunShapes are the shapes of the lucide sun icon.
func sunShapes() []any {
	return []any{
		svg.Circle(svg.CX(12), svg.CY(12), svg.R(4)),
		svg.Path(svg.D("M12 2v2")), svg.Path(svg.D("M12 20v2")),
		svg.Path(svg.D("m4.93 4.93 1.41 1.41")), svg.Path(svg.D("m17.66 17.66 1.41 1.41")),
		svg.Path(svg.D("M2 12h2")), svg.Path(svg.D("M20 12h2")),
		svg.Path(svg.D("m6.34 17.66-1.41 1.41")), svg.Path(svg.D("m19.07 4.93-1.41 1.41")),
	}
}

// lucideIcon renders an inline lucide icon (https://lucide.dev, the icon set
// shadcn/ui uses) with lucide's default attributes, sized by the classes of
// the button or menu item it is in; extraClass is added to its class.
func lucideIcon(name, extraClass string, shapes ...any) *mx.Element {
	class := "lucide lucide-" + name
	if extraClass != "" {
		class += " " + extraClass
	}
	return svg.SVG(append([]any{
		svg.XMLNS,
		svg.Width(24),
		svg.Height(24),
		svg.ViewBox(0, 0, 24, 24),
		svg.Fill("none"),
		svg.Stroke("currentColor"),
		svg.StrokeWidth(2),
		svg.StrokeLineCapRound,
		svg.StrokeLineJoinRound,
		svg.Class(class),
		html.Attrib("aria-hidden", "true"),
	}, shapes...)...)
}
