package html

// This file holds convenience constructors that bake a conventional attribute
// into a common element, the same pattern as InputType*, the *Button family and
// the OL* family in elements.go. They cut the boilerplate of writing the same
// element + attribute pairing (meta charset, script src, stylesheet link, …) at
// every call site.

import (
	"strings"

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

// ScriptJS is a <script> element wrapping the given raw JavaScript as its
// content, the inline counterpart to ScriptSrc,
// e.g. ScriptJS(`console.log("hi")`). Multiple arguments are joined with
// semicolons, so several statements can be passed as separate strings:
// ScriptJS(`const x = 1`, `console.log(x)`).
//
// The js is emitted verbatim as [mx.Raw] with no escaping — a script body
// cannot be HTML-escaped without changing what it executes. Pass only trusted,
// developer-controlled source. Never interpolate untrusted input: a "</script>"
// sequence in the content ends the element and everything after it is parsed as
// HTML, a cross-site scripting vector that escaping cannot defend against here.
func ScriptJS(js ...string) *mx.Element {
	return &mx.Element{
		Name:     "script",
		Children: []mx.Component{mx.Raw(strings.Join(js, ";"))},
	}
}

// ScriptSrc is a <script src="..."> loading an external classic script. Extra
// attributes may follow, e.g. ScriptSrc(url, Defer) or
// ScriptSrc(url, Async, CrossOrigin("anonymous")).
func ScriptSrc(url string, attribs ...mx.Attrib) *mx.Element {
	return &mx.Element{
		Name:     "script",
		Attribs:  append([]mx.Attrib{Src(url)}, attribs...),
		Children: []mx.Component{}, // not a void element: needs </script>, so non-nil empty children
	}
}

// ScriptModule is a <script type="module" ...> ES module script. Combine with
// Src for an external module: ScriptModule(Src(url)).
func ScriptModule(attribsChildren ...any) *mx.Element {
	return Script(append([]any{Type("module")}, attribsChildren...)...)
}

// ScriptModuleJS is a <script type="module"> wrapping the given raw ES module
// source as its content, the inline counterpart to ScriptModule(Src(url)),
// e.g. ScriptModuleJS(`import {x} from "./m.js"; x()`). Multiple arguments are
// joined with semicolons, so several statements can be passed as separate
// strings: ScriptModuleJS(`import {x} from "./m.js"`, `x()`).
//
// Like [ScriptJS], the source is emitted verbatim as [mx.Raw] with no escaping.
// Pass only trusted, developer-controlled source. Never interpolate untrusted
// input: a "</script>" sequence in the content ends the element and everything
// after it is parsed as HTML, a cross-site scripting vector that escaping cannot
// defend against here.
func ScriptModuleJS(js ...string) *mx.Element {
	return &mx.Element{
		Name:     "script",
		Attribs:  []mx.Attrib{Type("module")},
		Children: []mx.Component{mx.Raw(strings.Join(js, ";"))},
	}
}

// StyleSheet is a <link rel="stylesheet" href="..."> loading an external CSS
// stylesheet. Extra attributes (Media, Integrity, …) may follow.
func StyleSheet(href string, attribs ...mx.Attrib) *mx.Element {
	return &mx.Element{
		Name:    "link",
		Attribs: append([]mx.Attrib{Rel("stylesheet"), HRef(href)}, attribs...),
	}
}

// Icon is a <link rel="icon" href="..."> declaring the page favicon. Extra
// attributes (Sizes, Type, …) may follow.
func Icon(href string, attribs ...mx.Attrib) *mx.Element {
	return &mx.Element{
		Name:    "link",
		Attribs: append([]mx.Attrib{Rel("icon"), HRef(href)}, attribs...),
	}
}

// LinkPreload is a <link rel="preload" href="..." as="..."> preloading a
// resource. The as argument is the typed As destination enum (AsScript,
// AsStyle, AsFont, …). It is named LinkPreload, not Preload, because Preload is
// the media-element preload attribute enum.
func LinkPreload(href string, as As, attribs ...mx.Attrib) *mx.Element {
	return Link(append([]mx.Attrib{Rel("preload"), HRef(href), as}, attribs...)...)
}

// AHRef creates an <a> hyperlink to the given URL
// as a shortcut for A(HRef(url), attribsChildren...).
func AHRef(url string, attribsChildren ...any) *mx.Element {
	return A(append([]any{HRef(url)}, attribsChildren...)...)
}
