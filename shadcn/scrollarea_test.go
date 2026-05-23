package shadcn

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestScrollAreaDefault(t *testing.T) {
	out := render(t, ScrollArea(html.Class("h-72 w-48"), "Long content here"))
	for _, want := range []string{
		`data-slot="scroll-area"`,
		"relative",
		"overflow-auto",
		"[scrollbar-width:thin]",
		"h-72 w-48",
		">Long content here<",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}
