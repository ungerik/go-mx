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

func TestSSEAttribs(t *testing.T) {
	// Names: the extension spells these without the hx- prefix. Getting that
	// wrong yields attributes htmx silently ignores, which looks like a server
	// that never sends anything.
	//
	// Values: SSESwap joins several event names with a comma, which is what
	// subscribing one element to more than one event depends on.
	for _, tc := range []struct {
		attrib    mx.Attrib
		wantName  string
		wantValue string
	}{
		{SSEConnect("/stream"), "sse-connect", "/stream"},
		{SSEConnect("/chat/stream?id=1"), "sse-connect", "/chat/stream?id=1"},
		{SSESwap("message"), "sse-swap", "message"},
		{SSESwap("message", "done"), "sse-swap", "message,done"},
		{SSEClose("done"), "sse-close", "done"},
	} {
		if got := tc.attrib.AttribName(); got != tc.wantName {
			t.Errorf("AttribName = %q, want %q", got, tc.wantName)
		}
		if got := attribValue(t, tc.attrib); got != tc.wantValue {
			t.Errorf("AttribValue of %s = %q, want %q", tc.wantName, got, tc.wantValue)
		}
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
