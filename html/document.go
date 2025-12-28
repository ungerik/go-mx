package html

import (
	"bytes"
	"context"
	"maps"
	"net/http"
	"slices"

	"github.com/ungerik/go-mx"
)

const DOCTYPE Raw = `<!DOCTYPE html>`

type Document struct {
	Title        string
	Meta         map[string]string // name -> content
	MetaProperty map[string]string // property -> content
	Stylesheets  []string          // href for link rel="stylesheet"
	Style        string            // inline style after stylesheets
	HeadCustom   mx.Component      // Custom head content after all other head content
	Body         mx.Component
}

func NewDocument(title string, body ...any) *Document {
	return &Document{
		Title: title,
		Body:  mx.AsComponents(body...),
	}
}

func (d *Document) Render(ctx context.Context, w mx.Writer) error {
	return mx.Components{
		DOCTYPE,
		Raw("\n<html>"),
		Head(
			Meta(CharSet("UTF-8")),
			If(d.Title != "", TitleElem(d.Title)),
			ForEach(slices.Sorted(maps.Keys(d.Meta)),
				func(name string) *mx.Element {
					return Meta(Name(name), ContentAttr(d.Meta[name]))
				},
			),
			ForEach(slices.Sorted(maps.Keys(d.MetaProperty)),
				func(property string) *mx.Element {
					return Meta(Attrib(property, d.MetaProperty[property]))
				},
			),
			ForEach(d.Stylesheets,
				func(href string) *mx.Element {
					return Link(Rel("stylesheet"), HRef(href))
				},
			),
			If(d.Style != "", StyleElem(d.Style)),
			d.HeadCustom,
		),
		Body(
			d.Body,
		),
		Raw("\n</html>"),
	}.Render(ctx, w)
}

func (d *Document) Serve(addr string) error {
	return Serve(addr, d)
}

func (d *Document) HandleHTTP(response http.ResponseWriter, request *http.Request) {
	buf := bytes.NewBuffer(nil)
	writer := mx.NewCheckedWriter(buf).WithIndent("", "  ")
	err := d.Render(request.Context(), writer)
	if err != nil {
		mx.RespondNonContextError(response, err)
		return
	}
	response.Header().Set("Content-Type", mx.ContentTypeHTML)
	response.Write(buf.Bytes())
}
