package html

import (
	"context"
	"io"
	"net/http"
	"text/template"

	"github.com/ungerik/go-mx"
)

var _ mx.Component = Template{}

type Template struct {
	File string
	Data any
	// Funcs is a map of functions that can be called from the template.
	// For a nice collection of third party functions see:
	// https://masterminds.github.io/sprig/
	Funcs template.FuncMap
}

func (t Template) Render(ctx context.Context, w io.Writer) error {
	templ, err := template.New("").Funcs(t.Funcs).ParseGlob(t.File)
	if err != nil {
		return err
	}
	return templ.Execute(w, t.Data)
}

func (t Template) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mx.ServeComponent(w, r, contentTypeHTML, t)
}
