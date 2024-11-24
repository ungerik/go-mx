package html

import (
	"strconv"
	"strings"

	"github.com/ungerik/go-mx"
)

type Attribs = mx.Attribs

func Attrib(name, value string) mx.Attribute {
	return mx.Attribute{Name: name, Value: value}
}

type BoolAttrib string

func (a BoolAttrib) Attribute() (name, value string) {
	return string(a), string(a)
}

// See https://github.com/jozo/all-html-elements-and-attributes
// and https://html.spec.whatwg.org/multipage/indices.html#attributes-3

func Accept(contentTypes ...string) mx.Attrib {
	return mx.NewAttrib("accept", strings.Join(contentTypes, ","))
}
func AcceptCharset(charsets ...string) mx.Attrib {
	return mx.NewAttrib("accept-charset", strings.Join(charsets, " "))
}
func AccessKey(value string) mx.Attrib { return mx.NewAttrib("accesskey", value) }
func AccessKeyf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("accesskey", valueFmt, a...)
}
func Action(url string) mx.Attrib               { return mx.NewAttrib("action", url) }
func Actionf(urlFmt string, a ...any) mx.Attrib { return mx.NewAttribf("action", urlFmt, a...) }
func Align(value string) mx.Attrib              { return mx.NewAttrib("align", value) }
func Alignf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("align", valueFmt, a...)
}
func Allow(value string) mx.Attrib { return mx.NewAttrib("allow", value) }
func Allowf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("allow", valueFmt, a...)
}

const Alpha = BoolAttrib("alpha")

func Alt(text string) mx.Attrib               { return mx.NewAttrib("alt", text) }
func Altf(textFmt string, a ...any) mx.Attrib { return mx.NewAttribf("alt", textFmt, a...) }
func As(value string) mx.Attrib               { return mx.NewAttrib("as", value) }
func Asf(valueFmt string, a ...any) mx.Attrib { return mx.NewAttribf("as", valueFmt, a...) }

const Async = BoolAttrib("async")

var AutoCapitalizeNone = mx.NewAttrib("autocapitalize", "none")
var AutoCapitalizeSentences = mx.NewAttrib("autocapitalize", "sentences")
var AutoCapitalizeWords = mx.NewAttrib("autocapitalize", "words")
var AutoCapitalizeCharacters = mx.NewAttrib("autocapitalize", "characters")

func AutoComplete(tokens ...string) mx.Attrib {
	if len(tokens) == 0 {
		return AutoCompleteOn
	}
	return mx.NewAttrib("autocomplete", strings.Join(tokens, " "))
}

var AutoCompleteOn = mx.NewAttrib("autocomplete", "on")
var AutoCompleteOff = mx.NewAttrib("autocomplete", "off")
var AutoCorrectOn = mx.NewAttrib("autocorrect", "on")
var AutoCorrectOff = mx.NewAttrib("autocorrect", "off")
var AutoFocus = mx.NewAttrib("autofocus", "autofocus")
var AutoPlay = mx.NewAttrib("autoplay", "autoplay")

func Background(style string) mx.Attrib { return mx.NewAttrib("background", style) }
func BGColor(color string) mx.Attrib    { return mx.NewAttrib("bgcolor", color) }
func Border(value string) mx.Attrib     { return mx.NewAttrib("border", value) }
func Borderf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("border", valueFmt, a...)
}
func Capture(value string) mx.Attrib { return mx.NewAttrib("capture", value) }
func Capturef(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("capture", valueFmt, a...)
}
func CharSet(value string) mx.Attrib { return mx.NewAttrib("charset", value) }

const Checked = BoolAttrib("checked")

func CiteAttr(value string) mx.Attrib { return mx.NewAttrib("cite", value) }
func CiteAttrf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("cite", valueFmt, a...)
}
func Class(classes ...string) mx.Attrib          { return mx.NewAttrib("class", strings.Join(classes, " ")) }
func Classf(classFmt string, a ...any) mx.Attrib { return mx.NewAttribf("class", classFmt, a...) }
func Color(value string) mx.Attrib               { return mx.NewAttrib("color", value) }
func Colorf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("color", valueFmt, a...)
}
func Cols(numChars int) mx.Attrib       { return mx.NewAttrib("cols", strconv.Itoa(numChars)) }
func ColSpan(numCols int) mx.Attrib     { return mx.NewAttrib("colspan", strconv.Itoa(numCols)) }
func ContentAttr(text string) mx.Attrib { return mx.NewAttrib("content", text) }

var ContentEditableTrue = mx.NewAttrib("contenteditable", "true")
var ContentEditableFalse = mx.NewAttrib("contenteditable", "false")
var ContentEditablePlaintextOnly = mx.NewAttrib("contenteditable", "plaintext-only")
var Controls = mx.NewAttrib("controls", "controls")

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

var CrossOriginAnonymous = mx.NewAttrib("crossorigin", "anonymous")
var CrossOriginUseCredentials = mx.NewAttrib("crossorigin", "use-credentials")

func CSP(value string) mx.Attrib { return mx.NewAttrib("csp", value) }
func CSPf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("csp", valueFmt, a...)
}
func DataAttr(name, value string) mx.Attrib { return mx.NewAttrib("data-"+name, value) }
func DataAttrf(name, valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("data-"+name, valueFmt, a...)
}
func Datetime(value string) mx.Attrib { return mx.NewAttrib("datetime", value) }
func Datetimef(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("datetime", valueFmt, a...)
}

var DecodingAuto = mx.NewAttrib("decoding", "auto")
var DecodingAsync = mx.NewAttrib("decoding", "async")
var DecodingSync = mx.NewAttrib("decoding", "sync")

const Default = BoolAttrib("default")
const Defer = BoolAttrib("defer")

var DirLTR = mx.NewAttrib("dir", "ltr")
var DirRTL = mx.NewAttrib("dir", "rtl")
var DirAuto = mx.NewAttrib("dir", "auto")

func DirName(name string) mx.Attrib { return mx.NewAttrib("dirname", name) }

const Disabled = BoolAttrib("disabled")

func Download(filename string) mx.Attrib { return mx.NewAttrib("download", filename) }
func Draggable(value bool) mx.Attrib     { return mx.NewAttrib("draggable", strconv.FormatBool(value)) }

var EncTypeFormURLEndoced = mx.NewAttrib("enctype", "application/x-www-form-urlencoded")
var EncTypeMultipartFormData = mx.NewAttrib("enctype", "multipart/form-data")
var EncTypeTextPlain = mx.NewAttrib("enctype", "text/plain")
var EnterKeyHintEnter = mx.NewAttrib("enterkeyhint", "enter")
var EnterKeyHintDone = mx.NewAttrib("enterkeyhint", "done")
var EnterKeyHintGo = mx.NewAttrib("enterkeyhint", "go")
var EnterKeyHintNext = mx.NewAttrib("enterkeyhint", "next")
var EnterKeyHintPrevious = mx.NewAttrib("enterkeyhint", "previous")
var EnterKeyHintSearch = mx.NewAttrib("enterkeyhint", "search")
var EnterKeyHintSend = mx.NewAttrib("enterkeyhint", "send")
var FetchPriorityAuto = mx.NewAttrib("fetchpriority", "auto")
var FetchPriorityHigh = mx.NewAttrib("fetchpriority", "high")
var FetchPriorityLow = mx.NewAttrib("fetchpriority", "low")

func For(id string) mx.Attrib          { return mx.NewAttrib("for", id) }
func FormAttr(formID string) mx.Attrib { return mx.NewAttrib("form", formID) }
func FormAction(url string) mx.Attrib  { return mx.NewAttrib("formaction", url) }

var FormEncTypeFormURLEndoced = mx.NewAttrib("formenctype", "application/x-www-form-urlencoded")
var FormEncTypeMultipartFormData = mx.NewAttrib("formenctype", "multipart/form-data")
var FormEncTypeTextPlain = mx.NewAttrib("formenctype", "text/plain")
var FormMethodGET = mx.NewAttrib("formmethod", "get")
var FormMethodPOST = mx.NewAttrib("formmethod", "post")
var FormMethodDialog = mx.NewAttrib("formmethod", "dialog")

const FormNoValidate = BoolAttrib("formnovalidate")

func FormTarget(value string) mx.Attrib { return mx.NewAttrib("formtarget", value) }
func FormTargetf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("formtarget", valueFmt, a...)
}
func Headers(headerCellIDs ...string) mx.Attrib {
	return mx.NewAttrib("headers", strings.Join(headerCellIDs, " "))
}
func Height(value string) mx.Attrib { return mx.NewAttrib("height", value) }
func Heightf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("height", valueFmt, a...)
}

const Hidden = BoolAttrib("hidden")

var HiddenUntilFound = mx.NewAttrib("hidden", "until-found")

func High(limit float64) mx.Attrib {
	return mx.NewAttrib("high", strconv.FormatFloat(limit, 'f', -1, 64))
}
func Hight(pixels int) mx.Attrib      { return mx.NewAttrib("high", strconv.Itoa(pixels)) }
func HRef(url string) mx.Attrib       { return mx.NewAttrib("href", url) }
func HRefLang(value string) mx.Attrib { return mx.NewAttrib("hreflang", value) }
func HRefLangf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("hreflang", valueFmt, a...)
}

var HTTPEquivContentType = mx.NewAttrib("http-equiv", "content-type")
var HTTPEquivDefaultStyle = mx.NewAttrib("http-equiv", "default-style")
var HTTPEquivRefresh = mx.NewAttrib("http-equiv", "refresh")
var HTTPEquivXUACompatible = mx.NewAttrib("http-equiv", "x-ua-compatible")
var HTTPEquivContentSecurityPolicy = mx.NewAttrib("http-equiv", "content-security-policy")

func ID(value string) mx.Attrib               { return mx.NewAttrib("id", value) }
func IDf(valueFmt string, a ...any) mx.Attrib { return mx.NewAttribf("id", valueFmt, a...) }

// imagesizes ?
// imagesrcset ?

const Inert = BoolAttrib("inert")

var InputModeNone = mx.NewAttrib("inputmode", "none")
var InputModeText = mx.NewAttrib("inputmode", "text")
var InputModeTel = mx.NewAttrib("inputmode", "tel")
var InputModeEmail = mx.NewAttrib("inputmode", "email")
var InputModeURL = mx.NewAttrib("inputmode", "url")
var InputModeNumeric = mx.NewAttrib("inputmode", "numeric")
var InputModeDecimal = mx.NewAttrib("inputmode", "decimal")
var InputModeSearch = mx.NewAttrib("inputmode", "search")

func Integrity(value string) mx.Attrib { return mx.NewAttrib("integrity", value) }
func Integrityf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("integrity", valueFmt, a...)
}
func IntrinsicSize(value string) mx.Attrib { return mx.NewAttrib("intrinsicsize", value) }
func IntrinsicSizef(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("intrinsicsize", valueFmt, a...)
}

const IsMap = BoolAttrib("ismap")

func ItemID(url string) mx.Attrib        { return mx.NewAttrib("itemid", url) }
func ItemProp(props ...string) mx.Attrib { return mx.NewAttrib("itemprop", strings.Join(props, " ")) }
func ItemRef(ids ...string) mx.Attrib    { return mx.NewAttrib("itemref", strings.Join(ids, " ")) }

const ItemScope = BoolAttrib("itemscope")

func ItemType(urls ...string) mx.Attrib { return mx.NewAttrib("itemtype", strings.Join(urls, " ")) }

var KindSubtitles = mx.NewAttrib("kind", "subtitles")
var KindCaptions = mx.NewAttrib("kind", "captions")
var KindDescriptions = mx.NewAttrib("kind", "descriptions")
var KindChapters = mx.NewAttrib("kind", "chapters")
var KindMetadata = mx.NewAttrib("kind", "metadata")

func LabelAttr(value string) mx.Attrib { return mx.NewAttrib("label", value) }
func LabelAttrf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("label", valueFmt, a...)
}
func Lang(value string) mx.Attrib { return mx.NewAttrib("lang", value) }
func Langf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("lang", valueFmt, a...)
}
func Language(value string) mx.Attrib { return mx.NewAttrib("language", value) }
func Languagef(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("language", valueFmt, a...)
}
func List(id string) mx.Attrib { return mx.NewAttrib("list", id) }

var LoadingEager = mx.NewAttrib("loading", "eager")
var LoadingLazy = mx.NewAttrib("loading", "lazy")
var Loop = mx.NewAttrib("loop", "loop")

func Low(limit float64) mx.Attrib {
	return mx.NewAttrib("low", strconv.FormatFloat(limit, 'f', -1, 64))
}
func Max(value string) mx.Attrib { return mx.NewAttrib("max", value) }
func Maxf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("max", valueFmt, a...)
}
func MaxLength(length int) mx.Attrib { return mx.NewAttrib("maxlength", strconv.Itoa(length)) }

func Media(query string) mx.Attrib { return mx.NewAttrib("media", query) }

var MethodGET = mx.NewAttrib("method", "GET")
var MethodPOST = mx.NewAttrib("method", "POST")
var MethodDialog = mx.NewAttrib("method", "dialog")

func Min(value string) mx.Attrib { return mx.NewAttrib("min", value) }
func Minf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("min", valueFmt, a...)
}
func MinLength(length int) mx.Attrib { return mx.NewAttrib("minlength", strconv.Itoa(length)) }

const Multiple = BoolAttrib("multiple")
const Muted = BoolAttrib("muted")

func Name(value string) mx.Attrib { return mx.NewAttrib("name", value) }
func Namef(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("name", valueFmt, a...)
}

const NoModule = BoolAttrib("nomodule")

func Nonce(value string) mx.Attrib { return mx.NewAttrib("nonce", value) }
func Noncef(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("nonce", valueFmt, a...)
}

const NoValidate = BoolAttrib("novalidate")
const Open = BoolAttrib("open")

func Optimum(value float64) mx.Attrib {
	return mx.NewAttrib("optimum", strconv.FormatFloat(value, 'f', -1, 64))
}
func Pattern(value string) mx.Attrib { return mx.NewAttrib("pattern", value) }
func Patternf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("pattern", valueFmt, a...)
}
func Ping(value string) mx.Attrib { return mx.NewAttrib("ping", value) }
func Pingf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("ping", valueFmt, a...)
}
func Placeholder(value string) mx.Attrib { return mx.NewAttrib("placeholder", value) }
func Placeholderf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("placeholder", valueFmt, a...)
}
func PlaysInline(value string) mx.Attrib { return mx.NewAttrib("playsinline", value) }
func PlaysInlinef(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("playsinline", valueFmt, a...)
}
func Poster(value string) mx.Attrib { return mx.NewAttrib("poster", value) }
func Posterf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("poster", valueFmt, a...)
}
func Preload(value string) mx.Attrib { return mx.NewAttrib("preload", value) }
func Preloadf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("preload", valueFmt, a...)
}
func Readonly(value string) mx.Attrib { return mx.NewAttrib("readonly", value) }
func Readonlyf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("readonly", valueFmt, a...)
}

var ReferrerPolicyNoReferrer = mx.NewAttrib("referrerpolicy", "no-referrer")
var ReferrerPolicyNoReferrerWhenDowngrade = mx.NewAttrib("referrerpolicy", "no-referrer-when-downgrade")
var ReferrerPolicyOrigin = mx.NewAttrib("referrerpolicy", "origin")
var ReferrerPolicyOriginWhenCrossOrigin = mx.NewAttrib("referrerpolicy", "origin-when-cross-origin")
var ReferrerPolicySameOrigin = mx.NewAttrib("referrerpolicy", "same-origin")
var ReferrerPolicyStrictOrigin = mx.NewAttrib("referrerpolicy", "strict-origin")
var ReferrerPolicyStrictOriginWhenCrossOrigin = mx.NewAttrib("referrerpolicy", "strict-origin-when-cross-origin")
var ReferrerPolicyUnsafeUrl = mx.NewAttrib("referrerpolicy", "unsafe-url")

func Rel(keywords ...string) mx.Attrib { return mx.NewAttrib("rel", strings.Join(keywords, " ")) }

const Required = BoolAttrib("required")

func Reversed(value string) mx.Attrib { return mx.NewAttrib("reversed", value) }
func Reversedf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("reversed", valueFmt, a...)
}
func Role(value string) mx.Attrib { return mx.NewAttrib("role", value) }
func Rolef(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("role", valueFmt, a...)
}
func Rows(value string) mx.Attrib { return mx.NewAttrib("rows", value) }
func Rowsf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("rows", valueFmt, a...)
}
func RowSpan(value string) mx.Attrib { return mx.NewAttrib("rowspan", value) }
func RowSpanf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("rowspan", valueFmt, a...)
}
func Sandbox(value string) mx.Attrib { return mx.NewAttrib("sandbox", value) }
func Sandboxf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("sandbox", valueFmt, a...)
}
func Scope(value string) mx.Attrib { return mx.NewAttrib("scope", value) }
func Scopef(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("scope", valueFmt, a...)
}
func Scoped(value string) mx.Attrib { return mx.NewAttrib("scoped", value) }
func Scopedf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("scoped", valueFmt, a...)
}
func Selected(value string) mx.Attrib { return mx.NewAttrib("selected", value) }
func Selectedf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("selected", valueFmt, a...)
}

var ShapeDefault = mx.NewAttrib("shape", "default")
var ShapeRect = mx.NewAttrib("shape", "rect")
var ShapeCircle = mx.NewAttrib("shape", "circle")
var ShapePoly = mx.NewAttrib("shape", "poly")

func Size(value string) mx.Attrib { return mx.NewAttrib("size", value) }
func Sizef(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("size", valueFmt, a...)
}
func Sizes(sourceSizes ...string) mx.Attrib {
	return mx.NewAttrib("sizes", strings.Join(sourceSizes, ","))
}
func SlotAttr(value string) mx.Attrib { return mx.NewAttrib("slot", value) }
func SlotAttrf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("slot", valueFmt, a...)
}
func SpanAttr(value string) mx.Attrib { return mx.NewAttrib("span", value) }
func SpanAttrf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("span", valueFmt, a...)
}
func SpellCheck(value string) mx.Attrib { return mx.NewAttrib("spellcheck", value) }
func SpellCheckf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("spellcheck", valueFmt, a...)
}
func Src(url string) mx.Attrib      { return mx.NewAttrib("src", url) }
func SrcDoc(value string) mx.Attrib { return mx.NewAttrib("srcdoc", value) }
func SrcDocf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("srcdoc", valueFmt, a...)
}
func SrcLang(value string) mx.Attrib { return mx.NewAttrib("srclang", value) }
func SrcLangf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("srclang", valueFmt, a...)
}
func SrcSet(sources ...string) mx.Attrib {
	return mx.NewAttrib("srcset", strings.Join(sources, ","))
}
func Start(value string) mx.Attrib { return mx.NewAttrib("start", value) }
func Startf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("start", valueFmt, a...)
}
func Step(value string) mx.Attrib { return mx.NewAttrib("step", value) }
func Stepf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("step", valueFmt, a...)
}
func Style(value string) mx.Attrib { return mx.NewAttrib("style", value) }
func Stylef(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("style", valueFmt, a...)
}
func TabIndex(value string) mx.Attrib { return mx.NewAttrib("tabindex", value) }
func TabIndexf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("tabindex", valueFmt, a...)
}
func Target(value string) mx.Attrib { return mx.NewAttrib("target", value) }
func Targetf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("target", valueFmt, a...)
}

var TargetSelf = mx.NewAttrib("target", "_self")
var TargetBlank = mx.NewAttrib("target", "_blank")
var TargetParent = mx.NewAttrib("target", "_parent")
var TargetTop = mx.NewAttrib("target", "_top")
var TargetUnfencedTop = mx.NewAttrib("target", "_unfencedTop")

func Title(value string) mx.Attrib { return mx.NewAttrib("title", value) }
func Titlef(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("title", valueFmt, a...)
}
func Translate(value string) mx.Attrib { return mx.NewAttrib("translate", value) }
func Translatef(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("translate", valueFmt, a...)
}
func Type(value string) mx.Attrib { return mx.NewAttrib("type", value) }
func Typef(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("type", valueFmt, a...)
}
func UseMap(partialURL string) mx.Attrib { return mx.NewAttrib("usemap", partialURL) }
func Value(value string) mx.Attrib       { return mx.NewAttrib("value", value) }
func Valuef(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("value", valueFmt, a...)
}
func Width(pixels int) mx.Attrib  { return mx.NewAttrib("width", strconv.Itoa(pixels)) }
func Wrap(value string) mx.Attrib { return mx.NewAttrib("wrap", value) }
func Wrapf(valueFmt string, a ...any) mx.Attrib {
	return mx.NewAttribf("wrap", valueFmt, a...)
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
