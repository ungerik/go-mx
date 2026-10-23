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
