package xml_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ungerik/go-mx/xml"
)

func TestProcInst(t *testing.T) {
	require.Equal(t,
		`<?xml-stylesheet type="text/xsl" href="style.xsl"?>`,
		render(t, xml.ProcInst{Target: "xml-stylesheet", Data: `type="text/xsl" href="style.xsl"`}),
	)
	require.Equal(t, `<?php?>`, render(t, xml.ProcInst{Target: "php"}))

	for _, pi := range []xml.ProcInst{
		{Target: ""},
		{Target: "xml"},
		{Target: "XmL"},
		{Target: "bad target"},
		{Target: "t", Data: "ends ?> early"},
	} {
		_, err := xml.String(pi)
		require.Errorf(t, err, "expected error for %#v", pi)
	}
}

func TestProcInstIndentedAsChild(t *testing.T) {
	// A processing instruction used as an element child is broken and indented
	// to its siblings' depth, the same as a child element; the top-level
	// declaration stays at column 0.
	doc := &xml.Document{
		Declaration: xml.Declaration,
		Root: xml.Element("root",
			xml.ProcInst{Target: "xml-stylesheet", Data: `href="s.xsl"`},
			xml.Element("child", "x"),
		),
	}
	want := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n" +
		"<root>\n" +
		"  <?xml-stylesheet href=\"s.xsl\"?>\n" +
		"  <child>x</child>\n" +
		"</root>"
	require.Equal(t, want, renderIndent(t, doc))
}

func TestDeclaration(t *testing.T) {
	// Declaration and Decl are bare values; the writer adds the line break after
	// the closing "?>" only when content follows (see TestDocument).
	require.Equal(t, `<?xml version="1.0" encoding="UTF-8"?>`, render(t, xml.Declaration))
	require.Equal(t, `<?xml version="1.0" encoding="UTF-16"?>`, render(t, xml.Decl("1.0", "UTF-16")))
	require.Equal(t, `<?xml version="1.1"?>`, render(t, xml.Decl("1.1", "")))
}

func TestDoctype(t *testing.T) {
	require.Equal(t, `<!DOCTYPE note>`, render(t, xml.Doctype("note")))
	require.Equal(t,
		`<!DOCTYPE note SYSTEM "note.dtd">`,
		render(t, xml.Doctype(`note SYSTEM "note.dtd"`)),
	)
}
