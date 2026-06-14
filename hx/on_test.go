package hx

import (
	"context"
	"testing"
)

func TestOnRendersDOMEventAttribute(t *testing.T) {
	a := On("click", "alert('hi')")
	if got, want := a.AttribName(), "hx-on:click"; got != want {
		t.Errorf("AttribName = %q, want %q", got, want)
	}
	v, err := a.AttribValue(context.Background())
	if err != nil {
		t.Fatalf("AttribValue: %v", err)
	}
	if want := "alert('hi')"; v != want {
		t.Errorf("AttribValue = %q, want %q", v, want)
	}
}

func TestOnHTMXRendersHtmxEventAttribute(t *testing.T) {
	a := OnHTMX("after-request", "doStuff()")
	if got, want := a.AttribName(), "hx-on::after-request"; got != want {
		t.Errorf("AttribName = %q, want %q", got, want)
	}
	v, err := a.AttribValue(context.Background())
	if err != nil {
		t.Fatalf("AttribValue: %v", err)
	}
	if want := "doStuff()"; v != want {
		t.Errorf("AttribValue = %q, want %q", v, want)
	}
}
