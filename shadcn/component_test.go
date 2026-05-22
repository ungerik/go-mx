package shadcn

import (
	"context"
	"strings"
	"testing"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// render renders a component with a double-quote CheckedWriter and fails the
// test on any render error (which includes the duplicate-attribute check).
func render(t *testing.T, c mx.Component) string {
	t.Helper()
	var b strings.Builder
	if err := c.Render(context.Background(), mx.NewCheckedWriter(&b)); err != nil {
		t.Fatalf("render error: %v\npartial output:\n%s", err, b.String())
	}
	return b.String()
}

func TestFinish(t *testing.T) {
	t.Run("base only", func(t *testing.T) {
		out := render(t, finish(html.Div(), "thing", "p-2 text-sm"))
		if !strings.Contains(out, `data-slot="thing"`) {
			t.Errorf("missing data-slot: %s", out)
		}
		if !strings.Contains(out, `class="p-2 text-sm"`) {
			t.Errorf("missing merged class: %s", out)
		}
	})

	t.Run("caller class overrides base", func(t *testing.T) {
		out := render(t, finish(html.Div(html.Class("p-8")), "thing", "p-2 text-sm"))
		if strings.Contains(out, "p-2") {
			t.Errorf("conflicting base class p-2 should be dropped: %s", out)
		}
		if !strings.Contains(out, "p-8") || !strings.Contains(out, "text-sm") {
			t.Errorf("expected p-8 and text-sm: %s", out)
		}
	})

	t.Run("two caller class attribs merge into one", func(t *testing.T) {
		out := render(t, finish(html.Div(html.Class("p-2"), html.Class("m-4")), "thing", "text-sm"))
		if n := strings.Count(out, "class="); n != 1 {
			t.Errorf("expected exactly one class attribute, got %d: %s", n, out)
		}
		for _, c := range []string{"text-sm", "p-2", "m-4"} {
			if !strings.Contains(out, c) {
				t.Errorf("missing %s: %s", c, out)
			}
		}
	})

	t.Run("caller data-slot is dropped", func(t *testing.T) {
		out := render(t, finish(html.Div(html.DataAttr("slot", "evil")), "thing", ""))
		if !strings.Contains(out, `data-slot="thing"`) || strings.Contains(out, "evil") {
			t.Errorf("caller data-slot must not override component slot: %s", out)
		}
	})

	t.Run("duplicate attribute last wins, no render error", func(t *testing.T) {
		out := render(t, finish(html.Div(html.ID("a"), html.ID("b")), "thing", ""))
		if !strings.Contains(out, `id="b"`) || strings.Contains(out, `id="a"`) {
			t.Errorf("last id should win: %s", out)
		}
	})

	t.Run("attribute order: data-slot first, class last", func(t *testing.T) {
		out := render(t, finish(html.Div(html.ID("x")), "thing", "p-2"))
		if !strings.HasPrefix(out, `<div data-slot="thing" id="x" class="`) {
			t.Errorf("unexpected attribute order: %s", out)
		}
	})
}
