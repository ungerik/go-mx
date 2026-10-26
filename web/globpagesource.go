package web

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"iter"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/adrg/frontmatter"
	"github.com/domonda/go-errs"
	"github.com/ungerik/go-fs"
	"github.com/ungerik/go-fs/fsimpl"
	"github.com/ungerik/go-mx"
)

// https://github.com/adrg/frontmatter
// https://github.com/yuin/goldmark

// GlobPageSource is a PageSource that yields pages from files matching a glob
// Pattern below Dir. Markdown files are parsed for frontmatter metadata (title,
// description, lang, date, publishDate, draft, tags), which is not part of the
// page content, and HTML and plain-text files are loaded with their respective
// content types. A lang holding a locale like "en_US" becomes the language tag
// "en-US", and one that is no language tag fails the source with the name of
// the file, like front matter that cannot be parsed.
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
	//
	// Dir has to name a directory on the local file system; any other fails
	// the source. On Windows go-fs reads a file URI with a host, like
	// "file://server/share", as the network share `\\server\share`.
	//
	// Dir is trusted: the glob and the reads follow symlinks, so a symlink
	// below Dir publishes the file it points to, wherever that is, as does a
	// ".." in Pattern. Only those who may publish any file the process can
	// read should be able to write below Dir.
	Dir fs.File

	// Pattern is the glob matched below Dir, in the syntax of filepath.Match.
	// It must be relative when Dir is set. Like a shell glob, a wildcard does
	// not match a name beginning with ".": a pattern has to name the dot to
	// match hidden files, as ".well-known/*.txt" does. Directories, named
	// pipes, sockets and devices are skipped, as is a file that is gone by the
	// time its turn comes.
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
		// Checked here as well, so a cancelled ctx is not taken for an empty
		// site when no file matched.
		if err := ctx.Err(); err != nil {
			yield(nil, err)
			return
		}
		for _, file := range files {
			// Without content an HTML or text file is not read, so the file
			// reads alone would let a cancelled ctx go through every one of
			// them.
			if err := ctx.Err(); err != nil {
				yield(nil, err)
				return
			}
			if matchedHidden(pattern, file) {
				continue
			}
			// The glob matched a literal path, but go-fs expands a leading "~"
			// path element to the home directory. A leading "./" keeps a
			// directory named "~" literal, so the file read is the file
			// matched. Only such a path gets it: on Windows `\x` and `C:x` are
			// not absolute either, and "./" would change what they name.
			local := file
			if local == "~" || strings.HasPrefix(local, "~/") || strings.HasPrefix(local, "~"+string(filepath.Separator)) {
				local = "." + string(filepath.Separator) + local
			}
			content := fs.File(local)
			ext := content.ExtLower()
			// Only a file of a page type is looked at, so an unreadable image
			// a wide pattern like "*" matched does not fail the source.
			if ext != ".md" && ext != ".html" && ext != ".htm" && ext != ".txt" {
				continue
			}
			// The stat also dates the pages without front matter. Its error
			// fails the source rather than leaving the page out without a
			// word: without dates it would not be in the sitemap. It is the
			// os stat of the matched path, because go-fs reports every file
			// that is no directory and not regular as irregular, including
			// the OneDrive and compressed files of Windows, which are pages.
			info, err := os.Stat(file)
			if err != nil {
				// A file that is gone since the glob listed it, deleted or
				// replaced while a deploy runs, is no page. A dangling
				// symlink is still there and an error.
				_, lstatErr := os.Lstat(file)
				if errors.Is(lstatErr, os.ErrNotExist) {
					continue
				}
				yield(nil, err)
				return
			}
			// A pattern like "*.md" does not match a directory anyway, but a
			// wider one like "*" does, and a named pipe or a link to a device
			// would block or never end when read.
			if info.Mode()&(os.ModeDir|os.ModeNamedPipe|os.ModeSocket|os.ModeDevice|os.ModeCharDevice) != 0 {
				continue
			}
			switch ext {
			case ".md":
				data, err := content.ReadAll(ctx)
				if err != nil {
					yield(nil, err)
					return
				}
				var matter struct {
					Title       string    `yaml:"title"       toml:"title"       json:"title"`
					Description string    `yaml:"description" toml:"description" json:"description"`
					Lang        string    `yaml:"lang"        toml:"lang"        json:"lang"`
					Date        time.Time `yaml:"date"        toml:"date"        json:"date"`
					PublishDate time.Time `yaml:"publishDate" toml:"publishDate" json:"publishDate"`
					Draft       bool      `yaml:"draft"       toml:"draft"       json:"draft"`
					Tags        []string  `yaml:"tags"        toml:"tags"        json:"tags"`
				}
				// The front matter is metadata, parsed into the page fields
				// below. The content is the Markdown after it, which is what
				// a ContentRenderer renders and what a Markdown alternate
				// publishes (see [Page.MarkdownSource]).
				body, err := frontmatter.Parse(bytes.NewReader(data), &matter)
				if err != nil {
					yield(nil, errs.Errorf("%s: front matter: %w", file, err))
					return
				}
				lang, err := frontMatterLang(matter.Lang)
				if err != nil {
					yield(nil, fmt.Errorf("%s: front matter lang: %w", file, err))
					return
				}
				page := &Page{
					Path:        s.pagePath(file),
					ContentType: mx.ContentTypeMarkdown,
					Type:        s.PageType,
					Author:      s.DefaultAuthor,
					Title:       matter.Title,
					Description: matter.Description,
					Lang:        lang,
					Created:     matter.Date,
					LastUpdated: matter.Date,
					Published:   matter.PublishDate,
					Tags:        matter.Tags,
				}
				if matter.Draft {
					page.Published = time.Time{}
				}
				if withContent {
					page.Content = body
				}
				if !yield(page, nil) {
					return
				}

			case ".html", ".htm":
				// A file without frontmatter carries no dates, so its
				// modification time is the only publication signal there is.
				// Without it the page would never be [Page.Indexable] and
				// would silently stay out of every sitemap.
				modified := info.ModTime()
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
					page.Content, err = content.ReadAll(ctx)
					if err != nil {
						yield(nil, err)
						return
					}
				}
				if !yield(page, nil) {
					return
				}

			case ".txt":
				modified := info.ModTime()
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
					page.Content, err = content.ReadAll(ctx)
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

// frontMatterLang returns the lang of a page's front matter as a BCP 47
// language tag. Front matter written for Jekyll or Hugo often holds a locale
// like "en_US", which is turned into the tag "en-US" it stands for. Anything
// else that is no language tag is an error, reported with the file it is in
// rather than by every surface the page's language ends up in.
func frontMatterLang(lang string) (string, error) {
	lang = strings.ReplaceAll(strings.TrimSpace(lang), "_", "-")
	if lang == "" {
		return "", nil
	}
	if err := validateLang(lang); err != nil {
		return "", err
	}
	return lang, nil
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

// dirPath returns the local path of Dir, or "" if no Dir is set or if Dir
// names no directory on the local file system, which globPattern reports. The
// unset case is decided on the value rather than on its path, which must not
// depend on what go-fs makes of an empty File: a path of "." would be the
// working directory, a real directory to read from and not the same thing as
// "no directory given".
func (s *GlobPageSource) dirPath() string {
	if s.Dir == "" {
		return ""
	}
	return s.Dir.LocalPath()
}

// matchedHidden reports if a wildcard of pattern matched a name in the path
// of file that begins with ".". Like a shell glob, the source matches such a
// hidden file or directory only where pattern names the dot itself:
// ".well-known/*.txt" matches the files below ".well-known", but "*" does not
// match ".secret.txt". The matches of filepath.Glob are clean and have one
// name for each name of the cleaned pattern, so they compare name by name,
// without the volume name of Windows, which "C:.secret.txt" glues to the
// first name.
func matchedHidden(pattern, file string) bool {
	sep := string(filepath.Separator)
	pattern = filepath.Clean(pattern)
	patternNames := strings.Split(pattern[len(filepath.VolumeName(pattern)):], sep)
	fileNames := strings.Split(file[len(filepath.VolumeName(file)):], sep)
	if len(patternNames) != len(fileNames) {
		// A pattern without a wildcard matches only itself, as written.
		return false
	}
	for i, name := range fileNames {
		if strings.HasPrefix(name, ".") && !namesDot(patternNames[i]) {
			return true
		}
	}
	return false
}

// namesDot reports if a name of a glob pattern begins with a dot, written
// as is or, where a backslash is no separator, escaped as `\.`.
func namesDot(patternName string) bool {
	return strings.HasPrefix(patternName, ".") || (filepath.Separator != '\\' && strings.HasPrefix(patternName, `\.`))
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
		// The files are matched on the local disk, so a Dir of another file
		// system would match nothing, an empty source instead of an error.
		if s.Dir != "" {
			// Redacted, because the error goes to logs.
			return "", errs.Errorf("GlobPageSource.Dir %q names no directory on the local file system", fsimpl.RedactURI(string(s.Dir)))
		}
		return s.Pattern, nil
	}
	if filepath.IsAbs(s.Pattern) {
		return "", errs.Errorf("GlobPageSource.Pattern %q is absolute, it must be relative to Dir %q", s.Pattern, dir)
	}
	// The glob ignores a directory it cannot read, so a missing Dir would be
	// an empty source as well.
	info, err := os.Stat(dir)
	if err != nil {
		return "", errs.Errorf("GlobPageSource.Dir: %w", err)
	}
	if !info.IsDir() {
		return "", errs.Errorf("GlobPageSource.Dir %q is no directory", dir)
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
