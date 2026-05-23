package shadcn

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/hx"
)

func TestToggleDefault(t *testing.T) {
	out := render(t, Toggle("", "", "Bold"))
	for _, want := range []string{
		`data-slot="toggle"`,
		`type="button"`,
		`aria-pressed="false"`,
		`data-variant="default"`,
		`data-size="default"`,
		"aria-pressed:bg-accent", // rewritten from Radix data-[state=on]
		"h-9",                    // default size
		toggleAriaPressedFlip,    // default onclick body
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	if strings.Contains(out, "data-[state=on]") {
		t.Errorf("Radix data-[state=on] should have been rewritten: %s", out)
	}
}

func TestToggleVariantsAndSizes(t *testing.T) {
	out := render(t, Toggle(ToggleOutline, ToggleSizeLG, "x"))
	for _, want := range []string{
		`data-variant="outline"`,
		`data-size="lg"`,
		"border",
		"h-10",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}

func TestToggleCallerOnClickOverridesDefault(t *testing.T) {
	out := render(t, Toggle("", "", html.OnClick("doit()"), "x"))
	if !strings.Contains(out, `onclick="doit()"`) {
		t.Errorf("caller onclick should be present: %s", out)
	}
	if strings.Contains(out, toggleAriaPressedFlip) {
		t.Errorf("default aria-pressed flip should be skipped when caller onclick is set: %s", out)
	}
}

func TestToggleHXOptOut(t *testing.T) {
	out := render(t, Toggle("", "", hx.Post("/toggle-bold"), hx.Swap("outerHTML"), "Bold"))
	for _, want := range []string{`hx-post="/toggle-bold"`, `hx-swap="outerHTML"`, `aria-pressed="false"`} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	if strings.Contains(out, "onclick=") {
		t.Errorf("default onclick should be skipped when an hx-* attribute is present: %s", out)
	}
}

func TestToggleClassesReturnsString(t *testing.T) {
	cls := ToggleClasses(ToggleOutline, ToggleSizeSM)
	for _, want := range []string{"border", "h-8"} {
		if !strings.Contains(cls, want) {
			t.Errorf("ToggleClasses missing %q: %s", want, cls)
		}
	}
}

func TestToggleUnknownVariantFallsBack(t *testing.T) {
	out := render(t, Toggle(ToggleVariant("bogus"), ToggleSize("nope"), "x"))
	if !strings.Contains(out, `data-variant="default"`) || !strings.Contains(out, `data-size="default"`) {
		t.Errorf("unknown variant/size should fall back to default: %s", out)
	}
}
