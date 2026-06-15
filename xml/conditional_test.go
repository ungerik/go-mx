package xml_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/xml"
)

func TestConditionalAndForEach(t *testing.T) {
	got := render(t, xml.Element("list",
		xml.If(true, xml.Element("a")),
		xml.If(false, xml.Element("b")),
		xml.ForEach([]string{"x", "y"}, func(s string) *mx.Element { return xml.Element("item", s) }),
	))
	require.Equal(t, `<list><a></a><item>x</item><item>y</item></list>`, got)
}
