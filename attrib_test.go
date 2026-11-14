package mx

import (
	"fmt"
	"strings"
	"testing"
)

func TestValidateAttribName(t *testing.T) {
	// Names real markup uses, including the punctuation of htmx, Alpine
	// and XML namespace prefixes.
	for _, name := range []string{"class", "data-x", "aria-label", "hx-on::after-request", "@click", ":class", "x-on:click.prevent", "xlink:href", "viewBox"} {
		if err := ValidateAttribName(name); err != nil {
			t.Errorf("ValidateAttribName(%q) = %v, want nil", name, err)
		}
		if !ValidAttribName(name) {
			t.Errorf("ValidAttribName(%q) = false, want true", name)
		}
	}

	if err := ValidateAttribName(""); err == nil || !strings.Contains(err.Error(), "empty") {
		t.Errorf(`ValidateAttribName("") = %v, want an error saying the name is empty`, err)
	}

	// Each of these would end the name or the start tag when written
	// unescaped into it; the error names the character, so a caller sees
	// what to fix.
	for _, tt := range []struct {
		name string
		char rune
	}{
		{"a b", ' '}, {"a\tb", '\t'}, {"a\nb", '\n'}, {`a"b`, '"'}, {"a'b", '\''},
		{"a>b", '>'}, {"a/b", '/'}, {"a=b", '='}, {"a<b", '<'}, {"a\x00b", 0}, {"a\u0085b", '\u0085'},
	} {
		err := ValidateAttribName(tt.name)
		if err == nil {
			t.Errorf("ValidateAttribName(%q) = nil, want an error", tt.name)
			continue
		}
		if want := fmt.Sprintf("%q", tt.char); !strings.Contains(err.Error(), want) {
			t.Errorf("ValidateAttribName(%q) = %q, want it to name the character %s", tt.name, err, want)
		}
		if ValidAttribName(tt.name) {
			t.Errorf("ValidAttribName(%q) = true, want false", tt.name)
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
