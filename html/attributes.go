package html

import (
	"strconv"
	"strings"

	"github.com/ungerik/go-mx"
)

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
func Action(url string) mx.Attrib      { return mx.NewAttrib("action", url) }
func Align(value string) mx.Attrib     { return mx.NewAttrib("align", value) }
func Allow(value string) mx.Attrib     { return mx.NewAttrib("allow", value) }

const Alpha = BoolAttrib("alpha")

func Alt(text string) mx.Attrib { return mx.NewAttrib("alt", text) }
func As(value string) mx.Attrib { return mx.NewAttrib("as", value) }

const Async = BoolAttrib("async")

func AutoCapitalizeNone() mx.Attrib       { return mx.NewAttrib("autocapitalize", "none") }
func AutoCapitalizeSentences() mx.Attrib  { return mx.NewAttrib("autocapitalize", "sentences") }
func AutoCapitalizeWords() mx.Attrib      { return mx.NewAttrib("autocapitalize", "words") }
func AutoCapitalizeCharacters() mx.Attrib { return mx.NewAttrib("autocapitalize", "characters") }
func AutoComplete(tokens ...string) mx.Attrib {
	if len(tokens) == 0 {
		return AutoCompleteOn()
	}
	return mx.NewAttrib("autocomplete", strings.Join(tokens, " "))
}
func AutoCompleteOn() mx.Attrib         { return mx.NewAttrib("autocomplete", "on") }
func AutoCompleteOff() mx.Attrib        { return mx.NewAttrib("autocomplete", "off") }
func AutoCorrectOn() mx.Attrib          { return mx.NewAttrib("autocorrect", "on") }
func AutoCorrectOff() mx.Attrib         { return mx.NewAttrib("autocorrect", "off") }
func AutoFocus() mx.Attrib              { return mx.NewAttrib("autofocus", "autofocus") }
func AutoPlay() mx.Attrib               { return mx.NewAttrib("autoplay", "autoplay") }
func Background(style string) mx.Attrib { return mx.NewAttrib("background", style) }
func BGColor(color string) mx.Attrib    { return mx.NewAttrib("bgcolor", color) }
func Border(value string) mx.Attrib     { return mx.NewAttrib("border", value) }
func Capture(value string) mx.Attrib    { return mx.NewAttrib("capture", value) }
func CharSet(value string) mx.Attrib    { return mx.NewAttrib("charset", value) }

const Checked = BoolAttrib("checked")

func CiteAttr(value string) mx.Attrib   { return mx.NewAttrib("cite", value) }
func Class(classes ...string) mx.Attrib { return mx.NewAttrib("class", strings.Join(classes, " ")) }
func Color(value string) mx.Attrib      { return mx.NewAttrib("color", value) }
func Cols(numChars int) mx.Attrib       { return mx.NewAttrib("cols", strconv.Itoa(numChars)) }
func ColSpan(numCols int) mx.Attrib     { return mx.NewAttrib("colspan", strconv.Itoa(numCols)) }
func ContentAttr(text string) mx.Attrib { return mx.NewAttrib("content", text) }
func ContentEditableTrue() mx.Attrib    { return mx.NewAttrib("contenteditable", "true") }
func ContentEditableFalse() mx.Attrib   { return mx.NewAttrib("contenteditable", "false") }
func ContentEditablePlaintextOnly(value string) mx.Attrib {
	return mx.NewAttrib("contenteditable", "plaintext-only")
}
func Controls() mx.Attrib { return mx.NewAttrib("controls", "controls") }
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
func CrossOriginAnonymous() mx.Attrib       { return mx.NewAttrib("crossorigin", "anonymous") }
func CrossOriginUseCredentials() mx.Attrib  { return mx.NewAttrib("crossorigin", "use-credentials") }
func CSP(value string) mx.Attrib            { return mx.NewAttrib("csp", value) }
func DataAttr(name, value string) mx.Attrib { return mx.NewAttrib("data-"+name, value) }
func Datetime(value string) mx.Attrib       { return mx.NewAttrib("datetime", value) }
func DecodingAuto() mx.Attrib               { return mx.NewAttrib("decoding", "auto") }
func DecodingAsync() mx.Attrib              { return mx.NewAttrib("decoding", "async") }
func DecodingSync() mx.Attrib               { return mx.NewAttrib("decoding", "sync") }
func Default() mx.Attrib                    { return mx.NewAttrib("default", "default") }
func Defer() mx.Attrib                      { return mx.NewAttrib("defer", "defer") }
func DirLTR() mx.Attrib                     { return mx.NewAttrib("dir", "ltr") }
func DirRTL() mx.Attrib                     { return mx.NewAttrib("dir", "rtl") }
func DirAuto() mx.Attrib                    { return mx.NewAttrib("dir", "auto") }
func DirName(name string) mx.Attrib         { return mx.NewAttrib("dirname", name) }

const Disabled = BoolAttrib("disabled")

func Download(filename string) mx.Attrib { return mx.NewAttrib("download", filename) }
func Draggable(value bool) mx.Attrib     { return mx.NewAttrib("draggable", strconv.FormatBool(value)) }
func EncTypeFormURLEndoced(value string) mx.Attrib {
	return mx.NewAttrib("enctype", "application/x-www-form-urlencoded")
}
func EncTypeMultipartFormData(value string) mx.Attrib {
	return mx.NewAttrib("enctype", "multipart/form-data")
}
func EncTypeTextPlain(value string) mx.Attrib { return mx.NewAttrib("enctype", "text/plain") }
func EnterKeyHintEnter() mx.Attrib            { return mx.NewAttrib("enterkeyhint", "enter") }
func EnterKeyHintDone() mx.Attrib             { return mx.NewAttrib("enterkeyhint", "done") }
func EnterKeyHintGo() mx.Attrib               { return mx.NewAttrib("enterkeyhint", "go") }
func EnterKeyHintNext() mx.Attrib             { return mx.NewAttrib("enterkeyhint", "next") }
func EnterKeyHintPrevious() mx.Attrib         { return mx.NewAttrib("enterkeyhint", "previous") }
func EnterKeyHintSearch() mx.Attrib           { return mx.NewAttrib("enterkeyhint", "search") }
func EnterKeyHintSend() mx.Attrib             { return mx.NewAttrib("enterkeyhint", "send") }
func FetchPriorityAuto() mx.Attrib            { return mx.NewAttrib("fetchpriority", "auto") }
func FetchPriorityHigh() mx.Attrib            { return mx.NewAttrib("fetchpriority", "high") }
func FetchPriorityLow() mx.Attrib             { return mx.NewAttrib("fetchpriority", "low") }
func For(id string) mx.Attrib                 { return mx.NewAttrib("for", id) }
func FormAttr(formID string) mx.Attrib        { return mx.NewAttrib("form", formID) }
func FormAction(url string) mx.Attrib         { return mx.NewAttrib("formaction", url) }
func FormEncTypeFormURLEndoced(value string) mx.Attrib {
	return mx.NewAttrib("formenctype", "application/x-www-form-urlencoded")
}
func FormEncTypeMultipartFormData(value string) mx.Attrib {
	return mx.NewAttrib("formenctype", "multipart/form-data")
}
func FormEncTypeTextPlain(value string) mx.Attrib { return mx.NewAttrib("formenctype", "text/plain") }
func FormMethodGET() mx.Attrib                    { return mx.NewAttrib("formmethod", "get") }
func FormMethodPOST() mx.Attrib                   { return mx.NewAttrib("formmethod", "post") }
func FormMethodDialog() mx.Attrib                 { return mx.NewAttrib("formmethod", "dialog") }

const FormNoValidate = BoolAttrib("formnovalidate")

func FormTarget(value string) mx.Attrib { return mx.NewAttrib("formtarget", value) }
func Headers(headerCellIDs ...string) mx.Attrib {
	return mx.NewAttrib("headers", strings.Join(headerCellIDs, " "))
}
func Height(value string) mx.Attrib { return mx.NewAttrib("height", value) }

const Hidden = BoolAttrib("hidden")

func HiddenUntilFound() mx.Attrib { return mx.NewAttrib("hidden", "until-found") }
func High(limit float64) mx.Attrib {
	return mx.NewAttrib("high", strconv.FormatFloat(limit, 'f', -1, 64))
}
func Hight(pixels int) mx.Attrib        { return mx.NewAttrib("high", strconv.Itoa(pixels)) }
func HRef(url string) mx.Attrib         { return mx.NewAttrib("href", url) }
func HRefLang(value string) mx.Attrib   { return mx.NewAttrib("hreflang", value) }
func HTTPEquivContentType() mx.Attrib   { return mx.NewAttrib("http-equiv", "content-type") }
func HTTPEquivDefaultStyle() mx.Attrib  { return mx.NewAttrib("http-equiv", "default-style") }
func HTTPEquivRefresh() mx.Attrib       { return mx.NewAttrib("http-equiv", "refresh") }
func HTTPEquivXUACompatible() mx.Attrib { return mx.NewAttrib("http-equiv", "x-ua-compatible") }
func HTTPEquivContentSecurityPolicy() mx.Attrib {
	return mx.NewAttrib("http-equiv", "content-security-policy")
}
func ID(value string) mx.Attrib { return mx.NewAttrib("id", value) }

// imagesizes ?
// imagesrcset ?

const Inert = BoolAttrib("inert")

func InputModeNone() mx.Attrib             { return mx.NewAttrib("inputmode", "none") }
func InputModeText() mx.Attrib             { return mx.NewAttrib("inputmode", "text") }
func InputModeTel() mx.Attrib              { return mx.NewAttrib("inputmode", "tel") }
func InputModeEmail() mx.Attrib            { return mx.NewAttrib("inputmode", "email") }
func InputModeURL() mx.Attrib              { return mx.NewAttrib("inputmode", "url") }
func InputModeNumeric() mx.Attrib          { return mx.NewAttrib("inputmode", "numeric") }
func InputModeDecimal() mx.Attrib          { return mx.NewAttrib("inputmode", "decimal") }
func InputModeSearch() mx.Attrib           { return mx.NewAttrib("inputmode", "search") }
func Integrity(value string) mx.Attrib     { return mx.NewAttrib("integrity", value) }
func IntrinsicSize(value string) mx.Attrib { return mx.NewAttrib("intrinsicsize", value) }

const IsMap = BoolAttrib("ismap")

func ItemID(url string) mx.Attrib        { return mx.NewAttrib("itemid", url) }
func ItemProp(props ...string) mx.Attrib { return mx.NewAttrib("itemprop", strings.Join(props, " ")) }
func ItemRef(ids ...string) mx.Attrib    { return mx.NewAttrib("itemref", strings.Join(ids, " ")) }

const ItemScope = BoolAttrib("itemscope")

func ItemType(urls ...string) mx.Attrib   { return mx.NewAttrib("itemtype", strings.Join(urls, " ")) }
func KindSubtitles() mx.Attrib            { return mx.NewAttrib("kind", "subtitles") }
func KindCaptions() mx.Attrib             { return mx.NewAttrib("kind", "captions") }
func KindDescriptions() mx.Attrib         { return mx.NewAttrib("kind", "descriptions") }
func KindChapters() mx.Attrib             { return mx.NewAttrib("kind", "chapters") }
func KindMetadata() mx.Attrib             { return mx.NewAttrib("kind", "metadata") }
func LabelAttr(value string) mx.Attrib    { return mx.NewAttrib("label", value) }
func Lang(value string) mx.Attrib         { return mx.NewAttrib("lang", value) }
func Language(value string) mx.Attrib     { return mx.NewAttrib("language", value) }
func List(id string) mx.Attrib            { return mx.NewAttrib("list", id) }
func LoadingEager(value string) mx.Attrib { return mx.NewAttrib("loading", "eager") }
func LoadingLazy(value string) mx.Attrib  { return mx.NewAttrib("loading", "lazy") }
func Loop() mx.Attrib                     { return mx.NewAttrib("loop", "loop") }
func Low(limit float64) mx.Attrib {
	return mx.NewAttrib("low", strconv.FormatFloat(limit, 'f', -1, 64))
}
func Max(value string) mx.Attrib     { return mx.NewAttrib("max", value) }
func MaxLength(length int) mx.Attrib { return mx.NewAttrib("maxlength", strconv.Itoa(length)) }

func Media(query string) mx.Attrib   { return mx.NewAttrib("media", query) }
func MethodGET() mx.Attrib           { return mx.NewAttrib("method", "GET") }
func MethodPOST() mx.Attrib          { return mx.NewAttrib("method", "POST") }
func MethodDialog() mx.Attrib        { return mx.NewAttrib("method", "dialog") }
func Min(value string) mx.Attrib     { return mx.NewAttrib("min", value) }
func MinLength(length int) mx.Attrib { return mx.NewAttrib("minlength", strconv.Itoa(length)) }

const Multiple = BoolAttrib("multiple")
const Muted = BoolAttrib("muted")

func Name(value string) mx.Attrib { return mx.NewAttrib("name", value) }

const NoModule = BoolAttrib("nomodule")

func Nonce(value string) mx.Attrib { return mx.NewAttrib("nonce", value) }

const NoValidate = BoolAttrib("novalidate")
const Open = BoolAttrib("open")

func Optimum(value float64) mx.Attrib {
	return mx.NewAttrib("optimum", strconv.FormatFloat(value, 'f', -1, 64))
}
func Pattern(value string) mx.Attrib      { return mx.NewAttrib("pattern", value) }
func Ping(value string) mx.Attrib         { return mx.NewAttrib("ping", value) }
func Placeholder(value string) mx.Attrib  { return mx.NewAttrib("placeholder", value) }
func PlaysInline(value string) mx.Attrib  { return mx.NewAttrib("playsinline", value) }
func Poster(value string) mx.Attrib       { return mx.NewAttrib("poster", value) }
func Preload(value string) mx.Attrib      { return mx.NewAttrib("preload", value) }
func Readonly(value string) mx.Attrib     { return mx.NewAttrib("readonly", value) }
func ReferrerPolicyNoReferrer() mx.Attrib { return mx.NewAttrib("referrerpolicy", "no-referrer") }
func ReferrerPolicyNoReferrerWhenDowngrade() mx.Attrib {
	return mx.NewAttrib("referrerpolicy", "no-referrer-when-downgrade")
}
func ReferrerPolicyOrigin() mx.Attrib { return mx.NewAttrib("referrerpolicy", "origin") }
func ReferrerPolicyOriginWhenCrossOrigin() mx.Attrib {
	return mx.NewAttrib("referrerpolicy", "origin-when-cross-origin")
}
func ReferrerPolicySameOrigin() mx.Attrib { return mx.NewAttrib("referrerpolicy", "same-origin") }
func ReferrerPolicyStrictOrigin() mx.Attrib {
	return mx.NewAttrib("referrerpolicy", "strict-origin")
}
func ReferrerPolicyStrictOriginWhenCrossOrigin() mx.Attrib {
	return mx.NewAttrib("referrerpolicy", "strict-origin-when-cross-origin")
}
func ReferrerPolicyUnsafeUrl() mx.Attrib { return mx.NewAttrib("referrerpolicy", "unsafe-url") }
func Rel(keywords ...string) mx.Attrib   { return mx.NewAttrib("rel", strings.Join(keywords, " ")) }

const Required = BoolAttrib("required")

func Reversed(value string) mx.Attrib { return mx.NewAttrib("reversed", value) }
func Role(value string) mx.Attrib     { return mx.NewAttrib("role", value) }
func Rows(value string) mx.Attrib     { return mx.NewAttrib("rows", value) }
func RowSpan(value string) mx.Attrib  { return mx.NewAttrib("rowspan", value) }
func Sandbox(value string) mx.Attrib  { return mx.NewAttrib("sandbox", value) }
func Scope(value string) mx.Attrib    { return mx.NewAttrib("scope", value) }
func Scoped(value string) mx.Attrib   { return mx.NewAttrib("scoped", value) }
func Selected(value string) mx.Attrib { return mx.NewAttrib("selected", value) }
func ShapeDefault() mx.Attrib         { return mx.NewAttrib("shape", "default") }
func ShapeRect() mx.Attrib            { return mx.NewAttrib("shape", "rect") }
func ShapeCircle() mx.Attrib          { return mx.NewAttrib("shape", "circle") }
func ShapePoly() mx.Attrib            { return mx.NewAttrib("shape", "poly") }
func Size(value string) mx.Attrib     { return mx.NewAttrib("size", value) }
func Sizes(sourceSizes ...string) mx.Attrib {
	return mx.NewAttrib("sizes", strings.Join(sourceSizes, ","))
}
func SlotAttr(value string) mx.Attrib   { return mx.NewAttrib("slot", value) }
func SpanAttr(value string) mx.Attrib   { return mx.NewAttrib("span", value) }
func SpellCheck(value string) mx.Attrib { return mx.NewAttrib("spellcheck", value) }
func Src(url string) mx.Attrib          { return mx.NewAttrib("src", url) }
func SrcDoc(value string) mx.Attrib     { return mx.NewAttrib("srcdoc", value) }
func SrcLang(value string) mx.Attrib    { return mx.NewAttrib("srclang", value) }
func SrcSet(sources ...string) mx.Attrib {
	return mx.NewAttrib("srcset", strings.Join(sources, ","))
}
func Start(value string) mx.Attrib       { return mx.NewAttrib("start", value) }
func Step(value string) mx.Attrib        { return mx.NewAttrib("step", value) }
func Style(value string) mx.Attrib       { return mx.NewAttrib("style", value) }
func TabIndex(value string) mx.Attrib    { return mx.NewAttrib("tabindex", value) }
func Target(value string) mx.Attrib      { return mx.NewAttrib("target", value) }
func TargetSelf() mx.Attrib              { return mx.NewAttrib("target", "_self") }
func TargetBlank() mx.Attrib             { return mx.NewAttrib("target", "_blank") }
func TargetParent() mx.Attrib            { return mx.NewAttrib("target", "_parent") }
func TargetTop() mx.Attrib               { return mx.NewAttrib("target", "_top") }
func TargetUnfencedTop() mx.Attrib       { return mx.NewAttrib("target", "_unfencedTop") }
func Title(value string) mx.Attrib       { return mx.NewAttrib("title", value) }
func Translate(value string) mx.Attrib   { return mx.NewAttrib("translate", value) }
func Type(value string) mx.Attrib        { return mx.NewAttrib("type", value) }
func UseMap(partialURL string) mx.Attrib { return mx.NewAttrib("usemap", partialURL) }
func Value(value string) mx.Attrib       { return mx.NewAttrib("value", value) }
func Width(pixels int) mx.Attrib         { return mx.NewAttrib("width", strconv.Itoa(pixels)) }
func Wrap(value string) mx.Attrib        { return mx.NewAttrib("wrap", value) }

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
