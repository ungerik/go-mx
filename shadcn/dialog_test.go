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
