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

type Raw string

func (raw Raw) Render(_ context.Context, w Writer) error {
	_, err := w.Write([]byte(raw))
	return err
}

type RawBytes []byte

func (raw RawBytes) Render(_ context.Context, w Writer) error {
	_, err := w.Write([]byte(raw))
	return err
}

const RawNewline Raw = "\n"

const Newline newline = 0

type newline int

func (newline) Render(_ context.Context, w Writer) error {
	return w.Newline()
}
