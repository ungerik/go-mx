// Package xml builds XML markup with the same component model as the html and
// svg packages: an element is created with [Element], attributes with [Attrib],
// and the tree is written out by Render.
//
// Unlike html and svg, XML has no fixed vocabulary of element or attribute
// names, so this package provides generic constructors rather than one function
// per tag: [Element] and [EmptyElement] take any name, [ElementNS] and
// [EmptyElementNS] take a namespace prefix and a name, and [Attrib] takes any
// name and value. On top of those it adds the constructs that are specific to
// XML documents:
//
//   - [Comment]      — <!-- ... -->
//   - [CDATA]        — <![CDATA[ ... ]]>
//   - [ProcInst]     — a processing instruction <?target data?>
//   - [Declaration]  — the <?xml ...?> declaration (and [Decl] for a custom one)
//   - [Doctype]      — a <!DOCTYPE ...> document type declaration
//   - the predefined xml: namespace attributes ([XMLNS], [XMLLang], …)
//
// A complete document — declaration, optional prolog and a root element — can be
// assembled with [Document].
//
//	xml.Document{
//	    Root: xml.Element("note",
//	        xml.Attrib("id", 42),
//	        xml.Element("to", "Tove"),
//	        xml.Comment("the message body"),
//	        xml.Element("body", xml.CDATA("unescaped <raw> & text")),
//	    ),
//	}
//
// Element and attribute names are passed through as written: they are not
// checked for XML name syntax or schema validity. The package does handle the
// content-dependent mechanics of well-formed output — escaping text and
// attribute values, balancing tags, and enforcing the comment, CDATA, and
// processing-instruction constraints.
package xml

import (
	"github.com/ungerik/go-mx"
)

// Element creates a generic XML element with the given name, taking attributes
// and children as variadic arguments. An element with no children renders with
// an explicit close tag (<name></name>); use [EmptyElement] for the
// self-closing form (<name/>).
func Element(name string, attribsChildren ...any) *mx.Element {
	return mx.NewElement(name, attribsChildren...)
}

// EmptyElement creates a generic empty XML element with the given name, taking
// only attributes and no children. It renders in the self-closing form
// <name attr="..."/>.
func EmptyElement(name string, attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement(name, attribs...)
}

// ElementNS creates a namespace-prefixed XML element with the qualified name
// prefix:name, taking attributes and children as variadic arguments. It is a
// shortcut for Element(prefix+":"+name, …); for example
// ElementNS("soap", "Envelope", …) builds <soap:Envelope>…</soap:Envelope>.
// An empty prefix falls back to the unqualified name.
//
// ElementNS only writes the qualified name; bind the prefix to a namespace URI
// with an [XMLNSPrefix] attribute (usually on the root element).
func ElementNS(prefix, name string, attribsChildren ...any) *mx.Element {
	return mx.NewElement(qualify(prefix, name), attribsChildren...)
}

// EmptyElementNS creates an empty namespace-prefixed XML element with the
// qualified name prefix:name, taking only attributes and no children. It is the
// self-closing counterpart of [ElementNS], rendering <prefix:name attr="..."/>.
func EmptyElementNS(prefix, name string, attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement(qualify(prefix, name), attribs...)
}

// qualify joins a namespace prefix and a local name into an XML qualified name
// prefix:name, returning just name when prefix is empty.
func qualify(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + ":" + name
}

// Textf returns escaped [Text] formatted like fmt.Sprintf.
func Textf(format string, args ...any) mx.Text {
	return mx.Textf(format, args...)
}
