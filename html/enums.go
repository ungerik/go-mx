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
	AutoCapitalizeOff        AutoCapitalize = "off"
	AutoCapitalizeNone       AutoCapitalize = "none"
	AutoCapitalizeOn         AutoCapitalize = "on"
	AutoCapitalizeSentences  AutoCapitalize = "sentences"
	AutoCapitalizeWords      AutoCapitalize = "words"
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

func (v AutoCapitalize) AttribName() string                          { return "autocapitalize" }
func (v AutoCapitalize) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// AutoCorrect is the HTML autocorrect attribute.
type AutoCorrect string //#enum

const (
	AutoCorrectOn  AutoCorrect = "on"
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

func (v AutoCorrect) AttribName() string                          { return "autocorrect" }
func (v AutoCorrect) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// ContentEditable is the HTML contenteditable global attribute.
type ContentEditable string //#enum

const (
	ContentEditableTrue          ContentEditable = "true"
	ContentEditableFalse         ContentEditable = "false"
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

func (v ContentEditable) AttribName() string                          { return "contenteditable" }
func (v ContentEditable) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// CrossOrigin is the HTML crossorigin attribute.
type CrossOrigin string //#enum

const (
	CrossOriginAnonymous      CrossOrigin = "anonymous"
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

func (v CrossOrigin) AttribName() string                          { return "crossorigin" }
func (v CrossOrigin) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Decoding is the HTML decoding attribute of the img element.
type Decoding string //#enum

const (
	DecodingAuto  Decoding = "auto"
	DecodingAsync Decoding = "async"
	DecodingSync  Decoding = "sync"
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

func (v Decoding) AttribName() string                          { return "decoding" }
func (v Decoding) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Dir is the HTML dir global attribute (text directionality).
type Dir string //#enum

const (
	DirLTR  Dir = "ltr"
	DirRTL  Dir = "rtl"
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

func (v Dir) AttribName() string                          { return "dir" }
func (v Dir) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// EncType is the HTML enctype attribute of the form element.
type EncType string //#enum

const (
	EncTypeFormURLEncoded    EncType = "application/x-www-form-urlencoded"
	EncTypeMultipartFormData EncType = "multipart/form-data"
	EncTypeTextPlain         EncType = "text/plain"
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

func (v EncType) AttribName() string                          { return "enctype" }
func (v EncType) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// EnterKeyHint is the HTML enterkeyhint global attribute.
type EnterKeyHint string //#enum

const (
	EnterKeyHintEnter    EnterKeyHint = "enter"
	EnterKeyHintDone     EnterKeyHint = "done"
	EnterKeyHintGo       EnterKeyHint = "go"
	EnterKeyHintNext     EnterKeyHint = "next"
	EnterKeyHintPrevious EnterKeyHint = "previous"
	EnterKeyHintSearch   EnterKeyHint = "search"
	EnterKeyHintSend     EnterKeyHint = "send"
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

func (v EnterKeyHint) AttribName() string                          { return "enterkeyhint" }
func (v EnterKeyHint) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// FetchPriority is the HTML fetchpriority attribute.
type FetchPriority string //#enum

const (
	FetchPriorityAuto FetchPriority = "auto"
	FetchPriorityHigh FetchPriority = "high"
	FetchPriorityLow  FetchPriority = "low"
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

func (v FetchPriority) AttribName() string                          { return "fetchpriority" }
func (v FetchPriority) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// FormEncType is the HTML formenctype attribute of submit buttons.
type FormEncType string //#enum

const (
	FormEncTypeFormURLEncoded    FormEncType = "application/x-www-form-urlencoded"
	FormEncTypeMultipartFormData FormEncType = "multipart/form-data"
	FormEncTypeTextPlain         FormEncType = "text/plain"
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

func (v FormEncType) AttribName() string                          { return "formenctype" }
func (v FormEncType) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// FormMethod is the HTML formmethod attribute of submit buttons.
type FormMethod string //#enum

const (
	FormMethodGET    FormMethod = "get"
	FormMethodPOST   FormMethod = "post"
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

func (v FormMethod) AttribName() string                          { return "formmethod" }
func (v FormMethod) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// HTTPEquiv is the HTML http-equiv attribute of the meta element.
type HTTPEquiv string //#enum

const (
	HTTPEquivContentType           HTTPEquiv = "content-type"
	HTTPEquivDefaultStyle          HTTPEquiv = "default-style"
	HTTPEquivRefresh               HTTPEquiv = "refresh"
	HTTPEquivXUACompatible         HTTPEquiv = "x-ua-compatible"
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

func (v HTTPEquiv) AttribName() string                          { return "http-equiv" }
func (v HTTPEquiv) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// InputMode is the HTML inputmode global attribute.
type InputMode string //#enum

const (
	InputModeNone    InputMode = "none"
	InputModeText    InputMode = "text"
	InputModeTel     InputMode = "tel"
	InputModeEmail   InputMode = "email"
	InputModeURL     InputMode = "url"
	InputModeNumeric InputMode = "numeric"
	InputModeDecimal InputMode = "decimal"
	InputModeSearch  InputMode = "search"
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

func (v InputMode) AttribName() string                          { return "inputmode" }
func (v InputMode) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Kind is the HTML kind attribute of the track element.
type Kind string //#enum

const (
	KindSubtitles    Kind = "subtitles"
	KindCaptions     Kind = "captions"
	KindDescriptions Kind = "descriptions"
	KindChapters     Kind = "chapters"
	KindMetadata     Kind = "metadata"
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

func (v Kind) AttribName() string                          { return "kind" }
func (v Kind) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Loading is the HTML loading attribute of img and iframe elements.
type Loading string //#enum

const (
	LoadingEager Loading = "eager"
	LoadingLazy  Loading = "lazy"
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

func (v Loading) AttribName() string                          { return "loading" }
func (v Loading) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Method is the HTML method attribute of the form element.
type Method string //#enum

const (
	MethodGET    Method = "GET"
	MethodPOST   Method = "POST"
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
		"GET",
		"POST",
		"dialog",
	}
}

// String implements the fmt.Stringer interface for Method
func (v Method) String() string {
	return string(v)
}

func (v Method) AttribName() string                          { return "method" }
func (v Method) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// ReferrerPolicy is the HTML referrerpolicy attribute.
type ReferrerPolicy string //#enum

const (
	ReferrerPolicyNoReferrer                  ReferrerPolicy = "no-referrer"
	ReferrerPolicyNoReferrerWhenDowngrade     ReferrerPolicy = "no-referrer-when-downgrade"
	ReferrerPolicyOrigin                      ReferrerPolicy = "origin"
	ReferrerPolicyOriginWhenCrossOrigin       ReferrerPolicy = "origin-when-cross-origin"
	ReferrerPolicySameOrigin                  ReferrerPolicy = "same-origin"
	ReferrerPolicyStrictOrigin                ReferrerPolicy = "strict-origin"
	ReferrerPolicyStrictOriginWhenCrossOrigin ReferrerPolicy = "strict-origin-when-cross-origin"
	ReferrerPolicyUnsafeUrl                   ReferrerPolicy = "unsafe-url"
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

func (v ReferrerPolicy) AttribName() string                          { return "referrerpolicy" }
func (v ReferrerPolicy) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Shape is the HTML shape attribute of the area element.
type Shape string //#enum

const (
	ShapeDefault Shape = "default"
	ShapeRect    Shape = "rect"
	ShapeCircle  Shape = "circle"
	ShapePoly    Shape = "poly"
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

func (v Shape) AttribName() string                          { return "shape" }
func (v Shape) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// As is the HTML as attribute of the link element (preload destination).
type As string //#enum

const (
	AsAudio    As = "audio"
	AsDocument As = "document"
	AsEmbed    As = "embed"
	AsFetch    As = "fetch"
	AsFont     As = "font"
	AsImage    As = "image"
	AsObject   As = "object"
	AsScript   As = "script"
	AsStyle    As = "style"
	AsTrack    As = "track"
	AsVideo    As = "video"
	AsWorker   As = "worker"
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

func (v As) AttribName() string                          { return "as" }
func (v As) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Capture is the HTML capture attribute of file input elements.
type Capture string //#enum

const (
	CaptureUser        Capture = "user"
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

func (v Capture) AttribName() string                          { return "capture" }
func (v Capture) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Preload is the HTML preload attribute of audio and video elements.
type Preload string //#enum

const (
	PreloadNone     Preload = "none"
	PreloadMetadata Preload = "metadata"
	PreloadAuto     Preload = "auto"
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

func (v Preload) AttribName() string                          { return "preload" }
func (v Preload) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Scope is the HTML scope attribute of the th element.
type Scope string //#enum

const (
	ScopeRow      Scope = "row"
	ScopeCol      Scope = "col"
	ScopeRowGroup Scope = "rowgroup"
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

func (v Scope) AttribName() string                          { return "scope" }
func (v Scope) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// SpellCheck is the HTML spellcheck global attribute.
type SpellCheck string //#enum

const (
	SpellCheckTrue  SpellCheck = "true"
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

func (v SpellCheck) AttribName() string                          { return "spellcheck" }
func (v SpellCheck) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Translate is the HTML translate global attribute.
type Translate string //#enum

const (
	TranslateYes Translate = "yes"
	TranslateNo  Translate = "no"
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

func (v Translate) AttribName() string                          { return "translate" }
func (v Translate) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Wrap is the HTML wrap attribute of the textarea element.
type Wrap string //#enum

const (
	WrapSoft Wrap = "soft"
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

func (v Wrap) AttribName() string                          { return "wrap" }
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
