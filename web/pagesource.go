package web

import (
	"context"
	"iter"
)

// PageSource provides an iterator over the pages it contains.
// If withContent is false, only page metadata is loaded and Page.Content is left empty.
type PageSource interface {
	Pages(ctx context.Context, withContent bool) iter.Seq2[*Page, error]
}

// PageSourceFunc is an adapter that lets an ordinary function be used as a PageSource.
type PageSourceFunc func(ctx context.Context, withContent bool) iter.Seq2[*Page, error]

// Pages calls f and implements the PageSource interface.
func (f PageSourceFunc) Pages(ctx context.Context, withContent bool) iter.Seq2[*Page, error] {
	return f(ctx, withContent)
}
