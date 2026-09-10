package mx

import (
	"context"
	"testing"
)

func TestKeyedIDValue(t *testing.T) {
	for _, tc := range []struct {
		name  string
		parts []any
		want  string
	}{
		{"parts joined", []any{"msg", "abc", 3}, "_msg-abc-3"},
		{"mixed types formatted plainly", []any{"row", 42, true}, "_row-42-true"},
		// The '_' prefix is what keeps a numeric first part a valid HTML id.
		{"numeric first part", []any{7}, "_7"},
		// A uu.ID-shaped part must survive intact: its hyphens are meaningful.
		{"uuid part kept whole", []any{"msg", "6ba7b810-9dad-11d1-80b4-00c04fd430c8"},
			"_msg-6ba7b810-9dad-11d1-80b4-00c04fd430c8"},
		// Characters that would break a CSS selector or an htmx target reduce to
		// a separator rather than being passed through.
		{"unsafe characters reduced", []any{"a/b c.d"}, "_a-b-c-d"},
		{"runs collapse", []any{"a///b"}, "_a-b"},
		// Empty parts must not leave a dangling separator: "_-b" and "_a-" are
		// ugly ids and the trailing one is unstable under an appended part.
		{"empty leading part", []any{"", "b"}, "_b"},
		{"empty trailing part", []any{"a", ""}, "_a"},
		{"empty middle part", []any{"a", "", "b"}, "_a-b"},
		// Literal dashes are kept as written, so ids that differ only by them
		// stay distinct.
		{"literal dashes kept", []any{"a--b"}, "_a--b"},
		// Nothing usable at all is reported as empty for KeyedID to reject.
		{"no parts", nil, ""},
		{"only empty parts", []any{"", ""}, ""},
		{"only unsafe characters", []any{"///"}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := KeyedIDValue(tc.parts...); got != tc.want {
				t.Errorf("KeyedIDValue(%v) = %q, want %q", tc.parts, got, tc.want)
			}
		})
	}
}

func TestKeyedIDValue_StableAcrossCalls(t *testing.T) {
	// This is the entire reason KeyedID exists next to UniqueID: an out-of-band
	// swap rendered by a later request has to reproduce the id the first render
	// emitted, and UniqueID cannot.
	first := KeyedIDValue("msg", 12, "part", 3)
	second := KeyedIDValue("msg", 12, "part", 3)
	if first != second {
		t.Fatalf("KeyedIDValue is not deterministic: %q then %q", first, second)
	}

	a, err := UniqueID().AttribValue(t.Context())
	if err != nil {
		t.Fatalf("UniqueID: %v", err)
	}
	b, err := UniqueID().AttribValue(t.Context())
	if err != nil {
		t.Fatalf("UniqueID: %v", err)
	}
	if a == b {
		t.Error("UniqueID repeated a value, so the contrast KeyedID documents does not hold")
	}
}

func TestKeyedID_RendersIDAttribute(t *testing.T) {
	attrib := KeyedID("msg", 1)
	if got := attrib.AttribName(); got != "id" {
		t.Errorf("AttribName = %q, want \"id\"", got)
	}
	value, err := attrib.AttribValue(t.Context())
	if err != nil {
		t.Fatalf("AttribValue: %v", err)
	}
	if want := KeyedIDValue("msg", 1); value != want {
		t.Errorf("AttribValue = %q, want %q", value, want)
	}
}

func TestKeyedID_UnusableParts(t *testing.T) {
	// An empty id matches nothing, so a swap targeting it silently does nothing.
	// Deferring the error to render time keeps that from being invisible without
	// letting a bad key panic the program.
	for _, parts := range [][]any{nil, {""}, {"///"}} {
		attrib := KeyedID(parts...)
		if _, ok := attrib.(ErrAttrib); !ok {
			t.Errorf("KeyedID(%v) = %T, want ErrAttrib", parts, attrib)
			continue
		}
		if _, err := attrib.AttribValue(context.Background()); err == nil {
			t.Errorf("KeyedID(%v) rendered without an error", parts)
		}
	}
}
