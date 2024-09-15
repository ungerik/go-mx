package web

import (
	"context"

	"github.com/domonda/go-errs"
	"github.com/ungerik/go-mx/html"
)

var DefaultPageRenderer = PageRendererFunc(func(ctx context.Context, page *Page) (html.Document, error) {
	return html.Document{
		Title: page.Title,
		Body:  nil,
	}, nil
})

var _ PageRenderer = &PageTypeRenderer{}

type PageTypeRenderer struct {
	DefaultRenderer      PageRenderer
	PageTypeRenderers    map[string]PageRenderer
	ContentTypeRenderers map[string]ContentRenderer
}

func (t *PageTypeRenderer) RenderPage(ctx context.Context, page *Page) (html.Document, error) {
	r := t.PageTypeRenderers[page.Type]
	if r == nil {
		if t.DefaultRenderer != nil {
			r = t.DefaultRenderer
		} else {
			r = DefaultPageRenderer
		}
	}
	doc, err := r.RenderPage(ctx, page)
	if err != nil {
		return html.Document{}, err
	}
	if doc.Body != nil || page.Content == nil {
		return doc, nil
	}
	contentTempl := t.ContentTypeRenderers[page.ContentType]
	if contentTempl == nil {
		return html.Document{}, errs.Errorf("no ContentRenderer for Page.ContentType %q", page.ContentType)
	}
	doc.Body, err = contentTempl.RenderContent(ctx, page.ContentType, page.Content)
	if err != nil {
		return html.Document{}, err
	}
	return doc, nil
}
