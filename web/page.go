package web

import (
	"context"
	"io/fs"
	"iter"
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// https://github.com/adrg/frontmatter
// https://github.com/yuin/goldmark

type Page struct {
	Route      mx.Route
	PathValues map[string]any

	Title       string
	Author      string
	Type        string
	Tags        []string
	Created     time.Time
	LastUpdated time.Time
	Published   time.Time // Zero time means not published

	Resources   []fs.File // URLs or file paths
	ContentType string
	Content     []byte // Can be nil if only metadata is needed
}

type PageTemplate interface {
	TransformPage(ctx context.Context, page *Page) (html.Document, error)
}

type PageTemplateFunc func(ctx context.Context, page *Page) (html.Document, error)

func (f PageTemplateFunc) TransformPage(ctx context.Context, page *Page) (html.Document, error) {
	return f(ctx, page)
}

type ContentTemplate interface {
	TransformContent(ctx context.Context, contentType string, content []byte) (mx.Component, error)
}

type ContentTemplateFunc func(ctx context.Context, contentType string, content []byte) (mx.Component, error)

func (f ContentTemplateFunc) TransformContent(ctx context.Context, contentType string, content []byte) (mx.Component, error) {
	return f(ctx, contentType, content)
}

type PageIterator interface {
	IterPages(ctx context.Context, withConent bool) iter.Seq2[*Page, error]
}

type PageIteratorFunc func(ctx context.Context, withConent bool) iter.Seq2[*Page, error]

func (f PageIteratorFunc) IterPages(ctx context.Context, withConent bool) iter.Seq2[*Page, error] {
	return f(ctx, withConent)
}
