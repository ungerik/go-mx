package html

import (
	"bytes"
	"context"
	"maps"
	"net/http"
	"slices"

	"github.com/ungerik/go-mx"
)

// DOCTYPE is the HTML5 document type declaration `<!DOCTYPE html>` that
// must precede the root <html> element of a full document.
const DOCTYPE Raw = `<!DOCTYPE html>`

// Document is a Component for a complete HTML5 page. It renders the
// DOCTYPE, an <html> root, a <head> assembled from the metadata fields,
// and a <body> holding the Body component.
type Document struct {
	Title        string
	Meta         map[string]string // name -> content
	MetaProperty map[string]string // property -> content
	Stylesheets  []string          // href for link rel="stylesheet"
	Style        string            // inline style after stylesheets
	HeadCustom   mx.Component      // Custom head content after all other head content
	Body         mx.Component
}

// NewDocument returns a Document with the given <title> and the body
// arguments converted to its Body component via mx.AsComponents.
func NewDocument(title string, body ...any) *Document {
	return &Document{
		Title: title,
		Body:  mx.AsComponents(body...),
	}
}

// Render writes the complete HTML page to w, implementing the
// mx.Component interface.
func (d *Document) Render(ctx context.Context, w mx.Writer) error {
	return mx.Components{
		DOCTYPE,
		Raw("\n<html>"),
		Head(
			Meta(CharSet("UTF-8")),
			If(d.Title != "", TitleElem(d.Title)),
			ForEach(slices.Sorted(maps.Keys(d.Meta)),
				func(name string) *mx.Element {
					return Meta(Name(name), ContentAttr(d.Meta[name]))
				},
			),
			ForEach(slices.Sorted(maps.Keys(d.MetaProperty)),
				func(property string) *mx.Element {
					return Meta(Attrib(property, d.MetaProperty[property]))
				},
			),
			ForEach(d.Stylesheets,
				func(href string) *mx.Element {
					return Link(Rel("stylesheet"), HRef(href))
				},
			),
			If(d.Style != "", StyleElem(d.Style)),
			d.HeadCustom,
		),
		Body(
			d.Body,
		),
		Raw("\n</html>"),
	}.Render(ctx, w)
}

// Serve starts an HTTP server on addr that responds with this document
// for every request. It blocks until the server stops.
func (d *Document) Serve(addr string) error {
	return Serve(addr, d)
}

// HandleHTTP renders the document and writes it as an HTML response,
// implementing http.HandlerFunc. On a render error it responds with a
// generic 500 status via mx.RespondNonContextError.
func (d *Document) HandleHTTP(response http.ResponseWriter, request *http.Request) {
	buf := bytes.NewBuffer(nil)
	writer := mx.NewCheckedWriter(buf).WithIndent("", "  ")
	err := d.Render(request.Context(), writer)
	if err != nil {
		mx.RespondNonContextError(response, err)
		return
	}
	response.Header().Set("Content-Type", mx.ContentTypeHTML)
	response.Write(buf.Bytes())
}
