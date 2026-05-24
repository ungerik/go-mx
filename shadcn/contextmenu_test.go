package shadcn

import (
	"strings"
	"testing"
)

func TestContextMenuComposition(t *testing.T) {
	out := render(t, ContextMenu(
		ContextMenuTrigger("cm1", "Right-click me"),
		ContextMenuContent("cm1",
			ContextMenuLabel("Actions"),
			ContextMenuSeparator(),
			ContextMenuItem("Copy", ContextMenuShortcut("⌘C")),
			ContextMenuItem("Paste"),
			ContextMenuCheckboxItem(true, "Show grid"),
			ContextMenuRadioGroup("zoom",
				ContextMenuRadioItem("zoom", "100", true, "100%"),
				ContextMenuRadioItem("zoom", "150", false, "150%"),
			),
			ContextMenuSub(
				ContextMenuSubTrigger("cm1-more", "More"),
				ContextMenuSubContent("cm1-more",
					ContextMenuItem("Sub action"),
				),
			),
		),
	))
	for _, want := range []string{
		`data-slot="context-menu"`,
		`data-slot="context-menu-trigger"`,
		`oncontextmenu="contextMenuOpen(event,'cm1')"`,
		"window.contextMenuOpen",
		`data-slot="context-menu-content"`,
		`id="cm1"`,
		`popover="auto"`,
		`role="menu"`,
		"window.menuKeyNav",
		// Items
		`data-slot="context-menu-item"`,
		`role="menuitem"`,
		`data-slot="context-menu-label"`,
		`data-slot="context-menu-separator"`,
		`data-slot="context-menu-shortcut"`,
		`data-slot="context-menu-checkbox-item"`,
		`role="menuitemcheckbox"`,
		`aria-checked="true"`,
		`data-slot="context-menu-radio-group"`,
		`data-slot="context-menu-radio-item"`,
		`role="menuitemradio"`,
		// Sub
		`data-slot="context-menu-sub"`,
		`data-slot="context-menu-sub-trigger"`,
		`popovertarget="cm1-more"`,
		`data-slot="context-menu-sub-content"`,
		`data-submenu="true"`,
		">Right-click me<",
		">Copy<",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	// Critical: ContextMenuContent must NOT carry position-anchor — the
	// script sets pixel position from the cursor at click time.
	if strings.Contains(out, `data-slot="context-menu-content"`) {
		// Find the content element's style attribute
		idx := strings.Index(out, `data-slot="context-menu-content"`)
		// Look for "style=" in the same element (before the next ">").
		segment := out[idx : idx+500]
		if end := strings.Index(segment, ">"); end > 0 {
			segment = segment[:end]
		}
		if strings.Contains(segment, "position-anchor") {
			t.Errorf("ContextMenuContent should NOT carry position-anchor: %s", segment)
		}
	}
}

func TestContextMenuTriggerValidateIDPanics(t *testing.T) {
	for _, bad := range []string{"", "bad id"} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("expected ContextMenuTrigger panic for id %q", bad)
				}
			}()
			ContextMenuTrigger(bad)
		}()
	}
}
