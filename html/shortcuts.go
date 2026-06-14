package html

// This file holds convenience constructors that bake a conventional attribute
// into a common element, the same pattern as InputType*, the *Button family and
// the OL* family in elements.go. They cut the boilerplate of writing the same
// element + attribute pairing (meta charset, script src, stylesheet link, …) at
// every call site.

import (
	"github.com/ungerik/go-mx"
)

// MetaCharset is a <meta charset="..."> declaring the document character
// encoding, e.g. MetaCharset("UTF-8").
func MetaCharset(charset string) *mx.Element {
	return Meta(CharSet(charset))
}

// MetaCharsetUTF8 is the <meta charset="UTF-8"> declaration for the common
// UTF-8 case as a ready-to-use constant, the static equivalent of
// MetaCharset("UTF-8"). It can be dropped straight into a document head.
const MetaCharsetUTF8 Raw = /*html*/ `<meta charset="UTF-8">`

// MetaName is a <meta name="..." content="..."> document-level metadata pair,
// e.g. MetaName("description", "…").
func MetaName(name, content string) *mx.Element {
	return Meta(Name(name), ContentAttr(content))
}

// MetaProperty is a <meta property="..." content="..."> pair, used by metadata
// vocabularies such as Open Graph, e.g. MetaProperty("og:title", "…").
func MetaProperty(property, content string) *mx.Element {
	return Meta(Attrib("property", property), ContentAttr(content))
}

// MetaViewport is a <meta name="viewport" content="..."> controlling the layout
// viewport on mobile browsers. The usual responsive value is
// "width=device-width, initial-scale=1".
func MetaViewport(content string) *mx.Element {
	return MetaName("viewport", content)
}

// ScriptSrc is a <script src="..."> loading an external classic script. Extra
// attributes or children may follow, e.g. ScriptSrc(url, Defer) or
// ScriptSrc(url, Async, CrossOrigin("anonymous")).
func ScriptSrc(url string, attribsChildren ...any) *mx.Element {
	return Script(append([]any{Src(url)}, attribsChildren...)...)
}

// ScriptModule is a <script type="module" ...> ES module script. Combine with
// Src for an external module: ScriptModule(Src(url)).
func ScriptModule(attribsChildren ...any) *mx.Element {
	return Script(append([]any{Type("module")}, attribsChildren...)...)
}

// StyleSheet is a <link rel="stylesheet" href="..."> loading an external CSS
// stylesheet. Extra attributes (Media, Integrity, …) may follow.
func StyleSheet(href string, attribs ...mx.Attrib) *mx.Element {
	return Link(append([]mx.Attrib{Rel("stylesheet"), HRef(href)}, attribs...)...)
}

// Icon is a <link rel="icon" href="..."> declaring the page favicon. Extra
// attributes (Sizes, Type, …) may follow.
func Icon(href string, attribs ...mx.Attrib) *mx.Element {
	return Link(append([]mx.Attrib{Rel("icon"), HRef(href)}, attribs...)...)
}

// LinkPreload is a <link rel="preload" href="..." as="..."> preloading a
// resource. The as argument is the typed As destination enum (AsScript,
// AsStyle, AsFont, …). It is named LinkPreload, not Preload, because Preload is
// the media-element preload attribute enum.
func LinkPreload(href string, as As, attribs ...mx.Attrib) *mx.Element {
	return Link(append([]mx.Attrib{Rel("preload"), HRef(href), as}, attribs...)...)
}

// BlankLink is an <a href="..." target="_blank" rel="noopener noreferrer"> that
// opens in a new browsing context. The rel value prevents reverse tabnabbing:
// without rel="noopener" the opened page can navigate this one via
// window.opener. Extra attributes may follow the text.
func BlankLink(href, text string, attribs ...mx.Attrib) *mx.Element {
	return A(HRef(href), TargetBlank, Rel("noopener", "noreferrer"), attribs, text)
}
