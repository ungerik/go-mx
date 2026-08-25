package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ungerik/go-fs"
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/web"
)

func TestGlobPageSourceMarkdownFrontmatter(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "hello.md"), []byte(
		"---\n"+
			"title: Hello World\n"+
			"description: The first post\n"+
			"publishDate: 2026-06-01T08:00:00Z\n"+
			"tags: [go, web]\n"+
			"---\n"+
			"# Hello\n",
	), 0o600)
	require.NoError(t, err)

	source := &web.GlobPageSource{
		Pattern:       filepath.Join(dir, "*.md"),
		PageType:      "post",
		DefaultAuthor: "Site Author",
	}
	var pages []*web.Page
	for page, err := range source.Pages(context.Background(), true) {
		require.NoError(t, err)
		pages = append(pages, page)
	}
	require.Len(t, pages, 1)

	page := pages[0]
	// The description drives the meta description and og:description of the
	// rendered page, so it has to survive the front matter round trip.
	require.Equal(t, "The first post", page.Description)
	require.Equal(t, "Hello World", page.Title)
	require.Equal(t, []string{"go", "web"}, page.Tags)
	require.Equal(t, "post", page.Type)
	require.Equal(t, "Site Author", page.Author)
	require.Equal(t, mx.ContentTypeMarkdown, page.ContentType)
	require.True(t, page.IsPublished())
	require.NotNil(t, page.Content)
}

func TestGlobPageSourceDraft(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "draft.md"), []byte(
		"---\ntitle: Draft\ndraft: true\npublishDate: 2026-06-01T08:00:00Z\n---\nbody\n",
	), 0o600)
	require.NoError(t, err)

	source := &web.GlobPageSource{Pattern: filepath.Join(dir, "*.md")}
	for page, err := range source.Pages(context.Background(), false) {
		require.NoError(t, err)
		// A draft keeps its publish date out of the page, which is what keeps
		// it out of the sitemap and gets it a noindex meta tag.
		require.False(t, page.IsPublished())
		require.False(t, page.Indexable())
	}
}

func TestGlobPageSourcePagePath(t *testing.T) {
	// Regression: the source yielded pages without a Path, so every published
	// page failed Site.Sitemap and /sitemap.xml answered 500 — the built-in
	// file source could not feed the built-in sitemap.
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "blog"), 0o700))
	files := map[string]string{
		"index.md":      "---\ntitle: Home\npublishDate: 2026-06-01T08:00:00Z\n---\nhome\n",
		"blog/hello.md": "---\ntitle: Hello\npublishDate: 2026-06-01T08:00:00Z\n---\npost\n",
		"blog/index.md": "---\ntitle: Blog\npublishDate: 2026-06-01T08:00:00Z\n---\nlist\n",
		"blog/draft.md": "---\ntitle: Draft\n---\ndraft\n",
	}
	for name, content := range files {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600))
	}

	source := &web.GlobPageSource{Dir: fs.File(dir), Pattern: filepath.Join(dir, "*", "*.md")}
	paths := map[string]string{}
	for page, err := range source.Pages(context.Background(), false) {
		require.NoError(t, err)
		paths[page.Title] = page.Path
	}
	// An index file is the page of its directory, addressed by the directory
	// URL rather than by a URL ending in "/index".
	require.Equal(t, "/blog/hello", paths["Hello"])
	require.Equal(t, "/blog/", paths["Blog"])
}

func TestGlobPageSourceToSitemap(t *testing.T) {
	// The whole point of the source: markdown files in, sitemap out.
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "hello.md"),
		[]byte("---\ntitle: Hello\npublishDate: 2026-06-01T08:00:00Z\n---\npost\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "draft.md"),
		[]byte("---\ntitle: Draft\n---\ndraft\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "page.html"), []byte("<p>hi</p>"), 0o600))

	site := &web.Site{
		BaseURL: "https://example.com",
		Sources: []web.PageSource{&web.GlobPageSource{Dir: fs.File(dir), Pattern: filepath.Join(dir, "*.*")}},
	}
	sitemap, err := site.Sitemap(context.Background())
	require.NoError(t, err)

	locs := make([]string, len(sitemap.URLs))
	for i, u := range sitemap.URLs {
		locs[i] = u.Loc
	}
	// The draft is left out, the published markdown page and the HTML page
	// (dated by its modification time) are in.
	require.Equal(t, []string{"https://example.com/hello", "https://example.com/page"}, locs)

	response := httptest.NewRecorder()
	site.HandleSitemapXML(response, httptest.NewRequest(http.MethodGet, web.DefaultSitemapPath, nil))
	require.Equal(t, http.StatusOK, response.Code)
	require.Contains(t, response.Body.String(), "<loc>https://example.com/hello</loc>")
}

func TestGlobPageSourceEscapesFileNames(t *testing.T) {
	// Regression: file names went into Page.Path verbatim, so a "100%-off.md"
	// produced a path with an invalid percent escape that Site.URL rejects —
	// one legally named file broke the whole sitemap.
	dir := t.TempDir()
	for _, name := range []string{"100%-off.md", "faq?.md", "part#1.md"} {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name),
			[]byte("---\ntitle: T\npublishDate: 2026-06-01T08:00:00Z\n---\nx\n"), 0o600))
	}
	site := &web.Site{
		BaseURL: "https://example.com",
		Sources: []web.PageSource{&web.GlobPageSource{Dir: fs.File(dir), Pattern: filepath.Join(dir, "*.md")}},
	}
	sitemap, err := site.Sitemap(context.Background())
	require.NoError(t, err)
	locs := make([]string, len(sitemap.URLs))
	for i, u := range sitemap.URLs {
		locs[i] = u.Loc
	}
	require.Equal(t, []string{
		"https://example.com/100%25-off",
		"https://example.com/faq%3F",
		"https://example.com/part%231",
	}, locs)
}
