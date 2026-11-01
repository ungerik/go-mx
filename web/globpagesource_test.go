package web_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
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

func TestGlobPageSourcePercentInDir(t *testing.T) {
	// A percent escape in Dir is part of the directory name. Regression: go-fs
	// v0.1.0 percent-decoded the path of a local File, so a Dir named "a%20b"
	// was globbed as the nonexistent "a b" and the source yielded no pages and
	// no error.
	dir := filepath.Join(t.TempDir(), "a%20b")
	require.NoError(t, os.Mkdir(dir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "x%41.md"), []byte("---\ntitle: T\n---\nbody\n"), 0o600))

	source := &web.GlobPageSource{Dir: fs.File(dir), Pattern: "*.md"}
	var pages []*web.Page
	for page, err := range source.Pages(context.Background(), true) {
		require.NoError(t, err)
		pages = append(pages, page)
	}
	require.Len(t, pages, 1)
	// The file name keeps its "%" as well, escaped in the URL path.
	require.Equal(t, "/x%2541", pages[0].Path)
	markdown, ok := pages[0].MarkdownSource()
	require.True(t, ok)
	require.Equal(t, "body\n", string(markdown))
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
	// path, but go-fs v0.1.0 gave the zero fs.File the path "." rather than an
	// empty one, so the base became the working directory and every page came
	// out with an empty Path — which then failed the sitemap it was meant to
	// feed.
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

func TestGlobPageSourceHTMLAndTextContent(t *testing.T) {
	// HTML and plain-text files have no front matter, so the content of their
	// page is the whole file, and like any content it is only read when asked
	// for.
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "page.html"), []byte("<p>page</p>"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "note.txt"), []byte("a note\n"), 0o600))

	source := &web.GlobPageSource{Dir: fs.File(dir), Pattern: "*.*"}
	contents := map[string]string{}
	for page, err := range source.Pages(context.Background(), true) {
		require.NoError(t, err)
		// Content may be a string or a []byte, %s formats both.
		contents[page.Path] = fmt.Sprintf("%s", page.Content)
	}
	require.Equal(t, map[string]string{"/page": "<p>page</p>", "/note": "a note\n"}, contents)

	var paths []string
	for page, err := range source.Pages(context.Background(), false) {
		require.NoError(t, err)
		require.Nilf(t, page.Content, "page %s", page.Path)
		paths = append(paths, page.Path)
	}
	require.ElementsMatch(t, []string{"/page", "/note"}, paths)
}

func TestGlobPageSourceCancelledContext(t *testing.T) {
	// A caller that gives up, like a sitemap request whose client went away,
	// stops the source at the next file instead of it going through the rest
	// of the site. That has to hold without content too, where an HTML or
	// text file is not read at all.
	dir := t.TempDir()
	files := map[string]string{
		"post.md":   "---\ntitle: Post\n---\npost\n",
		"page.html": "<p>page</p>",
		"note.txt":  "a note\n",
	}
	for name, content := range files {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600))
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	for _, withContent := range []bool{true, false} {
		// "*.none" matches nothing, and a cancelled ctx must not pass for an
		// empty site then either.
		for _, pattern := range []string{"*.md", "*.html", "*.txt", "*.none"} {
			source := &web.GlobPageSource{Dir: fs.File(dir), Pattern: pattern}
			var results []error
			for page, err := range source.Pages(ctx, withContent) {
				require.Nilf(t, page, "pattern %q, withContent %t", pattern, withContent)
				results = append(results, err)
			}
			require.Lenf(t, results, 1, "pattern %q, withContent %t", pattern, withContent)
			require.ErrorIsf(t, results[0], context.Canceled, "pattern %q, withContent %t", pattern, withContent)
		}
	}
}

func TestGlobPageSourceReadError(t *testing.T) {
	// A matched file that cannot be read ends the source with the error
	// instead of its page being left out without a word. Regression: without
	// content an HTML or text file was only stat'ed for its dates, and a
	// failed stat gave it no dates, which dropped it from the sitemap.
	for _, withContent := range []bool{true, false} {
		for _, name := range []string{"broken.md", "broken.html", "broken.txt"} {
			dir := t.TempDir()
			// The glob matches a dangling symlink, but reading it fails.
			err := os.Symlink(filepath.Join(dir, "missing"), filepath.Join(dir, name))
			if err != nil {
				// Windows allows symlinks only with Developer Mode or elevation.
				t.Skipf("cannot create a symlink: %s", err)
			}

			source := &web.GlobPageSource{Dir: fs.File(dir), Pattern: "broken.*"}
			var results []error
			for page, err := range source.Pages(context.Background(), withContent) {
				require.Nilf(t, page, "file %s, withContent %t", name, withContent)
				results = append(results, err)
			}
			require.Lenf(t, results, 1, "file %s, withContent %t", name, withContent)
			require.ErrorIsf(t, results[0], os.ErrNotExist, "file %s, withContent %t", name, withContent)
		}
	}
}

func TestGlobPageSourceUnreadableFile(t *testing.T) {
	// A page file that has a stat but cannot be read ends the source with the
	// error. Every page file is opened once, also when no content is read,
	// so the source fails without content too, as the page would when served.
	for _, withContent := range []bool{true, false} {
		for _, name := range []string{"locked.md", "locked.html", "locked.txt"} {
			file := filepath.Join(t.TempDir(), name)
			require.NoError(t, os.WriteFile(file, []byte("locked\n"), 0o000))
			_, err := os.ReadFile(file)
			if err == nil {
				t.Skip("root and Windows read a file without read permission")
			}

			source := &web.GlobPageSource{Dir: fs.File(filepath.Dir(file)), Pattern: "locked.*"}
			var results []error
			for page, err := range source.Pages(context.Background(), withContent) {
				require.Nilf(t, page, "file %s, withContent %t", name, withContent)
				results = append(results, err)
			}
			require.Lenf(t, results, 1, "file %s, withContent %t", name, withContent)
			require.ErrorIsf(t, results[0], os.ErrPermission, "file %s, withContent %t", name, withContent)
		}
	}
}

func TestGlobPageSourceCancelledWhileIterating(t *testing.T) {
	// A caller that cancels while it iterates gets the context's error
	// instead of the next page. HTML files without content are not read, so
	// only the check before each file can notice.
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.html"), []byte("<p>a</p>"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "b.html"), []byte("<p>b</p>"), 0o600))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	source := &web.GlobPageSource{Dir: fs.File(dir), Pattern: "*.html"}
	var paths []string
	var results []error
	for page, err := range source.Pages(ctx, false) {
		if err != nil {
			results = append(results, err)
			continue
		}
		paths = append(paths, page.Path)
		cancel()
	}
	require.Equal(t, []string{"/a"}, paths)
	require.Len(t, results, 1)
	require.ErrorIs(t, results[0], context.Canceled)
}

func TestGlobPageSourceSkipsSpecialFiles(t *testing.T) {
	// Only a regular file is a page. Regression: a named pipe or a symlink to
	// a device with a page extension was read like a file, so a pipe blocked
	// the source past any deadline and /dev/zero was read without end.
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "hello.md"),
		[]byte("---\ntitle: Hello\n---\npost\n"), 0o600))
	err := os.Symlink(os.DevNull, filepath.Join(dir, "device.md"))
	if err != nil {
		t.Skipf("cannot create a symlink: %s", err)
	}
	// A file of no page type is not looked at, so an unreadable one that a
	// wide pattern matches does not fail the source.
	require.NoError(t, os.Symlink(filepath.Join(dir, "missing"), filepath.Join(dir, "image.png")))

	source := &web.GlobPageSource{Dir: fs.File(dir), Pattern: "*"}
	var titles []string
	for page, err := range source.Pages(context.Background(), true) {
		require.NoError(t, err)
		titles = append(titles, page.Title)
	}
	require.Equal(t, []string{"Hello"}, titles)
}

func TestGlobPageSourceSkipsHiddenFiles(t *testing.T) {
	// Like a shell glob, a wildcard does not match a name beginning with ".",
	// so a wide pattern does not publish dotfiles. Regression: Pattern "*"
	// published ".secret.txt" as the page "/.secret". A pattern that names
	// the dot still matches, as does a Dir below a hidden directory.
	dir := filepath.Join(t.TempDir(), ".site")
	for name, content := range map[string]string{
		"hello.md":                 "---\ntitle: Hello\n---\npost\n",
		".secret.txt":              "secret",
		".private/draft.md":        "---\ntitle: Private\n---\ndraft\n",
		".well-known/security.txt": "Contact: mailto:security@example.com\n",
	} {
		require.NoError(t, os.MkdirAll(filepath.Dir(filepath.Join(dir, name)), 0o700))
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600))
	}

	pagePaths := func(pattern string) []string {
		t.Helper()
		var paths []string
		source := &web.GlobPageSource{Dir: fs.File(dir), Pattern: pattern}
		for page, err := range source.Pages(context.Background(), false) {
			require.NoError(t, err)
			paths = append(paths, page.Path)
		}
		return paths
	}
	require.Equal(t, []string{"/hello"}, pagePaths("*"))
	require.Empty(t, pagePaths(filepath.Join("*", "*.md")))
	require.Equal(t, []string{"/.well-known/security"}, pagePaths(filepath.Join(".well-known", "*.txt")))
	require.Equal(t, []string{"/.secret"}, pagePaths(".*"))
	if filepath.Separator == '/' {
		// An escaped dot names the dot too, where a backslash is no separator.
		require.Equal(t, []string{"/.secret"}, pagePaths(`\.secret.txt`))
	}
}

func TestGlobPageSourceInvalidDir(t *testing.T) {
	// The files are matched on the local disk, so a Dir that names no local
	// directory is an error. Regression: the glob matched nothing below such
	// a Dir, so the source was empty and a sitemap of it had no page and no
	// error. go-fs v1 gives a URI of a file system that is not local its URI
	// as path, and a Dir that is missing, like a relative one under a service
	// started in another working directory, or a file is no better.
	dir := t.TempDir()
	file := filepath.Join(dir, "page.md")
	require.NoError(t, os.WriteFile(file, []byte("page\n"), 0o600))
	dirError := func(t *testing.T, sourceDir fs.File) error {
		t.Helper()
		source := &web.GlobPageSource{Dir: sourceDir, Pattern: "*.md"}
		var results []error
		for page, err := range source.Pages(context.Background(), false) {
			require.Nil(t, page)
			results = append(results, err)
		}
		require.Len(t, results, 1)
		return results[0]
	}

	require.ErrorContains(t, dirError(t, "nofs://host/content"), `GlobPageSource.Dir "nofs://host/content"`)
	require.ErrorIs(t, dirError(t, fs.File(filepath.Join(dir, "missing"))), os.ErrNotExist)
	require.ErrorContains(t, dirError(t, fs.File(file)), "is no directory")

	// The error goes to logs, so a password in the URI is redacted.
	err := dirError(t, "sftp://deploy:hunter2@example.com/site")
	require.ErrorContains(t, err, `GlobPageSource.Dir "sftp://example.com/site"`)
	require.NotContains(t, err.Error(), "hunter2")

	// Regression: only a password was redacted, so a token in the user info
	// or in the query went to the logs.
	for _, uri := range []string{
		"https://ghp_SECRET@github.com/x",
		"https://example.com/x?access_token=SECRET",
		"https://example.com/x#SECRET",
		"sftp://deploy:SECRET@example.com/%zz",
		// On Windows go-fs reads this as the local UNC path
		// `\\user:SECRET@server\share`, which every other Dir error printed.
		"file://user:SECRET@server/share",
	} {
		err := dirError(t, fs.File(uri))
		require.ErrorContains(t, err, "GlobPageSource.Dir", uri)
		require.NotContains(t, err.Error(), "SECRET", uri)
	}
}

// A matched path is read with os, as it is spelled. Regression: it was read
// through go-fs, which parsed a path spelling a file URI as that URI, so a
// local decoy named "file:" was stat'ed while the file the URI named was
// read and published.
func TestGlobPageSourceURIShapedPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("a colon is no character of a Windows file name")
	}
	secret := filepath.Join(t.TempDir(), "secret")
	require.NoError(t, os.Mkdir(secret, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(secret, "page.md"), []byte("secret\n"), 0o600))
	work := t.TempDir()
	decoy := filepath.Join(work, "file:", "localhost"+secret)
	require.NoError(t, os.MkdirAll(decoy, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(decoy, "page.md"), []byte("local\n"), 0o600))
	t.Chdir(work)

	source := &web.GlobPageSource{Pattern: "file://localhost" + secret + "/page.md"}
	var contents []string
	for page, err := range source.Pages(context.Background(), true) {
		require.NoError(t, err)
		contents = append(contents, fmt.Sprintf("%s", page.Content))
	}
	require.Equal(t, []string{"local\n"}, contents)
}

// cancelAfterContext is a context whose Err reports context.Canceled from
// its nth call on, which cancels the source at a point inside its work
// rather than before it starts.
type cancelAfterContext struct {
	context.Context
	n     int32
	calls atomic.Int32
}

func (c *cancelAfterContext) Err() error {
	if c.calls.Add(1) >= c.n {
		return context.Canceled
	}
	return nil
}

// The glob checks the context while it searches the directories, and a
// read checks it between the chunks of a file. Regression: the context was
// checked only between the glob, the stats and the reads, so a huge
// directory or a large file held a cancelled request until it was done.
func TestGlobPageSourceCancelledInsideGlobAndRead(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("needs Unix permissions that bind the user")
	}
	dir := t.TempDir()
	for _, sub := range []string{"a", "locked"} {
		require.NoError(t, os.Mkdir(filepath.Join(dir, sub), 0o700))
		require.NoError(t, os.WriteFile(filepath.Join(dir, sub, "page.md"), []byte("page\n"), 0o600))
	}
	require.NoError(t, os.Chmod(filepath.Join(dir, "locked"), 0))
	t.Cleanup(func() { os.Chmod(filepath.Join(dir, "locked"), 0o700) })

	// The first check passes, so only a check inside the glob can stop it
	// before it reaches the directory it cannot read.
	ctx := &cancelAfterContext{Context: context.Background(), n: 2}
	var results []error
	for page, err := range (&web.GlobPageSource{Dir: fs.File(dir), Pattern: "*/*.md"}).Pages(ctx, false) {
		require.Nil(t, page)
		results = append(results, err)
	}
	require.Len(t, results, 1)
	require.ErrorIs(t, results[0], context.Canceled)

	// A file of 4 MiB is read in 64 chunks, far more checks than come
	// before the read.
	big := filepath.Join(t.TempDir(), "big.md")
	require.NoError(t, os.WriteFile(big, []byte(strings.Repeat("x", 4<<20)), 0o600))
	ctx = &cancelAfterContext{Context: context.Background(), n: 20}
	results = nil
	for page, err := range (&web.GlobPageSource{Dir: fs.File(filepath.Dir(big)), Pattern: "*.md"}).Pages(ctx, true) {
		require.Nil(t, page)
		results = append(results, err)
	}
	require.Len(t, results, 1)
	require.ErrorIs(t, results[0], context.Canceled)
}

// Dir is the directory the glob runs below, not a pattern. Regression: Dir
// and Pattern were joined into one pattern, so a directory named
// "content[prod]" matched "contentp" and published its files, and one named
// "site[1]" matched nothing.
func TestGlobPageSourceGlobCharactersInDir(t *testing.T) {
	parent := t.TempDir()
	for name, content := range map[string]string{
		"content[prod]/real.md": "real\n",
		"contentp/other.md":     "other\n",
		"site[1]/page.md":       "page\n",
		"site1/decoy.md":        "decoy\n",
	} {
		file := filepath.Join(parent, filepath.FromSlash(name))
		require.NoError(t, os.MkdirAll(filepath.Dir(file), 0o700))
		require.NoError(t, os.WriteFile(file, []byte(content), 0o600))
	}
	for _, confine := range []bool{false, true} {
		for dir, want := range map[string][]string{"content[prod]": {"/real"}, "site[1]": {"/page"}} {
			var paths []string
			source := &web.GlobPageSource{Dir: fs.File(filepath.Join(parent, dir)), Pattern: "*.md", Confine: confine}
			for page, err := range source.Pages(context.Background(), true) {
				require.NoError(t, err, "Dir %s, Confine %t", dir, confine)
				paths = append(paths, page.Path)
			}
			require.Equal(t, want, paths, "Dir %s, Confine %t", dir, confine)
		}
	}
	// A pattern without a wildcard names one file below Dir, also literally.
	var paths []string
	for page, err := range (&web.GlobPageSource{Dir: fs.File(filepath.Join(parent, "site[1]")), Pattern: "page.md"}).Pages(context.Background(), false) {
		require.NoError(t, err)
		paths = append(paths, page.Path)
	}
	require.Equal(t, []string{"/page"}, paths)
}

func TestGlobPageSourceFileGoneWhileIterating(t *testing.T) {
	// A file the glob listed that is gone when its turn comes, deleted or
	// replaced while a deploy runs, is no page rather than an error that
	// would fail the sitemap and every other page with it.
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.html"), []byte("<p>a</p>"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "b.html"), []byte("<p>b</p>"), 0o600))

	source := &web.GlobPageSource{Dir: fs.File(dir), Pattern: "*.html"}
	var paths []string
	for page, err := range source.Pages(context.Background(), false) {
		require.NoError(t, err)
		paths = append(paths, page.Path)
		require.NoError(t, os.Remove(filepath.Join(dir, "b.html")))
	}
	require.Equal(t, []string{"/a"}, paths)
}

func TestGlobPageSourceLiteralTildeDir(t *testing.T) {
	// The page of a matched file has the content of that file. Regression: a
	// matched path was handed back to go-fs, which expands a leading "~" of a
	// relative path to the home directory, so a Dir of "./~/content" globbed
	// the directory named "~" but read the same-named files of the home
	// directory and published those.
	root := t.TempDir()
	home := filepath.Join(root, "home")
	work := filepath.Join(root, "work")
	require.NoError(t, os.MkdirAll(filepath.Join(home, "content"), 0o700))
	require.NoError(t, os.MkdirAll(filepath.Join(work, "~", "content"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(home, "content", "a.md"), []byte("home\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(work, "~", "content", "a.md"), []byte("literal\n"), 0o600))
	t.Setenv("HOME", home)
	t.Chdir(work)

	source := &web.GlobPageSource{Dir: fs.File("./~/content"), Pattern: "*.md"}
	var pages []*web.Page
	for page, err := range source.Pages(context.Background(), true) {
		require.NoError(t, err)
		pages = append(pages, page)
	}
	require.Len(t, pages, 1)
	markdown, ok := pages[0].MarkdownSource()
	require.True(t, ok)
	require.Equal(t, "literal\n", string(markdown))
}

// Confine keeps the source inside Dir: a symlink to a file or directory
// outside of it fails the source, and a Pattern reaching above it is an
// error, while a relative symlink that stays below Dir is followed. Without Confine, Dir is trusted and the symlink publishes
// the file it points to.
func TestGlobPageSourceConfine(t *testing.T) {
	page := []byte("---\ntitle: Page\npublishDate: 2026-06-01T08:00:00Z\n---\nbody\n")
	outside := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("secret"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(outside, "other.md"), page, 0o600))

	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "blog"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "blog", "hello.md"), page, 0o600))
	// os.Root follows only a relative symlink.
	require.NoError(t, os.Symlink("hello.md", filepath.Join(dir, "blog", "alias.md")))
	require.NoError(t, os.Symlink(outside, filepath.Join(dir, "linked")))

	paths := func(t *testing.T, source *web.GlobPageSource) []string {
		t.Helper()
		var paths []string
		for page, err := range source.Pages(context.Background(), true) {
			require.NoError(t, err)
			require.Equal(t, "body\n", fmt.Sprintf("%s", page.Content))
			paths = append(paths, page.Path)
		}
		return paths
	}
	require.Equal(t, []string{"/blog/alias", "/blog/hello"},
		paths(t, &web.GlobPageSource{Dir: fs.File(dir), Pattern: "blog/*.md", Confine: true}))
	var linkErr error
	for _, err := range (&web.GlobPageSource{Dir: fs.File(dir), Pattern: "*/*.md", Confine: true}).Pages(context.Background(), false) {
		linkErr = err
	}
	require.ErrorContains(t, linkErr, "linked", "a symlinked directory outside of Dir must fail the source")
	require.Equal(t, []string{"/blog/alias", "/blog/hello", "/linked/other"},
		paths(t, &web.GlobPageSource{Dir: fs.File(dir), Pattern: "*/*.md"}),
		"without Confine, Dir is trusted")

	require.NoError(t, os.Symlink(filepath.Join(outside, "secret.txt"), filepath.Join(dir, "leak.txt")))
	var leakErr error
	for page, err := range (&web.GlobPageSource{Dir: fs.File(dir), Pattern: "*.txt", Confine: true}).Pages(context.Background(), true) {
		require.Nil(t, page, "a symlink out of a confined Dir must not publish its target")
		leakErr = err
	}
	require.ErrorContains(t, leakErr, "leak.txt")

	for _, source := range []*web.GlobPageSource{
		{Dir: fs.File(dir), Pattern: "../*/*.md", Confine: true},
		{Dir: fs.File(dir), Pattern: "blog/../../*.md", Confine: true},
		{Pattern: filepath.Join(dir, "blog", "*.md"), Confine: true},
	} {
		var failed bool
		for _, err := range source.Pages(context.Background(), false) {
			require.Error(t, err, "Pattern %q", source.Pattern)
			failed = true
		}
		require.True(t, failed, "Pattern %q must fail", source.Pattern)
	}
	require.Equal(t, []string{"/blog/alias", "/blog/hello"},
		paths(t, &web.GlobPageSource{Dir: fs.File(dir), Pattern: "linked/../blog/*.md", Confine: true}),
		"a .. that stays below Dir is allowed")
}

// A directory the source cannot read fails it, Dir itself or one a wildcard
// matched, instead of leaving its pages out of every sitemap without a word.
// A hidden directory is not searched, so one that cannot be read is no
// error either.
func TestGlobPageSourceUnreadableDir(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("needs Unix permissions that bind the user")
	}
	page := []byte("---\ntitle: Page\npublishDate: 2026-06-01T08:00:00Z\n---\nbody\n")
	dir := t.TempDir()
	for _, sub := range []string{"blog", "locked", ".cache"} {
		require.NoError(t, os.MkdirAll(filepath.Join(dir, sub), 0o700))
		require.NoError(t, os.WriteFile(filepath.Join(dir, sub, "page.md"), page, 0o600))
	}
	for _, sub := range []string{"locked", ".cache"} {
		require.NoError(t, os.Chmod(filepath.Join(dir, sub), 0))
		t.Cleanup(func() { os.Chmod(filepath.Join(dir, sub), 0o700) })
	}

	firstErr := func(source *web.GlobPageSource) error {
		for _, err := range source.Pages(context.Background(), false) {
			if err != nil {
				return err
			}
		}
		return nil
	}
	require.ErrorContains(t, firstErr(&web.GlobPageSource{Dir: fs.File(dir), Pattern: "*/*.md"}), "locked")
	require.ErrorContains(t, firstErr(&web.GlobPageSource{Dir: fs.File(dir), Pattern: "*/*.md", Confine: true}), "locked")
	require.ErrorContains(t, firstErr(&web.GlobPageSource{Dir: fs.File(filepath.Join(dir, "locked")), Pattern: "*.md"}), "permission denied")
	require.NoError(t, firstErr(&web.GlobPageSource{Dir: fs.File(dir), Pattern: "blog/*.md"}))

	require.NoError(t, os.Chmod(filepath.Join(dir, "locked"), 0o700))
	require.NoError(t, firstErr(&web.GlobPageSource{Dir: fs.File(dir), Pattern: "*/*.md"}),
		"an unreadable hidden directory is not searched")
}

// A Pattern without Dir that names a hidden directory with an escaped dot
// gets its page paths below the fixed prefix before the escape, as it does
// with Dir.
func TestGlobPageSourceEscapedDotWithoutDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("a backslash is a separator on Windows")
	}
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".well-known"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".well-known", "security.txt"), []byte("Contact: x"), 0o600))
	var paths []string
	for page, err := range (&web.GlobPageSource{Pattern: dir + `/\.well-known/*.txt`}).Pages(context.Background(), false) {
		require.NoError(t, err)
		paths = append(paths, page.Path)
	}
	require.Equal(t, []string{"/.well-known/security"}, paths)
}

// A page file past MaxFileSize fails the source, also when no content is
// read. Without content only the head of a Markdown file is read for its
// front matter, and one whose front matter ends past the head is read whole.
func TestGlobPageSourceMaxFileSize(t *testing.T) {
	dir := t.TempDir()
	long := strings.Repeat("x", 100<<10)
	files := map[string]string{
		"short.md":      "---\ntitle: Short\npublishDate: 2026-06-01T08:00:00Z\n---\n" + long + "\n",
		"longmatter.md": "---\ntitle: Long\ndescription: " + long + "\npublishDate: 2026-06-01T08:00:00Z\n---\nbody\n",
		"plain.md":      long + "\n",
	}
	for name, content := range files {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600))
	}
	titles := map[string]string{}
	for page, err := range (&web.GlobPageSource{Dir: fs.File(dir), Pattern: "*.md"}).Pages(context.Background(), false) {
		require.NoError(t, err)
		require.Nil(t, page.Content)
		titles[page.Path] = page.Title
	}
	require.Equal(t, map[string]string{"/short": "Short", "/longmatter": "Long", "/plain": ""}, titles)

	for _, withContent := range []bool{true, false} {
		var sizeErr error
		for _, err := range (&web.GlobPageSource{Dir: fs.File(dir), Pattern: "short.md", MaxFileSize: 1 << 10}).Pages(context.Background(), withContent) {
			sizeErr = err
		}
		require.ErrorContains(t, sizeErr, "short.md", "withContent %t", withContent)
	}
	for _, err := range (&web.GlobPageSource{Dir: fs.File(dir), Pattern: "*.md", MaxFileSize: -1}).Pages(context.Background(), true) {
		require.NoError(t, err, "a negative MaxFileSize is no limit")
	}
}

// Serving a page's Markdown alternate reads that page's file alone, not
// every file the glob matched before it: a file that has grown past
// MaxFileSize since the routes were registered fails only its own page.
func TestGlobPageSourceMarkdownAlternateReadsOnlyItsFile(t *testing.T) {
	dir := t.TempDir()
	page := func(title string) string {
		return "---\ntitle: " + title + "\npublishDate: 2026-06-01T08:00:00Z\n---\nText of " + title + ".\n"
	}
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.md"), []byte(page("A")), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "b.md"), []byte(page("B")), 0o600))
	site := &web.Site{
		BaseURL:            "https://example.com",
		Sources:            []web.PageSource{&web.GlobPageSource{Dir: fs.File(dir), Pattern: "*.md", MaxFileSize: 1 << 10}},
		MarkdownAlternates: true,
	}
	mux := http.NewServeMux()
	require.NoError(t, site.RegisterRoutes(mux))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.md"), []byte(page(strings.Repeat("A", 2<<10))), 0o600))

	get := func(path string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		return rec
	}
	rec := get("/b.md")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Contains(t, rec.Body.String(), "Text of B.")
	require.Equal(t, http.StatusInternalServerError, get("/a.md").Code)
}

// Every field of the front matter is kept in Page.FrontMatter, the ones of
// no page field too, which a renderer reading fields of its own needs now
// that the front matter is not part of the content.
func TestGlobPageSourceFrontMatterFields(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"yaml.md": "---\ntitle: YAML\nimage: /img/yaml.png\naliases: [/old]\nhero:\n  alt: A hero\n  size: 2\n---\nbody\n",
		"toml.md": "+++\ntitle = \"TOML\"\nimage = \"/img/toml.png\"\n[hero]\nalt = \"A hero\"\n+++\nbody\n",
		"json.md": "{\n\"title\": \"JSON\",\n\"image\": \"/img/json.png\",\n\"hero\": {\"alt\": \"A hero\"}\n}\n\nbody\n",
		"none.md": "# No front matter\n",
	}
	for name, content := range files {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600))
	}
	for _, withContent := range []bool{true, false} {
		fields := map[string]map[string]any{}
		for page, err := range (&web.GlobPageSource{Dir: fs.File(dir), Pattern: "*.md"}).Pages(context.Background(), withContent) {
			require.NoError(t, err)
			fields[page.Path] = page.FrontMatter
		}
		require.Nil(t, fields["/none"], "withContent %t", withContent)
		for _, format := range []string{"yaml", "toml", "json"} {
			matter := fields["/"+format]
			require.NotNil(t, matter, "%s, withContent %t", format, withContent)
			require.Equal(t, strings.ToUpper(format), matter["title"], "a mapped field is kept too")
			require.Equal(t, "/img/"+format+".png", matter["image"])
			// A nested object is a map[string]any whatever the format; the
			// YAML decoder makes it a map[any]any.
			hero, ok := matter["hero"].(map[string]any)
			require.True(t, ok, "%s: hero is %T", format, matter["hero"])
			require.Equal(t, "A hero", hero["alt"])
		}
		require.Equal(t, []any{"/old"}, fields["/yaml"]["aliases"])
		require.Equal(t, 2, fields["/yaml"]["hero"].(map[string]any)["size"])
	}
}

// A cancelled context is checked before anything else, so a request nobody
// waits for does not stat Dir or run the glob, and gets the context's error
// rather than the one of a missing Dir.
func TestGlobPageSourceCancelledBeforeGlob(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	source := &web.GlobPageSource{Dir: fs.File(filepath.Join(t.TempDir(), "missing")), Pattern: "*.md"}
	var results []error
	for page, err := range source.Pages(ctx, false) {
		require.Nil(t, page)
		results = append(results, err)
	}
	require.Len(t, results, 1)
	require.ErrorIs(t, results[0], context.Canceled)
	require.NotErrorIs(t, results[0], os.ErrNotExist)
}

// A missing directory in the fixed part of Pattern fails the source like a
// missing Dir does, with or without Dir, rather than leaving an empty source.
func TestGlobPageSourceMissingPatternDir(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "blog"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "blog", "hello.md"), []byte("hello\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "page.md"), []byte("page\n"), 0o600))
	results := func(t *testing.T, source *web.GlobPageSource) (paths []string, errs []error) {
		t.Helper()
		for page, err := range source.Pages(context.Background(), false) {
			if err != nil {
				errs = append(errs, err)
				continue
			}
			paths = append(paths, page.Path)
		}
		return paths, errs
	}

	for _, source := range []*web.GlobPageSource{
		{Dir: fs.File(dir), Pattern: filepath.Join("news", "*.md")},
		{Dir: fs.File(dir), Pattern: filepath.Join("blog", "2026", "*.md")},
		{Pattern: filepath.Join(dir, "news", "*.md")},
		{Dir: fs.File(dir), Pattern: filepath.Join("news", "*.md"), Confine: true},
		// The directory of a pattern without a wildcard.
		{Dir: fs.File(dir), Pattern: filepath.Join("news", "hello.md")},
	} {
		paths, errs := results(t, source)
		require.Empty(t, paths, "Pattern %q", source.Pattern)
		require.Len(t, errs, 1, "Pattern %q", source.Pattern)
		require.ErrorIs(t, errs[0], os.ErrNotExist, "Pattern %q", source.Pattern)
		require.ErrorContains(t, errs[0], "GlobPageSource.Pattern", "Pattern %q", source.Pattern)
	}

	_, errs := results(t, &web.GlobPageSource{Dir: fs.File(dir), Pattern: filepath.Join("page.md", "*.md")})
	require.Len(t, errs, 1)
	require.ErrorContains(t, errs[0], "is no directory")

	// What a wildcard has to match, and the file a pattern without one names,
	// may well be missing: that is an empty match, not a broken source.
	for _, source := range []*web.GlobPageSource{
		{Dir: fs.File(dir), Pattern: filepath.Join("*", "2026", "*.md")},
		{Dir: fs.File(dir), Pattern: filepath.Join("blog", "missing.md")},
		{Dir: fs.File(dir), Pattern: "missing.md"},
	} {
		paths, errs := results(t, source)
		require.Empty(t, paths, "Pattern %q", source.Pattern)
		require.Empty(t, errs, "Pattern %q", source.Pattern)
	}

	// An existing prefix matches as before.
	for _, source := range []*web.GlobPageSource{
		{Dir: fs.File(dir), Pattern: filepath.Join("blog", "*.md")},
		{Dir: fs.File(dir), Pattern: filepath.Join("blog", "*.md"), Confine: true},
		{Pattern: filepath.Join(dir, "blog", "*.md")},
	} {
		paths, errs := results(t, source)
		require.Empty(t, errs, "Pattern %q", source.Pattern)
		require.Len(t, paths, 1, "Pattern %q", source.Pattern)
	}
}
