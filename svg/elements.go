// Package svg maps SVG elements and attributes to go-mx the same way the html
// package maps HTML: every element is a function returning a *mx.Element and
// every attribute is a function returning an mx.Attrib.
//
// The html package also provides an html.Svg constructor, but it only creates a
// bare <svg> element for inline embedding in HTML; it does not offer the full
// SVG element and attribute vocabulary, the xmlns namespace handling, or the
// numeric attribute values found here. Use this package to build SVG content;
// reach for html.Svg only as a thin inline <svg> wrapper.
//
// Unlike HTML, SVG has no void elements, so every element constructor accepts
// attributes and children. Attribute constructors are generic over Value, so
// both strings and number literals can be passed directly, e.g. svg.CX(50).
package svg

import (
	"github.com/ungerik/go-mx"
)

// NS is the SVG XML namespace.
const NS = "http://www.w3.org/2000/svg"

// XLinkNS is the deprecated XLink XML namespace, still referenced by xlink:href.
const XLinkNS = "http://www.w3.org/1999/xlink"

// Element creates an SVG element with the passed attributes and children.
//
// Unlike HTML, SVG has no void elements: every SVG element may contain child
// elements such as <title>, <desc>, <metadata> or animation elements, so all
// element constructors in this package accept attributes and children.
func Element(name string, attribsChildren ...any) *mx.Element {
	return mx.NewElement(name, attribsChildren...)
}

// VoidElement creates a self-closing SVG element without children.
// SVG has no required void elements, but this helper is available for
// rendering leaf graphics like <rect/> in their compact self-closing form.
func VoidElement(name string, attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement(name, attribs...)
}

// Textf returns an escaped text node formatted with fmt.Sprintf semantics,
// for use as the content of a Text element.
func Textf(format string, args ...any) mx.Text {
	return mx.Textf(format, args...)
}

// Root returns an <svg> element with the SVG xmlns namespace attribute
// prepended, suitable as the root of a standalone SVG document.
func Root(attribsChildren ...any) *mx.Element {
	return Element("svg", append([]any{XMLNS}, attribsChildren...)...)
}

// See https://developer.mozilla.org/en-US/docs/Web/SVG/Element

// A maps the SVG <a> element.
func A(attribsChildren ...any) *mx.Element { return Element("a", attribsChildren...) }

// Animate maps the SVG <animate> element.
func Animate(attribsChildren ...any) *mx.Element { return Element("animate", attribsChildren...) }

// AnimateMotion maps the SVG <animateMotion> element.
func AnimateMotion(attribsChildren ...any) *mx.Element {
	return Element("animateMotion", attribsChildren...)
}

// AnimateTransform maps the SVG <animateTransform> element.
func AnimateTransform(attribsChildren ...any) *mx.Element {
	return Element("animateTransform", attribsChildren...)
}

// Circle maps the SVG <circle> element.
func Circle(attribsChildren ...any) *mx.Element { return Element("circle", attribsChildren...) }

// ClipPath maps the SVG <clipPath> element.
func ClipPath(attribsChildren ...any) *mx.Element {
	return Element("clipPath", attribsChildren...)
}

// Defs maps the SVG <defs> element.
func Defs(attribsChildren ...any) *mx.Element { return Element("defs", attribsChildren...) }

// Desc maps the SVG <desc> element.
func Desc(attribsChildren ...any) *mx.Element { return Element("desc", attribsChildren...) }

// Ellipse maps the SVG <ellipse> element.
func Ellipse(attribsChildren ...any) *mx.Element { return Element("ellipse", attribsChildren...) }

// FeBlend maps the SVG <feBlend> element.
func FeBlend(attribsChildren ...any) *mx.Element { return Element("feBlend", attribsChildren...) }

// FeColorMatrix maps the SVG <feColorMatrix> element.
func FeColorMatrix(attribsChildren ...any) *mx.Element {
	return Element("feColorMatrix", attribsChildren...)
}

// FeComponentTransfer maps the SVG <feComponentTransfer> element.
func FeComponentTransfer(attribsChildren ...any) *mx.Element {
	return Element("feComponentTransfer", attribsChildren...)
}

// FeComposite maps the SVG <feComposite> element.
func FeComposite(attribsChildren ...any) *mx.Element {
	return Element("feComposite", attribsChildren...)
}

// FeConvolveMatrix maps the SVG <feConvolveMatrix> element.
func FeConvolveMatrix(attribsChildren ...any) *mx.Element {
	return Element("feConvolveMatrix", attribsChildren...)
}

// FeDiffuseLighting maps the SVG <feDiffuseLighting> element.
func FeDiffuseLighting(attribsChildren ...any) *mx.Element {
	return Element("feDiffuseLighting", attribsChildren...)
}

// FeDisplacementMap maps the SVG <feDisplacementMap> element.
func FeDisplacementMap(attribsChildren ...any) *mx.Element {
	return Element("feDisplacementMap", attribsChildren...)
}

// FeDistantLight maps the SVG <feDistantLight> element.
func FeDistantLight(attribsChildren ...any) *mx.Element {
	return Element("feDistantLight", attribsChildren...)
}

// FeDropShadow maps the SVG <feDropShadow> element.
func FeDropShadow(attribsChildren ...any) *mx.Element {
	return Element("feDropShadow", attribsChildren...)
}

// FeFlood maps the SVG <feFlood> element.
func FeFlood(attribsChildren ...any) *mx.Element { return Element("feFlood", attribsChildren...) }

// FeFuncA maps the SVG <feFuncA> element.
func FeFuncA(attribsChildren ...any) *mx.Element { return Element("feFuncA", attribsChildren...) }

// FeFuncB maps the SVG <feFuncB> element.
func FeFuncB(attribsChildren ...any) *mx.Element { return Element("feFuncB", attribsChildren...) }

// FeFuncG maps the SVG <feFuncG> element.
func FeFuncG(attribsChildren ...any) *mx.Element { return Element("feFuncG", attribsChildren...) }

// FeFuncR maps the SVG <feFuncR> element.
func FeFuncR(attribsChildren ...any) *mx.Element { return Element("feFuncR", attribsChildren...) }

// FeGaussianBlur maps the SVG <feGaussianBlur> element.
func FeGaussianBlur(attribsChildren ...any) *mx.Element {
	return Element("feGaussianBlur", attribsChildren...)
}

// FeImage maps the SVG <feImage> element.
func FeImage(attribsChildren ...any) *mx.Element { return Element("feImage", attribsChildren...) }

// FeMerge maps the SVG <feMerge> element.
func FeMerge(attribsChildren ...any) *mx.Element { return Element("feMerge", attribsChildren...) }

// FeMergeNode maps the SVG <feMergeNode> element.
func FeMergeNode(attribsChildren ...any) *mx.Element {
	return Element("feMergeNode", attribsChildren...)
}

// FeMorphology maps the SVG <feMorphology> element.
func FeMorphology(attribsChildren ...any) *mx.Element {
	return Element("feMorphology", attribsChildren...)
}

// FeOffset maps the SVG <feOffset> element.
func FeOffset(attribsChildren ...any) *mx.Element {
	return Element("feOffset", attribsChildren...)
}

// FePointLight maps the SVG <fePointLight> element.
func FePointLight(attribsChildren ...any) *mx.Element {
	return Element("fePointLight", attribsChildren...)
}

// FeSpecularLighting maps the SVG <feSpecularLighting> element.
func FeSpecularLighting(attribsChildren ...any) *mx.Element {
	return Element("feSpecularLighting", attribsChildren...)
}

// FeSpotLight maps the SVG <feSpotLight> element.
func FeSpotLight(attribsChildren ...any) *mx.Element {
	return Element("feSpotLight", attribsChildren...)
}

// FeTile maps the SVG <feTile> element.
func FeTile(attribsChildren ...any) *mx.Element { return Element("feTile", attribsChildren...) }

// FeTurbulence maps the SVG <feTurbulence> element.
func FeTurbulence(attribsChildren ...any) *mx.Element {
	return Element("feTurbulence", attribsChildren...)
}

// Filter maps the SVG <filter> element.
func Filter(attribsChildren ...any) *mx.Element { return Element("filter", attribsChildren...) }

// ForeignObject maps the SVG <foreignObject> element.
func ForeignObject(attribsChildren ...any) *mx.Element {
	return Element("foreignObject", attribsChildren...)
}

// G maps the SVG <g> element.
func G(attribsChildren ...any) *mx.Element { return Element("g", attribsChildren...) }

// Image maps the SVG <image> element.
func Image(attribsChildren ...any) *mx.Element { return Element("image", attribsChildren...) }

// Line maps the SVG <line> element.
func Line(attribsChildren ...any) *mx.Element { return Element("line", attribsChildren...) }

// LinearGradient maps the SVG <linearGradient> element.
func LinearGradient(attribsChildren ...any) *mx.Element {
	return Element("linearGradient", attribsChildren...)
}

// Marker maps the SVG <marker> element.
func Marker(attribsChildren ...any) *mx.Element { return Element("marker", attribsChildren...) }

// Mask maps the SVG <mask> element.
func Mask(attribsChildren ...any) *mx.Element { return Element("mask", attribsChildren...) }

// Metadata maps the SVG <metadata> element.
func Metadata(attribsChildren ...any) *mx.Element { return Element("metadata", attribsChildren...) }

// MPath maps the SVG <mpath> element.
func MPath(attribsChildren ...any) *mx.Element { return Element("mpath", attribsChildren...) }

// Path maps the SVG <path> element.
func Path(attribsChildren ...any) *mx.Element { return Element("path", attribsChildren...) }

// Pattern maps the SVG <pattern> element.
func Pattern(attribsChildren ...any) *mx.Element { return Element("pattern", attribsChildren...) }

// Polygon maps the SVG <polygon> element.
func Polygon(attribsChildren ...any) *mx.Element { return Element("polygon", attribsChildren...) }

// Polyline maps the SVG <polyline> element.
func Polyline(attribsChildren ...any) *mx.Element { return Element("polyline", attribsChildren...) }

// RadialGradient maps the SVG <radialGradient> element.
func RadialGradient(attribsChildren ...any) *mx.Element {
	return Element("radialGradient", attribsChildren...)
}

// Rect maps the SVG <rect> element.
func Rect(attribsChildren ...any) *mx.Element { return Element("rect", attribsChildren...) }

// Script maps the SVG <script> element.
func Script(attribsChildren ...any) *mx.Element { return Element("script", attribsChildren...) }

// Set maps the SVG <set> element.
func Set(attribsChildren ...any) *mx.Element { return Element("set", attribsChildren...) }

// Stop maps the SVG <stop> element.
func Stop(attribsChildren ...any) *mx.Element { return Element("stop", attribsChildren...) }

// StyleElem returns a <style> element with the raw (unescaped) CSS as content.
func StyleElem(css string) *mx.Element { return Element("style", mx.Raw(css)) }

// SVG returns an <svg> element. Use Root for a standalone document that needs
// the xmlns namespace attribute.
func SVG(attribsChildren ...any) *mx.Element { return Element("svg", attribsChildren...) }

// Switch maps the SVG <switch> element.
func Switch(attribsChildren ...any) *mx.Element { return Element("switch", attribsChildren...) }

// Symbol maps the SVG <symbol> element.
func Symbol(attribsChildren ...any) *mx.Element { return Element("symbol", attribsChildren...) }

// Text maps the SVG <text> element.
func Text(attribsChildren ...any) *mx.Element { return Element("text", attribsChildren...) }

// TextPath maps the SVG <textPath> element.
func TextPath(attribsChildren ...any) *mx.Element { return Element("textPath", attribsChildren...) }

// Title maps the SVG <title> element.
func Title(attribsChildren ...any) *mx.Element { return Element("title", attribsChildren...) }

// TSpan maps the SVG <tspan> element.
func TSpan(attribsChildren ...any) *mx.Element { return Element("tspan", attribsChildren...) }

// Use maps the SVG <use> element.
func Use(attribsChildren ...any) *mx.Element { return Element("use", attribsChildren...) }

// View maps the SVG <view> element.
func View(attribsChildren ...any) *mx.Element { return Element("view", attribsChildren...) }

// Deprecated SVG elements (removed from or never standardized in SVG2) are
// intentionally not provided:
//
//	altGlyph, altGlyphDef, altGlyphItem, animateColor, color-profile, cursor,
//	font, font-face, font-face-format, font-face-name, font-face-src,
//	font-face-uri, glyph, glyphRef, hkern, missing-glyph, tref, vkern
