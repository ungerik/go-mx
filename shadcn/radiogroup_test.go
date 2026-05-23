package shadcn

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestRadioGroupRoot(t *testing.T) {
	out := render(t, RadioGroup("plan",
		RadioGroupItem("plan", "free", html.ID("plan-free")),
		RadioGroupItem("plan", "pro", html.ID("plan-pro"), html.Checked),
	))
	for _, want := range []string{
		`data-slot="radio-group"`,
		`role="radiogroup"`,
		`data-name="plan"`,
		"grid gap-3",
		`data-slot="radio-group-item"`,
		`type="radio"`,
		`name="plan"`,
		`value="free"`,
		`value="pro"`,
		"checked:border-primary",
		"checked:before:scale-100",
		"checked",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}

func TestRadioGroupItemAttrsPassThrough(t *testing.T) {
	out := render(t, RadioGroupItem("g", "a", html.ID("g-a"), html.Disabled))
	for _, want := range []string{`id="g-a"`, "disabled"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}

func TestRadioGroupValidateIDPanics(t *testing.T) {
	for _, bad := range []string{"", "bad name", "dot.name"} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("expected panic for name %q", bad)
				}
			}()
			RadioGroup(bad)
		}()
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("expected RadioGroupItem panic for name %q", bad)
				}
			}()
			RadioGroupItem(bad, "v")
		}()
	}
}
