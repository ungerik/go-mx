package html

import (
	"bytes"
	"context"
	"maps"
	"net/http"
	"regexp"
	"slices"
	"strings"

	"github.com/domonda/go-errs"

	"github.com/ungerik/go-mx"
)

// DOCTYPE is the HTML5 document type declaration `<!DOCTYPE html>` that
// must precede the root <html> element of a full document.
const DOCTYPE Raw = `<!DOCTYPE html>`

// Document is a Component for a complete HTML5 page. It renders the
// DOCTYPE, an <html> root, a <head> assembled from the metadata fields,
// and a <body> holding BodyChildren. The head always starts with
// <meta charset="UTF-8">, followed by the fields in their declaration
// order, from Title to HeadCustom.
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
	Lang string
	// HTMLAttribs are extra attributes of the <html> root, rendered after
	// lang in the order given, for example Class("no-js") or a data-
	// attribute a theme script reads. They are written by a copy of the
	// CheckedWriter the page renders to (a default CheckedWriter for any
	// other Writer), so names are checked and values escaped and quoted
	// as on any element; an empty value renders as name="". An
	// attribute named "lang" (in any letter case) is refused by Render with
	// an error, because Lang is the one way to set it; so are a nil entry
	// and a duplicate name, also one differing only in letter case.
	HTMLAttribs []mx.Attrib

	// Title is the text of the <title> element, escaped as any text. The
	// empty string omits the element.
	Title string
	// Meta holds the <meta name="…" content="…"> elements of the head,
	// keyed by name, for example "description" or "viewport". They are
	// rendered sorted by name, so the output does not depend on map order,
	// and both name and content are escaped as attribute values.
	Meta map[string]string
	// MetaProperty holds the <meta property="…" content="…"> elements of
	// the head, keyed by property, as Open Graph uses them ("og:title",
	// "og:url"). They follow Meta, sorted by property, escaped as Meta.
	MetaProperty map[string]string
	// Stylesheets are the href URLs of the <link rel="stylesheet">
	// elements, rendered in the order given after the meta elements,
	// so a later stylesheet overrides an earlier one in the cascade.
	Stylesheets []string
	// Style is CSS for an inline <style> element after Stylesheets, so it
	// overrides them. The empty string omits the element. It is written
	// unescaped, as CSS must be, so it must not come from untrusted input:
	// a "</style>" inside it would end the element.
	Style string
	// HeadCustom is any further head content, rendered last in <head>,
	// after Style: scripts, JSON-LD, favicons, canonical links and the
	// like. Several components can be passed as mx.Components; nil adds
	// nothing.
	HeadCustom mx.Component

	// BodyAttribs are the attributes of <body>, for example the classes,
	// id and data- attributes a site theme's CSS selects on. They are
	// rendered through the normal element path, exactly as the attributes
	// of any other element. Render refuses a nil entry, an invalid name and
	// a duplicate name, also one differing only in letter case, before
	// writing anything.
	BodyAttribs []mx.Attrib
	// BodyChildren are the children of <body>, the content of the page,
	// as BodyAttribs are its attributes: Render builds the <body> element
	// from the two, so BodyChildren must not be a Body element, which would
	// render a second <body> nested in the first. Several children are
	// passed as mx.Components, or as the body arguments of NewDocument. A
	// nil BodyChildren renders an empty <body>.
	BodyChildren mx.Component
}

// langTagRegexp matches the BCP 47 language-tag charset
// [A-Za-z]{1,8}(-[A-Za-z0-9]{1,8})*. It is intentionally a charset check,
// not a registry validation: a tag matching it can only contain ASCII
// letters, digits and hyphens, so it is injection-proof by construction
// and cannot break out of the double-quoted lang attribute value.
var langTagRegexp = regexp.MustCompile(`^[A-Za-z]{1,8}(-[A-Za-z0-9]{1,8})*$`)

// ValidLangTag reports whether tag matches the BCP 47 language-tag charset,
// the check Document.Lang is rendered with. Every tag passing it is safe to
// put into an attribute value or a quoted header parameter as is.
func ValidLangTag(tag string) bool {
	return langTagRegexp.MatchString(tag)
}

// NewDocument returns a Document with the given <title> and the body
// arguments converted to its BodyChildren via mx.AsComponents.
func NewDocument(title string, body ...any) *Document {
	return &Document{
		Title:        title,
		BodyChildren: mx.AsComponents(body...),
	}
}

// Render writes the complete HTML page to w, implementing the
// mx.Component interface. An invalid Lang, any invalid HTMLAttribs entry
// and a nil, invalid or duplicate BodyAttribs name return an error before
// anything is written. A BodyAttribs value that fails returns its error
// when <body> is written, after the <head>, as on any element: values can
// depend on ctx, so they are computed only once, while writing.
func (d *Document) Render(ctx context.Context, w mx.Writer) error {
	err := checkAttribNames("BodyAttribs", d.BodyAttribs)
	if err != nil {
		return err
	}
	htmlOpen, err := d.htmlStartTag(ctx, w)
	if err != nil {
		return err
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
		// Built directly rather than with Body(…), which would sort its
		// arguments into attributes and children and so turn a BodyChildren
		// that also converts to attributes (an mx.AttribProvider, a struct
		// with attr tags) into attributes of <body>.
		&mx.Element{
			Name:     "body",
			Attribs:  d.BodyAttribs,
			Children: mx.Components{d.BodyChildren},
		},
		Raw("\n</html>"),
	}.Render(ctx, w)
}

// htmlStartTag returns the <html> start tag as Raw, so the element writer
// does not close it: the matching "</html>" is written separately.
//
// Without HTMLAttribs the tag is a constant, "<html>" or, for a Lang that
// passes ValidLangTag (whose charset admits no attribute-breaking
// characters), "<html lang=\"…\">", byte-for-byte as before the field
// existed. With HTMLAttribs the attributes are written by a scratch copy of
// w's own CheckedWriter (a default one if w is another Writer), so the
// root's attributes get the same name check, value escaping and quote style
// as those of every other element; only indentation is turned off, because
// the tag is the root and the indentation of the elements in it is w's.
func (d *Document) htmlStartTag(ctx context.Context, w mx.Writer) (Raw, error) {
	if d.Lang != "" && !ValidLangTag(d.Lang) {
		return "", errs.Errorf("html.Document.Lang %q is not a valid BCP 47 language tag", d.Lang)
	}
	if len(d.HTMLAttribs) == 0 {
		if d.Lang == "" {
			return "\n<html>", nil
		}
		return Raw("\n<html lang=\"" + d.Lang + "\">"), nil
	}

	err := checkAttribNames("HTMLAttribs", d.HTMLAttribs)
	if err != nil {
		return "", err
	}
	for i, a := range d.HTMLAttribs {
		// HTML attribute names are case-insensitive, so "LANG" would
		// give the root a second lang attribute just like "lang".
		if strings.EqualFold(a.AttribName(), "lang") {
			return "", errs.Errorf("html.Document.HTMLAttribs[%d] sets %q, use Document.Lang instead", i, a.AttribName())
		}
	}

	var b strings.Builder
	b.WriteByte('\n')
	var tw *mx.CheckedWriter
	if cw, ok := w.(*mx.CheckedWriter); ok {
		tw = cw.Clone(&b).WithIndent("", "")
	} else {
		tw = mx.NewCheckedWriter(&b)
	}
	err = tw.BeginElement("html")
	if err != nil {
		return "", err
	}
	if d.Lang != "" {
		err = tw.Attribute("lang", d.Lang)
		if err != nil {
			return "", err
		}
	}
	for _, a := range d.HTMLAttribs {
		value, err := a.AttribValue(ctx)
		if err != nil {
			return "", err
		}
		err = tw.Attribute(a.AttribName(), value)
		if err != nil {
			return "", err
		}
	}
	err = tw.CloseElementStartTag()
	if err != nil {
		return "", err
	}
	return Raw(b.String()), nil
}

// checkAttribNames returns an error naming the Document field if attribs
// holds a nil Attrib, a name mx.ValidAttribName rejects, or two names that
// differ only in letter case. The writer refuses invalid names and exact
// duplicates too, but only once it writes the element, after the <head>;
// checked here, they stop Render before anything is written. HTML attribute
// names are case-insensitive, so "class" and "CLASS" are two class
// attributes to a browser.
func checkAttribNames(field string, attribs []mx.Attrib) error {
	for i, a := range attribs {
		if a == nil {
			return errs.Errorf("html.Document.%s[%d] is nil", field, i)
		}
		if !mx.ValidAttribName(a.AttribName()) {
			return errs.Errorf("html.Document.%s[%d] has the invalid attribute name %q", field, i, a.AttribName())
		}
		for _, prev := range attribs[:i] {
			if strings.EqualFold(prev.AttribName(), a.AttribName()) {
				return errs.Errorf("duplicate attribute %s in html.Document.%s", a.AttribName(), field)
			}
		}
	}
	return nil
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
		mx.RespondNonContextError(response, request, err)
		return
	}
	response.Header().Set("Content-Type", mx.ContentTypeHTML)
	response.Write(buf.Bytes())
}
