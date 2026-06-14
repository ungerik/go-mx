package shadcn

import (
	"strings"
	"testing"
)

func TestSheetComposition(t *testing.T) {
	out := render(t, Sheet(
		SheetTrigger("s1", "Open"),
		SheetContent("s1", "",
			SheetHeader(
				SheetTitle("Edit profile"),
				SheetDescription("Make changes here."),
			),
			SheetFooter(SheetClose("Close")),
		),
	))
	for _, want := range []string{
		`data-slot="sheet"`,
		`data-slot="sheet-trigger"`,
		"document.getElementById('s1').showModal()",
		`data-slot="sheet-content"`,
		"<dialog ",
		`id="s1"`,
		"open:flex",
		"right-0", // default side
		"left-auto",
		"backdrop:bg-black/50",
		"if(event.target===this)this.close()",
		`data-slot="sheet-header"`,
		`data-slot="sheet-title"`,
		`data-slot="sheet-description"`,
		`data-slot="sheet-footer"`,
		`data-slot="sheet-close"`,
		"lucide-x",
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

func TestSheetSides(t *testing.T) {
	cases := map[SheetSide]string{
		SheetTop:    "top-0",
		SheetRight:  "right-0",
		SheetBottom: "bottom-0",
		SheetLeft:   "left-0",
	}
	for side, want := range cases {
		out := render(t, SheetContent("x", side))
		if !strings.Contains(out, want) {
			t.Errorf("side %q: missing %q in %s", side, want, out)
		}
	}
}

func TestSheetValidateIDPanics(t *testing.T) {
	for _, bad := range []string{"", "bad id"} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("expected SheetTrigger panic for id %q", bad)
				}
			}()
			SheetTrigger(bad)
		}()
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("expected SheetContent panic for id %q", bad)
				}
			}()
			SheetContent(bad, "")
		}()
	}
}
