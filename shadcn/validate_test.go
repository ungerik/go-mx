package shadcn

import (
	"context"
	"strings"
	"testing"

	"github.com/ungerik/go-mx"
)

// TestValidateIDReturnsError covers the PanicOnInvalidID=false branch of
// validateID. With the toggle off, validateID returns the error instead of
// panicking (the panic path is covered by TestValidateIDPanics), and the
// component constructors defer it to render time via mx.NewErrElement: a stray
// bad id then surfaces as a render error rather than a panic or injected markup.
//
// The original value is saved and restored so flipping this package-level global
// cannot leak into other tests in the package.
func TestValidateIDReturnsError(t *testing.T) {
	defer func(orig bool) { PanicOnInvalidID = orig }(PanicOnInvalidID)
	PanicOnInvalidID = false

	// validateID returns an error for invalid ids and nil for a valid one,
	// without panicking.
	for _, bad := range []string{"", "bad id", "has.dot", "quote'd"} {
		if err := validateID(bad); err == nil {
			t.Errorf("expected error for id %q, got nil", bad)
		}
	}
	if err := validateID("a-valid_ID-123"); err != nil {
		t.Errorf("unexpected error for valid id: %v", err)
	}

	// A constructor given a bad id defers the error to render time instead of
	// panicking: the element's Render returns the validation error.
	var sb strings.Builder
	err := RadioGroupItem("bad name", "v").Render(context.Background(), mx.NewCheckedWriter(&sb))
	if err == nil {
		t.Fatal("expected RadioGroupItem with invalid name to render a deferred error, got nil")
	}
	if !strings.Contains(err.Error(), "id must") {
		t.Errorf("expected an id-validation error, got: %v", err)
	}

	// A valid id renders cleanly with no deferred error.
	sb.Reset()
	if err := RadioGroupItem("group-1", "v").Render(context.Background(), mx.NewCheckedWriter(&sb)); err != nil {
		t.Errorf("unexpected render error for valid name: %v", err)
	}
}

func TestValidateIDAcceptsEveryKeyedIDValue(t *testing.T) {
	// validateID and mx.KeyedIDValue share mx.ValidIDRune precisely so that an
	// id built by mx can be handed to a component here — the streaming case:
	// mx.KeyedID marks the element, the same key builds the selector, and a
	// shadcn component wraps it. If the two rules drifted apart the mismatch
	// would surface as a panic (PanicOnInvalidID is on by default) at render
	// time, from an id the caller never typed and cannot see anything wrong with.
	defer func(orig bool) { PanicOnInvalidID = orig }(PanicOnInvalidID)
	PanicOnInvalidID = false

	for _, parts := range [][]any{
		{"msg", 3},
		{"msg", "6ba7b810-9dad-11d1-80b4-00c04fd430c8"},
		{"gallery", "stick-to-bottom", "log"},
		// Keys a real application passes through unchecked: non-ASCII letters
		// and punctuation are reduced by KeyedIDValue, and what it produces
		// still has to pass here.
		{"user", "Müller"},
		{"path", "a/b c.d"},
		{7},
	} {
		id := mx.KeyedIDValue(parts...)
		if id == "" {
			t.Errorf("KeyedIDValue(%v) is empty, so the case tests nothing", parts)
			continue
		}
		if err := validateID(id); err != nil {
			t.Errorf("validateID(%q) built from KeyedIDValue(%v): %v", id, parts, err)
		}
	}
}
