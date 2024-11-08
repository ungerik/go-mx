package html

import (
	"strconv"
	"strings"

	"github.com/ungerik/go-mx"
)

type Attrib = mx.Attrib

func Attribute(name, value string) Attrib {
	return mx.Attribute(name, value)
}

// See https://github.com/jozo/all-html-elements-and-attributes
// and https://html.spec.whatwg.org/multipage/indices.html#attributes-3

func Accept(contentTypes ...string) Attrib {
	return mx.Attribute("accept", strings.Join(contentTypes, ","))
}
func AcceptCharset(charsets ...string) Attrib {
	return mx.Attribute("accept-charset", strings.Join(charsets, " "))
}
func AccessKey(value string) Attrib    { return mx.Attribute("accesskey", value) }
func Action(url string) Attrib         { return mx.Attribute("action", url) }
func Align(value string) Attrib        { return mx.Attribute("align", value) }
func Allow(value string) Attrib        { return mx.Attribute("allow", value) }
func Alpha() Attrib                    { return mx.Attribute("alpha", "alpha") }
func Alt(text string) Attrib           { return mx.Attribute("alt", text) }
func As(value string) Attrib           { return mx.Attribute("as", value) }
func Async() Attrib                    { return mx.Attribute("async", "async") }
func AutoCapitalizeNone() Attrib       { return mx.Attribute("autocapitalize", "none") }
func AutoCapitalizeSentences() Attrib  { return mx.Attribute("autocapitalize", "sentences") }
func AutoCapitalizeWords() Attrib      { return mx.Attribute("autocapitalize", "words") }
func AutoCapitalizeCharacters() Attrib { return mx.Attribute("autocapitalize", "characters") }
func AutoComplete(tokens ...string) Attrib {
	if len(tokens) == 0 {
		return AutoCompleteOn()
	}
	return mx.Attribute("autocomplete", strings.Join(tokens, " "))
}
func AutoCompleteOn() Attrib         { return mx.Attribute("autocomplete", "on") }
func AutoCompleteOff() Attrib        { return mx.Attribute("autocomplete", "off") }
func AutoCorrectOn() Attrib          { return mx.Attribute("autocorrect", "on") }
func AutoCorrectOff() Attrib         { return mx.Attribute("autocorrect", "off") }
func AutoFocus() Attrib              { return mx.Attribute("autofocus", "autofocus") }
func AutoPlay() Attrib               { return mx.Attribute("autoplay", "autoplay") }
func Background(value string) Attrib { return mx.Attribute("background", value) }
func BGColor(value string) Attrib    { return mx.Attribute("bgcolor", value) }
func Border(value string) Attrib     { return mx.Attribute("border", value) }
func Capture(value string) Attrib    { return mx.Attribute("capture", value) }
func CharSet(value string) Attrib    { return mx.Attribute("charset", value) }
func Checked() Attrib                { return mx.Attribute("checked", "checked") }
func CiteAttr(value string) Attrib   { return mx.Attribute("cite", value) }
func Class(classes ...string) Attrib { return mx.Attribute("class", strings.Join(classes, " ")) }
func Color(value string) Attrib      { return mx.Attribute("color", value) }
func Cols(numChars int) Attrib       { return mx.Attribute("cols", strconv.Itoa(numChars)) }
func ColSpan(numCols int) Attrib     { return mx.Attribute("colspan", strconv.Itoa(numCols)) }
func ContentAttr(text string) Attrib { return mx.Attribute("content", text) }
func ContentEditableTrue() Attrib    { return mx.Attribute("contenteditable", "true") }
func ContentEditableFalse() Attrib   { return mx.Attribute("contenteditable", "false") }
func ContentEditablePlaintextOnly(value string) Attrib {
	return mx.Attribute("contenteditable", "plaintext-only")
}
func Controls() Attrib { return mx.Attribute("controls", "controls") }
func Coords(coords ...float64) Attrib {
	var b strings.Builder
	for i, coord := range coords {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.FormatFloat(coord, 'f', -1, 64))
	}
	return mx.Attribute("coords", b.String())
}
func CrossOriginAnonymous() Attrib       { return mx.Attribute("crossorigin", "anonymous") }
func CrossOriginUseCredentials() Attrib  { return mx.Attribute("crossorigin", "use-credentials") }
func CSP(value string) Attrib            { return mx.Attribute("csp", value) }
func DataAttr(name, value string) Attrib { return mx.Attribute("data-"+name, value) }
func Datetime(value string) Attrib       { return mx.Attribute("datetime", value) }
func DecodingAuto() Attrib               { return mx.Attribute("decoding", "auto") }
func DecodingAsync() Attrib              { return mx.Attribute("decoding", "async") }
func DecodingSync() Attrib               { return mx.Attribute("decoding", "sync") }
func Default() Attrib                    { return mx.Attribute("default", "default") }
func Defer() Attrib                      { return mx.Attribute("defer", "defer") }
func DirLTR() Attrib                     { return mx.Attribute("dir", "ltr") }
func DirRTL() Attrib                     { return mx.Attribute("dir", "rtl") }
func DirAuto() Attrib                    { return mx.Attribute("dir", "auto") }
func DirName(name string) Attrib         { return mx.Attribute("dirname", name) }
func Disabled() Attrib                   { return mx.Attribute("disabled", "disabled") }
func Download(filename string) Attrib    { return mx.Attribute("download", filename) }
func Draggable(value bool) Attrib        { return mx.Attribute("draggable", strconv.FormatBool(value)) }
func EncTypeFormURLEndoced(value string) Attrib {
	return mx.Attribute("enctype", "application/x-www-form-urlencoded")
}
func EncTypeMultipartFormData(value string) Attrib {
	return mx.Attribute("enctype", "multipart/form-data")
}
func EncTypeTextPlain(value string) Attrib { return mx.Attribute("enctype", "text/plain") }
func EnterKeyHintEnter() Attrib            { return mx.Attribute("enterkeyhint", "enter") }
func EnterKeyHintDone() Attrib             { return mx.Attribute("enterkeyhint", "done") }
func EnterKeyHintGo() Attrib               { return mx.Attribute("enterkeyhint", "go") }
func EnterKeyHintNext() Attrib             { return mx.Attribute("enterkeyhint", "next") }
func EnterKeyHintPrevious() Attrib         { return mx.Attribute("enterkeyhint", "previous") }
func EnterKeyHintSearch() Attrib           { return mx.Attribute("enterkeyhint", "search") }
func EnterKeyHintSend() Attrib             { return mx.Attribute("enterkeyhint", "send") }
func FetchPriorityAuto() Attrib            { return mx.Attribute("fetchpriority", "auto") }
func FetchPriorityHigh() Attrib            { return mx.Attribute("fetchpriority", "high") }
func FetchPriorityLow() Attrib             { return mx.Attribute("fetchpriority", "low") }
func For(id string) Attrib                 { return mx.Attribute("for", id) }
func FormAttr(formID string) Attrib        { return mx.Attribute("form", formID) }
func FormAction(url string) Attrib         { return mx.Attribute("formaction", url) }
func FormEncTypeFormURLEndoced(value string) Attrib {
	return mx.Attribute("formenctype", "application/x-www-form-urlencoded")
}
func FormEncTypeMultipartFormData(value string) Attrib {
	return mx.Attribute("formenctype", "multipart/form-data")
}
func FormEncTypeTextPlain(value string) Attrib { return mx.Attribute("formenctype", "text/plain") }
func FormMethodGET() Attrib                    { return mx.Attribute("formmethod", "get") }
func FormMethodPOST() Attrib                   { return mx.Attribute("formmethod", "post") }
func FormMethodDialog() Attrib                 { return mx.Attribute("formmethod", "dialog") }
func FormNoValidate() Attrib                   { return mx.Attribute("formnovalidate", "formnovalidate") }
func FormTarget(value string) Attrib           { return mx.Attribute("formtarget", value) }
func Headers(headerCellIDs ...string) Attrib {
	return mx.Attribute("headers", strings.Join(headerCellIDs, " "))
}
func Height(value string) Attrib { return mx.Attribute("height", value) }
func Hidden() Attrib             { return mx.Attribute("hidden", "hidden") }
func HiddenUntilFound() Attrib   { return mx.Attribute("hidden", "until-found") }
func High(limit float64) Attrib {
	return mx.Attribute("high", strconv.FormatFloat(limit, 'f', -1, 64))
}
func Hight(pixels int) Attrib        { return mx.Attribute("high", strconv.Itoa(pixels)) }
func HRef(url string) Attrib         { return mx.Attribute("href", url) }
func HRefLang(value string) Attrib   { return mx.Attribute("hreflang", value) }
func HTTPEquivContentType() Attrib   { return mx.Attribute("http-equiv", "content-type") }
func HTTPEquivDefaultStyle() Attrib  { return mx.Attribute("http-equiv", "default-style") }
func HTTPEquivRefresh() Attrib       { return mx.Attribute("http-equiv", "refresh") }
func HTTPEquivXUACompatible() Attrib { return mx.Attribute("http-equiv", "x-ua-compatible") }
func HTTPEquivContentSecurityPolicy() Attrib {
	return mx.Attribute("http-equiv", "content-security-policy")
}
func ID(value string) Attrib { return mx.Attribute("id", value) }

// imagesizes ?
// imagesrcset ?

func Inert() Attrib                     { return mx.Attribute("inert", "inert") }
func InputModeNone() Attrib             { return mx.Attribute("inputmode", "none") }
func InputModeText() Attrib             { return mx.Attribute("inputmode", "text") }
func InputModeTel() Attrib              { return mx.Attribute("inputmode", "tel") }
func InputModeEmail() Attrib            { return mx.Attribute("inputmode", "email") }
func InputModeURL() Attrib              { return mx.Attribute("inputmode", "url") }
func InputModeNumeric() Attrib          { return mx.Attribute("inputmode", "numeric") }
func InputModeDecimal() Attrib          { return mx.Attribute("inputmode", "decimal") }
func InputModeSearch() Attrib           { return mx.Attribute("inputmode", "search") }
func Integrity(value string) Attrib     { return mx.Attribute("integrity", value) }
func IntrinsicSize(value string) Attrib { return mx.Attribute("intrinsicsize", value) }
func IsMap() Attrib                     { return mx.Attribute("ismap", "ismap") }
func ItemID(url string) Attrib          { return mx.Attribute("itemid", url) }
func ItemProp(props ...string) Attrib   { return mx.Attribute("itemprop", strings.Join(props, " ")) }
func ItemRef(ids ...string) Attrib      { return mx.Attribute("itemref", strings.Join(ids, " ")) }
func ItemScope() Attrib                 { return mx.Attribute("itemscope", "itemscope") }
func ItemType(urls ...string) Attrib    { return mx.Attribute("itemtype", strings.Join(urls, " ")) }
func KindSubtitles() Attrib             { return mx.Attribute("kind", "subtitles") }
func KindCaptions() Attrib              { return mx.Attribute("kind", "captions") }
func KindDescriptions() Attrib          { return mx.Attribute("kind", "descriptions") }
func KindChapters() Attrib              { return mx.Attribute("kind", "chapters") }
func KindMetadata() Attrib              { return mx.Attribute("kind", "metadata") }
func LabelAttr(value string) Attrib     { return mx.Attribute("label", value) }
func Lang(value string) Attrib          { return mx.Attribute("lang", value) }
func Language(value string) Attrib      { return mx.Attribute("language", value) }
func List(id string) Attrib             { return mx.Attribute("list", id) }
func LoadingEager(value string) Attrib  { return mx.Attribute("loading", "eager") }
func LoadingLazy(value string) Attrib   { return mx.Attribute("loading", "lazy") }
func Loop() Attrib                      { return mx.Attribute("loop", "loop") }
func Low(limit float64) Attrib          { return mx.Attribute("low", strconv.FormatFloat(limit, 'f', -1, 64)) }
func Max(value string) Attrib           { return mx.Attribute("max", value) }
func MaxLength(length int) Attrib       { return mx.Attribute("maxlength", strconv.Itoa(length)) }

// TODO
func Media(value string) Attrib        { return mx.Attribute("media", value) }
func MethodGET() Attrib                { return mx.Attribute("method", "get") }
func MethodPOST() Attrib               { return mx.Attribute("method", "post") }
func MethodDialog() Attrib             { return mx.Attribute("method", "dialog") }
func Min(value string) Attrib          { return mx.Attribute("min", value) }
func MinLength(length int) Attrib      { return mx.Attribute("minlength", strconv.Itoa(length)) }
func Multiple(value string) Attrib     { return mx.Attribute("multiple", value) }
func Muted() Attrib                    { return mx.Attribute("muted", "muted") }
func Name(value string) Attrib         { return mx.Attribute("name", value) }
func NoValidate() Attrib               { return mx.Attribute("novalidate", "novalidate") }
func Open(value string) Attrib         { return mx.Attribute("open", value) }
func Optimum(value string) Attrib      { return mx.Attribute("optimum", value) }
func Pattern(value string) Attrib      { return mx.Attribute("pattern", value) }
func Ping(value string) Attrib         { return mx.Attribute("ping", value) }
func Placeholder(value string) Attrib  { return mx.Attribute("placeholder", value) }
func PlaysInline(value string) Attrib  { return mx.Attribute("playsinline", value) }
func Poster(value string) Attrib       { return mx.Attribute("poster", value) }
func Preload(value string) Attrib      { return mx.Attribute("preload", value) }
func Readonly(value string) Attrib     { return mx.Attribute("readonly", value) }
func ReferrerPolicyNoReferrer() Attrib { return mx.Attribute("referrerpolicy", "no-referrer") }
func ReferrerPolicyNoReferrerWhenDowngrade() Attrib {
	return mx.Attribute("referrerpolicy", "no-referrer-when-downgrade")
}
func ReferrerPolicyOrigin() Attrib { return mx.Attribute("referrerpolicy", "origin") }
func ReferrerPolicyOriginWhenCrossOrigin() Attrib {
	return mx.Attribute("referrerpolicy", "origin-when-cross-origin")
}
func ReferrerPolicySameOrigin() Attrib { return mx.Attribute("referrerpolicy", "same-origin") }
func ReferrerPolicyStrictOrigin() Attrib {
	return mx.Attribute("referrerpolicy", "strict-origin")
}
func ReferrerPolicyStrictOriginWhenCrossOrigin() Attrib {
	return mx.Attribute("referrerpolicy", "strict-origin-when-cross-origin")
}
func ReferrerPolicyUnsafeUrl() Attrib { return mx.Attribute("referrerpolicy", "unsafe-url") }
func Rel(keywords ...string) Attrib   { return mx.Attribute("rel", strings.Join(keywords, " ")) }
func Required(value string) Attrib    { return mx.Attribute("required", value) }
func Reversed(value string) Attrib    { return mx.Attribute("reversed", value) }
func Role(value string) Attrib        { return mx.Attribute("role", value) }
func Rows(value string) Attrib        { return mx.Attribute("rows", value) }
func RowSpan(value string) Attrib     { return mx.Attribute("rowspan", value) }
func Sandbox(value string) Attrib     { return mx.Attribute("sandbox", value) }
func Scope(value string) Attrib       { return mx.Attribute("scope", value) }
func Scoped(value string) Attrib      { return mx.Attribute("scoped", value) }
func Selected(value string) Attrib    { return mx.Attribute("selected", value) }
func ShapeDefault() Attrib            { return mx.Attribute("shape", "default") }
func ShapeRect() Attrib               { return mx.Attribute("shape", "rect") }
func ShapeCircle() Attrib             { return mx.Attribute("shape", "circle") }
func ShapePoly() Attrib               { return mx.Attribute("shape", "poly") }
func Size(value string) Attrib        { return mx.Attribute("size", value) }
func Sizes(sourceSizes ...string) Attrib {
	return mx.Attribute("sizes", strings.Join(sourceSizes, ","))
}
func SlotAttr(value string) Attrib   { return mx.Attribute("slot", value) }
func SpanAttr(value string) Attrib   { return mx.Attribute("span", value) }
func Spellcheck(value string) Attrib { return mx.Attribute("spellcheck", value) }
func Src(url string) Attrib          { return mx.Attribute("src", url) }
func SrcDoc(value string) Attrib     { return mx.Attribute("srcdoc", value) }
func SrcLang(value string) Attrib    { return mx.Attribute("srclang", value) }
func SrcSet(sources ...string) Attrib {
	return mx.Attribute("srcset", strings.Join(sources, ","))
}
func Start(value string) Attrib       { return mx.Attribute("start", value) }
func Step(value string) Attrib        { return mx.Attribute("step", value) }
func Style(value string) Attrib       { return mx.Attribute("style", value) }
func TabIndex(value string) Attrib    { return mx.Attribute("tabindex", value) }
func Target(value string) Attrib      { return mx.Attribute("target", value) }
func TargetSelf() Attrib              { return mx.Attribute("target", "_self") }
func TargetBlank() Attrib             { return mx.Attribute("target", "_blank") }
func TargetParent() Attrib            { return mx.Attribute("target", "_parent") }
func TargetTop() Attrib               { return mx.Attribute("target", "_top") }
func TargetUnfencedTop() Attrib       { return mx.Attribute("target", "_unfencedTop") }
func Title(value string) Attrib       { return mx.Attribute("title", value) }
func Translate(value string) Attrib   { return mx.Attribute("translate", value) }
func Type(value string) Attrib        { return mx.Attribute("type", value) }
func UseMap(partialURL string) Attrib { return mx.Attribute("usemap", partialURL) }
func Value(value string) Attrib       { return mx.Attribute("value", value) }
func Width(pixels int) Attrib         { return mx.Attribute("width", strconv.Itoa(pixels)) }
func Wrap(value string) Attrib        { return mx.Attribute("wrap", value) }

// Event handlers, see https://html.spec.whatwg.org/multipage/indices.html#events-2

// OnAfterPrint `afterprint` event handler for Window object (body element)
func OnAfterPrint(execute string) Attrib { return mx.Attribute("onafterprint", execute) }

// OnAuxClick `auxclick` event handler (all HTML elements)
func OnAuxClick(execute string) Attrib { return mx.Attribute("onauxclick", execute) }

// OnBeforeInput `beforeinput` event handler (all HTML elements)
func OnBeforeInput(execute string) Attrib { return mx.Attribute("onbeforeinput", execute) }

// OnBeforeMatch `beforematch` event handler (all HTML elements)
func OnBeforeMatch(execute string) Attrib { return mx.Attribute("onbeforematch", execute) }

// OnBeforePrint `beforeprint` event handler for Window object (body element)
func OnBeforePrint(execute string) Attrib { return mx.Attribute("onbeforeprint", execute) }

// OnBeforeUnload `beforeunload` event handler for Window object (body element)
func OnBeforeUnload(execute string) Attrib { return mx.Attribute("onbeforeunload", execute) }

// OnBeforeToggle `beforetoggle` event handler (all HTML elements)
func OnBeforeToggle(execute string) Attrib { return mx.Attribute("onbeforetoggle", execute) }

// OnBlur `blur` event handler (all HTML elements)
func OnBlur(execute string) Attrib { return mx.Attribute("onblur", execute) }

// OnCancel `cancel` event handler (all HTML elements)
func OnCancel(execute string) Attrib { return mx.Attribute("oncancel", execute) }

// OnCanplay `canplay` event handler (all HTML elements)
func OnCanplay(execute string) Attrib { return mx.Attribute("oncanplay", execute) }

// OnCanPlayThrough `canplaythrough` event handler (all HTML elements)
func OnCanPlayThrough(execute string) Attrib { return mx.Attribute("oncanplaythrough", execute) }

// OnChange `change` event handler (all HTML elements)
func OnChange(execute string) Attrib { return mx.Attribute("onchange", execute) }

// OnClick `click` event handler (all HTML elements)
func OnClick(execute string) Attrib { return mx.Attribute("onclick", execute) }

// OnClose `close` event handler (all HTML elements)
func OnClose(execute string) Attrib { return mx.Attribute("onclose", execute) }

// OnContextLost `contextlost` event handler (all HTML elements)
func OnContextLost(execute string) Attrib { return mx.Attribute("oncontextlost", execute) }

// OnContextMenu `contextmenu` event handler (all HTML elements)
func OnContextMenu(execute string) Attrib { return mx.Attribute("oncontextmenu", execute) }

// OnContextRestored `contextrestored` event handler (all HTML elements)
func OnContextRestored(execute string) Attrib { return mx.Attribute("oncontextrestored", execute) }

// OnCopy `copy` event handler (all HTML elements)
func OnCopy(execute string) Attrib { return mx.Attribute("oncopy", execute) }

// OnCueChange `cuechange` event handler (all HTML elements)
func OnCueChange(execute string) Attrib { return mx.Attribute("oncuechange", execute) }

// OnCut `cut` event handler (all HTML elements)
func OnCut(execute string) Attrib { return mx.Attribute("oncut", execute) }

// OnDblClick `dblclick` event handler (all HTML elements)
func OnDblClick(execute string) Attrib { return mx.Attribute("ondblclick", execute) }

// OnDrag `drag` event handler (all HTML elements)
func OnDrag(execute string) Attrib { return mx.Attribute("ondrag", execute) }

// OnDragEnd `dragend` event handler (all HTML elements)
func OnDragEnd(execute string) Attrib { return mx.Attribute("ondragend", execute) }

// OnDragEnter `dragenter` event handler (all HTML elements)
func OnDragEnter(execute string) Attrib { return mx.Attribute("ondragenter", execute) }

// OnDragLeave `dragleave` event handler (all HTML elements)
func OnDragLeave(execute string) Attrib { return mx.Attribute("ondragleave", execute) }

// OnDragOver `dragover` event handler (all HTML elements)
func OnDragOver(execute string) Attrib { return mx.Attribute("ondragover", execute) }

// OnDragStart `dragstart` event handler (all HTML elements)
func OnDragStart(execute string) Attrib { return mx.Attribute("ondragstart", execute) }

// OnDrop `drop` event handler (all HTML elements)
func OnDrop(execute string) Attrib { return mx.Attribute("ondrop", execute) }

// OnDurationChange `durationchange` event handler (all HTML elements)
func OnDurationChange(execute string) Attrib { return mx.Attribute("ondurationchange", execute) }

// OnEmptied `emptied` event handler (all HTML elements)
func OnEmptied(execute string) Attrib { return mx.Attribute("onemptied", execute) }

// OnEnded `ended` event handler (all HTML elements)
func OnEnded(execute string) Attrib { return mx.Attribute("onended", execute) }

// OnError `error` event handler (all HTML elements)
func OnError(execute string) Attrib { return mx.Attribute("onerror", execute) }

// OnFocus `focus` event handler (all HTML elements)
func OnFocus(execute string) Attrib { return mx.Attribute("onfocus", execute) }

// OnFormData `formdata` event handler (all HTML elements)
func OnFormData(execute string) Attrib { return mx.Attribute("onformdata", execute) }

// OnHashChange `hashchange` event handler for Window object (body element)
func OnHashChange(execute string) Attrib { return mx.Attribute("onhashchange", execute) }

// OnInput `input` event handler (all HTML elements)
func OnInput(execute string) Attrib { return mx.Attribute("oninput", execute) }

// OnInvalid `invalid` event handler (all HTML elements)
func OnInvalid(execute string) Attrib { return mx.Attribute("oninvalid", execute) }

// OnKeyDown `keydown` event handler (all HTML elements)
func OnKeyDown(execute string) Attrib { return mx.Attribute("onkeydown", execute) }

// OnKeyPress `keypress` event handler (all HTML elements)
func OnKeyPress(execute string) Attrib { return mx.Attribute("onkeypress", execute) }

// OnKeyUp `keyup` event handler (all HTML elements)
func OnKeyUp(execute string) Attrib { return mx.Attribute("onkeyup", execute) }

// OnLanguageChange `languagechange` event handler for Window object (body element)
func OnLanguageChange(execute string) Attrib { return mx.Attribute("onlanguagechange", execute) }

// OnLoad `load` event handler (all HTML elements)
func OnLoad(execute string) Attrib { return mx.Attribute("onload", execute) }

// OnLoadedData `loadeddata` event handler (all HTML elements)
func OnLoadedData(execute string) Attrib { return mx.Attribute("onloadeddata", execute) }

// OnLoadedMetadata `loadedmetadata` event handler (all HTML elements)
func OnLoadedMetadata(execute string) Attrib { return mx.Attribute("onloadedmetadata", execute) }

// OnLoadStart `loadstart` event handler (all HTML elements)
func OnLoadStart(execute string) Attrib { return mx.Attribute("onloadstart", execute) }

// OnMessage `message` event handler for Window object (body element)
func OnMessage(execute string) Attrib { return mx.Attribute("onmessage", execute) }

// OnMessageError `messageerror` event handler for Window object (body element)
func OnMessageError(execute string) Attrib { return mx.Attribute("onmessageerror", execute) }

// OnMouseDown `mousedown` event handler (all HTML elements)
func OnMouseDown(execute string) Attrib { return mx.Attribute("onmousedown", execute) }

// OnMouseEnter `mouseenter` event handler (all HTML elements)
func OnMouseEnter(execute string) Attrib { return mx.Attribute("onmouseenter", execute) }

// OnMouseLeave `mouseleave` event handler (all HTML elements)
func OnMouseLeave(execute string) Attrib { return mx.Attribute("onmouseleave", execute) }

// OnMouseMove `mousemove` event handler (all HTML elements)
func OnMouseMove(execute string) Attrib { return mx.Attribute("onmousemove", execute) }

// OnMouseOut `mouseout` event handler (all HTML elements)
func OnMouseOut(execute string) Attrib { return mx.Attribute("onmouseout", execute) }

// OnMouseOver `mouseover` event handler (all HTML elements)
func OnMouseOver(execute string) Attrib { return mx.Attribute("onmouseover", execute) }

// OnMouseUp `mouseup` event handler (all HTML elements)
func OnMouseUp(execute string) Attrib { return mx.Attribute("onmouseup", execute) }

// OnOffline `offline` event handler for Window object (body element)
func OnOffline(execute string) Attrib { return mx.Attribute("onoffline", execute) }

// OnOnline `online` event handler for Window object (body element)
func OnOnline(execute string) Attrib { return mx.Attribute("ononline", execute) }

// OnPageHide `pagehide` event handler for Window object (body element)
func OnPageHide(execute string) Attrib { return mx.Attribute("onpagehide", execute) }

// OnPageReveal `pagereveal` event handler for Window object (body element)
func OnPageReveal(execute string) Attrib { return mx.Attribute("onpagereveal", execute) }

// OnPageShow `pageshow` event handler for Window object (body element)
func OnPageShow(execute string) Attrib { return mx.Attribute("onpageshow", execute) }

// OnPageSwap `pageswap` event handler for Window object (body element)
func OnPageSwap(execute string) Attrib { return mx.Attribute("onpageswap", execute) }

// OnPaste `paste` event handler (all HTML elements)
func OnPaste(execute string) Attrib { return mx.Attribute("onpaste", execute) }

// OnPause `pause` event handler (all HTML elements)
func OnPause(execute string) Attrib { return mx.Attribute("onpause", execute) }

// OnPlay `play` event handler (all HTML elements)
func OnPlay(execute string) Attrib { return mx.Attribute("onplay", execute) }

// OnPlaying `playing` event handler (all HTML elements)
func OnPlaying(execute string) Attrib { return mx.Attribute("onplaying", execute) }

// OnPopState `popstate` event handler for Window object (body element)
func OnPopState(execute string) Attrib { return mx.Attribute("onpopstate", execute) }

// OnProgress `progress` event handler (all HTML elements)
func OnProgress(execute string) Attrib { return mx.Attribute("onprogress", execute) }

// OnRateChange `ratechange` event handler (all HTML elements)
func OnRateChange(execute string) Attrib { return mx.Attribute("onratechange", execute) }

// OnReset `reset` event handler (all HTML elements)
func OnReset(execute string) Attrib { return mx.Attribute("onreset", execute) }

// OnResize `resize` event handler (all HTML elements)
func OnResize(execute string) Attrib { return mx.Attribute("onresize", execute) }

// OnRejectionHandled `rejectionhandled` event handler for Window object (body element)
func OnRejectionHandled(execute string) Attrib { return mx.Attribute("onrejectionhandled", execute) }

// OnScroll `scroll` event handler (all HTML elements)
func OnScroll(execute string) Attrib { return mx.Attribute("onscroll", execute) }

// OnScrollEnd `scrollend` event handler (all HTML elements)
func OnScrollEnd(execute string) Attrib { return mx.Attribute("onscrollend", execute) }

// OnSecurityPolicyViolation `securitypolicyviolation` event handler (all HTML elements)
func OnSecurityPolicyViolation(execute string) Attrib {
	return mx.Attribute("onsecuritypolicyviolation", execute)
}

// OnSeeked `seeked` event handler (all HTML elements)
func OnSeeked(execute string) Attrib { return mx.Attribute("onseeked", execute) }

// OnSeeking `seeking` event handler (all HTML elements)
func OnSeeking(execute string) Attrib { return mx.Attribute("onseeking", execute) }

// OnSelect `select` event handler (all HTML elements)
func OnSelect(execute string) Attrib { return mx.Attribute("onselect", execute) }

// OnSlotChange `slotchange` event handler (all HTML elements)
func OnSlotChange(execute string) Attrib { return mx.Attribute("onslotchange", execute) }

// OnStalled `stalled` event handler (all HTML elements)
func OnStalled(execute string) Attrib { return mx.Attribute("onstalled", execute) }

// OnStorage `storage` event handler for Window object (body element)
func OnStorage(execute string) Attrib { return mx.Attribute("onstorage", execute) }

// OnSubmit `submit` event handler (all HTML elements)
func OnSubmit(execute string) Attrib { return mx.Attribute("onsubmit", execute) }

// OnSuspend `suspend` event handler (all HTML elements)
func OnSuspend(execute string) Attrib { return mx.Attribute("onsuspend", execute) }

// OnTimeUpdate `timeupdate` event handler (all HTML elements)
func OnTimeUpdate(execute string) Attrib { return mx.Attribute("ontimeupdate", execute) }

// OnToggle `toggle` event handler (all HTML elements)
func OnToggle(execute string) Attrib { return mx.Attribute("ontoggle", execute) }

// OnUnhandledRejection `unhandledrejection` event handler for Window object (body element)
func OnUnhandledRejection(execute string) Attrib {
	return mx.Attribute("onunhandledrejection", execute)
}

// OnUnload `unload` event handler for Window object (body element)
func OnUnload(execute string) Attrib { return mx.Attribute("onunload", execute) }

// OnVolumeChange `volumechange` event handler (all HTML elements)
func OnVolumeChange(execute string) Attrib { return mx.Attribute("onvolumechange", execute) }

// OnWaiting `waiting` event handler (all HTML elements)
func OnWaiting(execute string) Attrib { return mx.Attribute("onwaiting", execute) }

// OnWheel `wheel` event handler (all HTML elements)
func OnWheel(execute string) Attrib { return mx.Attribute("onwheel", execute) }
