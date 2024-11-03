package html

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/ungerik/go-mx"
)

type Document struct {
	Title        string
	Meta         map[string]string // name -> content
	MetaProperty map[string]string // property -> content
	Stylesheets  []string          // href for link rel="stylesheet"
	Style        string            // inline style after stylesheets
	HeadCustom   mx.Component      // Custom head content after all other head content
	Body         mx.Component
}

func (html Document) Render(ctx context.Context, w mx.Writer) error {
	_, err := fmt.Fprint(w, "<!DOCTYPE html>\n<html>\n<head>\n")
	if err != nil {
		return err
	}
	err = Meta(Charset("UTF-8")).Render(ctx, w)
	if err != nil {
		return err
	}
	if html.Title != "" {
		err := TitleElem(html.Title).Render(ctx, w)
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
	_, err = fmt.Fprint(w, "</body>\n</html>")
	return err
}
