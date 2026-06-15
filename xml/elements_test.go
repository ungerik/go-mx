package xml_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/xml"
)

func TestElement(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"empty close tag", render(t, xml.Element("note")), `<note></note>`},
		{
			"attribs and nested children",
			render(t, xml.Element("note", xml.Attrib("id", 1), xml.Element("to", "Tove"), xml.Element("from", "Jani"))),
			`<note id="1"><to>Tove</to><from>Jani</from></note>`,
		},
		{"self-closing empty element", render(t, xml.EmptyElement("br")), `<br/>`},
		{
			"empty element with attribs",
			render(t, xml.EmptyElement("img", xml.Attrib("src", "x.png"))),
			`<img src="x.png"/>`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.got)
		})
	}
}

func TestElementNS(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{
			"prefixed element",
			render(t, xml.ElementNS("soap", "Envelope", xml.Attrib("id", 1))),
			`<soap:Envelope id="1"></soap:Envelope>`,
		},
		{"empty prefixed element", render(t, xml.EmptyElementNS("ns", "leaf")), `<ns:leaf/>`},
		{"empty prefix falls back", render(t, xml.ElementNS("", "plain")), `<plain></plain>`},
		{
			"prefixed attribute",
			render(t, xml.Element("e", xml.AttribNS("xlink", "href", "https://example.com"))),
			`<e xlink:href="https://example.com"></e>`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.got)
		})
	}
}

// ExampleElementNS renders a typical namespaced XML document — a SOAP envelope
// with a prefixed namespace on the envelope and a second one on a nested
// element — to stdout with indented formatting. The go test framework compares
// the printed output against the Output comment below.
func ExampleElementNS() {
	doc := xml.NewDocument(
		xml.ElementNS("soap", "Envelope",
			xml.XMLNSPrefix("soap", "http://www.w3.org/2003/05/soap-envelope"),
			xml.ElementNS("soap", "Body",
				xml.ElementNS("m", "GetPrice",
					xml.XMLNSPrefix("m", "https://example.com/prices"),
					xml.ElementNS("m", "Item", "Apples"),
				),
			),
		),
	)

	w := mx.NewCheckedWriter(os.Stdout).WithIndent("", "  ")
	if err := doc.Render(context.Background(), w); err != nil {
		panic(err)
	}

	// Output:
	// <?xml version="1.0" encoding="UTF-8"?>
	// <soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	//   <soap:Body>
	//     <m:GetPrice xmlns:m="https://example.com/prices">
	//       <m:Item>Apples</m:Item>
	//     </m:GetPrice>
	//   </soap:Body>
	// </soap:Envelope>
}
