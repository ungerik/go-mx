package shadcn

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestButtonDefault(t *testing.T) {
	out := render(t, Button("", "", "Click"))
	for _, want := range []string{
		`data-slot="button"`,
		`data-variant="default"`,
		`data-size="default"`,
		`type="button"`,
		">Click<",
		"inline-flex", // base classes
		"bg-primary",  // default variant
		"h-9",         // default size
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}

func TestButtonVariantsAndSizes(t *testing.T) {
	variants := map[ButtonVariant]string{
		ButtonDestructive: "bg-destructive",
		ButtonOutline:     "border",
		ButtonSecondary:   "bg-secondary",
		ButtonGhost:       "hover:bg-accent",
		ButtonLink:        "underline-offset-4",
	}
	for v, cls := range variants {
		out := render(t, Button(v, "", "x"))
		if !strings.Contains(out, `data-variant="`+string(v)+`"`) {
			t.Errorf("variant %s: missing data-variant: %s", v, out)
		}
		if !strings.Contains(out, cls) {
			t.Errorf("variant %s: missing class %q: %s", v, cls, out)
		}
	}
	sizes := map[ButtonSize]string{
		SizeSM:     "h-8",
		SizeLG:     "h-10",
		SizeIcon:   "size-9",
		SizeIconSM: "size-8",
	}
	for s, cls := range sizes {
		out := render(t, Button("", s, "x"))
		if !strings.Contains(out, `data-size="`+string(s)+`"`) {
			t.Errorf("size %s: missing data-size: %s", s, out)
		}
		if !strings.Contains(out, cls) {
			t.Errorf("size %s: missing class %q: %s", s, cls, out)
		}
	}
}

func TestButtonCallerTypeOverride(t *testing.T) {
	out := render(t, Button("", "", html.Type("submit"), "x"))
	if !strings.Contains(out, `type="submit"`) || strings.Contains(out, `type="button"`) {
		t.Errorf("caller type should override the default: %s", out)
	}
}

func TestButtonCallerClassMerges(t *testing.T) {
	out := render(t, Button(ButtonOutline, SizeSM, html.Class("w-full")))
	for _, want := range []string{"w-full", "border", "h-8"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}

func TestButtonUnknownVariantFallsBack(t *testing.T) {
	out := render(t, Button(ButtonVariant("bogus"), ButtonSize("nope"), "x"))
	if !strings.Contains(out, `data-variant="bogus"`) {
		t.Errorf("unknown variant should still be echoed verbatim: %s", out)
	}
	if !strings.Contains(out, "bg-primary") || !strings.Contains(out, "h-9") {
		t.Errorf("unknown variant/size should fall back to default classes: %s", out)
	}
}
