package shadcn

import (
	"strings"
	"testing"
)

func TestEmpty(t *testing.T) {
	out := render(t, Empty(
		EmptyHeader(
			EmptyMedia(EmptyMediaIcon, "!"),
			EmptyTitle("No results"),
			EmptyDescription("Try a different search."),
		),
		EmptyContent(Button("", "", "Clear filters")),
	))
	for _, want := range []string{
		`data-slot="empty"`, `data-slot="empty-header"`, `data-slot="empty-title"`,
		`data-slot="empty-description"`, `data-slot="empty-content"`,
		// Upstream names the media part's slot "empty-icon"; keep parity.
		`data-slot="empty-icon"`,
		"border-dashed", "text-lg",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	// cn-font-heading (a theme font var this port does not define) must be
	// dropped, not leaked, or the title would reference a missing utility.
	if strings.Contains(out, "cn-") {
		t.Errorf("unresolved cn-* token leaked into output: %s", out)
	}
}

func TestEmptyMediaVariants(t *testing.T) {
	def := render(t, EmptyMedia("", "x"))
	if !strings.Contains(def, "bg-transparent") || !strings.Contains(def, `data-variant="default"`) {
		t.Errorf("empty variant should resolve to default: %s", def)
	}
	icon := render(t, EmptyMedia(EmptyMediaIcon, "x"))
	if !strings.Contains(icon, "bg-muted") || !strings.Contains(icon, "size-10") {
		t.Errorf("icon variant should frame in a muted square: %s", icon)
	}
}
