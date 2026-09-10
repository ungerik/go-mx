package hx

import (
	"testing"
)

func TestOnRendersDOMEventAttribute(t *testing.T) {
	a := On("click", "alert('hi')")
	if got, want := a.AttribName(), "hx-on:click"; got != want {
		t.Errorf("AttribName = %q, want %q", got, want)
	}
	if got, want := attribValue(t, a), "alert('hi')"; got != want {
		t.Errorf("AttribValue = %q, want %q", got, want)
	}
}

func TestOnHTMXRendersHtmxEventAttribute(t *testing.T) {
	a := OnHTMX("after-request", "doStuff()")
	if got, want := a.AttribName(), "hx-on::after-request"; got != want {
		t.Errorf("AttribName = %q, want %q", got, want)
	}
	if got, want := attribValue(t, a), "doStuff()"; got != want {
		t.Errorf("AttribValue = %q, want %q", got, want)
	}
}
