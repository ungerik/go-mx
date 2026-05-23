package shadcn

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestCheckboxDefault(t *testing.T) {
	out := render(t, Checkbox(html.ID("agree")))
	for _, want := range []string{
		`<input `,
		`data-slot="checkbox"`,
		`type="checkbox"`,
		`id="agree"`,
		"peer",
		"appearance-none",
		"checked:bg-primary",
		"checked:border-primary",
		"checked:bg-[url(",
		"data:image/svg+xml",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	if strings.Contains(out, "data-[state=") {
		t.Errorf("Radix data-[state= selector should have been rewritten: %s", out)
	}
}

func TestCheckboxCallerAttrsPassThrough(t *testing.T) {
	out := render(t, Checkbox(html.Name("agree"), html.Value("yes"), html.Checked, html.Disabled))
	for _, want := range []string{`name="agree"`, `value="yes"`, "checked", "disabled"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}

func TestCheckboxIndeterminateScript(t *testing.T) {
	out := render(t, CheckboxIndeterminateScript("agree"))
	for _, want := range []string{
		"<script>",
		`document.getElementById('agree').indeterminate=true`,
		"</script>",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}

func TestCheckboxIndeterminateScriptValidatesID(t *testing.T) {
	for _, bad := range []string{"", "bad id", "x.y", "evil');alert(1"} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("expected panic for id %q", bad)
				}
			}()
			CheckboxIndeterminateScript(bad)
		}()
	}
}
