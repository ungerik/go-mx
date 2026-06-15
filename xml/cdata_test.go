package xml_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/xml"
)

func TestCDATA(t *testing.T) {
	require.Equal(t, `<![CDATA[a < b & c]]>`, render(t, xml.CDATA("a < b & c")))

	// A CDATA section cannot contain its own terminator.
	_, err := xml.String(xml.CDATA("oops ]]> here"))
	require.Error(t, err)
}

// ExampleCDATA renders an element tree to stdout with indented formatting,
// contrasting normal text content — where &, < and > are escaped — with a CDATA
// section, whose content (here an HTML fragment) is emitted verbatim. The go
// test framework compares the printed output against the Output comment below.
func ExampleCDATA() {
	item := xml.Element("item",
		// Plain text is escaped: & < > become entities.
		xml.Element("title", `Books & Comics: <50% off>`),
		xml.Element("description",
			xml.Comment("A CDATA section keeps its markup raw, with no escaping:"),
			xml.CDATA(`<p>Save <strong>50%</strong> on all items!</p>`),
		),
	)

	w := mx.NewCheckedWriter(os.Stdout).WithIndent("", "  ")
	if err := item.Render(context.Background(), w); err != nil {
		panic(err)
	}

	// Output:
	// <item>
	//   <title>Books &amp; Comics: &lt;50% off&gt;</title>
	//   <description>
	//     <!-- A CDATA section keeps its markup raw, with no escaping: -->
	//     <![CDATA[<p>Save <strong>50%</strong> on all items!</p>]]>
	//   </description>
	// </item>
}
