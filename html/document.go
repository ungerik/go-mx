package html

import (
	"bytes"
	"context"
	"maps"
	"net/http"
	"regexp"
	"slices"

	"github.com/domonda/go-errs"
	"github.com/ungerik/go-mx"
)

// DOCTYPE is the HTML5 document type declaration `<!DOCTYPE html>` that
// must precede the root <html> element of a full document.
const DOCTYPE Raw = `<!DOCTYPE html>`

// Document is a Component for a complete HTML5 page. It renders the
// DOCTYPE, an <html> root, a <head> assembled from the metadata fields,
// and a <body> holding the Body component.
type Document struct {
	// Lang sets the lang attribute of the <html> root, the language of
	// the whole page. The empty string omits the attribute, rendering the
	// bare <html> tag as before. Setting it is the fix for WCAG 3.1.1
	// (Language of Page): a screen reader needs it to pick pronunciation
	// rules and browsers need it to offer translation. The value must be a
	// BCP 47 tag charset ([A-Za-z]{1,8}(-[A-Za-z0-9]{1,8})*, e.g. "de" or
	// "de-AT"); an invalid value is rejected by Render rather than escaped
	// into markup, so a caller mistake surfaces instead of silently
	// mangling the tag.
	Lang         string
	Title        string
	Meta         map[string]string // name -> content
	MetaProperty map[string]string // property -> content
	Stylesheets  []string          // href for link rel="stylesheet"
	Style        string            // inline style after stylesheets
	HeadCustom   mx.Component      // Custom head content after all other head content
	Body         mx.Component
}

// langTagRegexp matches the BCP 47 language-tag charset
// [A-Za-z]{1,8}(-[A-Za-z0-9]{1,8})*. It is intentionally a charset check,
// not a registry validation: a tag matching it can only contain ASCII
// letters, digits and hyphens, so it is injection-proof by construction
// and cannot break out of the double-quoted lang attribute value.
var langTagRegexp = regexp.MustCompile(`^[A-Za-z]{1,8}(-[A-Za-z0-9]{1,8})*$`)

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
	// Build the root open tag. An empty Lang renders the bare <html> tag
	// byte-for-byte as before; a set Lang is validated against the BCP 47
	// charset and, once matched, is safe to interpolate directly (no
	// escaping needed) because the charset admits no attribute-breaking
	// characters. An invalid Lang aborts the render rather than emitting
	// broken or unescaped markup.
	htmlOpen := Raw("\n<html>")
	if d.Lang != "" {
		if !langTagRegexp.MatchString(d.Lang) {
			return errs.Errorf("html.Document.Lang %q is not a valid BCP 47 language tag", d.Lang)
		}
		htmlOpen = Raw("\n<html lang=\"" + d.Lang + "\">")
	}
	return mx.Components{
		DOCTYPE,
		htmlOpen,
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
					return Meta(Attrib("property", property), ContentAttr(d.MetaProperty[property]))
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
