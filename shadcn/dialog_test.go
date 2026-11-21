package shadcn

import (
	"slices"
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
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

// TestModalTransitionsCallerOverride checks that a caller's class replaces
// the modal defaults through Cn, the way a className overrides upstream's.
//
// Value: protects=a caller's transition-none or backdrop:bg-* replacing the modal transition and backdrop color in every browser; fails_when=the transition list moves into supports-prefixed classes a caller's unprefixed class no longer conflicts with, or the backdrop color stops being one backdrop:bg-* class; why_new=no test passed a caller class to a modal; seam=none; no_claim=the rendered animation itself (no Tailwind build in the test)
func TestModalTransitionsCallerOverride(t *testing.T) {
	out := render(t, DialogContent("d", html.Class("transition-none backdrop:bg-black/80")))
	classes := strings.Fields(classAttr(t, out))
	for _, want := range []string{"transition-none", "backdrop:bg-black/80"} {
		if !slices.Contains(classes, want) {
			t.Errorf("the caller's %q is missing: %v", want, classes)
		}
	}
	// The default list must be gone under any variant prefix too: a
	// supports-[…]: copy would not conflict with the caller's unprefixed
	// class and would bring the transition back where the prefix matches.
	for _, c := range classes {
		if strings.HasSuffix(c, "transition-[opacity,scale,translate,display,overlay]") || c == "backdrop:bg-black/50" {
			t.Errorf("the default %q should give way to the caller's class: %v", c, classes)
		}
	}
}

// classAttr returns the value of the first class attribute in out.
func classAttr(t *testing.T, out string) string {
	t.Helper()
	_, rest, ok := strings.Cut(out, `class="`)
	if !ok {
		t.Fatalf("no class attribute in %s", out)
	}
	value, _, _ := strings.Cut(rest, `"`)
	return value
}
