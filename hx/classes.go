package hx

// CSS classes htmx applies to elements during the request lifecycle.
// Use them with html.Class(...) to style requests, indicators and swaps.
// See https://htmx.org/reference/#classes
const (
	ClassAdded     = "htmx-added"     // applied to new content before it is swapped, removed after settling
	ClassIndicator = "htmx-indicator" // a request indicator, shown while a request is in flight
	ClassRequest   = "htmx-request"   // applied to an element or its hx-indicator while a request is in flight
	ClassSettling  = "htmx-settling"  // applied to target content between the swap and settle steps
	ClassSwapping  = "htmx-swapping"  // applied to target content between the request and swap steps
)
