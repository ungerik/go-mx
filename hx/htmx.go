package hx

import "github.com/ungerik/go-mx/html"

var (
	// ScriptFromCDN is a <script> element that loads the minified htmx
	// 2.0.10 library from the unpkg CDN with a Subresource Integrity hash
	// and anonymous crossorigin. Place it in the document <head> to enable
	// htmx on the page.
	ScriptFromCDN = html.Script(
		html.Src("https://unpkg.com/htmx.org@2.0.10"),
		html.Integrity("sha384-H5SrcfygHmAuTDZphMHqBJLc3FhssKjG7w/CeCpFReSfwBWDTKpkzPP8c+cLsK+V"),
		html.CrossOriginAnonymous,
	)

	// ScriptDebugFromCDN is a <script> element that loads the unminified
	// htmx 2.0.10 library from the unpkg CDN with a Subresource Integrity
	// hash and anonymous crossorigin. Use it instead of [ScriptFromCDN]
	// during development to get readable stack traces and logging.
	ScriptDebugFromCDN = html.Script(
		html.Src("https://unpkg.com/htmx.org@2.0.10/dist/htmx.js"),
		html.Integrity("sha384-Q+Dky3iHVJOr6wUjQ4ulh6uQ76an/t+ak1+PjMVaxRjbZamFLAG+u9InkfjbsEQf"),
		html.CrossOriginAnonymous,
	)
)

// htmx request and response headers are available as the Header* constants
// (see headers.go), with reader helpers like [IsRequest] / [IsBoosted] and
// response-header setters like [SetRedirect], [SetReswap], [SetTrigger], etc.
// A handler returning no content can use http.StatusNoContent (204).
