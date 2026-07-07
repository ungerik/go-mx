package shadcn

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestTextarea(t *testing.T) {
	out := render(t, Textarea(html.Rows(4), "Hello"))
	for _, want := range []string{
		`data-slot="textarea"`,
		`rows="4"`,
		">Hello<",
		"min-h-16",
		"field-sizing-content",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}

func TestTextareaID(t *testing.T) {
	// TextareaID prepends the id so a Label(html.For(id)) can bind to it.
	out := render(t, TextareaID("msg", "Hello"))
	for _, want := range []string{`data-slot="textarea"`, `id="msg"`, ">Hello<"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}
