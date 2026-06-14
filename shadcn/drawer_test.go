package shadcn

import (
	"strings"
	"testing"
)

func TestDrawerComposition(t *testing.T) {
	out := render(t, Drawer(
		DrawerTrigger("dr1", "Open"),
		DrawerContent("dr1",
			DrawerHeader(
				DrawerTitle("Move Goal"),
				DrawerDescription("Set your daily activity goal."),
			),
			DrawerFooter(DrawerClose("Cancel")),
		),
	))
	for _, want := range []string{
		`data-slot="drawer"`,
		`data-slot="drawer-trigger"`,
		"document.getElementById('dr1').showModal()",
		`data-slot="drawer-content"`,
		"<dialog ",
		`id="dr1"`,
		"open:flex",
		"bottom-0",
		"rounded-t-lg",
		"backdrop:bg-black/50",
		"if(event.target===this)this.close()",
		`data-slot="drawer-handle"`,
		`onpointerdown="drawerStart(event,this)"`,
		"window.drawerStart",
		`data-slot="drawer-header"`,
		`data-slot="drawer-title"`,
		`data-slot="drawer-footer"`,
		`data-slot="drawer-close"`,
		"Move Goal",
		">Open<",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}

func TestDrawerValidateIDPanics(t *testing.T) {
	for _, bad := range []string{"", "bad id"} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("expected DrawerTrigger panic for id %q", bad)
				}
			}()
			DrawerTrigger(bad)
		}()
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("expected DrawerContent panic for id %q", bad)
				}
			}()
			DrawerContent(bad)
		}()
	}
}
