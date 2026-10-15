package web_test

import (
	"testing"

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

func TestPageLastModified(t *testing.T) {
	require.True(t, (&web.Page{}).LastModified().IsZero(), "draft")
	// Never updated: the publication is the last change.
	require.Equal(t, published, (&web.Page{Published: published}).LastModified())
	require.Equal(t, updated, (&web.Page{Published: published, LastUpdated: updated}).LastModified())
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
		// The Markdown version at the page's MarkdownPath.
		{Path: "/en/", Alternates: []web.Alternate{{Type: "text/markdown; charset=utf-8", Path: "/en/index.md"}}},
		// A document describing the page is no Markdown version of it,
		// whatever its type.
		{Path: "/en/faq", Alternates: []web.Alternate{{Rel: web.RelDescribedBy, Type: "text/markdown", Path: "/en/faq-notes.md"}}},
		// Without a Path there is no Markdown path to compare with.
		{Alternates: []web.Alternate{{Type: "text/markdown", Path: "/x.md"}}},
		// The Markdown path in another spelling of its escapes.
		{Path: "/café", Alternates: []web.Alternate{{Type: "text/markdown", Path: "/caf%c3%a9.md"}}},
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
		// llms.txt and the MCP tools link the page's MarkdownPath, so the
		// head may not name another Markdown version.
		"markdown elsewhere": {&web.Page{Path: "/en/faq", Alternates: []web.Alternate{
			{Type: "text/markdown; charset=utf-8", Path: "/en/faq.markdown"},
		}}, `not at the page's Markdown path "/en/faq.md"`},
		"markdown elsewhere, other case": {&web.Page{Path: "/en/faq", Alternates: []web.Alternate{
			{Type: "Text/Markdown", Path: "/en/faq.markdown"},
		}}, "not at the page's Markdown path"},
		"markdown without markdown path": {&web.Page{Path: "/en/index", Alternates: []web.Alternate{
			{Type: "text/markdown", Path: "/en/index.md"},
		}}, "ends in /index"},
		"translation of itself": {&web.Page{Path: "/en", Alternates: []web.Alternate{
			{HrefLang: "de", Path: "/en"},
		}}, "points at the page itself"},
		// Another spelling of the same path is still the same URL.
		"translation of itself escaped": {&web.Page{Path: "/café", Alternates: []web.Alternate{
			{HrefLang: "de", Path: "/caf%C3%A9"},
		}}, "points at the page itself"},
		// Percent escapes are case-insensitive.
		"translation of itself lower-case escapes": {&web.Page{Path: "/café", Alternates: []web.Alternate{
			{HrefLang: "de", Path: "/caf%c3%a9"},
		}}, "points at the page itself"},
	}
	for name, tt := range invalid {
		t.Run(name, func(t *testing.T) {
			require.ErrorContains(t, tt.page.Validate(), tt.wantErr)
		})
	}
}
