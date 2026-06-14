package hx

import (
	"strconv"
	"strings"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// See https://htmx.org/reference/#attributes

// Get issues a GET to the specified URL
func Get(url string) mx.Attrib { return mx.NewAttrib("hx-get", url) }

// Post issues a POST to the specified URL
func Post(url string) mx.Attrib { return mx.NewAttrib("hx-post", url) }

// On handles a DOM event with an inline script on the element.
// It renders as the hx-on:<event> attribute, so On("click", "alert('hi')")
// renders as hx-on:click="alert('hi')".
// For htmx events use [OnHTMX].
// See https://htmx.org/attributes/hx-on/
func On(event, handler string) mx.Attrib { return mx.NewAttrib("hx-on:"+event, handler) }

// OnHTMX handles an htmx event with an inline script on the element.
// It renders as the hx-on::<event> attribute, which is shorthand for
// hx-on:htmx:<event>, so OnHTMX("after-request", "doStuff()")
// renders as hx-on::after-request="doStuff()".
// For DOM events use [On].
// See https://htmx.org/attributes/hx-on/
func OnHTMX(event, handler string) mx.Attrib { return mx.NewAttrib("hx-on::"+event, handler) }

// PushURL push a URL into the browser location bar to create history
func PushURL(url string) mx.Attrib { return mx.NewAttrib("hx-push-url", url) }

// Select select content to swap in from a response
func Select(value string) mx.Attrib { return mx.NewAttrib("hx-select", value) }

// SelectOOB select content to swap in from a response, somewhere other than the target (out of band)
func SelectOOB(value string) mx.Attrib { return mx.NewAttrib("hx-select-oob", value) }

// Swap controls how content will swap in, using a typed [SwapStyle] keyword and
// optional modifiers (e.g. "swap:1s", "settle:1s", "scroll:bottom"). The
// modifiers are passed through verbatim; an invalid style defers an error to
// render time.
func Swap(style SwapStyle, modifiers ...string) mx.Attrib {
	if err := style.Validate(); err != nil {
		return mx.ErrAttrib{Name: "hx-swap", Err: err}
	}
	value := string(style)
	if len(modifiers) > 0 {
		value += " " + strings.Join(modifiers, " ")
	}
	return mx.NewAttrib("hx-swap", value)
}

// SwapOOB marks the element to swap in from a response out of band, using the
// given [SwapStyle] and optional target CSS selector(s). Use [SwapOOBTrue] for
// the plain hx-swap-oob="true" form. An invalid style defers an error to render
// time.
func SwapOOB(style SwapStyle, selector ...string) mx.Attrib {
	if err := style.Validate(); err != nil {
		return mx.ErrAttrib{Name: "hx-swap-oob", Err: err}
	}
	value := string(style)
	if len(selector) > 0 {
		value += ":" + strings.Join(selector, ",")
	}
	return mx.NewAttrib("hx-swap-oob", value)
}

// SwapOOBTrue marks the element to swap in out of band using its own id,
// rendering hx-swap-oob="true".
const SwapOOBTrue = mx.ConstAttrib("hx-swap-oob=true")

// Target specifies the target element to be swapped
func Target(value string) mx.Attrib { return mx.NewAttrib("hx-target", value) }

// Trigger specifies the event that triggers the request
func Trigger(value string) mx.Attrib { return mx.NewAttrib("hx-trigger", value) }

// Vals add values to submit with the request (JSON format)
func Vals(value string) mx.Attrib { return mx.NewAttrib("hx-vals", value) }

// Boost enables (true) or disables (false) progressive enhancement for links
// and forms. htmx accepts only "true" or "false" as the attribute value.
func Boost(enable bool) mx.Attrib { return mx.NewAttrib("hx-boost", strconv.FormatBool(enable)) }

// Confirm shows a confirm() dialog before issuing a request
func Confirm(value string) mx.Attrib { return mx.NewAttrib("hx-confirm", value) }

// Delete issues a DELETE to the specified URL
func Delete(value string) mx.Attrib { return mx.NewAttrib("hx-delete", value) }

// Disable disables htmx processing for the given node and any children nodes.
// htmx ignores the attribute value, so it renders as a bare boolean attribute.
const Disable = html.BoolAttrib("hx-disable")

// DisabledElt adds the disabled attribute to the specified elements while a request is in flight
func DisabledElt(value string) mx.Attrib { return mx.NewAttrib("hx-disabled-elt", value) }

// Disinherit control and disable automatic attribute inheritance for child nodes
func Disinherit(value string) mx.Attrib { return mx.NewAttrib("hx-disinherit", value) }

// EncodingMultipart changes the request encoding type to multipart/form-data,
// the only value htmx accepts (used for file uploads).
const EncodingMultipart = mx.ConstAttrib("hx-encoding=multipart/form-data")

// Ext extensions to use for this element
func Ext(value string) mx.Attrib { return mx.NewAttrib("hx-ext", value) }

// Headers adds to the headers that will be submitted with the request
func Headers(value string) mx.Attrib { return mx.NewAttrib("hx-headers", value) }

// History controls whether the page state is saved to the history cache. Set it
// to false on any element to prevent sensitive data being saved. htmx accepts
// only "true" or "false" as the attribute value.
func History(enable bool) mx.Attrib { return mx.NewAttrib("hx-history", strconv.FormatBool(enable)) }

// HistoryElt marks the element to snapshot and restore during history
// navigation. htmx takes no value, so it renders as a bare boolean attribute.
const HistoryElt = html.BoolAttrib("hx-history-elt")

// Include include additional data in requests
func Include(value string) mx.Attrib { return mx.NewAttrib("hx-include", value) }

// Indicator the element to put the htmx-request class on during the request
func Indicator(value string) mx.Attrib { return mx.NewAttrib("hx-indicator", value) }

// Inherit control and enable automatic attribute inheritance for child nodes if it has been disabled by default
func Inherit(value string) mx.Attrib { return mx.NewAttrib("hx-inherit", value) }

// Params filters the parameters that will be submitted with a request
func Params(value string) mx.Attrib { return mx.NewAttrib("hx-params", value) }

// Patch issues a PATCH to the specified URL
func Patch(value string) mx.Attrib { return mx.NewAttrib("hx-patch", value) }

// Preserve keeps the element unchanged between requests. htmx ignores the
// attribute value, so it renders as a bare boolean attribute.
const Preserve = html.BoolAttrib("hx-preserve")

// Prompt shows a prompt() before submitting a request
func Prompt(value string) mx.Attrib { return mx.NewAttrib("hx-prompt", value) }

// Put issues a PUT to the specified URL
func Put(value string) mx.Attrib { return mx.NewAttrib("hx-put", value) }

// ReplaceURL replace the URL in the browser location bar
func ReplaceURL(value string) mx.Attrib { return mx.NewAttrib("hx-replace-url", value) }

// Request configures various aspects of the request
func Request(value string) mx.Attrib { return mx.NewAttrib("hx-request", value) }

// Sync control how requests made by different elements are synchronized
func Sync(value string) mx.Attrib { return mx.NewAttrib("hx-sync", value) }

// Validate forces the element to validate itself before a request when true.
// htmx accepts only "true" or "false" as the attribute value.
func Validate(enable bool) mx.Attrib { return mx.NewAttrib("hx-validate", strconv.FormatBool(enable)) }
