package shadcn

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestFormComposition(t *testing.T) {
	out := render(t, Form(
		FormItem(
			FormLabel(html.For("u"), html.DataAttr("error", "true"), "Username"),
			Input(html.ID("u")),
			FormDescription("Your public display name."),
			FormMessage("Username is required."),
		),
	))
	for _, want := range []string{
		`data-slot="form"`,
		"<form ",
		`data-slot="form-item"`,
		"grid gap-2",
		`data-slot="form-label"`,
		"data-[error=true]:text-destructive",
		`data-error="true"`,
		`data-slot="form-description"`,
		"text-muted-foreground",
		`data-slot="form-message"`,
		"text-destructive",
		"Username",
		"Your public display name.",
		"Username is required.",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}
