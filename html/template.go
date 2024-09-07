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
}

func (t Template) Render(ctx context.Context, w io.Writer) error {
	templ, err := template.New("").ParseGlob(t.File)
	if err != nil {
		return err
	}
	return templ.Execute(w, t.Data)
}

func (t Template) GetChildren(ctx context.Context) ([]mx.Component, error) { return nil, nil }

func (t Template) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mx.ServeHTTP(w, r, httpHeaderContentTypeHTML, t)
}
