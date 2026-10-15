package mcp_test

import (
	"context"
	"errors"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
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

func TestGuardHandlerOriginTrailingSlash(t *testing.T) {
	// An Origin header never ends in a slash, so an allowed origin written
	// with one would otherwise never match anything.
	request := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	request.Header.Set("Origin", "https://example.com")
	response := serveGuarded(mxmcp.GuardOptions{AllowedOrigins: []string{"https://example.com/"}}, request)
	require.Equal(t, http.StatusOK, response.Code)
}

func TestGuardHandlerOriginDefaultPort(t *testing.T) {
	// Browsers leave the default port out of the Origin header.
	for allowed, origin := range map[string]string{
		"https://example.com:443": "https://example.com",
		"http://example.com:80":   "http://example.com",
		"https://example.com":     "https://example.com:443",
	} {
		request := httptest.NewRequest(http.MethodPost, "/mcp", nil)
		request.Header.Set("Origin", origin)
		response := serveGuarded(mxmcp.GuardOptions{AllowedOrigins: []string{allowed}}, request)
		require.Equal(t, http.StatusOK, response.Code, "allowed %q, origin %q", allowed, origin)
	}
	// Only the default port of the scheme: 443 is no default for http.
	request := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	request.Header.Set("Origin", "http://example.com")
	response := serveGuarded(mxmcp.GuardOptions{AllowedOrigins: []string{"http://example.com:443"}}, request)
	require.Equal(t, http.StatusForbidden, response.Code)
}

func TestGuardHandlerHost(t *testing.T) {
	opts := mxmcp.GuardOptions{AllowedHosts: []string{"example.com", "localhost:8080", "[::1]"}}

	tests := []struct {
		host       string
		wantStatus int
	}{
		{host: "example.com", wantStatus: http.StatusOK},
		{host: "EXAMPLE.com", wantStatus: http.StatusOK},
		// A host name without a port is allowed on every port.
		{host: "example.com:8443", wantStatus: http.StatusOK},
		{host: "localhost:8080", wantStatus: http.StatusOK},
		// A host with a port only on that one.
		{host: "localhost:9090", wantStatus: http.StatusForbidden},
		{host: "localhost", wantStatus: http.StatusForbidden},
		{host: "[::1]", wantStatus: http.StatusOK},
		{host: "[::1]:8080", wantStatus: http.StatusOK},
		// DNS rebinding: the attacker's domain pointed at the server.
		{host: "evil.example", wantStatus: http.StatusForbidden},
		{host: "example.com.evil.example", wantStatus: http.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			// A rebinding page's GET is same-origin and carries no Origin.
			request := httptest.NewRequest(http.MethodGet, "/mcp", nil)
			request.Host = tt.host
			require.Equal(t, tt.wantStatus, serveGuarded(opts, request).Code)
		})
	}

	t.Run("no allowed hosts", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/mcp", nil)
		request.Host = "evil.example"
		require.Equal(t, http.StatusOK, serveGuarded(mxmcp.GuardOptions{}, request).Code)
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
	require.Equal(t, "no-store, no-transform", serveGuarded(mxmcp.GuardOptions{}, allowed).Header().Get("Cache-Control"))

	// Rejections are not to be cached either: the next request may be fine.
	rejected := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	rejected.Header.Set("Origin", "https://evil.example")
	response := serveGuarded(mxmcp.GuardOptions{}, rejected)
	require.Equal(t, http.StatusForbidden, response.Code)
	require.Equal(t, "no-store, no-transform", response.Header().Get("Cache-Control"))
}

func TestGuardHandlerWithStreamableHTTPHandler(t *testing.T) {
	// The documented wiring end to end: the page tools served statelessly by
	// the SDK's handler behind the guard, over a real HTTP connection.
	server := sdk.NewServer(&sdk.Implementation{Name: "test-server", Version: "v0.0.1"}, nil)
	require.NoError(t, mxmcp.AddPageTools(server, testPages(), testSite(), mxmcp.PageToolsOptions{}))
	handler := sdk.NewStreamableHTTPHandler(
		func(*http.Request) *sdk.Server { return server },
		&sdk.StreamableHTTPOptions{Stateless: true, JSONResponse: true},
	)
	httpServer := httptest.NewServer(mxmcp.GuardHandler(handler, mxmcp.GuardOptions{
		AllowedOrigins: []string{"https://example.com"},
	}))
	defer httpServer.Close()

	session, err := sdk.NewClient(&sdk.Implementation{Name: "test-client", Version: "v0.0.1"}, nil).
		Connect(context.Background(), &sdk.StreamableClientTransport{Endpoint: httpServer.URL}, nil)
	require.NoError(t, err)
	defer session.Close()
	result, err := session.CallTool(context.Background(), &sdk.CallToolParams{Name: "get_page", Arguments: map[string]any{"path": "/en/faq"}})
	require.NoError(t, err)
	require.False(t, result.IsError)
	require.Contains(t, text(t, result), "## Pricing")

	// Regression: the SDK sets "no-cache, no-transform" on its responses,
	// which replaced the no-store the guard had set before calling it.
	request, err := http.NewRequest(http.MethodPost, httpServer.URL, strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}`))
	require.NoError(t, err)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json, text/event-stream")
	request.Header.Set("Mcp-Protocol-Version", "2025-06-18")
	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	response.Body.Close()
	require.Equal(t, http.StatusOK, response.StatusCode)
	require.Equal(t, "no-store, no-transform", response.Header.Get("Cache-Control"))

	request.Header.Set("Origin", "https://evil.example")
	request.Body = io.NopCloser(strings.NewReader(`{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`))
	response, err = http.DefaultClient.Do(request)
	require.NoError(t, err)
	response.Body.Close()
	require.Equal(t, http.StatusForbidden, response.StatusCode)
}

func TestGuardHandlerFlushSetsNoStore(t *testing.T) {
	// A streamed response is flushed before anything is written, which
	// writes the header too.
	flushing := http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Cache-Control", "no-cache")
		require.NoError(t, http.NewResponseController(response).Flush())
	})
	response := httptest.NewRecorder()
	mxmcp.GuardHandler(flushing, mxmcp.GuardOptions{}).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/mcp", nil))
	require.True(t, response.Flushed)
	require.Equal(t, "no-store, no-transform", response.Header().Get("Cache-Control"))
}

func TestGuardHandlerNoStoreWithoutWrite(t *testing.T) {
	// Regression: a handler writing nothing gets its header written by
	// net/http after it returned, past the wrapper.
	empty := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	response := httptest.NewRecorder()
	mxmcp.GuardHandler(empty, mxmcp.GuardOptions{}).ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/mcp", nil))
	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, "no-store, no-transform", response.Header().Get("Cache-Control"))
}

func TestGuardHandlerNoStoreAfterEarlyHints(t *testing.T) {
	// An informational response is not the response: the final header is
	// written again and has to carry no-store too.
	hinting := http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.WriteHeader(http.StatusEarlyHints)
		response.Header().Set("Cache-Control", "public, max-age=3600")
		response.WriteHeader(http.StatusOK)
	})
	server := httptest.NewServer(mxmcp.GuardHandler(hinting, mxmcp.GuardOptions{}))
	defer server.Close()
	response, err := http.Get(server.URL)
	require.NoError(t, err)
	response.Body.Close()
	require.Equal(t, "no-store, no-transform", response.Header.Get("Cache-Control"))
}

func TestGuardHandlerResponseController(t *testing.T) {
	// The writer the handler gets still reaches the connection's features.
	deadlineErr := make(chan error, 1)
	handler := http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		deadlineErr <- http.NewResponseController(response).SetWriteDeadline(time.Now().Add(time.Minute))
	})
	server := httptest.NewServer(mxmcp.GuardHandler(handler, mxmcp.GuardOptions{}))
	defer server.Close()
	response, err := http.Get(server.URL)
	require.NoError(t, err)
	response.Body.Close()
	require.NoError(t, <-deadlineErr)
}

// plainWriter is a ResponseWriter that cannot flush. The recorder is a named
// field, because embedding it would promote its Flush method.
type plainWriter struct{ recorder *httptest.ResponseRecorder }

func (w plainWriter) Header() http.Header         { return w.recorder.Header() }
func (w plainWriter) Write(b []byte) (int, error) { return w.recorder.Write(b) }
func (w plainWriter) WriteHeader(statusCode int)  { w.recorder.WriteHeader(statusCode) }

func TestGuardHandlerKeepsFlushability(t *testing.T) {
	// A handler probing for streaming support, like mx.NewSSEResponse, must
	// learn that it cannot stream while a 500 can still be sent, so the
	// guard's writer flushes only where the wrapped one can.
	var flushErr error
	probing := http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		_, flushable := response.(http.Flusher)
		require.False(t, flushable)
		flushErr = http.NewResponseController(response).Flush()
		http.Error(response, "cannot stream", http.StatusInternalServerError)
	})
	recorder := httptest.NewRecorder()
	mxmcp.GuardHandler(probing, mxmcp.GuardOptions{}).ServeHTTP(plainWriter{recorder}, httptest.NewRequest(http.MethodGet, "/mcp", nil))
	require.True(t, errors.Is(flushErr, http.ErrNotSupported), "flush error %v", flushErr)
	require.Equal(t, http.StatusInternalServerError, recorder.Code)
	require.Equal(t, "no-store, no-transform", recorder.Header().Get("Cache-Control"))
}

func TestGuardHandlerBatches(t *testing.T) {
	// Regression: one POST with hundreds of tool calls in a JSON-RPC batch
	// went past every per-call bound and any rate limit counting requests.
	// Any JSON whitespace may come first, a carriage return included.
	batch := " \r\n\t[" + strings.Repeat(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"list_pages"}},`, 3) + `{"jsonrpc":"2.0","id":2,"method":"ping"}]`
	response := serveGuarded(mxmcp.GuardOptions{}, httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(batch)))
	require.Equal(t, http.StatusBadRequest, response.Code)
	require.Contains(t, response.Body.String(), "batches are not supported")

	// A single call is no batch, and AllowBatches lets a batch through.
	single := `{"jsonrpc":"2.0","id":1,"method":"ping"}`
	response = serveGuarded(mxmcp.GuardOptions{}, httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(single)))
	require.Equal(t, http.StatusOK, response.Code)
	response = serveGuarded(mxmcp.GuardOptions{AllowBatches: true}, httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(batch)))
	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, batch, response.Body.String())
}

func TestGuardHandlerBatchesBeforeStreamableHTTPHandler(t *testing.T) {
	// The SDK itself still accepts a batch from a client that sends no
	// protocol version header; behind the guard it never sees one.
	server := sdk.NewServer(&sdk.Implementation{Name: "test-server", Version: "v0.0.1"}, nil)
	require.NoError(t, mxmcp.AddPageTools(server, testPages(), testSite(), mxmcp.PageToolsOptions{}))
	handler := sdk.NewStreamableHTTPHandler(
		func(*http.Request) *sdk.Server { return server },
		&sdk.StreamableHTTPOptions{Stateless: true, JSONResponse: true},
	)
	httpServer := httptest.NewServer(mxmcp.GuardHandler(handler, mxmcp.GuardOptions{}))
	defer httpServer.Close()

	batch := `[{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"list_pages","arguments":{}}},` +
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_pages","arguments":{}}}]`
	request, err := http.NewRequest(http.MethodPost, httpServer.URL, strings.NewReader(batch))
	require.NoError(t, err)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json, text/event-stream")
	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	response.Body.Close()
	require.Equal(t, http.StatusBadRequest, response.StatusCode)
}

// unwrapOnly is middleware's writer that only offers Unwrap.
type unwrapOnly struct{ http.ResponseWriter }

func (w unwrapOnly) Unwrap() http.ResponseWriter { return w.ResponseWriter }

// failingFlusher is a writer whose flush fails, like one to a client gone.
type failingFlusher struct{ *httptest.ResponseRecorder }

func (w failingFlusher) FlushError() error { return errors.New("connection gone") }

func TestGuardHandlerFlushing(t *testing.T) {
	t.Run("through Unwrap", func(t *testing.T) {
		// Middleware in front of the guard that only offers Unwrap must not
		// cost the handler its flushing.
		var flushErr error
		handler := http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			_, flushable := response.(http.Flusher)
			require.True(t, flushable)
			flushErr = http.NewResponseController(response).Flush()
		})
		recorder := httptest.NewRecorder()
		mxmcp.GuardHandler(handler, mxmcp.GuardOptions{}).ServeHTTP(unwrapOnly{recorder}, httptest.NewRequest(http.MethodGet, "/mcp", nil))
		require.NoError(t, flushErr)
		require.True(t, recorder.Flushed)
	})

	t.Run("error passed on", func(t *testing.T) {
		// A stream has to learn that its client is gone.
		var flushErr error
		handler := http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			flushErr = http.NewResponseController(response).Flush()
		})
		mxmcp.GuardHandler(handler, mxmcp.GuardOptions{}).ServeHTTP(failingFlusher{httptest.NewRecorder()}, httptest.NewRequest(http.MethodGet, "/mcp", nil))
		require.EqualError(t, flushErr, "connection gone")
	})

	t.Run("http.Flusher", func(t *testing.T) {
		handler := http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			response.Header().Set("Cache-Control", "no-cache")
			response.(http.Flusher).Flush()
		})
		recorder := httptest.NewRecorder()
		mxmcp.GuardHandler(handler, mxmcp.GuardOptions{}).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/mcp", nil))
		require.True(t, recorder.Flushed)
		require.Equal(t, "no-store, no-transform", recorder.Header().Get("Cache-Control"))
	})
}

func TestGuardHandlerBodyLimitOfOuterMiddleware(t *testing.T) {
	// A lower limit set in front of the guard is a body too large as well.
	outer := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			request.Body = http.MaxBytesReader(response, request.Body, 1024)
			next.ServeHTTP(response, request)
		})
	}
	request := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(strings.Repeat("x", 4096)))
	request.ContentLength = -1
	response := httptest.NewRecorder()
	outer(mxmcp.GuardHandler(echoHandler, mxmcp.GuardOptions{})).ServeHTTP(response, request)
	require.Equal(t, http.StatusRequestEntityTooLarge, response.Code)
}
