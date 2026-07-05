package html

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/ungerik/go-mx"
)

// Attribs is an alias for mx.Attribs, a collection of HTML attributes.
type Attribs = mx.Attribs

// AttribValue is the type set accepted by the mixed-type attribute constructors
// in this package — those whose HTML value may be either a number or a string,
// because the attribute accepts a bare number as well as a keyword, a date/time
// or a length. Strings pass through unchanged; numbers are formatted as plain
// decimals, so Width(100), Width("100%"), Max("2025-01-01") and Step("any") all
// work.
//
// Attributes are typed according to what their HTML value can be:
//   - always a plain integer → int (e.g. Cols, ColSpan, MaxLength)
//   - always a plain number → float64 (e.g. High, Low, Optimum)
//   - a list of plain numbers → ...float64, rendered comma-separated like Coords
//   - number or string (keyword/date/length) → generic over AttribValue
//   - never numeric (URL, keyword, CSS, script) → string
type AttribValue interface {
	~string |
		~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

// attribValueString formats an attribute value as a string. Strings pass through
// and integer types go through fmt.Sprint, but floats use strconv.FormatFloat
// with the 'f' format so small or large magnitudes render as plain decimals
// instead of fmt's scientific notation (e.g. 0.00005 not "5e-05").
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

// Attrib constructs an arbitrary HTML attribute with the given name and value.
// The value may be a string or any number type (see [AttribValue]): strings pass
// through unchanged and float values are formatted as plain decimals (never
// scientific notation). It also serves as an escape hatch for attributes not
// covered by a dedicated constructor.
func Attrib[T AttribValue](name string, value T) mx.Attribute {
	return mx.Attribute{Name: name, Value: attribValueString(value)}
}

// BoolAttrib implements the mx.Attrib interface
// and returns its string value for both name and value.
type BoolAttrib string

var _ mx.Attrib = BoolAttrib("")

// AttribName returns the attribute name (the BoolAttrib's string value).
func (a BoolAttrib) AttribName() string {
	return string(a)
}

// AttribValue returns the attribute value (the BoolAttrib's string value) and a nil error.
func (a BoolAttrib) AttribValue(context.Context) (string, error) {
	return string(a), nil
}

// See https://github.com/jozo/all-html-elements-and-attributes
// and https://html.spec.whatwg.org/multipage/indices.html#attributes-3

// Accept sets the accept attribute, the comma-separated list of file types an <input type=file> accepts.
func Accept(contentTypes ...string) mx.Attrib {
	return mx.NewAttrib("accept", strings.Join(contentTypes, ","))
}

// AcceptCharset sets the accept-charset attribute, the character encodings a <form> accepts on submission.
func AcceptCharset(charsets ...string) mx.Attrib {
	return mx.NewAttrib("accept-charset", strings.Join(charsets, " "))
}

// AccessKey sets the accesskey attribute, a keyboard shortcut hint for activating or focusing the element.
func AccessKey(value string) mx.Attrib { return mx.NewAttrib("accesskey", value) }

// AccessKeyf is like AccessKey but builds the value with a fmt format string.
func AccessKeyf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("accesskey", valueFmt, a...)
}

// Action sets the action attribute, the URL a <form> submits to.
func Action(url string) mx.Attrib { return mx.NewAttrib("action", url) }

// Actionf is like Action but builds the value with a fmt format string.
func Actionf(urlFmt string, a ...any) mx.Attrib { return mx.NewAttribf("action", urlFmt, a...) }

// Align sets the align attribute. Obsolete presentational attribute; use CSS instead.
func Align(value string) mx.Attrib { return mx.NewAttrib("align", value) }

// Alignf is like Align but builds the value with a fmt format string.
func Alignf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("align", valueFmt, a...)
}

// Allow sets the allow attribute, the permissions policy applied to an <iframe>'s content.
func Allow(value string) mx.Attrib { return mx.NewAttrib("allow", value) }

// Allowf is like Allow but builds the value with a fmt format string.
func Allowf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("allow", valueFmt, a...)
}

// Alpha is the boolean alpha attribute enabling the alpha (opacity) channel on a color <input>.
const Alpha = BoolAttrib("alpha")

// Alt sets the alt attribute, the alternative text describing an image or area.
func Alt(text string) mx.Attrib { return mx.NewAttrib("alt", text) }

// Altf is like Alt but builds the value with a fmt format string.
func Altf(textFmt string, a ...any) mx.Attrib { return mx.NewAttribf("alt", textFmt, a...) }

// Async is the boolean async attribute marking a <script> to be fetched and executed asynchronously.
const Async = BoolAttrib("async")

// AutoComplete sets the autocomplete attribute, controlling automated value completion;
// with no tokens it defaults to AutoCompleteOn.
func AutoComplete(tokens ...string) mx.Attrib {
	if len(tokens) == 0 {
		return AutoCompleteOn
	}
	return mx.NewAttrib("autocomplete", strings.Join(tokens, " "))
}

// AutoCompleteOn sets the autocomplete attribute to "on", enabling automated value completion.
const AutoCompleteOn = mx.ConstAttrib("autocomplete=on")

// AutoCompleteOff sets the autocomplete attribute to "off", disabling automated value completion.
const AutoCompleteOff = mx.ConstAttrib("autocomplete=off")

// AutoFocus is the boolean autofocus attribute focusing the element on page load.
const AutoFocus = BoolAttrib("autofocus")

// AutoPlay is the boolean autoplay attribute making media start playing as soon as it can.
const AutoPlay = BoolAttrib("autoplay")

// Background sets the background attribute. Obsolete presentational attribute; use CSS instead.
func Background(style string) mx.Attrib { return mx.NewAttrib("background", style) }

// BGColor sets the bgcolor attribute. Obsolete presentational attribute; use CSS instead.
func BGColor(color string) mx.Attrib { return mx.NewAttrib("bgcolor", color) }

// Border sets the border attribute. Obsolete presentational attribute; use CSS instead.
func Border(value string) mx.Attrib { return mx.NewAttrib("border", value) }

// Borderf is like Border but builds the value with a fmt format string.
func Borderf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("border", valueFmt, a...)
}

// CharSet sets the charset attribute, the character encoding (e.g. on <meta> or <script>).
func CharSet(value string) mx.Attrib { return mx.NewAttrib("charset", value) }

// Checked is the boolean checked attribute marking a checkbox or radio <input> as selected.
const Checked = BoolAttrib("checked")

// CiteAttr sets the cite attribute, a URL referencing the source of a quotation or edit.
func CiteAttr(value string) mx.Attrib { return mx.NewAttrib("cite", value) }

// CiteAttrf is like CiteAttr but builds the value with a fmt format string.
func CiteAttrf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("cite", valueFmt, a...)
}

// Class sets the class attribute, a space-separated list of CSS class names.
func Class(classes ...string) mx.Attrib { return mx.NewAttrib("class", strings.Join(classes, " ")) }

// Classf is like Class but builds the value with a fmt format string.
func Classf(classFmt string, a ...any) mx.Attrib { return mx.NewAttribf("class", classFmt, a...) }

// Color sets the color attribute. Obsolete presentational attribute; use CSS instead.
func Color(value string) mx.Attrib { return mx.NewAttrib("color", value) }

// Colorf is like Color but builds the value with a fmt format string.
func Colorf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("color", valueFmt, a...)
}

// Cols sets the cols attribute, the visible width of a <textarea> in characters.
func Cols(numChars int) mx.Attrib { return mx.NewAttrib("cols", strconv.Itoa(numChars)) }

// ColSpan sets the colspan attribute, the number of columns a table cell spans.
func ColSpan(numCols int) mx.Attrib { return mx.NewAttrib("colspan", strconv.Itoa(numCols)) }

// Command sets the command attribute of a <button>, the action it invokes on its CommandFor target.
// Pass a built-in keyword (the Command* constants) or an author-defined custom command, which
// must start with two hyphens (e.g. Command("--rotate")).
func Command(value string) mx.Attrib { return mx.NewAttrib("command", value) }

const (
	// CommandTogglePopover toggles the targeted popover between shown and hidden.
	CommandTogglePopover = mx.ConstAttrib("command=toggle-popover")
	// CommandShowPopover shows the targeted popover.
	CommandShowPopover = mx.ConstAttrib("command=show-popover")
	// CommandHidePopover hides the targeted popover.
	CommandHidePopover = mx.ConstAttrib("command=hide-popover")
	// CommandShowModal opens the targeted <dialog> as a modal.
	CommandShowModal = mx.ConstAttrib("command=show-modal")
	// CommandClose closes the targeted <dialog>.
	CommandClose = mx.ConstAttrib("command=close")
	// CommandRequestClose requests closing the targeted <dialog>, firing a cancelable close request.
	CommandRequestClose = mx.ConstAttrib("command=request-close")
)

// CommandFor sets the commandfor attribute of a <button>, the id of the element its Command acts on.
func CommandFor(id string) mx.Attrib { return mx.NewAttrib("commandfor", id) }

// ContentAttr sets the content attribute, the value associated with a <meta> name or http-equiv.
func ContentAttr(text string) mx.Attrib { return mx.NewAttrib("content", text) }

// Controls is the boolean controls attribute showing the browser's built-in media playback controls.
const Controls = BoolAttrib("controls")

// Coords sets the coords attribute of an image-map <area>, the comma-separated coordinates of its shape.
func Coords(coords ...float64) mx.Attrib {
	var b strings.Builder
	for i, coord := range coords {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.FormatFloat(coord, 'f', -1, 64))
	}
	return mx.NewAttrib("coords", b.String())
}

// CSP sets the csp attribute of an <iframe>, the Content Security Policy enforced on the
// embedded document. Experimental.
func CSP(value string) mx.Attrib { return mx.NewAttrib("csp", value) }

// CSPf is like CSP but builds the value with a fmt format string.
func CSPf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("csp", valueFmt, a...)
}

// DataAttr sets a custom data-* attribute named "data-"+name with the given value.
func DataAttr(name, value string) mx.Attrib { return mx.NewAttrib("data-"+name, value) }

// DataAttrf is like DataAttr but builds the value with a fmt format string.
func DataAttrf(name, valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("data-"+name, valueFmt, a...)
}

// Datetime sets the datetime attribute, the machine-readable date/time of a <time>, <ins> or <del>.
func Datetime(value string) mx.Attrib { return mx.NewAttrib("datetime", value) }

// Datetimef is like Datetime but builds the value with a fmt format string.
func Datetimef(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("datetime", valueFmt, a...)
}

// Default is the boolean default attribute marking a <track> as the default text track.
const Default = BoolAttrib("default")

// Defer is the boolean defer attribute deferring <script> execution until the document is parsed.
const Defer = BoolAttrib("defer")

// DirName sets the dirname attribute, naming the field that submits the control's text directionality.
func DirName(name string) mx.Attrib { return mx.NewAttrib("dirname", name) }

// Disabled is the boolean disabled attribute marking a form control as non-interactive and excluded from submission.
const Disabled = BoolAttrib("disabled")

// Download sets the download attribute, prompting to save the link target under the given filename.
func Download(filename string) mx.Attrib { return mx.NewAttrib("download", filename) }

// Draggable sets the draggable attribute (enumerated true/false) controlling whether the element can be dragged.
func Draggable(value bool) mx.Attrib { return mx.NewAttrib("draggable", strconv.FormatBool(value)) }

// For sets the for attribute, associating a <label> or <output> with the control's id.
func For(id string) mx.Attrib { return mx.NewAttrib("for", id) }

// FormAttr sets the form attribute, associating a control with the <form> of the given id.
func FormAttr(formID string) mx.Attrib { return mx.NewAttrib("form", formID) }

// FormAction sets the formaction attribute, overriding the form's action URL for a submit control.
func FormAction(url string) mx.Attrib { return mx.NewAttrib("formaction", url) }

// FormNoValidate is the boolean formnovalidate attribute skipping form validation when submitting via this control.
const FormNoValidate = BoolAttrib("formnovalidate")

// FormTarget sets the formtarget attribute, overriding the form's target browsing context for a submit control.
func FormTarget(value string) mx.Attrib { return mx.NewAttrib("formtarget", value) }

// FormTargetf is like FormTarget but builds the value with a fmt format string.
func FormTargetf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("formtarget", valueFmt, a...)
}

// Headers sets the headers attribute, the ids of the header cells a table cell is associated with.
func Headers(headerCellIDs ...string) mx.Attrib {
	return mx.NewAttrib("headers", strings.Join(headerCellIDs, " "))
}

// Height sets the height attribute, the rendered height of an element in pixels.
func Height[T AttribValue](value T) mx.Attrib {
	return mx.NewAttrib("height", attribValueString(value))
}

// Heightf is like Height but builds the value with a fmt format string.
func Heightf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("height", valueFmt, a...)
}

// HeightPx sets the height attribute to the given value with a "px" unit suffix.
func HeightPx(pixels float64) mx.Attrib {
	return Height(strconv.FormatFloat(pixels, 'f', -1, 64) + "px")
}

// HeightEm sets the height attribute to the given value with an "em" unit suffix.
func HeightEm(ems float64) mx.Attrib {
	return Height(strconv.FormatFloat(ems, 'f', -1, 64) + "em")
}

// Hidden is the boolean hidden attribute hiding the element from rendering.
const Hidden = BoolAttrib("hidden")

// HiddenUntilFound sets the hidden attribute to "until-found", hiding content that can still be
// revealed by find-in-page and fragment navigation.
const HiddenUntilFound = mx.ConstAttrib("hidden=until-found")

// High sets the high attribute of a <meter>, the lower bound of the high (upper) value range.
func High(limit float64) mx.Attrib {
	return mx.NewAttrib("high", strconv.FormatFloat(limit, 'f', -1, 64))
}

// HRef sets the href attribute, the URL a link or hyperlink points to.
func HRef(url string) mx.Attrib { return mx.NewAttrib("href", url) }

// HRefLang sets the hreflang attribute, the language of the linked resource.
func HRefLang(value string) mx.Attrib { return mx.NewAttrib("hreflang", value) }

// HRefLangf is like HRefLang but builds the value with a fmt format string.
func HRefLangf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("hreflang", valueFmt, a...)
}

// ID sets the id attribute, the element's unique identifier within the document.
func ID(value string) mx.Attrib { return mx.NewAttrib("id", value) }

// IDf is like ID but builds the value with a fmt format string.
func IDf(valueFmt string, a ...any) mx.Attrib { return mx.NewAttribf("id", valueFmt, a...) }

// imagesizes ?
// imagesrcset ?

// Inert is the boolean inert attribute making the element and its subtree non-interactive.
const Inert = BoolAttrib("inert")

// Integrity sets the integrity attribute, the Subresource Integrity hash for a <script> or <link>.
func Integrity(value string) mx.Attrib { return mx.NewAttrib("integrity", value) }

// Integrityf is like Integrity but builds the value with a fmt format string.
func Integrityf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("integrity", valueFmt, a...)
}

// IntrinsicSize sets the intrinsicsize attribute. Non-standard (a removed Chrome experiment).
func IntrinsicSize(value string) mx.Attrib { return mx.NewAttrib("intrinsicsize", value) }

// IntrinsicSizef is like IntrinsicSize but builds the value with a fmt format string.
func IntrinsicSizef(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("intrinsicsize", valueFmt, a...)
}

// IsMap is the boolean ismap attribute marking an image inside an <a> as a server-side image map.
const IsMap = BoolAttrib("ismap")

// ItemID sets the itemid attribute, the global identifier of a microdata item.
func ItemID(url string) mx.Attrib { return mx.NewAttrib("itemid", url) }

// ItemProp sets the itemprop attribute, the microdata property names the element provides.
func ItemProp(props ...string) mx.Attrib { return mx.NewAttrib("itemprop", strings.Join(props, " ")) }

// ItemRef sets the itemref attribute, the ids of microdata properties located elsewhere in the document.
func ItemRef(ids ...string) mx.Attrib { return mx.NewAttrib("itemref", strings.Join(ids, " ")) }

// ItemScope is the boolean itemscope attribute declaring a new microdata item.
const ItemScope = BoolAttrib("itemscope")

// ItemType sets the itemtype attribute, the vocabulary URLs defining a microdata item's properties.
func ItemType(urls ...string) mx.Attrib { return mx.NewAttrib("itemtype", strings.Join(urls, " ")) }

// LabelAttr sets the label attribute, the user-visible label of an <option> or <optgroup>.
func LabelAttr(value string) mx.Attrib { return mx.NewAttrib("label", value) }

// LabelAttrf is like LabelAttr but builds the value with a fmt format string.
func LabelAttrf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("label", valueFmt, a...)
}

// Lang sets the lang attribute, the language of the element's content.
func Lang(value string) mx.Attrib { return mx.NewAttrib("lang", value) }

// Langf is like Lang but builds the value with a fmt format string.
func Langf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("lang", valueFmt, a...)
}

// Language sets the language attribute of a <script>. Obsolete; use Type instead.
func Language(value string) mx.Attrib { return mx.NewAttrib("language", value) }

// Languagef is like Language but builds the value with a fmt format string.
func Languagef(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("language", valueFmt, a...)
}

// List sets the list attribute, the id of the <datalist> providing autocomplete options for an <input>.
func List(id string) mx.Attrib { return mx.NewAttrib("list", id) }

// Loop is the boolean loop attribute making media restart from the beginning when it ends.
const Loop = BoolAttrib("loop")

// Low sets the low attribute of a <meter>, the upper bound of the low (lower) value range.
func Low(limit float64) mx.Attrib {
	return mx.NewAttrib("low", strconv.FormatFloat(limit, 'f', -1, 64))
}

// Max sets the max attribute, the maximum allowed value of a control or range.
// The value may be a number or a string such as a date or time (see [AttribValue]).
func Max[T AttribValue](value T) mx.Attrib { return mx.NewAttrib("max", attribValueString(value)) }

// Maxf is like Max but builds the value with a fmt format string.
func Maxf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("max", valueFmt, a...)
}

// MaxLength sets the maxlength attribute, the maximum number of characters a control accepts.
func MaxLength(length int) mx.Attrib { return mx.NewAttrib("maxlength", strconv.Itoa(length)) }

// Media sets the media attribute, the media query the linked resource applies to.
func Media(query string) mx.Attrib { return mx.NewAttrib("media", query) }

// Min sets the min attribute, the minimum allowed value of a control or range.
// The value may be a number or a string such as a date or time (see [AttribValue]).
func Min[T AttribValue](value T) mx.Attrib { return mx.NewAttrib("min", attribValueString(value)) }

// Minf is like Min but builds the value with a fmt format string.
func Minf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("min", valueFmt, a...)
}

// MinLength sets the minlength attribute, the minimum number of characters a control accepts.
func MinLength(length int) mx.Attrib { return mx.NewAttrib("minlength", strconv.Itoa(length)) }

// Multiple is the boolean multiple attribute allowing multiple values in a <select>, file or email <input>.
const Multiple = BoolAttrib("multiple")

// Muted is the boolean muted attribute making media play with audio muted by default.
const Muted = BoolAttrib("muted")

// Name sets the name attribute, the name used when submitting a form control's value.
func Name(value string) mx.Attrib { return mx.NewAttrib("name", value) }

// Namef is like Name but builds the value with a fmt format string.
func Namef(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("name", valueFmt, a...)
}

// NoModule is the boolean nomodule attribute preventing a <script> from running in module-supporting browsers.
const NoModule = BoolAttrib("nomodule")

// Nonce sets the nonce attribute, the cryptographic nonce used by Content-Security-Policy.
func Nonce(value string) mx.Attrib { return mx.NewAttrib("nonce", value) }

// Noncef is like Nonce but builds the value with a fmt format string.
func Noncef(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("nonce", valueFmt, a...)
}

// NoValidate is the boolean novalidate attribute disabling validation when a <form> is submitted.
const NoValidate = BoolAttrib("novalidate")

// Open is the boolean open attribute making a <details> or <dialog> initially open.
const Open = BoolAttrib("open")

// Optimum sets the optimum attribute of a <meter>, the value considered the optimal point.
func Optimum(value float64) mx.Attrib {
	return mx.NewAttrib("optimum", strconv.FormatFloat(value, 'f', -1, 64))
}

// Pattern sets the pattern attribute, a regular expression an <input>'s value must match.
func Pattern(value string) mx.Attrib { return mx.NewAttrib("pattern", value) }

// Patternf is like Pattern but builds the value with a fmt format string.
func Patternf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("pattern", valueFmt, a...)
}

// Ping sets the ping attribute, the space-separated URLs notified when an <a> or <area> is followed.
func Ping(value string) mx.Attrib { return mx.NewAttrib("ping", value) }

// Pingf is like Ping but builds the value with a fmt format string.
func Pingf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("ping", valueFmt, a...)
}

// Placeholder sets the placeholder attribute, the hint text shown in an empty input.
func Placeholder(value string) mx.Attrib { return mx.NewAttrib("placeholder", value) }

// Placeholderf is like Placeholder but builds the value with a fmt format string.
func Placeholderf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("placeholder", valueFmt, a...)
}

// PlaysInline is the boolean playsinline attribute hinting that video should play inline rather than fullscreen.
const PlaysInline = BoolAttrib("playsinline")

// PopoverTarget sets the popovertarget attribute of a <button> or <input>, the id of the popover element it controls.
func PopoverTarget(id string) mx.Attrib { return mx.NewAttrib("popovertarget", id) }

// Poster sets the poster attribute, the image shown for a <video> before playback begins.
func Poster(value string) mx.Attrib { return mx.NewAttrib("poster", value) }

// Posterf is like Poster but builds the value with a fmt format string.
func Posterf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("poster", valueFmt, a...)
}

// Readonly is the boolean readonly attribute making a control's value non-editable.
const Readonly = BoolAttrib("readonly")

// Rel sets the rel attribute, the space-separated link relationship keywords of an <a>, <link> or <area>.
func Rel(keywords ...string) mx.Attrib { return mx.NewAttrib("rel", strings.Join(keywords, " ")) }

// Required is the boolean required attribute making a form control mandatory before submission.
const Required = BoolAttrib("required")

// Reversed is the boolean reversed attribute numbering an ordered <ol> list in descending order.
const Reversed = BoolAttrib("reversed")

// Role sets the role attribute, the ARIA role of the element.
func Role(value string) mx.Attrib { return mx.NewAttrib("role", value) }

// Rolef is like Role but builds the value with a fmt format string.
func Rolef(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("role", valueFmt, a...)
}

// Rows sets the rows attribute, the visible number of text lines in a <textarea>.
func Rows(numChars int) mx.Attrib { return mx.NewAttrib("rows", strconv.Itoa(numChars)) }

// RowSpan sets the rowspan attribute, the number of rows a table cell spans.
func RowSpan(numRows int) mx.Attrib { return mx.NewAttrib("rowspan", strconv.Itoa(numRows)) }

// Sandbox sets the sandbox attribute of an <iframe>, the space-separated token list restricting the embedded content.
func Sandbox(value string) mx.Attrib { return mx.NewAttrib("sandbox", value) }

// Sandboxf is like Sandbox but builds the value with a fmt format string.
func Sandboxf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("sandbox", valueFmt, a...)
}

// Scoped was removed from the HTML spec

// Selected is the boolean selected attribute marking an <option> as initially selected.
const Selected = BoolAttrib("selected")

// Size sets the size attribute, the visible width of an <input> or number of
// rows of a <select>. The value may be a number or a string (see [AttribValue]).
func Size[T AttribValue](value T) mx.Attrib { return mx.NewAttrib("size", attribValueString(value)) }

// Sizef is like Size but builds the value with a fmt format string.
func Sizef(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("size", valueFmt, a...)
}

// Sizes sets the sizes attribute, the source sizes used to pick a responsive image candidate.
func Sizes(sourceSizes ...string) mx.Attrib {
	return mx.NewAttrib("sizes", strings.Join(sourceSizes, ","))
}

// SlotAttr sets the slot attribute, the name of the shadow-DOM slot the element is assigned to.
func SlotAttr(value string) mx.Attrib { return mx.NewAttrib("slot", value) }

// SlotAttrf is like SlotAttr but builds the value with a fmt format string.
func SlotAttrf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("slot", valueFmt, a...)
}

// SpanAttr sets the span attribute, the number of columns a <col> or <colgroup>
// spans. The value may be a number or a string (see [AttribValue]).
func SpanAttr[T AttribValue](value T) mx.Attrib {
	return mx.NewAttrib("span", attribValueString(value))
}

// SpanAttrf is like SpanAttr but builds the value with a fmt format string.
func SpanAttrf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("span", valueFmt, a...)
}

// Src sets the src attribute, the URL of the resource embedded by the element.
func Src(url string) mx.Attrib { return mx.NewAttrib("src", url) }

// SrcDoc sets the srcdoc attribute, the inline HTML rendered as an <iframe>'s document.
func SrcDoc(value string) mx.Attrib { return mx.NewAttrib("srcdoc", value) }

// SrcDocf is like SrcDoc but builds the value with a fmt format string.
func SrcDocf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("srcdoc", valueFmt, a...)
}

// SrcLang sets the srclang attribute, the language of a <track>'s text track.
func SrcLang(value string) mx.Attrib { return mx.NewAttrib("srclang", value) }

// SrcLangf is like SrcLang but builds the value with a fmt format string.
func SrcLangf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("srclang", valueFmt, a...)
}

// SrcSet sets the srcset attribute, the candidate image sources for responsive images.
func SrcSet(sources ...string) mx.Attrib {
	return mx.NewAttrib("srcset", strings.Join(sources, ","))
}

// Start sets the start attribute, the starting number of an ordered <ol> list.
// The value may be a number or a string (see [AttribValue]).
func Start[T AttribValue](value T) mx.Attrib { return mx.NewAttrib("start", attribValueString(value)) }

// Startf is like Start but builds the value with a fmt format string.
func Startf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("start", valueFmt, a...)
}

// Step sets the step attribute, the granularity of allowed values for a numeric
// or date <input>. The value may be a number or the keyword "any" (see [AttribValue]).
func Step[T AttribValue](value T) mx.Attrib { return mx.NewAttrib("step", attribValueString(value)) }

// Stepf is like Step but builds the value with a fmt format string.
func Stepf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("step", valueFmt, a...)
}

// Style sets the style attribute, inline CSS declarations applied to the element.
func Style(value string) mx.Attrib { return mx.NewAttrib("style", value) }

// Stylef is like Style but builds the value with a fmt format string.
func Stylef(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("style", valueFmt, a...)
}

// TabIndex sets the tabindex attribute, controlling the element's keyboard tab
// order. The value may be a number or a string (see [AttribValue]).
func TabIndex[T AttribValue](value T) mx.Attrib {
	return mx.NewAttrib("tabindex", attribValueString(value))
}

// TabIndexf is like TabIndex but builds the value with a fmt format string.
func TabIndexf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("tabindex", valueFmt, a...)
}

// Target sets the target attribute, the browsing context where a link or form opens.
func Target(value string) mx.Attrib { return mx.NewAttrib("target", value) }

// Targetf is like Target but builds the value with a fmt format string.
func Targetf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("target", valueFmt, a...)
}

// TargetSelf sets the target attribute to "_self", opening in the current browsing context.
const TargetSelf = mx.ConstAttrib("target=_self")

// TargetBlank sets the target attribute to "_blank", opening in a new browsing context.
const TargetBlank = mx.ConstAttrib("target=_blank")

// TargetParent sets the target attribute to "_parent", opening in the parent browsing context.
const TargetParent = mx.ConstAttrib("target=_parent")

// TargetTop sets the target attribute to "_top", opening in the topmost browsing context.
const TargetTop = mx.ConstAttrib("target=_top")

// TargetUnfencedTop sets the target attribute to "_unfencedTop", opening at the top of a fenced frame.
const TargetUnfencedTop = mx.ConstAttrib("target=_unfencedTop")

// Title sets the title attribute, the advisory tooltip text for the element.
func Title(value string) mx.Attrib { return mx.NewAttrib("title", value) }

// Titlef is like Title but builds the value with a fmt format string.
func Titlef(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("title", valueFmt, a...)
}

// Type sets the type attribute, the kind of control, link, button or media (e.g. an <input> type).
func Type(value string) mx.Attrib { return mx.NewAttrib("type", value) }

// Typef is like Type but builds the value with a fmt format string.
func Typef(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("type", valueFmt, a...)
}

// UseMap sets the usemap attribute, associating an image with a <map> via a hash-name fragment.
func UseMap(partialURL string) mx.Attrib { return mx.NewAttrib("usemap", partialURL) }

// Value sets the value attribute, the initial or current value of a control or
// option. The value may be a string or a number (see [AttribValue]), so a number
// for a <progress>, <meter> or numeric <input> works as well as a text value.
func Value[T AttribValue](value T) mx.Attrib { return mx.NewAttrib("value", attribValueString(value)) }

// Valuef is like Value but builds the value with a fmt format string.
func Valuef(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("value", valueFmt, a...)
}

// Width sets the width attribute, the rendered width of an element in pixels.
func Width[T AttribValue](value T) mx.Attrib { return mx.NewAttrib("width", attribValueString(value)) }

// Widthf is like Width but builds the value with a fmt format string.
func Widthf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("width", valueFmt, a...)
}

// WidthPx sets the width attribute to the given value with a "px" unit suffix.
func WidthPx(pixels float64) mx.Attrib {
	return Width(strconv.FormatFloat(pixels, 'f', -1, 64) + "px")
}

// WidthEm sets the width attribute to the given value with an "em" unit suffix.
func WidthEm(ems float64) mx.Attrib {
	return Width(strconv.FormatFloat(ems, 'f', -1, 64) + "em")
}

// Event handlers, see https://html.spec.whatwg.org/multipage/indices.html#events-2

// OnAfterPrint `afterprint` event handler for Window object (body element)
func OnAfterPrint(execute string) mx.Attrib { return mx.NewAttrib("onafterprint", execute) }

// OnAuxClick `auxclick` event handler (all HTML elements)
func OnAuxClick(execute string) mx.Attrib { return mx.NewAttrib("onauxclick", execute) }

// OnBeforeInput `beforeinput` event handler (all HTML elements)
func OnBeforeInput(execute string) mx.Attrib { return mx.NewAttrib("onbeforeinput", execute) }

// OnBeforeMatch `beforematch` event handler (all HTML elements)
func OnBeforeMatch(execute string) mx.Attrib { return mx.NewAttrib("onbeforematch", execute) }

// OnBeforePrint `beforeprint` event handler for Window object (body element)
func OnBeforePrint(execute string) mx.Attrib { return mx.NewAttrib("onbeforeprint", execute) }

// OnBeforeUnload `beforeunload` event handler for Window object (body element)
func OnBeforeUnload(execute string) mx.Attrib { return mx.NewAttrib("onbeforeunload", execute) }

// OnBeforeToggle `beforetoggle` event handler (all HTML elements)
func OnBeforeToggle(execute string) mx.Attrib { return mx.NewAttrib("onbeforetoggle", execute) }

// OnBlur `blur` event handler (all HTML elements)
func OnBlur(execute string) mx.Attrib { return mx.NewAttrib("onblur", execute) }

// OnCancel `cancel` event handler (all HTML elements)
func OnCancel(execute string) mx.Attrib { return mx.NewAttrib("oncancel", execute) }

// OnCanplay `canplay` event handler (all HTML elements)
func OnCanplay(execute string) mx.Attrib { return mx.NewAttrib("oncanplay", execute) }

// OnCanPlayThrough `canplaythrough` event handler (all HTML elements)
func OnCanPlayThrough(execute string) mx.Attrib { return mx.NewAttrib("oncanplaythrough", execute) }

// OnChange `change` event handler (all HTML elements)
func OnChange(execute string) mx.Attrib { return mx.NewAttrib("onchange", execute) }

// OnClick `click` event handler (all HTML elements)
func OnClick(execute string) mx.Attrib { return mx.NewAttrib("onclick", execute) }

// OnClose `close` event handler (all HTML elements)
func OnClose(execute string) mx.Attrib { return mx.NewAttrib("onclose", execute) }

// OnContextLost `contextlost` event handler (all HTML elements)
func OnContextLost(execute string) mx.Attrib { return mx.NewAttrib("oncontextlost", execute) }

// OnContextMenu `contextmenu` event handler (all HTML elements)
func OnContextMenu(execute string) mx.Attrib { return mx.NewAttrib("oncontextmenu", execute) }

// OnContextRestored `contextrestored` event handler (all HTML elements)
func OnContextRestored(execute string) mx.Attrib { return mx.NewAttrib("oncontextrestored", execute) }

// OnCopy `copy` event handler (all HTML elements)
func OnCopy(execute string) mx.Attrib { return mx.NewAttrib("oncopy", execute) }

// OnCueChange `cuechange` event handler (all HTML elements)
func OnCueChange(execute string) mx.Attrib { return mx.NewAttrib("oncuechange", execute) }

// OnCut `cut` event handler (all HTML elements)
func OnCut(execute string) mx.Attrib { return mx.NewAttrib("oncut", execute) }

// OnDblClick `dblclick` event handler (all HTML elements)
func OnDblClick(execute string) mx.Attrib { return mx.NewAttrib("ondblclick", execute) }

// OnDrag `drag` event handler (all HTML elements)
func OnDrag(execute string) mx.Attrib { return mx.NewAttrib("ondrag", execute) }

// OnDragEnd `dragend` event handler (all HTML elements)
func OnDragEnd(execute string) mx.Attrib { return mx.NewAttrib("ondragend", execute) }

// OnDragEnter `dragenter` event handler (all HTML elements)
func OnDragEnter(execute string) mx.Attrib { return mx.NewAttrib("ondragenter", execute) }

// OnDragLeave `dragleave` event handler (all HTML elements)
func OnDragLeave(execute string) mx.Attrib { return mx.NewAttrib("ondragleave", execute) }

// OnDragOver `dragover` event handler (all HTML elements)
func OnDragOver(execute string) mx.Attrib { return mx.NewAttrib("ondragover", execute) }

// OnDragStart `dragstart` event handler (all HTML elements)
func OnDragStart(execute string) mx.Attrib { return mx.NewAttrib("ondragstart", execute) }

// OnDrop `drop` event handler (all HTML elements)
func OnDrop(execute string) mx.Attrib { return mx.NewAttrib("ondrop", execute) }

// OnDurationChange `durationchange` event handler (all HTML elements)
func OnDurationChange(execute string) mx.Attrib { return mx.NewAttrib("ondurationchange", execute) }

// OnEmptied `emptied` event handler (all HTML elements)
func OnEmptied(execute string) mx.Attrib { return mx.NewAttrib("onemptied", execute) }

// OnEnded `ended` event handler (all HTML elements)
func OnEnded(execute string) mx.Attrib { return mx.NewAttrib("onended", execute) }

// OnError `error` event handler (all HTML elements)
func OnError(execute string) mx.Attrib { return mx.NewAttrib("onerror", execute) }

// OnFocus `focus` event handler (all HTML elements)
func OnFocus(execute string) mx.Attrib { return mx.NewAttrib("onfocus", execute) }

// OnFormData `formdata` event handler (all HTML elements)
func OnFormData(execute string) mx.Attrib { return mx.NewAttrib("onformdata", execute) }

// OnHashChange `hashchange` event handler for Window object (body element)
func OnHashChange(execute string) mx.Attrib { return mx.NewAttrib("onhashchange", execute) }

// OnInput `input` event handler (all HTML elements)
func OnInput(execute string) mx.Attrib { return mx.NewAttrib("oninput", execute) }

// OnInvalid `invalid` event handler (all HTML elements)
func OnInvalid(execute string) mx.Attrib { return mx.NewAttrib("oninvalid", execute) }

// OnKeyDown `keydown` event handler (all HTML elements)
func OnKeyDown(execute string) mx.Attrib { return mx.NewAttrib("onkeydown", execute) }

// OnKeyPress `keypress` event handler (all HTML elements)
func OnKeyPress(execute string) mx.Attrib { return mx.NewAttrib("onkeypress", execute) }

// OnKeyUp `keyup` event handler (all HTML elements)
func OnKeyUp(execute string) mx.Attrib { return mx.NewAttrib("onkeyup", execute) }

// OnLanguageChange `languagechange` event handler for Window object (body element)
func OnLanguageChange(execute string) mx.Attrib { return mx.NewAttrib("onlanguagechange", execute) }

// OnLoad `load` event handler (all HTML elements)
func OnLoad(execute string) mx.Attrib { return mx.NewAttrib("onload", execute) }

// OnLoadedData `loadeddata` event handler (all HTML elements)
func OnLoadedData(execute string) mx.Attrib { return mx.NewAttrib("onloadeddata", execute) }

// OnLoadedMetadata `loadedmetadata` event handler (all HTML elements)
func OnLoadedMetadata(execute string) mx.Attrib { return mx.NewAttrib("onloadedmetadata", execute) }

// OnLoadStart `loadstart` event handler (all HTML elements)
func OnLoadStart(execute string) mx.Attrib { return mx.NewAttrib("onloadstart", execute) }

// OnMessage `message` event handler for Window object (body element)
func OnMessage(execute string) mx.Attrib { return mx.NewAttrib("onmessage", execute) }

// OnMessageError `messageerror` event handler for Window object (body element)
func OnMessageError(execute string) mx.Attrib { return mx.NewAttrib("onmessageerror", execute) }

// OnMouseDown `mousedown` event handler (all HTML elements)
func OnMouseDown(execute string) mx.Attrib { return mx.NewAttrib("onmousedown", execute) }

// OnMouseEnter `mouseenter` event handler (all HTML elements)
func OnMouseEnter(execute string) mx.Attrib { return mx.NewAttrib("onmouseenter", execute) }

// OnMouseLeave `mouseleave` event handler (all HTML elements)
func OnMouseLeave(execute string) mx.Attrib { return mx.NewAttrib("onmouseleave", execute) }

// OnMouseMove `mousemove` event handler (all HTML elements)
func OnMouseMove(execute string) mx.Attrib { return mx.NewAttrib("onmousemove", execute) }

// OnMouseOut `mouseout` event handler (all HTML elements)
func OnMouseOut(execute string) mx.Attrib { return mx.NewAttrib("onmouseout", execute) }

// OnMouseOver `mouseover` event handler (all HTML elements)
func OnMouseOver(execute string) mx.Attrib { return mx.NewAttrib("onmouseover", execute) }

// OnMouseUp `mouseup` event handler (all HTML elements)
func OnMouseUp(execute string) mx.Attrib { return mx.NewAttrib("onmouseup", execute) }

// OnOffline `offline` event handler for Window object (body element)
func OnOffline(execute string) mx.Attrib { return mx.NewAttrib("onoffline", execute) }

// OnOnline `online` event handler for Window object (body element)
func OnOnline(execute string) mx.Attrib { return mx.NewAttrib("ononline", execute) }

// OnPageHide `pagehide` event handler for Window object (body element)
func OnPageHide(execute string) mx.Attrib { return mx.NewAttrib("onpagehide", execute) }

// OnPageReveal `pagereveal` event handler for Window object (body element)
func OnPageReveal(execute string) mx.Attrib { return mx.NewAttrib("onpagereveal", execute) }

// OnPageShow `pageshow` event handler for Window object (body element)
func OnPageShow(execute string) mx.Attrib { return mx.NewAttrib("onpageshow", execute) }

// OnPageSwap `pageswap` event handler for Window object (body element)
func OnPageSwap(execute string) mx.Attrib { return mx.NewAttrib("onpageswap", execute) }

// OnPaste `paste` event handler (all HTML elements)
func OnPaste(execute string) mx.Attrib { return mx.NewAttrib("onpaste", execute) }

// OnPause `pause` event handler (all HTML elements)
func OnPause(execute string) mx.Attrib { return mx.NewAttrib("onpause", execute) }

// OnPlay `play` event handler (all HTML elements)
func OnPlay(execute string) mx.Attrib { return mx.NewAttrib("onplay", execute) }

// OnPlaying `playing` event handler (all HTML elements)
func OnPlaying(execute string) mx.Attrib { return mx.NewAttrib("onplaying", execute) }

// OnPopState `popstate` event handler for Window object (body element)
func OnPopState(execute string) mx.Attrib { return mx.NewAttrib("onpopstate", execute) }

// OnProgress `progress` event handler (all HTML elements)
func OnProgress(execute string) mx.Attrib { return mx.NewAttrib("onprogress", execute) }

// OnRateChange `ratechange` event handler (all HTML elements)
func OnRateChange(execute string) mx.Attrib { return mx.NewAttrib("onratechange", execute) }

// OnReset `reset` event handler (all HTML elements)
func OnReset(execute string) mx.Attrib { return mx.NewAttrib("onreset", execute) }

// OnResize `resize` event handler (all HTML elements)
func OnResize(execute string) mx.Attrib { return mx.NewAttrib("onresize", execute) }

// OnRejectionHandled `rejectionhandled` event handler for Window object (body element)
func OnRejectionHandled(execute string) mx.Attrib { return mx.NewAttrib("onrejectionhandled", execute) }

// OnScroll `scroll` event handler (all HTML elements)
func OnScroll(execute string) mx.Attrib { return mx.NewAttrib("onscroll", execute) }

// OnScrollEnd `scrollend` event handler (all HTML elements)
func OnScrollEnd(execute string) mx.Attrib { return mx.NewAttrib("onscrollend", execute) }

// OnSecurityPolicyViolation `securitypolicyviolation` event handler (all HTML elements)
func OnSecurityPolicyViolation(execute string) mx.Attrib {
	return mx.NewAttrib("onsecuritypolicyviolation", execute)
}

// OnSeeked `seeked` event handler (all HTML elements)
func OnSeeked(execute string) mx.Attrib { return mx.NewAttrib("onseeked", execute) }

// OnSeeking `seeking` event handler (all HTML elements)
func OnSeeking(execute string) mx.Attrib { return mx.NewAttrib("onseeking", execute) }

// OnSelect `select` event handler (all HTML elements)
func OnSelect(execute string) mx.Attrib { return mx.NewAttrib("onselect", execute) }

// OnSlotChange `slotchange` event handler (all HTML elements)
func OnSlotChange(execute string) mx.Attrib { return mx.NewAttrib("onslotchange", execute) }

// OnStalled `stalled` event handler (all HTML elements)
func OnStalled(execute string) mx.Attrib { return mx.NewAttrib("onstalled", execute) }

// OnStorage `storage` event handler for Window object (body element)
func OnStorage(execute string) mx.Attrib { return mx.NewAttrib("onstorage", execute) }

// OnSubmit `submit` event handler (all HTML elements)
func OnSubmit(execute string) mx.Attrib { return mx.NewAttrib("onsubmit", execute) }

// OnSuspend `suspend` event handler (all HTML elements)
func OnSuspend(execute string) mx.Attrib { return mx.NewAttrib("onsuspend", execute) }

// OnTimeUpdate `timeupdate` event handler (all HTML elements)
func OnTimeUpdate(execute string) mx.Attrib { return mx.NewAttrib("ontimeupdate", execute) }

// OnToggle `toggle` event handler (all HTML elements)
func OnToggle(execute string) mx.Attrib { return mx.NewAttrib("ontoggle", execute) }

// OnUnhandledRejection `unhandledrejection` event handler for Window object (body element)
func OnUnhandledRejection(execute string) mx.Attrib {
	return mx.NewAttrib("onunhandledrejection", execute)
}

// OnUnload `unload` event handler for Window object (body element)
func OnUnload(execute string) mx.Attrib { return mx.NewAttrib("onunload", execute) }

// OnVolumeChange `volumechange` event handler (all HTML elements)
func OnVolumeChange(execute string) mx.Attrib { return mx.NewAttrib("onvolumechange", execute) }

// OnWaiting `waiting` event handler (all HTML elements)
func OnWaiting(execute string) mx.Attrib { return mx.NewAttrib("onwaiting", execute) }

// OnWheel `wheel` event handler (all HTML elements)
func OnWheel(execute string) mx.Attrib { return mx.NewAttrib("onwheel", execute) }
