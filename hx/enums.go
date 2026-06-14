//go:generate go -C ../tools tool go-enum ../hx/$GOFILE

// This file defines htmx attribute values that form a strict, closed set of
// keyword tokens as enum types. Each type is a string that implements mx.Attrib,
// so the typed constants can be used directly as element attributes, e.g.
// html.Div(hx.SwapOuterHTML). A conversion such as SwapStyle("innerHTML") also
// works for dynamic values. AttribValue returns the value together with the
// result of the generated Validate method, so rendering an element that holds an
// invalid enum value fails with a descriptive error instead of emitting an
// invalid keyword.
//
// The go:generate directive runs the go-enum tool pinned in the nested tools
// module (kept out of the shipped go-mx dependency tree). Run it with:
//
//	go generate ./hx/...
//
// go-enum appends the Valid, Validate, Enums, EnumStrings and String methods for
// every //#enum type; the hand-written AttribName/AttribValue methods are left
// untouched.

package hx

import (
	"context"
	"fmt"

	"github.com/ungerik/go-mx"
)

// SwapStyle is the htmx hx-swap style keyword controlling how response content
// is swapped into the target (an enumerated keyword). Swap modifiers such as
// "swap:1s" are appended separately via [Swap]; the style itself is one of the
// constants below.
type SwapStyle string //#enum

const (
	// SwapInnerHTML replaces the inner HTML of the target element.
	SwapInnerHTML SwapStyle = "innerHTML"
	// SwapOuterHTML replaces the entire target element with the response.
	SwapOuterHTML SwapStyle = "outerHTML"
	// SwapTextContent replaces the text content of the target without parsing the response as HTML.
	SwapTextContent SwapStyle = "textContent"
	// SwapBeforeBegin inserts the response before the target element.
	SwapBeforeBegin SwapStyle = "beforebegin"
	// SwapAfterBegin inserts the response before the first child of the target element.
	SwapAfterBegin SwapStyle = "afterbegin"
	// SwapBeforeEnd inserts the response after the last child of the target element.
	SwapBeforeEnd SwapStyle = "beforeend"
	// SwapAfterEnd inserts the response after the target element.
	SwapAfterEnd SwapStyle = "afterend"
	// SwapDelete deletes the target element regardless of the response.
	SwapDelete SwapStyle = "delete"
	// SwapNone does not append content from the response.
	SwapNone SwapStyle = "none"
)

// Valid indicates if v is any of the valid values for SwapStyle
func (v SwapStyle) Valid() bool {
	switch v {
	case
		SwapInnerHTML,
		SwapOuterHTML,
		SwapTextContent,
		SwapBeforeBegin,
		SwapAfterBegin,
		SwapBeforeEnd,
		SwapAfterEnd,
		SwapDelete,
		SwapNone:
		return true
	}
	return false
}

// Validate returns an error if v is none of the valid values for SwapStyle
func (v SwapStyle) Validate() error {
	if !v.Valid() {
		return fmt.Errorf("invalid value %#v for type hx.SwapStyle", v)
	}
	return nil
}

// Enums returns all valid values for SwapStyle
func (SwapStyle) Enums() []SwapStyle {
	return []SwapStyle{
		SwapInnerHTML,
		SwapOuterHTML,
		SwapTextContent,
		SwapBeforeBegin,
		SwapAfterBegin,
		SwapBeforeEnd,
		SwapAfterEnd,
		SwapDelete,
		SwapNone,
	}
}

// EnumStrings returns all valid values for SwapStyle as strings
func (SwapStyle) EnumStrings() []string {
	return []string{
		"innerHTML",
		"outerHTML",
		"textContent",
		"beforebegin",
		"afterbegin",
		"beforeend",
		"afterend",
		"delete",
		"none",
	}
}

// String implements the fmt.Stringer interface for SwapStyle
func (v SwapStyle) String() string {
	return string(v)
}

// AttribName returns the "hx-swap" attribute name.
func (v SwapStyle) AttribName() string { return "hx-swap" }

// AttribValue returns the keyword value together with the result of Validate,
// so rendering an invalid value fails with a descriptive error.
func (v SwapStyle) AttribValue(context.Context) (string, error) { return string(v), v.Validate() }

// Compile-time checks that every enum type is usable directly as an attribute.
var (
	_ mx.Attrib = SwapStyle("")
)
