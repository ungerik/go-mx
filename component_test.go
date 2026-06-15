package mx

import (
	"testing"
)

type stringerType struct{ s string }

func (s stringerType) String() string { return s.s }

type errorType struct{ msg string }

func (e errorType) Error() string { return e.msg }

type namedString string

// plainStruct implements neither Component, fmt.Stringer nor error, so it falls
// through to the fmt.Sprint default case.
type plainStruct struct{ X int }

func TestDefaultAsComponent_nilAndComponentPassthrough(t *testing.T) {
	if c := DefaultAsComponent(nil); c != nil {
		t.Errorf("DefaultAsComponent(nil) = %#v, want nil", c)
	}

	want := Text("kept")
	if got := DefaultAsComponent(want); got != want {
		t.Errorf("DefaultAsComponent(Component) = %#v, want same instance %#v", got, want)
	}
}

// Residual types (anything not matched by an explicit case) get their standard
// fmt.Sprint textual representation, rendered as an escaped Text node.
func TestDefaultAsComponent_stringifiesResidualTypes(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want string
	}{
		{name: "string", in: "hello", want: "hello"},
		{name: "int", in: 42, want: "42"},
		{name: "negative int", in: -7, want: "-7"},
		{name: "uint", in: uint(7), want: "7"},
		{name: "bool", in: true, want: "true"},
		{name: "float", in: 3.5, want: "3.5"},
		// A named string type is not the builtin string, so it falls through
		// to pretty.Sprint, which backtick-quotes it.
		{name: "named string", in: namedString("nm"), want: "`nm`"},
		{name: "fmt.Stringer", in: stringerType{"S"}, want: "S"},
		{name: "error", in: errorType{"boom"}, want: "boom"},
		// pretty.Sprint tags the type and names fields, and dereferences the
		// pointer (no leading & to escape, unlike fmt.Sprint).
		{name: "struct", in: plainStruct{1}, want: "plainStruct{X:1}"},
		{name: "pointer to struct", in: &plainStruct{1}, want: "plainStruct{X:1}"},
		{name: "slice", in: []int{1, 2, 3}, want: "[1,2,3]"},
		{name: "map", in: map[string]int{"a": 1}, want: "{`a`:1}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := DefaultAsComponent(tt.in)
			if c == nil {
				t.Fatalf("DefaultAsComponent(%#v) = nil", tt.in)
			}
			if got := renderToString(t, c); got != tt.want {
				t.Errorf("render = %q, want %q", got, tt.want)
			}
		})
	}
}

// A stringified residual value must never be able to inject markup: the Writer
// escapes the Text node it is wrapped in. This holds for every markup target,
// since html and svg render through the same CheckedWriter / TextEscaper.
func TestDefaultAsComponent_residualTextIsEscaped(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want string
	}{
		{name: "stringer with tags", in: stringerType{`<script>alert(1)</script>`}, want: `&lt;script&gt;alert(1)&lt;/script&gt;`},
		{name: "stringer with entities", in: stringerType{`a & b`}, want: `a &amp; b`},
		{name: "error with quotes", in: errorType{`he said "hi" & 'bye'`}, want: `he said &quot;hi&quot; &amp; &apos;bye&apos;`},
		{name: "struct with markup field", in: struct{ S string }{S: "<b>"}, want: "{S:`&lt;b&gt;`}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := renderToString(t, DefaultAsComponent(tt.in)); got != tt.want {
				t.Errorf("render = %q, want %q (markup not escaped)", got, tt.want)
			}
		})
	}
}

// Escaping a residual value inside an element must produce the same output as
// escaping the equivalent string, i.e. residual stringification is just text.
func TestDefaultAsComponent_residualMatchesEquivalentString(t *testing.T) {
	withResidual := renderToString(t, NewElement("div", stringerType{"<x>"}))
	withString := renderToString(t, NewElement("div", "<x>"))
	if withResidual != withString {
		t.Errorf("residual %q != string %q", withResidual, withString)
	}
}
