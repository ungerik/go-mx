package shadcn

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestCollapsibleComposition(t *testing.T) {
	out := render(t, Collapsible(
		CollapsibleTrigger("Show more"),
		CollapsibleContent("Hidden text"),
	))
	for _, want := range []string{
		"<details ",
		`data-slot="collapsible"`,
		"group",
		"<summary ",
		`data-slot="collapsible-trigger"`,
		"[&amp;::-webkit-details-marker]:hidden",
		"<div ",
		`data-slot="collapsible-content"`,
		">Show more<",
		">Hidden text<",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}

func TestCollapsibleOpenPassesThrough(t *testing.T) {
	out := render(t, Collapsible(html.Open, CollapsibleTrigger("t")))
	if !strings.Contains(out, "open") {
		t.Errorf("html.Open should pass through to <details>: %s", out)
	}
}
