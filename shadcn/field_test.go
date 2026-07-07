package shadcn

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestFieldComposition(t *testing.T) {
	out := render(t, Field("",
		FieldLabelFor("u", "Username"),
		InputID("u"),
		FieldDescription("Your public display name."),
		FieldError("Username is required."),
	))
	for _, want := range []string{
		`data-slot="field"`, `role="group"`, `data-orientation="vertical"`,
		`data-slot="field-label"`, `for="u"`,
		`data-slot="field-description"`,
		// FieldError announces to screen readers as it appears.
		`data-slot="field-error"`, `role="alert"`, "text-destructive",
		"Username is required.",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	// Class strings are reconstructed from the cn-* style sheet; a leaked
	// cn-* token would silently render unstyled.
	if strings.Contains(out, "cn-") {
		t.Errorf("unresolved cn-* token leaked into output: %s", out)
	}
	// Base UI's has-data-checked:/group-has-data-horizontal: selectors must
	// be rewritten to native/attribute forms this port actually emits.
	for _, baseUI := range []string{"has-data-checked:", "data-horizontal", "[role=checkbox]"} {
		if strings.Contains(out, baseUI) {
			t.Errorf("Base UI selector %q should have been rewritten: %s", baseUI, out)
		}
	}
}

func TestFieldOrientations(t *testing.T) {
	horizontal := render(t, Field(FieldHorizontal, "x"))
	if !strings.Contains(horizontal, `data-orientation="horizontal"`) ||
		!strings.Contains(horizontal, "flex-row") {
		t.Errorf("horizontal orientation missing: %s", horizontal)
	}
	// The checkbox/radio nudge must scope BOTH comma branches with &> to
	// direct children; an unscoped radio branch would leak mt-px to every
	// radio-group-item on the page (render escapes & to &amp;).
	if !strings.Contains(horizontal, "&amp;>[data-slot=radio-group-item]]:mt-px") {
		t.Errorf("radio branch must be direct-child scoped (&>): %s", horizontal)
	}
	responsive := render(t, Field(FieldResponsive, "x"))
	// The responsive layout must query the FieldGroup container, not the
	// viewport, so fields adapt to the space they actually get.
	if !strings.Contains(responsive, "@md/field-group:flex-row") {
		t.Errorf("responsive orientation should use container queries: %s", responsive)
	}
}

func TestFieldInvalidState(t *testing.T) {
	// Error state is author-set data-invalid on the Field root (the Form-era
	// data-error label attribute is gone with FormLabel).
	out := render(t, Field("", html.DataAttr("invalid", "true"), "x"))
	if !strings.Contains(out, `data-invalid="true"`) ||
		!strings.Contains(out, "data-[invalid=true]:text-destructive") {
		t.Errorf("invalid state classes missing: %s", out)
	}
}

func TestFieldSetLegend(t *testing.T) {
	out := render(t, FieldSet(
		FieldLegend("", "Contact"),
		FieldGroup(Field("", "x")),
	))
	for _, want := range []string{
		"<fieldset", `data-slot="field-set"`,
		"<legend", `data-slot="field-legend"`, `data-variant="legend"`,
		`data-slot="field-group"`, "@container/field-group",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	label := render(t, FieldLegend(FieldLegendLabel, "x"))
	if !strings.Contains(label, `data-variant="label"`) {
		t.Errorf("label variant missing: %s", label)
	}
}

func TestFieldLabelAsCard(t *testing.T) {
	out := render(t, FieldLabel(Field("", FieldTitle("Choice"), Checkbox())))
	// The checked-card look must key off native :checked, not Base UI's
	// data-checked attribute (our Checkbox is a native input).
	if !strings.Contains(out, "has-checked:bg-primary/5") {
		t.Errorf("missing native has-checked card classes: %s", out)
	}
	// FieldTitle deliberately shares the field-label slot (upstream parity),
	// so the label plus the title make exactly two field-label slots.
	if n := strings.Count(out, `data-slot="field-label"`); n != 2 {
		t.Errorf("expected exactly 2 field-label slots, got %d: %s", n, out)
	}
}

func TestFieldContent(t *testing.T) {
	// FieldContent stacks a label and description as one column inside a Field.
	out := render(t, FieldContent(FieldTitle("Name"), FieldDescription("Helper")))
	for _, want := range []string{
		`data-slot="field-content"`, "flex-col",
		`data-slot="field-label"`, `data-slot="field-description"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	if strings.Contains(out, "cn-") {
		t.Errorf("unresolved cn-* token leaked into output: %s", out)
	}
}

func TestFieldSeparator(t *testing.T) {
	plain := render(t, FieldSeparator())
	if !strings.Contains(plain, `data-content="false"`) ||
		strings.Contains(plain, "field-separator-content") {
		t.Errorf("plain separator should have no content span: %s", plain)
	}
	labeled := render(t, FieldSeparator("Or continue with"))
	for _, want := range []string{
		`data-content="true"`, `data-slot="field-separator-content"`,
		"Or continue with", `data-slot="separator"`,
	} {
		if !strings.Contains(labeled, want) {
			t.Errorf("missing %q in %s", want, labeled)
		}
	}
}
