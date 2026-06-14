package html

import (
	"io"
	"strings"
)

var htmlEscaper = strings.NewReplacer(
	`&`, "&amp;",
	`'`, "&apos;",
	`<`, "&lt;",
	`>`, "&gt;",
	`"`, "&quot;",
)

// Escape returns s with the HTML special characters &, ', <, > and "
// replaced by their character references so the result is safe to embed
// as HTML text or an attribute value.
func Escape[S ~string](s S) string {
	return htmlEscaper.Replace(string(s))
}

// WriteEscaped writes s to w with the HTML special characters &, ', <, >
// and " replaced by their character references.
func WriteEscaped[S ~string](w io.Writer, s S) error {
	_, err := htmlEscaper.WriteString(w, string(s))
	return err
}

// WriteRaw writes s to w verbatim without any HTML escaping; the caller
// is responsible for ensuring s is safe HTML.
func WriteRaw[S ~string](w io.Writer, s S) error {
	_, err := io.WriteString(w, string(s))
	return err
}
