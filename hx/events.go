package hx

// htmx event names, for use with [On], [OnHTMX], [Trigger] and hx-on attributes.
// Each constant holds the full event name including the htmx: prefix, so it can
// be used directly: Trigger(hx.EventLoad) or On(hx.EventAfterSwap, "...").
// See https://htmx.org/events/
const (
	// Request lifecycle

	EventConfigRequest = "htmx:configRequest" // before a request, to inspect or modify it
	EventBeforeRequest = "htmx:beforeRequest" // before an AJAX request is issued
	EventBeforeSend    = "htmx:beforeSend"    // just before an AJAX request is sent
	EventAfterRequest  = "htmx:afterRequest"  // after an AJAX request has finished
	EventBeforeOnLoad  = "htmx:beforeOnLoad"  // before any response processing
	EventAfterOnLoad   = "htmx:afterOnLoad"   // after a successful AJAX request, before swapping
	EventResponseError = "htmx:responseError" // when an HTTP error response is received
	EventSendError     = "htmx:sendError"     // when a network error prevents a request
	EventSendAbort     = "htmx:sendAbort"     // when a request is aborted
	EventTimeout       = "htmx:timeout"       // when a request times out
	EventAbort         = "htmx:abort"         // send to an element to abort its in-flight request
	EventConfirm       = "htmx:confirm"       // to customize or cancel the hx-confirm step

	// Swap lifecycle

	EventBeforeSwap       = "htmx:beforeSwap"       // before content is swapped in
	EventAfterSwap        = "htmx:afterSwap"        // after content is swapped in
	EventBeforeTransition = "htmx:beforeTransition" // before a View Transition wrapped swap
	EventAfterSettle      = "htmx:afterSettle"      // after the settle step
	EventSwapError        = "htmx:swapError"        // when an error occurs during the swap step
	EventOOBBeforeSwap    = "htmx:oobBeforeSwap"    // before an out-of-band element is swapped
	EventOOBAfterSwap     = "htmx:oobAfterSwap"     // after an out-of-band element is swapped
	EventOOBErrorNoTarget = "htmx:oobErrorNoTarget" // when an OOB element has no matching target

	// Node processing

	EventLoad                 = "htmx:load"                 // when a new node is loaded into the DOM by htmx
	EventBeforeProcessNode    = "htmx:beforeProcessNode"    // before htmx processes a node
	EventAfterProcessNode     = "htmx:afterProcessNode"     // after htmx processes a node
	EventBeforeCleanupElement = "htmx:beforeCleanupElement" // before htmx cleans up an element being removed
	EventOnLoadError          = "htmx:onLoadError"          // when an exception occurs during load handling
	EventTrigger              = "htmx:trigger"              // when an element is triggered with no AJAX request

	// History

	EventBeforeHistorySave         = "htmx:beforeHistorySave"         // before the page state is snapshotted into history
	EventBeforeHistoryUpdate       = "htmx:beforeHistoryUpdate"       // before the history is updated
	EventPushedIntoHistory         = "htmx:pushedIntoHistory"         // after a URL is pushed into history
	EventReplacedInHistory         = "htmx:replacedInHistory"         // after a URL is replaced in history
	EventHistoryRestore            = "htmx:historyRestore"            // when the page is restored from history
	EventHistoryCacheHit           = "htmx:historyCacheHit"           // on a cache hit during history restoration
	EventHistoryCacheMiss          = "htmx:historyCacheMiss"          // on a cache miss during history restoration
	EventHistoryCacheMissLoad      = "htmx:historyCacheMissLoad"      // on a successful remote retrieval after a cache miss
	EventHistoryCacheMissLoadError = "htmx:historyCacheMissLoadError" // on a bad remote retrieval after a cache miss
	EventHistoryCacheError         = "htmx:historyCacheError"         // when an error occurs accessing the history cache

	// Prompt, URL and validation

	EventPrompt             = "htmx:prompt"              // after an hx-prompt is shown
	EventValidateURL        = "htmx:validateUrl"         // to validate the URL of a request (can be canceled)
	EventValidationValidate = "htmx:validation:validate" // before an element is validated
	EventValidationFailed   = "htmx:validation:failed"   // when an element fails validation
	EventValidationHalted   = "htmx:validation:halted"   // when a request is halted due to validation errors

	// XHR progress

	EventXHRAbort     = "htmx:xhr:abort"     // when an AJAX request aborts
	EventXHRLoadStart = "htmx:xhr:loadstart" // when an AJAX request starts
	EventXHRLoadEnd   = "htmx:xhr:loadend"   // when an AJAX request finishes
	EventXHRProgress  = "htmx:xhr:progress"  // periodically during an AJAX request

	// SSE extension (htmx 2.0 moved SSE out of core into the sse extension)

	EventNoSSESourceError = "htmx:noSSESourceError" // when an element references a missing SSE source
	EventSSEError         = "htmx:sseError"         // when an error occurs in an SSE source
)
