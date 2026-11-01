package shadcn

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestMarker(t *testing.T) {
	out := render(t, Marker("", html.Role("status"),
		MarkerIcon(Spinner()),
		MarkerContent("Thinking"),
	))
	for _, want := range []string{
		`<div data-slot="marker"`, `data-variant="default"`, `role="status"`,
		`data-slot="marker-icon"`, `aria-hidden="true"`,
		`data-slot="marker-content"`, "group/marker",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	if strings.Contains(out, "cn-") {
		t.Errorf("unresolved cn-* token leaked into output: %s", out)
	}
}

func TestMarkerVariants(t *testing.T) {
	separator := render(t, Marker(MarkerSeparator, MarkerContent("Worked for 42s")))
	// The rules are pseudo-elements on both sides, and the content shrinks
	// between them through the group variant keyed on data-variant.
	for _, want := range []string{"before:flex-1", "after:flex-1", `data-variant="separator"`,
		"group-data-[variant=separator]/marker:flex-none"} {
		if !strings.Contains(separator, want) {
			t.Errorf("missing %q in %s", want, separator)
		}
	}
	border := render(t, Marker(MarkerBorder, "x"))
	if !strings.Contains(border, "border-b") || strings.Contains(border, "before:flex-1") {
		t.Errorf("border variant classes wrong: %s", border)
	}
}

func TestMarkerRenderTag(t *testing.T) {
	// The render prop is replaced by the attributes: an href makes it a
	// link, a type a button; neither is valid on the default <div>.
	if out := render(t, Marker("", html.HRef("/log"), "x")); !strings.HasPrefix(out, "<a ") {
		t.Errorf("href should render an <a>: %s", out)
	}
	if out := render(t, Marker("", html.Type("button"), "x")); !strings.HasPrefix(out, "<button ") {
		t.Errorf("type should render a <button>: %s", out)
	}
}
