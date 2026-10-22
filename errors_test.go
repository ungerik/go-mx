package mx

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/domonda/go-errs"
)

func TestRespondNonContextError_HidesTheDetailByDefault(t *testing.T) {
	// The 500 half of the policy [SSEResponse.SendError] now shares: an internal
	// error reaches the client as a status, never as the message that names a
	// host, a query or a file path. Both reporters read the same
	// hiddenErrorText, so the response side needs its own test or the
	// shared helper is only ever checked through the stream.
	rec := httptest.NewRecorder()
	RespondNonContextError(rec, httptest.NewRequest(http.MethodGet, "/", nil), errs.New("connection to 10.0.0.4 refused"))
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	if strings.Contains(rec.Body.String(), "10.0.0.4") {
		t.Errorf("RespondNonContextError leaked the error detail: %q", rec.Body.String())
	}
	if got, want := strings.TrimSpace(rec.Body.String()), "Internal Server Error"; got != want {
		t.Errorf("body = %q, want %q", got, want)
	}
}

func TestRespondNonContextError_RevealsWhenConfigured(t *testing.T) {
	// The development toggle has to reach the 500 as well as the stream,
	// otherwise the two reporters disagree about the same error.
	defer func(prev bool) { RevealInternalServerErrors = prev }(RevealInternalServerErrors)
	RevealInternalServerErrors = true

	rec := httptest.NewRecorder()
	RespondNonContextError(rec, httptest.NewRequest(http.MethodGet, "/", nil), errs.New("connection to 10.0.0.4 refused"))
	if !strings.Contains(rec.Body.String(), "connection to 10.0.0.4 refused") {
		t.Errorf("body = %q, want the error message", rec.Body.String())
	}
}

func TestRespondNonContextError_SilentForContextErrors(t *testing.T) {
	// A canceled context means the client that would read the 500 has
	// disconnected, so there is nobody to tell — and net/http logs a
	// superfluous WriteHeader for the attempt. The wrapped case is the
	// realistic one: a cancellation surfaces from a database driver or an HTTP
	// call with its own message around it, so the check has to unwrap.
	for _, err := range []error{
		context.Canceled,
		context.DeadlineExceeded,
		errs.Errorf("querying the ledger: %w", context.Canceled),
		errs.Errorf("calling the pricing service: %w", context.DeadlineExceeded),
	} {
		rec := httptest.NewRecorder()
		RespondNonContextError(rec, httptest.NewRequest(http.MethodGet, "/", nil), err)
		if rec.Body.Len() != 0 {
			t.Errorf("RespondNonContextError(%v) wrote %q, want nothing", err, rec.Body.String())
		}
		// The recorder starts at 200, so an unchanged code means no status was
		// committed either.
		if rec.Code != http.StatusOK {
			t.Errorf("RespondNonContextError(%v) responded with status %d, want no response at all", err, rec.Code)
		}
	}
}

// ibanError is a domain error that marks itself as fixable by the user with
// its method set alone: it uses no go-mx type. Its Error text carries the
// internal context a log wants, its UserMessage the text the user reads.
type ibanError struct{ iban string }

func (e ibanError) Error() string {
	return fmt.Sprintf("payments: IBAN %q fails its checksum", e.iban)
}

func (e ibanError) UserMessage() string { return "The IBAN is not valid." }

var (
	_ UserMessageError = ibanError{}
	_ UserMessageError = FormMessage("")
)

// requestKey carries a value through a request's context, so a test can
// check that ReportError gets the context of the request.
type requestKey struct{}

// newReportRequest returns a request whose context carries requestKey,
// posting form when it is not nil.
func newReportRequest(method string, form url.Values) *http.Request {
	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}
	r := httptest.NewRequest(method, "/", body)
	if form != nil {
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	return r.WithContext(context.WithValue(r.Context(), requestKey{}, "request-1"))
}

// serveReportRequest serves newReportRequest(method, form) with h.
func serveReportRequest(h http.Handler, method string, form url.Values) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, newReportRequest(method, form))
	return rec
}

// reports records what ReportError receives.
type reports struct {
	mu   sync.Mutex
	errs []error
	ctxs []context.Context
}

// recordReports sets ReportError for the duration of the test and returns
// the record of its calls.
func recordReports(t *testing.T) *reports {
	t.Helper()
	prev := ReportError
	t.Cleanup(func() { ReportError = prev })
	rep := new(reports)
	ReportError = func(ctx context.Context, err error) {
		rep.mu.Lock()
		defer rep.mu.Unlock()
		rep.errs = append(rep.errs, err)
		rep.ctxs = append(rep.ctxs, ctx)
	}
	return rep
}

// got returns the reported errors.
func (r *reports) got() []error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.Clone(r.errs)
}

// texts returns the text of each reported error.
func (r *reports) texts() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	texts := make([]string, len(r.errs))
	for i, err := range r.errs {
		texts[i] = err.Error()
	}
	return texts
}

// reset forgets the calls so far.
func (r *reports) reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.errs, r.ctxs = nil, nil
}

// checkReported fails the test unless ReportError got exactly want, in
// order, each with the context of newReportRequest when fromRequest is set.
func (r *reports) checkReported(t *testing.T, fromRequest bool, want ...error) {
	t.Helper()
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.errs) != len(want) {
		t.Errorf("ReportError got %d errors %q, want %d %q", len(r.errs), r.errs, len(want), want)
		return
	}
	for i, err := range r.errs {
		if !errors.Is(err, want[i]) {
			t.Errorf("report %d = %v, want %v", i, err, want[i])
		}
		if fromRequest && r.ctxs[i].Value(requestKey{}) != "request-1" {
			t.Errorf("report %d did not get the request context", i)
		}
	}
}

// A domain package marks a user-fixable error with a UserMessage method,
// without importing go-mx. Its UserMessage stands alone: the text of an
// error wrapping it, or joined with it, may carry internal context.
func TestUserMessage_RecognizedByMethodSet(t *testing.T) {
	domain := ibanError{"DE00 1234"}
	for _, c := range []struct {
		name string
		err  error
	}{
		{"bare", domain},
		{"wrapped", fmt.Errorf("saving partner 7: %w", domain)},
		{"joined", errors.Join(errors.New("pq: deadlock detected"), domain)},
		{"wrapped in a join", errors.Join(errors.New("pq: deadlock detected"), fmt.Errorf("normalizing: %w", domain))},
	} {
		msg, ok := UserMessage(c.err)
		if !ok || msg != "The IBAN is not valid." {
			t.Errorf("%s: UserMessage = %q, %v; want the domain error's UserMessage", c.name, msg, ok)
		}
	}
	if msg, ok := UserMessage(fmt.Errorf("saving: %w", FormMessage("the plan is full"))); !ok || msg != "the plan is full" {
		t.Errorf("FormMessage: UserMessage = %q, %v", msg, ok)
	}
	for _, err := range []error{nil, errors.New("pq: connection refused"), FieldErrors{"Name": nil}} {
		if msg, ok := UserMessage(err); ok {
			t.Errorf("UserMessage(%v) = %q, want none", err, msg)
		}
	}
}

// Every 500 RespondNonContextError answers drops the error for the client,
// so the application gets it, with the request's context to log it with.
// It does whether or not the error is revealed, and not for a client that
// has disconnected.
func TestRespondNonContextError_ReportsTheHiddenError(t *testing.T) {
	rep := recordReports(t)
	hidden := errs.New("connection to 10.0.0.4 refused")
	RespondNonContextError(httptest.NewRecorder(), newReportRequest(http.MethodGet, nil), hidden)
	rep.checkReported(t, true, hidden)

	rep.reset()
	defer func(prev bool) { RevealInternalServerErrors = prev }(RevealInternalServerErrors)
	RevealInternalServerErrors = true
	RespondNonContextError(httptest.NewRecorder(), newReportRequest(http.MethodGet, nil), hidden)
	rep.checkReported(t, true, hidden)

	rep.reset()
	RespondNonContextError(httptest.NewRecorder(), newReportRequest(http.MethodGet, nil), fmt.Errorf("querying: %w", context.Canceled))
	RespondNonContextError(httptest.NewRecorder(), newReportRequest(http.MethodGet, nil), context.DeadlineExceeded)
	rep.checkReported(t, false)
}

// The hook gets the request's values but not its cancellation, so writing a
// report is not cut short by a client that went away meanwhile. Nothing is
// reported for a nil error, and a nil hook is no hook.
func TestReportHiddenError(t *testing.T) {
	ReportHiddenError(t.Context(), errors.New("no hook set")) // must not panic

	rep := recordReports(t)
	ctx, cancel := context.WithCancel(context.WithValue(t.Context(), requestKey{}, "request-1"))
	cancel()
	hidden := errors.New("pq: connection refused")
	ReportHiddenError(ctx, hidden)
	rep.checkReported(t, true, hidden)
	if len(rep.ctxs) == 1 && rep.ctxs[0].Err() != nil {
		t.Errorf("the hook's context is done: %v", rep.ctxs[0].Err())
	}

	rep.reset()
	ReportHiddenError(t.Context(), nil)
	ReportHiddenError(t.Context(), context.Canceled)
	ReportHiddenError(t.Context(), fmt.Errorf("calling the pricing service: %w", context.DeadlineExceeded))
	rep.checkReported(t, false)
}
