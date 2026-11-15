package mx

import (
	"context"
	"testing"
)

// func TestComponentSlice(t *testing.T) {
// 	c := RawComponent("Test")
// 	tests := []struct {
// 		name string
// 		c    Component
// 		want []Component
// 	}{
// 		{name: "nil", c: nil, want: nil},
// 		{name: "single", c: c, want: []Component{c}},
// 		{name: "Components empty", c: Components{}, want: []Component{}},
// 		{name: "Components single", c: Components{c}, want: []Component{c}},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := ComponentSlice(tt.c); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("ComponentSlice(%#v) = %#v, want %#v", tt.c, got, tt.want)
// 			}
// 		})
// 	}
// }

// TestAsComponentFilterAttribs runs the AsComponent docs example that drops
// an attribute passed as a child instead of rendering it as text.
func TestAsComponentFilterAttribs(t *testing.T) {
	base := AsComponent
	t.Cleanup(func() { AsComponent = base })
	AsComponent = func(c any) Component {
		if _, isComponent := c.(Component); !isComponent {
			if _, isAttribs := AsAttribs(c); isAttribs {
				return nil
			}
		}
		return base(c)
	}

	// Every attribute shape AsAttribs accepts is dropped, not only an Attrib.
	cs := AsComponents(
		NewAttrib("class", "x"),
		[]Attrib{NewAttrib("id", "y")},
		Attribs{"title": "z"},
		"text",
	)
	if len(cs) != 1 {
		t.Fatalf("AsComponents = %#v, want the text only", cs)
	}
	if got, want := renderToString(t, cs), "text"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	// A Component that also provides attributes stays content.
	if got, want := renderToString(t, AsComponents(componentWithAttribs{})), "content"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// componentWithAttribs is a Component that is also an AttribProvider.
type componentWithAttribs struct{}

func (componentWithAttribs) Render(ctx context.Context, w Writer) error {
	return Text("content").Render(ctx, w)
}

func (componentWithAttribs) Attribs() []Attrib { return []Attrib{NewAttrib("class", "stray")} }
