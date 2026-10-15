package html

// https://json-ld.org/spec/latest/json-ld/#embedding-json-ld-in-html-documents
// https://developers.google.com/search/docs/appearance/structured-data/intro-structured-data

import (
	"encoding/json"

	"github.com/domonda/go-errs"
	"github.com/ungerik/go-mx"
)

// JSONLDType is the MIME type of JSON-LD. As the type attribute of a <script>
// it tells the browser the content is a data block, not code to run, and tells
// crawlers to read it as linked data.
const JSONLDType = "application/ld+json"

// JSONLD is a <script type="application/ld+json"> holding v marshaled with
// [json.Marshal], the way structured data such as a schema.org description of
// the page is embedded into HTML. It belongs in the document head, e.g. as
// [Document.HeadCustom].
//
// Unlike [ScriptJS] the content is safe for untrusted values. A script body
// cannot be HTML-escaped, so the JSON goes in as [mx.Raw], but json.Marshal
// escapes "<", ">" and "&" as \u003c, \u003e and \u0026 inside strings, the
// only place JSON can hold them, and does so for the output of a
// [json.Marshaler] too. No "</script" or "<!--" can therefore reach the markup
// to end the element early.
//
// v is marshaled right away, so the returned element can be checked through
// its Err field before rendering. A marshal error, a nil v and a v that
// marshals to null (a nil pointer, map or slice) are deferred to render time
// (see [mx.NewErrElement]): a script holding null describes nothing and is
// almost certainly a bug.
func JSONLD(v any) *mx.Element {
	if v == nil {
		return mx.NewErrElement(errs.New("html.JSONLD: nil value"))
	}
	data, err := json.Marshal(v)
	if err != nil {
		return mx.NewErrElement(errs.Errorf("html.JSONLD: %w", err))
	}
	if string(data) == "null" {
		return mx.NewErrElement(errs.Errorf("html.JSONLD: %T value marshals to null", v))
	}
	return &mx.Element{
		Name:     "script",
		Attribs:  []mx.Attrib{Type(JSONLDType)},
		Children: []mx.Component{mx.Raw(data)},
	}
}
