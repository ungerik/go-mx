package hx

import "github.com/ungerik/go-mx/html"

var (
	// ScriptFromCDN is a <script> element that loads the minified htmx
	// 2.0.11 library from the jsDelivr CDN with a Subresource Integrity hash
	// and anonymous crossorigin. Place it in the document <head> to enable
	// htmx on the page.
	ScriptFromCDN = html.Script(
		html.Src("https://cdn.jsdelivr.net/npm/htmx.org@2.0.11/dist/htmx.min.js"),
		html.Integrity("sha384-2OatzQy1H+Zd/IIrjr1TcuDGqLXeHhbooAyJY1KdQMKnr4LZ22k31GBLdYKHmVjg"),
		html.CrossOriginAnonymous,
	)

	// ScriptDebugFromCDN is a <script> element that loads the unminified
	// htmx 2.0.11 library from the jsDelivr CDN with a Subresource Integrity
	// hash and anonymous crossorigin. Use it instead of [ScriptFromCDN]
	// during development to get readable stack traces and logging.
	ScriptDebugFromCDN = html.Script(
		html.Src("https://cdn.jsdelivr.net/npm/htmx.org@2.0.11/dist/htmx.js"),
		html.Integrity("sha384-gmJEF2eAKY4e+FDN+qtKIivWyb6ANwDB7JUdUybKgQspPKyEX/pIRZG/0uaRoW2C"),
		html.CrossOriginAnonymous,
	)

	// ScriptSSEFromCDN is a <script> element that loads the htmx SSE
	// extension 2.2.4 from the jsDelivr CDN with a Subresource Integrity hash
	// and anonymous crossorigin. htmx 2.0 moved SSE out of core, so the
	// [SSEConnect], [SSESwap] and [SSEClose] attributes need this loaded in
	// addition to [ScriptFromCDN], and Ext("sse") on the connecting element.
	ScriptSSEFromCDN = html.Script(
		html.Src("https://cdn.jsdelivr.net/npm/htmx-ext-sse@2.2.4/dist/sse.min.js"),
		html.Integrity("sha384-A986SAtodyH8eg8x8irJnYUk7i9inVQqYigD6qZ9evobksGNIXfeFvDwLSHcp31N"),
		html.CrossOriginAnonymous,
	)
)

// htmx request and response headers are available as the Header* constants
// (see headers.go), with reader helpers like [IsRequest] / [IsBoosted] and
// response-header setters like [SetRedirect], [SetReswap], [SetTrigger], etc.
// A handler returning no content can use http.StatusNoContent (204).
