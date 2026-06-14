package html

import (
	"github.com/ungerik/go-mx"
)

func Element(name string, attribsChildren ...any) *mx.Element {
	return mx.NewElement(name, attribsChildren...)
}

func VoidElement(name string, attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement(name, attribs...)
}

func Textf(format string, args ...any) Text {
	return mx.Textf(format, args...)
}

func Hyperlink(href, text string, attribs ...mx.Attrib) *mx.Element {
	return A(HRef(href), attribs, text)
}

// See https://github.com/jozo/all-html-elements-and-attributes

func A(attribsChildren ...any) *mx.Element       { return Element("a", attribsChildren...) }
func Abbr(attribsChildren ...any) *mx.Element    { return Element("abbr", attribsChildren...) }
func Address(attribsChildren ...any) *mx.Element { return Element("address", attribsChildren...) }
func Area(attribs ...mx.Attrib) *mx.Element      { return VoidElement("area", attribs...) }
func Article(attribsChildren ...any) *mx.Element { return Element("article", attribsChildren...) }
func Aside(attribsChildren ...any) *mx.Element   { return Element("aside", attribsChildren...) }
func Audio(attribsChildren ...any) *mx.Element   { return Element("audio", attribsChildren...) }

// Use `Strong` for semantic emphasis
func B(attribsChildren ...any) *mx.Element   { return Element("b", attribsChildren...) }
func Base(attribs ...mx.Attrib) *mx.Element  { return VoidElement("base", attribs...) }
func BDi(attribsChildren ...any) *mx.Element { return Element("bdi", attribsChildren...) }
func BDo(attribsChildren ...any) *mx.Element { return Element("bdo", attribsChildren...) }
func Blockquote(attribsChildren ...any) *mx.Element {
	return mx.NewElement("blockquote", attribsChildren...)
}
func Body(attribsChildren ...any) *mx.Element    { return Element("body", attribsChildren...) }
func Br(attribs ...mx.Attrib) *mx.Element        { return VoidElement("br", attribs...) }
func Button(attribsChildren ...any) *mx.Element  { return Element("button", attribsChildren...) }
func Canvas(attribsChildren ...any) *mx.Element  { return Element("canvas", attribsChildren...) }
func Caption(attribsChildren ...any) *mx.Element { return Element("caption", attribsChildren...) }
func Cite(attribsChildren ...any) *mx.Element    { return Element("cite", attribsChildren...) }
func Code(attribsChildren ...any) *mx.Element    { return Element("code", attribsChildren...) }
func Col(attribs ...mx.Attrib) *mx.Element       { return VoidElement("col", attribs...) }
func ColGroup(attribsChildren ...any) *mx.Element {
	return mx.NewElement("colgroup", attribsChildren...)
}

// Command was removed from HTML spec
func Content(attribsChildren ...any) *mx.Element { return Element("content", attribsChildren...) }
func Data(attribsChildren ...any) *mx.Element    { return Element("data", attribsChildren...) }
func DataList(attribsChildren ...any) *mx.Element {
	return mx.NewElement("datalist", attribsChildren...)
}
func DD(attribsChildren ...any) *mx.Element      { return Element("dd", attribsChildren...) }
func Del(attribsChildren ...any) *mx.Element     { return Element("del", attribsChildren...) }
func Details(attribsChildren ...any) *mx.Element { return Element("details", attribsChildren...) }
func Dfn(attribsChildren ...any) *mx.Element     { return Element("dfn", attribsChildren...) }
func Dialog(attribsChildren ...any) *mx.Element  { return Element("dialog", attribsChildren...) }
func Div(attribsChildren ...any) *mx.Element     { return Element("div", attribsChildren...) }
func DL(attribsChildren ...any) *mx.Element      { return Element("dl", attribsChildren...) }
func DT(attribsChildren ...any) *mx.Element      { return Element("dt", attribsChildren...) }
func Em(attribsChildren ...any) *mx.Element      { return Element("em", attribsChildren...) }

// Use `Object` or `IFrame` for better browser compatibility
func Embed(attribs ...mx.Attrib) *mx.Element { return VoidElement("embed", attribs...) }
func FieldSet(attribsChildren ...any) *mx.Element {
	return mx.NewElement("fieldset", attribsChildren...)
}
func FigCaption(attribsChildren ...any) *mx.Element {
	return mx.NewElement("figcaption", attribsChildren...)
}
func Figure(attribsChildren ...any) *mx.Element { return Element("figure", attribsChildren...) }
func Footer(attribsChildren ...any) *mx.Element { return Element("footer", attribsChildren...) }
func Form(attribsChildren ...any) *mx.Element   { return Element("form", attribsChildren...) }
func H1(attribsChildren ...any) *mx.Element     { return Element("h1", attribsChildren...) }
func H2(attribsChildren ...any) *mx.Element     { return Element("h2", attribsChildren...) }
func H3(attribsChildren ...any) *mx.Element     { return Element("h3", attribsChildren...) }
func H4(attribsChildren ...any) *mx.Element     { return Element("h4", attribsChildren...) }
func H5(attribsChildren ...any) *mx.Element     { return Element("h5", attribsChildren...) }
func H6(attribsChildren ...any) *mx.Element     { return Element("h6", attribsChildren...) }
func Head(attribsChildren ...any) *mx.Element   { return Element("head", attribsChildren...) }
func Header(attribsChildren ...any) *mx.Element { return Element("header", attribsChildren...) }
func HGroup(attribsChildren ...any) *mx.Element { return Element("hgroup", attribsChildren...) }
func HR(attribs ...mx.Attrib) *mx.Element       { return VoidElement("hr", attribs...) }
func HTML(attribsChildren ...any) *mx.Element   { return Element("html", attribsChildren...) }

// Use `Em` for semantic emphasis
func I(attribsChildren ...any) *mx.Element      { return Element("i", attribsChildren...) }
func IFrame(attribsChildren ...any) *mx.Element { return Element("iframe", attribsChildren...) }

// Image is SVG only, use Img for HTML
func Img(attribs ...mx.Attrib) *mx.Element   { return VoidElement("img", attribs...) }
func Input(attribs ...mx.Attrib) *mx.Element { return VoidElement("input", attribs...) }
func InputTypeButton(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "button", attribs)...)
}
func InputTypeCheckbox(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "checkbox", attribs)...)
}
func InputTypeColor(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "color", attribs)...)
}
func InputTypeDate(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "date", attribs)...)
}
func InputTypeDatetimeLocal(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "datetime-local", attribs)...)
}
func InputTypeEmail(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "email", attribs)...)
}
func InputTypeFile(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "file", attribs)...)
}
func InputTypeHidden(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "hidden", attribs)...)
}
func InputTypeImage(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "image", attribs)...)
}
func InputTypeMonth(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "month", attribs)...)
}
func InputTypeNumber(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "number", attribs)...)
}
func InputTypePassword(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "password", attribs)...)
}
func InputTypeRadio(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "radio", attribs)...)
}
func InputTypeRange(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "range", attribs)...)
}
func InputTypeReset(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "reset", attribs)...)
}
func InputTypeSearch(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "search", attribs)...)
}
func InputTypeSubmit(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "submit", attribs)...)
}
func InputTypeTel(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "tel", attribs)...)
}
func InputTypeText(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "text", attribs)...)
}
func InputTypeTime(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "time", attribs)...)
}
func InputTypeURL(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "url", attribs)...)
}
func InputTypeWeek(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "week", attribs)...)
}
func Ins(attribsChildren ...any) *mx.Element     { return Element("ins", attribsChildren...) }
func Kbd(attribsChildren ...any) *mx.Element     { return Element("kbd", attribsChildren...) }
func Label(attribsChildren ...any) *mx.Element   { return Element("label", attribsChildren...) }
func Legend(attribsChildren ...any) *mx.Element  { return Element("legend", attribsChildren...) }
func LI(attribsChildren ...any) *mx.Element      { return Element("li", attribsChildren...) }
func Link(attribs ...mx.Attrib) *mx.Element      { return VoidElement("link", attribs...) }
func Main(attribsChildren ...any) *mx.Element    { return Element("main", attribsChildren...) }
func Map(attribsChildren ...any) *mx.Element     { return Element("map", attribsChildren...) }
func Mark(attribsChildren ...any) *mx.Element    { return Element("mark", attribsChildren...) }
func Math(attribsChildren ...any) *mx.Element    { return Element("math", attribsChildren...) }
func Meta(attribs ...mx.Attrib) *mx.Element      { return VoidElement("meta", attribs...) }
func Menu(attribsChildren ...any) *mx.Element    { return Element("menu", attribsChildren...) }
func Meter(attribsChildren ...any) *mx.Element   { return Element("meter", attribsChildren...) }
func Nav(attribsChildren ...any) *mx.Element     { return Element("nav", attribsChildren...) }
func NoEmbed(attribsChildren ...any) *mx.Element { return Element("noembed", attribsChildren...) }
func NoScript(attribsChildren ...any) *mx.Element {
	return mx.NewElement("noscript", attribsChildren...)
}
func Object(attribsChildren ...any) *mx.Element { return Element("object", attribsChildren...) }
func OL(attribsChildren ...any) *mx.Element     { return Element("ol", attribsChildren...) }
func OptGroup(attribsChildren ...any) *mx.Element {
	return mx.NewElement("optgroup", attribsChildren...)
}
func Option(attribsChildren ...any) *mx.Element  { return Element("option", attribsChildren...) }
func Output(attribsChildren ...any) *mx.Element  { return Element("output", attribsChildren...) }
func P(attribsChildren ...any) *mx.Element       { return Element("p", attribsChildren...) }
func Picture(attribsChildren ...any) *mx.Element { return Element("picture", attribsChildren...) }
func Portal(attribsChildren ...any) *mx.Element  { return Element("portal", attribsChildren...) }
func Pre(attribsChildren ...any) *mx.Element     { return Element("pre", attribsChildren...) }
func Progress(attribsChildren ...any) *mx.Element {
	return mx.NewElement("progress", attribsChildren...)
}
func Q(attribsChildren ...any) *mx.Element       { return Element("q", attribsChildren...) }
func RB(attribsChildren ...any) *mx.Element      { return Element("rb", attribsChildren...) }
func RP(attribsChildren ...any) *mx.Element      { return Element("rp", attribsChildren...) }
func RT(attribsChildren ...any) *mx.Element      { return Element("rt", attribsChildren...) }
func RTC(attribsChildren ...any) *mx.Element     { return Element("rtc", attribsChildren...) }
func Ruby(attribsChildren ...any) *mx.Element    { return Element("ruby", attribsChildren...) }
func S(attribsChildren ...any) *mx.Element       { return Element("s", attribsChildren...) }
func Samp(attribsChildren ...any) *mx.Element    { return Element("samp", attribsChildren...) }
func Script(attribsChildren ...any) *mx.Element  { return Element("script", attribsChildren...) }
func Search(attribsChildren ...any) *mx.Element  { return Element("search", attribsChildren...) }
func Section(attribsChildren ...any) *mx.Element { return Element("section", attribsChildren...) }
func Select(attribsChildren ...any) *mx.Element  { return Element("select", attribsChildren...) }
func Shadow(attribsChildren ...any) *mx.Element  { return Element("shadow", attribsChildren...) }
func Slot(attribsChildren ...any) *mx.Element    { return Element("slot", attribsChildren...) }

// Small is still valid in HTML5 but use CSS for better control.
func Small(attribsChildren ...any) *mx.Element   { return Element("small", attribsChildren...) }
func Source(attribs ...mx.Attrib) *mx.Element    { return VoidElement("source", attribs...) }
func Span(attribsChildren ...any) *mx.Element    { return Element("span", attribsChildren...) }
func Strong(attribsChildren ...any) *mx.Element  { return Element("strong", attribsChildren...) }
func StyleElem(css string) *mx.Element           { return Element("style", Raw(css)) }
func Sub(attribsChildren ...any) *mx.Element     { return Element("sub", attribsChildren...) }
func Summary(attribsChildren ...any) *mx.Element { return Element("summary", attribsChildren...) }
func Sup(attribsChildren ...any) *mx.Element     { return Element("sup", attribsChildren...) }

// Svg creates a bare <svg> element for inline embedding in HTML.
// For the full SVG element and attribute vocabulary, namespace handling, and
// numeric attribute values use the svg package (github.com/ungerik/go-mx/svg).
func Svg(attribsChildren ...any) *mx.Element   { return Element("svg", attribsChildren...) }
func Table(attribsChildren ...any) *mx.Element { return Element("table", attribsChildren...) }
func TBody(attribsChildren ...any) *mx.Element { return Element("tbody", attribsChildren...) }
func TD(attribsChildren ...any) *mx.Element    { return Element("td", attribsChildren...) }
func TemplateElem(attribsChildren ...any) *mx.Element {
	return mx.NewElement("template", attribsChildren...)
}
func TextArea(attribsChildren ...any) *mx.Element {
	return mx.NewElement("textarea", attribsChildren...)
}
func TFoot(attribsChildren ...any) *mx.Element     { return Element("tfoot", attribsChildren...) }
func TH(attribsChildren ...any) *mx.Element        { return Element("th", attribsChildren...) }
func THead(attribsChildren ...any) *mx.Element     { return Element("thead", attribsChildren...) }
func Time(attribsChildren ...any) *mx.Element      { return Element("time", attribsChildren...) }
func TitleElem(attribsChildren ...any) *mx.Element { return Element("title", attribsChildren...) }
func TR(attribsChildren ...any) *mx.Element        { return Element("tr", attribsChildren...) }
func Track(attribs ...mx.Attrib) *mx.Element       { return VoidElement("track", attribs...) }
func UL(attribsChildren ...any) *mx.Element        { return Element("ul", attribsChildren...) }
func Var(attribsChildren ...any) *mx.Element       { return Element("var", attribsChildren...) }
func Video(attribsChildren ...any) *mx.Element     { return Element("video", attribsChildren...) }
func WBr(attribs ...mx.Attrib) *mx.Element         { return VoidElement("wbr", attribs...) }

// Deprecated in HTML5:
//
// func Acronym(attribsChildren ...any) *mx.Element { return Element("acronym", attribsChildren...) }
// func Big(attribsChildren ...any) *mx.Element     { return Element("big", attribsChildren...) }
// func Center(attribsChildren ...any) *mx.Element  { return Element("center", attribsChildren...) }
// func Font(attribsChildren ...any) *mx.Element    { return Element("font", attribsChildren...) }
// func Frame(attribsChildren ...any) *mx.Element   { return Element("frame", attribsChildren...) }
// func FrameSet(attribsChildren ...any) *mx.Element {
// 	return mx.NewElement("frameset", attribsChildren...)
// }
// func NoFrames(attribsChildren ...any) *mx.Element {
// 	return mx.NewElement("noframes", attribsChildren...)
// }
// func Marquee(attribsChildren ...any) *mx.Element { return Element("marquee", attribsChildren...) }
// Menu was reinstated in HTML5 Living Standard, see active definition above
// func MenuItem(attribsChildren ...any) *mx.Element {
// 	return mx.NewElement("menuitem", attribsChildren...)
// }
// func NoBr(attribsChildren ...any) *mx.Element { return Element("nobr", attribsChildren...) }
// func Plaintext(attribsChildren ...any) *mx.Element {
// 	return mx.NewElement("plaintext", attribsChildren...)
// }
// func Strike(attribsChildren ...any) *mx.Element { return Element("strike", attribsChildren...) }
// func TT(attribsChildren ...any) *mx.Element     { return Element("tt", attribsChildren...) }
// func U(attribsChildren ...any) *mx.Element      { return Element("u", attribsChildren...) }
