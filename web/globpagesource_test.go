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
			"lang: en-GB\n"+
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
	require.Equal(t, "en-GB", page.Lang)
	require.True(t, page.IsPublished())
	// The front matter is parsed into the page fields, not published again
	// as part of the content.
	markdown, ok := page.MarkdownSource()
	require.True(t, ok)
	require.Equal(t, "# Hello\n", string(markdown))
}

func TestGlobPageSourceMarkdownWithoutFrontmatter(t *testing.T) {
	// A Markdown file without front matter is all content: nothing of it may
	// be taken for metadata and dropped.
	dir := t.TempDir()
	const content = "# Plain\n\n---\n\nNo front matter, but a thematic break.\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "plain.md"), []byte(content), 0o600))

	source := &web.GlobPageSource{Pattern: filepath.Join(dir, "*.md")}
	var pages []*web.Page
	for page, err := range source.Pages(context.Background(), true) {
		require.NoError(t, err)
		pages = append(pages, page)
	}
	require.Len(t, pages, 1)
	require.Equal(t, "/plain", pages[0].Path)
	require.Empty(t, pages[0].Title)
	markdown, ok := pages[0].MarkdownSource()
	require.True(t, ok)
	require.Equal(t, content, string(markdown))
}

func TestGlobPageSourceBrokenFrontmatterNamesFile(t *testing.T) {
	// Like an invalid lang, front matter that cannot be parsed has to say
	// which of possibly hundreds of files to fix.
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "broken.md"), []byte("---\ntitle: [unclosed\n---\nbody\n"), 0o600)
	require.NoError(t, err)

	source := &web.GlobPageSource{Pattern: filepath.Join(dir, "*.md")}
	for _, err := range source.Pages(context.Background(), false) {
		require.ErrorContains(t, err, "broken.md: front matter")
		return
	}
	t.Fatal("no result")
}

func TestGlobPageSourceFrontmatterLang(t *testing.T) {
	pageLang := func(t *testing.T, lang string) (string, error) {
		t.Helper()
		dir := t.TempDir()
		err := os.WriteFile(filepath.Join(dir, "page.md"), []byte("---\nlang: "+lang+"\n---\nbody\n"), 0o600)
		require.NoError(t, err)
		source := &web.GlobPageSource{Pattern: filepath.Join(dir, "*.md")}
		for page, err := range source.Pages(context.Background(), false) {
			if err != nil {
				return "", err
			}
			return page.Lang, nil
		}
		t.Fatal("no page")
		return "", nil
	}

	// Regression: Jekyll writes a locale like "en_US", which used to be
	// ignored and then failed every surface the page's language reached. It
	// is the language tag "en-US".
	for lang, want := range map[string]string{"en": "en", "de-AT": "de-AT", "en_US": "en-US", "zh_Hant_TW": "zh-Hant-TW"} {
		got, err := pageLang(t, lang)
		require.NoErrorf(t, err, "lang %q", lang)
		require.Equalf(t, want, got, "lang %q", lang)
	}

	// Anything else is an error naming the file, like broken front matter.
	for _, lang := range []string{"en US", "en.US", "deutschland", "x-default"} {
		_, err := pageLang(t, `"`+lang+`"`)
		require.ErrorContainsf(t, err, "page.md", "lang %q", lang)
	}
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

	source := &web.GlobPageSource{Dir: fs.File(dir), Pattern: filepath.Join("*", "*.md")}
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
		Sources: []web.PageSource{&web.GlobPageSource{Dir: fs.File(dir), Pattern: "*.*"}},
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
		Sources: []web.PageSource{&web.GlobPageSource{Dir: fs.File(dir), Pattern: "*.md"}},
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

func TestGlobPageSourceDirScopesGlob(t *testing.T) {
	// Dir is both what the glob runs below and what page paths are relative
	// to, so the two cannot name different directories.
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "blog"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "blog", "hello.md"),
		[]byte("---\ntitle: Hello\npublishDate: 2026-06-01T08:00:00Z\n---\npost\n"), 0o600))
	// A sibling directory the source must not reach into.
	outside := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(outside, "other.md"),
		[]byte("---\ntitle: Other\npublishDate: 2026-06-01T08:00:00Z\n---\nx\n"), 0o600))

	source := &web.GlobPageSource{Dir: fs.File(dir), Pattern: filepath.Join("blog", "*.md")}
	var pages []*web.Page
	for page, err := range source.Pages(context.Background(), false) {
		require.NoError(t, err)
		pages = append(pages, page)
	}
	require.Len(t, pages, 1)
	require.Equal(t, "/blog/hello", pages[0].Path)

	// Regression: Dir used to be ignored for matching, so a Pattern naming
	// another directory was globbed anyway and its files got a page path
	// derived from the wrong root — wrong URLs in the sitemap rather than an
	// empty one that shows something is off.
	source = &web.GlobPageSource{Dir: fs.File(dir), Pattern: filepath.Join(outside, "*.md")}
	for _, err := range source.Pages(context.Background(), false) {
		require.ErrorContains(t, err, "absolute")
	}

	source = &web.GlobPageSource{Dir: fs.File(dir)}
	for _, err := range source.Pages(context.Background(), false) {
		require.ErrorContains(t, err, "Pattern is empty")
	}
}

func TestGlobPageSourceSkipsDirectories(t *testing.T) {
	// A pattern wide enough to match directories must not try to read one as a
	// page.
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "images"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "hello.md"),
		[]byte("---\ntitle: Hello\npublishDate: 2026-06-01T08:00:00Z\n---\npost\n"), 0o600))

	source := &web.GlobPageSource{Dir: fs.File(dir), Pattern: "*"}
	var titles []string
	for page, err := range source.Pages(context.Background(), false) {
		require.NoError(t, err)
		titles = append(titles, page.Title)
	}
	require.Equal(t, []string{"Hello"}, titles)
}

func TestGlobPageSourceWithoutDir(t *testing.T) {
	// Without a Dir the Pattern carries the whole path and its fixed prefix is
	// the base for page paths. Regression: the unset Dir was detected by its
	// path, but the zero fs.File has the path "." rather than an empty one, so
	// the base became the working directory and every page came out with an
	// empty Path — which then failed the sitemap it was meant to feed.
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "hello.md"),
		[]byte("---\ntitle: Hello\npublishDate: 2026-06-01T08:00:00Z\n---\npost\n"), 0o600))

	source := &web.GlobPageSource{Pattern: filepath.Join(dir, "*.md")}
	var pages []*web.Page
	for page, err := range source.Pages(context.Background(), false) {
		require.NoError(t, err)
		pages = append(pages, page)
	}
	require.Len(t, pages, 1)
	require.Equal(t, "/hello", pages[0].Path)

	site := &web.Site{BaseURL: "https://example.com", Sources: []web.PageSource{source}}
	sitemap, err := site.Sitemap(context.Background())
	require.NoError(t, err)
	require.Len(t, sitemap.URLs, 1)
	require.Equal(t, "https://example.com/hello", sitemap.URLs[0].Loc)
}
