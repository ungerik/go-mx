package web

import (
	"context"
	"iter"
)

type PageSource interface {
	Pages(ctx context.Context, withContent bool) iter.Seq2[*Page, error]
}

type PageSourceFunc func(ctx context.Context, withContent bool) iter.Seq2[*Page, error]

func (f PageSourceFunc) Pages(ctx context.Context, withContent bool) iter.Seq2[*Page, error] {
	return f(ctx, withContent)
}
