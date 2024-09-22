package html

import (
	"context"
	"fmt"
	"io"
	"maps"
	"net/http"
	"slices"

	"github.com/ungerik/go-mx"
)

var _ mx.Component = Document{}

type Document struct {
	Title        string
	Meta         map[string]string // name -> content
	MetaProperty map[string]string // property -> content
	Stylesheets  []string          // href for link rel="stylesheet"
	Style        string            // inline style after stylesheets
	HeadCustom   mx.Component      // Custom head content after all other head content
	Body         mx.Component
}

func (html Document) Render(ctx context.Context, w io.Writer) error {
	_, err := fmt.Fprint(w, "<!DOCTYPE html>\n<html>\n<head>\n<meta charset='UTF-8'/>\n")
	if err != nil {
		return err
	}

	if html.Title != "" {
		_, err := fmt.Fprintf(w, "<title>%s</title>\n", Escape(html.Title))
		if err != nil {
			return err
		}
	}
	for _, name := range slices.Sorted(maps.Keys(html.Meta)) {
		content := html.Meta[name]
		_, err := fmt.Fprintf(w, "<meta name='%s' content='%s'/>\n", Escape(name), Escape(content))
		if err != nil {
			return err
		}
	}
	for _, property := range slices.Sorted(maps.Keys(html.MetaProperty)) {
		content := html.MetaProperty[property]
		_, err := fmt.Fprintf(w, "<meta property='%s' content='%s'/>\n", Escape(property), Escape(content))
		if err != nil {
			return err
		}
	}
	// <link rel="stylesheet" type="text/css" href="mystyle.css">
	for _, href := range html.Stylesheets {
		_, err := fmt.Fprintf(w, "<link rel='stylesheet' href='%s'/>\n", href)
		if err != nil {
			return err
		}
	}
	if html.Style != "" {
		_, err := fmt.Fprintf(w, "<style>%s</style>\n", html.Style)
		if err != nil {
			return err
		}
	}
	if html.HeadCustom != nil {
		err = html.HeadCustom.Render(ctx, w)
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprint(w, "</head>\n<body>\n")
	if err != nil {
		return err
	}
	if html.Body != nil {
		err = html.Body.Render(ctx, w)
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprint(w, "\n</body>\n</html>")
	return err
}

func (html Document) GetChildren(ctx context.Context) ([]mx.Component, error) {
	return mx.ComponentSlice(html.Body), ctx.Err()
}

func (html Document) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mx.ServeComponent(w, r, contentTypeHTML, html)
}
