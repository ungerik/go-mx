package html

import (
	"io"
	"net/http"
	"strings"

	"github.com/ungerik/go-mx"
)

var contentTypeHTML = http.Header{"Content-Type": []string{mx.ContentTypeHTML}}

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
