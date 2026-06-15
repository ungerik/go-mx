package html

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// markupStringer.String returns text that looks like markup. Passed as a child
// it must be treated as text and escaped, never injected as elements.
type markupStringer struct{ s string }

func (m markupStringer) String() string { return m.s }

// unexpectedStruct implements neither mx.Component, fmt.Stringer nor error, so
// the element constructors fall back to fmt.Sprint and escape the result. This
// is the deliberate fmt.Sprint trade-off: a struct mistaken for markup renders
// as its escaped Go representation (never injecting elements) instead of
// erroring, with no compile-time error.
type unexpectedStruct struct{ Name string }

// Non-Component children that represent invalid markup must be escaped when
// rendered through the public html element constructors, so a stringified value
// can never inject tags into the output.
func TestElementEscapesNonComponentMarkup(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{
			"raw string with script tag",
			Div(`<script>alert(1)</script>`).String(),
			`<div>&lt;script&gt;alert(1)&lt;/script&gt;</div>`,
		},
		{
			"raw string with ampersand and quotes",
			P(`a & "b" 'c'`).String(),
			`<p>a &amp; &quot;b&quot; &apos;c&apos;</p>`,
		},
		{
			"stringer returning markup",
			Span(markupStringer{`<img src=x onerror=alert(1)>`}).String(),
			`<span>&lt;img src=x onerror=alert(1)&gt;</span>`,
		},
		{
			// Unexpected struct child: go-pretty gives a type-tagged,
			// single-line dump and the markup in the field value is escaped.
			"unexpected struct child stringified and escaped",
			Div(unexpectedStruct{Name: "<b>"}).String(),
			"<div>unexpectedStruct{Name:`&lt;b&gt;`}</div>",
		},
		{
			// go-pretty dereferences the pointer (no leading &), so it renders
			// the same as the value struct, still fully escaped.
			"pointer to unexpected struct dereferenced and escaped",
			Div(&unexpectedStruct{Name: "x"}).String(),
			"<div>unexpectedStruct{Name:`x`}</div>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.got)
		})
	}
}
