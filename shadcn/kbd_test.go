package shadcn

import (
	"strings"
	"testing"
)

func TestKbd(t *testing.T) {
	out := render(t, Kbd("⌘"))
	for _, want := range []string{"<kbd", `data-slot="kbd"`, ">⌘</kbd>", "bg-muted", "font-sans"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	// The class string is reconstructed from the cn-* style sheet; a leaked
	// cn-* token would silently render unstyled (no rule ships with this port).
	if strings.Contains(out, "cn-") {
		t.Errorf("unresolved cn-* token leaked into output: %s", out)
	}
}

func TestKbdGroup(t *testing.T) {
	out := render(t, KbdGroup(Kbd("Ctrl"), Kbd("B")))
	// Upstream renders the group as a <kbd> element too, not a <div>.
	if !strings.HasPrefix(out, "<kbd") {
		t.Errorf("KbdGroup should render a <kbd> element: %s", out)
	}
	if !strings.Contains(out, `data-slot="kbd-group"`) {
		t.Errorf("missing data-slot: %s", out)
	}
	if strings.Count(out, `data-slot="kbd"`) != 2 {
		t.Errorf("expected two nested Kbd keys: %s", out)
	}
}
