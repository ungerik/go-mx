package mx

import (
	"io"
	"strings"
)

// The variables in this file are the package-level configuration of mx. Each
// has a working default and is consulted while a markup tree is built or
// rendered, so assigning a different value changes that behavior for the whole
// program. They are plain package variables with no locking: set them once
// during initialization (before any concurrent rendering), not while rendering.
var (
	// DefaultWriterFactory builds the Writer used wherever a caller does not
	// supply one of its own. By default it returns a CheckedWriter, which
	// validates element nesting and escapes text and attribute values. Replace
	// it to change the default writer program-wide (for example to enable
	// indentation with NewCheckedWriter(w).WithIndent(...)).
	DefaultWriterFactory WriterFactory = WriterFactoryFunc(func(w io.Writer) Writer {
		return NewCheckedWriter(w)
	})

	// TextEscaper escapes text nodes; it is also the default text escaper of
	// CheckedWriter. It replaces the five characters significant in markup text
	// content (&, ', <, >, "). That set is identical for HTML, XHTML, SVG and
	// XML, so a single escaper serves every target — escaping is a Writer
	// concern, which keeps value-to-Component conversion (see AsComponent)
	// independent of the target syntax. Reassign it to customize text escaping
	// globally. (Attribute values use the writer's own quote-aware escapers, not
	// this one.)
	TextEscaper = strings.NewReplacer(
		`&`, "&amp;",
		`'`, "&apos;",
		`<`, "&lt;",
		`>`, "&gt;",
		`"`, "&quot;",
	)

	// AsAttribs converts a value passed in attribute position into a slice of
	// Attrib, reporting whether the value was recognized as attributes at all
	// (so element constructors can fall back to treating it as a child). It
	// defaults to DefaultAsAttribs. Reassign it to extend or replace attribute
	// conversion for the whole program.
	AsAttribs = DefaultAsAttribs

	// AsComponent converts a value passed as a child into a Component. It
	// defaults to DefaultAsComponent — see its docs for the recognized types and
	// the github.com/domonda/go-pretty fallback used to render an unexpected
	// value as escaped text. Element and component constructors route their
	// children through AsComponent, so assigning a different func changes child
	// conversion everywhere. (Attribute conversion is separate — see AsAttribs.)
	//
	// For example, to turn the silent "unexpected value becomes text" fallback
	// into a hard failure during development (widen the accepted set to taste):
	//
	//	base := mx.AsComponent
	//	mx.AsComponent = func(c any) mx.Component {
	//		switch c.(type) {
	//		case nil, mx.Component, string:
	//			return base(c)
	//		default:
	//			panic(fmt.Sprintf("mx: unexpected child of type %T", c))
	//		}
	//	}
	//
	// Other uses: wrap the default to log unexpected types, or replace the
	// pretty representation with fmt.Sprint or a custom pretty.Printer.
	AsComponent = DefaultAsComponent
)
