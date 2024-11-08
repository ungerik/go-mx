package html

import (
	"strconv"
	"strings"

	"github.com/ungerik/go-mx"
)

func Attrib(name, value string) mx.Attrib {
	return mx.Attrib{Name: name, Value: value}
}

func BoolAttrib(name string) mx.Attrib {
	return mx.Attrib{Name: name, Value: name}
}

// See https://github.com/jozo/all-html-elements-and-attributes
// and https://html.spec.whatwg.org/multipage/indices.html#attributes-3

func Accept(contentTypes ...string) mx.Attrib {
	return Attrib("accept", strings.Join(contentTypes, ","))
}
func AcceptCharset(charsets ...string) mx.Attrib {
	return Attrib("accept-charset", strings.Join(charsets, " "))
}
func AccessKey(value string) mx.Attrib    { return Attrib("accesskey", value) }
func Action(url string) mx.Attrib         { return Attrib("action", url) }
func Align(value string) mx.Attrib        { return Attrib("align", value) }
func Allow(value string) mx.Attrib        { return Attrib("allow", value) }
func Alpha() mx.Attrib                    { return Attrib("alpha", "alpha") }
func Alt(text string) mx.Attrib           { return Attrib("alt", text) }
func As(value string) mx.Attrib           { return Attrib("as", value) }
func Async() mx.Attrib                    { return Attrib("async", "async") }
func AutoCapitalizeNone() mx.Attrib       { return Attrib("autocapitalize", "none") }
func AutoCapitalizeSentences() mx.Attrib  { return Attrib("autocapitalize", "sentences") }
func AutoCapitalizeWords() mx.Attrib      { return Attrib("autocapitalize", "words") }
func AutoCapitalizeCharacters() mx.Attrib { return Attrib("autocapitalize", "characters") }
func AutoComplete(tokens ...string) mx.Attrib {
	if len(tokens) == 0 {
		return AutoCompleteOn()
	}
	return Attrib("autocomplete", strings.Join(tokens, " "))
}
func AutoCompleteOn() mx.Attrib         { return Attrib("autocomplete", "on") }
func AutoCompleteOff() mx.Attrib        { return Attrib("autocomplete", "off") }
func AutoCorrectOn() mx.Attrib          { return Attrib("autocorrect", "on") }
func AutoCorrectOff() mx.Attrib         { return Attrib("autocorrect", "off") }
func AutoFocus() mx.Attrib              { return Attrib("autofocus", "autofocus") }
func AutoPlay() mx.Attrib               { return Attrib("autoplay", "autoplay") }
func Background(value string) mx.Attrib { return Attrib("background", value) }
func BGColor(value string) mx.Attrib    { return Attrib("bgcolor", value) }
func Border(value string) mx.Attrib     { return Attrib("border", value) }
func Capture(value string) mx.Attrib    { return Attrib("capture", value) }
func CharSet(value string) mx.Attrib    { return Attrib("charset", value) }
func Checked() mx.Attrib                { return Attrib("checked", "checked") }
func CiteAttr(value string) mx.Attrib   { return Attrib("cite", value) }
func Class(classes ...string) mx.Attrib { return Attrib("class", strings.Join(classes, " ")) }
func Color(value string) mx.Attrib      { return Attrib("color", value) }
func Cols(numChars int) mx.Attrib       { return Attrib("cols", strconv.Itoa(numChars)) }
func ColSpan(numCols int) mx.Attrib     { return Attrib("colspan", strconv.Itoa(numCols)) }
func ContentAttr(text string) mx.Attrib { return Attrib("content", text) }
func ContentEditableTrue() mx.Attrib    { return Attrib("contenteditable", "true") }
func ContentEditableFalse() mx.Attrib   { return Attrib("contenteditable", "false") }
func ContentEditablePlaintextOnly(value string) mx.Attrib {
	return Attrib("contenteditable", "plaintext-only")
}
func Controls() mx.Attrib { return Attrib("controls", "controls") }
func Coords(coords ...float64) mx.Attrib {
	var b strings.Builder
	for i, coord := range coords {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.FormatFloat(coord, 'f', -1, 64))
	}
	return Attrib("coords", b.String())
}
func CrossOriginAnonymous() mx.Attrib       { return Attrib("crossorigin", "anonymous") }
func CrossOriginUseCredentials() mx.Attrib  { return Attrib("crossorigin", "use-credentials") }
func CSP(value string) mx.Attrib            { return Attrib("csp", value) }
func DataAttr(name, value string) mx.Attrib { return Attrib("data-"+name, value) }
func Datetime(value string) mx.Attrib       { return Attrib("datetime", value) }
func DecodingAuto() mx.Attrib               { return Attrib("decoding", "auto") }
func DecodingAsync() mx.Attrib              { return Attrib("decoding", "async") }
func DecodingSync() mx.Attrib               { return Attrib("decoding", "sync") }
func Default() mx.Attrib                    { return Attrib("default", "default") }
func Defer() mx.Attrib                      { return Attrib("defer", "defer") }
func DirLTR() mx.Attrib                     { return Attrib("dir", "ltr") }
func DirRTL() mx.Attrib                     { return Attrib("dir", "rtl") }
func DirAuto() mx.Attrib                    { return Attrib("dir", "auto") }
func DirName(name string) mx.Attrib         { return Attrib("dirname", name) }
func Disabled() mx.Attrib                   { return Attrib("disabled", "disabled") }
func Download(filename string) mx.Attrib    { return Attrib("download", filename) }
func Draggable(value bool) mx.Attrib        { return Attrib("draggable", strconv.FormatBool(value)) }
func EncTypeFormURLEndoced(value string) mx.Attrib {
	return Attrib("enctype", "application/x-www-form-urlencoded")
}
func EncTypeMultipartFormData(value string) mx.Attrib {
	return Attrib("enctype", "multipart/form-data")
}
func EncTypeTextPlain(value string) mx.Attrib { return Attrib("enctype", "text/plain") }
func EnterKeyHintEnter() mx.Attrib            { return Attrib("enterkeyhint", "enter") }
func EnterKeyHintDone() mx.Attrib             { return Attrib("enterkeyhint", "done") }
func EnterKeyHintGo() mx.Attrib               { return Attrib("enterkeyhint", "go") }
func EnterKeyHintNext() mx.Attrib             { return Attrib("enterkeyhint", "next") }
func EnterKeyHintPrevious() mx.Attrib         { return Attrib("enterkeyhint", "previous") }
func EnterKeyHintSearch() mx.Attrib           { return Attrib("enterkeyhint", "search") }
func EnterKeyHintSend() mx.Attrib             { return Attrib("enterkeyhint", "send") }
func FetchPriorityAuto() mx.Attrib            { return Attrib("fetchpriority", "auto") }
func FetchPriorityHigh() mx.Attrib            { return Attrib("fetchpriority", "high") }
func FetchPriorityLow() mx.Attrib             { return Attrib("fetchpriority", "low") }
func For(id string) mx.Attrib                 { return Attrib("for", id) }
func FormAttr(formID string) mx.Attrib        { return Attrib("form", formID) }
func FormAction(url string) mx.Attrib         { return Attrib("formaction", url) }
func FormEncTypeFormURLEndoced(value string) mx.Attrib {
	return Attrib("formenctype", "application/x-www-form-urlencoded")
}
func FormEncTypeMultipartFormData(value string) mx.Attrib {
	return Attrib("formenctype", "multipart/form-data")
}
func FormEncTypeTextPlain(value string) mx.Attrib { return Attrib("formenctype", "text/plain") }
func FormMethodGET() mx.Attrib                    { return Attrib("formmethod", "get") }
func FormMethodPOST() mx.Attrib                   { return Attrib("formmethod", "post") }
func FormMethodDialog() mx.Attrib                 { return Attrib("formmethod", "dialog") }
func FormNoValidate() mx.Attrib                   { return Attrib("formnovalidate", "formnovalidate") }
func FormTarget(value string) mx.Attrib           { return Attrib("formtarget", value) }
func Headers(headerCellIDs ...string) mx.Attrib {
	return Attrib("headers", strings.Join(headerCellIDs, " "))
}
func Height(value string) mx.Attrib { return Attrib("height", value) }
func Hidden() mx.Attrib             { return Attrib("hidden", "hidden") }
func HiddenUntilFound() mx.Attrib   { return Attrib("hidden", "until-found") }
func High(limit float64) mx.Attrib {
	return Attrib("high", strconv.FormatFloat(limit, 'f', -1, 64))
}
func Hight(pixels int) mx.Attrib        { return Attrib("high", strconv.Itoa(pixels)) }
func HRef(url string) mx.Attrib         { return Attrib("href", url) }
func HRefLang(value string) mx.Attrib   { return Attrib("hreflang", value) }
func HTTPEquivContentType() mx.Attrib   { return Attrib("http-equiv", "content-type") }
func HTTPEquivDefaultStyle() mx.Attrib  { return Attrib("http-equiv", "default-style") }
func HTTPEquivRefresh() mx.Attrib       { return Attrib("http-equiv", "refresh") }
func HTTPEquivXUACompatible() mx.Attrib { return Attrib("http-equiv", "x-ua-compatible") }
func HTTPEquivContentSecurityPolicy() mx.Attrib {
	return Attrib("http-equiv", "content-security-policy")
}
func ID(value string) mx.Attrib { return Attrib("id", value) }

// imagesizes ?
// imagesrcset ?

func Inert() mx.Attrib                     { return Attrib("inert", "inert") }
func InputModeNone() mx.Attrib             { return Attrib("inputmode", "none") }
func InputModeText() mx.Attrib             { return Attrib("inputmode", "text") }
func InputModeTel() mx.Attrib              { return Attrib("inputmode", "tel") }
func InputModeEmail() mx.Attrib            { return Attrib("inputmode", "email") }
func InputModeURL() mx.Attrib              { return Attrib("inputmode", "url") }
func InputModeNumeric() mx.Attrib          { return Attrib("inputmode", "numeric") }
func InputModeDecimal() mx.Attrib          { return Attrib("inputmode", "decimal") }
func InputModeSearch() mx.Attrib           { return Attrib("inputmode", "search") }
func Integrity(value string) mx.Attrib     { return Attrib("integrity", value) }
func IntrinsicSize(value string) mx.Attrib { return Attrib("intrinsicsize", value) }
func IsMap() mx.Attrib                     { return Attrib("ismap", "ismap") }
func ItemID(url string) mx.Attrib          { return Attrib("itemid", url) }
func ItemProp(props ...string) mx.Attrib   { return Attrib("itemprop", strings.Join(props, " ")) }
func ItemRef(ids ...string) mx.Attrib      { return Attrib("itemref", strings.Join(ids, " ")) }
func ItemScope() mx.Attrib                 { return Attrib("itemscope", "itemscope") }
func ItemType(urls ...string) mx.Attrib    { return Attrib("itemtype", strings.Join(urls, " ")) }
func KindSubtitles() mx.Attrib             { return Attrib("kind", "subtitles") }
func KindCaptions() mx.Attrib              { return Attrib("kind", "captions") }
func KindDescriptions() mx.Attrib          { return Attrib("kind", "descriptions") }
func KindChapters() mx.Attrib              { return Attrib("kind", "chapters") }
func KindMetadata() mx.Attrib              { return Attrib("kind", "metadata") }
func LabelAttr(value string) mx.Attrib     { return Attrib("label", value) }
func Lang(value string) mx.Attrib          { return Attrib("lang", value) }
func Language(value string) mx.Attrib      { return Attrib("language", value) }
func List(id string) mx.Attrib             { return Attrib("list", id) }
func LoadingEager(value string) mx.Attrib  { return Attrib("loading", "eager") }
func LoadingLazy(value string) mx.Attrib   { return Attrib("loading", "lazy") }
func Loop() mx.Attrib                      { return Attrib("loop", "loop") }
func Low(limit float64) mx.Attrib          { return Attrib("low", strconv.FormatFloat(limit, 'f', -1, 64)) }
func Max(value string) mx.Attrib           { return Attrib("max", value) }
func MaxLength(length int) mx.Attrib       { return Attrib("maxlength", strconv.Itoa(length)) }

func Media(query string) mx.Attrib   { return Attrib("media", query) }
func MethodGET() mx.Attrib           { return Attrib("method", "GET") }
func MethodPOST() mx.Attrib          { return Attrib("method", "POST") }
func MethodDialog() mx.Attrib        { return Attrib("method", "dialog") }
func Min(value string) mx.Attrib     { return Attrib("min", value) }
func MinLength(length int) mx.Attrib { return Attrib("minlength", strconv.Itoa(length)) }
func Multiple() mx.Attrib            { return BoolAttrib("multiple") }
func Muted() mx.Attrib               { return BoolAttrib("muted") }
func Name(value string) mx.Attrib    { return Attrib("name", value) }
func NoModule() mx.Attrib            { return BoolAttrib("nomodule") }
func Nonce(value string) mx.Attrib   { return Attrib("nonce", value) }
func NoValidate() mx.Attrib          { return BoolAttrib("novalidate") }
func Open() mx.Attrib                { return BoolAttrib("open") }
func Optimum(value float64) mx.Attrib {
	return Attrib("optimum", strconv.FormatFloat(value, 'f', -1, 64))
}
func Pattern(value string) mx.Attrib      { return Attrib("pattern", value) }
func Ping(value string) mx.Attrib         { return Attrib("ping", value) }
func Placeholder(value string) mx.Attrib  { return Attrib("placeholder", value) }
func PlaysInline(value string) mx.Attrib  { return Attrib("playsinline", value) }
func Poster(value string) mx.Attrib       { return Attrib("poster", value) }
func Preload(value string) mx.Attrib      { return Attrib("preload", value) }
func Readonly(value string) mx.Attrib     { return Attrib("readonly", value) }
func ReferrerPolicyNoReferrer() mx.Attrib { return Attrib("referrerpolicy", "no-referrer") }
func ReferrerPolicyNoReferrerWhenDowngrade() mx.Attrib {
	return Attrib("referrerpolicy", "no-referrer-when-downgrade")
}
func ReferrerPolicyOrigin() mx.Attrib { return Attrib("referrerpolicy", "origin") }
func ReferrerPolicyOriginWhenCrossOrigin() mx.Attrib {
	return Attrib("referrerpolicy", "origin-when-cross-origin")
}
func ReferrerPolicySameOrigin() mx.Attrib { return Attrib("referrerpolicy", "same-origin") }
func ReferrerPolicyStrictOrigin() mx.Attrib {
	return Attrib("referrerpolicy", "strict-origin")
}
func ReferrerPolicyStrictOriginWhenCrossOrigin() mx.Attrib {
	return Attrib("referrerpolicy", "strict-origin-when-cross-origin")
}
func ReferrerPolicyUnsafeUrl() mx.Attrib { return Attrib("referrerpolicy", "unsafe-url") }
func Rel(keywords ...string) mx.Attrib   { return Attrib("rel", strings.Join(keywords, " ")) }
func Required(value string) mx.Attrib    { return Attrib("required", value) }
func Reversed(value string) mx.Attrib    { return Attrib("reversed", value) }
func Role(value string) mx.Attrib        { return Attrib("role", value) }
func Rows(value string) mx.Attrib        { return Attrib("rows", value) }
func RowSpan(value string) mx.Attrib     { return Attrib("rowspan", value) }
func Sandbox(value string) mx.Attrib     { return Attrib("sandbox", value) }
func Scope(value string) mx.Attrib       { return Attrib("scope", value) }
func Scoped(value string) mx.Attrib      { return Attrib("scoped", value) }
func Selected(value string) mx.Attrib    { return Attrib("selected", value) }
func ShapeDefault() mx.Attrib            { return Attrib("shape", "default") }
func ShapeRect() mx.Attrib               { return Attrib("shape", "rect") }
func ShapeCircle() mx.Attrib             { return Attrib("shape", "circle") }
func ShapePoly() mx.Attrib               { return Attrib("shape", "poly") }
func Size(value string) mx.Attrib        { return Attrib("size", value) }
func Sizes(sourceSizes ...string) mx.Attrib {
	return Attrib("sizes", strings.Join(sourceSizes, ","))
}
func SlotAttr(value string) mx.Attrib   { return Attrib("slot", value) }
func SpanAttr(value string) mx.Attrib   { return Attrib("span", value) }
func SpellCheck(value string) mx.Attrib { return Attrib("spellcheck", value) }
func Src(url string) mx.Attrib          { return Attrib("src", url) }
func SrcDoc(value string) mx.Attrib     { return Attrib("srcdoc", value) }
func SrcLang(value string) mx.Attrib    { return Attrib("srclang", value) }
func SrcSet(sources ...string) mx.Attrib {
	return Attrib("srcset", strings.Join(sources, ","))
}
func Start(value string) mx.Attrib       { return Attrib("start", value) }
func Step(value string) mx.Attrib        { return Attrib("step", value) }
func Style(value string) mx.Attrib       { return Attrib("style", value) }
func TabIndex(value string) mx.Attrib    { return Attrib("tabindex", value) }
func Target(value string) mx.Attrib      { return Attrib("target", value) }
func TargetSelf() mx.Attrib              { return Attrib("target", "_self") }
func TargetBlank() mx.Attrib             { return Attrib("target", "_blank") }
func TargetParent() mx.Attrib            { return Attrib("target", "_parent") }
func TargetTop() mx.Attrib               { return Attrib("target", "_top") }
func TargetUnfencedTop() mx.Attrib       { return Attrib("target", "_unfencedTop") }
func Title(value string) mx.Attrib       { return Attrib("title", value) }
func Translate(value string) mx.Attrib   { return Attrib("translate", value) }
func Type(value string) mx.Attrib        { return Attrib("type", value) }
func UseMap(partialURL string) mx.Attrib { return Attrib("usemap", partialURL) }
func Value(value string) mx.Attrib       { return Attrib("value", value) }
func Width(pixels int) mx.Attrib         { return Attrib("width", strconv.Itoa(pixels)) }
func Wrap(value string) mx.Attrib        { return Attrib("wrap", value) }

// Event handlers, see https://html.spec.whatwg.org/multipage/indices.html#events-2

// OnAfterPrint `afterprint` event handler for Window object (body element)
func OnAfterPrint(execute string) mx.Attrib { return Attrib("onafterprint", execute) }

// OnAuxClick `auxclick` event handler (all HTML elements)
func OnAuxClick(execute string) mx.Attrib { return Attrib("onauxclick", execute) }

// OnBeforeInput `beforeinput` event handler (all HTML elements)
func OnBeforeInput(execute string) mx.Attrib { return Attrib("onbeforeinput", execute) }

// OnBeforeMatch `beforematch` event handler (all HTML elements)
func OnBeforeMatch(execute string) mx.Attrib { return Attrib("onbeforematch", execute) }

// OnBeforePrint `beforeprint` event handler for Window object (body element)
func OnBeforePrint(execute string) mx.Attrib { return Attrib("onbeforeprint", execute) }

// OnBeforeUnload `beforeunload` event handler for Window object (body element)
func OnBeforeUnload(execute string) mx.Attrib { return Attrib("onbeforeunload", execute) }

// OnBeforeToggle `beforetoggle` event handler (all HTML elements)
func OnBeforeToggle(execute string) mx.Attrib { return Attrib("onbeforetoggle", execute) }

// OnBlur `blur` event handler (all HTML elements)
func OnBlur(execute string) mx.Attrib { return Attrib("onblur", execute) }

// OnCancel `cancel` event handler (all HTML elements)
func OnCancel(execute string) mx.Attrib { return Attrib("oncancel", execute) }

// OnCanplay `canplay` event handler (all HTML elements)
func OnCanplay(execute string) mx.Attrib { return Attrib("oncanplay", execute) }

// OnCanPlayThrough `canplaythrough` event handler (all HTML elements)
func OnCanPlayThrough(execute string) mx.Attrib { return Attrib("oncanplaythrough", execute) }

// OnChange `change` event handler (all HTML elements)
func OnChange(execute string) mx.Attrib { return Attrib("onchange", execute) }

// OnClick `click` event handler (all HTML elements)
func OnClick(execute string) mx.Attrib { return Attrib("onclick", execute) }

// OnClose `close` event handler (all HTML elements)
func OnClose(execute string) mx.Attrib { return Attrib("onclose", execute) }

// OnContextLost `contextlost` event handler (all HTML elements)
func OnContextLost(execute string) mx.Attrib { return Attrib("oncontextlost", execute) }

// OnContextMenu `contextmenu` event handler (all HTML elements)
func OnContextMenu(execute string) mx.Attrib { return Attrib("oncontextmenu", execute) }

// OnContextRestored `contextrestored` event handler (all HTML elements)
func OnContextRestored(execute string) mx.Attrib { return Attrib("oncontextrestored", execute) }

// OnCopy `copy` event handler (all HTML elements)
func OnCopy(execute string) mx.Attrib { return Attrib("oncopy", execute) }

// OnCueChange `cuechange` event handler (all HTML elements)
func OnCueChange(execute string) mx.Attrib { return Attrib("oncuechange", execute) }

// OnCut `cut` event handler (all HTML elements)
func OnCut(execute string) mx.Attrib { return Attrib("oncut", execute) }

// OnDblClick `dblclick` event handler (all HTML elements)
func OnDblClick(execute string) mx.Attrib { return Attrib("ondblclick", execute) }

// OnDrag `drag` event handler (all HTML elements)
func OnDrag(execute string) mx.Attrib { return Attrib("ondrag", execute) }

// OnDragEnd `dragend` event handler (all HTML elements)
func OnDragEnd(execute string) mx.Attrib { return Attrib("ondragend", execute) }

// OnDragEnter `dragenter` event handler (all HTML elements)
func OnDragEnter(execute string) mx.Attrib { return Attrib("ondragenter", execute) }

// OnDragLeave `dragleave` event handler (all HTML elements)
func OnDragLeave(execute string) mx.Attrib { return Attrib("ondragleave", execute) }

// OnDragOver `dragover` event handler (all HTML elements)
func OnDragOver(execute string) mx.Attrib { return Attrib("ondragover", execute) }

// OnDragStart `dragstart` event handler (all HTML elements)
func OnDragStart(execute string) mx.Attrib { return Attrib("ondragstart", execute) }

// OnDrop `drop` event handler (all HTML elements)
func OnDrop(execute string) mx.Attrib { return Attrib("ondrop", execute) }

// OnDurationChange `durationchange` event handler (all HTML elements)
func OnDurationChange(execute string) mx.Attrib { return Attrib("ondurationchange", execute) }

// OnEmptied `emptied` event handler (all HTML elements)
func OnEmptied(execute string) mx.Attrib { return Attrib("onemptied", execute) }

// OnEnded `ended` event handler (all HTML elements)
func OnEnded(execute string) mx.Attrib { return Attrib("onended", execute) }

// OnError `error` event handler (all HTML elements)
func OnError(execute string) mx.Attrib { return Attrib("onerror", execute) }

// OnFocus `focus` event handler (all HTML elements)
func OnFocus(execute string) mx.Attrib { return Attrib("onfocus", execute) }

// OnFormData `formdata` event handler (all HTML elements)
func OnFormData(execute string) mx.Attrib { return Attrib("onformdata", execute) }

// OnHashChange `hashchange` event handler for Window object (body element)
func OnHashChange(execute string) mx.Attrib { return Attrib("onhashchange", execute) }

// OnInput `input` event handler (all HTML elements)
func OnInput(execute string) mx.Attrib { return Attrib("oninput", execute) }

// OnInvalid `invalid` event handler (all HTML elements)
func OnInvalid(execute string) mx.Attrib { return Attrib("oninvalid", execute) }

// OnKeyDown `keydown` event handler (all HTML elements)
func OnKeyDown(execute string) mx.Attrib { return Attrib("onkeydown", execute) }

// OnKeyPress `keypress` event handler (all HTML elements)
func OnKeyPress(execute string) mx.Attrib { return Attrib("onkeypress", execute) }

// OnKeyUp `keyup` event handler (all HTML elements)
func OnKeyUp(execute string) mx.Attrib { return Attrib("onkeyup", execute) }

// OnLanguageChange `languagechange` event handler for Window object (body element)
func OnLanguageChange(execute string) mx.Attrib { return Attrib("onlanguagechange", execute) }

// OnLoad `load` event handler (all HTML elements)
func OnLoad(execute string) mx.Attrib { return Attrib("onload", execute) }

// OnLoadedData `loadeddata` event handler (all HTML elements)
func OnLoadedData(execute string) mx.Attrib { return Attrib("onloadeddata", execute) }

// OnLoadedMetadata `loadedmetadata` event handler (all HTML elements)
func OnLoadedMetadata(execute string) mx.Attrib { return Attrib("onloadedmetadata", execute) }

// OnLoadStart `loadstart` event handler (all HTML elements)
func OnLoadStart(execute string) mx.Attrib { return Attrib("onloadstart", execute) }

// OnMessage `message` event handler for Window object (body element)
func OnMessage(execute string) mx.Attrib { return Attrib("onmessage", execute) }

// OnMessageError `messageerror` event handler for Window object (body element)
func OnMessageError(execute string) mx.Attrib { return Attrib("onmessageerror", execute) }

// OnMouseDown `mousedown` event handler (all HTML elements)
func OnMouseDown(execute string) mx.Attrib { return Attrib("onmousedown", execute) }

// OnMouseEnter `mouseenter` event handler (all HTML elements)
func OnMouseEnter(execute string) mx.Attrib { return Attrib("onmouseenter", execute) }

// OnMouseLeave `mouseleave` event handler (all HTML elements)
func OnMouseLeave(execute string) mx.Attrib { return Attrib("onmouseleave", execute) }

// OnMouseMove `mousemove` event handler (all HTML elements)
func OnMouseMove(execute string) mx.Attrib { return Attrib("onmousemove", execute) }

// OnMouseOut `mouseout` event handler (all HTML elements)
func OnMouseOut(execute string) mx.Attrib { return Attrib("onmouseout", execute) }

// OnMouseOver `mouseover` event handler (all HTML elements)
func OnMouseOver(execute string) mx.Attrib { return Attrib("onmouseover", execute) }

// OnMouseUp `mouseup` event handler (all HTML elements)
func OnMouseUp(execute string) mx.Attrib { return Attrib("onmouseup", execute) }

// OnOffline `offline` event handler for Window object (body element)
func OnOffline(execute string) mx.Attrib { return Attrib("onoffline", execute) }

// OnOnline `online` event handler for Window object (body element)
func OnOnline(execute string) mx.Attrib { return Attrib("ononline", execute) }

// OnPageHide `pagehide` event handler for Window object (body element)
func OnPageHide(execute string) mx.Attrib { return Attrib("onpagehide", execute) }

// OnPageReveal `pagereveal` event handler for Window object (body element)
func OnPageReveal(execute string) mx.Attrib { return Attrib("onpagereveal", execute) }

// OnPageShow `pageshow` event handler for Window object (body element)
func OnPageShow(execute string) mx.Attrib { return Attrib("onpageshow", execute) }

// OnPageSwap `pageswap` event handler for Window object (body element)
func OnPageSwap(execute string) mx.Attrib { return Attrib("onpageswap", execute) }

// OnPaste `paste` event handler (all HTML elements)
func OnPaste(execute string) mx.Attrib { return Attrib("onpaste", execute) }

// OnPause `pause` event handler (all HTML elements)
func OnPause(execute string) mx.Attrib { return Attrib("onpause", execute) }

// OnPlay `play` event handler (all HTML elements)
func OnPlay(execute string) mx.Attrib { return Attrib("onplay", execute) }

// OnPlaying `playing` event handler (all HTML elements)
func OnPlaying(execute string) mx.Attrib { return Attrib("onplaying", execute) }

// OnPopState `popstate` event handler for Window object (body element)
func OnPopState(execute string) mx.Attrib { return Attrib("onpopstate", execute) }

// OnProgress `progress` event handler (all HTML elements)
func OnProgress(execute string) mx.Attrib { return Attrib("onprogress", execute) }

// OnRateChange `ratechange` event handler (all HTML elements)
func OnRateChange(execute string) mx.Attrib { return Attrib("onratechange", execute) }

// OnReset `reset` event handler (all HTML elements)
func OnReset(execute string) mx.Attrib { return Attrib("onreset", execute) }

// OnResize `resize` event handler (all HTML elements)
func OnResize(execute string) mx.Attrib { return Attrib("onresize", execute) }

// OnRejectionHandled `rejectionhandled` event handler for Window object (body element)
func OnRejectionHandled(execute string) mx.Attrib { return Attrib("onrejectionhandled", execute) }

// OnScroll `scroll` event handler (all HTML elements)
func OnScroll(execute string) mx.Attrib { return Attrib("onscroll", execute) }

// OnScrollEnd `scrollend` event handler (all HTML elements)
func OnScrollEnd(execute string) mx.Attrib { return Attrib("onscrollend", execute) }

// OnSecurityPolicyViolation `securitypolicyviolation` event handler (all HTML elements)
func OnSecurityPolicyViolation(execute string) mx.Attrib {
	return Attrib("onsecuritypolicyviolation", execute)
}

// OnSeeked `seeked` event handler (all HTML elements)
func OnSeeked(execute string) mx.Attrib { return Attrib("onseeked", execute) }

// OnSeeking `seeking` event handler (all HTML elements)
func OnSeeking(execute string) mx.Attrib { return Attrib("onseeking", execute) }

// OnSelect `select` event handler (all HTML elements)
func OnSelect(execute string) mx.Attrib { return Attrib("onselect", execute) }

// OnSlotChange `slotchange` event handler (all HTML elements)
func OnSlotChange(execute string) mx.Attrib { return Attrib("onslotchange", execute) }

// OnStalled `stalled` event handler (all HTML elements)
func OnStalled(execute string) mx.Attrib { return Attrib("onstalled", execute) }

// OnStorage `storage` event handler for Window object (body element)
func OnStorage(execute string) mx.Attrib { return Attrib("onstorage", execute) }

// OnSubmit `submit` event handler (all HTML elements)
func OnSubmit(execute string) mx.Attrib { return Attrib("onsubmit", execute) }

// OnSuspend `suspend` event handler (all HTML elements)
func OnSuspend(execute string) mx.Attrib { return Attrib("onsuspend", execute) }

// OnTimeUpdate `timeupdate` event handler (all HTML elements)
func OnTimeUpdate(execute string) mx.Attrib { return Attrib("ontimeupdate", execute) }

// OnToggle `toggle` event handler (all HTML elements)
func OnToggle(execute string) mx.Attrib { return Attrib("ontoggle", execute) }

// OnUnhandledRejection `unhandledrejection` event handler for Window object (body element)
func OnUnhandledRejection(execute string) mx.Attrib {
	return Attrib("onunhandledrejection", execute)
}

// OnUnload `unload` event handler for Window object (body element)
func OnUnload(execute string) mx.Attrib { return Attrib("onunload", execute) }

// OnVolumeChange `volumechange` event handler (all HTML elements)
func OnVolumeChange(execute string) mx.Attrib { return Attrib("onvolumechange", execute) }

// OnWaiting `waiting` event handler (all HTML elements)
func OnWaiting(execute string) mx.Attrib { return Attrib("onwaiting", execute) }

// OnWheel `wheel` event handler (all HTML elements)
func OnWheel(execute string) mx.Attrib { return Attrib("onwheel", execute) }
