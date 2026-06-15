package web

import (
	"context"

	"github.com/domonda/go-errs"
	"github.com/ungerik/go-mx/html"
)

var _ PageRenderer = &PageTypeRenderer{}

// PageTypeRenderer is a PageRenderer that dispatches to a per-Page.Type
// PageRenderer and a per-Page.ContentType ContentRenderer. When the selected
// page renderer produces a document without a body, the matching content
// renderer is used to render Page.Content into the body.
type PageTypeRenderer struct {
	DefaultRenderer      PageRenderer
	PageTypeRenderers    map[string]PageRenderer
	ContentTypeRenderers map[string]ContentRenderer
}

// RenderPage selects a PageRenderer by Page.Type (falling back to
// DefaultRenderer or DefaultPageRenderer), then, if the resulting document has
// no body and the page has content, renders Page.Content via the
// ContentRenderer registered for Page.ContentType. It implements the
// PageRenderer interface.
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
