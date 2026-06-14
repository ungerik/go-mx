package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
	"github.com/ungerik/go-mx/svg"
)

// navIcon builds a small lucide-style inline SVG from its path data.
func navIcon(paths ...string) mx.Component {
	shapes := []any{
		svg.XMLNS, svg.Width(16), svg.Height(16),
		svg.ViewBox(0, 0, 24, 24), svg.Fill("none"), svg.Stroke("currentColor"),
		svg.StrokeWidth(2), svg.StrokeLineCap("round"), svg.StrokeLineJoin("round"),
	}
	for _, d := range paths {
		shapes = append(shapes, svg.Path(svg.D(d)))
	}
	return svg.SVG(shapes...)
}

// navItem builds one SidebarMenuItem with a leading icon, label, optional badge
// and active state.
func navItem(active bool, badge, label string, icon mx.Component) mx.Component {
	btn := []any{icon, html.Span(label)}
	if active {
		btn = append([]any{html.DataAttr("active", "true")}, btn...)
	}
	item := []any{shadcn.SidebarMenuButton(btn...)}
	if badge != "" {
		item = append(item, shadcn.SidebarMenuBadge(badge))
	}
	return shadcn.SidebarMenuItem(item...)
}

func SidebarDemo() mx.Component {
	return shadcn.SidebarProvider(html.Class("h-[480px] min-h-0 overflow-hidden rounded-lg border"),
		shadcn.Sidebar(
			shadcn.SidebarHeader(
				html.Div(html.Class("flex items-center gap-2 px-1 py-1.5"),
					html.Div(html.Class("bg-sidebar-primary text-sidebar-primary-foreground flex aspect-square size-8 shrink-0 items-center justify-center rounded-lg"),
						navIcon("M3 3h7v7H3z", "M14 3h7v7h-7z", "M14 14h7v7h-7z", "M3 14h7v7H3z")),
					html.Div(html.Class("grid flex-1 text-left text-sm leading-tight"),
						html.Span(html.Class("truncate font-semibold"), "Acme Inc"),
						html.Span(html.Class("text-sidebar-foreground/70 truncate text-xs"), "Enterprise"),
					),
				),
			),
			shadcn.SidebarContent(
				shadcn.SidebarGroup(
					shadcn.SidebarGroupLabel("Platform"),
					shadcn.SidebarGroupContent(
						shadcn.SidebarMenu(
							navItem(true, "", "Dashboard", navIcon("M3 3h7v7H3z", "M14 3h7v7h-7z", "M14 14h7v7h-7z", "M3 14h7v7H3z")),
							navItem(false, "3", "Inbox", navIcon("M22 12h-6l-2 3h-4l-2-3H2", "M5.5 5 2 12v6a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-6l-3.5-7z")),
							navItem(false, "", "Calendar", navIcon("M8 2v4", "M16 2v4", "M3 10h18", "M5 4h14a2 2 0 0 1 2 2v13a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2z")),
							navItem(false, "", "Search", navIcon("m21 21-4.3-4.3", "M11 4a7 7 0 1 0 0 14 7 7 0 0 0 0-14z")),
							navItem(false, "", "Settings", navIcon("M12 8a4 4 0 1 0 0 8 4 4 0 0 0 0-8z", "M19.4 13a8 8 0 0 0 0-2l2-1.5-2-3.5-2.3 1a8 8 0 0 0-1.7-1L17 3.5h-4l-.7 2.5a8 8 0 0 0-1.7 1l-2.3-1-2 3.5L8.6 11a8 8 0 0 0 0 2l-2 1.5 2 3.5 2.3-1a8 8 0 0 0 1.7 1l.7 2.5h4l.7-2.5a8 8 0 0 0 1.7-1l2.3 1 2-3.5z")),
						),
					),
				),
			),
			shadcn.SidebarFooter(
				shadcn.SidebarMenu(
					shadcn.SidebarMenuItem(
						shadcn.SidebarMenuButton(
							navIcon("M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2", "M12 7a4 4 0 1 0 0 0z"),
							html.Span("Ada Lovelace"),
						),
					),
				),
			),
		),
		shadcn.SidebarInset(
			html.Element("header", html.Class("flex h-12 shrink-0 items-center gap-2 border-b px-4"),
				shadcn.SidebarTrigger(),
				html.Span(html.Class("text-sm font-medium"), "Dashboard"),
			),
			html.Div(html.Class("text-muted-foreground flex flex-1 items-center justify-center p-6 text-sm"),
				"Toggle the sidebar with the button or ⌘B / Ctrl+B."),
		),
	)
}
