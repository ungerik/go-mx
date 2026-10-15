package mcp

import (
	"bytes"
	"io"
	"math"
	"net/http"
	"slices"
	"strings"
)

// DefaultMaxRequestBytes is the request body size [GuardHandler] accepts when
// GuardOptions.MaxRequestBytes is zero. An MCP request is a JSON-RPC message
// of a few hundred bytes, so this leaves plenty of room while keeping a client
// from making the server buffer megabytes.
const DefaultMaxRequestBytes = 64 << 10

// GuardOptions configure [GuardHandler].
type GuardOptions struct {
	// AllowedOrigins are the origins a browser may call the endpoint from, in
	// the form of an Origin header: "https://example.com", or with a port
	// "http://localhost:8080". They are compared case-insensitively.
	AllowedOrigins []string

	// RequireOrigin rejects requests without an Origin header. By default they
	// are let through, because only browsers send one: a desktop client or a
	// server-side agent calling the endpoint does not.
	RequireOrigin bool

	// MaxRequestBytes is the largest request body accepted,
	// [DefaultMaxRequestBytes] when zero.
	MaxRequestBytes int64
}

// GuardHandler wraps the HTTP handler of an MCP endpoint, typically the SDK's
// StreamableHTTPHandler, with the checks a public endpoint needs before a
// request reaches it:
//
//   - A request with an Origin header that is not one of AllowedOrigins is
//     rejected with 403 Forbidden. The MCP specification requires servers to
//     validate it, because a web page could otherwise make its visitors'
//     browsers call the endpoint, including one only reachable from their
//     network (DNS rebinding).
//   - A request body larger than MaxRequestBytes is rejected with 413 Content
//     Too Large before the handler reads any of it.
//   - Every response is marked "Cache-Control: no-store", because tool results
//     are answers to one call and not documents for a cache to keep.
//
// Rate limiting is not done here; it belongs in the caller's middleware,
// which knows its clients.
func GuardHandler(next http.Handler, opts GuardOptions) http.Handler {
	maxBytes := opts.MaxRequestBytes
	if maxBytes <= 0 {
		maxBytes = DefaultMaxRequestBytes
	}
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Cache-Control", "no-store")

		origin := request.Header.Get("Origin")
		switch {
		case origin == "" && opts.RequireOrigin:
			http.Error(response, "Origin header required", http.StatusForbidden)
			return
		case origin != "" && !slices.ContainsFunc(opts.AllowedOrigins, func(allowed string) bool { return strings.EqualFold(allowed, origin) }):
			http.Error(response, "Origin not allowed", http.StatusForbidden)
			return
		}

		if request.ContentLength > maxBytes {
			http.Error(response, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)
			return
		}
		if request.Body != nil && request.Body != http.NoBody {
			// A chunked or HTTP/2 body has no length to check up front, so
			// it is read here, one byte past the limit to tell a body of
			// exactly the limit from a larger one. Reading it whole is
			// what lets an oversized body get a 413 rather than whatever
			// error the handler makes of a read failing half way.
			limit := maxBytes
			if limit < math.MaxInt64 {
				limit++
			}
			body, err := io.ReadAll(io.LimitReader(request.Body, limit))
			if err != nil {
				http.Error(response, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			if int64(len(body)) > maxBytes {
				http.Error(response, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)
				return
			}
			request.Body = io.NopCloser(bytes.NewReader(body))
		}
		next.ServeHTTP(response, request)
	})
}
