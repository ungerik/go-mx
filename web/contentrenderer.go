package web

import (
	"context"

	"github.com/ungerik/go-mx"
)

// ContentRenderer renders a page's content of the given contentType into a Component.
type ContentRenderer interface {
	RenderContent(ctx context.Context, contentType string, content any) (mx.Component, error)
}

// ContentRendererFunc is an adapter that lets an ordinary function be used as a ContentRenderer.
type ContentRendererFunc func(ctx context.Context, contentType string, content any) (mx.Component, error)

// RenderContent calls f and implements the ContentRenderer interface.
func (f ContentRendererFunc) RenderContent(ctx context.Context, contentType string, content any) (mx.Component, error) {
	return f(ctx, contentType, content)
}
