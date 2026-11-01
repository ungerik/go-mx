package logview

import (
	"context"
	"strings"
)

// Matches reports whether the line raw, as [Config.Line] renders it, contains
// query, ignoring case and the query's surrounding white space. An empty query
// matches every line.
//
// It is the test the view's filter applies in the browser, so a handler that
// answers [Config.SearchParam] with it sends exactly the lines the filter
// would have shown. That is why the line is matched as rendered rather than
// as raw: the reader types what they see, which is "status=200" for a record
// that carries "status":200, and never sees the escape sequences or the JSON
// quoting of the raw line.
//
// It renders the line to find out, so it costs about what rendering the line
// does. A handler searching a large store should narrow the candidates with
// its own index first and use Matches to decide the rest.
func (c *Config) Matches(raw, query string) bool {
	query = strings.TrimSpace(query)
	if query == "" {
		return true
	}
	var text textWriter
	if err := c.lineElement(raw).Render(context.Background(), &text); err != nil {
		// A line that cannot render cannot be shown, so it matches nothing.
		return false
	}
	return strings.Contains(strings.ToLower(text.String()), strings.ToLower(query))
}

// textWriter is an [mx.Writer] that keeps only the text content of what is
// rendered into it, the server-side counterpart of a line's textContent in the
// browser. Markup is dropped, and so is raw output, which a line never
// produces and which would be markup if it did.
type textWriter struct {
	text strings.Builder
}

func (t *textWriter) Write(p []byte) (int, error)    { return len(p), nil }
func (t *textWriter) BeginElement(string) error      { return nil }
func (t *textWriter) Attribute(string, string) error { return nil }
func (t *textWriter) CloseElementStartTag() error    { return nil }
func (t *textWriter) EscapeText(text string) error   { t.text.WriteString(text); return nil }
func (t *textWriter) EndElement() error              { return nil }
func (t *textWriter) Comment(string) error           { return nil }
func (t *textWriter) CDATA(text string) error        { t.text.WriteString(text); return nil }
func (t *textWriter) Newline() error                 { return nil }
func (t *textWriter) String() string                 { return t.text.String() }
