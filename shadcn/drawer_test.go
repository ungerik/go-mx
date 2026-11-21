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

// TestDrawerDragInChrome drives the drag script in headless Chrome: a drag
// past the threshold closes the drawer, a short one snaps it back, and both
// clear the inline translate and transition the drag set as soon as the
// pointer is released, so the CSS open/close transitions take over at once
// and no timer is left to interfere with the next gesture.
func TestDrawerDragInChrome(t *testing.T) {
	body := DrawerContent("drw", DrawerTitle("Drawer"))
	const script = /*js*/ `
var d=document.getElementById('drw'),h=d.querySelector('[data-slot=drawer-handle]');
var drag=async function(to){
	d.showModal();
	h.dispatchEvent(new PointerEvent('pointerdown',{bubbles:true,clientY:0}));
	document.dispatchEvent(new PointerEvent('pointermove',{bubbles:true,clientY:to}));
	var during=d.style.translate+'|'+d.style.transition;
	document.dispatchEvent(new PointerEvent('pointerup',{bubbles:true,clientY:to}));
	return {during:during,open:d.open,inline:d.style.translate+'|'+d.style.transition};
};
var far=await drag(5000);
out.farDuring=far.during;
out.farOpen=far.open;
out.farInline=far.inline;
var near=await drag(5);
out.nearOpen=near.open;
out.nearInline=near.inline;
`
	got := runInChrome(t, body, script, nil)
	checkRecorded(t, got, map[string]any{
		"farDuring":  "0px 5000px|none",
		"farOpen":    false,
		"farInline":  "|",
		"nearOpen":   true,
		"nearInline": "|",
	})
}
