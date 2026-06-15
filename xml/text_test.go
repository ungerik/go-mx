package xml_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ungerik/go-mx/xml"
)

func TestTextEscaping(t *testing.T) {
	got := render(t, xml.Element("p", `a < b & "c" 'd' >`))
	require.Equal(t, `<p>a &lt; b &amp; &quot;c&quot; &apos;d&apos; &gt;</p>`, got)
}
