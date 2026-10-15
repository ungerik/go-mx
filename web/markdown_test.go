package web_test

import (
	"bytes"
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/web"
)

var updateGolden = flag.Bool("update", false, "update the golden files in testdata/")

// requireGolden compares got with the golden file testdata/name, rewriting it
// instead when the tests run with -update.
func requireGolden(t *testing.T, name string, got []byte) {
	t.Helper()
	path := filepath.Join("testdata", name)
	if *updateGolden {
		require.NoError(t, os.WriteFile(path, got, 0o644))
	}
	want, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, string(want), string(got))
}

func TestMarkdownPath(t *testing.T) {
	tests := []struct {
		path    string
		want    string
		wantErr string
	}{
		{path: "/en/faq", want: "/en/faq.md"},
		{path: "/en/", want: "/en/index.md"},
		{path: "/", want: "/index.md"},
		{path: "/blog/release-1.2", want: "/blog/release-1.2.md"},
		{path: "/a%20b", want: "/a%20b.md"},
		// The page and its alternate would be the same URL.
		{path: "/en/faq.md", wantErr: "already ends in .md"},
		// Its alternate would be the one of "/en/".
		{path: "/en/index", wantErr: "ends in /index"},
		{path: "/index", wantErr: "ends in /index"},
		{path: "/en/indexes", want: "/en/indexes.md"},
		// The checks look at the decoded path a server routes by.
		{path: "/en/faq%2Emd", wantErr: "already ends in .md"},
		{path: "/en/%69ndex", wantErr: "ends in /index"},
		{path: "/README.MD", wantErr: "already ends in .md"},
		{path: "", wantErr: "empty"},
		{path: "en/faq", wantErr: "does not start with '/'"},
		{path: "/en/../faq", wantErr: ".."},
		{path: "/en/faq?x=1", wantErr: "query"},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got, err := web.MarkdownPath(tt.path)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				require.Empty(t, got)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

// markdownSite and markdownPage are a translated page with a Markdown source,
// the input of the Markdown alternate golden files.
func markdownSite() *web.Site {
	return &web.Site{BaseURL: "https://example.com", Lang: "en", DefaultLang: "en", Description: "Not the page's"}
}

func markdownPage() *web.Page {
	return &web.Page{
		Path:        "/en/faq",
		Title:       "Frequently\nasked questions",
		Description: "Answers to what people ask most.\n\nUpdated whenever a new question comes up.",
		Published:   published,
		LastUpdated: updated,
		Alternates: []web.Alternate{
			{HrefLang: "de", Path: "/de/faq"},
			{HrefLang: "fr", Path: "/fr/faq(1)"},
			{Type: "text/markdown", Path: "/en/faq.md"},
		},
		ContentType: mx.ContentTypeMarkdown,
		Content:     "## What is this?\n\nAn example.\n",
	}
}

func TestSiteMarkdownAlternate(t *testing.T) {
	doc, err := markdownSite().MarkdownAlternate(markdownPage())
	require.NoError(t, err)

	require.Equal(t, "https://example.com/en/faq", doc.URL)
	require.Equal(t, "en", doc.Lang)
	require.Equal(t, updated, doc.LastUpdated)
	// The page itself and the x-default link are not translations of it.
	require.Equal(t,
		[]web.Translation{
			{Lang: "de", URL: "https://example.com/de/faq"},
			{Lang: "fr", URL: "https://example.com/fr/faq(1)"},
		},
		doc.Translations,
	)
	requireGolden(t, "markdown_alternate.md", doc.Bytes())

	// Sections are the caller's: added after building, rendered before the
	// page's own source.
	doc.Sections = append(doc.Sections,
		web.MarkdownSection{Title: "Facts", Body: "- Founded: 2026\n- Employees: 3\n"},
		web.MarkdownSection{Body: "A section without a heading."},
	)
	requireGolden(t, "markdown_alternate_sections.md", doc.Bytes())
}

func TestSiteMarkdownAlternateTranslationOfDefault(t *testing.T) {
	// A translation of the version in Site.DefaultLang gets an x-default link
	// pointing at that version, but x-default is no language: only the
	// translation itself is listed.
	page := markdownPage()
	page.Path = "/de/faq"
	page.Lang = "de"
	page.Alternates = []web.Alternate{{HrefLang: "en", Path: "/en/faq"}}

	headers, err := markdownSite().PageHeaders(page)
	require.NoError(t, err)
	require.Contains(t, headers.Values("Link"), `<https://example.com/en/faq>; rel="alternate"; hreflang="x-default"`)

	doc, err := markdownSite().MarkdownAlternate(page)
	require.NoError(t, err)
	require.Equal(t, "de", doc.Lang)
	require.Equal(t, []web.Translation{{Lang: "en", URL: "https://example.com/en/faq"}}, doc.Translations)
}

func TestSiteMarkdownAlternateMinimal(t *testing.T) {
	// Without a description, a language or translations, the header is just
	// the title and the facts that always exist. LastUpdated falls back to
	// the publication date.
	site := &web.Site{BaseURL: "https://example.com"}
	page := &web.Page{
		Path:        "/notes/",
		Title:       "Notes",
		Published:   published,
		ContentType: "text/markdown",
		Content:     []byte("Some notes, no trailing newline"),
	}
	doc, err := site.MarkdownAlternate(page)
	require.NoError(t, err)
	require.Equal(t, published, doc.LastUpdated)
	requireGolden(t, "markdown_alternate_minimal.md", doc.Bytes())
}

func TestSiteMarkdownAlternateErrors(t *testing.T) {
	site := markdownSite()

	page := markdownPage()
	page.ContentType = mx.ContentTypeHTML
	_, err := site.MarkdownAlternate(page)
	require.ErrorContains(t, err, "no Markdown alternate")

	page = markdownPage()
	page.Path = ""
	_, err = site.MarkdownAlternate(page)
	require.ErrorContains(t, err, "path is empty")

	page = markdownPage()
	page.Alternates = append(page.Alternates, web.Alternate{HrefLang: "de", Path: "/de/faq-2"})
	_, err = site.MarkdownAlternate(page)
	require.ErrorContains(t, err, "repeats")
}

func TestSiteMarkdownAlternateURL(t *testing.T) {
	site := markdownSite()

	got, err := site.MarkdownAlternateURL(markdownPage())
	require.NoError(t, err)
	require.Equal(t, "https://example.com/en/faq.md", got)

	// Every page MarkdownAlternate fails for gets no URL either, so nothing
	// lists an alternate that answers with an error.
	page := markdownPage()
	page.Path = ""
	_, err = site.MarkdownAlternateURL(page)
	require.ErrorContains(t, err, "path is empty")

	page = markdownPage()
	page.Alternates = append(page.Alternates, web.Alternate{HrefLang: "de", Path: "/de/faq-2"})
	_, err = site.MarkdownAlternateURL(page)
	require.ErrorContains(t, err, "repeats")

	// Translations need the page's own language for the hreflang links.
	_, err = (&web.Site{BaseURL: "https://example.com"}).MarkdownAlternateURL(markdownPage())
	require.ErrorContains(t, err, "no language")

	// A path whose Markdown path would collide has no alternate URL.
	page = markdownPage()
	page.Path = "/en/index"
	_, err = site.MarkdownAlternateURL(page)
	require.ErrorContains(t, err, "ends in /index")
}

func TestMarkdownDocumentLineBreaks(t *testing.T) {
	// CommonMark ends a line at a lone carriage return, so every kind of line
	// break in a description has to continue the block quote, or what
	// follows it could be a heading. Control characters are dropped from
	// one-line values.
	doc := &web.MarkdownDocument{
		Title:       "Title\x00\x1b[31m",
		Description: "First\r# Injected\r\nSecond\nThird",
		Sections:    []web.MarkdownSection{{Title: "Facts\r## Injected", Body: "Body"}},
	}
	require.Equal(t,
		"# Title[31m\n\n"+
			"> First\n> # Injected\n> Second\n> Third\n\n"+
			"## Facts ## Injected\n\nBody\n\n"+
			"---\n\n",
		string(doc.Bytes()),
	)
}

func TestMarkdownDocumentOneLineMetadata(t *testing.T) {
	// The fields of a document are the caller's to set, and every metadata
	// item is one list item, so none of them may start a block of its own.
	doc := &web.MarkdownDocument{
		URL:          "https://example.com/a\n# Injected",
		Translations: []web.Translation{{Lang: "de\n# Injected](x) [y", URL: "https://example.com/de\n# Injected"}},
	}
	require.Equal(t,
		"- Canonical URL: https://example.com/a # Injected\n"+
			"- Translations: [de # Injected)(x) (y](https://example.com/de%20#%20Injected)\n\n"+
			"---\n\n",
		string(doc.Bytes()),
	)

	// A trailing backslash would escape the closing bracket or parenthesis,
	// and a translation without a usable label is left out.
	doc = &web.MarkdownDocument{Translations: []web.Translation{
		{Lang: `de\`, URL: `https://example.com/de\`},
		{Lang: `\`, URL: "https://example.com/x"},
	}}
	require.Equal(t, "- Translations: [de](https://example.com/de%5C)\n\n---\n\n", string(doc.Bytes()))
}

func TestMarkdownDocumentWriteTo(t *testing.T) {
	doc := &web.MarkdownDocument{Title: "Hello", Source: []byte("Body\n")}
	var b bytes.Buffer
	n, err := doc.WriteTo(&b)
	require.NoError(t, err)
	require.Equal(t, int64(b.Len()), n)
	require.Equal(t, "# Hello\n\n---\n\nBody\n", b.String())
	require.Equal(t, b.Bytes(), doc.Bytes())
}

func TestSiteServeMarkdown(t *testing.T) {
	site := markdownSite()
	page := markdownPage()

	response := httptest.NewRecorder()
	site.ServeMarkdown(response, httptest.NewRequest(http.MethodGet, "/en/faq.md", nil), page, "rev-42", updated)

	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, mx.ContentTypeMarkdown, response.Header().Get("Content-Type"))
	// The alternate points search engines at the HTML page.
	require.Equal(t, []string{`<https://example.com/en/faq>; rel="canonical"`}, response.Header().Values("Link"))
	require.Equal(t, "Accept", response.Header().Get("Vary"))
	// http.ServeContent cannot match an unquoted entity tag, and the tag of
	// the alternate must not be the one of the HTML page sharing its URL.
	require.Equal(t, `"rev-42-md"`, response.Header().Get("ETag"))
	require.Equal(t, updated.Format(http.TimeFormat), response.Header().Get("Last-Modified"))
	require.Empty(t, response.Header().Get("X-Robots-Tag"))
	doc, err := site.MarkdownAlternate(page)
	require.NoError(t, err)
	require.Equal(t, string(doc.Bytes()), response.Body.String())

	t.Run("if-none-match", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/en/faq.md", nil)
		request.Header.Set("If-None-Match", `"rev-42-md"`)
		response := httptest.NewRecorder()
		site.ServeMarkdown(response, request, page, `"rev-42"`, updated)
		require.Equal(t, http.StatusNotModified, response.Code)
		require.Empty(t, response.Body.String())
		require.Equal(t, "Accept", response.Header().Get("Vary"))
	})

	t.Run("vary set by the caller", func(t *testing.T) {
		// A handler negotiating HTML and Markdown on the page's URL sets
		// "Vary: Accept" before it knows which one it answers with.
		response := httptest.NewRecorder()
		response.Header().Add("Vary", "Origin, accept")
		site.ServeMarkdown(response, httptest.NewRequest(http.MethodGet, "/en/faq", nil), page, "", updated)
		require.Equal(t, []string{"Origin, accept"}, response.Header().Values("Vary"))
	})

	t.Run("if-none-match outdated", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/en/faq.md", nil)
		request.Header.Set("If-None-Match", `"rev-41"`)
		response := httptest.NewRecorder()
		site.ServeMarkdown(response, request, page, "rev-42", updated)
		require.Equal(t, http.StatusOK, response.Code)
	})

	t.Run("weak etag", func(t *testing.T) {
		response := httptest.NewRecorder()
		site.ServeMarkdown(response, httptest.NewRequest(http.MethodGet, "/en/faq.md", nil), page, `W/"rev-42"`, updated)
		require.Equal(t, `W/"rev-42-md"`, response.Header().Get("ETag"))
	})

	t.Run("malformed etag", func(t *testing.T) {
		response := httptest.NewRecorder()
		site.ServeMarkdown(response, httptest.NewRequest(http.MethodGet, "/en/faq.md", nil), page, `"rev-42`, updated)
		require.Equal(t, `"rev-42-md"`, response.Header().Get("ETag"))
	})

	t.Run("without etag", func(t *testing.T) {
		response := httptest.NewRecorder()
		site.ServeMarkdown(response, httptest.NewRequest(http.MethodGet, "/en/faq.md", nil), page, "", updated)
		require.Equal(t, http.StatusOK, response.Code)
		require.Empty(t, response.Header().Get("ETag"))
	})

	t.Run("not indexable", func(t *testing.T) {
		// Serving a draft is the caller's decision; if it does, the
		// alternate carries the noindex its HTML page would.
		draft := markdownPage()
		draft.Published = scheduled
		response := httptest.NewRecorder()
		site.ServeMarkdown(response, httptest.NewRequest(http.MethodGet, "/en/faq.md", nil), draft, "", updated)
		require.Equal(t, http.StatusOK, response.Code)
		require.Equal(t, "noindex, nofollow", response.Header().Get("X-Robots-Tag"))
	})

	t.Run("no Markdown alternate", func(t *testing.T) {
		htmlPage := markdownPage()
		htmlPage.ContentType = mx.ContentTypeHTML
		response := httptest.NewRecorder()
		site.ServeMarkdown(response, httptest.NewRequest(http.MethodGet, "/en/faq.md", nil), htmlPage, "", updated)
		// Regression: an HTML-only page answered every agent preferring
		// Markdown with a 500 when wired as the docs show.
		require.Equal(t, http.StatusNotFound, response.Code)
	})

	t.Run("broken alternates", func(t *testing.T) {
		broken := markdownPage()
		broken.Alternates = append(broken.Alternates, web.Alternate{HrefLang: "de", Path: "/de/faq-2"})
		response := httptest.NewRecorder()
		site.ServeMarkdown(response, httptest.NewRequest(http.MethodGet, "/en/faq.md", nil), broken, "", updated)
		require.Equal(t, http.StatusInternalServerError, response.Code)
	})
}

func TestAcceptsMarkdown(t *testing.T) {
	tests := []struct {
		name   string
		accept []string
		want   bool
	}{
		{name: "missing", accept: nil, want: false},
		{name: "markdown", accept: []string{"text/markdown"}, want: true},
		{name: "markdown with charset", accept: []string{"text/markdown; charset=utf-8"}, want: true},
		{name: "upper case", accept: []string{"Text/Markdown"}, want: true},
		{name: "html preferred", accept: []string{"text/html, text/markdown;q=0.9"}, want: false},
		{name: "markdown preferred", accept: []string{"text/markdown, text/html;q=0.9"}, want: true},
		// At equal quality the order decides, which is how agents rank.
		{name: "tie, markdown first", accept: []string{"text/markdown, text/html"}, want: true},
		{name: "tie, html first", accept: []string{"text/html, text/markdown"}, want: false},
		{name: "agent", accept: []string{"text/markdown, text/html, */*"}, want: true},
		{name: "tie with wildcard", accept: []string{"text/markdown, */*"}, want: true},
		{name: "tie across header lines", accept: []string{"text/html", "text/markdown"}, want: false},
		{name: "wildcard only", accept: []string{"*/*"}, want: false},
		{name: "text wildcard only", accept: []string{"text/*"}, want: false},
		{
			name:   "browser",
			accept: []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8"},
			want:   false,
		},
		{name: "markdown over wildcard", accept: []string{"text/markdown, */*;q=0.1"}, want: true},
		{name: "wildcard over markdown", accept: []string{"text/markdown;q=0.5, */*"}, want: false},
		// text/html is matched by the most specific range, so a low text/*
		// is what counts against Markdown here, not the high */*.
		{name: "text wildcard is more specific", accept: []string{"text/markdown;q=0.6, text/*;q=0.5, */*"}, want: true},
		{name: "html explicit over wildcard", accept: []string{"text/markdown;q=0.6, text/html;q=0.7, */*;q=0.1"}, want: false},
		{name: "refused", accept: []string{"text/markdown;q=0"}, want: false},
		{name: "malformed q", accept: []string{"text/markdown;q=high"}, want: false},
		{name: "q out of range", accept: []string{"text/markdown;q=2"}, want: false},
		// NaN compares false with everything, so it must not become the
		// quality of text/html and lose to nothing.
		{name: "NaN q", accept: []string{"text/html;q=nan, text/markdown;q=0.5"}, want: true},
		{name: "NaN q for markdown", accept: []string{"text/markdown;q=NaN"}, want: false},
		{name: "spaces and upper-case Q", accept: []string{"text/html ; Q=0.4 , text/markdown ; q=0.5"}, want: true},
		{name: "several header lines", accept: []string{"text/html;q=0.4", "text/markdown"}, want: true},
		{name: "empty ranges", accept: []string{", ,text/markdown"}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			for _, accept := range tt.accept {
				request.Header.Add("Accept", accept)
			}
			require.Equal(t, tt.want, web.AcceptsMarkdown(request))
		})
	}
}
