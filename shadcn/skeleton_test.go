package shadcn

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestSkeleton(t *testing.T) {
	out := render(t, Skeleton(html.Class("h-4 w-32")))
	for _, want := range []string{
		`data-slot="skeleton"`,
		"animate-pulse",
		"bg-accent",
		"rounded-md",
		"h-4",
		"w-32",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}
