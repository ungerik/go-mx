package shadcn

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestInput(t *testing.T) {
	out := render(t, Input(html.Type("email"), html.Placeholder("you@example.com")))
	for _, want := range []string{
		`data-slot="input"`,
		`type="email"`,
		`placeholder="you@example.com"`,
		"rounded-md",
		"border-input",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	if strings.Contains(out, "</input>") {
		t.Errorf("input must render as a void element: %s", out)
	}
}
