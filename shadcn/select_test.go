package shadcn

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestSelectComposition(t *testing.T) {
	out := render(t, Select(html.Name("country"),
		SelectGroup("Europe",
			SelectOption("de", "Germany"),
			SelectOption("fr", "France"),
		),
		SelectGroup("Americas",
			SelectOption("us", "United States", html.Selected),
		),
	))
	for _, want := range []string{
		`<select `,
		`data-slot="select"`,
		`name="country"`,
		"[appearance:base-select]",
		`<optgroup `,
		`data-slot="select-group"`,
		`label="Europe"`,
		`label="Americas"`,
		`<option `,
		`data-slot="select-option"`,
		`value="de"`,
		`value="fr"`,
		`value="us"`,
		"selected",
		">Germany<",
		">France<",
		">United States<",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}

func TestSelectCallerAttrsPassThrough(t *testing.T) {
	out := render(t, Select(
		html.Name("size"), html.Required, html.Disabled,
		SelectOption("s", "Small"),
		SelectOption("m", "Medium"),
	))
	for _, want := range []string{`name="size"`, "required", "disabled"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}
