package xml

import (
	"context"

	"github.com/ungerik/go-mx"
)

// CDATA is a CDATA section rendered as <![CDATA[text]]>.
//
// The text inside a CDATA section is not parsed, so characters that would
// otherwise need escaping (&, <, >) appear verbatim — this is the convenient
// way to embed a block of markup, code or other text without escaping each
// character. The one sequence a CDATA section cannot contain is its own
// terminator "]]>"; text containing it fails to render with an error.
type CDATA string

var _ mx.Component = CDATA("")

// Render writes the CDATA section, implementing [mx.Component]. It returns an
// error if the text contains the section terminator "]]>".
func (d CDATA) Render(_ context.Context, w mx.Writer) error {
	return w.CDATA(string(d))
}
