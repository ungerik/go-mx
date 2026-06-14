package html

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ungerik/go-mx"
)

func TestHyperlink(t *testing.T) {
	tests := []struct {
		href    string
		text    string
		attribs []mx.Attrib
		want    string
	}{
		{
			href: "https://example.com",
			text: "Example",
			want: `<a href='https://example.com'>Example</a>`,
		},
		{
			href:    "https://example.com",
			text:    "Example",
			attribs: []mx.Attrib{TargetBlank},
			want:    `<a href='https://example.com' target='_blank'>Example</a>`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := Hyperlink(tt.href, tt.text, tt.attribs...)
			require.Equal(t, tt.want, got.String())
		})
	}
}

func TestButtonTypeConstructors(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"submit", SubmitButton(Text("Save")).String(), `<button type='submit'>Save</button>`},
		{"reset", ResetButton(Text("Reset")).String(), `<button type='reset'>Reset</button>`},
		{"button", ButtonButton(Text("Toggle")).String(), `<button type='button'>Toggle</button>`},
		{"bare defaults to submit", Button(Text("Save")).String(), `<button>Save</button>`},
		{
			"attribs and children mix",
			SubmitButton(Class("primary"), ID("save"), Text("Save")).String(),
			`<button type='submit' class='primary' id='save'>Save</button>`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.got)
		})
	}
}
