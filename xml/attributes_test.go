package xml_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ungerik/go-mx/xml"
)

func TestAttribValueTypes(t *testing.T) {
	got := render(t, xml.Element("v",
		xml.Attrib("s", "text"),
		xml.Attrib("i", 100),
		xml.Attrib("f", 0.00005),
	))
	require.Equal(t, `<v s="text" i="100" f="0.00005"></v>`, got)
}

func TestNamespaceAttribs(t *testing.T) {
	got := render(t, xml.Element("root",
		xml.XMLNS("urn:example"),
		xml.XMLNSPrefix("x", "urn:x"),
		xml.XMLLang("en"),
		xml.XMLSpacePreserve,
	))
	require.Equal(t,
		`<root xmlns="urn:example" xmlns:x="urn:x" xml:lang="en" xml:space="preserve"></root>`,
		got,
	)
}
