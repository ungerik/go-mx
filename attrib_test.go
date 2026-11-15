package mx

import (
	"context"
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

// TestConstAttribString guards that a ConstAttrib formats like Attribute
// wherever fmt prints it (%v, error messages), as name='value' instead of its
// raw "name=value". Rendering one passed as a child is guarded by
// TestDefaultAsComponentAttrib.
func TestConstAttribString(t *testing.T) {
	for _, tt := range []struct {
		a    ConstAttrib
		want string
	}{
		{"type=text", "type='text'"},
		{"title=it's", "title='it&apos;s'"},
		{"data-x=a=b", "data-x='a=b'"},
		{"hidden=", "hidden=''"},
		{"hidden", "hidden='hidden'"},
	} {
		if got := tt.a.String(); got != tt.want {
			t.Errorf("ConstAttrib(%q).String() = %q, want %q", string(tt.a), got, tt.want)
		}
	}
}

// TestConstAttribWithoutEquals guards that a ConstAttrib without '=' is both
// name and value, as a BoolAttrib, instead of panicking, and renders so.
func TestConstAttribWithoutEquals(t *testing.T) {
	a := ConstAttrib("hidden")
	if got := a.AttribName(); got != "hidden" {
		t.Errorf("AttribName() = %q, want %q", got, "hidden")
	}
	value, err := a.AttribValue(t.Context())
	if err != nil || value != "hidden" {
		t.Errorf("AttribValue() = %q, %v, want %q, nil", value, err, "hidden")
	}
	var b strings.Builder
	err = NewElement("div", a).Render(t.Context(), NewCheckedWriter(&b))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := b.String(), `<div hidden="hidden"></div>`; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

// TestAttribStringContext guards that AttribStringContext formats like
// AttribString, values the attribute with the given context and returns a
// failing value's error instead of formatting it into the text.
func TestAttribStringContext(t *testing.T) {
	got, err := AttribStringContext(t.Context(), ConstAttrib("title=it's"))
	if err != nil || got != "title='it&apos;s'" {
		t.Errorf("got %q, %v, want %q, nil", got, err, "title='it&apos;s'")
	}

	ctx := context.WithValue(t.Context(), ctxAttribKey{}, "ann")
	got, err = AttribStringContext(ctx, ctxAttrib{})
	if err != nil || got != "data-user='ann'" {
		t.Errorf("got %q, %v, want %q, nil", got, err, "data-user='ann'")
	}

	got, err = AttribStringContext(t.Context(), ErrAttribf("class", "bad class"))
	if err == nil || got != "" || !strings.Contains(err.Error(), "bad class") {
		t.Errorf("got %q, %v, want \"\" and the ErrAttrib error", got, err)
	}
}

// TestAttribStringEscapesName guards that AttribString and
// AttribStringContext escape an invalid name as they escape the value, so
// their result, documented as safe as text content, cannot inject markup
// through the name.
func TestAttribStringEscapesName(t *testing.T) {
	a := NewAttrib("<img src=x onerror=alert(1)>", "v")
	const want = "&lt;img src=x onerror=alert(1)>='v'"
	if got := AttribString(a); got != want {
		t.Errorf("AttribString = %q, want %q", got, want)
	}
	got, err := AttribStringContext(t.Context(), a)
	if err != nil || got != want {
		t.Errorf("AttribStringContext = %q, %v, want %q, nil", got, err, want)
	}
}

// ptrAttrib has pointer receivers that dereference, so calling a method of
// a nil *ptrAttrib panics, as with most pointer Attrib types.
type ptrAttrib struct{ name, value string }

func (a *ptrAttrib) AttribName() string { return a.name }

func (a *ptrAttrib) AttribValue(context.Context) (string, error) { return a.value, nil }

// funcAttrib is an Attrib whose value is computed by calling it, so calling
// AttribValue on a nil funcAttrib panics.
type funcAttrib func() string

func (funcAttrib) AttribName() string { return "data-f" }

func (f funcAttrib) AttribValue(context.Context) (string, error) { return f(), nil }

// TestIsNilAttrib guards that IsNilAttrib catches a typed nil pointer or func
// as well as a nil interface, and reports false for a usable Attrib.
func TestIsNilAttrib(t *testing.T) {
	for _, tt := range []struct {
		name string
		a    Attrib
		want bool
	}{
		{"nil interface", nil, true},
		{"nil pointer", (*ptrAttrib)(nil), true},
		{"nil func", funcAttrib(nil), true},
		{"pointer", &ptrAttrib{"title", "x"}, false},
		{"func", funcAttrib(func() string { return "x" }), false},
		{"value", NewAttrib("title", "x"), false},
		{"empty ConstAttrib", ConstAttrib(""), false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNilAttrib(tt.a); got != tt.want {
				t.Errorf("IsNilAttrib = %v, want %v", got, tt.want)
			}
		})
	}
}
