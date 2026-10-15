package mcp

import (
	"bytes"
	"errors"
	"io"
	"math"
	"net"
	"net/http"
	"slices"
	"strings"
)

// DefaultMaxRequestBytes is the request body size [GuardHandler] accepts when
// GuardOptions.MaxRequestBytes is zero or negative. An MCP request is a JSON-RPC message
// of a few hundred bytes, so this leaves plenty of room while keeping a client
// from making the server buffer megabytes.
const DefaultMaxRequestBytes = 64 << 10

// cacheControl is the Cache-Control of every response of [GuardHandler]. It
// keeps the "no-transform" the SDK sets for its event streams, which a proxy
// that compresses or buffers responses must not hold back.
const cacheControl = "no-store, no-transform"

// GuardOptions configure [GuardHandler].
type GuardOptions struct {
	// AllowedOrigins are the origins whose requests are let through, in the
	// form of an Origin header: "https://example.com", or with a port
	// "http://localhost:8080". They are compared case-insensitively, a
	// trailing slash is ignored, and so is the default port of the scheme,
	// which browsers leave out: "https://example.com:443" allows
	// "https://example.com".
	//
	// Browsers send an Origin header with every cross-origin request and with
	// same-origin ones other than GET and HEAD, so a page of the site calling
	// its own endpoint needs the site's origin listed too. GuardHandler only
	// rejects requests and sends no CORS headers: a page on another origin
	// can call the endpoint only if CORS is set up for it by other middleware.
	AllowedOrigins []string

	// AllowedHosts, if not empty, are the only Host header values accepted:
	// a host name like "example.com", which matches it on any port, or a host
	// and port like "localhost:8080", which matches only that port. They are
	// compared case-insensitively. A browser leaves the default port out of
	// the Host header, so list "example.com" rather than "example.com:443".
	//
	// It is the defense against DNS rebinding, where a page points its own
	// domain at an address only its visitors' network can reach. The
	// browser's requests to it are then same-origin, so a GET carries no
	// Origin header to check, but the Host header still names the page's
	// domain. The SDK's StreamableHTTPHandler already rejects a foreign Host
	// on a loopback address; AllowedHosts covers a server on any other one.
	AllowedHosts []string

	// RequireOrigin rejects requests without an Origin header. By default they
	// are let through, because only browsers send one: a desktop client or a
	// server-side agent calling the endpoint does not.
	RequireOrigin bool

	// AllowBatches lets JSON-RPC batches through, which are rejected by
	// default. MCP dropped batching in its 2025-06-18 revision, but the SDK
	// still accepts a batch from a client that sends no protocol version
	// header, and one request carrying hundreds of tool calls defeats every
	// per-call bound and any rate limit that counts requests. Set it only for
	// clients of the 2025-03-26 protocol that batch, and rate-limit tool
	// calls rather than requests then.
	AllowBatches bool

	// MaxRequestBytes is the largest request body accepted,
	// [DefaultMaxRequestBytes] when zero or negative.
	MaxRequestBytes int64
}

// GuardHandler wraps the HTTP handler of an MCP endpoint, typically the SDK's
// StreamableHTTPHandler, with the checks a public endpoint needs before a
// request reaches it:
//
//   - A request with an Origin header that is not one of AllowedOrigins is
//     rejected with 403 Forbidden. The MCP specification requires servers to
//     validate it, because a web page could otherwise make its visitors'
//     browsers call the endpoint.
//   - A request whose Host header is not one of AllowedHosts, if there are
//     any, is rejected with 403 Forbidden, which keeps a page from reaching
//     an endpoint in its visitors' network by DNS rebinding.
//   - A request body larger than MaxRequestBytes is rejected with 413 Content
//     Too Large before the handler reads any of it.
//   - A JSON-RPC batch, a body that is a JSON array, is rejected with 400 Bad
//     Request unless AllowBatches is set.
//   - Every response is marked "Cache-Control: no-store, no-transform",
//     because tool results are answers to one call and not documents for a
//     cache to keep. It is set before the handler runs and again when the
//     response is written, so it also replaces the "no-cache" the SDK sets on
//     its own responses.
//
// Rate limiting is not done here; it belongs in the caller's middleware,
// which knows its clients.
func GuardHandler(next http.Handler, opts GuardOptions) http.Handler {
	maxBytes := opts.MaxRequestBytes
	if maxBytes <= 0 {
		maxBytes = DefaultMaxRequestBytes
	}
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		// Set up front for a response the handler never writes itself, and
		// again by the writer over whatever the handler set.
		response.Header().Set("Cache-Control", cacheControl)
		response = wrapNoStore(response)

		if len(opts.AllowedHosts) > 0 && !hostAllowed(opts.AllowedHosts, request.Host) {
			http.Error(response, "Host not allowed", http.StatusForbidden)
			return
		}
		origin := request.Header.Get("Origin")
		switch {
		case origin == "" && opts.RequireOrigin:
			http.Error(response, "Origin header required", http.StatusForbidden)
			return
		case origin != "" && !slices.ContainsFunc(opts.AllowedOrigins, func(allowed string) bool {
			return strings.EqualFold(normalizeOrigin(allowed), normalizeOrigin(origin))
		}):
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
			// A lower limit of middleware in front of the guard is a body
			// too large as well, not a broken one.
			if maxBytesErr := new(http.MaxBytesError); errors.As(err, &maxBytesErr) {
				http.Error(response, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)
				return
			}
			if err != nil {
				http.Error(response, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			if int64(len(body)) > maxBytes {
				http.Error(response, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)
				return
			}
			if !opts.AllowBatches && isJSONArray(body) {
				http.Error(response, "JSON-RPC batches are not supported", http.StatusBadRequest)
				return
			}
			request.Body = io.NopCloser(bytes.NewReader(body))
		}
		next.ServeHTTP(response, request)
	})
}

// isJSONArray reports whether body is a JSON array, the form of a JSON-RPC
// batch: its first byte that is not JSON whitespace is a '['.
func isJSONArray(body []byte) bool {
	body = bytes.TrimLeft(body, " \t\r\n")
	return len(body) > 0 && body[0] == '['
}

// wrapNoStore wraps response in a [noStoreWriter], one that can flush only if
// response can: a handler probing for streaming support, like
// mx.NewSSEResponse, has to see the truth to answer a 500 while it still can.
func wrapNoStore(response http.ResponseWriter) http.ResponseWriter {
	w := &noStoreWriter{ResponseWriter: response}
	if canFlush(response) {
		return &noStoreFlusher{w}
	}
	return w
}

// noStoreWriter sets [cacheControl] on the response the moment its header is
// written, after the wrapped handler has set headers of its own.
type noStoreWriter struct {
	http.ResponseWriter
	wroteHeader bool
}

func (w *noStoreWriter) WriteHeader(statusCode int) {
	// An informational response like 103 Early Hints is not the response:
	// the header is written again with the final status.
	if !w.wroteHeader && (statusCode >= http.StatusOK || statusCode == http.StatusSwitchingProtocols) {
		w.wroteHeader = true
		w.Header().Set("Cache-Control", cacheControl)
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *noStoreWriter) Write(data []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(data)
}

// Unwrap returns the wrapped writer, so an http.ResponseController reaches
// its other features, like write deadlines.
func (w *noStoreWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// noStoreFlusher is a [noStoreWriter] around a writer that can flush.
type noStoreFlusher struct {
	*noStoreWriter
}

// FlushError flushes a streamed response, the method http.ResponseController
// prefers over http.Flusher. Flushing writes the header too, so it passes
// through WriteHeader first.
func (w *noStoreFlusher) FlushError() error {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return http.NewResponseController(w.ResponseWriter).Flush()
}

// Flush implements http.Flusher with [noStoreFlusher.FlushError].
func (w *noStoreFlusher) Flush() {
	_ = w.FlushError()
}

// canFlush reports whether w, or a writer it wraps, can flush, walking the
// Unwrap chain the way http.ResponseController.Flush does. It mirrors canFlush
// in the root mx package, which is not exported.
func canFlush(w http.ResponseWriter) bool {
	for {
		switch t := w.(type) {
		case interface{ FlushError() error }, http.Flusher:
			return true
		case interface{ Unwrap() http.ResponseWriter }:
			w = t.Unwrap()
		default:
			return false
		}
	}
}

// normalizeOrigin returns origin without a trailing slash and without the
// default port of its scheme, the form a browser sends it in.
func normalizeOrigin(origin string) string {
	origin = strings.TrimSuffix(origin, "/")
	switch scheme, _, _ := strings.Cut(strings.ToLower(origin), "://"); scheme {
	case "https":
		return strings.TrimSuffix(origin, ":443")
	case "http":
		return strings.TrimSuffix(origin, ":80")
	}
	return origin
}

// hostAllowed reports whether the Host header host is one of allowed: equal to
// an entry, or on any port of an entry without one.
func hostAllowed(allowed []string, host string) bool {
	hostname := host
	if h, _, err := net.SplitHostPort(host); err == nil {
		hostname = h
	}
	return slices.ContainsFunc(allowed, func(a string) bool {
		// An IPv6 entry is written "[::1]", but SplitHostPort drops the
		// brackets of the host it splits off the port.
		return strings.EqualFold(a, host) || strings.EqualFold(strings.TrimSuffix(strings.TrimPrefix(a, "["), "]"), hostname)
	})
}
