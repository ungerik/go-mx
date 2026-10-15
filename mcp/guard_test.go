package mcp_test

import (
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/require"

	mxmcp "github.com/ungerik/go-mx/mcp"
)

// echoHandler responds with the request body it reads.
var echoHandler = http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	response.Write(body)
})

func serveGuarded(opts mxmcp.GuardOptions, request *http.Request) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	mxmcp.GuardHandler(echoHandler, opts).ServeHTTP(response, request)
	return response
}

func TestGuardHandlerOrigin(t *testing.T) {
	opts := mxmcp.GuardOptions{AllowedOrigins: []string{"https://example.com", "http://localhost:8080"}}

	tests := []struct {
		name          string
		origin        string
		requireOrigin bool
		wantStatus    int
	}{
		{name: "allowed", origin: "https://example.com", wantStatus: http.StatusOK},
		{name: "allowed with port", origin: "http://localhost:8080", wantStatus: http.StatusOK},
		{name: "case-insensitive", origin: "HTTPS://EXAMPLE.COM", wantStatus: http.StatusOK},
		{name: "other origin", origin: "https://evil.example", wantStatus: http.StatusForbidden},
		{name: "other port", origin: "http://localhost:9090", wantStatus: http.StatusForbidden},
		{name: "other scheme", origin: "http://example.com", wantStatus: http.StatusForbidden},
		// A sandboxed iframe or a file: page sends "null".
		{name: "null", origin: "null", wantStatus: http.StatusForbidden},
		// Desktop clients and server-side agents send no Origin.
		{name: "missing", origin: "", wantStatus: http.StatusOK},
		{name: "missing but required", origin: "", requireOrigin: true, wantStatus: http.StatusForbidden},
		{name: "allowed and required", origin: "https://example.com", requireOrigin: true, wantStatus: http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(`{"jsonrpc":"2.0"}`))
			if tt.origin != "" {
				request.Header.Set("Origin", tt.origin)
			}
			opts := opts
			opts.RequireOrigin = tt.requireOrigin
			response := serveGuarded(opts, request)
			require.Equal(t, tt.wantStatus, response.Code)
			if tt.wantStatus == http.StatusOK {
				require.Equal(t, `{"jsonrpc":"2.0"}`, response.Body.String())
			}
		})
	}

	t.Run("no allowed origins", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/mcp", nil)
		request.Header.Set("Origin", "https://example.com")
		require.Equal(t, http.StatusForbidden, serveGuarded(mxmcp.GuardOptions{}, request).Code)
	})

	t.Run("checked for GET too", func(t *testing.T) {
		// A GET opens the server's event stream, which is as worth
		// protecting as a POST.
		request := httptest.NewRequest(http.MethodGet, "/mcp", nil)
		request.Header.Set("Origin", "https://evil.example")
		require.Equal(t, http.StatusForbidden, serveGuarded(opts, request).Code)
	})
}

func TestGuardHandlerBodySize(t *testing.T) {
	t.Run("default limit", func(t *testing.T) {
		atLimit := strings.Repeat("x", mxmcp.DefaultMaxRequestBytes)
		response := serveGuarded(mxmcp.GuardOptions{}, httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(atLimit)))
		require.Equal(t, http.StatusOK, response.Code)
		require.Equal(t, atLimit, response.Body.String())

		response = serveGuarded(mxmcp.GuardOptions{}, httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(atLimit+"x")))
		require.Equal(t, http.StatusRequestEntityTooLarge, response.Code)
	})

	t.Run("content length", func(t *testing.T) {
		// Rejected on the declared length, before reading anything.
		request := httptest.NewRequest(http.MethodPost, "/mcp", iotest.ErrReader(io.ErrUnexpectedEOF))
		request.ContentLength = 11
		response := serveGuarded(mxmcp.GuardOptions{MaxRequestBytes: 10}, request)
		require.Equal(t, http.StatusRequestEntityTooLarge, response.Code)
	})

	t.Run("chunked", func(t *testing.T) {
		// No length to check up front, as with chunked encoding or HTTP/2.
		request := httptest.NewRequest(http.MethodPost, "/mcp", io.MultiReader(strings.NewReader("0123456789"), strings.NewReader("a")))
		request.ContentLength = -1
		response := serveGuarded(mxmcp.GuardOptions{MaxRequestBytes: 10}, request)
		require.Equal(t, http.StatusRequestEntityTooLarge, response.Code)
	})

	t.Run("read error", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/mcp", iotest.ErrReader(io.ErrUnexpectedEOF))
		request.ContentLength = -1
		response := serveGuarded(mxmcp.GuardOptions{}, request)
		require.Equal(t, http.StatusBadRequest, response.Code)
	})

	t.Run("no limit", func(t *testing.T) {
		// Regression: the limit plus the one byte that tells a body at the
		// limit from a larger one overflowed, and every body arrived empty.
		request := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader("body"))
		request.ContentLength = -1
		response := serveGuarded(mxmcp.GuardOptions{MaxRequestBytes: math.MaxInt64}, request)
		require.Equal(t, http.StatusOK, response.Code)
		require.Equal(t, "body", response.Body.String())
	})

	t.Run("no body", func(t *testing.T) {
		response := serveGuarded(mxmcp.GuardOptions{}, httptest.NewRequest(http.MethodGet, "/mcp", nil))
		require.Equal(t, http.StatusOK, response.Code)
	})
}

func TestGuardHandlerNoStore(t *testing.T) {
	allowed := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	require.Equal(t, "no-store", serveGuarded(mxmcp.GuardOptions{}, allowed).Header().Get("Cache-Control"))

	// Rejections are not to be cached either: the next request may be fine.
	rejected := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	rejected.Header.Set("Origin", "https://evil.example")
	response := serveGuarded(mxmcp.GuardOptions{}, rejected)
	require.Equal(t, http.StatusForbidden, response.Code)
	require.Equal(t, "no-store", response.Header().Get("Cache-Control"))
}
