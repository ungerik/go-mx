package mx

import (
	"context"
	"fmt"
)

// Text is a string [Component] rendered as escaped text content. It is the
// default conversion for string children (see [DefaultAsComponent]); use [Raw]
// to write trusted markup without escaping.
type Text string

// Render writes the string to w as escaped text via [Writer.EscapeText].
func (t Text) Render(_ context.Context, w Writer) error {
	return w.EscapeText(string(t))
}

// Textf returns a [Text] built from a fmt.Sprintf format and arguments. The
// result is still escaped when rendered, so it is safe with untrusted arguments.
func Textf(format string, args ...any) Text {
	return Text(fmt.Sprintf(format, args...))
}
