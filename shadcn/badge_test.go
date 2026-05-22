package shadcn

import (
	"strings"
	"testing"
)

func TestBadgeDefault(t *testing.T) {
	out := render(t, Badge("", "New"))
	for _, want := range []string{`data-slot="badge"`, ">New<", "inline-flex", "bg-primary"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}

func TestBadgeVariants(t *testing.T) {
	variants := map[BadgeVariant]string{
		BadgeSecondary:   "bg-secondary",
		BadgeDestructive: "bg-destructive",
		BadgeOutline:     "text-foreground",
	}
	for v, cls := range variants {
		out := render(t, Badge(v, "x"))
		if !strings.Contains(out, cls) {
			t.Errorf("variant %s: missing class %q: %s", v, cls, out)
		}
	}
}

func TestBadgeClasses(t *testing.T) {
	if !strings.Contains(BadgeClasses(BadgeDefault), "bg-primary") {
		t.Error("BadgeClasses(BadgeDefault) should contain bg-primary")
	}
	if !strings.Contains(BadgeClasses(BadgeVariant("bogus")), "bg-primary") {
		t.Error("an unknown variant should fall back to the default classes")
	}
}
