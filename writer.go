package mx

import (
	"io"
)

// Writer is the markup sink that components render into. Beyond the raw
// io.Writer (used for unescaped output such as [Raw]), it exposes the element
// lifecycle the renderer drives: BeginElement opens a start tag, Attribute adds
// an attribute to the open start tag, CloseElementStartTag closes it, EscapeText
// writes escaped text content, and EndElement closes the current element (as a
// void element if the start tag is still open). Comment and CDATA write those
// constructs and Newline emits an (indentation-aware) line break. The methods
// must be called in a valid order; a [CheckedWriter] validates this and reports
// misuse as an error. Implementations may escape, validate and indent the output.
type Writer interface {
	io.Writer
	BeginElement(elem string) error
	Attribute(name, value string) error
	CloseElementStartTag() error
	EscapeText(text string) error
	EndElement() error
	Comment(text string) error
	CDATA(text string) error
	Newline() error
}

// WriterFactory creates a [Writer] that renders into the given io.Writer. It
// lets callers configure how markup is written (escaping, validation,
// indentation) once and reuse that configuration per output destination.
type WriterFactory interface {
	NewWriter(w io.Writer) Writer
}

// WriterFactoryFunc adapts a function to the [WriterFactory] interface.
type WriterFactoryFunc func(w io.Writer) Writer

// NewWriter calls the function, satisfying the [WriterFactory] interface.
func (f WriterFactoryFunc) NewWriter(w io.Writer) Writer {
	return f(w)
}
