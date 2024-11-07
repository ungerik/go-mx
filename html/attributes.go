package html

import (
	"strings"

	"github.com/ungerik/go-mx"
)

type Attrib = mx.Attrib

// See https://github.com/jozo/all-html-elements-and-attributes

func Accept(contentTypes ...string) Attrib {
	return Attrib{Name: "accept", Value: strings.Join(contentTypes, ",")}
}
func AcceptCharset(charsets ...string) Attrib {
	return Attrib{Name: "accept-charset", Value: strings.Join(charsets, " ")}
}
func AccessKey(value string) Attrib    { return Attrib{Name: "accesskey", Value: value} }
func Action(value string) Attrib       { return Attrib{Name: "action", Value: value} }
func Align(value string) Attrib        { return Attrib{Name: "align", Value: value} }
func Allow(value string) Attrib        { return Attrib{Name: "allow", Value: value} }
func Alt(value string) Attrib          { return Attrib{Name: "alt", Value: value} }
func As(value string) Attrib           { return Attrib{Name: "as", Value: value} }
func Async(value string) Attrib        { return Attrib{Name: "async", Value: value} }
func AutoCapitalizeNone() Attrib       { return Attrib{Name: "autocapitalize", Value: "none"} }
func AutoCapitalizeSentences() Attrib  { return Attrib{Name: "autocapitalize", Value: "sentences"} }
func AutoCapitalizeWords() Attrib      { return Attrib{Name: "autocapitalize", Value: "words"} }
func AutoCapitalizeCharacters() Attrib { return Attrib{Name: "autocapitalize", Value: "characters"} }
func AutoComplete(tokens ...string) Attrib {
	if len(tokens) == 0 {
		return AutoCompleteOn()
	}
	return Attrib{Name: "autocomplete", Value: strings.Join(tokens, " ")}
}
func AutoCompleteOn() Attrib              { return Attrib{Name: "autocomplete", Value: "on"} }
func AutoCompleteOff() Attrib             { return Attrib{Name: "autocomplete", Value: "off"} }
func AutoPlay(value string) Attrib        { return Attrib{Name: "autoplay", Value: value} }
func Background(value string) Attrib      { return Attrib{Name: "background", Value: value} }
func BGColor(value string) Attrib         { return Attrib{Name: "bgcolor", Value: value} }
func Border(value string) Attrib          { return Attrib{Name: "border", Value: value} }
func Capture(value string) Attrib         { return Attrib{Name: "capture", Value: value} }
func Charset(value string) Attrib         { return Attrib{Name: "charset", Value: value} }
func Checked(value string) Attrib         { return Attrib{Name: "checked", Value: value} }
func CiteAttr(value string) Attrib        { return Attrib{Name: "cite", Value: value} }
func Class(value string) Attrib           { return Attrib{Name: "class", Value: value} }
func Color(value string) Attrib           { return Attrib{Name: "color", Value: value} }
func Cols(value string) Attrib            { return Attrib{Name: "cols", Value: value} }
func ColSpan(value string) Attrib         { return Attrib{Name: "colspan", Value: value} }
func ContentAttr(value string) Attrib     { return Attrib{Name: "content", Value: value} }
func ContentEditable(value string) Attrib { return Attrib{Name: "contenteditable", Value: value} }
func Controls(value string) Attrib        { return Attrib{Name: "controls", Value: value} }
func Coords(value string) Attrib          { return Attrib{Name: "coords", Value: value} }
func CrossOrigin(value string) Attrib     { return Attrib{Name: "crossorigin", Value: value} }
func CSP(value string) Attrib             { return Attrib{Name: "csp", Value: value} }
func DataAttr(name, value string) Attrib  { return Attrib{Name: "data-" + name, Value: value} }
func Datetime(value string) Attrib        { return Attrib{Name: "datetime", Value: value} }
func Decoding(value string) Attrib        { return Attrib{Name: "decoding", Value: value} }
func Default(value string) Attrib         { return Attrib{Name: "default", Value: value} }
func Defer(value string) Attrib           { return Attrib{Name: "defer", Value: value} }
func Dir(value string) Attrib             { return Attrib{Name: "dir", Value: value} }
func DirName(value string) Attrib         { return Attrib{Name: "dirname", Value: value} }
func Disabled(value string) Attrib        { return Attrib{Name: "disabled", Value: value} }
func Download(value string) Attrib        { return Attrib{Name: "download", Value: value} }
func Draggable(value string) Attrib       { return Attrib{Name: "draggable", Value: value} }
func EncType(value string) Attrib         { return Attrib{Name: "enctype", Value: value} }
func EnterKeyHint(value string) Attrib    { return Attrib{Name: "enterkeyhint", Value: value} }
func For(value string) Attrib             { return Attrib{Name: "for", Value: value} }
func FormAttr(value string) Attrib        { return Attrib{Name: "form", Value: value} }
func FormAction(value string) Attrib      { return Attrib{Name: "formaction", Value: value} }
func FormEncType(value string) Attrib     { return Attrib{Name: "formenctype", Value: value} }
func FormMethodGET() Attrib               { return Attrib{Name: "formmethod", Value: "get"} }
func FormMethodPOST() Attrib              { return Attrib{Name: "formmethod", Value: "post"} }
func FormMethodDialog() Attrib            { return Attrib{Name: "formmethod", Value: "dialog"} }
func FormNoValidate() Attrib              { return Attrib{Name: "formnovalidate", Value: "formnovalidate"} }
func FormTarget(value string) Attrib      { return Attrib{Name: "formtarget", Value: value} }
func Headers(value string) Attrib         { return Attrib{Name: "headers", Value: value} }
func Height(value string) Attrib          { return Attrib{Name: "height", Value: value} }
func Hidden(value string) Attrib          { return Attrib{Name: "hidden", Value: value} }
func High(value string) Attrib            { return Attrib{Name: "high", Value: value} }
func HRef(url string) Attrib              { return Attrib{Name: "href", Value: url} }
func HRefLang(value string) Attrib        { return Attrib{Name: "hreflang", Value: value} }
func HTTPEquiv(value string) Attrib       { return Attrib{Name: "http-equiv", Value: value} }
func ID(value string) Attrib              { return Attrib{Name: "id", Value: value} }
func Inputmode(value string) Attrib       { return Attrib{Name: "inputmode", Value: value} }
func Integrity(value string) Attrib       { return Attrib{Name: "integrity", Value: value} }
func IntrinsicSize(value string) Attrib   { return Attrib{Name: "intrinsicsize", Value: value} }
func IsMap(value string) Attrib           { return Attrib{Name: "ismap", Value: value} }
func ItemProp(value string) Attrib        { return Attrib{Name: "itemprop", Value: value} }
func Kind(value string) Attrib            { return Attrib{Name: "kind", Value: value} }
func LabelAttr(value string) Attrib       { return Attrib{Name: "label", Value: value} }
func Lang(value string) Attrib            { return Attrib{Name: "lang", Value: value} }
func Language(value string) Attrib        { return Attrib{Name: "language", Value: value} }
func List(value string) Attrib            { return Attrib{Name: "list", Value: value} }
func Loading(value string) Attrib         { return Attrib{Name: "loading", Value: value} }
func Loop(value string) Attrib            { return Attrib{Name: "loop", Value: value} }
func Low(value string) Attrib             { return Attrib{Name: "low", Value: value} }
func Max(value string) Attrib             { return Attrib{Name: "max", Value: value} }
func MaxLength(value string) Attrib       { return Attrib{Name: "maxlength", Value: value} }
func Media(value string) Attrib           { return Attrib{Name: "media", Value: value} }
func MethodGET() Attrib                   { return Attrib{Name: "method", Value: "get"} }
func MethodPOST() Attrib                  { return Attrib{Name: "method", Value: "post"} }
func MethodDialog() Attrib                { return Attrib{Name: "method", Value: "dialog"} }
func Min(value string) Attrib             { return Attrib{Name: "min", Value: value} }
func MinLength(value string) Attrib       { return Attrib{Name: "minlength", Value: value} }
func Multiple(value string) Attrib        { return Attrib{Name: "multiple", Value: value} }
func Muted(value string) Attrib           { return Attrib{Name: "muted", Value: value} }
func Name(value string) Attrib            { return Attrib{Name: "name", Value: value} }
func NoValidate() Attrib                  { return Attrib{Name: "novalidate", Value: "novalidate"} }
func Open(value string) Attrib            { return Attrib{Name: "open", Value: value} }
func Optimum(value string) Attrib         { return Attrib{Name: "optimum", Value: value} }
func Pattern(value string) Attrib         { return Attrib{Name: "pattern", Value: value} }
func Ping(value string) Attrib            { return Attrib{Name: "ping", Value: value} }
func Placeholder(value string) Attrib     { return Attrib{Name: "placeholder", Value: value} }
func PlaysInline(value string) Attrib     { return Attrib{Name: "playsinline", Value: value} }
func Poster(value string) Attrib          { return Attrib{Name: "poster", Value: value} }
func Preload(value string) Attrib         { return Attrib{Name: "preload", Value: value} }
func Readonly(value string) Attrib        { return Attrib{Name: "readonly", Value: value} }
func ReferrerPolicyNoReferrer() Attrib    { return Attrib{Name: "referrerpolicy", Value: "no-referrer"} }
func ReferrerPolicyNoReferrerWhenDowngrade() Attrib {
	return Attrib{Name: "referrerpolicy", Value: "no-referrer-when-downgrade"}
}
func ReferrerPolicyOrigin() Attrib { return Attrib{Name: "referrerpolicy", Value: "origin"} }
func ReferrerPolicyOriginWhenCrossOrigin() Attrib {
	return Attrib{Name: "referrerpolicy", Value: "origin-when-cross-origin"}
}
func ReferrerPolicySameOrigin() Attrib { return Attrib{Name: "referrerpolicy", Value: "same-origin"} }
func ReferrerPolicyStrictOrigin() Attrib {
	return Attrib{Name: "referrerpolicy", Value: "strict-origin"}
}
func ReferrerPolicyStrictOriginWhenCrossOrigin() Attrib {
	return Attrib{Name: "referrerpolicy", Value: "strict-origin-when-cross-origin"}
}
func ReferrerPolicyUnsafeUrl() Attrib { return Attrib{Name: "referrerpolicy", Value: "unsafe-url"} }
func Rel(keywords ...string) Attrib   { return Attrib{Name: "rel", Value: strings.Join(keywords, " ")} }
func Required(value string) Attrib    { return Attrib{Name: "required", Value: value} }
func Reversed(value string) Attrib    { return Attrib{Name: "reversed", Value: value} }
func Role(value string) Attrib        { return Attrib{Name: "role", Value: value} }
func Rows(value string) Attrib        { return Attrib{Name: "rows", Value: value} }
func RowSpan(value string) Attrib     { return Attrib{Name: "rowspan", Value: value} }
func Sandbox(value string) Attrib     { return Attrib{Name: "sandbox", Value: value} }
func Scope(value string) Attrib       { return Attrib{Name: "scope", Value: value} }
func Scoped(value string) Attrib      { return Attrib{Name: "scoped", Value: value} }
func Selected(value string) Attrib    { return Attrib{Name: "selected", Value: value} }
func Shape(value string) Attrib       { return Attrib{Name: "shape", Value: value} }
func Size(value string) Attrib        { return Attrib{Name: "size", Value: value} }
func Sizes(value string) Attrib       { return Attrib{Name: "sizes", Value: value} }
func SlotAttr(value string) Attrib    { return Attrib{Name: "slot", Value: value} }
func SpanAttr(value string) Attrib    { return Attrib{Name: "span", Value: value} }
func Spellcheck(value string) Attrib  { return Attrib{Name: "spellcheck", Value: value} }
func Src(value string) Attrib         { return Attrib{Name: "src", Value: value} }
func SrcDoc(value string) Attrib      { return Attrib{Name: "srcdoc", Value: value} }
func SrcLang(value string) Attrib     { return Attrib{Name: "srclang", Value: value} }
func SrcSet(value string) Attrib      { return Attrib{Name: "srcset", Value: value} }
func Start(value string) Attrib       { return Attrib{Name: "start", Value: value} }
func Step(value string) Attrib        { return Attrib{Name: "step", Value: value} }
func Style(value string) Attrib       { return Attrib{Name: "style", Value: value} }
func TabIndex(value string) Attrib    { return Attrib{Name: "tabindex", Value: value} }
func Target(value string) Attrib      { return Attrib{Name: "target", Value: value} }
func TargetSelf() Attrib              { return Attrib{Name: "target", Value: "_self"} }
func TargetBlank() Attrib             { return Attrib{Name: "target", Value: "_blank"} }
func TargetParent() Attrib            { return Attrib{Name: "target", Value: "_parent"} }
func TargetTop() Attrib               { return Attrib{Name: "target", Value: "_top"} }
func TargetUnfencedTop() Attrib       { return Attrib{Name: "target", Value: "_unfencedTop"} }
func Title(value string) Attrib       { return Attrib{Name: "title", Value: value} }
func Translate(value string) Attrib   { return Attrib{Name: "translate", Value: value} }
func Type(value string) Attrib        { return Attrib{Name: "type", Value: value} }
func Usemap(value string) Attrib      { return Attrib{Name: "usemap", Value: value} }
func Value(value string) Attrib       { return Attrib{Name: "value", Value: value} }
func Width(value string) Attrib       { return Attrib{Name: "width", Value: value} }
func Wrap(value string) Attrib        { return Attrib{Name: "wrap", Value: value} }
