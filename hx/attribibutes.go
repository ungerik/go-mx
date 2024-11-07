package hx

import "github.com/ungerik/go-mx"

// See https://htmx.org/reference/#attributes

func Get(value string) mx.Attrib         { return mx.Attrib{Name: "hx-get", Value: value} }          // issues a GET to the specified URL
func Post(value string) mx.Attrib        { return mx.Attrib{Name: "hx-post", Value: value} }         // issues a POST to the specified URL
func On(value string) mx.Attrib          { return mx.Attrib{Name: "hx-on*", Value: value} }          // handle events with inline scripts on elements
func PushURL(value string) mx.Attrib     { return mx.Attrib{Name: "hx-push-url", Value: value} }     // push a URL into the browser location bar to create history
func Select(value string) mx.Attrib      { return mx.Attrib{Name: "hx-select", Value: value} }       // select content to swap in from a response
func SelectOOB(value string) mx.Attrib   { return mx.Attrib{Name: "hx-select-oob", Value: value} }   // select content to swap in from a response, somewhere other than the target (out of band)
func Swap(value string) mx.Attrib        { return mx.Attrib{Name: "hx-swap", Value: value} }         // controls how content will swap in (outerHTML, beforeend, afterend, …)
func SwapOOB(value string) mx.Attrib     { return mx.Attrib{Name: "hx-swap-oob", Value: value} }     // mark element to swap in from a response (out of band)
func Target(value string) mx.Attrib      { return mx.Attrib{Name: "hx-target", Value: value} }       // specifies the target element to be swapped
func Trigger(value string) mx.Attrib     { return mx.Attrib{Name: "hx-trigger", Value: value} }      // specifies the event that triggers the request
func Vals(value string) mx.Attrib        { return mx.Attrib{Name: "hx-vals", Value: value} }         // add values to submit with the request (JSON format)
func Boost(value string) mx.Attrib       { return mx.Attrib{Name: "hx-boost", Value: value} }        // add progressive enhancement for links and forms
func Confirm(value string) mx.Attrib     { return mx.Attrib{Name: "hx-confirm", Value: value} }      // shows a confirm() dialog before issuing a request
func Delete(value string) mx.Attrib      { return mx.Attrib{Name: "hx-delete", Value: value} }       // issues a DELETE to the specified URL
func Disable(value string) mx.Attrib     { return mx.Attrib{Name: "hx-disable", Value: value} }      // disables htmx processing for the given node and any children nodes
func DisabledElt(value string) mx.Attrib { return mx.Attrib{Name: "hx-disabled-elt", Value: value} } // adds the disabled attribute to the specified elements while a request is in flight
func Disinherit(value string) mx.Attrib  { return mx.Attrib{Name: "hx-disinherit", Value: value} }   // control and disable automatic attribute inheritance for child nodes
func Encoding(value string) mx.Attrib    { return mx.Attrib{Name: "hx-encoding", Value: value} }     // changes the request encoding type
func Ext(value string) mx.Attrib         { return mx.Attrib{Name: "hx-ext", Value: value} }          // extensions to use for this element
func Headers(value string) mx.Attrib     { return mx.Attrib{Name: "hx-headers", Value: value} }      // adds to the headers that will be submitted with the request
func History(value string) mx.Attrib     { return mx.Attrib{Name: "hx-history", Value: value} }      // prevent sensitive data being saved to the history cache
func HistoryElt(value string) mx.Attrib  { return mx.Attrib{Name: "hx-history-elt", Value: value} }  // the element to snapshot and restore during history navigation
func Include(value string) mx.Attrib     { return mx.Attrib{Name: "hx-include", Value: value} }      // include additional data in requests
func Indicator(value string) mx.Attrib   { return mx.Attrib{Name: "hx-indicator", Value: value} }    // the element to put the htmx-request class on during the request
func Inherit(value string) mx.Attrib     { return mx.Attrib{Name: "hx-inherit", Value: value} }      // control and enable automatic attribute inheritance for child nodes if it has been disabled by default
func Params(value string) mx.Attrib      { return mx.Attrib{Name: "hx-params", Value: value} }       // filters the parameters that will be submitted with a request
func Patch(value string) mx.Attrib       { return mx.Attrib{Name: "hx-patch", Value: value} }        // issues a PATCH to the specified URL
func Preserve(value string) mx.Attrib    { return mx.Attrib{Name: "hx-preserve", Value: value} }     // specifies elements to keep unchanged between requests
func Prompt(value string) mx.Attrib      { return mx.Attrib{Name: "hx-prompt", Value: value} }       // shows a prompt() before submitting a request
func Put(value string) mx.Attrib         { return mx.Attrib{Name: "hx-put", Value: value} }          // issues a PUT to the specified URL
func ReplaceURL(value string) mx.Attrib  { return mx.Attrib{Name: "hx-replace-url", Value: value} }  // replace the URL in the browser location bar
func Request(value string) mx.Attrib     { return mx.Attrib{Name: "hx-request", Value: value} }      // configures various aspects of the request
func Sync(value string) mx.Attrib        { return mx.Attrib{Name: "hx-sync", Value: value} }         // control how requests made by different elements are synchronized
func Validate(value string) mx.Attrib    { return mx.Attrib{Name: "hx-validate", Value: value} }     // force elements to validate themselves before a request
