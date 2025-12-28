package hx

import "github.com/ungerik/go-mx"

// See https://htmx.org/reference/#attributes

// Get issues a GET to the specified URL
func Get(url string) mx.Attrib { return mx.NewAttrib("hx-get", url) }

// Post issues a POST to the specified URL
func Post(url string) mx.Attrib { return mx.NewAttrib("hx-post", url) }

// On handle events with inline scripts on elements
func On(execute string) mx.Attrib { return mx.NewAttrib("hx-on*", execute) }

// PushURL push a URL into the browser location bar to create history
func PushURL(url string) mx.Attrib { return mx.NewAttrib("hx-push-url", url) }

// Select select content to swap in from a response
func Select(value string) mx.Attrib { return mx.NewAttrib("hx-select", value) }

// SelectOOB select content to swap in from a response, somewhere other than the target (out of band)
func SelectOOB(value string) mx.Attrib { return mx.NewAttrib("hx-select-oob", value) }

// Swap controls how content will swap in (outerHTML, beforeend, afterend, …)
func Swap(value string) mx.Attrib { return mx.NewAttrib("hx-swap", value) }

// SwapOOB mark element to swap in from a response (out of band)
func SwapOOB(value string) mx.Attrib { return mx.NewAttrib("hx-swap-oob", value) }

// Target specifies the target element to be swapped
func Target(value string) mx.Attrib { return mx.NewAttrib("hx-target", value) }

// Trigger specifies the event that triggers the request
func Trigger(value string) mx.Attrib { return mx.NewAttrib("hx-trigger", value) }

// Vals add values to submit with the request (JSON format)
func Vals(value string) mx.Attrib { return mx.NewAttrib("hx-vals", value) }

// Boost add progressive enhancement for links and forms
func Boost(value string) mx.Attrib { return mx.NewAttrib("hx-boost", value) }

// Confirm shows a confirm() dialog before issuing a request
func Confirm(value string) mx.Attrib { return mx.NewAttrib("hx-confirm", value) }

// Delete issues a DELETE to the specified URL
func Delete(value string) mx.Attrib { return mx.NewAttrib("hx-delete", value) }

// Disable disables htmx processing for the given node and any children nodes
func Disable(value string) mx.Attrib { return mx.NewAttrib("hx-disable", value) }

// DisabledElt adds the disabled attribute to the specified elements while a request is in flight
func DisabledElt(value string) mx.Attrib { return mx.NewAttrib("hx-disabled-elt", value) }

// Disinherit control and disable automatic attribute inheritance for child nodes
func Disinherit(value string) mx.Attrib { return mx.NewAttrib("hx-disinherit", value) }

// Encoding changes the request encoding type
func Encoding(value string) mx.Attrib { return mx.NewAttrib("hx-encoding", value) }

// Ext extensions to use for this element
func Ext(value string) mx.Attrib { return mx.NewAttrib("hx-ext", value) }

// Headers adds to the headers that will be submitted with the request
func Headers(value string) mx.Attrib { return mx.NewAttrib("hx-headers", value) }

// History prevent sensitive data being saved to the history cache
func History(value string) mx.Attrib { return mx.NewAttrib("hx-history", value) }

// HistoryElt the element to snapshot and restore during history navigation
func HistoryElt(value string) mx.Attrib { return mx.NewAttrib("hx-history-elt", value) }

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

// Preserve specifies elements to keep unchanged between requests
func Preserve(value string) mx.Attrib { return mx.NewAttrib("hx-preserve", value) }

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

// Validate force elements to validate themselves before a request
func Validate(value string) mx.Attrib { return mx.NewAttrib("hx-validate", value) }
