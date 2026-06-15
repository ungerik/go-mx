package web

import (
	"context"

	"github.com/ungerik/go-mx/html"
)

// PageRenderer renders a Page into a complete html.Document.
type PageRenderer interface {
	RenderPage(ctx context.Context, page *Page) (html.Document, error)
}

// PageRendererFunc is an adapter that lets an ordinary function be used as a PageRenderer.
type PageRendererFunc func(ctx context.Context, page *Page) (html.Document, error)

// RenderPage calls f and implements the PageRenderer interface.
func (f PageRendererFunc) RenderPage(ctx context.Context, page *Page) (html.Document, error) {
	return f(ctx, page)
}

// DefaultPageRenderer is the PageRenderer used when no other renderer is configured.
// It wraps DefaultRenderPage.
var DefaultPageRenderer PageRenderer = PageRendererFunc(DefaultRenderPage)

// DefaultRenderPage renders a Page into an html.Document containing only the
// document-level metadata derived from the page (title and robots meta),
// without rendering the page content into the document body.
func DefaultRenderPage(ctx context.Context, page *Page) (doc html.Document, err error) {
	// https://learn.microsoft.com/en-us/previous-versions/windows/internet-explorer/ie-developer/platform-apis/dn255024(v=vs.85)
	// https://ogp.me/

	doc.Title = page.Title
	if page.NoIndex {
		doc.Meta["robots"] = "noindex, nofollow"
	}

	return doc, nil
}
