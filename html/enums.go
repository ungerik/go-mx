//go:generate go -C ../tools tool go-enum ../html/$GOFILE

// This file defines the HTML attributes whose value is a strict, closed set of
// keyword tokens as enum types. Each type is a string that implements mx.Attrib,
// so the typed constants are used directly as element attributes, e.g.
// html.Div(html.DirRTL). A conversion such as Dir("rtl") also works for dynamic
// values. AttribValue returns the value together with the result of the
// generated Validate method, so rendering an element that holds an invalid enum
// value fails with a descriptive error instead of emitting an invalid keyword.
//
// The go:generate directive runs the go-enum tool pinned in the nested tools
// module (kept out of the shipped go-mx dependency tree). Run it with:
//
//	go generate ./html/...
//
// go-enum appends the Valid, Validate, Enums, EnumStrings and String methods for
// every //#enum type; the hand-written AttribName/AttribValue methods are left
// untouched.

package html

import (
	"context"
	"fmt"

	"github.com/ungerik/go-mx"
)

// AutoCapitalize is the HTML autocapitalize global attribute.
type AutoCapitalize string //#enum

const (
	// AutoCapitalizeOff disables automatic capitalization (no autocapitalization).
	AutoCapitalizeOff AutoCapitalize = "off"
	// AutoCapitalizeNone disables automatic capitalization (synonym of "off").
	AutoCapitalizeNone AutoCapitalize = "none"
	// AutoCapitalizeOn enables default platform autocapitalization (synonym of "sentences").
	AutoCapitalizeOn AutoCapitalize = "on"
	// AutoCapitalizeSentences autocapitalizes the first letter of each sentence.
	AutoCapitalizeSentences AutoCapitalize = "sentences"
	// AutoCapitalizeWords autocapitalizes the first letter of each word.
	AutoCapitalizeWords AutoCapitalize = "words"
	// AutoCapitalizeCharacters autocapitalizes every character.
	AutoCapitalizeCharacters AutoCapitalize = "characters"
)

// Valid indicates if v is any of the valid values for AutoCapitalize
func (v AutoCapitalize) Valid() bool {
	switch v {
	case
		AutoCapitalizeOff,
		AutoCapitalizeNone,
		AutoCapitalizeOn,
		AutoCapitalizeSentences,
		AutoCapitalizeWords,
		AutoCapitalizeCharacters:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for AutoCapitalize
func (v AutoCapitalize) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type html.AutoCapitalize", v)
	}
	return nil
}

// Enums returns all valid values for AutoCapitalize
func (AutoCapitalize) Enums() []AutoCapitalize {
	return []AutoCapitalize{
		AutoCapitalizeOff,
		AutoCapitalizeNone,
		AutoCapitalizeOn,
		AutoCapitalizeSentences,
		AutoCapitalizeWords,
		AutoCapitalizeCharacters,
	}
}

// EnumStrings returns all valid values for AutoCapitalize as strings
func (AutoCapitalize) EnumStrings() []string {
	return []string{
		"off",
		"none",
		"on",
		"sentences",
		"words",
		"characters",
	}
}

// String implements the fmt.Stringer interface for AutoCapitalize
func (v AutoCapitalize) String() string {
	return string(v)
}

// AttribName returns the "autocapitalize" attribute name.
func (v AutoCapitalize) AttribName() string { return "autocapitalize" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v AutoCapitalize) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// AutoCorrect is the HTML autocorrect attribute.
type AutoCorrect string //#enum

const (
	// AutoCorrectOn enables automatic spelling correction and text substitutions.
	AutoCorrectOn AutoCorrect = "on"
	// AutoCorrectOff disables automatic spelling correction and text substitutions.
	AutoCorrectOff AutoCorrect = "off"
)

// Valid indicates if v is any of the valid values for AutoCorrect
func (v AutoCorrect) Valid() bool {
	switch v {
	case
		AutoCorrectOn,
		AutoCorrectOff:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for AutoCorrect
func (v AutoCorrect) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type html.AutoCorrect", v)
	}
	return nil
}

// Enums returns all valid values for AutoCorrect
func (AutoCorrect) Enums() []AutoCorrect {
	return []AutoCorrect{
		AutoCorrectOn,
		AutoCorrectOff,
	}
}

// EnumStrings returns all valid values for AutoCorrect as strings
func (AutoCorrect) EnumStrings() []string {
	return []string{
		"on",
		"off",
	}
}

// String implements the fmt.Stringer interface for AutoCorrect
func (v AutoCorrect) String() string {
	return string(v)
}

// AttribName returns the "autocorrect" attribute name.
func (v AutoCorrect) AttribName() string { return "autocorrect" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v AutoCorrect) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// ContentEditable is the HTML contenteditable global attribute.
type ContentEditable string //#enum

const (
	// ContentEditableTrue makes the element editable, including rich text formatting.
	ContentEditableTrue ContentEditable = "true"
	// ContentEditableFalse makes the element not editable.
	ContentEditableFalse ContentEditable = "false"
	// ContentEditablePlaintextOnly makes the element editable as raw text without rich formatting.
	ContentEditablePlaintextOnly ContentEditable = "plaintext-only"
)

// Valid indicates if v is any of the valid values for ContentEditable
func (v ContentEditable) Valid() bool {
	switch v {
	case
		ContentEditableTrue,
		ContentEditableFalse,
		ContentEditablePlaintextOnly:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for ContentEditable
func (v ContentEditable) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type html.ContentEditable", v)
	}
	return nil
}

// Enums returns all valid values for ContentEditable
func (ContentEditable) Enums() []ContentEditable {
	return []ContentEditable{
		ContentEditableTrue,
		ContentEditableFalse,
		ContentEditablePlaintextOnly,
	}
}

// EnumStrings returns all valid values for ContentEditable as strings
func (ContentEditable) EnumStrings() []string {
	return []string{
		"true",
		"false",
		"plaintext-only",
	}
}

// String implements the fmt.Stringer interface for ContentEditable
func (v ContentEditable) String() string {
	return string(v)
}

// AttribName returns the "contenteditable" attribute name.
func (v ContentEditable) AttribName() string { return "contenteditable" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v ContentEditable) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// CrossOrigin is the HTML crossorigin attribute.
type CrossOrigin string //#enum

const (
	// CrossOriginAnonymous performs a CORS request without sending credentials (cookies, TLS client certs, or HTTP auth).
	CrossOriginAnonymous CrossOrigin = "anonymous"
	// CrossOriginUseCredentials performs a CORS request that sends credentials (cookies, TLS client certs, or HTTP auth).
	CrossOriginUseCredentials CrossOrigin = "use-credentials"
)

// Valid indicates if v is any of the valid values for CrossOrigin
func (v CrossOrigin) Valid() bool {
	switch v {
	case
		CrossOriginAnonymous,
		CrossOriginUseCredentials:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for CrossOrigin
func (v CrossOrigin) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type html.CrossOrigin", v)
	}
	return nil
}

// Enums returns all valid values for CrossOrigin
func (CrossOrigin) Enums() []CrossOrigin {
	return []CrossOrigin{
		CrossOriginAnonymous,
		CrossOriginUseCredentials,
	}
}

// EnumStrings returns all valid values for CrossOrigin as strings
func (CrossOrigin) EnumStrings() []string {
	return []string{
		"anonymous",
		"use-credentials",
	}
}

// String implements the fmt.Stringer interface for CrossOrigin
func (v CrossOrigin) String() string {
	return string(v)
}

// AttribName returns the "crossorigin" attribute name.
func (v CrossOrigin) AttribName() string { return "crossorigin" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v CrossOrigin) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Decoding is the HTML decoding attribute of the img element.
type Decoding string //#enum

const (
	// DecodingAuto lets the browser decide the best image decoding strategy (the default).
	DecodingAuto Decoding = "auto"
	// DecodingAsync decodes the image asynchronously to avoid blocking presentation of other content.
	DecodingAsync Decoding = "async"
	// DecodingSync decodes the image synchronously for atomic presentation with other content.
	DecodingSync Decoding = "sync"
)

// Valid indicates if v is any of the valid values for Decoding
func (v Decoding) Valid() bool {
	switch v {
	case
		DecodingAuto,
		DecodingAsync,
		DecodingSync:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for Decoding
func (v Decoding) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type html.Decoding", v)
	}
	return nil
}

// Enums returns all valid values for Decoding
func (Decoding) Enums() []Decoding {
	return []Decoding{
		DecodingAuto,
		DecodingAsync,
		DecodingSync,
	}
}

// EnumStrings returns all valid values for Decoding as strings
func (Decoding) EnumStrings() []string {
	return []string{
		"auto",
		"async",
		"sync",
	}
}

// String implements the fmt.Stringer interface for Decoding
func (v Decoding) String() string {
	return string(v)
}

// AttribName returns the "decoding" attribute name.
func (v Decoding) AttribName() string { return "decoding" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v Decoding) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Dir is the HTML dir global attribute (text directionality).
type Dir string //#enum

const (
	// DirLTR sets the text direction to left-to-right.
	DirLTR Dir = "ltr"
	// DirRTL sets the text direction to right-to-left.
	DirRTL Dir = "rtl"
	// DirAuto lets the user agent determine the direction from the element's content.
	DirAuto Dir = "auto"
)

// Valid indicates if v is any of the valid values for Dir
func (v Dir) Valid() bool {
	switch v {
	case
		DirLTR,
		DirRTL,
		DirAuto:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for Dir
func (v Dir) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type html.Dir", v)
	}
	return nil
}

// Enums returns all valid values for Dir
func (Dir) Enums() []Dir {
	return []Dir{
		DirLTR,
		DirRTL,
		DirAuto,
	}
}

// EnumStrings returns all valid values for Dir as strings
func (Dir) EnumStrings() []string {
	return []string{
		"ltr",
		"rtl",
		"auto",
	}
}

// String implements the fmt.Stringer interface for Dir
func (v Dir) String() string {
	return string(v)
}

// AttribName returns the "dir" attribute name.
func (v Dir) AttribName() string { return "dir" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v Dir) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// EncType is the HTML enctype attribute of the form element.
type EncType string //#enum

const (
	// EncTypeFormURLEncoded URL-encodes the form data (the default form encoding).
	EncTypeFormURLEncoded EncType = "application/x-www-form-urlencoded"
	// EncTypeMultipartFormData sends the form as multipart data, required for file uploads.
	EncTypeMultipartFormData EncType = "multipart/form-data"
	// EncTypeTextPlain sends the form data as plain text without encoding (debugging only).
	EncTypeTextPlain EncType = "text/plain"
)

// Valid indicates if v is any of the valid values for EncType
func (v EncType) Valid() bool {
	switch v {
	case
		EncTypeFormURLEncoded,
		EncTypeMultipartFormData,
		EncTypeTextPlain:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for EncType
func (v EncType) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type html.EncType", v)
	}
	return nil
}

// Enums returns all valid values for EncType
func (EncType) Enums() []EncType {
	return []EncType{
		EncTypeFormURLEncoded,
		EncTypeMultipartFormData,
		EncTypeTextPlain,
	}
}

// EnumStrings returns all valid values for EncType as strings
func (EncType) EnumStrings() []string {
	return []string{
		"application/x-www-form-urlencoded",
		"multipart/form-data",
		"text/plain",
	}
}

// String implements the fmt.Stringer interface for EncType
func (v EncType) String() string {
	return string(v)
}

// AttribName returns the "enctype" attribute name.
func (v EncType) AttribName() string { return "enctype" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v EncType) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// EnterKeyHint is the HTML enterkeyhint global attribute.
type EnterKeyHint string //#enum

const (
	// EnterKeyHintEnter hints at inserting a new line (typically a "return" key).
	EnterKeyHintEnter EnterKeyHint = "enter"
	// EnterKeyHintDone hints that there is nothing more to input and the input method editor closes.
	EnterKeyHintDone EnterKeyHint = "done"
	// EnterKeyHintGo hints at taking the user to the target of the typed text.
	EnterKeyHintGo EnterKeyHint = "go"
	// EnterKeyHintNext hints at moving to the next field that accepts text.
	EnterKeyHintNext EnterKeyHint = "next"
	// EnterKeyHintPrevious hints at moving to the previous field that accepts text.
	EnterKeyHintPrevious EnterKeyHint = "previous"
	// EnterKeyHintSearch hints at taking the user to the results of searching the typed text.
	EnterKeyHintSearch EnterKeyHint = "search"
	// EnterKeyHintSend hints at delivering the typed text to its target.
	EnterKeyHintSend EnterKeyHint = "send"
)

// Valid indicates if v is any of the valid values for EnterKeyHint
func (v EnterKeyHint) Valid() bool {
	switch v {
	case
		EnterKeyHintEnter,
		EnterKeyHintDone,
		EnterKeyHintGo,
		EnterKeyHintNext,
		EnterKeyHintPrevious,
		EnterKeyHintSearch,
		EnterKeyHintSend:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for EnterKeyHint
func (v EnterKeyHint) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type html.EnterKeyHint", v)
	}
	return nil
}

// Enums returns all valid values for EnterKeyHint
func (EnterKeyHint) Enums() []EnterKeyHint {
	return []EnterKeyHint{
		EnterKeyHintEnter,
		EnterKeyHintDone,
		EnterKeyHintGo,
		EnterKeyHintNext,
		EnterKeyHintPrevious,
		EnterKeyHintSearch,
		EnterKeyHintSend,
	}
}

// EnumStrings returns all valid values for EnterKeyHint as strings
func (EnterKeyHint) EnumStrings() []string {
	return []string{
		"enter",
		"done",
		"go",
		"next",
		"previous",
		"search",
		"send",
	}
}

// String implements the fmt.Stringer interface for EnterKeyHint
func (v EnterKeyHint) String() string {
	return string(v)
}

// AttribName returns the "enterkeyhint" attribute name.
func (v EnterKeyHint) AttribName() string { return "enterkeyhint" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v EnterKeyHint) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// FetchPriority is the HTML fetchpriority attribute.
type FetchPriority string //#enum

const (
	// FetchPriorityAuto lets the browser decide the fetch priority (the default).
	FetchPriorityAuto FetchPriority = "auto"
	// FetchPriorityHigh fetches the resource at high priority relative to other resources of the same type.
	FetchPriorityHigh FetchPriority = "high"
	// FetchPriorityLow fetches the resource at low priority relative to other resources of the same type.
	FetchPriorityLow FetchPriority = "low"
)

// Valid indicates if v is any of the valid values for FetchPriority
func (v FetchPriority) Valid() bool {
	switch v {
	case
		FetchPriorityAuto,
		FetchPriorityHigh,
		FetchPriorityLow:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for FetchPriority
func (v FetchPriority) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type html.FetchPriority", v)
	}
	return nil
}

// Enums returns all valid values for FetchPriority
func (FetchPriority) Enums() []FetchPriority {
	return []FetchPriority{
		FetchPriorityAuto,
		FetchPriorityHigh,
		FetchPriorityLow,
	}
}

// EnumStrings returns all valid values for FetchPriority as strings
func (FetchPriority) EnumStrings() []string {
	return []string{
		"auto",
		"high",
		"low",
	}
}

// String implements the fmt.Stringer interface for FetchPriority
func (v FetchPriority) String() string {
	return string(v)
}

// AttribName returns the "fetchpriority" attribute name.
func (v FetchPriority) AttribName() string { return "fetchpriority" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v FetchPriority) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// FormEncType is the HTML formenctype attribute of submit buttons.
type FormEncType string //#enum

const (
	// FormEncTypeFormURLEncoded URL-encodes the form data, overriding the form's enctype (the default encoding).
	FormEncTypeFormURLEncoded FormEncType = "application/x-www-form-urlencoded"
	// FormEncTypeMultipartFormData sends the form as multipart data, overriding the form's enctype (required for file uploads).
	FormEncTypeMultipartFormData FormEncType = "multipart/form-data"
	// FormEncTypeTextPlain sends the form data as plain text, overriding the form's enctype (debugging only).
	FormEncTypeTextPlain FormEncType = "text/plain"
)

// Valid indicates if v is any of the valid values for FormEncType
func (v FormEncType) Valid() bool {
	switch v {
	case
		FormEncTypeFormURLEncoded,
		FormEncTypeMultipartFormData,
		FormEncTypeTextPlain:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for FormEncType
func (v FormEncType) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type html.FormEncType", v)
	}
	return nil
}

// Enums returns all valid values for FormEncType
func (FormEncType) Enums() []FormEncType {
	return []FormEncType{
		FormEncTypeFormURLEncoded,
		FormEncTypeMultipartFormData,
		FormEncTypeTextPlain,
	}
}

// EnumStrings returns all valid values for FormEncType as strings
func (FormEncType) EnumStrings() []string {
	return []string{
		"application/x-www-form-urlencoded",
		"multipart/form-data",
		"text/plain",
	}
}

// String implements the fmt.Stringer interface for FormEncType
func (v FormEncType) String() string {
	return string(v)
}

// AttribName returns the "formenctype" attribute name.
func (v FormEncType) AttribName() string { return "formenctype" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v FormEncType) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// FormMethod is the HTML formmethod attribute of submit buttons.
type FormMethod string //#enum

const (
	// FormMethodGET submits the form with the HTTP GET method, overriding the form's method.
	FormMethodGET FormMethod = "get"
	// FormMethodPOST submits the form with the HTTP POST method, overriding the form's method.
	FormMethodPOST FormMethod = "post"
	// FormMethodDialog closes the enclosing dialog and submits without sending data, overriding the form's method.
	FormMethodDialog FormMethod = "dialog"
)

// Valid indicates if v is any of the valid values for FormMethod
func (v FormMethod) Valid() bool {
	switch v {
	case
		FormMethodGET,
		FormMethodPOST,
		FormMethodDialog:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for FormMethod
func (v FormMethod) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type html.FormMethod", v)
	}
	return nil
}

// Enums returns all valid values for FormMethod
func (FormMethod) Enums() []FormMethod {
	return []FormMethod{
		FormMethodGET,
		FormMethodPOST,
		FormMethodDialog,
	}
}

// EnumStrings returns all valid values for FormMethod as strings
func (FormMethod) EnumStrings() []string {
	return []string{
		"get",
		"post",
		"dialog",
	}
}

// String implements the fmt.Stringer interface for FormMethod
func (v FormMethod) String() string {
	return string(v)
}

// AttribName returns the "formmethod" attribute name.
func (v FormMethod) AttribName() string { return "formmethod" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v FormMethod) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// HTTPEquiv is the HTML http-equiv attribute of the meta element.
type HTTPEquiv string //#enum

const (
	// HTTPEquivContentType declares the document's character encoding (the content attribute must be "text/html; charset=utf-8").
	HTTPEquivContentType HTTPEquiv = "content-type"
	// HTTPEquivDefaultStyle sets the name of the preferred alternate style sheet.
	HTTPEquivDefaultStyle HTTPEquiv = "default-style"
	// HTTPEquivRefresh reloads or redirects the page after the number of seconds given in the content attribute.
	HTTPEquivRefresh HTTPEquiv = "refresh"
	// HTTPEquivXUACompatible selects the legacy Internet Explorer compatibility rendering mode.
	HTTPEquivXUACompatible HTTPEquiv = "x-ua-compatible"
	// HTTPEquivContentSecurityPolicy declares a Content Security Policy for the document.
	HTTPEquivContentSecurityPolicy HTTPEquiv = "content-security-policy"
)

// Valid indicates if v is any of the valid values for HTTPEquiv
func (v HTTPEquiv) Valid() bool {
	switch v {
	case
		HTTPEquivContentType,
		HTTPEquivDefaultStyle,
		HTTPEquivRefresh,
		HTTPEquivXUACompatible,
		HTTPEquivContentSecurityPolicy:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for HTTPEquiv
func (v HTTPEquiv) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type html.HTTPEquiv", v)
	}
	return nil
}

// Enums returns all valid values for HTTPEquiv
func (HTTPEquiv) Enums() []HTTPEquiv {
	return []HTTPEquiv{
		HTTPEquivContentType,
		HTTPEquivDefaultStyle,
		HTTPEquivRefresh,
		HTTPEquivXUACompatible,
		HTTPEquivContentSecurityPolicy,
	}
}

// EnumStrings returns all valid values for HTTPEquiv as strings
func (HTTPEquiv) EnumStrings() []string {
	return []string{
		"content-type",
		"default-style",
		"refresh",
		"x-ua-compatible",
		"content-security-policy",
	}
}

// String implements the fmt.Stringer interface for HTTPEquiv
func (v HTTPEquiv) String() string {
	return string(v)
}

// AttribName returns the "http-equiv" attribute name.
func (v HTTPEquiv) AttribName() string { return "http-equiv" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v HTTPEquiv) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// InputMode is the HTML inputmode global attribute.
type InputMode string //#enum

const (
	// InputModeNone shows no virtual keyboard, useful when the page implements its own input control.
	InputModeNone InputMode = "none"
	// InputModeText shows a standard text keyboard in the user's locale (the default).
	InputModeText InputMode = "text"
	// InputModeTel shows a telephone keypad with digits and the * and # keys.
	InputModeTel InputMode = "tel"
	// InputModeEmail shows a text keyboard optimized for entering email addresses.
	InputModeEmail InputMode = "email"
	// InputModeURL shows a text keyboard optimized for entering URLs.
	InputModeURL InputMode = "url"
	// InputModeNumeric shows a numeric keypad for entering digits.
	InputModeNumeric InputMode = "numeric"
	// InputModeDecimal shows a numeric keypad including the decimal separator.
	InputModeDecimal InputMode = "decimal"
	// InputModeSearch shows a keyboard optimized for search, typically with a "search" action key.
	InputModeSearch InputMode = "search"
)

// Valid indicates if v is any of the valid values for InputMode
func (v InputMode) Valid() bool {
	switch v {
	case
		InputModeNone,
		InputModeText,
		InputModeTel,
		InputModeEmail,
		InputModeURL,
		InputModeNumeric,
		InputModeDecimal,
		InputModeSearch:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for InputMode
func (v InputMode) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type html.InputMode", v)
	}
	return nil
}

// Enums returns all valid values for InputMode
func (InputMode) Enums() []InputMode {
	return []InputMode{
		InputModeNone,
		InputModeText,
		InputModeTel,
		InputModeEmail,
		InputModeURL,
		InputModeNumeric,
		InputModeDecimal,
		InputModeSearch,
	}
}

// EnumStrings returns all valid values for InputMode as strings
func (InputMode) EnumStrings() []string {
	return []string{
		"none",
		"text",
		"tel",
		"email",
		"url",
		"numeric",
		"decimal",
		"search",
	}
}

// String implements the fmt.Stringer interface for InputMode
func (v InputMode) String() string {
	return string(v)
}

// AttribName returns the "inputmode" attribute name.
func (v InputMode) AttribName() string { return "inputmode" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v InputMode) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Kind is the HTML kind attribute of the track element.
type Kind string //#enum

const (
	// KindSubtitles provides a translation or transcription of dialogue, shown when audio is available but not understood.
	KindSubtitles Kind = "subtitles"
	// KindCaptions provides a transcription of dialogue and sound effects, suitable when audio is unavailable.
	KindCaptions Kind = "captions"
	// KindDescriptions provides textual descriptions of the video for synthesis when the visuals are unavailable.
	KindDescriptions Kind = "descriptions"
	// KindChapters provides chapter titles used for navigating the media resource.
	KindChapters Kind = "chapters"
	// KindMetadata provides tracks intended for use by scripts and not displayed to the user.
	KindMetadata Kind = "metadata"
)

// Valid indicates if v is any of the valid values for Kind
func (v Kind) Valid() bool {
	switch v {
	case
		KindSubtitles,
		KindCaptions,
		KindDescriptions,
		KindChapters,
		KindMetadata:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for Kind
func (v Kind) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type html.Kind", v)
	}
	return nil
}

// Enums returns all valid values for Kind
func (Kind) Enums() []Kind {
	return []Kind{
		KindSubtitles,
		KindCaptions,
		KindDescriptions,
		KindChapters,
		KindMetadata,
	}
}

// EnumStrings returns all valid values for Kind as strings
func (Kind) EnumStrings() []string {
	return []string{
		"subtitles",
		"captions",
		"descriptions",
		"chapters",
		"metadata",
	}
}

// String implements the fmt.Stringer interface for Kind
func (v Kind) String() string {
	return string(v)
}

// AttribName returns the "kind" attribute name.
func (v Kind) AttribName() string { return "kind" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v Kind) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Loading is the HTML loading attribute of img and iframe elements.
type Loading string //#enum

const (
	// LoadingEager loads the resource immediately, regardless of its position in the viewport (the default).
	LoadingEager Loading = "eager"
	// LoadingLazy defers loading the resource until it nears the viewport.
	LoadingLazy Loading = "lazy"
)

// Valid indicates if v is any of the valid values for Loading
func (v Loading) Valid() bool {
	switch v {
	case
		LoadingEager,
		LoadingLazy:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for Loading
func (v Loading) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type html.Loading", v)
	}
	return nil
}

// Enums returns all valid values for Loading
func (Loading) Enums() []Loading {
	return []Loading{
		LoadingEager,
		LoadingLazy,
	}
}

// EnumStrings returns all valid values for Loading as strings
func (Loading) EnumStrings() []string {
	return []string{
		"eager",
		"lazy",
	}
}

// String implements the fmt.Stringer interface for Loading
func (v Loading) String() string {
	return string(v)
}

// AttribName returns the "loading" attribute name.
func (v Loading) AttribName() string { return "loading" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v Loading) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Method is the HTML method attribute of the form element.
type Method string //#enum

const (
	// MethodGET submits the form with the HTTP GET method, appending the data to the action URL (the default).
	MethodGET Method = "get"
	// MethodPOST submits the form with the HTTP POST method, sending the data in the request body.
	MethodPOST Method = "post"
	// MethodDialog closes the enclosing dialog and submits without sending data.
	MethodDialog Method = "dialog"
)

// Valid indicates if v is any of the valid values for Method
func (v Method) Valid() bool {
	switch v {
	case
		MethodGET,
		MethodPOST,
		MethodDialog:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for Method
func (v Method) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type html.Method", v)
	}
	return nil
}

// Enums returns all valid values for Method
func (Method) Enums() []Method {
	return []Method{
		MethodGET,
		MethodPOST,
		MethodDialog,
	}
}

// EnumStrings returns all valid values for Method as strings
func (Method) EnumStrings() []string {
	return []string{
		"get",
		"post",
		"dialog",
	}
}

// String implements the fmt.Stringer interface for Method
func (v Method) String() string {
	return string(v)
}

// AttribName returns the "method" attribute name.
func (v Method) AttribName() string { return "method" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v Method) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// ReferrerPolicy is the HTML referrerpolicy attribute.
type ReferrerPolicy string //#enum

const (
	// ReferrerPolicyNoReferrer never sends the Referer header.
	ReferrerPolicyNoReferrer ReferrerPolicy = "no-referrer"
	// ReferrerPolicyNoReferrerWhenDowngrade sends the full URL, but omits the Referer header when downgrading from HTTPS to HTTP.
	ReferrerPolicyNoReferrerWhenDowngrade ReferrerPolicy = "no-referrer-when-downgrade"
	// ReferrerPolicyOrigin sends only the origin (scheme, host, and port) as the referrer for all requests.
	ReferrerPolicyOrigin ReferrerPolicy = "origin"
	// ReferrerPolicyOriginWhenCrossOrigin sends the full URL for same-origin requests but only the origin for cross-origin requests.
	ReferrerPolicyOriginWhenCrossOrigin ReferrerPolicy = "origin-when-cross-origin"
	// ReferrerPolicySameOrigin sends the full URL for same-origin requests and no referrer for cross-origin requests.
	ReferrerPolicySameOrigin ReferrerPolicy = "same-origin"
	// ReferrerPolicyStrictOrigin sends only the origin, and nothing when downgrading from HTTPS to HTTP.
	ReferrerPolicyStrictOrigin ReferrerPolicy = "strict-origin"
	// ReferrerPolicyStrictOriginWhenCrossOrigin sends the full URL for same-origin requests, only the origin cross-origin, and nothing on an HTTPS-to-HTTP downgrade (the default).
	ReferrerPolicyStrictOriginWhenCrossOrigin ReferrerPolicy = "strict-origin-when-cross-origin"
	// ReferrerPolicyUnsafeUrl always sends the full URL, including on downgrades, which can leak data (unsafe).
	ReferrerPolicyUnsafeUrl ReferrerPolicy = "unsafe-url"
)

// Valid indicates if v is any of the valid values for ReferrerPolicy
func (v ReferrerPolicy) Valid() bool {
	switch v {
	case
		ReferrerPolicyNoReferrer,
		ReferrerPolicyNoReferrerWhenDowngrade,
		ReferrerPolicyOrigin,
		ReferrerPolicyOriginWhenCrossOrigin,
		ReferrerPolicySameOrigin,
		ReferrerPolicyStrictOrigin,
		ReferrerPolicyStrictOriginWhenCrossOrigin,
		ReferrerPolicyUnsafeUrl:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for ReferrerPolicy
func (v ReferrerPolicy) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type html.ReferrerPolicy", v)
	}
	return nil
}

// Enums returns all valid values for ReferrerPolicy
func (ReferrerPolicy) Enums() []ReferrerPolicy {
	return []ReferrerPolicy{
		ReferrerPolicyNoReferrer,
		ReferrerPolicyNoReferrerWhenDowngrade,
		ReferrerPolicyOrigin,
		ReferrerPolicyOriginWhenCrossOrigin,
		ReferrerPolicySameOrigin,
		ReferrerPolicyStrictOrigin,
		ReferrerPolicyStrictOriginWhenCrossOrigin,
		ReferrerPolicyUnsafeUrl,
	}
}

// EnumStrings returns all valid values for ReferrerPolicy as strings
func (ReferrerPolicy) EnumStrings() []string {
	return []string{
		"no-referrer",
		"no-referrer-when-downgrade",
		"origin",
		"origin-when-cross-origin",
		"same-origin",
		"strict-origin",
		"strict-origin-when-cross-origin",
		"unsafe-url",
	}
}

// String implements the fmt.Stringer interface for ReferrerPolicy
func (v ReferrerPolicy) String() string {
	return string(v)
}

// AttribName returns the "referrerpolicy" attribute name.
func (v ReferrerPolicy) AttribName() string { return "referrerpolicy" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v ReferrerPolicy) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Shape is the HTML shape attribute of the area element.
type Shape string //#enum

const (
	// ShapeDefault makes the entire enclosing image the active area.
	ShapeDefault Shape = "default"
	// ShapeRect defines a rectangular active area from the coords attribute.
	ShapeRect Shape = "rect"
	// ShapeCircle defines a circular active area from the coords attribute.
	ShapeCircle Shape = "circle"
	// ShapePoly defines a polygonal active area from the coords attribute.
	ShapePoly Shape = "poly"
)

// Valid indicates if v is any of the valid values for Shape
func (v Shape) Valid() bool {
	switch v {
	case
		ShapeDefault,
		ShapeRect,
		ShapeCircle,
		ShapePoly:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for Shape
func (v Shape) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type html.Shape", v)
	}
	return nil
}

// Enums returns all valid values for Shape
func (Shape) Enums() []Shape {
	return []Shape{
		ShapeDefault,
		ShapeRect,
		ShapeCircle,
		ShapePoly,
	}
}

// EnumStrings returns all valid values for Shape as strings
func (Shape) EnumStrings() []string {
	return []string{
		"default",
		"rect",
		"circle",
		"poly",
	}
}

// String implements the fmt.Stringer interface for Shape
func (v Shape) String() string {
	return string(v)
}

// AttribName returns the "shape" attribute name.
func (v Shape) AttribName() string { return "shape" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v Shape) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// As is the HTML as attribute of the link element (preload destination).
type As string //#enum

const (
	// AsAudio preloads the resource as audio data, as used by an audio element.
	AsAudio As = "audio"
	// AsDocument preloads the resource as an HTML document for embedding in a frame or iframe.
	AsDocument As = "document"
	// AsEmbed preloads the resource as content for an embed element.
	AsEmbed As = "embed"
	// AsFetch preloads the resource for retrieval via fetch or XHR (requires the crossorigin attribute).
	AsFetch As = "fetch"
	// AsFont preloads the resource as a font (requires the crossorigin attribute).
	AsFont As = "font"
	// AsImage preloads the resource as an image, as used by an img element.
	AsImage As = "image"
	// AsObject preloads the resource as content for an object element.
	AsObject As = "object"
	// AsScript preloads the resource as a script, as used by a script element.
	AsScript As = "script"
	// AsStyle preloads the resource as a style sheet, as used by a link rel=stylesheet element.
	AsStyle As = "style"
	// AsTrack preloads the resource as a WebVTT text track, as used by a track element.
	AsTrack As = "track"
	// AsVideo preloads the resource as video data, as used by a video element.
	AsVideo As = "video"
	// AsWorker preloads the resource as a Web Worker or Shared Worker script.
	AsWorker As = "worker"
)

// Valid indicates if v is any of the valid values for As
func (v As) Valid() bool {
	switch v {
	case
		AsAudio,
		AsDocument,
		AsEmbed,
		AsFetch,
		AsFont,
		AsImage,
		AsObject,
		AsScript,
		AsStyle,
		AsTrack,
		AsVideo,
		AsWorker:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for As
func (v As) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type html.As", v)
	}
	return nil
}

// Enums returns all valid values for As
func (As) Enums() []As {
	return []As{
		AsAudio,
		AsDocument,
		AsEmbed,
		AsFetch,
		AsFont,
		AsImage,
		AsObject,
		AsScript,
		AsStyle,
		AsTrack,
		AsVideo,
		AsWorker,
	}
}

// EnumStrings returns all valid values for As as strings
func (As) EnumStrings() []string {
	return []string{
		"audio",
		"document",
		"embed",
		"fetch",
		"font",
		"image",
		"object",
		"script",
		"style",
		"track",
		"video",
		"worker",
	}
}

// String implements the fmt.Stringer interface for As
func (v As) String() string {
	return string(v)
}

// AttribName returns the "as" attribute name.
func (v As) AttribName() string { return "as" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v As) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Capture is the HTML capture attribute of file input elements.
type Capture string //#enum

const (
	// CaptureUser requests the user-facing camera or microphone for capture.
	CaptureUser Capture = "user"
	// CaptureEnvironment requests the outward-facing camera or microphone for capture.
	CaptureEnvironment Capture = "environment"
)

// Valid indicates if v is any of the valid values for Capture
func (v Capture) Valid() bool {
	switch v {
	case
		CaptureUser,
		CaptureEnvironment:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for Capture
func (v Capture) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type html.Capture", v)
	}
	return nil
}

// Enums returns all valid values for Capture
func (Capture) Enums() []Capture {
	return []Capture{
		CaptureUser,
		CaptureEnvironment,
	}
}

// EnumStrings returns all valid values for Capture as strings
func (Capture) EnumStrings() []string {
	return []string{
		"user",
		"environment",
	}
}

// String implements the fmt.Stringer interface for Capture
func (v Capture) String() string {
	return string(v)
}

// AttribName returns the "capture" attribute name.
func (v Capture) AttribName() string { return "capture" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v Capture) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Preload is the HTML preload attribute of audio and video elements.
type Preload string //#enum

const (
	// PreloadNone hints that the media should not be preloaded, minimizing unnecessary traffic.
	PreloadNone Preload = "none"
	// PreloadMetadata hints that only the media metadata (such as duration) should be fetched.
	PreloadMetadata Preload = "metadata"
	// PreloadAuto hints that the whole media file may be downloaded even if the user is not expected to use it.
	PreloadAuto Preload = "auto"
)

// Valid indicates if v is any of the valid values for Preload
func (v Preload) Valid() bool {
	switch v {
	case
		PreloadNone,
		PreloadMetadata,
		PreloadAuto:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for Preload
func (v Preload) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type html.Preload", v)
	}
	return nil
}

// Enums returns all valid values for Preload
func (Preload) Enums() []Preload {
	return []Preload{
		PreloadNone,
		PreloadMetadata,
		PreloadAuto,
	}
}

// EnumStrings returns all valid values for Preload as strings
func (Preload) EnumStrings() []string {
	return []string{
		"none",
		"metadata",
		"auto",
	}
}

// String implements the fmt.Stringer interface for Preload
func (v Preload) String() string {
	return string(v)
}

// AttribName returns the "preload" attribute name.
func (v Preload) AttribName() string { return "preload" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v Preload) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Scope is the HTML scope attribute of the th element.
type Scope string //#enum

const (
	// ScopeRow marks the header as applying to the cells of its row.
	ScopeRow Scope = "row"
	// ScopeCol marks the header as applying to the cells of its column.
	ScopeCol Scope = "col"
	// ScopeRowGroup marks the header as applying to the cells of its row group.
	ScopeRowGroup Scope = "rowgroup"
	// ScopeColGroup marks the header as applying to the cells of its column group.
	ScopeColGroup Scope = "colgroup"
)

// Valid indicates if v is any of the valid values for Scope
func (v Scope) Valid() bool {
	switch v {
	case
		ScopeRow,
		ScopeCol,
		ScopeRowGroup,
		ScopeColGroup:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for Scope
func (v Scope) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type html.Scope", v)
	}
	return nil
}

// Enums returns all valid values for Scope
func (Scope) Enums() []Scope {
	return []Scope{
		ScopeRow,
		ScopeCol,
		ScopeRowGroup,
		ScopeColGroup,
	}
}

// EnumStrings returns all valid values for Scope as strings
func (Scope) EnumStrings() []string {
	return []string{
		"row",
		"col",
		"rowgroup",
		"colgroup",
	}
}

// String implements the fmt.Stringer interface for Scope
func (v Scope) String() string {
	return string(v)
}

// AttribName returns the "scope" attribute name.
func (v Scope) AttribName() string { return "scope" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v Scope) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// SpellCheck is the HTML spellcheck global attribute.
type SpellCheck string //#enum

const (
	// SpellCheckTrue enables spelling and grammar checking for the element's content.
	SpellCheckTrue SpellCheck = "true"
	// SpellCheckFalse disables spelling and grammar checking for the element's content.
	SpellCheckFalse SpellCheck = "false"
)

// Valid indicates if v is any of the valid values for SpellCheck
func (v SpellCheck) Valid() bool {
	switch v {
	case
		SpellCheckTrue,
		SpellCheckFalse:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for SpellCheck
func (v SpellCheck) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type html.SpellCheck", v)
	}
	return nil
}

// Enums returns all valid values for SpellCheck
func (SpellCheck) Enums() []SpellCheck {
	return []SpellCheck{
		SpellCheckTrue,
		SpellCheckFalse,
	}
}

// EnumStrings returns all valid values for SpellCheck as strings
func (SpellCheck) EnumStrings() []string {
	return []string{
		"true",
		"false",
	}
}

// String implements the fmt.Stringer interface for SpellCheck
func (v SpellCheck) String() string {
	return string(v)
}

// AttribName returns the "spellcheck" attribute name.
func (v SpellCheck) AttribName() string { return "spellcheck" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v SpellCheck) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Translate is the HTML translate global attribute.
type Translate string //#enum

const (
	// TranslateYes marks the element's content as translatable (the default).
	TranslateYes Translate = "yes"
	// TranslateNo marks the element's content as not to be translated.
	TranslateNo Translate = "no"
)

// Valid indicates if v is any of the valid values for Translate
func (v Translate) Valid() bool {
	switch v {
	case
		TranslateYes,
		TranslateNo:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for Translate
func (v Translate) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type html.Translate", v)
	}
	return nil
}

// Enums returns all valid values for Translate
func (Translate) Enums() []Translate {
	return []Translate{
		TranslateYes,
		TranslateNo,
	}
}

// EnumStrings returns all valid values for Translate as strings
func (Translate) EnumStrings() []string {
	return []string{
		"yes",
		"no",
	}
}

// String implements the fmt.Stringer interface for Translate
func (v Translate) String() string {
	return string(v)
}

// AttribName returns the "translate" attribute name.
func (v Translate) AttribName() string { return "translate" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v Translate) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Wrap is the HTML wrap attribute of the textarea element.
type Wrap string //#enum

const (
	// WrapSoft wraps text for display but submits it without inserted line breaks (the default).
	WrapSoft Wrap = "soft"
	// WrapHard inserts line breaks into the submitted text so it wraps; requires the cols attribute.
	WrapHard Wrap = "hard"
)

// Valid indicates if v is any of the valid values for Wrap
func (v Wrap) Valid() bool {
	switch v {
	case
		WrapSoft,
		WrapHard:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for Wrap
func (v Wrap) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type html.Wrap", v)
	}
	return nil
}

// Enums returns all valid values for Wrap
func (Wrap) Enums() []Wrap {
	return []Wrap{
		WrapSoft,
		WrapHard,
	}
}

// EnumStrings returns all valid values for Wrap as strings
func (Wrap) EnumStrings() []string {
	return []string{
		"soft",
		"hard",
	}
}

// String implements the fmt.Stringer interface for Wrap
func (v Wrap) String() string {
	return string(v)
}

// AttribName returns the "wrap" attribute name.
func (v Wrap) AttribName() string { return "wrap" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v Wrap) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Compile-time checks that every enum type is usable directly as an attribute.
var (
	_ mx.Attrib = AutoCapitalize("")
	_ mx.Attrib = AutoCorrect("")
	_ mx.Attrib = ContentEditable("")
	_ mx.Attrib = CrossOrigin("")
	_ mx.Attrib = Decoding("")
	_ mx.Attrib = Dir("")
	_ mx.Attrib = EncType("")
	_ mx.Attrib = EnterKeyHint("")
	_ mx.Attrib = FetchPriority("")
	_ mx.Attrib = FormEncType("")
	_ mx.Attrib = FormMethod("")
	_ mx.Attrib = HTTPEquiv("")
	_ mx.Attrib = InputMode("")
	_ mx.Attrib = Kind("")
	_ mx.Attrib = Loading("")
	_ mx.Attrib = Method("")
	_ mx.Attrib = ReferrerPolicy("")
	_ mx.Attrib = Shape("")
	_ mx.Attrib = As("")
	_ mx.Attrib = Capture("")
	_ mx.Attrib = Preload("")
	_ mx.Attrib = Scope("")
	_ mx.Attrib = SpellCheck("")
	_ mx.Attrib = Translate("")
	_ mx.Attrib = Wrap("")
)
