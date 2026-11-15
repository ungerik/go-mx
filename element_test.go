package mx

import (
	"context"
	"strings"
	"testing"
)

// TestElementRenderNilAttrib guards that a nil Attrib, which a []Attrib
// passed to NewElement can carry, is an error rather than a panic, also a
// typed nil whose AttribValue would panic.
func TestElementRenderNilAttrib(t *testing.T) {
	for _, tt := range []struct {
		name string
		a    Attrib
	}{
		{"nil", nil},
		{"typed nil pointer", (*ptrAttrib)(nil)},
		{"typed nil func", funcAttrib(nil)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var b strings.Builder
			err := NewElement("p", []Attrib{NewAttrib("class", "a"), tt.a}).Render(context.Background(), NewCheckedWriter(&b))
			if err == nil {
				t.Fatalf("want an error, got %s", b.String())
			}
		})
	}
}

// TestElementAttribIndexNilAttrib guards that AttribIndex skips a nil
// Attrib, also a typed nil, instead of panicking on it.
func TestElementAttribIndexNilAttrib(t *testing.T) {
	e := NewElement("p", []Attrib{nil, (*ptrAttrib)(nil), NewAttrib("class", "a")})
	if i := e.AttribIndex("class"); i != 2 {
		t.Errorf("AttribIndex(class) = %d, want 2", i)
	}
	if i := e.AttribIndex("id"); i != -1 {
		t.Errorf("AttribIndex(id) = %d, want -1", i)
	}
}

// TestNewElementDropsNilComponents guards that a child AsComponent converts
// to nil is left out of Children, as AsComponents leaves it out, so an
// AsComponent that filters unsupported types removes them entirely.
func TestNewElementDropsNilComponents(t *testing.T) {
	base := AsComponent
	t.Cleanup(func() { AsComponent = base })
	AsComponent = func(c any) Component {
		if _, ok := c.(int); ok {
			return nil
		}
		return base(c)
	}

	e := NewElement("p", "a", 42, "b")
	if len(e.Children) != 2 {
		t.Fatalf("Children = %#v, want the two strings only", e.Children)
	}
	if got, want := renderToString(t, e), "<p>ab</p>"; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}
