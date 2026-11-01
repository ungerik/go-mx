package mx

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
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

// cancelledRequest returns a GET request whose context is done, as the
// request of a client that has disconnected is.
func cancelledRequest() *http.Request {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return httptest.NewRequestWithContext(ctx, http.MethodGet, "/", nil)
}

func TestRespondNonContextError_SilentForContextErrors(t *testing.T) {
	// A cancelled request context means the client that would read the 500
	// has disconnected, so there is nobody to tell — and net/http logs a
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
		RespondNonContextError(rec, cancelledRequest(), err)
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

// A timeout of a call the handler made, such as a query under its own
// context.WithTimeout, fails with context.DeadlineExceeded while the client
// still waits: that is a 500, not a client that went away. So is a real
// error joined with the context.Canceled of a fan-out's own cancel.
func TestRespondNonContextError_OwnTimeoutIsAServerError(t *testing.T) {
	for _, err := range []error{
		errs.Errorf("querying the ledger: %w", context.DeadlineExceeded),
		errors.Join(errors.New("pricing service: 503"), context.Canceled),
	} {
		rec := httptest.NewRecorder()
		RespondNonContextError(rec, httptest.NewRequest(http.MethodGet, "/", nil), err)
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("RespondNonContextError(%v) responded with status %d, want 500", err, rec.Code)
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

// blankMessageError is a UserMessageError whose UserMessage is empty, which
// gives the user nothing to read.
type blankMessageError struct{}

func (blankMessageError) Error() string       { return "crm: partner 7 has no display name" }
func (blankMessageError) UserMessage() string { return "" }

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

// serveCancelledReportRequest serves newReportRequest(method, form) with h
// under a context that is done, as the request of a client that has
// disconnected is.
func serveCancelledReportRequest(h http.Handler, method string, form url.Values) *httptest.ResponseRecorder {
	r := newReportRequest(method, form)
	ctx, cancel := context.WithCancel(r.Context())
	cancel()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, r.WithContext(ctx))
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
	for _, err := range []error{nil, errors.New("pq: connection refused"), FieldErrors{"Name": nil}, blankMessageError{}, FormMessage("")} {
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
	RespondNonContextError(httptest.NewRecorder(), cancelledRequest(), fmt.Errorf("querying: %w", context.Canceled))
	RespondNonContextError(httptest.NewRecorder(), cancelledRequest(), context.DeadlineExceeded)
	rep.checkReported(t, false)

	// A timeout of the handler's own call, while the client waits, is reported.
	timeout := fmt.Errorf("querying: %w", context.DeadlineExceeded)
	RespondNonContextError(httptest.NewRecorder(), newReportRequest(http.MethodGet, nil), timeout)
	rep.checkReported(t, true, timeout)
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
	ReportHiddenError(ctx, context.Canceled)
	ReportHiddenError(ctx, fmt.Errorf("calling the pricing service: %w", context.DeadlineExceeded))
	rep.checkReported(t, false)

	// The same errors while the request's context is not done come from a
	// timeout or cancel of the server's own, and are reported.
	timeout := fmt.Errorf("calling the pricing service: %w", context.DeadlineExceeded)
	ReportHiddenError(t.Context(), timeout)
	rep.checkReported(t, false, timeout)
}

// A sensitive field shows only user messages, all of them, never the text
// of what wraps them or is joined with them, which is reported, or else the
// mask, and reports the error it hides.
func TestMaskSensitiveError(t *testing.T) {
	rep := recordReports(t)
	hidden := errors.New(`"hunter2" is too short`)
	if got := MaskSensitiveError(t.Context(), hidden, "not a valid value"); got != error(FormMessage("not a valid value")) {
		t.Errorf("MaskSensitiveError = %q, want the mask", got)
	}
	rep.checkReported(t, false, hidden)

	rep.reset()
	wrapped := fmt.Errorf("checking %q: %w", "hunter2", FormMessage("Use at least 12 characters."))
	if got := MaskSensitiveError(t.Context(), wrapped, "not a valid value"); got != error(FormMessage("Use at least 12 characters.")) {
		t.Errorf("MaskSensitiveError = %q, want the user message alone", got)
	}
	rep.checkReported(t, false)

	quoting := fmt.Errorf("%q has no digit", "hunter2")
	joined := errors.Join(FormMessage("Use at least 12 characters."), quoting, FormMessage("Add a digit."))
	got := MaskSensitiveError(t.Context(), joined, "not a valid value")
	if texts := FieldErrorTexts(got); !slices.Equal(texts, []string{"Use at least 12 characters.", "Add a digit."}) {
		t.Errorf("MaskSensitiveError of a join shows %q, want both user messages", texts)
	}
	rep.checkReported(t, false, quoting)
}

// asMessage supplies its user message only through an As method.
type asMessage struct{}

func (asMessage) Error() string { return "internal: rule 7 at db-host-3" }

func (asMessage) As(target any) bool {
	if m, ok := target.(*UserMessageError); ok {
		*m = FormMessage("Check the IBAN.")
		return true
	}
	return false
}

// The user message of a field's error is found as errors.As finds it, also
// through an As method; an empty one does not hide a real one deeper in the
// tree; and an error without text still says something.
func TestFieldErrorTexts_EdgeCases(t *testing.T) {
	if texts := FieldErrorTexts(asMessage{}); !slices.Equal(texts, []string{"Check the IBAN."}) {
		t.Errorf("an As method's message: %q", texts)
	}
	if msg, ok := UserMessage(errors.Join(FormMessage(""), FormMessage("x"))); !ok || msg != "x" {
		t.Errorf("UserMessage behind an empty one = %q, %t", msg, ok)
	}
	if texts := FieldErrorTexts(errors.Join(FormMessage(""), FormMessage("x"))); !slices.Equal(texts, []string{"x"}) {
		t.Errorf("FieldErrorTexts behind an empty message: %q", texts)
	}
	if texts := FieldErrorTexts(FieldErrors{}); !slices.Equal(texts, []string{"not a valid value"}) {
		t.Errorf("an error without text: %q", texts)
	}
}

// wrappedQuote keeps the quoting error it wraps behind a short text.
type wrappedQuote struct{ cause error }

func (e wrappedQuote) Error() string { return "invalid code" }
func (e wrappedQuote) Unwrap() error { return e.cause }

// panicOnError fails the test the moment its text is built.
type panicOnError struct{}

func (panicOnError) Error() string { panic("Error() built for text far past the bound") }

// The bound of a field's error: for posted text far past it the error's text
// is never built; a long error of short text is replaced and reported, as no
// forged value built it; a short error of long text keeps only its text, so
// the hook does not get the forged value it wraps. A user message is bounded
// too, and so is each part of an error joined with one.
func TestBoundFieldError(t *testing.T) {
	rep := recordReports(t)
	quote := func(raw string) error { return fmt.Errorf("%q is no code", raw) }
	long := strings.Repeat("x", 600)

	const notValid = FormMessage("the text is not a valid value")
	if err := BoundFieldError(t.Context(), panicOnError{}, strings.Repeat("x", 64*MaxFieldErrorLen+64), notValid); err != error(notValid) {
		t.Errorf("far past the bound: %v", err)
	}
	if err := BoundFieldError(t.Context(), quote(long), long, notValid); err != error(notValid) {
		t.Errorf("a quote of long text: %v", err)
	}
	// Built from a value that may be forged, neither error is reported as it
	// is, only as a stand-in naming its type and lengths.
	if texts := rep.texts(); len(texts) != 2 || strings.Contains(strings.Join(texts, ""), "xxxx") ||
		!strings.Contains(texts[0], "mx.panicOnError") || !strings.Contains(texts[0], "not built") || !strings.Contains(texts[1], "613 bytes") {
		t.Errorf("reported %q, want two stand-ins", texts)
	}
	rep.reset()
	hint := errors.New("cannot parse value into the type: " + long)
	if err := BoundFieldError(t.Context(), hint, "x", notValid); err != error(notValid) {
		t.Errorf("a long error of short text: %v", err)
	}
	rep.checkReported(t, false, hint)
	rep.reset()
	kept := BoundFieldError(t.Context(), wrappedQuote{quote(long)}, long, notValid)
	if kept.Error() != "invalid code" || errors.Unwrap(kept) != nil {
		t.Errorf("a short error of long text: %v, wraps %v; want only its text", kept, errors.Unwrap(kept))
	}
	short := quote("abc")
	if err := BoundFieldError(t.Context(), short, "abc", notValid); err != short {
		t.Errorf("a short error of short text: %v, want it as it is", err)
	}
	if msg := FormMessage("Use a shorter code."); BoundFieldError(t.Context(), msg, long, notValid) != error(msg) {
		t.Error("a short user message must stay as it is")
	}
	if err := BoundFieldError(t.Context(), FormMessage(long), long, notValid); err != error(notValid) {
		t.Errorf("a user message past the bound: %v", err)
	}
	joined := errors.Join(FormMessage("Use a shorter code."), quote(long))
	if texts := FieldErrorTexts(BoundFieldError(t.Context(), joined, long, notValid)); !slices.Equal(texts, []string{"Use a shorter code.", string(notValid)}) {
		t.Errorf("a part joined with a user message: %q", texts)
	}
	wrapped := fmt.Errorf("checking: %w", joined)
	if texts := FieldErrorTexts(BoundFieldError(t.Context(), wrapped, long, notValid)); !slices.Equal(texts, []string{"Use a shorter code.", string(notValid)}) {
		t.Errorf("a wrapped join: %q", texts)
	}
	shortJoin := errors.Join(FormMessage("a"), errors.New("b"))
	if err := BoundFieldError(t.Context(), shortJoin, "x", notValid); err != shortJoin {
		t.Errorf("a short join must stay as it is: %v", err)
	}

	// A bound too large to multiply means no bound in practice, not none at all.
	prev := MaxFieldErrorLen
	t.Cleanup(func() { MaxFieldErrorLen = prev })
	MaxFieldErrorLen = math.MaxInt
	if err := BoundFieldError(t.Context(), short, "abc", notValid); err != short {
		t.Errorf("a bound of MaxInt: %v, want the error as it is", err)
	}
}

// addressErrors is a multi-error whose own UserMessage speaks for the
// internal causes it wraps.
type addressErrors []error

func (e addressErrors) Error() string       { return errors.Join(e...).Error() }
func (e addressErrors) Unwrap() []error     { return e }
func (e addressErrors) UserMessage() string { return "Please check the address." }

// A field shows the UserMessage of an error, never the Error text of a
// wrapper around it or of the causes an error with its own message wraps;
// every message of a joined error, also a wrapped one, is a text of its own,
// in a stable order, and an error with no user message shows itself.
func TestFieldErrorTexts(t *testing.T) {
	for _, c := range []struct {
		name string
		err  error
		want []string
	}{
		{"nil", nil, nil},
		{"plain", errors.New("strconv: invalid syntax"), []string{"strconv: invalid syntax"}},
		{"plain wrapped", fmt.Errorf("parsing: %w", errors.New("invalid syntax")), []string{"parsing: invalid syntax"}},
		{"domain", ibanError{"DE00"}, []string{"The IBAN is not valid."}},
		{"wrapped", fmt.Errorf("partner 7: %w", ibanError{"DE00"}), []string{"The IBAN is not valid."}},
		{"two messages", errors.Join(FormMessage("at least 12 characters"), FormMessage("one digit")), []string{"at least 12 characters", "one digit"}},
		{"wrapped join", fmt.Errorf("password: %w", errors.Join(FormMessage("at least 12 characters"), FormMessage("one digit"))), []string{"at least 12 characters", "one digit"}},
		{"message and plain", errors.Join(fmt.Errorf("rule: %w", FormMessage("one digit")), errors.New("too short")), []string{"one digit", "too short"}},
		{"own message over causes", addressErrors{errors.New("geo: lookup at 10.0.0.5 failed"), errors.New("pq: row 7 locked")}, []string{"Please check the address."}},
		{"field errors by path", FieldErrors{"B": FormMessage("b"), "A": FormMessage("a"), "C": nil}, []string{"a", "b"}},
		{"empty message", blankMessageError{}, []string{"crm: partner 7 has no display name"}},
		{"plain join", errors.Join(errors.New("too short"), errors.New("no digits")), []string{"too short\nno digits"}},
		{"plain multi-wrap", fmt.Errorf("must be between %w and %w", errors.New("10"), errors.New("20")), []string{"must be between 10 and 20"}},
	} {
		if got := FieldErrorTexts(c.err); !slices.Equal(got, c.want) {
			t.Errorf("%s: FieldErrorTexts = %q, want %q", c.name, got, c.want)
		}
	}
	// A map's order is random; the texts follow the paths every time.
	many := FieldErrors{"E": FormMessage("e"), "D": FormMessage("d"), "C": FormMessage("c"), "B": FormMessage("b"), "A": FormMessage("a")}
	for range 20 {
		if got := FieldErrorTexts(many); !slices.Equal(got, []string{"a", "b", "c", "d", "e"}) {
			t.Fatalf("FieldErrorTexts(FieldErrors) = %q, want path order", got)
		}
	}
}
