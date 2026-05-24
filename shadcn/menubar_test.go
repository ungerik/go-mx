package shadcn

import (
	"strings"
	"testing"
)

func TestMenubarComposition(t *testing.T) {
	out := render(t, Menubar(
		MenubarMenu(
			MenubarTrigger("file", "File"),
			MenubarContent("file", "",
				MenubarItem("New", MenubarShortcut("⌘N")),
				MenubarItem("Open"),
				MenubarSeparator(),
				MenubarSub(
					MenubarSubTrigger("file-recent", "Open Recent"),
					MenubarSubContent("file-recent",
						MenubarItem("project-a"),
					),
				),
			),
		),
		MenubarMenu(
			MenubarTrigger("edit", "Edit"),
			MenubarContent("edit", "",
				MenubarItem("Cut"),
				MenubarItem("Paste"),
			),
		),
	))
	for _, want := range []string{
		`data-slot="menubar"`,
		`role="menubar"`,
		"window.menubarHover",
		`data-slot="menubar-menu"`,
		`data-slot="menubar-trigger"`,
		`popovertarget="file"`,
		`popovertarget="edit"`,
		`aria-haspopup="menu"`,
		`onmouseenter="menubarHover(this)"`,
		"anchor-name: --file",
		"anchor-name: --edit",
		"aria-expanded:bg-accent", // rewritten from data-[state=open]
		`data-slot="menubar-content"`,
		`id="file"`,
		`id="edit"`,
		`popover="auto"`,
		`role="menu"`,
		`data-slot="menubar-item"`,
		`data-slot="menubar-shortcut"`,
		`data-slot="menubar-separator"`,
		`data-slot="menubar-sub-trigger"`,
		`data-slot="menubar-sub-content"`,
		`data-submenu="true"`,
		">File<",
		">Edit<",
		">New<",
		">project-a<",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	if strings.Contains(out, "data-[state=open]") {
		t.Errorf("Radix data-[state=open] should have been rewritten: %s", out)
	}
}

func TestMenubarTriggerValidateIDPanics(t *testing.T) {
	for _, bad := range []string{"", "bad id"} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("expected MenubarTrigger panic for id %q", bad)
				}
			}()
			MenubarTrigger(bad)
		}()
	}
}
