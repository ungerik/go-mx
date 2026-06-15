package xml

import (
	"fmt"
	"strconv"

	"github.com/ungerik/go-mx"
)

// Attribs is an alias for [mx.Attribs], a collection of XML attributes.
type Attribs = mx.Attribs

// AttribValue is the type set accepted by the generic [Attrib] constructor: a
// string or any integer or floating-point type (not uintptr or complex).
// Strings pass through unchanged; integers are formatted as plain decimals and
// floats with strconv.FormatFloat so small or large magnitudes render as plain
// decimals instead of scientific notation.
type AttribValue interface {
	~string |
		~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

// attribValueString formats an attribute value as a string. Strings pass
// through and integer types go through fmt.Sprint, but floats use
// strconv.FormatFloat with the 'f' format so small or large magnitudes render
// as plain decimals instead of fmt's scientific notation (e.g. 0.00005 not
// "5e-05").
func attribValueString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		return fmt.Sprint(value)
	}
}

// Attrib constructs an XML attribute name="value". The value may be a string or
// any number type (see [AttribValue]): strings pass through unchanged and float
// values are formatted as plain decimals (never scientific notation). Since XML
// has no fixed attribute vocabulary this is the general attribute constructor;
// the xml: namespace attributes ([XMLNS], [XMLLang], …) are the only named ones.
func Attrib[T AttribValue](name string, value T) mx.Attribute {
	return mx.Attribute{Name: name, Value: attribValueString(value)}
}

// AttribNS constructs a namespace-prefixed XML attribute prefix:name="value".
// It is the attribute counterpart of [ElementNS]; an empty prefix falls back to
// the unqualified name.
func AttribNS[T AttribValue](prefix, name string, value T) mx.Attribute {
	return mx.Attribute{Name: qualify(prefix, name), Value: attribValueString(value)}
}

// The xml: namespace and xmlns attributes are predefined by the XML
// specification and have the same meaning in every document, so unlike generic
// element and attribute names they get dedicated constructors here.

// XMLNS sets the default-namespace declaration xmlns="uri", binding the URI as
// the default namespace for the element and its unprefixed descendants.
func XMLNS(uri string) mx.Attrib { return mx.NewAttrib("xmlns", uri) }

// XMLNSPrefix sets a namespace-prefix declaration xmlns:prefix="uri", binding
// the prefix used by [ElementNS]/[AttribNS] to the given namespace URI.
func XMLNSPrefix(prefix, uri string) mx.Attrib {
	return mx.NewAttrib("xmlns:"+prefix, uri)
}

// XMLLang sets the xml:lang attribute, the natural language of the element's
// content and attribute values (an RFC 5646 / BCP 47 language tag like "en-US").
func XMLLang(lang string) mx.Attrib { return mx.NewAttrib("xml:lang", lang) }

// XMLBase sets the xml:base attribute, the base URI for resolving relative
// references within the element.
func XMLBase(uri string) mx.Attrib { return mx.NewAttrib("xml:base", uri) }

// XMLID sets the xml:id attribute, an element identifier of type ID independent
// of any DTD or schema.
func XMLID(id string) mx.Attrib { return mx.NewAttrib("xml:id", id) }

// XMLSpace sets the xml:space attribute, signalling how whitespace should be
// handled. The defined values are "default" and "preserve"; use the
// [XMLSpaceDefault] and [XMLSpacePreserve] constants for those.
func XMLSpace(value string) mx.Attrib { return mx.NewAttrib("xml:space", value) }

const (
	// XMLSpaceDefault is the xml:space="default" attribute, letting the
	// application apply its default whitespace handling.
	XMLSpaceDefault mx.ConstAttrib = "xml:space=default"
	// XMLSpacePreserve is the xml:space="preserve" attribute, requesting that
	// whitespace be preserved.
	XMLSpacePreserve mx.ConstAttrib = "xml:space=preserve"
)
