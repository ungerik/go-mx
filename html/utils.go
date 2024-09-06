package html

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/ungerik/go-mx"
)

var httpHeaderContentTypeHTML = http.Header{"Content-Type": []string{"text/html; charset=utf-8"}}

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

var attribEscaper = strings.NewReplacer(
	`&`, "&amp;",
	`<`, "&lt;",
	`"`, "&quot;",
)

func QuoteAttribute(value string) string {
	return `"` + attribEscaper.Replace(value) + `"`
}

func WriteStructAsStartTagWithAttribs(ctx context.Context, w io.Writer, elem string, s any) error {
	renderer := RendererFromContext(ctx)
	err := renderer.OpenElement(w, elem)
	if err != nil {
		return err
	}
	for field, val := range mx.FlatExportedStructFields(reflect.ValueOf(s)) {
		if a, _ := val.Interface().(Attribs); a != nil {
			for name, value := range a.Iter() {
				err = renderer.Attribute(w, name, value)
				if err != nil {
					return err
				}
			}
			continue
		}
		name := field.Tag.Get("attr")
		if name == "" || name == "-" {
			continue
		}
		value := fmt.Sprint(val.Interface())
		if value == "" {
			continue
		}
		err = renderer.Attribute(w, name, value)
		if err != nil {
			return err
		}
	}
	return renderer.CloseElement(w)
}
