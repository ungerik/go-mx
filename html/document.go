package html

import (
	"context"
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
	return mx.Components{
		Raw("<!DOCTYPE html>\n<html>"),
		Head(
			Meta(Charset("UTF-8")),
			If(html.Title != "", TitleElem(html.Title)),
			ForEachSlice(slices.Sorted(maps.Keys(html.Meta)),
				func(name string) *Element {
					return Meta(Name(name), ContentAttr(html.Meta[name]))
				},
			),
			ForEachSlice(slices.Sorted(maps.Keys(html.MetaProperty)),
				func(property string) *Element {
					return Meta(mx.Attrib{Name: property, Value: html.MetaProperty[property]})
				},
			),
			ForEachSlice(html.Stylesheets,
				func(href string) *Element {
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
