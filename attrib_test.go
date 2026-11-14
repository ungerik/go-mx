package mx

import "testing"

func TestValidAttribName(t *testing.T) {
	// Names real markup uses, including the punctuation of htmx, Alpine
	// and XML namespace prefixes.
	for _, name := range []string{"class", "data-x", "aria-label", "hx-on::after-request", "@click", ":class", "x-on:click.prevent", "xlink:href", "viewBox"} {
		if !ValidAttribName(name) {
			t.Errorf("ValidAttribName(%q) = false, want true", name)
		}
	}
	// Each of these would end the name or the start tag when written
	// unescaped into it.
	for _, name := range []string{"", "a b", "a\tb", "a\nb", `a"b`, "a'b", "a>b", "a/b", "a=b", "a<b", "a\x00b", "a\u0085b"} {
		if ValidAttribName(name) {
			t.Errorf("ValidAttribName(%q) = true, want false", name)
		}
	}
}

func TestAttributeValidate(t *testing.T) {
	if err := (Attribute{Name: "class"}).Validate(); err != nil {
		t.Errorf("valid name: %v", err)
	}
	for _, name := range []string{"", "x onload"} {
		if (Attribute{Name: name}).Valid() {
			t.Errorf("Attribute{Name: %q} is valid, want invalid", name)
		}
	}
}
