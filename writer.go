package mx

import (
	"io"
)

type Writer interface {
	io.Writer
	BeginElement(elem string) error
	Attribute(name, value string) error
	CloseAndEndElement() error
	CloseElement() error
	EscapeText(text string) error
	EndElement() error
	Comment(text string) error
	CDATA(text string) error
}

type WriterFactory interface {
	NewWriter(w io.Writer) Writer
}

type WriterFactoryFunc func(w io.Writer) Writer

func (f WriterFactoryFunc) NewWriter(w io.Writer) Writer {
	return f(w)
}
