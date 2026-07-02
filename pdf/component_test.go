package pdf

import (
	"testing"
)

type pdfStringer struct{ s string }

func (s pdfStringer) String() string { return s.s }

type pdfErr struct{ msg string }

func (e pdfErr) Error() string { return e.msg }

type pdfPlainStruct struct{ X int }

// Residual types get a go-pretty textual representation as a Text node (error
// and fmt.Stringer keep their own text). A PDF is not markup, so there is no
// escaping step — the value is simply drawn as text.
func TestAsComponent_stringifiesResidualTypes(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want Text
	}{
		{name: "string", in: "hi", want: Text("hi")},
		{name: "int", in: 42, want: Text("42")},
		{name: "bool", in: false, want: Text("false")},
		{name: "float", in: 1.5, want: Text("1.5")},
		{name: "fmt.Stringer", in: pdfStringer{"S"}, want: Text("S")},
		{name: "error", in: pdfErr{"boom"}, want: Text("boom")},
		// pretty.Sprint tags the type and names fields (vs fmt's anonymous "{1}").
		{name: "struct", in: pdfPlainStruct{1}, want: Text("pdfPlainStruct{X:1}")},
		{name: "slice", in: []int{1, 2}, want: Text("[1,2]")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := AsComponent(tt.in)
			txt, ok := c.(Text)
			if !ok {
				t.Fatalf("AsComponent(%#v) = %#v, want Text", tt.in, c)
			}
			if txt != tt.want {
				t.Errorf("AsComponent(%#v) = %q, want %q", tt.in, txt, tt.want)
			}
		})
	}
}
