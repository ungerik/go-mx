package web

import (
	"context"

	"github.com/ungerik/go-mx/html"
)

type PageRenderer interface {
	RenderPage(ctx context.Context, page *Page) (html.Document, error)
}

type PageRendererFunc func(ctx context.Context, page *Page) (html.Document, error)

func (f PageRendererFunc) RenderPage(ctx context.Context, page *Page) (html.Document, error) {
	return f(ctx, page)
}

var DefaultPageRenderer PageRenderer = PageRendererFunc(DefaultRenderPage)

func DefaultRenderPage(ctx context.Context, page *Page) (doc html.Document, err error) {
	// https://learn.microsoft.com/en-us/previous-versions/windows/internet-explorer/ie-developer/platform-apis/dn255024(v=vs.85)
	// https://ogp.me/

	doc.Title = page.Title
	if page.NoIndex {
		doc.Meta["robots"] = "noindex, nofollow"
	}

	return doc, nil
}
