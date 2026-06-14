package highlight

import (
	"go/format"
	"strconv"
	"strings"
)

// GoSource highlights Go source and returns, instead of HTML, the Go source
// code that builds that highlighted markup with the html package. It is a
// generator: the returned code is not the input echoed back but a tree of
// html.Pre / html.Code / html.Span calls.
//
// For the input
//
//	func main() {}
//
// it returns roughly
//
//	html.Pre(html.Class("hl"),
//		html.Code(
//			html.Span(html.Class("hl-keyword"), "func"),
//			" ",
//			html.Span(html.Class("hl-function"), "main"),
//			html.Span(html.Class("hl-punctuation"), "()"),
//			" ",
//			html.Span(html.Class("hl-punctuation"), "{}"),
//		),
//	)
//
// The result is gofmt-formatted. If the generated expression somehow fails to
// format, the unformatted but valid expression is returned instead.
func (h *Highlighter) GoSource(src string) string {
	expr := h.goSourceExpr(TokenizeGo(src))

	// format.Source accepts a list of statements, so wrap the expression in a
	// throwaway assignment, format it, then strip the wrapper back off.
	formatted, err := format.Source([]byte("_ = " + expr))
	if err != nil {
		return expr
	}
	return strings.TrimPrefix(string(formatted), "_ = ")
}

// goSourceExpr builds the unformatted html.Pre(...) expression. Every child is
// placed on its own line with a trailing comma so gofmt produces the indented,
// multi-line layout (gofmt never breaks a single line into several itself).
func (h *Highlighter) goSourceExpr(tokens []Token) string {
	prefix := h.prefix()
	var b strings.Builder
	b.WriteString("html.Pre(html.Class(")
	b.WriteString(strconv.Quote(h.BlockClass()))
	b.WriteString("),\nhtml.Code(\n")
	for _, it := range h.items(tokens) {
		if it.Class == ClassPlain {
			b.WriteString(strconv.Quote(it.Text))
		} else {
			b.WriteString("html.Span(html.Class(")
			b.WriteString(strconv.Quote(prefix + string(it.Class)))
			b.WriteString("), ")
			b.WriteString(strconv.Quote(it.Text))
			b.WriteString(")")
		}
		b.WriteString(",\n")
	}
	b.WriteString("),\n)")
	return b.String()
}

// GoSource highlights Go source with the [Default] highlighter. See
// [Highlighter.GoSource].
func GoSource(src string) string { return Default.GoSource(src) }
