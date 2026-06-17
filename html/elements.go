package html

import (
	"github.com/ungerik/go-mx"
)

// Element creates a generic HTML element with the given tag name, taking
// attributes and children as variadic arguments.
func Element(name string, attribsChildren ...any) *mx.Element {
	return mx.NewElement(name, attribsChildren...)
}

// VoidElement creates a generic void HTML element with the given tag name,
// taking only attributes and no children.
func VoidElement(name string, attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement(name, attribs...)
}

// Textf returns escaped Text formatted like fmt.Sprintf.
func Textf(format string, args ...any) Text {
	return mx.Textf(format, args...)
}

// See https://github.com/jozo/all-html-elements-and-attributes

// A creates an <a> anchor element, a hyperlink to other pages, files, or locations.
func A(attribsChildren ...any) *mx.Element { return Element("a", attribsChildren...) }

// Abbr creates an <abbr> element marking an abbreviation or acronym.
func Abbr(attribsChildren ...any) *mx.Element { return Element("abbr", attribsChildren...) }

// Address creates an <address> element for contact information of its nearest article or body.
func Address(attribsChildren ...any) *mx.Element { return Element("address", attribsChildren...) }

// Area creates a void <area> element defining a clickable region inside an image Map.
// It takes only attributes and has no children.
func Area(attribs ...mx.Attrib) *mx.Element { return VoidElement("area", attribs...) }

// Article creates an <article> element for self-contained, independently distributable content.
func Article(attribsChildren ...any) *mx.Element { return Element("article", attribsChildren...) }

// Aside creates an <aside> element for content tangentially related to the surrounding content.
func Aside(attribsChildren ...any) *mx.Element { return Element("aside", attribsChildren...) }

// Audio creates an <audio> element for embedding sound content.
func Audio(attribsChildren ...any) *mx.Element { return Element("audio", attribsChildren...) }

// B creates a <b> element for text stylistically offset from normal prose
// without conveying extra importance, such as keywords or product names.
// Use Strong for semantic emphasis.
func B(attribsChildren ...any) *mx.Element { return Element("b", attribsChildren...) }

// Base creates a void <base> element setting the document base URL and default
// link target. It takes only attributes and has no children.
func Base(attribs ...mx.Attrib) *mx.Element { return VoidElement("base", attribs...) }

// BDi creates a <bdi> element isolating a span of text whose directionality may differ.
func BDi(attribsChildren ...any) *mx.Element { return Element("bdi", attribsChildren...) }

// BDo creates a <bdo> element overriding the text directionality via the dir attribute.
func BDo(attribsChildren ...any) *mx.Element { return Element("bdo", attribsChildren...) }

// Blockquote creates a <blockquote> element for an extended quotation.
func Blockquote(attribsChildren ...any) *mx.Element {
	return mx.NewElement("blockquote", attribsChildren...)
}

// Body creates the <body> element containing the document's content.
func Body(attribsChildren ...any) *mx.Element { return Element("body", attribsChildren...) }

// Br creates a void <br> line break element. It takes only attributes and has no children.
func Br(attribs ...mx.Attrib) *mx.Element { return VoidElement("br", attribs...) }

// Button creates a <button> element without a type attribute. A missing type
// defaults to the submit state in every context, but its effect depends on the
// button's form association:
//   - Inside a <form> (or associated via the form attribute) it acts as a submit
//     button: activating it (click, or Enter while the form has focus) submits
//     the form. This is a common gotcha for buttons meant only to run scripts.
//   - Outside any form the submit default has nothing to submit, so the button
//     has no default action and does something only when scripted.
//
// Use SubmitButton, ResetButton, or ButtonButton to set the type explicitly.
func Button(attribsChildren ...any) *mx.Element { return Element("button", attribsChildren...) }

// SubmitButton is a <button type="submit"> that submits its form when activated.
func SubmitButton(attribsChildren ...any) *mx.Element {
	return Element("button", append([]any{Type("submit")}, attribsChildren...)...)
}

// ResetButton is a <button type="reset"> that resets its form's controls.
func ResetButton(attribsChildren ...any) *mx.Element {
	return Element("button", append([]any{Type("reset")}, attribsChildren...)...)
}

// ButtonButton is a <button type="button"> with no default behavior: unlike a
// bare Button (which defaults to type="submit") it does not submit a form, so
// use it for buttons that only trigger scripts.
func ButtonButton(attribsChildren ...any) *mx.Element {
	return Element("button", append([]any{Type("button")}, attribsChildren...)...)
}

// Canvas creates a <canvas> element used to draw graphics via scripting (2D or WebGL).
func Canvas(attribsChildren ...any) *mx.Element { return Element("canvas", attribsChildren...) }

// Caption creates a <caption> element giving a Table its title.
func Caption(attribsChildren ...any) *mx.Element { return Element("caption", attribsChildren...) }

// Cite creates a <cite> element marking the title of a referenced creative work.
func Cite(attribsChildren ...any) *mx.Element { return Element("cite", attribsChildren...) }

// Code creates a <code> element marking a fragment of computer code.
func Code(attribsChildren ...any) *mx.Element { return Element("code", attribsChildren...) }

// Col creates a void <col> element defining a column within a ColGroup.
// It takes only attributes and has no children.
func Col(attribs ...mx.Attrib) *mx.Element { return VoidElement("col", attribs...) }

// ColGroup creates a <colgroup> element grouping one or more columns of a Table.
func ColGroup(attribsChildren ...any) *mx.Element {
	return mx.NewElement("colgroup", attribsChildren...)
}

// Command was removed from HTML spec

// Content was an obsolete Shadow DOM v0 feature that was never standardized; use Slot instead.

// Data creates a <data> element linking its content to a machine-readable value via the value attribute.
func Data(attribsChildren ...any) *mx.Element { return Element("data", attribsChildren...) }

// DataList creates a <datalist> element holding a set of Option suggestions for an Input.
func DataList(attribsChildren ...any) *mx.Element {
	return mx.NewElement("datalist", attribsChildren...)
}

// DD creates a <dd> element giving the description or value for the preceding DT term in a DL.
func DD(attribsChildren ...any) *mx.Element { return Element("dd", attribsChildren...) }

// Del creates a <del> element marking text that has been deleted from the document.
func Del(attribsChildren ...any) *mx.Element { return Element("del", attribsChildren...) }

// Details creates a <details> element disclosing additional content, toggled via its Summary.
func Details(attribsChildren ...any) *mx.Element { return Element("details", attribsChildren...) }

// Dfn creates a <dfn> element marking the defining instance of a term.
func Dfn(attribsChildren ...any) *mx.Element { return Element("dfn", attribsChildren...) }

// Dialog creates a <dialog> element representing a modal or non-modal dialog box.
func Dialog(attribsChildren ...any) *mx.Element { return Element("dialog", attribsChildren...) }

// Div creates a <div> generic block-level container with no semantic meaning.
func Div(attribsChildren ...any) *mx.Element { return Element("div", attribsChildren...) }

// DivClass creates a <div> with the given space separated class names
// as a shortcut for Div(Class(classes), attribsChildren...).
func DivClass(classes string, attribsChildren ...any) *mx.Element {
	return Element("div", append([]any{Class(classes)}, attribsChildren...)...)
}

// DivID creates a <div> with the given id attribute
// as a shortcut for Div(ID(id), attribsChildren...).
func DivID(id string, attribsChildren ...any) *mx.Element {
	return Element("div", append([]any{ID(id)}, attribsChildren...)...)
}

// DL creates a <dl> description list of DT term and DD description pairs.
func DL(attribsChildren ...any) *mx.Element { return Element("dl", attribsChildren...) }

// DT creates a <dt> element naming a term in a DL description list.
func DT(attribsChildren ...any) *mx.Element { return Element("dt", attribsChildren...) }

// Em creates an <em> element marking text with stress emphasis.
func Em(attribsChildren ...any) *mx.Element { return Element("em", attribsChildren...) }

// Embed creates a void <embed> element that embeds external content such as a
// plug-in or media at the insertion point. It takes only attributes and has no
// children. Use Object or IFrame for better browser compatibility.
func Embed(attribs ...mx.Attrib) *mx.Element { return VoidElement("embed", attribs...) }

// FieldSet creates a <fieldset> element grouping related form controls, optionally labeled by a Legend.
func FieldSet(attribsChildren ...any) *mx.Element {
	return mx.NewElement("fieldset", attribsChildren...)
}

// FigCaption creates a <figcaption> element providing a caption for its parent Figure.
func FigCaption(attribsChildren ...any) *mx.Element {
	return mx.NewElement("figcaption", attribsChildren...)
}

// Figure creates a <figure> element for self-contained content such as an image, optionally with a FigCaption.
func Figure(attribsChildren ...any) *mx.Element { return Element("figure", attribsChildren...) }

// Footer creates a <footer> element for the footer of its nearest sectioning content.
func Footer(attribsChildren ...any) *mx.Element { return Element("footer", attribsChildren...) }

// Form creates a <form> element for submitting user input to a server.
func Form(attribsChildren ...any) *mx.Element { return Element("form", attribsChildren...) }

// H1 creates an <h1> top-level section heading.
func H1(attribsChildren ...any) *mx.Element { return Element("h1", attribsChildren...) }

// H1Class creates an <h1> with the given space separated class names
// as a shortcut for H1(Class(classes), attribsChildren...).
func H1Class(classes string, attribsChildren ...any) *mx.Element {
	return Element("h1", append([]any{Class(classes)}, attribsChildren...)...)
}

// H2 creates an <h2> second-level section heading.
func H2(attribsChildren ...any) *mx.Element { return Element("h2", attribsChildren...) }

// H2Class creates an <h2> with the given space separated class names
// as a shortcut for H2(Class(classes), attribsChildren...).
func H2Class(classes string, attribsChildren ...any) *mx.Element {
	return Element("h2", append([]any{Class(classes)}, attribsChildren...)...)
}

// H3 creates an <h3> third-level section heading.
func H3(attribsChildren ...any) *mx.Element { return Element("h3", attribsChildren...) }

// H3Class creates an <h3> with the given space separated class names
// as a shortcut for H3(Class(classes), attribsChildren...).
func H3Class(classes string, attribsChildren ...any) *mx.Element {
	return Element("h3", append([]any{Class(classes)}, attribsChildren...)...)
}

// H4 creates an <h4> fourth-level section heading.
func H4(attribsChildren ...any) *mx.Element { return Element("h4", attribsChildren...) }

// H4Class creates an <h4> with the given space separated class names
// as a shortcut for H4(Class(classes), attribsChildren...).
func H4Class(classes string, attribsChildren ...any) *mx.Element {
	return Element("h4", append([]any{Class(classes)}, attribsChildren...)...)
}

// H5 creates an <h5> fifth-level section heading.
func H5(attribsChildren ...any) *mx.Element { return Element("h5", attribsChildren...) }

// H5Class creates an <h5> with the given space separated class names
// as a shortcut for H5(Class(classes), attribsChildren...).
func H5Class(classes string, attribsChildren ...any) *mx.Element {
	return Element("h5", append([]any{Class(classes)}, attribsChildren...)...)
}

// H6 creates an <h6> sixth-level section heading.
func H6(attribsChildren ...any) *mx.Element { return Element("h6", attribsChildren...) }

// H6Class creates an <h6> with the given space separated class names
// as a shortcut for H6(Class(classes), attribsChildren...).
func H6Class(classes string, attribsChildren ...any) *mx.Element {
	return Element("h6", append([]any{Class(classes)}, attribsChildren...)...)
}

// Head creates the <head> element containing document metadata.
func Head(attribsChildren ...any) *mx.Element { return Element("head", attribsChildren...) }

// Header creates a <header> element for introductory content of its nearest sectioning content.
func Header(attribsChildren ...any) *mx.Element { return Element("header", attribsChildren...) }

// HGroup creates an <hgroup> element grouping a heading with associated tagline paragraphs.
func HGroup(attribsChildren ...any) *mx.Element { return Element("hgroup", attribsChildren...) }

// HR creates a void <hr> thematic-break element. It takes only attributes and has no children.
func HR(attribs ...mx.Attrib) *mx.Element { return VoidElement("hr", attribs...) }

// HTML creates the root <html> element of the document.
func HTML(attribsChildren ...any) *mx.Element { return Element("html", attribsChildren...) }

// I creates an <i> element for text in an alternate voice or mood, such as a
// technical term, foreign phrase, or thought. Use Em for semantic emphasis.
func I(attribsChildren ...any) *mx.Element { return Element("i", attribsChildren...) }

// IFrame creates an <iframe> element embedding another HTML page into the current one.
func IFrame(attribsChildren ...any) *mx.Element { return Element("iframe", attribsChildren...) }

// Img creates a void <img> element that embeds an image via its Src attribute.
// It takes only attributes and has no children. (Image is the SVG element; use
// Img for HTML.)
func Img(attribs ...mx.Attrib) *mx.Element { return VoidElement("img", attribs...) }

// ImgSrc creates a void <img> element loading the image at the given URL
// as a shortcut for Img(Src(url), attribs...).
func ImgSrc(url string, attribs ...mx.Attrib) *mx.Element {
	return VoidElement("img", append([]mx.Attrib{Src(url)}, attribs...)...)
}

// Input creates a void <input> form control element. It takes only attributes
// and has no children; the type attribute determines the kind of control.
func Input(attribs ...mx.Attrib) *mx.Element { return VoidElement("input", attribs...) }

// InputTypeButton creates an <input type="button"> push button with no default behavior.
func InputTypeButton(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "button", attribs)...)
}

// InputTypeCheckbox creates an <input type="checkbox"> toggle control.
func InputTypeCheckbox(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "checkbox", attribs)...)
}

// InputTypeColor creates an <input type="color"> color picker control.
func InputTypeColor(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "color", attribs)...)
}

// InputTypeDate creates an <input type="date"> date picker (year, month, day).
func InputTypeDate(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "date", attribs)...)
}

// InputTypeDatetimeLocal creates an <input type="datetime-local"> date and time picker without time zone.
func InputTypeDatetimeLocal(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "datetime-local", attribs)...)
}

// InputTypeEmail creates an <input type="email"> field for an email address with validation.
func InputTypeEmail(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "email", attribs)...)
}

// InputTypeFile creates an <input type="file"> control for selecting files to upload.
func InputTypeFile(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "file", attribs)...)
}

// InputTypeHidden creates an <input type="hidden"> control submitted with the form but not displayed.
func InputTypeHidden(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "hidden", attribs)...)
}

// InputTypeImage creates an <input type="image"> graphical submit button.
func InputTypeImage(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "image", attribs)...)
}

// InputTypeMonth creates an <input type="month"> month and year picker.
func InputTypeMonth(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "month", attribs)...)
}

// InputTypeNumber creates an <input type="number"> numeric entry control.
func InputTypeNumber(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "number", attribs)...)
}

// InputTypePassword creates an <input type="password"> field whose value is obscured.
func InputTypePassword(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "password", attribs)...)
}

// InputTypeRadio creates an <input type="radio"> radio button; same-name radios form an exclusive group.
func InputTypeRadio(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "radio", attribs)...)
}

// InputTypeRange creates an <input type="range"> slider for a value within a range.
func InputTypeRange(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "range", attribs)...)
}

// InputTypeReset creates an <input type="reset"> button that resets the form's controls.
func InputTypeReset(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "reset", attribs)...)
}

// InputTypeSearch creates an <input type="search"> single-line text field for search queries.
func InputTypeSearch(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "search", attribs)...)
}

// InputTypeSubmit creates an <input type="submit"> button that submits the form.
func InputTypeSubmit(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "submit", attribs)...)
}

// InputTypeTel creates an <input type="tel"> field for a telephone number.
func InputTypeTel(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "tel", attribs)...)
}

// InputTypeText creates an <input type="text"> single-line text field.
func InputTypeText(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "text", attribs)...)
}

// InputTypeTime creates an <input type="time"> time picker without a time zone.
func InputTypeTime(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "time", attribs)...)
}

// InputTypeURL creates an <input type="url"> field for a URL with validation.
func InputTypeURL(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "url", attribs)...)
}

// InputTypeWeek creates an <input type="week"> week and year picker.
func InputTypeWeek(attribs ...mx.Attrib) *mx.Element {
	return mx.NewVoidElement("input", mx.PrependAttrib("type", "week", attribs)...)
}

// Ins creates an <ins> element marking text that has been inserted into the document.
func Ins(attribsChildren ...any) *mx.Element { return Element("ins", attribsChildren...) }

// Kbd creates a <kbd> element marking user keyboard input.
func Kbd(attribsChildren ...any) *mx.Element { return Element("kbd", attribsChildren...) }

// Label creates a <label> element captioning a form control via its for attribute or by wrapping it.
func Label(attribsChildren ...any) *mx.Element { return Element("label", attribsChildren...) }

// LabelFor creates a <label> bound to the control with the given id
// as a shortcut for Label(For(id), attribsChildren...).
func LabelFor(id string, attribsChildren ...any) *mx.Element {
	return Element("label", append([]any{For(id)}, attribsChildren...)...)
}

// Legend creates a <legend> element giving a caption for its parent FieldSet.
func Legend(attribsChildren ...any) *mx.Element { return Element("legend", attribsChildren...) }

// LI creates an <li> list item within an OL, UL, or Menu.
func LI(attribsChildren ...any) *mx.Element { return Element("li", attribsChildren...) }

// Link creates a void <link> element relating the document to external resources such as stylesheets.
// It takes only attributes and has no children.
func Link(attribs ...mx.Attrib) *mx.Element { return VoidElement("link", attribs...) }

// Main creates the <main> element wrapping the dominant content of the document body.
func Main(attribsChildren ...any) *mx.Element { return Element("main", attribsChildren...) }

// Map creates a <map> element defining an image map together with its Area children.
func Map(attribsChildren ...any) *mx.Element { return Element("map", attribsChildren...) }

// Mark creates a <mark> element highlighting text for reference or relevance.
func Mark(attribsChildren ...any) *mx.Element { return Element("mark", attribsChildren...) }

// Math creates a <math> element, the root of MathML mathematical notation.
func Math(attribsChildren ...any) *mx.Element { return Element("math", attribsChildren...) }

// Meta creates a void <meta> element carrying document metadata. It takes only attributes and has no children.
func Meta(attribs ...mx.Attrib) *mx.Element { return VoidElement("meta", attribs...) }

// Menu creates a <menu> element representing a semantic list of commands, similar to UL.
func Menu(attribsChildren ...any) *mx.Element { return Element("menu", attribsChildren...) }

// Meter creates a <meter> element displaying a scalar measurement within a known range (a gauge).
func Meter(attribsChildren ...any) *mx.Element { return Element("meter", attribsChildren...) }

// Nav creates a <nav> element for a section of major navigation links.
func Nav(attribsChildren ...any) *mx.Element { return Element("nav", attribsChildren...) }

// NoEmbed creates an obsolete <noembed> element; provide fallback content via Object instead.
func NoEmbed(attribsChildren ...any) *mx.Element { return Element("noembed", attribsChildren...) }

// NoScript creates a <noscript> element with content shown when scripting is disabled or unsupported.
func NoScript(attribsChildren ...any) *mx.Element {
	return mx.NewElement("noscript", attribsChildren...)
}

// Object creates an <object> element embedding an external resource such as an image, PDF, or plugin.
func Object(attribsChildren ...any) *mx.Element { return Element("object", attribsChildren...) }

// OL creates an <ol> ordered-list element without a type attribute. Item markers
// default to decimal numbers (1, 2, 3, …); the CSS list-style-type property can
// change the rendered style without setting the attribute. Combine with Start to
// change the first number and the Reversed boolean attribute to count downward.
//
// Use OLDecimal, OLLowerAlpha, OLUpperAlpha, OLLowerRoman, or OLUpperRoman to set
// the marker type explicitly via the attribute.
func OL(attribsChildren ...any) *mx.Element { return Element("ol", attribsChildren...) }

// OLDecimal is an <ol type="1"> with decimal number markers (1, 2, 3, …) — the
// same markers as a bare OL, but with the type attribute set explicitly.
func OLDecimal(attribsChildren ...any) *mx.Element {
	return Element("ol", append([]any{Type("1")}, attribsChildren...)...)
}

// OLLowerAlpha is an <ol type="a"> with lowercase letter markers (a, b, c, …).
func OLLowerAlpha(attribsChildren ...any) *mx.Element {
	return Element("ol", append([]any{Type("a")}, attribsChildren...)...)
}

// OLUpperAlpha is an <ol type="A"> with uppercase letter markers (A, B, C, …).
func OLUpperAlpha(attribsChildren ...any) *mx.Element {
	return Element("ol", append([]any{Type("A")}, attribsChildren...)...)
}

// OLLowerRoman is an <ol type="i"> with lowercase roman numeral markers (i, ii, iii, …).
func OLLowerRoman(attribsChildren ...any) *mx.Element {
	return Element("ol", append([]any{Type("i")}, attribsChildren...)...)
}

// OLUpperRoman is an <ol type="I"> with uppercase roman numeral markers (I, II, III, …).
func OLUpperRoman(attribsChildren ...any) *mx.Element {
	return Element("ol", append([]any{Type("I")}, attribsChildren...)...)
}

// OptGroup creates an <optgroup> element grouping related Option items within a Select.
func OptGroup(attribsChildren ...any) *mx.Element {
	return mx.NewElement("optgroup", attribsChildren...)
}

// Option creates an <option> item within a Select, OptGroup, or DataList.
func Option(attribsChildren ...any) *mx.Element { return Element("option", attribsChildren...) }

// Output creates an <output> element holding the result of a calculation or user action.
func Output(attribsChildren ...any) *mx.Element { return Element("output", attribsChildren...) }

// P creates a <p> paragraph element.
func P(attribsChildren ...any) *mx.Element { return Element("p", attribsChildren...) }

// PClass creates a <p> with the given space separated class names
// as a shortcut for P(Class(classes), attribsChildren...).
func PClass(classes string, attribsChildren ...any) *mx.Element {
	return Element("p", append([]any{Class(classes)}, attribsChildren...)...)
}

// Picture creates a <picture> element offering Source alternatives for its contained Img.
func Picture(attribsChildren ...any) *mx.Element { return Element("picture", attribsChildren...) }

// Portal was an experimental feature that was never standardized and is not supported by browsers.

// Pre creates a <pre> element whose text is rendered preformatted, preserving whitespace.
func Pre(attribsChildren ...any) *mx.Element { return Element("pre", attribsChildren...) }

// Progress creates a <progress> element displaying the completion progress of a task (a gauge).
func Progress(attribsChildren ...any) *mx.Element {
	return mx.NewElement("progress", attribsChildren...)
}

// Q creates a <q> element for a short inline quotation.
func Q(attribsChildren ...any) *mx.Element { return Element("q", attribsChildren...) }

// RB creates an obsolete <rb> ruby base element; modern Ruby markup nests the base text directly.
func RB(attribsChildren ...any) *mx.Element { return Element("rb", attribsChildren...) }

// RP creates an <rp> element providing fallback parentheses around ruby text for browsers lacking Ruby support.
func RP(attribsChildren ...any) *mx.Element { return Element("rp", attribsChildren...) }

// RT creates an <rt> element giving the pronunciation annotation within a Ruby.
func RT(attribsChildren ...any) *mx.Element { return Element("rt", attribsChildren...) }

// RTC creates an obsolete <rtc> ruby text container element.
func RTC(attribsChildren ...any) *mx.Element { return Element("rtc", attribsChildren...) }

// Ruby creates a <ruby> element for annotating East Asian characters with pronunciation.
func Ruby(attribsChildren ...any) *mx.Element { return Element("ruby", attribsChildren...) }

// S creates an <s> element marking text that is no longer accurate or relevant.
// Use Del for content removed by document edits.
func S(attribsChildren ...any) *mx.Element { return Element("s", attribsChildren...) }

// Samp creates a <samp> element marking sample or quoted output from a program or system.
func Samp(attribsChildren ...any) *mx.Element { return Element("samp", attribsChildren...) }

// Script creates a <script> element embedding or referencing executable code.
func Script(attribsChildren ...any) *mx.Element { return Element("script", attribsChildren...) }

// Search creates a <search> element grouping form controls and content related to a search or filtering operation.
func Search(attribsChildren ...any) *mx.Element { return Element("search", attribsChildren...) }

// Section creates a <section> element for a standalone thematic grouping of content.
func Section(attribsChildren ...any) *mx.Element { return Element("section", attribsChildren...) }

// Select creates a <select> element offering a menu of Option choices.
func Select(attribsChildren ...any) *mx.Element { return Element("select", attribsChildren...) }

// Shadow was an obsolete Shadow DOM v0 feature that was never standardized; use Slot instead.

// Slot creates a <slot> element, a placeholder inside a web component filled with markup from the light DOM.
func Slot(attribsChildren ...any) *mx.Element { return Element("slot", attribsChildren...) }

// Small is still valid in HTML5 but use CSS for better control.
func Small(attribsChildren ...any) *mx.Element { return Element("small", attribsChildren...) }

// Source creates a void <source> element specifying one media or image alternative for a
// Picture, Audio, or Video. It takes only attributes and has no children.
func Source(attribs ...mx.Attrib) *mx.Element { return VoidElement("source", attribs...) }

// Span creates a <span> generic inline container with no semantic meaning.
func Span(attribsChildren ...any) *mx.Element { return Element("span", attribsChildren...) }

// SpanClass creates a <span> with the given space separated class names
// as a shortcut for Span(Class(classes), attribsChildren...).
func SpanClass(classes string, attribsChildren ...any) *mx.Element {
	return Element("span", append([]any{Class(classes)}, attribsChildren...)...)
}

// Strong creates a <strong> element marking content of strong importance, seriousness, or urgency.
func Strong(attribsChildren ...any) *mx.Element { return Element("strong", attribsChildren...) }

// StyleElem creates a <style> element wrapping the given raw CSS as its content.
func StyleElem(css string) *mx.Element { return Element("style", Raw(css)) }

// Sub creates a <sub> element rendering its content as subscript.
func Sub(attribsChildren ...any) *mx.Element { return Element("sub", attribsChildren...) }

// Summary creates a <summary> element providing the visible heading and toggle for a Details element.
func Summary(attribsChildren ...any) *mx.Element { return Element("summary", attribsChildren...) }

// Sup creates a <sup> element rendering its content as superscript.
func Sup(attribsChildren ...any) *mx.Element { return Element("sup", attribsChildren...) }

// Table creates a <table> element presenting data in rows and columns.
func Table(attribsChildren ...any) *mx.Element { return Element("table", attribsChildren...) }

// TBody creates a <tbody> element grouping the body rows of a Table.
func TBody(attribsChildren ...any) *mx.Element { return Element("tbody", attribsChildren...) }

// TD creates a <td> standard data cell within a Table row.
func TD(attribsChildren ...any) *mx.Element { return Element("td", attribsChildren...) }

// TemplateElem creates a <template> element holding markup that is not rendered until cloned by script.
func TemplateElem(attribsChildren ...any) *mx.Element {
	return mx.NewElement("template", attribsChildren...)
}

// TextArea creates a <textarea> element for multi-line plain-text input.
func TextArea(attribsChildren ...any) *mx.Element {
	return mx.NewElement("textarea", attribsChildren...)
}

// TFoot creates a <tfoot> element grouping the footer rows of a Table.
func TFoot(attribsChildren ...any) *mx.Element { return Element("tfoot", attribsChildren...) }

// TH creates a <th> header cell within a Table row.
func TH(attribsChildren ...any) *mx.Element { return Element("th", attribsChildren...) }

// THead creates a <thead> element grouping the header rows of a Table.
func THead(attribsChildren ...any) *mx.Element { return Element("thead", attribsChildren...) }

// Time creates a <time> element linking its content to a machine-readable date or time via the datetime attribute.
func Time(attribsChildren ...any) *mx.Element { return Element("time", attribsChildren...) }

// TitleElem creates the document <title> element shown in the browser title bar or tab.
func TitleElem(attribsChildren ...any) *mx.Element { return Element("title", attribsChildren...) }

// TR creates a <tr> table row holding TD or TH cells.
func TR(attribsChildren ...any) *mx.Element { return Element("tr", attribsChildren...) }

// Track creates a void <track> element supplying timed text tracks (subtitles, captions) for
// Audio or Video. It takes only attributes and has no children.
func Track(attribs ...mx.Attrib) *mx.Element { return VoidElement("track", attribs...) }

// UL creates a <ul> unordered list of LI items.
func UL(attribsChildren ...any) *mx.Element { return Element("ul", attribsChildren...) }

// Var creates a <var> element marking a variable in a mathematical expression or program.
func Var(attribsChildren ...any) *mx.Element { return Element("var", attribsChildren...) }

// Video creates a <video> element for embedding video content.
func Video(attribsChildren ...any) *mx.Element { return Element("video", attribsChildren...) }

// WBr creates a void <wbr> element marking an optional line-break opportunity.
// It takes only attributes and has no children.
func WBr(attribs ...mx.Attrib) *mx.Element { return VoidElement("wbr", attribs...) }

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
