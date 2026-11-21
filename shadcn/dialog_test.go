package shadcn

import (
	"strings"
	"testing"
)

func TestDialogComposition(t *testing.T) {
	out := render(t, Dialog(
		DialogTrigger("d1", "Open"),
		DialogContent("d1",
			DialogHeader(
				DialogTitle("Edit profile"),
				DialogDescription("Make changes to your profile here."),
			),
			DialogFooter(
				DialogClose("Cancel"),
			),
		),
	))
	for _, want := range []string{
		`data-slot="dialog"`,
		`data-slot="dialog-trigger"`,
		"document.getElementById('d1').showModal()",
		`data-slot="dialog-content"`,
		"<dialog ",
		`id="d1"`,
		"open:grid",
		"m-auto",
		"backdrop:bg-black/50",
		"if(event.target===this)this.close()", // light dismiss
		`data-slot="dialog-header"`,
		`data-slot="dialog-title"`,
		"<h2 ",
		`data-slot="dialog-description"`,
		`data-slot="dialog-footer"`,
		`data-slot="dialog-close"`,
		"this.closest('dialog').close()",
		"lucide-x", // built-in close button icon
		"Edit profile",
		">Open<",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	if strings.Contains(out, "data-[state=") {
		t.Errorf("Radix data-[state= should have been dropped: %s", out)
	}
}

func TestDialogValidateIDPanics(t *testing.T) {
	for _, bad := range []string{"", "bad id"} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("expected DialogTrigger panic for id %q", bad)
				}
			}()
			DialogTrigger(bad)
		}()
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("expected DialogContent panic for id %q", bad)
				}
			}()
			DialogContent(bad)
		}()
	}
}

// TestModalTransitions checks that every modal <dialog> carries the CSS
// transitions that replace upstream's data-state animations, with the
// movement of its kind, and that Cn kept every class while merging.
func TestModalTransitions(t *testing.T) {
	for _, c := range []struct {
		name     string
		out      string
		movement []string
	}{
		{"Dialog", render(t, DialogContent("d")), []string{"opacity-0 scale-95 open:opacity-100 open:scale-100 starting:open:opacity-0 starting:open:scale-95", "duration-200"}},
		{"AlertDialog", render(t, AlertDialogContent("a", "")), []string{"opacity-0 scale-95 open:opacity-100 open:scale-100 starting:open:opacity-0 starting:open:scale-95", "duration-200"}},
		{"Sheet", render(t, SheetContent("s", "")), []string{"translate-x-full open:translate-x-0 starting:open:translate-x-full", "duration-300 open:duration-500"}},
		{"Drawer", render(t, DrawerContent("w")), []string{"translate-y-full open:translate-y-0 starting:open:translate-y-full", "duration-500"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			for _, want := range append(c.movement, modalTransitionClasses) {
				if !strings.Contains(c.out, want) {
					t.Errorf("missing %q in %s", want, c.out)
				}
			}
			// Radix's data-state animations have no attribute to match here.
			for _, gone := range []string{"data-[state=", "animate-in", "animate-out"} {
				if strings.Contains(c.out, gone) {
					t.Errorf("unrewritten upstream animation %q in %s", gone, c.out)
				}
			}
		})
	}
}
