package web

import (
	"context"

	"github.com/domonda/go-errs"
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

type DefaultVisualizer struct {
	ContentTypeVisualizers map[string]ContentVisualizer
}

func (v *DefaultVisualizer) VisualizePage(ctx context.Context, page *Page) (mx.Component, error) {
	contentVisualizer := v.ContentTypeVisualizers[page.ContentType]
	if contentVisualizer == nil {
		return nil, errs.Errorf("no ContentVisualizer for ContentType %q", page.ContentType)
	}
	body, err := contentVisualizer.VisualizeContent(ctx, page.ContentType, page.Content)
	if err != nil {
		return nil, err
	}
	return &html.HTML{
		Title: page.Title,
		Body:  body,
	}, nil
}
