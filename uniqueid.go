package mx

import (
	"context"
	"fmt"
	"hash/fnv"
	"strconv"
	"strings"
)

// UniqueID returns an "id" [Attrib] whose value is unique within the running
// process. It draws from an atomic counter, so it is safe for concurrent use,
// and formats the value as "_" followed by the count in base 36, yielding a
// valid HTML id that does not start with a digit.
func UniqueID() Attrib {
	return uniqueID(idCounter.Add(1))
}

type uniqueID uint64

// AttribName returns "id".
func (id uniqueID) AttribName() string {
	return "id"
}

// AttribValue returns the unique value formatted as "_" followed by the count
// in base 36, and a nil error.
func (id uniqueID) AttribValue(context.Context) (string, error) {
	return "_" + strconv.FormatUint(uint64(id), 36), nil
}

// ValidIDRune reports whether r may appear in an HTML id generated or accepted
// by go-mx: a letter, a digit, '_' or '-'.
//
// It is the single definition of that rule, so the sanitizer that builds ids
// ([KeyedIDValue]) and the validators that reject them cannot drift apart. A
// drift would be silent until a sanitized id made a validator fail, so this is
// deliberately one function rather than the same predicate written twice.
//
// The set is ASCII-only, which reduces non-ASCII letters away in [KeyedIDValue]
// rather than passing through characters that are valid HTML but awkward in a
// CSS selector or a JavaScript string.
func ValidIDRune(r rune) bool {
	return r == '_' || r == '-' ||
		(r >= '0' && r <= '9') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= 'a' && r <= 'z')
}

// KeyedID returns an "id" [Attrib] whose value is derived deterministically from
// parts, so the same parts yield the same id in every render and in every
// process.
//
// That determinism is the whole difference to [UniqueID], which draws from a
// process-lifetime counter: two renders of the same entity get different ids,
// and a restart starts the sequence over. An out-of-band swap addresses its
// target by id, so appending to an element that an earlier request rendered — a
// streamed message growing token by token, say — needs an id the later render
// can reproduce. Use UniqueID when an id only has to be distinct within one
// document, and KeyedID when it has to be found again later.
//
// The value is built by [KeyedIDValue]: the parts are formatted with
// fmt.Sprint, joined with '-', reduced to letters, digits, '_' and '-', and
// prefixed with '_', so KeyedID("msg", id, 3) yields something like
// "_msg-6ba7b810-9dad-11d1-80b4-00c04fd430c8-3". Parts that reduce to no usable
// characters yield an [ErrAttrib], because an empty id would silently match
// nothing.
//
// Use [KeyedIDValue] to build a selector referring to the same id.
func KeyedID(parts ...any) Attrib {
	value := KeyedIDValue(parts...)
	if value == "" {
		return ErrAttribf("id", "mx: KeyedID parts contain no usable characters: %v", parts)
	}
	return Attribute{Name: "id", Value: value}
}

// KeyedIDValue returns the id value that [KeyedID] renders for parts, so a
// caller can refer to that element: "#"+KeyedIDValue("msg", id) is the selector
// of an out-of-band swap into the element KeyedID("msg", id) marked, as in
// hx.SwapOOB(hx.SwapBeforeEnd, "#"+KeyedIDValue("msg", id)).
//
// Each part is formatted with fmt.Sprint rather than the pretty printer used
// for components (see [DefaultAsComponent]) — an id wants the plain form of a
// value, not a type-tagged dump. Characters outside letters, digits, '_' and
// '-' become a single '-' separator, as does each part boundary, and the result
// is prefixed with '_' so it is a valid HTML id that does not start with a
// digit. The kept characters are exactly [ValidIDRune], which the shadcn
// components validate ids against too, so a KeyedIDValue is always accepted
// there; it deliberately reduces non-ASCII letters away.
//
// Reducing characters away would make different keys collide — "a.b" and "a/b"
// are one id otherwise, and so are "Müller" and "Mäller" — and a collision is
// invisible: an out-of-band swap simply lands in the first element that
// matches. So whenever a rune was reduced, a short digest of the exact parts is
// appended, which keeps distinct keys distinct while leaving an id built from
// usable characters alone, fully readable.
//
// The digest does not cover the part boundaries: KeyedIDValue("a-b") and
// KeyedIDValue("a", "b") still produce the same id, because keeping a literal
// '-' readable is worth more than that case, which only a caller mixing both
// spellings for the same entity can reach.
//
// It returns an empty string when parts reduce to no usable characters at all,
// which is the case [KeyedID] reports as an [ErrAttrib].
func KeyedIDValue(parts ...any) string {
	var b strings.Builder
	b.WriteByte('_')
	// separator records that a part boundary or a reduced character is pending.
	// Writing it only before the next kept rune collapses runs and drops
	// leading and trailing separators in one pass. Literal '-' runes are kept
	// as they are, so they still distinguish ids that only differ by them.
	separator := false
	// reduced records that at least one rune did not survive, which is what
	// makes the readable form ambiguous and calls for the digest suffix.
	reduced := false
	for i, part := range parts {
		if i > 0 {
			separator = true
		}
		text := fmt.Sprint(part)
		if text == "" {
			// An empty part contributes no runes, so the loop below cannot
			// notice it. Without this, ("user", "", "x") and ("user", "x", "")
			// both reduce to "_user-x" — the same invisible collision the digest
			// exists to prevent, just reached by a different route.
			reduced = true
		}
		for _, r := range text {
			if !ValidIDRune(r) {
				separator = true
				reduced = true
				continue
			}
			if separator && b.Len() > 1 {
				b.WriteByte('-')
			}
			separator = false
			b.WriteRune(r)
		}
	}
	if b.Len() == 1 {
		return ""
	}
	if reduced {
		b.WriteByte('-')
		b.WriteString(keyedIDDigest(parts))
	}
	return b.String()
}

// keyedIDDigest returns a deterministic digest of the exact parts in base 36,
// which [KeyedIDValue] appends to an id whose readable form dropped characters.
//
// FNV-1a is chosen for being stable across processes and Go versions, not for
// being cryptographic: this distinguishes keys a caller chose, it does not
// defend against keys chosen to collide.
func keyedIDDigest(parts []any) string {
	hash := fnv.New64a()
	for _, part := range parts {
		s := fmt.Sprint(part)
		// Length-prefixed, so the parts cannot run into each other and make
		// ("ab", "c") and ("a", "bc") the same digest.
		fmt.Fprintf(hash, "%d:%s", len(s), s)
	}
	return strconv.FormatUint(hash.Sum64(), 36)
}
