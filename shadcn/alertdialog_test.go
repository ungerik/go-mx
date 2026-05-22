package shadcn

import (
	"strings"
	"testing"
)

func TestAlertDialogContent(t *testing.T) {
	out := render(t, AlertDialogContent("confirm-delete", AlertDialogHeader()))
	for _, want := range []string{
		"<dialog ",
		`id="confirm-delete"`,
		`data-slot="alert-dialog-content"`,
		`data-size="default"`,
		"backdrop:bg-black/50",
		"<form ",
		`method="dialog"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	// Radix-only positioning / animation classes must be gone.
	for _, gone := range []string{"data-[state=", "animate-", "z-50", "top-[50%]", "translate-x"} {
		if strings.Contains(out, gone) {
			t.Errorf("Radix-only class %q should have been dropped: %s", gone, out)
		}
	}
}

func TestAlertDialogTrigger(t *testing.T) {
	out := render(t, AlertDialogTrigger("confirm-delete", "Delete"))
	for _, want := range []string{
		`data-slot="alert-dialog-trigger"`,
		`type="button"`,
		"document.getElementById('confirm-delete').showModal()",
		">Delete<",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}

func TestAlertDialogActionAndCancel(t *testing.T) {
	action := render(t, AlertDialogAction("Continue"))
	for _, want := range []string{
		`data-slot="alert-dialog-action"`,
		`type="submit"`,
		`value="confirm"`,
		"bg-primary", // default Button variant
	} {
		if !strings.Contains(action, want) {
			t.Errorf("action: missing %q in %s", want, action)
		}
	}
	cancel := render(t, AlertDialogCancel("Cancel"))
	for _, want := range []string{
		`data-slot="alert-dialog-cancel"`,
		`type="submit"`,
		`value="cancel"`,
		"border", // outline Button variant
	} {
		if !strings.Contains(cancel, want) {
			t.Errorf("cancel: missing %q in %s", want, cancel)
		}
	}
}

// TestAlertDialogComposition renders a full trigger + dialog tree; render
// fails the test on any writer error (e.g. a duplicate attribute).
func TestAlertDialogComposition(t *testing.T) {
	out := render(t, AlertDialog(
		AlertDialogTrigger("d1", "Open"),
		AlertDialogContent("d1",
			AlertDialogHeader(
				AlertDialogTitle("Are you sure?"),
				AlertDialogDescription("This action cannot be undone."),
			),
			AlertDialogFooter(
				AlertDialogCancel("Cancel"),
				AlertDialogAction("Continue"),
			),
		),
	))
	for _, want := range []string{
		`data-slot="alert-dialog"`,
		`data-slot="alert-dialog-trigger"`,
		`data-slot="alert-dialog-content"`,
		`data-slot="alert-dialog-title"`,
		`data-slot="alert-dialog-description"`,
		`data-slot="alert-dialog-action"`,
		`data-slot="alert-dialog-cancel"`,
		"<dialog ",
		"<h2 ",
		"Are you sure?",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in composition output", want)
		}
	}
}

func TestValidateIDPanics(t *testing.T) {
	for _, bad := range []string{"", "bad id", "has.dot", "quote'd"} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("expected panic for id %q", bad)
				}
			}()
			validateID(bad)
		}()
	}
	// A valid id must not panic.
	validateID("a-valid_ID-123")
}
