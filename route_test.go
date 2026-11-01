package mx

import (
	"errors"
	"net/http"
	"testing"
)

// A route whose component fails to render, or panics, answers a 500 and is
// reported, a panic with the debug stack its error carries.
func TestComponentFuncHandler_ReportsErrorsAndPanics(t *testing.T) {
	rep := recordReports(t)
	renderErr := errors.New("pq: relation \"invoices\" does not exist")
	failing := NewComponentFuncHandler(func(struct{}) Component { return NewErrElement(renderErr) }, nil)
	if rec := serveReportRequest(failing, http.MethodGet, nil); rec.Code != http.StatusInternalServerError {
		t.Errorf("status %d, want 500", rec.Code)
	}
	rep.checkReported(t, true, renderErr)

	rep.reset()
	panicking := NewComponentFuncHandler(func(struct{}) Component { panic("route table not loaded") }, nil)
	if rec := serveReportRequest(panicking, http.MethodGet, nil); rec.Code != http.StatusInternalServerError {
		t.Errorf("status %d, want 500", rec.Code)
	}
	checkReportedPanic(t, rep, "route table not loaded")
}

func TestTypedRoutePath(t *testing.T) {
	page := func(struct{}) Component { return nil }
	tests := []struct {
		pattern string
		values  map[string]any
		want    string
	}{
		{"/about", nil, "/about"},
		{"GET /blog/{slug}", map[string]any{"slug": "hello-world"}, "/blog/hello-world"},
		{"GET /invoices/{id}/lines/{line}", map[string]any{"id": 42, "line": 3}, "/invoices/42/lines/3"},
		// A value is escaped as one segment, so a slash in it cannot add one.
		{"/tags/{tag}", map[string]any{"tag": "a/b c"}, "/tags/a%2Fb%20c"},
		// A remainder wildcard keeps the slashes of its value as separators
		// and escapes each segment on its own.
		{"GET /files/{path...}", map[string]any{"path": "docs/read me.md"}, "/files/docs/read%20me.md"},
		// {$} only anchors the pattern; the trailing slash before it stays.
		{"GET /{$}", nil, "/"},
		{"/blog/{$}", nil, "/blog/"},
		// The host of a host-qualified pattern is not part of the path.
		{"GET example.com/blog/{slug}", map[string]any{"slug": "x"}, "/blog/x"},
	}
	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			got, err := NewTypedRoute(tt.pattern, page, nil).Path(tt.values)
			if err != nil {
				t.Fatalf("Path(%v) error: %v", tt.values, err)
			}
			if got != tt.want {
				t.Errorf("Path(%v) = %q, want %q", tt.values, got, tt.want)
			}
		})
	}
}

// Path must never hand back a URL with a wildcard left in it, which is what
// the stub it replaced did for every pattern with one.
func TestTypedRoutePathErrors(t *testing.T) {
	page := func(struct{}) Component { return nil }
	tests := []struct {
		name    string
		pattern string
		values  map[string]any
	}{
		{"missing value", "GET /blog/{slug}", nil},
		{"missing remainder value", "/files/{path...}", map[string]any{}},
		{"misspelt name", "/blog/{slug}", map[string]any{"slug": "x", "slgu": "y"}},
		{"value for a pattern without wildcards", "/about", map[string]any{"id": 1}},
		// "/users/../edit" would resolve to "/edit", another route.
		{"dot-dot value", "/users/{id}/edit", map[string]any{"id": ".."}},
		{"dot value", "/users/{id}/edit", map[string]any{"id": "."}},
		{"dot-dot part of a remainder value", "/files/{path...}", map[string]any{"path": "a/../../admin"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTypedRoute(tt.pattern, page, nil).Path(tt.values)
			if err == nil {
				t.Errorf("Path(%v) = %q, want an error", tt.values, got)
			}
		})
	}
}
