package mx

import (
	"context"
)

var (
	_ Component = Raw("")
	_ Component = RawBytes(nil)
	_ Component = RawNewline
	_ Component = Newline
)

// Raw is a string [Component] that is written to the output verbatim, without
// escaping. Use it only for trusted, already-valid markup; caller-supplied data
// should go through [Text] (or the conversion in [DefaultAsComponent]) instead.
type Raw string

// Render writes the raw string to w without escaping.
func (raw Raw) Render(_ context.Context, w Writer) error {
	_, err := w.Write([]byte(raw))
	return err
}

// RawBytes is the []byte counterpart of [Raw]: a byte-slice [Component] written
// to the output verbatim, without escaping.
type RawBytes []byte

// Render writes the raw bytes to w without escaping.
func (raw RawBytes) Render(_ context.Context, w Writer) error {
	_, err := w.Write([]byte(raw))
	return err
}

// RawNewline is a [Raw] component that writes a literal "\n" newline character
// to the output, independent of any writer indentation. For an indentation-aware
// line break use [Newline].
const RawNewline Raw = "\n"

// Newline is a [Component] that emits a line break via [Writer.Newline], which
// on an indenting writer also writes the current indentation prefix. For a plain
// literal "\n" use [RawNewline].
const Newline newline = 0

type newline int

// Render emits a line break by calling [Writer.Newline] on w.
func (newline) Render(_ context.Context, w Writer) error {
	return w.Newline()
}
