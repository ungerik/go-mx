package mx

import (
	"context"
)

var (
	_ Component = Raw("")
	_ Component = RawBytes(nil)
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

const (
	Newline Raw = "\n"
)
