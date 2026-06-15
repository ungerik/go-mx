package xml

import (
	"context"
	"strings"

	"github.com/domonda/go-errs"
	"github.com/ungerik/go-mx"
)

// Comment is an XML comment rendered as <!-- text -->.
//
// XML forbids the string "--" inside a comment (which also rules out a comment
// that would close early with "-->"); a Comment containing "--" therefore fails
// to render with an error instead of emitting malformed markup. The text is
// written verbatim — comments are not parsed, so no escaping is applied. When
// rendered with an indenting writer the comment starts on its own line.
type Comment string

var _ mx.Component = Comment("")

// Render writes the comment, implementing [mx.Component]. It returns an error if
// the text contains the forbidden "--" sequence.
func (c Comment) Render(_ context.Context, w mx.Writer) error {
	if strings.Contains(string(c), "--") {
		return errs.Errorf("xml.Comment must not contain %q: %q", "--", string(c))
	}
	return w.Comment(string(c))
}
