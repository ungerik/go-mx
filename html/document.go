package html

import (
	"context"
	"maps"
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

func (html *Document) Render(ctx context.Context, w mx.Writer) error {
	return mx.Components{
		DOCTYPE,
		Raw("\n<html>"),
		Head(
			Meta(CharSet("UTF-8")),
			If(html.Title != "", TitleElem(html.Title)),
			ForEachSlice(slices.Sorted(maps.Keys(html.Meta)),
				func(name string) *mx.Element {
					return Meta(Name(name), ContentAttr(html.Meta[name]))
				},
			),
			ForEachSlice(slices.Sorted(maps.Keys(html.MetaProperty)),
				func(property string) *mx.Element {
					return Meta(Attrib(property, html.MetaProperty[property]))
				},
			),
			ForEachSlice(html.Stylesheets,
				func(href string) *mx.Element {
					return Link(Rel("stylesheet"), HRef(href))
				},
			),
			If(html.Style != "", StyleElem(html.Style)),
			html.HeadCustom,
		),
		Body(
			html.Body,
		),
		Raw("\n</html>"),
	}.Render(ctx, w)
}
