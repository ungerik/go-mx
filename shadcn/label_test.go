package shadcn

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestLabel(t *testing.T) {
	out := render(t, Label(html.For("email"), "Email"))
	for _, want := range []string{
		`data-slot="label"`,
		`for="email"`,
		">Email<",
		"font-medium",
		"select-none",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}
