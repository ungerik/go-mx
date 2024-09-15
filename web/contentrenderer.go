package web

import (
	"context"

	"github.com/ungerik/go-mx"
)

type ContentRenderer interface {
	RenderContent(ctx context.Context, contentType string, content any) (mx.Component, error)
}

type ContentRendererFunc func(ctx context.Context, contentType string, content any) (mx.Component, error)

func (f ContentRendererFunc) RenderContent(ctx context.Context, contentType string, content any) (mx.Component, error) {
	return f(ctx, contentType, content)
}
