package web_test

import (
	"bytes"
	"context"
	"errors"
	"iter"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/web"
)

func exampleLLMsTxt() *web.LLMsTxt {
	return &web.LLMsTxt{
		Title:   "Example",
		Summary: "A site about examples.",
		Context: "Every page links its Markdown version.",
		Sections: []web.LLMsSection{
			{
				Title: "Docs",
				Links: []web.LLMsLink{
					{Title: "Getting started", URL: "https://example.com/docs/start.md", Note: "Install and first steps"},
					{Title: "Reference", URL: "/docs/reference.md"},
				},
			},
			{Title: "Empty"},
			{
				Title: "Optional",
				Links: []web.LLMsLink{{Title: "Changelog", URL: "http://example.com/changelog.md#latest"}},
			},
		},
	}
}

func TestLLMsTxtMarshalText(t *testing.T) {
	text, err := exampleLLMsTxt().MarshalText()
	require.NoError(t, err)
	requireGolden(t, "llms.txt", text)

	// Only the title is required.
	minimal := &web.LLMsTxt{Title: "Example"}
	require.Equal(t, "# Example\n", minimal.String())
}

func TestLLMsTxtValidate(t *testing.T) {
	tests := map[string]struct {
		modify  func(*web.LLMsTxt)
		wantErr string
	}{
		"empty title":           {func(l *web.LLMsTxt) { l.Title = " " }, "title is empty"},
		"line break in title":   {func(l *web.LLMsTxt) { l.Title = "Example\n## Injected" }, "line break"},
		"line break in summary": {func(l *web.LLMsTxt) { l.Summary = "a\r\nb" }, "line break"},
		"line break in context": {func(l *web.LLMsTxt) { l.Context = "a\nb" }, "line break"},
		"control character":     {func(l *web.LLMsTxt) { l.Summary = "a\x00b" }, "control character"},
		"context is a heading":  {func(l *web.LLMsTxt) { l.Context = "  ## Injected" }, "heading"},
		"context is a quote":    {func(l *web.LLMsTxt) { l.Context = " > Not the summary" }, "block quote"},
		// Each of these makes the context a code or HTML block instead of a
		// paragraph; an unclosed fence or comment takes the sections along.
		"context is a fence":          {func(l *web.LLMsTxt) { l.Context = "```" }, "code or HTML block"},
		"context is a tilde fence":    {func(l *web.LLMsTxt) { l.Context = " ~~~ go" }, "code or HTML block"},
		"context is HTML":             {func(l *web.LLMsTxt) { l.Context = "<!-- note" }, "code or HTML block"},
		"context is indented code":    {func(l *web.LLMsTxt) { l.Context = "    code" }, "code or HTML block"},
		"context is tab-indented":     {func(l *web.LLMsTxt) { l.Context = "\tcode" }, "code or HTML block"},
		"angle bracket in link title": {func(l *web.LLMsTxt) { l.Sections[0].Links[0].Title = `Docs <b title="` }, "'<'"},
		"empty section title":         {func(l *web.LLMsTxt) { l.Sections[0].Title = "" }, "section 0: section title is empty"},
		"line break in section":       {func(l *web.LLMsTxt) { l.Sections[1].Title = "a\nb" }, "section 1"},
		"empty link title":            {func(l *web.LLMsTxt) { l.Sections[0].Links[1].Title = "" }, "link 1: link title is empty"},
		"line break in link":          {func(l *web.LLMsTxt) { l.Sections[0].Links[0].Title = "a\nb" }, "line break"},
		"bracket in link title":       {func(l *web.LLMsTxt) { l.Sections[0].Links[0].Title = "a]b" }, "square bracket"},
		"opening bracket":             {func(l *web.LLMsTxt) { l.Sections[0].Links[0].Title = "a[b" }, "square bracket"},
		"trailing backslash":          {func(l *web.LLMsTxt) { l.Sections[0].Links[0].Title = `a\` }, "backslash"},
		"link title x":                {func(l *web.LLMsTxt) { l.Sections[0].Links[0].Title = "X" }, "task list checkbox"},
		"backtick in link title":      {func(l *web.LLMsTxt) { l.Sections[0].Links[0].Title = "use `go" }, "backtick"},
		"line break in note":          {func(l *web.LLMsTxt) { l.Sections[0].Links[0].Note = "a\n- [x](/x)" }, "line break"},
		"empty URL":                   {func(l *web.LLMsTxt) { l.Sections[0].Links[0].URL = "" }, "URL is empty"},
		"parenthesis in URL":          {func(l *web.LLMsTxt) { l.Sections[0].Links[0].URL = "/a)b" }, "parenthesis"},
		"space in URL":                {func(l *web.LLMsTxt) { l.Sections[0].Links[0].URL = "/a b" }, "whitespace"},
		"relative URL":                {func(l *web.LLMsTxt) { l.Sections[0].Links[0].URL = "docs/a.md" }, "neither"},
		"protocol-relative URL":       {func(l *web.LLMsTxt) { l.Sections[0].Links[0].URL = "//evil.example/a.md" }, "neither"},
		"other scheme":                {func(l *web.LLMsTxt) { l.Sections[0].Links[0].URL = "javascript:alert(1)" }, "parenthesis"},
		"tab-stop indentation":        {func(l *web.LLMsTxt) { l.Context = " \tcode" }, "code or HTML block"},
		"backslash in URL":            {func(l *web.LLMsTxt) { l.Sections[0].Links[0].URL = `https://example.com/a\` }, "backslash"},
		"other scheme no parens":      {func(l *web.LLMsTxt) { l.Sections[0].Links[0].URL = "ftp://example.com/a.md" }, "neither"},
		"URL without host":            {func(l *web.LLMsTxt) { l.Sections[0].Links[0].URL = "https:///a.md" }, "neither"},
		"invalid URL":                 {func(l *web.LLMsTxt) { l.Sections[0].Links[0].URL = "https://example.com/%zz" }, "invalid URL escape"},
		"invalid path":                {func(l *web.LLMsTxt) { l.Sections[0].Links[0].URL = "/%zz" }, "invalid URL escape"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			llms := exampleLLMsTxt()
			tt.modify(llms)
			require.ErrorContains(t, llms.Validate(), tt.wantErr)

			// An invalid file is never written.
			_, err := llms.MarshalText()
			require.Error(t, err)
			require.Contains(t, llms.String(), "!ERROR: ")
		})
	}
	require.NoError(t, exampleLLMsTxt().Validate())

	// Up to three columns of indentation keep the context a paragraph.
	llms := exampleLLMsTxt()
	llms.Context = "   Three spaces are a paragraph."
	require.NoError(t, llms.Validate())
}

func TestLLMsSectionValidate(t *testing.T) {
	require.NoError(t, web.LLMsSection{Title: "Docs"}.Validate())
	require.ErrorContains(t, web.LLMsSection{}.Validate(), "section title is empty")
}

func TestLLMsLinkValidate(t *testing.T) {
	require.NoError(t, web.LLMsLink{Title: "Docs", URL: "/docs.md"}.Validate())
	require.ErrorContains(t, web.LLMsLink{Title: "Docs"}.Validate(), "URL is empty")
}

func TestLLMsTxtWriteTo(t *testing.T) {
	var b bytes.Buffer
	n, err := exampleLLMsTxt().WriteTo(&b)
	require.NoError(t, err)
	require.Equal(t, int64(b.Len()), n)
	require.Equal(t, exampleLLMsTxt().String(), b.String())

	_, err = (&web.LLMsTxt{}).WriteTo(&b)
	require.Error(t, err)
}

func TestLLMsTxtHandleHTTP(t *testing.T) {
	llms := exampleLLMsTxt()
	response := httptest.NewRecorder()
	llms.HandleHTTP(response, httptest.NewRequest(http.MethodGet, web.LLMsTxtPath, nil))

	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, mx.ContentTypeMarkdown, response.Header().Get("Content-Type"))
	require.Equal(t, llms.String(), response.Body.String())
	etag := response.Header().Get("ETag")
	require.Regexp(t, `^"[0-9a-f]{32}"$`, etag)

	// The ETag is derived from the content, so an unchanged file is not
	// sent again...
	request := httptest.NewRequest(http.MethodGet, web.LLMsTxtPath, nil)
	request.Header.Set("If-None-Match", etag)
	response = httptest.NewRecorder()
	llms.HandleHTTP(response, request)
	require.Equal(t, http.StatusNotModified, response.Code)

	// ...and a changed one is.
	llms.Summary = "Changed."
	response = httptest.NewRecorder()
	llms.HandleHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	require.NotEqual(t, etag, response.Header().Get("ETag"))

	t.Run("invalid", func(t *testing.T) {
		response := httptest.NewRecorder()
		(&web.LLMsTxt{}).HandleHTTP(response, httptest.NewRequest(http.MethodGet, web.LLMsTxtPath, nil))
		require.Equal(t, http.StatusInternalServerError, response.Code)
	})
}

// contentPages is a PageSource that only fills Page.Content when asked to,
// like a real source.
type contentPages []web.Page

func (p contentPages) Pages(ctx context.Context, withContent bool) iter.Seq2[*web.Page, error] {
	return func(yield func(*web.Page, error) bool) {
		for _, page := range p {
			if !withContent {
				page.Content = nil
			}
			if !yield(&page, nil) {
				return
			}
		}
	}
}

func TestSiteLLMsTxt(t *testing.T) {
	markdown := func(title, path, pageType, description string) web.Page {
		return web.Page{
			Title:       title,
			Path:        path,
			Type:        pageType,
			Description: description,
			Published:   published,
			ContentType: mx.ContentTypeMarkdown,
			Content:     "# " + title,
		}
	}
	site := &web.Site{
		BaseURL:     "https://example.com",
		Title:       "Example",
		Description: "A site\nabout examples.",
		Sources: []web.PageSource{contentPages{
			markdown("Getting started", "/docs/start", "doc", "Install\nand first steps"),
			markdown("Hello [world]", "/blog/hello", "post", ""),
			markdown("", "/docs/", "doc", ""),
			markdown("Imprint", "/imprint", "", ""),
			markdown("Paren", "/docs/a(b)", "guide", ""),
			// Site content is made to fit instead of failing the file.
			markdown("Use `go\x00 [run] <b>\\", "/docs/run", "doc", "Runs \x1b[1mit\x07"),
			markdown(`\\`, "/docs/backslashes", "doc", ""),
			// A lone "x" would render as a checkbox, so the path stands in.
			markdown("x", "/docs/x", "doc", ""),
			// Not listed: a draft, a page marked NoIndex, and pages without
			// a Markdown source.
			func() web.Page { p := markdown("Draft", "/blog/draft", "post", ""); p.Published = scheduled; return p }(),
			func() web.Page { p := markdown("Hidden", "/hidden", "", ""); p.NoIndex = true; return p }(),
			{Title: "HTML", Path: "/html", Published: published, ContentType: mx.ContentTypeHTML, Content: "<p>"},
			{Title: "Metadata only", Path: "/meta", Published: published, ContentType: mx.ContentTypeMarkdown},
		}},
	}
	llms, err := site.LLMsTxt(context.Background(), web.LLMsTxtOptions{
		Context:       "Every link is the Markdown version of a page.",
		SectionTitles: map[string]string{"doc": "Docs", "guide": "Docs", "post": "Blog"},
	})
	require.NoError(t, err)
	text, err := llms.MarshalText()
	require.NoError(t, err)
	requireGolden(t, "llms_site.txt", text)

	t.Run("options override site", func(t *testing.T) {
		llms, err := site.LLMsTxt(context.Background(), web.LLMsTxtOptions{
			Title:         "Override",
			Summary:       "Overridden summary.",
			SectionTitles: map[string]string{"": "Other"},
		})
		require.NoError(t, err)
		require.Equal(t, "Override", llms.Title)
		require.Equal(t, "Overridden summary.", llms.Summary)
		require.Equal(t, []string{"doc", "post", "Other", "guide"}, sectionTitles(llms))

		llms, err = site.LLMsTxt(context.Background(), web.LLMsTxtOptions{})
		require.NoError(t, err)
		require.Equal(t, []string{"doc", "post", web.DefaultLLMsSectionTitle, "guide"}, sectionTitles(llms))
	})

	t.Run("no title", func(t *testing.T) {
		_, err := (&web.Site{BaseURL: "https://example.com"}).LLMsTxt(context.Background(), web.LLMsTxtOptions{})
		require.ErrorContains(t, err, "title is empty")
	})

	t.Run("backslashes around a trailing space", func(t *testing.T) {
		// Regression: the trailing backslashes were trimmed before the space
		// between them, which left one at the end of the link title and
		// failed the whole file.
		site := &web.Site{BaseURL: "https://example.com", Title: "Example", Sources: []web.PageSource{contentPages{
			markdown(`Foo \ \`, "/foo", "", ""),
		}}}
		llms, err := site.LLMsTxt(context.Background(), web.LLMsTxtOptions{})
		require.NoError(t, err)
		require.Equal(t, "Foo", llms.Sections[0].Links[0].Title)
	})

	t.Run("source error", func(t *testing.T) {
		failing := web.PageSourceFunc(func(ctx context.Context, withContent bool) iter.Seq2[*web.Page, error] {
			return func(yield func(*web.Page, error) bool) { yield(nil, errors.New("source failed")) }
		})
		_, err := (&web.Site{BaseURL: "https://example.com", Title: "X", Sources: []web.PageSource{failing}}).
			LLMsTxt(context.Background(), web.LLMsTxtOptions{})
		require.ErrorContains(t, err, "source failed")
	})

	t.Run("page without alternate", func(t *testing.T) {
		// Linking it would point agents at a 500.
		page := markdown("Translated", "/faq", "", "")
		page.Alternates = []web.Alternate{{HrefLang: "de", Path: "/de/faq"}}
		site := &web.Site{BaseURL: "https://example.com", Title: "X", Sources: []web.PageSource{contentPages{page}}}
		_, err := site.LLMsTxt(context.Background(), web.LLMsTxtOptions{})
		require.ErrorContains(t, err, "no language")
	})

	t.Run("page without path", func(t *testing.T) {
		// Like the sitemap: a public page that cannot be linked is a bug in
		// the source, not something to leave out silently.
		site := &web.Site{BaseURL: "https://example.com", Title: "X", Sources: []web.PageSource{
			contentPages{markdown("Lost", "", "", "")},
		}}
		_, err := site.LLMsTxt(context.Background(), web.LLMsTxtOptions{})
		require.ErrorContains(t, err, "Lost")
	})

	t.Run("invalid base URL", func(t *testing.T) {
		site := &web.Site{BaseURL: "example.com", Title: "X", Sources: []web.PageSource{
			contentPages{markdown("Page", "/page", "", "")},
		}}
		_, err := site.LLMsTxt(context.Background(), web.LLMsTxtOptions{})
		require.ErrorContains(t, err, "BaseURL")
	})
}

func sectionTitles(llms *web.LLMsTxt) []string {
	var titles []string
	for _, section := range llms.Sections {
		titles = append(titles, section.Title)
	}
	return titles
}
