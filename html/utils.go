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

func Escape[S ~string](s S) string {
	return htmlEscaper.Replace(string(s))
}

func WriteEscaped[S ~string](w io.Writer, s S) error {
	_, err := htmlEscaper.WriteString(w, string(s))
	return err
}

func WriteRaw[S ~string](w io.Writer, s S) error {
	_, err := io.WriteString(w, string(s))
	return err
}
