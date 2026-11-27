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

// Value: protects=a label passed in attribsChildren wins over the SelectGroup
// label parameter and is written once; fails_when=SelectGroup appends the
// parameter label although the caller set one; why_new=TestSelectComposition
// passes only the parameter; seam=none; no_claim=does not cover how browsers
// render an optgroup label.
func TestSelectGroupCallerLabelWins(t *testing.T) {
	out := render(t, SelectGroup("Param", html.LabelAttr("Caller"), SelectOption("a", "A")))
	if strings.Count(out, `label="Caller"`) != 1 || strings.Contains(out, "Param") {
		t.Errorf("want exactly the caller's label: %s", out)
	}
}
