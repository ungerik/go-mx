package mx

import (
	"context"
	"fmt"
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
		// a separator rather than being passed through, and the digest suffix
		// records what the reduction dropped.
		{"unsafe characters reduced", []any{"a/b c.d"}, "_a-b-c-d-3nfm9cgxexhrt"},
		{"runs collapse", []any{"a///b"}, "_a-b-3t6m3knugd6kq"},
		// Empty parts must not leave a dangling separator: "_-b" and "_a-" are
		// ugly ids and the trailing one is unstable under an appended part. They
		// are still lossy, though — the readable form cannot say where the empty
		// part was — so they carry the disambiguating digest like any other
		// reduction. See TestKeyedIDValue_EmptyPartsDoNotCollide.
		{"empty leading part", []any{"", "b"}, "_b-12hxi64wxvnba"},
		{"empty trailing part", []any{"a", ""}, "_a-d4t3nyqq7gr7"},
		{"empty middle part", []any{"a", "", "b"}, "_a-b-2yc2swuiyqk6a"},
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

func TestKeyedIDValue_EmptyPartsDoNotCollide(t *testing.T) {
	// An empty part contributes no characters, so the readable form of
	// ("user", "", "x") and ("user", "x", "") is identical. Without the digest
	// both render "_user-x" and an out-of-band swap for one lands in the
	// other's element — the collision is invisible because both ids are
	// individually valid and the swap simply hits whichever matches first.
	distinct := map[string][]any{}
	for _, parts := range [][]any{
		{"user", "", "x"},
		{"user", "x", ""},
		{"", "user", "x"},
		{"user", "x"},
	} {
		id := KeyedIDValue(parts...)
		if id == "" {
			t.Fatalf("KeyedIDValue(%v) is empty", parts)
		}
		if other, clash := distinct[id]; clash {
			t.Errorf("KeyedIDValue(%v) and KeyedIDValue(%v) both yield %q", parts, other, id)
			continue
		}
		distinct[id] = parts
	}
}

func TestKeyedIDValue_ReducedCharactersDoNotCollide(t *testing.T) {
	// An id is a swap target: two keys that reduce to the same characters would
	// make every out-of-band swap for one land in the other's element, and the
	// first match silently wins. Keys that differ only in reduced characters
	// are exactly the ones a caller cannot see the difference in, so they are
	// the ones that have to stay distinct.
	for _, keys := range [][2][]any{
		{{"msg", "a.b"}, {"msg", "a/b"}},
		{{"user", "Müller"}, {"user", "Mäller"}},
		{{"msg", "a.b"}, {"msg", "a", "b"}},
	} {
		first, second := KeyedIDValue(keys[0]...), KeyedIDValue(keys[1]...)
		if first == second {
			t.Errorf("KeyedIDValue(%v) and KeyedIDValue(%v) both yield %q", keys[0], keys[1], first)
		}
	}
}

func TestKeyedIDValue_DigestOnlyWhenSomethingWasReduced(t *testing.T) {
	// The digest is the price of a lossy reduction, not a tax on every id: a
	// readable id is the whole reason KeyedIDValue does not simply hash, and a
	// key made of usable characters loses nothing to begin with.
	for _, parts := range [][]any{
		{"msg", "abc", 3},
		{"msg", "6ba7b810-9dad-11d1-80b4-00c04fd430c8"},
		{"gallery", "stick-to-bottom", "log"},
	} {
		want := "_" + fmt.Sprint(parts[0])
		for _, part := range parts[1:] {
			want += "-" + fmt.Sprint(part)
		}
		if got := KeyedIDValue(parts...); got != want {
			t.Errorf("KeyedIDValue(%v) = %q, want %q", parts, got, want)
		}
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

func TestKeyedIDValue_DigestDistinguishesPartLengths(t *testing.T) {
	// These pairs reduce to the same readable id *and* to the same concatenated
	// text, so only length-prefixing each part inside the digest keeps them
	// apart. Without it two different keys would produce one id, and the
	// collision is silent: an out-of-band swap lands in whichever element the
	// selector matches first, so one entity's updates appear under another.
	for _, keys := range [][2][]any{
		{{"a.", "b"}, {"a", ".b"}},
		{{"msg", "1.", "2"}, {"msg", "1", ".2"}},
	} {
		first, second := KeyedIDValue(keys[0]...), KeyedIDValue(keys[1]...)
		if first == second {
			t.Errorf("KeyedIDValue(%v) and KeyedIDValue(%v) both yield %q", keys[0], keys[1], first)
		}
	}
}
