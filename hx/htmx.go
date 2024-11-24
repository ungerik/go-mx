package hx

import "github.com/ungerik/go-mx/html"

var (
	ScriptFromCDN = html.Script(
		html.Src("https://unpkg.com/htmx.org@2.0.3"),
		html.Integrity("sha384-0895/pl2MU10Hqc6jd4RvrthNlDiE9U1tWmX7WRESftEDRosgxNsQG/Ze9YMRzHq"),
		html.CrossOriginAnonymous,
	)

	ScriptDebugFromCDN = html.Script(
		html.Src("https://unpkg.com/htmx.org@2.0.3/dist/htmx.js"),
		html.Integrity("sha384-BBDmZzVt6vjz5YbQqZPtFZW82o8QotoM7RUp5xOxV3nSJ8u2pSdtzFAbGKzTlKtg"),
		html.CrossOriginAnonymous,
	)
)

// 204 - No Content

// htmx supports some htmx-specific response headers:
// HX-Location - allows you to do a client-side redirect that does not do a full page reload
// HX-Push-Url - pushes a new url into the history stack
// HX-Redirect - can be used to do a client-side redirect to a new location
// HX-Refresh - if set to “true” the client-side will do a full refresh of the page
// HX-Replace-Url - replaces the current URL in the location bar
// HX-Reswap - allows you to specify how the response will be swapped. See hx-swap for possible values
// HX-Retarget - a CSS selector that updates the target of the content update to a different element on the page
// HX-Reselect - a CSS selector that allows you to choose which part of the response is used to be swapped in. Overrides an existing hx-select on the triggering element
// HX-Trigger - allows you to trigger client-side events
// HX-Trigger-After-Settle - allows you to trigger client-side events after the settle step
// HX-Trigger-After-Swap - allows you to trigger client-side events after the swap step
