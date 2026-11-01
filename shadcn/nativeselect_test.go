package shadcn

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestNativeSelect(t *testing.T) {
	out := render(t, NativeSelect("", html.ID("fruit"), html.Name("fruit"), html.Class("w-48"),
		NativeSelectOption("", "Select a fruit"),
		NativeSelectOptGroup("Fruits",
			NativeSelectOption("apple", "Apple"),
		),
	))
	for _, want := range []string{
		`data-slot="native-select-wrapper"`, `data-slot="native-select"`,
		`data-slot="native-select-icon"`, `data-slot="native-select-option"`,
		`data-slot="native-select-optgroup"`, `label="Fruits"`, `value="apple"`,
		`data-size="default"`, "appearance-none", "lucide-chevron-down",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	// As upstream does with className, the caller class sizes the wrapper,
	// while id and name must reach the <select> so a label and the form find
	// it.
	wrapper, rest, _ := strings.Cut(out, "<select")
	if !strings.Contains(wrapper, "w-48") || strings.Contains(wrapper, `id="fruit"`) {
		t.Errorf("class belongs on the wrapper, id on the select: %s", out)
	}
	sel, _, _ := strings.Cut(rest, ">")
	if !strings.Contains(sel, `id="fruit"`) || !strings.Contains(sel, `name="fruit"`) || strings.Contains(sel, "w-48") {
		t.Errorf("select start tag wrong: %s", sel)
	}
	if strings.Contains(out, "cn-") || strings.Contains(out, "base-select") {
		t.Errorf("unexpected class token: %s", out)
	}
}

func TestNativeSelectSizeSM(t *testing.T) {
	out := render(t, NativeSelect(NativeSelectSizeSM, NativeSelectOption("a", "A")))
	// Both the wrapper and the select carry the size; the select's
	// data-[size=sm]:h-8 keys on its own attribute.
	if strings.Count(out, `data-size="sm"`) != 2 || !strings.Contains(out, "data-[size=sm]:h-8") {
		t.Errorf("small size not applied: %s", out)
	}
}
