package mx

import (
	"context"
	"strings"
	"testing"
)

// TestElementRenderNilAttrib guards that a nil Attrib, which a []Attrib
// passed to NewElement can carry, is an error rather than a panic.
func TestElementRenderNilAttrib(t *testing.T) {
	var b strings.Builder
	err := NewElement("p", []Attrib{NewAttrib("class", "a"), nil}).Render(context.Background(), NewCheckedWriter(&b))
	if err == nil {
		t.Fatalf("want an error, got %s", b.String())
	}
}

// TestElementAttribIndexNilAttrib guards that AttribIndex skips a nil
// Attrib instead of panicking on it.
func TestElementAttribIndexNilAttrib(t *testing.T) {
	e := NewElement("p", []Attrib{nil, NewAttrib("class", "a")})
	if i := e.AttribIndex("class"); i != 1 {
		t.Errorf("AttribIndex(class) = %d, want 1", i)
	}
	if i := e.AttribIndex("id"); i != -1 {
		t.Errorf("AttribIndex(id) = %d, want -1", i)
	}
}
