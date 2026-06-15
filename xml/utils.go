package xml

import (
	"context"
	"io"
	"strings"

	"github.com/ungerik/go-mx"
)

// xmlEscaper replaces the five characters with predefined XML entity references.
var xmlEscaper = strings.NewReplacer(
	`&`, "&amp;",
	`<`, "&lt;",
	`>`, "&gt;",
	`"`, "&quot;",
	`'`, "&apos;",
)

// Escape returns s with the XML special characters &, <, >, " and ' replaced by
// their predefined entity references so the result is safe to embed as XML
// character data or an attribute value.
func Escape[S ~string](s S) string {
	return xmlEscaper.Replace(string(s))
}

// WriteEscaped writes s to w with the XML special characters &, <, >, " and '
// replaced by their predefined entity references.
func WriteEscaped[S ~string](w io.Writer, s S) error {
	_, err := xmlEscaper.WriteString(w, string(s))
	return err
}

// WriteRaw writes s to w verbatim without any XML escaping; the caller is
// responsible for ensuring s is well-formed XML.
func WriteRaw[S ~string](w io.Writer, s S) error {
	_, err := io.WriteString(w, string(s))
	return err
}

// String renders component to an XML string using a default (double-quoting)
// [mx.CheckedWriter], returning any render error. It is a convenience for tests
// and one-off rendering; use [Serve] or [Document.HandleHTTP] for indented
// output over HTTP.
func String(component mx.Component) (string, error) {
	var b strings.Builder
	if err := component.Render(context.Background(), mx.NewCheckedWriter(&b)); err != nil {
		return "", err
	}
	return b.String(), nil
}
