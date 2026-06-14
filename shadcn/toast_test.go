package shadcn

import (
	"strings"
	"testing"
)

func TestToaster(t *testing.T) {
	out := render(t, Toaster())
	for _, want := range []string{
		`data-slot="toaster"`,
		"fixed",
		"z-[100]",
		"pointer-events-none",
		"window.toast",
		"<script>",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}
