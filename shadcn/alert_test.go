package shadcn

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestAlertDefault(t *testing.T) {
	out := render(t, Alert(AlertDefault, "Heads up"))
	for _, want := range []string{
		`data-slot="alert"`,
		`role="alert"`,
		">Heads up<",
		"relative", // base classes
		"bg-card",  // default variant
		"text-card-foreground",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}

// TestAlertDestructive doubles as a regression guard: the destructive variant
// relies on the arbitrary-variant class *:data-[slot=alert-description]:
// text-destructive/90, which Cn must preserve.
func TestAlertDestructive(t *testing.T) {
	out := render(t, Alert(AlertDestructive))
	for _, want := range []string{
		"text-destructive",
		"*:data-[slot=alert-description]:text-destructive/90",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}

func TestAlertCallerRoleOverride(t *testing.T) {
	out := render(t, Alert(AlertDefault, html.Role("status")))
	if !strings.Contains(out, `role="status"`) || strings.Contains(out, `role="alert"`) {
		t.Errorf("caller role should override the default: %s", out)
	}
}

func TestAlertTitleAndDescription(t *testing.T) {
	title := render(t, AlertTitle("Title"))
	if !strings.Contains(title, `data-slot="alert-title"`) ||
		!strings.Contains(title, "font-medium") || !strings.Contains(title, ">Title<") {
		t.Errorf("unexpected AlertTitle output: %s", title)
	}
	desc := render(t, AlertDescription("Body"))
	if !strings.Contains(desc, `data-slot="alert-description"`) ||
		!strings.Contains(desc, "text-muted-foreground") || !strings.Contains(desc, ">Body<") {
		t.Errorf("unexpected AlertDescription output: %s", desc)
	}
}
