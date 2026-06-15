package xml_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ungerik/go-mx/xml"
)

func TestEscape(t *testing.T) {
	require.Equal(t, `a &lt; b &amp; &quot;c&quot; &apos;d&apos; &gt;`, xml.Escape(`a < b & "c" 'd' >`))
}

func TestString(t *testing.T) {
	got, err := xml.String(xml.Element("body", xml.CDATA("unescaped <raw> & text")))
	require.NoError(t, err)
	require.Equal(t, `<body><![CDATA[unescaped <raw> & text]]></body>`, got)
}
