package hx

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResponseHeaderSetters(t *testing.T) {
	tests := []struct {
		name   string
		set    func(w http.ResponseWriter)
		header string
		want   string
	}{
		{"Location", func(w http.ResponseWriter) { SetLocation(w, "/next") }, HeaderLocation, "/next"},
		{"PushURL", func(w http.ResponseWriter) { SetPushURL(w, "/page") }, HeaderPushURL, "/page"},
		{"Redirect", func(w http.ResponseWriter) { SetRedirect(w, "/login") }, HeaderRedirect, "/login"},
		{"Refresh", func(w http.ResponseWriter) { SetRefresh(w) }, HeaderRefresh, "true"},
		{"ReplaceURL", func(w http.ResponseWriter) { SetReplaceURL(w, "false") }, HeaderReplaceURL, "false"},
		{"Reswap", func(w http.ResponseWriter) { SetReswap(w, "beforeend") }, HeaderReswap, "beforeend"},
		{"Retarget", func(w http.ResponseWriter) { SetRetarget(w, "#main") }, HeaderRetarget, "#main"},
		{"Reselect", func(w http.ResponseWriter) { SetReselect(w, "#frag") }, HeaderReselect, "#frag"},
		{"Trigger", func(w http.ResponseWriter) { SetTrigger(w, "a", "b") }, HeaderTrigger, "a, b"},
		{"TriggerAfterSettle", func(w http.ResponseWriter) { SetTriggerAfterSettle(w, "x") }, HeaderTriggerAfterSettle, "x"},
		{"TriggerAfterSwap", func(w http.ResponseWriter) { SetTriggerAfterSwap(w, "y") }, HeaderTriggerAfterSwap, "y"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			tt.set(w)
			if got := w.Header().Get(tt.header); got != tt.want {
				t.Errorf("%s = %q, want %q", tt.header, got, tt.want)
			}
		})
	}
}

func TestSetTriggerJSON(t *testing.T) {
	w := httptest.NewRecorder()
	if err := SetTriggerJSON(w, map[string]any{"showMessage": "hello"}); err != nil {
		t.Fatalf("SetTriggerJSON: %v", err)
	}
	if got, want := w.Header().Get(HeaderTrigger), `{"showMessage":"hello"}`; got != want {
		t.Errorf("HX-Trigger = %q, want %q", got, want)
	}
}

func TestSetTriggerJSON_MarshalError(t *testing.T) {
	w := httptest.NewRecorder()
	// A channel is not JSON-serializable, so Marshal must fail.
	err := SetTriggerJSON(w, map[string]any{"bad": make(chan int)})
	if err == nil {
		t.Fatal("expected marshal error, got nil")
	}
	if got := w.Header().Get(HeaderTrigger); got != "" {
		t.Errorf("HX-Trigger should be unset on error, got %q", got)
	}
}

func TestSetTriggerEmptyIsNoOp(t *testing.T) {
	w := httptest.NewRecorder()
	SetTrigger(w)            // no events
	SetTriggerAfterSettle(w) // no events
	SetTriggerAfterSwap(w)   // no events
	if err := SetTriggerJSON(w, nil); err != nil {
		t.Fatalf("SetTriggerJSON(nil): %v", err)
	}
	for _, h := range []string{HeaderTrigger, HeaderTriggerAfterSettle, HeaderTriggerAfterSwap} {
		if got := w.Header().Get(h); got != "" {
			t.Errorf("%s should be unset for empty events, got %q", h, got)
		}
	}
}

func TestRequestHeaderReaders(t *testing.T) {
	tests := []struct {
		name   string
		header string
		value  string
		read   func(*http.Request) bool
		want   bool
	}{
		{"IsRequest true", HeaderRequest, "true", IsRequest, true},
		{"IsRequest absent", HeaderRequest, "", IsRequest, false},
		{"IsBoosted true", HeaderBoosted, "true", IsBoosted, true},
		{"IsBoosted false", HeaderBoosted, "false", IsBoosted, false},
		{"IsHistoryRestoreRequest true", HeaderHistoryRestoreRequest, "true", IsHistoryRestoreRequest, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.value != "" {
				r.Header.Set(tt.header, tt.value)
			}
			if got := tt.read(r); got != tt.want {
				t.Errorf("%s = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
