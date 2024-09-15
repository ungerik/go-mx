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
