package web_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/web"
)

func TestPageIsPublished(t *testing.T) {
	require.False(t, (&web.Page{}).IsPublished(), "draft")
	require.False(t, (&web.Page{Published: scheduled}).IsPublished(), "scheduled")
	require.True(t, (&web.Page{Published: published}).IsPublished())
}

func TestPageIndexable(t *testing.T) {
	require.True(t, (&web.Page{Published: published}).Indexable())
	require.False(t, (&web.Page{Published: published, NoIndex: true}).Indexable())
	require.False(t, (&web.Page{}).Indexable())
}

func TestPageMarkdownSource(t *testing.T) {
	tests := map[string]struct {
		page   web.Page
		want   string
		wantOK bool
	}{
		"bytes":             {web.Page{ContentType: mx.ContentTypeMarkdown, Content: []byte("# Hi")}, "# Hi", true},
		"string":            {web.Page{ContentType: mx.ContentTypeMarkdown, Content: "# Hi"}, "# Hi", true},
		"without charset":   {web.Page{ContentType: "text/markdown", Content: "# Hi"}, "# Hi", true},
		"charset no space":  {web.Page{ContentType: "text/markdown;charset=utf-8", Content: "# Hi"}, "# Hi", true},
		"other parameter":   {web.Page{ContentType: "text/markdown; variant=GFM", Content: "# Hi"}, "# Hi", true},
		"malformed":         {web.Page{ContentType: "text/markdown; =", Content: "# Hi"}, "", false},
		"other case":        {web.Page{ContentType: "Text/Markdown", Content: "# Hi"}, "# Hi", true},
		"HTML":              {web.Page{ContentType: mx.ContentTypeHTML, Content: "<h1>Hi</h1>"}, "", false},
		"content not given": {web.Page{ContentType: mx.ContentTypeMarkdown}, "", false},
		// A component renders HTML; it is no Markdown source to publish.
		"component": {web.Page{ContentType: mx.ContentTypeMarkdown, Content: mx.Raw("# Hi")}, "", false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			source, ok := tt.page.MarkdownSource()
			require.Equal(t, tt.wantOK, ok)
			require.Equal(t, tt.want, string(source))
		})
	}
}

func TestPageValidate(t *testing.T) {
	valid := []*web.Page{
		{},
		{Path: "/", Lang: "en"},
		{Path: "/en/faq", Lang: "en-GB", Alternates: []web.Alternate{
			{HrefLang: "de", Path: "/de/faq"},
			{HrefLang: "fr", Path: "/fr/faq"},
			// The default version may be the page itself.
			{HrefLang: web.HrefLangDefault, Path: "/en/faq"},
			// The same URL in another format, served by content negotiation.
			{Type: "text/markdown", Path: "/en/faq"},
			// Same type, different languages: not duplicates.
			{Type: "text/markdown", HrefLang: "de", Path: "/de/faq.md"},
		}},
	}
	for _, page := range valid {
		require.NoError(t, page.Validate(), "page %q", page.Path)
	}

	invalid := map[string]struct {
		page    *web.Page
		wantErr string
	}{
		"invalid path":      {&web.Page{Path: "faq"}, "does not start with '/'"},
		"invalid lang":      {&web.Page{Path: "/", Lang: "en_US"}, "language tag"},
		"x-default as lang": {&web.Page{Path: "/", Lang: "x-default"}, "not a language"},
		"invalid alternate": {&web.Page{Path: "/", Alternates: []web.Alternate{{Path: "/other"}}}, "alternate 0"},
		"duplicate hreflang": {&web.Page{Path: "/en", Alternates: []web.Alternate{
			{HrefLang: "de", Path: "/de"},
			{HrefLang: "de", Path: "/de-2"},
		}}, "alternate 1 repeats"},
		"duplicate type": {&web.Page{Path: "/en", Alternates: []web.Alternate{
			{Type: "text/markdown", Path: "/en.md"},
			{Type: "TEXT/markdown", Path: "/en.markdown"},
		}}, "repeats"},
		"translation of itself": {&web.Page{Path: "/en", Alternates: []web.Alternate{
			{HrefLang: "de", Path: "/en"},
		}}, "points at the page itself"},
	}
	for name, tt := range invalid {
		t.Run(name, func(t *testing.T) {
			require.ErrorContains(t, tt.page.Validate(), tt.wantErr)
		})
	}
}

func TestPagePublishedBoundary(t *testing.T) {
	// Published exactly now counts as published, not as scheduled.
	page := &web.Page{Published: time.Now().Add(-time.Millisecond)}
	require.True(t, page.IsPublished())
}
