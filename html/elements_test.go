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
