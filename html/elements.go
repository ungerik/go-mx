package html

import (
	"github.com/ungerik/go-mx"
)

type (
	Element  = mx.Element
	Text     = mx.Text
	Raw      = mx.Raw
	RawBytes = mx.RawBytes
)

func NewElement(name string, attribsChildren ...any) *Element {
	return mx.NewElement(name, attribsChildren...)
}

func NewVoidElement(name string, attribs ...mx.Attrib) *Element {
	return mx.NewVoidElement(name, attribs...)
}

func Textf(format string, args ...any) Text {
	return mx.Textf(format, args...)
}

// See https://github.com/jozo/all-html-elements-and-attributes

func A(attribsChildren ...any) *Element       { return mx.NewElement("a", attribsChildren...) }
func Abbr(attribsChildren ...any) *Element    { return mx.NewElement("abbr", attribsChildren...) }
func Acronym(attribsChildren ...any) *Element { return mx.NewElement("acronym", attribsChildren...) }
func Address(attribsChildren ...any) *Element { return mx.NewElement("address", attribsChildren...) }
func Area(attribs ...Attrib) *Element         { return mx.NewVoidElement("area", attribs...) }
func Article(attribsChildren ...any) *Element { return mx.NewElement("article", attribsChildren...) }
func Aside(attribsChildren ...any) *Element   { return mx.NewElement("aside", attribsChildren...) }
func Audio(attribsChildren ...any) *Element   { return mx.NewElement("audio", attribsChildren...) }
func B(attribsChildren ...any) *Element       { return mx.NewElement("b", attribsChildren...) }
func Base(attribs ...Attrib) *Element         { return mx.NewVoidElement("base", attribs...) }
func BDI(attribsChildren ...any) *Element     { return mx.NewElement("bdi", attribsChildren...) }
func BDO(attribsChildren ...any) *Element     { return mx.NewElement("bdo", attribsChildren...) }
func Big(attribsChildren ...any) *Element     { return mx.NewElement("big", attribsChildren...) }
func Blockquote(attribsChildren ...any) *Element {
	return mx.NewElement("blockquote", attribsChildren...)
}
func Body(attribsChildren ...any) *Element     { return mx.NewElement("body", attribsChildren...) }
func Br(attribs ...Attrib) *Element            { return mx.NewVoidElement("br", attribs...) }
func Button(attribsChildren ...any) *Element   { return mx.NewElement("button", attribsChildren...) }
func Canvas(attribsChildren ...any) *Element   { return mx.NewElement("canvas", attribsChildren...) }
func Caption(attribsChildren ...any) *Element  { return mx.NewElement("caption", attribsChildren...) }
func Center(attribsChildren ...any) *Element   { return mx.NewElement("center", attribsChildren...) }
func Cite(attribsChildren ...any) *Element     { return mx.NewElement("cite", attribsChildren...) }
func Code(attribsChildren ...any) *Element     { return mx.NewElement("code", attribsChildren...) }
func Col(attribs ...Attrib) *Element           { return mx.NewVoidElement("col", attribs...) }
func ColGroup(attribsChildren ...any) *Element { return mx.NewElement("colgroup", attribsChildren...) }
func Content(attribsChildren ...any) *Element  { return mx.NewElement("content", attribsChildren...) }
func Data(attribsChildren ...any) *Element     { return mx.NewElement("data", attribsChildren...) }
func DataList(attribsChildren ...any) *Element { return mx.NewElement("datalist", attribsChildren...) }
func DD(attribsChildren ...any) *Element       { return mx.NewElement("dd", attribsChildren...) }
func Del(attribsChildren ...any) *Element      { return mx.NewElement("del", attribsChildren...) }
func Details(attribsChildren ...any) *Element  { return mx.NewElement("details", attribsChildren...) }
func Dfn(attribsChildren ...any) *Element      { return mx.NewElement("dfn", attribsChildren...) }
func Dialog(attribsChildren ...any) *Element   { return mx.NewElement("dialog", attribsChildren...) }
func Div(attribsChildren ...any) *Element      { return mx.NewElement("div", attribsChildren...) }
func DL(attribsChildren ...any) *Element       { return mx.NewElement("dl", attribsChildren...) }
func DT(attribsChildren ...any) *Element       { return mx.NewElement("dt", attribsChildren...) }
func Em(attribsChildren ...any) *Element       { return mx.NewElement("em", attribsChildren...) }
func Embed(attribs ...Attrib) *Element         { return mx.NewVoidElement("embed", attribs...) }
func FieldSet(attribsChildren ...any) *Element { return mx.NewElement("fieldset", attribsChildren...) }
func FigCaption(attribsChildren ...any) *Element {
	return mx.NewElement("figcaption", attribsChildren...)
}
func Figure(attribsChildren ...any) *Element   { return mx.NewElement("figure", attribsChildren...) }
func Font(attribsChildren ...any) *Element     { return mx.NewElement("font", attribsChildren...) }
func Footer(attribsChildren ...any) *Element   { return mx.NewElement("footer", attribsChildren...) }
func Form(attribsChildren ...any) *Element     { return mx.NewElement("form", attribsChildren...) }
func Frame(attribsChildren ...any) *Element    { return mx.NewElement("frame", attribsChildren...) }
func FrameSet(attribsChildren ...any) *Element { return mx.NewElement("frameset", attribsChildren...) }
func H1(attribsChildren ...any) *Element       { return mx.NewElement("h1", attribsChildren...) }
func H2(attribsChildren ...any) *Element       { return mx.NewElement("h2", attribsChildren...) }
func H3(attribsChildren ...any) *Element       { return mx.NewElement("h3", attribsChildren...) }
func H4(attribsChildren ...any) *Element       { return mx.NewElement("h4", attribsChildren...) }
func H5(attribsChildren ...any) *Element       { return mx.NewElement("h5", attribsChildren...) }
func H6(attribsChildren ...any) *Element       { return mx.NewElement("h6", attribsChildren...) }
func Head(attribsChildren ...any) *Element     { return mx.NewElement("head", attribsChildren...) }
func Header(attribsChildren ...any) *Element   { return mx.NewElement("header", attribsChildren...) }
func Hgroup(attribsChildren ...any) *Element   { return mx.NewElement("hgroup", attribsChildren...) }
func HR(attribs ...Attrib) *Element            { return mx.NewVoidElement("hr", attribs...) }
func HTML(attribsChildren ...any) *Element     { return mx.NewElement("html", attribsChildren...) }
func I(attribsChildren ...any) *Element        { return mx.NewElement("i", attribsChildren...) }
func IFrame(attribsChildren ...any) *Element   { return mx.NewElement("iframe", attribsChildren...) }
func Image(attribsChildren ...any) *Element    { return mx.NewElement("image", attribsChildren...) }
func Img(attribs ...Attrib) *Element           { return mx.NewVoidElement("img", attribs...) }
func Input(attribs ...Attrib) *Element         { return mx.NewVoidElement("input", attribs...) }
func Ins(attribsChildren ...any) *Element      { return mx.NewElement("ins", attribsChildren...) }
func Kbd(attribsChildren ...any) *Element      { return mx.NewElement("kbd", attribsChildren...) }
func Label(attribsChildren ...any) *Element    { return mx.NewElement("label", attribsChildren...) }
func Legend(attribsChildren ...any) *Element   { return mx.NewElement("legend", attribsChildren...) }
func LI(attribsChildren ...any) *Element       { return mx.NewElement("li", attribsChildren...) }
func Link(attribs ...Attrib) *Element          { return mx.NewVoidElement("link", attribs...) }
func Main(attribsChildren ...any) *Element     { return mx.NewElement("main", attribsChildren...) }
func Map(attribsChildren ...any) *Element      { return mx.NewElement("map", attribsChildren...) }
func Mark(attribsChildren ...any) *Element     { return mx.NewElement("mark", attribsChildren...) }
func Marquee(attribsChildren ...any) *Element  { return mx.NewElement("marquee", attribsChildren...) }
func Math(attribsChildren ...any) *Element     { return mx.NewElement("math", attribsChildren...) }
func Menu(attribsChildren ...any) *Element     { return mx.NewElement("menu", attribsChildren...) }
func MenuItem(attribsChildren ...any) *Element { return mx.NewElement("menuitem", attribsChildren...) }
func Meta(attribs ...Attrib) *Element          { return mx.NewVoidElement("meta", attribs...) }
func Meter(attribsChildren ...any) *Element    { return mx.NewElement("meter", attribsChildren...) }
func Nav(attribsChildren ...any) *Element      { return mx.NewElement("nav", attribsChildren...) }
func NoBr(attribsChildren ...any) *Element     { return mx.NewElement("nobr", attribsChildren...) }
func NoEmbed(attribsChildren ...any) *Element  { return mx.NewElement("noembed", attribsChildren...) }
func NoFrames(attribsChildren ...any) *Element { return mx.NewElement("noframes", attribsChildren...) }
func NoScript(attribsChildren ...any) *Element { return mx.NewElement("noscript", attribsChildren...) }
func Object(attribsChildren ...any) *Element   { return mx.NewElement("object", attribsChildren...) }
func OL(attribsChildren ...any) *Element       { return mx.NewElement("ol", attribsChildren...) }
func OptGroup(attribsChildren ...any) *Element { return mx.NewElement("optgroup", attribsChildren...) }
func Option(attribsChildren ...any) *Element   { return mx.NewElement("option", attribsChildren...) }
func Output(attribsChildren ...any) *Element   { return mx.NewElement("output", attribsChildren...) }
func P(attribsChildren ...any) *Element        { return mx.NewElement("p", attribsChildren...) }
func Picture(attribsChildren ...any) *Element  { return mx.NewElement("picture", attribsChildren...) }
func Plaintext(attribsChildren ...any) *Element {
	return mx.NewElement("plaintext", attribsChildren...)
}
func Portal(attribsChildren ...any) *Element   { return mx.NewElement("portal", attribsChildren...) }
func Pre(attribsChildren ...any) *Element      { return mx.NewElement("pre", attribsChildren...) }
func Progress(attribsChildren ...any) *Element { return mx.NewElement("progress", attribsChildren...) }
func Q(attribsChildren ...any) *Element        { return mx.NewElement("q", attribsChildren...) }
func RB(attribsChildren ...any) *Element       { return mx.NewElement("rb", attribsChildren...) }
func RP(attribsChildren ...any) *Element       { return mx.NewElement("rp", attribsChildren...) }
func RT(attribsChildren ...any) *Element       { return mx.NewElement("rt", attribsChildren...) }
func RTC(attribsChildren ...any) *Element      { return mx.NewElement("rtc", attribsChildren...) }
func Ruby(attribsChildren ...any) *Element     { return mx.NewElement("ruby", attribsChildren...) }
func S(attribsChildren ...any) *Element        { return mx.NewElement("s", attribsChildren...) }
func Samp(attribsChildren ...any) *Element     { return mx.NewElement("samp", attribsChildren...) }
func Script(attribsChildren ...any) *Element   { return mx.NewElement("script", attribsChildren...) }
func Search(attribsChildren ...any) *Element   { return mx.NewElement("search", attribsChildren...) }
func Section(attribsChildren ...any) *Element  { return mx.NewElement("section", attribsChildren...) }
func Select(attribsChildren ...any) *Element   { return mx.NewElement("select", attribsChildren...) }
func Shadow(attribsChildren ...any) *Element   { return mx.NewElement("shadow", attribsChildren...) }
func Slot(attribsChildren ...any) *Element     { return mx.NewElement("slot", attribsChildren...) }
func Small(attribsChildren ...any) *Element    { return mx.NewElement("small", attribsChildren...) }
func Source(attribs ...Attrib) *Element        { return mx.NewVoidElement("source", attribs...) }
func Span(attribsChildren ...any) *Element     { return mx.NewElement("span", attribsChildren...) }
func Strike(attribsChildren ...any) *Element   { return mx.NewElement("strike", attribsChildren...) }
func Strong(attribsChildren ...any) *Element   { return mx.NewElement("strong", attribsChildren...) }
func StyleElem(css string) *Element            { return mx.NewElement("style", Raw(css)) }
func Sub(attribsChildren ...any) *Element      { return mx.NewElement("sub", attribsChildren...) }
func Summary(attribsChildren ...any) *Element  { return mx.NewElement("summary", attribsChildren...) }
func Sup(attribsChildren ...any) *Element      { return mx.NewElement("sup", attribsChildren...) }
func Svg(attribsChildren ...any) *Element      { return mx.NewElement("svg", attribsChildren...) }
func Table(attribsChildren ...any) *Element    { return mx.NewElement("table", attribsChildren...) }
func TBody(attribsChildren ...any) *Element    { return mx.NewElement("tbody", attribsChildren...) }
func TD(attribsChildren ...any) *Element       { return mx.NewElement("td", attribsChildren...) }
func TemplateElem(attribsChildren ...any) *Element {
	return mx.NewElement("template", attribsChildren...)
}
func TextArea(attribsChildren ...any) *Element  { return mx.NewElement("textarea", attribsChildren...) }
func TFoot(attribsChildren ...any) *Element     { return mx.NewElement("tfoot", attribsChildren...) }
func TH(attribsChildren ...any) *Element        { return mx.NewElement("th", attribsChildren...) }
func THead(attribsChildren ...any) *Element     { return mx.NewElement("thead", attribsChildren...) }
func Time(attribsChildren ...any) *Element      { return mx.NewElement("time", attribsChildren...) }
func TitleElem(attribsChildren ...any) *Element { return mx.NewElement("title", attribsChildren...) }
func TR(attribsChildren ...any) *Element        { return mx.NewElement("tr", attribsChildren...) }
func Track(attribsChildren ...any) *Element     { return mx.NewElement("track", attribsChildren...) }
func TT(attribsChildren ...any) *Element        { return mx.NewElement("tt", attribsChildren...) }
func U(attribsChildren ...any) *Element         { return mx.NewElement("u", attribsChildren...) }
func UL(attribsChildren ...any) *Element        { return mx.NewElement("ul", attribsChildren...) }
func Var(attribsChildren ...any) *Element       { return mx.NewElement("var", attribsChildren...) }
func Video(attribsChildren ...any) *Element     { return mx.NewElement("video", attribsChildren...) }
func WBr(attribs ...Attrib) *Element            { return mx.NewVoidElement("wbr", attribs...) }
