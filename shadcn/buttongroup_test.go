package shadcn

import (
	"strings"
	"testing"
)

func TestButtonGroup(t *testing.T) {
	out := render(t, ButtonGroup("",
		Button(ButtonOutline, "", "One"),
		Button(ButtonOutline, "", "Two"),
	))
	for _, want := range []string{
		`data-slot="button-group"`, `role="group"`, `data-orientation="horizontal"`,
		// The joining classes key off the children's data-slot attributes.
		"rounded-r-none", "rounded-l-none",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	if strings.Contains(out, "cn-") {
		t.Errorf("unresolved cn-* token leaked into output: %s", out)
	}
	// The select width rule must target this port's native <select> slot, not
	// Radix's select-trigger, or it silently never matches (the README says
	// ButtonGroup works with Select).
	// (render escapes & to &amp; in the class attribute).
	if !strings.Contains(out, "[&amp;>[data-slot=select]:not([class*='w-'])]:w-fit") {
		t.Errorf("select w-fit rule must target data-slot=select: %s", out)
	}
	if strings.Contains(out, "select-trigger") {
		t.Errorf("select-trigger is a Radix slot this port never emits: %s", out)
	}
}

func TestButtonGroupVertical(t *testing.T) {
	out := render(t, ButtonGroup(ButtonGroupVertical, Button("", "", "x")))
	for _, want := range []string{`data-orientation="vertical"`, "flex-col", "rounded-b-none"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}

func TestButtonGroupClasses(t *testing.T) {
	if !strings.Contains(ButtonGroupClasses(""), "rounded-r-none") {
		t.Error("empty orientation should resolve to horizontal classes")
	}
	if !strings.Contains(ButtonGroupClasses(ButtonGroupVertical), "flex-col") {
		t.Error("vertical orientation should include flex-col")
	}
}

func TestButtonGroupText(t *testing.T) {
	out := render(t, ButtonGroupText("https://"))
	if !strings.Contains(out, `data-slot="button-group-text"`) || !strings.Contains(out, "bg-muted") {
		t.Errorf("missing slot or classes: %s", out)
	}
}

func TestButtonGroupSeparator(t *testing.T) {
	out := render(t, ButtonGroupSeparator(""))
	// Default is vertical (a horizontal group needs a vertical rule) —
	// the opposite of Separator's default.
	for _, want := range []string{
		`data-slot="button-group-separator"`, `data-orientation="vertical"`,
		`aria-orientation="vertical"`, "bg-input", "self-stretch",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	// Base UI's bare data-vertical:/data-horizontal: selectors must be
	// rewritten to the data-[orientation=…]: attribute this port emits —
	// and the h-auto rewrite must win the merge against Separator's h-full,
	// or a vertical separator would overflow the group.
	if strings.Contains(out, "data-vertical:") || strings.Contains(out, "data-horizontal:") {
		t.Errorf("Base UI bare data-orientation selector should have been rewritten: %s", out)
	}
	if strings.Contains(out, "data-[orientation=vertical]:h-full") {
		t.Errorf("h-auto should have replaced the separator h-full: %s", out)
	}
	if !strings.Contains(out, "data-[orientation=vertical]:h-auto") {
		t.Errorf("missing h-auto rewrite: %s", out)
	}
}

func TestButtonGroupSeparatorHorizontal(t *testing.T) {
	out := render(t, ButtonGroupSeparator(SeparatorHorizontal))
	if !strings.Contains(out, `data-orientation="horizontal"`) {
		t.Errorf("horizontal separator should emit data-orientation=horizontal: %s", out)
	}
	// aria-orientation is only added for the vertical default (a horizontal
	// rule is the default reading direction, so it stays implicit).
	if strings.Contains(out, "aria-orientation=") {
		t.Errorf("horizontal separator should not set aria-orientation: %s", out)
	}
}
