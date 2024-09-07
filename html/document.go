package html

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/ungerik/go-mx"
)

var _ mx.Component = Document{}

type Document struct {
	Title string
	Body  mx.Component
}

func (html Document) Render(ctx context.Context, w io.Writer) error {
	_, err := fmt.Fprint(w, "<!DOCTYPE html>\n<html>\n<head>")
	if err != nil {
		return err
	}
	if html.Title != "" {
		_, err := fmt.Fprintf(w, "\n<title>%s</title>", Escape(html.Title))
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprint(w, "\n</head>\n<body>\n")
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
