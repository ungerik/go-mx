package xml

import (
	"bytes"
	"context"
	"net/http"

	"github.com/ungerik/go-mx"
)

// Document is a [mx.Component] for a complete XML document: an optional XML
// declaration, optional prolog content and a single root element.
//
// The components are rendered in order with no whitespace inserted between them.
// The [mx.CheckedWriter] recognizes the "?>" ending a declaration or processing
// instruction and breaks the line after it, so the prolog or root starts on the
// line after the declaration in both compact and indented output (an indenting
// writer further indents nested elements).
type Document struct {
	// Declaration is rendered first when non-nil. Use the [Declaration]
	// constant for the standard <?xml version="1.0" encoding="UTF-8"?> or
	// [Decl] for a custom version/encoding.
	Declaration mx.Component
	// Prolog holds optional content between the declaration and the root, such
	// as [Comment]s, [ProcInst]s or a [Doctype]. Pass several with mx.Components.
	Prolog mx.Component
	// Root is the document's single root element.
	Root mx.Component
}

// NewDocument returns a Document with the standard XML [Declaration] and the
// given root element. A zero Document value has no declaration.
func NewDocument(root mx.Component) *Document {
	return &Document{Declaration: Declaration, Root: root}
}

// Render writes the document, implementing [mx.Component].
func (d *Document) Render(ctx context.Context, w mx.Writer) error {
	var comps mx.Components
	if d.Declaration != nil {
		comps = append(comps, d.Declaration)
	}
	if d.Prolog != nil {
		comps = append(comps, d.Prolog)
	}
	if d.Root != nil {
		comps = append(comps, d.Root)
	}
	return comps.Render(ctx, w)
}

// Serve starts an HTTP server on addr that responds with this document for every
// request. It blocks until the server stops.
func (d *Document) Serve(addr string) error {
	return Serve(addr, d)
}

// HandleHTTP renders the document and writes it as an XML response,
// implementing http.HandlerFunc. On a render error it responds with a generic
// 500 status via [mx.RespondNonContextError].
func (d *Document) HandleHTTP(response http.ResponseWriter, request *http.Request) {
	serve(response, request, d)
}

// Serve starts an HTTP server on addr that renders component as the XML response
// for every request. It blocks until the server stops, returning the error from
// http.ListenAndServe.
func Serve(addr string, component mx.Component) error {
	return http.ListenAndServe(addr, http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		serve(response, request, component)
	}))
}

// serve renders component to an indented XML response, the shared body of
// [Serve] and [Document.HandleHTTP].
func serve(response http.ResponseWriter, request *http.Request, component mx.Component) {
	buf := bytes.NewBuffer(nil)
	writer := mx.NewCheckedWriter(buf).WithIndent("", "  ")
	if err := component.Render(request.Context(), writer); err != nil {
		mx.RespondNonContextError(response, err)
		return
	}
	response.Header().Set("Content-Type", mx.ContentTypeXML)
	response.Write(buf.Bytes())
}
