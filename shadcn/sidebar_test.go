package shadcn

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestSidebarComposition(t *testing.T) {
	out := render(t, SidebarProvider(
		Sidebar(
			SidebarHeader(SidebarTrigger()),
			SidebarContent(
				SidebarGroup(
					SidebarGroupLabel("Platform"),
					SidebarGroupContent(
						SidebarMenu(
							SidebarMenuItem(
								SidebarMenuButton(html.DataAttr("active", "true"), "Home"),
								SidebarMenuBadge("3"),
							),
							SidebarMenuItem(
								SidebarMenuButton("More"),
								SidebarMenuSub(
									SidebarMenuSubItem(SidebarMenuSubButton(html.HRef("#"), "Sub")),
								),
							),
						),
					),
				),
			),
			SidebarSeparator(),
			SidebarFooter("footer"),
		),
		SidebarInset("main content"),
	))
	for _, want := range []string{
		`data-slot="sidebar-wrapper"`,
		"group/sidebar-wrapper",
		`data-state="expanded"`,
		"--sidebar-width:",
		"window.sidebarToggle",
		"sidebar_state",
		`data-slot="sidebar"`,
		"w-[var(--sidebar-width)]",
		"group-data-[state=collapsed]/sidebar-wrapper:w-[var(--sidebar-width-icon)]",
		`data-slot="sidebar-trigger"`,
		`onclick="sidebarToggle()"`,
		"lucide-panel-left",
		`data-slot="sidebar-header"`,
		`data-slot="sidebar-content"`,
		`data-slot="sidebar-group-label"`,
		`data-slot="sidebar-menu"`,
		"<ul ",
		`data-slot="sidebar-menu-item"`,
		"<li ",
		`data-slot="sidebar-menu-button"`,
		`data-active="true"`,
		`data-slot="sidebar-menu-badge"`,
		`data-slot="sidebar-menu-sub"`,
		`data-slot="sidebar-menu-sub-button"`,
		`data-slot="sidebar-separator"`,
		`data-slot="sidebar-inset"`,
		"<main ",
		"Platform",
		"main content",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}
