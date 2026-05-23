package shadcn

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/hx"
)

func TestToggleGroupSingle(t *testing.T) {
	out := render(t, ToggleGroup(ToggleGroupSingle, ToggleOutline, ToggleSizeSM, "align",
		ToggleGroupItem("align", "left", ToggleOutline, ToggleSizeSM, "L"),
		ToggleGroupItem("align", "center", ToggleOutline, ToggleSizeSM, "C"),
		ToggleGroupItem("align", "right", ToggleOutline, ToggleSizeSM, "R"),
	))
	for _, want := range []string{
		`data-slot="toggle-group"`,
		`role="group"`,
		`data-type="single"`,
		`data-variant="outline"`,
		`data-size="sm"`,
		`data-toggle-group="align"`,
		`data-slot="toggle-group-item"`,
		`type="button"`,
		`aria-pressed="false"`,
		`data-toggle-group-value="left"`,
		`onclick="toggleGroupClick(this)"`,
		"border",                 // outline variant from ToggleClasses
		"first:rounded-l-md",     // join class
		"<script>",
		"window.toggleGroupClick",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	if n := strings.Count(out, "<script>"); n != 1 {
		t.Errorf("toggleGroupClick script should be emitted once, got %d: %s", n, out)
	}
}

func TestToggleGroupMultiple(t *testing.T) {
	out := render(t, ToggleGroup(ToggleGroupMultiple, "", "", "fmt",
		ToggleGroupItem("fmt", "bold", "", "", "B"),
	))
	if !strings.Contains(out, `data-type="multiple"`) {
		t.Errorf("expected data-type=multiple: %s", out)
	}
}

func TestToggleGroupItemHXOptOut(t *testing.T) {
	out := render(t, ToggleGroupItem("g", "x", "", "", hx.Post("/x"), "X"))
	if !strings.Contains(out, `hx-post="/x"`) {
		t.Errorf("hx-post should pass through: %s", out)
	}
	if strings.Contains(out, "onclick=") {
		t.Errorf("default onclick should be skipped when hx-* present: %s", out)
	}
}

func TestToggleGroupValidateIDPanics(t *testing.T) {
	for _, bad := range []string{"", "bad id"} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("expected ToggleGroup panic for id %q", bad)
				}
			}()
			ToggleGroup(ToggleGroupSingle, "", "", bad)
		}()
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("expected ToggleGroupItem panic for groupID %q", bad)
				}
			}()
			ToggleGroupItem(bad, "v", "", "")
		}()
	}
}
