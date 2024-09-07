package web

import (
	"context"

	"github.com/domonda/go-errs"
	"github.com/ungerik/go-mx/html"
)

var DefaultPageTemplate = PageTemplateFunc(func(ctx context.Context, page *Page) (html.Document, error) {
	return html.Document{
		Title: page.Title,
		Body:  nil,
	}, nil
})

var _ PageTemplate = &MultiPageTemplate{}

type MultiPageTemplate struct {
	DefaultTemplate      PageTemplate
	PageTypeTemplates    map[string]PageTemplate
	ContentTypeTemplates map[string]ContentTemplate
}

func (t *MultiPageTemplate) TransformPage(ctx context.Context, page *Page) (html.Document, error) {
	templ := t.PageTypeTemplates[page.Type]
	if templ == nil {
		if t.DefaultTemplate != nil {
			templ = t.DefaultTemplate
		} else {
			templ = DefaultPageTemplate
		}
	}
	doc, err := templ.TransformPage(ctx, page)
	if err != nil {
		return html.Document{}, err
	}
	if doc.Body != nil || page.Content == nil {
		return doc, nil
	}
	contentTempl := t.ContentTypeTemplates[page.ContentType]
	if contentTempl == nil {
		return html.Document{}, errs.Errorf("no ContentTemplate for Page.ContentType %q", page.ContentType)
	}
	doc.Body, err = contentTempl.TransformContent(ctx, page.ContentType, page.Content)
	if err != nil {
		return html.Document{}, err
	}
	return doc, nil
}
