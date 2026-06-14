package html

import (
	"context"
	"text/template"

	"github.com/ungerik/go-mx"
)

var _ mx.Component = Template{}

// Template is a Component that renders a Go text/template parsed from
// files matching the File glob pattern, executed with Data. It implements
// mx.Component so templates can be composed with the rest of the package.
type Template struct {
	File string
	Data any
	// Funcs is a map of functions that can be called from the template.
	// For a nice collection of third party functions see:
	// https://masterminds.github.io/sprig/
	Funcs template.FuncMap
}

// Render parses the templates matching t.File and writes the result of
// executing them with t.Data to w, implementing the mx.Component interface.
func (t Template) Render(ctx context.Context, w mx.Writer) error {
	templ, err := template.New("").Funcs(t.Funcs).ParseGlob(t.File)
	if err != nil {
		return err
	}
	return templ.Execute(w, t.Data)
}
