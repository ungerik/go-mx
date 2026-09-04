package web

import (
	"bytes"
	"context"
	"errors"
	"iter"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/adrg/frontmatter"
	"github.com/domonda/go-errs"
	"github.com/ungerik/go-fs"
	"github.com/ungerik/go-mx"
)

// https://github.com/adrg/frontmatter
// https://github.com/yuin/goldmark

// GlobPageSource is a PageSource that yields pages from files matching a glob
// Pattern below Dir. Markdown files are parsed for frontmatter metadata, and
// HTML and plain-text files are loaded with their respective content types.
//
//	source := &web.GlobPageSource{Dir: fs.File("content/blog"), Pattern: "*.md"}
//	// content/blog/hello.md -> Page{Path: "/hello"}
type GlobPageSource struct {
	// Dir is the directory the source reads from. It is both what Pattern is
	// matched below and what the URL path of a page is relative to, so the two
	// cannot disagree: "content/blog/hello.md" matched by a Dir of
	// "content/blog" and a Pattern of "*.md" becomes the page path "/hello".
	//
	// An empty Dir keeps Pattern as the whole path to match, relative to the
	// working directory, and takes the base for page paths from the part of
	// Pattern before its first wildcard — the same result written differently.
	Dir fs.File

	// Pattern is the glob matched below Dir, in the syntax of filepath.Match.
	// It must be relative when Dir is set.
	Pattern string

	PageType      string
	DefaultAuthor string
}

// Pages iterates over the files matching the glob Pattern and yields a Page for
// each supported file type, implementing the PageSource interface.
func (s *GlobPageSource) Pages(ctx context.Context, withContent bool) iter.Seq2[*Page, error] {
	return func(yield func(*Page, error) bool) {
		pattern, err := s.globPattern()
		if err != nil {
			yield(nil, err)
			return
		}
		files, err := filepath.Glob(pattern)
		if err != nil {
			yield(nil, err)
			return
		}
		for _, file := range files {
			content := fs.File(file)
			// A directory has no content to render. A pattern like "*.md" does
			// not match one anyway; a wider pattern like "*" does, and skipping
			// it here keeps that from being read as a file.
			if content.IsDir() {
				continue
			}
			switch content.ExtLower() {
			case ".md":
				data, err := content.ReadAll()
				if err != nil {
					yield(nil, err)
					return
				}
				var matter struct {
					Title       string    `yaml:"title"       toml:"title"       json:"title"`
					Description string    `yaml:"description" toml:"description" json:"description"`
					Date        time.Time `yaml:"date"        toml:"date"        json:"date"`
					PublishDate time.Time `yaml:"publishDate" toml:"publishDate" json:"publishDate"`
					Draft       bool      `yaml:"draft"       toml:"draft"       json:"draft"`
					Tags        []string  `yaml:"tags"        toml:"tags"        json:"tags"`
				}
				_, err = frontmatter.Parse(bytes.NewReader(data), &matter)
				if errors.Is(err, frontmatter.ErrNotFound) {
				}
				if err != nil {
					yield(nil, err)
					return
				}
				page := &Page{
					Path:        s.pagePath(file),
					ContentType: mx.ContentTypeMarkdown,
					Type:        s.PageType,
					Author:      s.DefaultAuthor,
					Title:       matter.Title,
					Description: matter.Description,
					Created:     matter.Date,
					LastUpdated: matter.Date,
					Published:   matter.PublishDate,
					Tags:        matter.Tags,
				}
				if matter.Draft {
					page.Published = time.Time{}
				}
				if withContent {
					page.Content = data
				}
				if !yield(page, nil) {
					return
				}

			case ".html", ".htm":
				// A file without frontmatter carries no dates, so its
				// modification time is the only publication signal there is.
				// Without it the page would never be [Page.Indexable] and
				// would silently stay out of every sitemap.
				modified := content.Modified()
				page := &Page{
					Path:        s.pagePath(file),
					ContentType: mx.ContentTypeHTML,
					Type:        s.PageType,
					Author:      s.DefaultAuthor,
					Created:     modified,
					LastUpdated: modified,
					Published:   modified,
				}
				// TODO parse HTML title
				if withContent {
					page.Content, err = content.ReadAll()
					if err != nil {
						yield(nil, err)
						return
					}
				}
				if !yield(page, nil) {
					return
				}

			case ".txt":
				modified := content.Modified()
				page := &Page{
					Path:        s.pagePath(file),
					ContentType: mx.ContentTypePlainText,
					Type:        s.PageType,
					Title:       content.Name(), // Use filename as title
					Author:      s.DefaultAuthor,
					Created:     modified,
					LastUpdated: modified,
					Published:   modified,
				}
				if withContent {
					page.Content, err = content.ReadAll()
					if err != nil {
						yield(nil, err)
						return
					}
				}
				if !yield(page, nil) {
					return
				}
			}
		}
	}
}

// pagePath derives the URL path of a page from the path of its file: the path
// relative to the source's base directory, with the file extension dropped and
// an "index" file mapping to its directory. "blog/hello.md" below the base
// directory becomes "/blog/hello", "blog/index.md" becomes "/blog/".
//
// A [Site] needs the path to build a page's canonical URL and its sitemap
// entry, so a source that leaves it empty makes every page it yields
// unreachable for both. It returns an empty path only if the file lies outside
// the base directory, which cannot happen for a file the glob matched.
func (s *GlobPageSource) pagePath(file string) string {
	rel, err := filepath.Rel(s.baseDir(), file)
	if err != nil {
		return ""
	}
	rel = filepath.ToSlash(rel)
	// A file outside the base directory has no path within the site. Cleaning
	// it would drop the leading ".." and invent a path that looks legitimate,
	// so it yields an empty path instead, which fails loudly wherever it is
	// needed rather than landing in a sitemap as a wrong URL.
	if rel == ".." || strings.HasPrefix(rel, "../") {
		return ""
	}
	p := path.Clean("/" + escapePathSegments(rel))
	p = strings.TrimSuffix(p, path.Ext(p))
	// An index file is the page of its directory, addressed by the directory
	// URL with its trailing slash, not by a URL ending in "/index".
	if path.Base(p) == "index" {
		p = path.Dir(p)
		if !strings.HasSuffix(p, "/") {
			p += "/"
		}
	}
	return p
}

// baseDir returns the directory page paths are relative to: Dir if it is set,
// otherwise the fixed directory prefix of Pattern, that is everything before
// its first wildcard.
func (s *GlobPageSource) baseDir() string {
	if dir := s.dirPath(); dir != "" {
		return dir
	}
	pattern := s.Pattern
	if i := strings.IndexAny(pattern, "*?["); i >= 0 {
		pattern = pattern[:i]
	}
	return filepath.Dir(pattern)
}

// dirPath returns the path of Dir, or "" if no Dir is set. The zero fs.File
// has the path "." — the working directory — which is a real directory to read
// from and not the same thing as "no directory given", so the unset case is
// decided on the value rather than on its path.
func (s *GlobPageSource) dirPath() string {
	if s.Dir == "" {
		return ""
	}
	return s.Dir.Path()
}

// globPattern returns the pattern to match files with: Pattern below Dir, or
// Pattern itself when Dir is empty.
//
// An absolute Pattern together with a Dir is an error rather than a silent
// choice between them. The two would name different directories, and because
// Dir is also what page paths are relative to, the pages of the matched files
// would either be dropped or get a path derived from the wrong root — wrong
// URLs in the sitemap, not an empty one that shows something is off.
func (s *GlobPageSource) globPattern() (string, error) {
	if s.Pattern == "" {
		return "", errs.New("GlobPageSource.Pattern is empty")
	}
	dir := s.dirPath()
	if dir == "" {
		return s.Pattern, nil
	}
	if filepath.IsAbs(s.Pattern) {
		return "", errs.Errorf("GlobPageSource.Pattern %q is absolute, it must be relative to Dir %q", s.Pattern, dir)
	}
	return filepath.Join(dir, s.Pattern), nil
}

// escapePathSegments percent-escapes every segment of a slash separated file
// path so the result is usable as a URL path. File names are not URL paths:
// a "100%-off.md" would otherwise become an invalid escape sequence and a
// "faq?.md" a query, and both would be rejected wherever the page path is
// turned into a URL.
func escapePathSegments(p string) string {
	segments := strings.Split(p, "/")
	for i, segment := range segments {
		segments[i] = url.PathEscape(segment)
	}
	return strings.Join(segments, "/")
}
