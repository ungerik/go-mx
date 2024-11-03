package mx

import (
	"io"
	"strings"
)

var (
	DefaultWriterFactory WriterFactory = WriterFactoryFunc(func(w io.Writer) Writer {
		return NewCheckedWriter(w)
	})

	TextEscaper = strings.NewReplacer(
		`&`, "&amp;",
		`'`, "&apos;",
		`<`, "&lt;",
		`>`, "&gt;",
		`"`, "&quot;",
	)

	AsComponent = DefaultAsComponent
)
