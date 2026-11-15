package mx

import (
	"iter"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAttribs_Attribs(t *testing.T) {
	tests := []struct {
		name string
		a    Attribs
		want []Attrib
	}{
		{name: "nil", a: nil, want: nil},
		{name: "empty", a: Attribs{}, want: nil},
		{name: "single", a: Attribs{"a": "val_a"}, want: []Attrib{Attribute{"a", "val_a"}}},
		{name: "sort 3", a: Attribs{"c": "val_c", "b": "val_b", "a": "val_a"}, want: []Attrib{Attribute{"a", "val_a"}, Attribute{"b", "val_b"}, Attribute{"c", "val_c"}}},
		{name: "sort id", a: Attribs{"c": "val_c", "id": "val_id", "a": "val_a"}, want: []Attrib{Attribute{"id", "val_id"}, Attribute{"a", "val_a"}, Attribute{"c", "val_c"}}},
		{name: "sort id class", a: Attribs{"class": "val_class", "id": "val_id", "a": "val_a", "b": "val_b"}, want: []Attrib{Attribute{"id", "val_id"}, Attribute{"class", "val_class"}, Attribute{"a", "val_a"}, Attribute{"b", "val_b"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.a.Attribs()
			require.Equal(t, tt.want, got, "Attribs.Attribs()")
		})
	}
}

// TestDefaultAsAttribsNilFunc guards that a nil attribute func or iterator,
// or a typed-nil AttribProvider, is recognized as attributes but yields
// none, instead of being called and panicking, so NewElement, AsComponents and ModifyOnRender skip it like a
// nil Attrib.
// Value: protects=a nil func() Attrib/Attribute/[]Attrib/[]Attribute yields no attribs; fails_when=DefaultAsAttribs calls the nil func; why_new=no test passes a nil attribute func; seam=none; no_claim=custom AsAttribs replacements
func TestDefaultAsAttribsNilFunc(t *testing.T) {
	for _, tt := range []struct {
		name string
		arg  any
	}{
		{"func() Attrib", (func() Attrib)(nil)},
		{"func() Attribute", (func() Attribute)(nil)},
		{"func() []Attrib", (func() []Attrib)(nil)},
		{"func() []Attribute", (func() []Attribute)(nil)},
		{"iter.Seq[Attrib]", iter.Seq[Attrib](nil)},
		{"iter.Seq[Attribute]", iter.Seq[Attribute](nil)},
		{"typed-nil AttribProvider", (*ptrAttribProvider)(nil)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			attribs, ok := DefaultAsAttribs(tt.arg)
			require.True(t, ok)
			require.Empty(t, attribs)

			e := NewElement("p", tt.arg, "x")
			require.Empty(t, e.Attribs)
			require.Equal(t, "<p>x</p>", renderToString(t, e))
			require.Equal(t, "x", renderToString(t, AsComponents(tt.arg, "x")))
		})
	}
}

// ptrAttribProvider is an AttribProvider whose pointer-receiver Attribs
// dereferences the receiver, so a nil one panics when called.
type ptrAttribProvider struct{ attribs []Attrib }

func (p *ptrAttribProvider) Attribs() []Attrib { return p.attribs }

// TestHasAttrTags guards that hasAttrTags finds a tagged field exactly where
// ReflectAttribs would read one, so skipping ReflectAttribs for the struct
// types it rejects never drops an attribute.
func TestHasAttrTags(t *testing.T) {
	type tagged struct {
		Class string `attr:"class"`
	}
	type embedsTagged struct{ tagged }
	type unexportedTagged struct {
		class string `attr:"class"`
	}
	for _, tt := range []struct {
		name string
		v    any
		want bool
	}{
		{"tagged", tagged{}, true},
		{"promoted from embedded", embedsTagged{}, true},
		{"untagged", struct{ Class string }{}, false},
		{"skipped tag", struct {
			Class string `attr:"-"`
		}{}, false},
		{"unexported", unexportedTagged{}, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			typ := reflect.TypeOf(tt.v)
			// Twice: the first call fills the cache, the second reads it.
			for range 2 {
				if got := hasAttrTags(typ); got != tt.want {
					t.Errorf("hasAttrTags = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
