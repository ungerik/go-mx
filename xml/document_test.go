package xml_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/xml"
)

func TestDocument(t *testing.T) {
	doc := xml.NewDocument(xml.Element("note", xml.Element("to", "Tove")))

	// Compact: the writer breaks the line after the declaration's "?>", so the
	// root starts on the next line; the rest stays compact.
	require.Equal(t,
		"<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<note><to>Tove</to></note>",
		render(t, doc),
	)

	// Indented: no blank line after the declaration — indentation's own line
	// break replaces the writer's "?>" break rather than stacking on it.
	wantIndent := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n" +
		"<note>\n" +
		"  <to>Tove</to>\n" +
		"</note>"
	require.Equal(t, wantIndent, renderIndent(t, doc))
}

func TestDocumentWithProlog(t *testing.T) {
	doc := &xml.Document{
		Declaration: xml.Declaration,
		Prolog:      mx.Components{xml.ProcInst{Target: "xml-stylesheet", Data: `href="s.xsl"`}, xml.Comment("a note")},
		Root:        xml.Element("r"),
	}
	// The declaration and the processing instruction each end with "?>", so each
	// is followed by a line break in indented output.
	want := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n" +
		"<?xml-stylesheet href=\"s.xsl\"?>\n" +
		"<!-- a note -->\n" +
		"<r></r>"
	require.Equal(t, want, renderIndent(t, doc))
}
