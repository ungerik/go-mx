package pdf

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ungerik/go-mx"
)

// A document that fails to render answers a generic 500, and the error goes
// to mx.ReportError with the request's context.
func TestDocument_ServeHTTPReportsTheError(t *testing.T) {
	prev := mx.ReportError
	t.Cleanup(func() { mx.ReportError = prev })
	type requestKey struct{}
	var reported []error
	mx.ReportError = func(ctx context.Context, err error) {
		if ctx.Value(requestKey{}) != "request-1" {
			t.Error("ReportError did not get the request context")
		}
		reported = append(reported, err)
	}
	renderErr := errors.New("invoice 7: pq: connection refused")
	doc := &Document{Body: ComponentFunc(func(context.Context, *Renderer) error { return renderErr })}
	req := httptest.NewRequest(http.MethodGet, "/invoice.pdf", nil)
	req = req.WithContext(context.WithValue(req.Context(), requestKey{}, "request-1"))
	rec := httptest.NewRecorder()
	doc.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status %d, want 500", rec.Code)
	}
	if len(reported) != 1 || !errors.Is(reported[0], renderErr) {
		t.Errorf("reported %v, want the render error once", reported)
	}
}
