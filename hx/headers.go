package hx

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/domonda/go-errs"
)

// HTTP header names used by htmx.
// See https://htmx.org/reference/#request_headers
// and https://htmx.org/reference/#response_headers
const (
	// Request headers htmx sends with each request.

	HeaderRequest               = "HX-Request"                 // always "true" for htmx requests
	HeaderBoosted               = "HX-Boosted"                 // "true" if the request is via an element using hx-boost
	HeaderTrigger               = "HX-Trigger"                 // request: id of the triggered element; response: client-side events to trigger
	HeaderTriggerName           = "HX-Trigger-Name"            // name of the triggered element if it has one
	HeaderTarget                = "HX-Target"                  // id of the target element if it has one
	HeaderCurrentURL            = "HX-Current-URL"             // current URL of the browser
	HeaderPrompt                = "HX-Prompt"                  // user response to an hx-prompt
	HeaderHistoryRestoreRequest = "HX-History-Restore-Request" // "true" if for history restoration after a cache miss

	// Response headers a handler can set to control htmx.
	// HeaderTrigger above doubles as a response header.

	HeaderLocation           = "HX-Location"             // client-side redirect without a full page reload
	HeaderPushURL            = "HX-Push-Url"             // push a new URL into the history stack
	HeaderRedirect           = "HX-Redirect"             // client-side redirect to a new location
	HeaderRefresh            = "HX-Refresh"              // "true" makes the client do a full refresh of the page
	HeaderReplaceURL         = "HX-Replace-Url"          // replace the current URL in the location bar
	HeaderReswap             = "HX-Reswap"               // how the response is swapped, see hx-swap
	HeaderRetarget           = "HX-Retarget"             // CSS selector that updates the target of the content update
	HeaderReselect           = "HX-Reselect"             // CSS selector to choose the part of the response to swap in
	HeaderTriggerAfterSettle = "HX-Trigger-After-Settle" // client-side events to trigger after the settle step
	HeaderTriggerAfterSwap   = "HX-Trigger-After-Swap"   // client-side events to trigger after the swap step
)

// IsRequest reports whether r was made by htmx (HX-Request: true).
func IsRequest(r *http.Request) bool {
	return r.Header.Get(HeaderRequest) == "true"
}

// IsBoosted reports whether r was made via an element using hx-boost (HX-Boosted: true).
func IsBoosted(r *http.Request) bool {
	return r.Header.Get(HeaderBoosted) == "true"
}

// IsHistoryRestoreRequest reports whether r is for history restoration
// after a miss in the local history cache (HX-History-Restore-Request: true).
func IsHistoryRestoreRequest(r *http.Request) bool {
	return r.Header.Get(HeaderHistoryRestoreRequest) == "true"
}

// SetLocation sets the HX-Location response header, instructing htmx to do a
// client-side redirect to url that does not do a full page reload.
// For a redirect with extra context (target, swap, values, …) set
// HeaderLocation to a JSON object directly.
func SetLocation(w http.ResponseWriter, url string) {
	w.Header().Set(HeaderLocation, url)
}

// SetPushURL sets the HX-Push-Url response header, pushing url into the browser
// history stack. Pass "false" to prevent the browser history from being updated.
func SetPushURL(w http.ResponseWriter, url string) {
	w.Header().Set(HeaderPushURL, url)
}

// SetRedirect sets the HX-Redirect response header, making htmx do a
// client-side redirect to url.
func SetRedirect(w http.ResponseWriter, url string) {
	w.Header().Set(HeaderRedirect, url)
}

// SetRefresh sets the HX-Refresh response header to "true", making htmx do a
// full refresh of the page.
func SetRefresh(w http.ResponseWriter) {
	w.Header().Set(HeaderRefresh, "true")
}

// SetReplaceURL sets the HX-Replace-Url response header, replacing the current
// URL in the location bar with url. Pass "false" to prevent the browser
// location from being updated.
func SetReplaceURL(w http.ResponseWriter, url string) {
	w.Header().Set(HeaderReplaceURL, url)
}

// SetReswap sets the HX-Reswap response header, overriding how the response is
// swapped in. value uses the hx-swap syntax (outerHTML, beforeend, …).
func SetReswap(w http.ResponseWriter, value string) {
	w.Header().Set(HeaderReswap, value)
}

// SetRetarget sets the HX-Retarget response header to a CSS selector that
// updates the target of the content update to a different element on the page.
func SetRetarget(w http.ResponseWriter, selector string) {
	w.Header().Set(HeaderRetarget, selector)
}

// SetReselect sets the HX-Reselect response header to a CSS selector that
// chooses which part of the response is swapped in, overriding hx-select on
// the triggering element.
func SetReselect(w http.ResponseWriter, selector string) {
	w.Header().Set(HeaderReselect, selector)
}

// SetTrigger sets the HX-Trigger response header, triggering the named
// client-side events as soon as the response is received. For events that
// carry a detail payload use [SetTriggerJSON]. Calling it with no events is a
// no-op (it never sets an empty header).
func SetTrigger(w http.ResponseWriter, events ...string) {
	setTriggerNames(w, HeaderTrigger, events)
}

// SetTriggerAfterSettle sets the HX-Trigger-After-Settle response header,
// triggering the named client-side events after the settle step.
// For events that carry a detail payload use [SetTriggerAfterSettleJSON].
// Calling it with no events is a no-op.
func SetTriggerAfterSettle(w http.ResponseWriter, events ...string) {
	setTriggerNames(w, HeaderTriggerAfterSettle, events)
}

// SetTriggerAfterSwap sets the HX-Trigger-After-Swap response header,
// triggering the named client-side events after the swap step.
// For events that carry a detail payload use [SetTriggerAfterSwapJSON].
// Calling it with no events is a no-op.
func SetTriggerAfterSwap(w http.ResponseWriter, events ...string) {
	setTriggerNames(w, HeaderTriggerAfterSwap, events)
}

func setTriggerNames(w http.ResponseWriter, header string, events []string) {
	if len(events) == 0 {
		return
	}
	w.Header().Set(header, strings.Join(events, ", "))
}

// SetTriggerJSON sets the HX-Trigger response header to a JSON object mapping
// each event name to the detail passed to the client-side event.
func SetTriggerJSON(w http.ResponseWriter, events map[string]any) error {
	return setTriggerJSON(w, HeaderTrigger, events)
}

// SetTriggerAfterSettleJSON sets the HX-Trigger-After-Settle response header to
// a JSON object mapping each event name to its client-side event detail.
func SetTriggerAfterSettleJSON(w http.ResponseWriter, events map[string]any) error {
	return setTriggerJSON(w, HeaderTriggerAfterSettle, events)
}

// SetTriggerAfterSwapJSON sets the HX-Trigger-After-Swap response header to a
// JSON object mapping each event name to its client-side event detail.
func SetTriggerAfterSwapJSON(w http.ResponseWriter, events map[string]any) error {
	return setTriggerJSON(w, HeaderTriggerAfterSwap, events)
}

func setTriggerJSON(w http.ResponseWriter, header string, events map[string]any) error {
	if len(events) == 0 {
		return nil
	}
	value, err := json.Marshal(events)
	if err != nil {
		return errs.Errorf("hx: marshal %s events: %w", header, err)
	}
	w.Header().Set(header, string(value))
	return nil
}
