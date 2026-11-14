package web_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/web"
)

// titledRenderer returns a PageRenderer that renders only a title, so a test
// can tell which renderer PageTypeRenderer picked.
func titledRenderer(title string) web.PageRenderer {
	return web.PageRendererFunc(func(context.Context, *web.Page) (html.Document, error) {
		return html.Document{Title: title}, nil
	})
}

// textContentRenderer renders a string Content as a text node.
var textContentRenderer = web.ContentRendererFunc(func(_ context.Context, _ string, content any) (mx.Component, error) {
	return mx.Text(content.(string)), nil
})

func TestPageTypeRendererSelectsPageRenderer(t *testing.T) {
	ctx := context.Background()
	published := time.Now().Add(-time.Hour)

	t.Run("DefaultPageRenderer without any renderer", func(t *testing.T) {
		doc, err := (&web.PageTypeRenderer{}).RenderPage(ctx, &web.Page{Title: "Home", Published: published})
		require.NoError(t, err)
		require.Equal(t, "Home", doc.Title)
		require.Nil(t, doc.BodyChildren)
	})

	t.Run("DefaultRenderer for an unregistered type", func(t *testing.T) {
		r := &web.PageTypeRenderer{
			DefaultRenderer:   titledRenderer("default"),
			PageTypeRenderers: map[string]web.PageRenderer{"post": titledRenderer("post")},
		}
		doc, err := r.RenderPage(ctx, &web.Page{Type: "page"})
		require.NoError(t, err)
		require.Equal(t, "default", doc.Title)
	})

	t.Run("PageTypeRenderers by Page.Type", func(t *testing.T) {
		r := &web.PageTypeRenderer{
			DefaultRenderer:   titledRenderer("default"),
			PageTypeRenderers: map[string]web.PageRenderer{"post": titledRenderer("post")},
		}
		doc, err := r.RenderPage(ctx, &web.Page{Type: "post"})
		require.NoError(t, err)
		require.Equal(t, "post", doc.Title)
	})

	t.Run("page renderer error", func(t *testing.T) {
		r := &web.PageTypeRenderer{
			DefaultRenderer: web.PageRendererFunc(func(context.Context, *web.Page) (html.Document, error) {
				return html.Document{Title: "partial"}, errors.New("page failed")
			}),
		}
		doc, err := r.RenderPage(ctx, &web.Page{})
		require.EqualError(t, err, "page failed")
		require.Equal(t, html.Document{}, doc, "a failed render returns the zero Document")
	})
}

func TestPageTypeRendererContent(t *testing.T) {
	ctx := context.Background()
	render := func(t *testing.T, doc html.Document) string {
		t.Helper()
		var b strings.Builder
		require.NoError(t, doc.Render(ctx, mx.NewCheckedWriter(&b)))
		return b.String()
	}

	t.Run("Content rendered into BodyChildren", func(t *testing.T) {
		r := &web.PageTypeRenderer{
			ContentTypeRenderers: map[string]web.ContentRenderer{"text/plain": textContentRenderer},
		}
		doc, err := r.RenderPage(ctx, &web.Page{ContentType: "text/plain", Content: "a < b"})
		require.NoError(t, err)
		require.Contains(t, render(t, doc), "<body>a &lt; b</body>")
	})

	// A page renderer that already filled the body wins: the content
	// renderer is not consulted, not even to report that none is registered.
	t.Run("BodyChildren set by the page renderer kept", func(t *testing.T) {
		r := &web.PageTypeRenderer{
			DefaultRenderer: web.PageRendererFunc(func(context.Context, *web.Page) (html.Document, error) {
				return html.Document{BodyChildren: mx.Text("own body")}, nil
			}),
		}
		doc, err := r.RenderPage(ctx, &web.Page{ContentType: "text/unknown", Content: "ignored"})
		require.NoError(t, err)
		require.Contains(t, render(t, doc), "<body>own body</body>")
	})

	t.Run("nil Content needs no content renderer", func(t *testing.T) {
		doc, err := (&web.PageTypeRenderer{}).RenderPage(ctx, &web.Page{ContentType: "text/unknown"})
		require.NoError(t, err)
		require.Nil(t, doc.BodyChildren)
	})

	t.Run("no content renderer for the ContentType", func(t *testing.T) {
		r := &web.PageTypeRenderer{
			ContentTypeRenderers: map[string]web.ContentRenderer{"text/plain": textContentRenderer},
		}
		doc, err := r.RenderPage(ctx, &web.Page{ContentType: "text/unknown", Content: "x"})
		require.ErrorContains(t, err, `"text/unknown"`)
		require.Equal(t, html.Document{}, doc)
	})

	t.Run("content renderer error", func(t *testing.T) {
		r := &web.PageTypeRenderer{
			ContentTypeRenderers: map[string]web.ContentRenderer{
				"text/plain": web.ContentRendererFunc(func(context.Context, string, any) (mx.Component, error) {
					return nil, errors.New("content failed")
				}),
			},
		}
		doc, err := r.RenderPage(ctx, &web.Page{ContentType: "text/plain", Content: "x"})
		require.EqualError(t, err, "content failed")
		require.Equal(t, html.Document{}, doc)
	})
}
