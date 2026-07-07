package shadcn

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestForm(t *testing.T) {
	// Form is just the tagged native <form>; the field layout lives in the
	// Field system (see field_test.go).
	out := render(t, Form(html.Attrib("method", "post"),
		Field("", FieldLabelFor("u", "Username"), InputID("u")),
	))
	for _, want := range []string{"<form ", `data-slot="form"`, `method="post"`, `data-slot="field"`} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}
