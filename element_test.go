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

// TestAsAttribsOrComponent guards the per-argument sorting that NewElement
// and html.NewDocument share: attributes when AsAttribs accepts the
// argument (asked first, so a Component that is also an Attrib counts as
// attributes), otherwise the AsComponent child, and neither for nil.
func TestAsAttribsOrComponent(t *testing.T) {
	for _, tt := range []struct {
		name        string
		arg         any
		wantAttribs int
		wantChild   bool
	}{
		{"nil", nil, 0, false},
		{"Attrib", NewAttrib("class", "a"), 1, false},
		{"[]Attrib", []Attrib{NewAttrib("class", "a"), NewAttrib("id", "b")}, 2, false},
		{"empty []Attrib", []Attrib{}, 0, false},
		{"string", "text", 0, true},
		{"Component", Text("text"), 0, true},
		{"Component and Attrib", componentAttrib{}, 1, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			attribs, child := AsAttribsOrComponent(tt.arg)
			if len(attribs) != tt.wantAttribs || (child != nil) != tt.wantChild {
				t.Errorf("got %d attribs and child %v, want %d attribs and a child: %v", len(attribs), child, tt.wantAttribs, tt.wantChild)
			}
		})
	}

	t.Run("AsComponent converts to nil", func(t *testing.T) {
		base := AsComponent
		t.Cleanup(func() { AsComponent = base })
		AsComponent = func(any) Component { return nil }
		attribs, child := AsAttribsOrComponent("text")
		if attribs != nil || child != nil {
			t.Errorf("got %v and %v, want neither", attribs, child)
		}
	})
}
