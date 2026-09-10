package mx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/domonda/go-errs"
)

func TestRespondNonContextError_HidesTheDetailByDefault(t *testing.T) {
	// The 500 half of the policy [SSEResponse.SendError] now shares: an internal
	// error reaches the client as a status, never as the message that names a
	// host, a query or a file path. Both reporters read the same
	// nonContextErrorMessage, so the response side needs its own test or the
	// shared helper is only ever checked through the stream.
	rec := httptest.NewRecorder()
	RespondNonContextError(rec, errs.New("connection to 10.0.0.4 refused"))
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
	RespondNonContextError(rec, errs.New("connection to 10.0.0.4 refused"))
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
		RespondNonContextError(rec, err)
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
