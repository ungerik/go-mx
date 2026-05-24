package shadcn

import (
	"strings"
	"testing"
)

func TestDropdownMenuComposition(t *testing.T) {
	out := render(t, DropdownMenu(
		DropdownMenuTrigger("dm1", "Menu"),
		DropdownMenuContent("dm1", "",
			DropdownMenuLabel("Account"),
			DropdownMenuSeparator(),
			DropdownMenuGroup(
				DropdownMenuItem("Profile", DropdownMenuShortcut("⌘P")),
				DropdownMenuItem("Settings"),
			),
			DropdownMenuSeparator(),
			DropdownMenuCheckboxItem(true, "Notifications"),
			DropdownMenuRadioGroup("theme",
				DropdownMenuRadioItem("theme", "light", true, "Light"),
				DropdownMenuRadioItem("theme", "dark", false, "Dark"),
			),
			DropdownMenuSub(
				DropdownMenuSubTrigger("dm1-more", "More"),
				DropdownMenuSubContent("dm1-more",
					DropdownMenuItem("Sub item"),
				),
			),
		),
	))
	for _, want := range []string{
		`data-slot="dropdown-menu"`,
		`data-slot="dropdown-menu-trigger"`,
		`popovertarget="dm1"`,
		`aria-haspopup="menu"`,
		`aria-expanded="false"`,
		"anchor-name: --dm1",
		`data-slot="dropdown-menu-content"`,
		`id="dm1"`,
		`popover="auto"`,
		`role="menu"`,
		`onkeydown="menuKeyNav(event)"`,
		`ontoggle="menuOpen(event)"`,
		"position-area: bottom",
		"window.menuKeyNav",
		"window.menuOpen",
		// Items
		`data-slot="dropdown-menu-item"`,
		`role="menuitem"`,
		`tabindex="-1"`,
		`data-slot="dropdown-menu-label"`,
		`data-slot="dropdown-menu-separator"`,
		`role="separator"`,
		`data-slot="dropdown-menu-group"`,
		`role="group"`,
		`data-slot="dropdown-menu-shortcut"`,
		// Checkbox item
		`data-slot="dropdown-menu-checkbox-item"`,
		`role="menuitemcheckbox"`,
		`aria-checked="true"`,
		"lucide-check",
		// Radio group / items
		`data-slot="dropdown-menu-radio-group"`,
		`data-radio-group="theme"`,
		`data-slot="dropdown-menu-radio-item"`,
		`role="menuitemradio"`,
		`data-value="light"`,
		`data-value="dark"`,
		"lucide-circle",
		// Sub
		`data-slot="dropdown-menu-sub"`,
		`data-slot="dropdown-menu-sub-trigger"`,
		`popovertarget="dm1-more"`,
		"lucide-chevron-right",
		`data-slot="dropdown-menu-sub-content"`,
		`data-submenu="true"`,
		"position-area: right",
		">Profile<",
		">Sub item<",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	if strings.Contains(out, "data-[state=") {
		t.Errorf("Radix data-[state= should have been dropped: %s", out)
	}
}

func TestDropdownMenuRadioItemUnselected(t *testing.T) {
	out := render(t, DropdownMenuRadioItem("g", "v", false, "Off"))
	if !strings.Contains(out, `aria-checked="false"`) {
		t.Errorf("expected aria-checked=false: %s", out)
	}
	if strings.Contains(out, "lucide-circle") {
		t.Errorf("unselected radio should NOT render the circle indicator: %s", out)
	}
}

func TestDropdownMenuCheckboxItemUnchecked(t *testing.T) {
	out := render(t, DropdownMenuCheckboxItem(false, "Off"))
	if !strings.Contains(out, `aria-checked="false"`) {
		t.Errorf("expected aria-checked=false: %s", out)
	}
	if strings.Contains(out, "lucide-check") {
		t.Errorf("unchecked item should NOT render the check indicator: %s", out)
	}
}

func TestDropdownMenuValidateIDPanics(t *testing.T) {
	for _, bad := range []string{"", "bad id"} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("expected DropdownMenuTrigger panic for id %q", bad)
				}
			}()
			DropdownMenuTrigger(bad)
		}()
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("expected DropdownMenuContent panic for id %q", bad)
				}
			}()
			DropdownMenuContent(bad, "")
		}()
	}
}
