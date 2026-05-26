package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/shadcn"
)

func newTestServer(t *testing.T) (http.Handler, *store) {
	t.Helper()
	s := &store{p: Profile{ID: "user-1", Name: "Ada", Email: "ada@example.com", Age: 36}}
	mux := http.NewServeMux()
	mux.Handle("/admin/profile", mx.ReflectFormHandler(s.Load, s.Save))
	return mx.Middleware(shadcn.FieldDecider)(mux), s
}

func TestExampleGETRendersAllFields(t *testing.T) {
	srv, _ := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/admin/profile", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{
		`name="Name"`, `value="Ada"`,
		`name="Email"`, `type="email"`,
		`name="Bio"`, `<textarea`,
		`name="Age"`, `type="number"`,
		`name="Active"`, `data-slot="switch"`,
		`name="Tier"`, `<select`,
		`name="Features"`, `data-slot="checkbox"`,
		`name="Password"`, `type="password"`,
		`name="Account-VATNumber"`, // section recursion uses hyphen path
		`__present__Name`,          // sentinel emitted by handler
		`method="POST"`,
		`enctype="multipart/form-data"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing %q", want)
		}
	}
}

func TestExamplePOSTUpdatesProfile(t *testing.T) {
	srv, s := newTestServer(t)

	form := url.Values{}
	form.Set(mx.PresentSentinelName("Name"), "1")
	form.Set("Name", "Grace Hopper")
	form.Set(mx.PresentSentinelName("Email"), "1")
	form.Set("Email", "grace@example.com")
	form.Set(mx.PresentSentinelName("Age"), "1")
	form.Set("Age", "85")
	form.Set(mx.PresentSentinelName("Active"), "1")
	form.Set("Active", "on")
	form.Set(mx.PresentSentinelName("Tier"), "1")
	form.Set("Tier", "enterprise")
	form.Set(mx.PresentSentinelName("Features"), "1")
	form["Features"] = []string{"sso", "exports"}
	// Sensitive password input is not sent on round-trip; skip.
	// Joined is readonly so __present is permitted but the parser
	// will refuse to overwrite.
	form.Set(mx.PresentSentinelName("Joined"), "1")
	form.Set("Joined", "2999-01-01T00:00:00")

	req := httptest.NewRequest(http.MethodPost, "/admin/profile",
		strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}

	got := s.p
	if got.Name != "Grace Hopper" || got.Email != "grace@example.com" || got.Age != 85 {
		t.Errorf("save: got %+v", got)
	}
	if got.Tier != "enterprise" {
		t.Errorf("tier=%q", got.Tier)
	}
	if _, ok := got.Features["sso"]; !ok {
		t.Errorf("missing sso: %v", got.Features)
	}
	if _, ok := got.Features["exports"]; !ok {
		t.Errorf("missing exports: %v", got.Features)
	}
	if !got.Joined.IsZero() && got.Joined.Year() == 2999 {
		t.Errorf("readonly field overwritten: %v", got.Joined)
	}
}

// Per-field required check is the framework's responsibility — verify
// it intercepts before onSubmit.
func TestExamplePOSTRequiredCheckFires(t *testing.T) {
	srv, _ := newTestServer(t)
	form := url.Values{}
	form.Set(mx.PresentSentinelName("Name"), "1")
	form.Set("Name", "") // empty + required

	req := httptest.NewRequest(http.MethodPost, "/admin/profile",
		strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d (want 200 re-render)", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `data-error="Name"`) {
		t.Errorf("expected inline Name error: %s", body)
	}
	if !strings.Contains(body, "required") {
		t.Errorf("expected required-check error message: %s", body)
	}
}

// Cross-field invariant is the caller's responsibility — verify that
// onSubmit's FieldErrors route to the right field inline.
func TestExamplePOSTCrossFieldError(t *testing.T) {
	srv, _ := newTestServer(t)
	form := url.Values{}
	form.Set(mx.PresentSentinelName("Name"), "1")
	form.Set("Name", "Ada")
	form.Set(mx.PresentSentinelName("Tier"), "1")
	form.Set("Tier", "enterprise")
	form.Set(mx.PresentSentinelName("Features"), "1")
	form["Features"] = []string{"audit-log"} // no SSO

	req := httptest.NewRequest(http.MethodPost, "/admin/profile",
		strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d (want 200 re-render)", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `data-error="Features"`) {
		t.Errorf("expected inline Features error: %s", body)
	}
	if !strings.Contains(body, "enterprise tier requires SSO") {
		t.Errorf("expected cross-field message: %s", body)
	}
}
