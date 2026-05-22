package shadcn

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestSeparatorHorizontal(t *testing.T) {
	out := render(t, Separator(""))
	for _, want := range []string{
		`data-slot="separator"`,
		`role="separator"`,
		`data-orientation="horizontal"`,
		"bg-border",
		"shrink-0",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	if strings.Contains(out, "aria-orientation") {
		t.Errorf("horizontal separator should not set aria-orientation: %s", out)
	}
}

func TestSeparatorVertical(t *testing.T) {
	out := render(t, Separator(SeparatorVertical))
	for _, want := range []string{`data-orientation="vertical"`, `aria-orientation="vertical"`} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}

func TestSeparatorCallerRoleOverride(t *testing.T) {
	out := render(t, Separator("", html.Role("none")))
	if !strings.Contains(out, `role="none"`) || strings.Contains(out, `role="separator"`) {
		t.Errorf("caller role should override the default: %s", out)
	}
}
