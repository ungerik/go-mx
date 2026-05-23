package shadcn

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestSwitchDefault(t *testing.T) {
	out := render(t, Switch())
	for _, want := range []string{
		`<input `,
		`data-slot="switch"`,
		`type="checkbox"`,
		`role="switch"`,
		"peer",            // base
		"checked:bg-primary",
		"before:content-['']",
		"checked:before:translate-x-[calc(100%-2px)]",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	// Radix-only data-state selectors must be gone.
	if strings.Contains(out, "data-[state=") {
		t.Errorf("Radix data-[state= selector should have been rewritten: %s", out)
	}
}

func TestSwitchCallerAttrsPassThrough(t *testing.T) {
	out := render(t, Switch(html.Name("notify"), html.Value("on"), html.Checked, html.Disabled))
	for _, want := range []string{`name="notify"`, `value="on"`, "checked", "disabled"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}

func TestSwitchCallerTypeOverride(t *testing.T) {
	// A caller-supplied type wins over the default; the input still renders.
	out := render(t, Switch(html.Type("checkbox"), html.Class("w-12")))
	if !strings.Contains(out, `type="checkbox"`) || !strings.Contains(out, "w-12") {
		t.Errorf("caller type/class should pass through: %s", out)
	}
}
