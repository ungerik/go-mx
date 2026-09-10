package mx

import (
	"context"
	"fmt"
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
// caller can refer to that element: "#"+KeyedIDValue("msg", id) is the
// hx-target of an out-of-band swap into the element KeyedID("msg", id) marked.
//
// Each part is formatted with fmt.Sprint rather than the pretty printer used
// for components (see [DefaultAsComponent]) — an id wants the plain form of a
// value, not a type-tagged dump. Characters outside letters, digits, '_' and
// '-' become a single '-' separator, as does each part boundary, and the result
// is prefixed with '_' so it is a valid HTML id that does not start with a
// digit. That allowlist is the same one the shadcn components validate ids
// against, and it deliberately reduces non-ASCII letters away.
//
// A readable id is worth more in devtools than a hash would be, at the price
// that the separator is not escaped: KeyedIDValue("a-b") and
// KeyedIDValue("a", "b") produce the same id. Passing one part per key avoids
// it, and only a caller mixing both spellings for the same entity can hit it.
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
	for i, part := range parts {
		if i > 0 {
			separator = true
		}
		for _, r := range fmt.Sprint(part) {
			keep := r == '_' || r == '-' ||
				(r >= '0' && r <= '9') ||
				(r >= 'A' && r <= 'Z') ||
				(r >= 'a' && r <= 'z')
			if !keep {
				separator = true
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
	return b.String()
}
