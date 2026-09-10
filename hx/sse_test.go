package hx

import (
	"testing"

	"github.com/ungerik/go-mx"
)

func attribValue(t *testing.T, a mx.Attrib) string {
	t.Helper()
	value, err := a.AttribValue(t.Context())
	if err != nil {
		t.Fatalf("AttribValue: %v", err)
	}
	return value
}

func TestSSEAttribNames(t *testing.T) {
	// The extension spells these without the hx- prefix. Getting that wrong
	// yields attributes htmx silently ignores, which looks like a server that
	// never sends anything.
	for _, tc := range []struct {
		attrib mx.Attrib
		want   string
	}{
		{SSEConnect("/stream"), "sse-connect"},
		{SSESwap("message"), "sse-swap"},
		{SSEClose("done"), "sse-close"},
	} {
		if got := tc.attrib.AttribName(); got != tc.want {
			t.Errorf("AttribName = %q, want %q", got, tc.want)
		}
	}
}

func TestSSESwap_JoinsEventNames(t *testing.T) {
	if got := attribValue(t, SSESwap("message")); got != "message" {
		t.Errorf("SSESwap(one) = %q, want %q", got, "message")
	}
	if got, want := attribValue(t, SSESwap("message", "done")), "message,done"; got != want {
		t.Errorf("SSESwap(two) = %q, want %q", got, want)
	}
}

func TestSSESwap_NoEventsIsAnError(t *testing.T) {
	// An empty sse-swap subscribes to nothing, so the element would just never
	// update — a failure with no symptom other than silence.
	attrib := SSESwap()
	if _, ok := attrib.(mx.ErrAttrib); !ok {
		t.Fatalf("SSESwap() = %T, want mx.ErrAttrib", attrib)
	}
	if _, err := attrib.AttribValue(t.Context()); err == nil {
		t.Error("SSESwap() rendered without an error")
	}
}

func TestSSEAttribValues(t *testing.T) {
	if got := attribValue(t, SSEConnect("/chat/stream?id=1")); got != "/chat/stream?id=1" {
		t.Errorf("SSEConnect = %q", got)
	}
	if got := attribValue(t, SSEClose("done")); got != "done" {
		t.Errorf("SSEClose = %q", got)
	}
}
