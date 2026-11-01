package web

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	iofs "io/fs"
	"iter"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"slices"
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
	// Dir is no pattern: a directory named "content[prod]" is matched below
	// as it is named, only Pattern holds wildcards.
	//
	// An empty Dir keeps Pattern as the whole path to match, relative to the
	// working directory, and takes the base for page paths from the part of
	// Pattern before its first wildcard — the same result written differently.
	//
	// Dir has to name a directory on the local file system; any other fails
	// the source. On Windows go-fs reads a file URI with a host, like
	// "file://server/share", as the network share `\\server\share`.
	//
	// Dir is trusted unless Confine is set: the glob and the reads follow
	// symlinks, so a symlink below Dir publishes the file it points to,
	// wherever that is, as does a ".." in Pattern. Only those who may publish
	// any file the process can read should be able to write below Dir.
	Dir fs.File

	// Confine keeps the source inside Dir, for content that people who may
	// not read the server can write, like a contributed content repository.
	// The glob and the reads go through an [os.Root] opened on Dir: a symlink
	// is followed only if it is relative and stays below Dir, a file or
	// directory reached through any other symlink fails the source, and a
	// Pattern reaching above Dir with ".." is an error. Confine requires Dir. It is off by default because deploy tools
	// symlink content directories, which it would reject.
	Confine bool

	// Pattern is the glob matched below Dir, in the syntax of filepath.Match.
	// It must be relative when Dir is set. Like a shell glob, a wildcard does
	// not match a name beginning with ".": a pattern has to name the dot to
	// match hidden files, as ".well-known/*.txt" does. Directories, named
	// pipes, sockets and devices are skipped, as is a file that is gone by the
	// time its turn comes. A page file is opened also when no content is
	// read, so one that cannot be read fails the source then too. A directory the glob has to search but cannot read
	// fails the source, rather than leaving its pages out without a word; a
	// hidden directory is not searched unless the pattern names its dot.
	Pattern string

	// MaxFileSize is the size in bytes past which a page file fails the
	// source instead of being read into memory. 0 means
	// [DefaultMaxPageFileSize], a negative value no limit. A sitemap, llms.txt
	// or page list, which need no content, read only the head of a Markdown
	// file holding its front matter.
	MaxFileSize int64

	PageType      string
	DefaultAuthor string
}

// DefaultMaxPageFileSize is the default of [GlobPageSource.MaxFileSize]:
// the size in bytes past which a page file fails its source.
var DefaultMaxPageFileSize int64 = 10 << 20

// frontMatterHeadSize is how much of a Markdown file is read for its front
// matter when no content is wanted. A file whose front matter does not end
// within it is read whole.
const frontMatterHeadSize = 64 << 10

// Pages iterates over the files matching the glob Pattern and yields a Page for
// each supported file type, implementing the PageSource interface.
//
// The context is checked while the glob reads each directory and between
// the chunks of every file read, so a cancelled request stops at the next
// of them. A single call into the operating system that stalls, a stat, an
// open or a read on a hung network or FUSE mount, cannot be interrupted:
// the source returns with the context's error once that call does.
func (s *GlobPageSource) Pages(ctx context.Context, withContent bool) iter.Seq2[*Page, error] {
	return s.pages(ctx, withContent, "")
}

// pagesAt is [GlobPageSource.Pages] restricted to the pages at pagePath: a
// file with another path is neither opened nor read, so serving one page's
// Markdown alternate reads that page's file alone.
func (s *GlobPageSource) pagesAt(ctx context.Context, pagePath string, withContent bool) iter.Seq2[*Page, error] {
	return s.pages(ctx, withContent, pagePath)
}

// pages iterates the pages of the files matching the glob Pattern, only
// those at onlyPath unless it is "".
func (s *GlobPageSource) pages(ctx context.Context, withContent bool, onlyPath string) iter.Seq2[*Page, error] {
	return func(yield func(*Page, error) bool) {
		pattern, err := s.globPattern()
		if err != nil {
			yield(nil, err)
			return
		}
		var files pageFiles = localFiles{dir: s.dirPath()}
		if s.Confine {
			root, err := os.OpenRoot(s.dirPath())
			if err != nil {
				yield(nil, errs.Errorf("GlobPageSource.Dir: %w", err))
				return
			}
			defer root.Close()
			files = rootFiles{root: root, dir: s.dirPath()}
		}
		// Pattern, not the pattern joined to Dir, so the glob runs below Dir
		// and a glob character in the name of Dir stays a character.
		matches, err := files.glob(ctx, s.Pattern)
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
		for _, file := range matches {
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
			ext := strings.ToLower(filepath.Ext(file))
			// Only a file of a page type is looked at, so an unreadable image
			// a wide pattern like "*" matched does not fail the source.
			if ext != ".md" && ext != ".html" && ext != ".htm" && ext != ".txt" {
				continue
			}
			if onlyPath != "" && s.pagePath(file) != onlyPath {
				continue
			}
			// The file is opened once, also when no content is read, and
			// its stat is that of the open file, so what is checked is what
			// is read. Its error fails the source rather than leaving the
			// page out without a word: without dates it would not be in the
			// sitemap, and it fails whenever it is served.
			f, info, err := openPageFile(files, file)
			if err != nil {
				yield(nil, err)
				return
			}
			if f == nil {
				continue
			}
			page, err := s.readPage(ctx, f, info, file, ext, withContent)
			f.Close()
			if err != nil {
				yield(nil, err)
				return
			}
			if !yield(page, nil) {
				return
			}
		}
	}
}

// readPage returns the page of the open file f, which info describes, with
// its content if withContent is set.
func (s *GlobPageSource) readPage(ctx context.Context, f *os.File, info os.FileInfo, file, ext string, withContent bool) (*Page, error) {
	// Also a file that is not read now: it fails whenever it is served.
	limit := s.maxFileSize()
	if limit > 0 && info.Size() > limit {
		return nil, errs.Errorf("%s: %d bytes, more than the %d bytes a page file may have", file, info.Size(), limit)
	}
	switch ext {
	case ".md":
		head := int64(-1)
		if !withContent {
			head = frontMatterHeadSize
		}
		data, whole, err := readPageFile(ctx, f, file, limit, head)
		if err != nil {
			return nil, err
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
		if err == nil && !whole && len(body) == len(data) {
			// The head held no complete front matter: either the file
			// has none, or it ends past the head. The rest is read from
			// the same open file, not from whatever has the name now.
			if _, err := f.Seek(0, io.SeekStart); err != nil {
				return nil, err
			}
			if data, _, err = readPageFile(ctx, f, file, limit, -1); err != nil {
				return nil, err
			}
			body, err = frontmatter.Parse(bytes.NewReader(data), &matter)
		}
		if err != nil {
			return nil, errs.Errorf("%s: front matter: %w", file, err)
		}
		lang, err := frontMatterLang(matter.Lang)
		if err != nil {
			return nil, fmt.Errorf("%s: front matter lang: %w", file, err)
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
		return page, nil

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
			content, _, err := readPageFile(ctx, f, file, limit, -1)
			if err != nil {
				return nil, err
			}
			page.Content = content
		}
		return page, nil

	default: // ".txt"
		modified := info.ModTime()
		page := &Page{
			Path:        s.pagePath(file),
			ContentType: mx.ContentTypePlainText,
			Type:        s.PageType,
			Title:       filepath.Base(file), // Use filename as title
			Author:      s.DefaultAuthor,
			Created:     modified,
			LastUpdated: modified,
			Published:   modified,
		}
		if withContent {
			content, _, err := readPageFile(ctx, f, file, limit, -1)
			if err != nil {
				return nil, err
			}
			page.Content = content
		}
		return page, nil
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

// maxFileSize returns the size in bytes past which a page file fails the
// source, or 0 for no limit.
func (s *GlobPageSource) maxFileSize() int64 {
	switch {
	case s.MaxFileSize > 0:
		return s.MaxFileSize
	case s.MaxFileSize < 0:
		return 0
	}
	return max(DefaultMaxPageFileSize, 0)
}

// readPageFile reads the open file f, named file: whole, failing past limit
// bytes (none for 0) also when it grew since its stat, or for head >= 0 only
// its first head bytes, whole reporting whether that was all of it. The read
// stops with the context's error once ctx is done (see [contextReader]).
func readPageFile(ctx context.Context, f *os.File, file string, limit, head int64) (data []byte, whole bool, err error) {
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}
	var r io.Reader = contextReader{ctx: ctx, r: f}
	switch {
	case head >= 0:
		r = io.LimitReader(r, head+1)
	case limit > 0:
		r = io.LimitReader(r, limit+1)
	}
	data, err = io.ReadAll(r)
	if err != nil {
		return nil, false, err
	}
	if head >= 0 {
		if int64(len(data)) > head {
			return data[:head], false, nil
		}
		return data, true, nil
	}
	if limit > 0 && int64(len(data)) > limit {
		return nil, false, errs.Errorf("%s: more than the %d bytes a page file may have", file, limit)
	}
	return data, true, nil
}

// readChunkSize is the most a [contextReader] reads at once, so a large file
// is read in steps the context is checked between.
const readChunkSize = 64 << 10

// contextReader reads from r in chunks of at most [readChunkSize] bytes and
// returns the error of ctx instead of reading once ctx is done, so the read
// of a large page file ends soon after its request was cancelled. A read
// already blocked in the operating system, like one on a stalled network
// mount, returns only when the operating system lets it.
type contextReader struct {
	ctx context.Context
	r   io.Reader
}

func (c contextReader) Read(p []byte) (int, error) {
	if err := c.ctx.Err(); err != nil {
		return 0, err
	}
	if len(p) > readChunkSize {
		p = p[:readChunkSize]
	}
	return c.r.Read(p)
}

// openPageFile opens a matched file and stats the open file rather than the
// name, so the page is read from the file that was checked: a file replaced
// by a named pipe between a stat and an open by name would block the read
// without a timeout. The open itself does not wait for a writer of a named
// pipe (see [openPageFileFlags]).
//
// It returns a nil file and no error for a file that is no page: a
// directory, named pipe, socket or device, which a pattern like "*" matches
// and which would block or never end when read, and a file that is gone
// since the glob listed it, deleted or replaced while a deploy runs. A
// dangling symlink is still there and an error, as is one leading out of a
// confined Dir. The check is a list of what to skip rather than
// [os.FileMode.IsRegular], because the OneDrive and compressed files of
// Windows are irregular to Go and still pages.
func openPageFile(files pageFiles, file string) (*os.File, os.FileInfo, error) {
	f, err := files.open(file)
	if err != nil {
		if _, lstatErr := files.lstat(file); errors.Is(lstatErr, os.ErrNotExist) {
			return nil, nil, nil
		}
		// A socket cannot be opened at all.
		if info, statErr := files.stat(file); statErr == nil && noPageFile(info.Mode()) {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, nil, err
	}
	if noPageFile(info.Mode()) {
		f.Close()
		return nil, nil, nil
	}
	return f, info, nil
}

// noPageFile reports whether a file of mode is no page to read: a
// directory, named pipe, socket or device.
func noPageFile(mode os.FileMode) bool {
	return mode&(os.ModeDir|os.ModeNamedPipe|os.ModeSocket|os.ModeDevice|os.ModeCharDevice) != 0
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
// otherwise the fixed directory prefix of Pattern, the names before the
// first one holding a wildcard or, where a backslash is no separator, an
// escape like the `\.` of a hidden name.
func (s *GlobPageSource) baseDir() string {
	if dir := s.dirPath(); dir != "" {
		return dir
	}
	pattern := filepath.Clean(s.Pattern)
	base, rest := splitGlobBase(pattern)
	if rest == "" {
		// A pattern without a wildcard names one file.
		return filepath.Dir(pattern)
	}
	return base
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
			return "", errs.Errorf("GlobPageSource.Dir %q names no directory on the local file system", redactURI(string(s.Dir)))
		}
		if s.Confine {
			return "", errs.New("GlobPageSource.Confine requires a Dir to confine the source to")
		}
		return s.Pattern, nil
	}
	// On Windows go-fs reads "file://user:pass@server/share" as the UNC path
	// `\\user:pass@server\share`, which every error naming Dir or one of its
	// files would print. No UNC path takes credentials that way, so such a
	// Dir is refused before anything prints it.
	if u, err := url.Parse(string(s.Dir)); err == nil && u.Scheme != "" && u.User != nil {
		return "", errs.Errorf("GlobPageSource.Dir %q has user info, which a local directory cannot use", redactURI(string(s.Dir)))
	}
	if filepath.IsAbs(s.Pattern) {
		return "", errs.Errorf("GlobPageSource.Pattern %q is absolute, it must be relative to Dir %q", s.Pattern, dir)
	}
	// Opened rather than stat'ed, which needs no permission on Dir itself,
	// so a Dir that cannot be read fails here with its own error.
	f, err := os.Open(dir)
	if err != nil {
		return "", errs.Errorf("GlobPageSource.Dir: %w", err)
	}
	info, err := f.Stat()
	f.Close()
	if err != nil {
		return "", errs.Errorf("GlobPageSource.Dir: %w", err)
	}
	if !info.IsDir() {
		return "", errs.Errorf("GlobPageSource.Dir %q is no directory", dir)
	}
	pattern := filepath.Join(dir, s.Pattern)
	if s.Confine {
		if rel, err := filepath.Rel(dir, pattern); err != nil || !iofs.ValidPath(filepath.ToSlash(rel)) {
			return "", errs.Errorf("GlobPageSource.Pattern %q reaches outside the confined Dir %q", s.Pattern, dir)
		}
	}
	return pattern, nil
}

// redactURI returns uri for an error message, which goes to logs: without
// the user info, query and fragment, where a URI carries a password or an
// access token, or only its scheme if it does not parse. A string without a
// scheme is returned as it is.
func redactURI(uri string) string {
	u, err := url.Parse(uri)
	if err != nil {
		if scheme, _, ok := strings.Cut(uri, ":"); ok && !strings.ContainsAny(scheme, "/@?#") {
			return scheme + ":…"
		}
		return "…"
	}
	if u.Scheme == "" {
		return uri
	}
	u.User = nil
	u.RawQuery = ""
	u.ForceQuery = false
	u.Fragment = ""
	u.RawFragment = ""
	return u.String()
}

// pageFiles is where a [GlobPageSource] matches and reads its files: the
// local file system, or with Confine an [os.Root] on Dir. The glob takes
// the Pattern of the source, relative to Dir, and files are named by their
// local path, Dir and the matched name joined.
type pageFiles interface {
	glob(ctx context.Context, pattern string) ([]string, error)
	stat(file string) (os.FileInfo, error)
	lstat(file string) (os.FileInfo, error)
	// open opens file for reading without blocking on a named pipe.
	open(file string) (*os.File, error)
}

// localFiles reads from the local file system, following every symlink.
// The glob runs below dir, or for an empty dir matches the whole pattern.
type localFiles struct {
	dir string
}

func (l localFiles) glob(ctx context.Context, pattern string) ([]string, error) {
	// Only the pattern is split at its first wildcard, and dir joined to the
	// fixed part before it: a directory named "content[prod]" is that
	// directory, not a pattern matching "contentp". Joining also resolves a
	// leading ".." of the pattern, which an [iofs.FS] would refuse.
	base, rest := splitGlobBase(filepath.Clean(pattern))
	if l.dir != "" {
		base = filepath.Join(l.dir, base)
	}
	if rest == "" {
		// A pattern without a wildcard matches the file it names, also a
		// dangling symlink, which then fails the source.
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if _, err := os.Lstat(base); err != nil {
			if errors.Is(err, iofs.ErrNotExist) {
				return nil, nil
			}
			return nil, err
		}
		return []string{base}, nil
	}
	names, err := globFS(ctx, os.DirFS(base), rest)
	if err != nil {
		return nil, err
	}
	files := make([]string, len(names))
	for i, name := range names {
		files[i] = filepath.Join(base, filepath.FromSlash(name))
	}
	return files, nil
}

func (localFiles) stat(file string) (os.FileInfo, error) { return os.Stat(file) }

func (localFiles) lstat(file string) (os.FileInfo, error) { return os.Lstat(file) }

// open opens the matched path with os, as it is: a leading "~" names a
// directory "~", not the home directory, and "file://localhost/x" a
// directory "file:", not the file URI go-fs would read.
func (localFiles) open(file string) (*os.File, error) {
	return os.OpenFile(file, openPageFileFlags, 0)
}

// rootFiles reads through root, opened on the directory dir, so a symlink is
// followed only if it is relative and stays below dir: a file or directory
// reached through any other fails to stat, read or search.
type rootFiles struct {
	root *os.Root
	dir  string
}

// glob matches pattern below the root, which a glob character in the name
// of dir cannot change, because the root is opened on dir by name.
func (r rootFiles) glob(ctx context.Context, pattern string) ([]string, error) {
	names, err := globFS(ctx, r.root.FS(), filepath.ToSlash(filepath.Clean(pattern)))
	if err != nil {
		return nil, err
	}
	files := make([]string, len(names))
	for i, name := range names {
		files[i] = filepath.Join(r.dir, filepath.FromSlash(name))
	}
	return files, nil
}

// name returns the name of file below the root.
func (r rootFiles) name(file string) (string, error) {
	return filepath.Rel(r.dir, file)
}

func (r rootFiles) stat(file string) (os.FileInfo, error) {
	name, err := r.name(file)
	if err != nil {
		return nil, err
	}
	return r.root.Stat(name)
}

func (r rootFiles) lstat(file string) (os.FileInfo, error) {
	name, err := r.name(file)
	if err != nil {
		return nil, err
	}
	return r.root.Lstat(name)
}

func (r rootFiles) open(file string) (*os.File, error) {
	name, err := r.name(file)
	if err != nil {
		return nil, err
	}
	return r.root.OpenFile(name, openPageFileFlags, 0)
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

// splitGlobBase splits a clean local pattern into the directory it is
// matched below, the names before the first one holding a wildcard (see
// [hasGlobMeta]), and the rest in slash form, for an [iofs.FS] on that
// directory. rest is "" for a pattern without a wildcard.
func splitGlobBase(pattern string) (base, rest string) {
	volume := filepath.VolumeName(pattern)
	names := strings.Split(filepath.ToSlash(pattern[len(volume):]), "/")
	for i, name := range names {
		if !hasGlobMeta(name) {
			continue
		}
		base = strings.Join(names[:i], "/")
		if base == "" && i > 0 {
			base = "/" // the names of an absolute pattern start with ""
		}
		base = volume + filepath.FromSlash(base)
		if base == "" {
			base = "."
		}
		return base, strings.Join(names[i:], "/")
	}
	return pattern, ""
}

// hasGlobMeta reports whether a name of a slash-form pattern holds a
// wildcard or an escape: path.Match reads a backslash as one, and on
// Windows, where it is a separator, filepath.ToSlash has removed it.
func hasGlobMeta(name string) bool {
	return strings.ContainsAny(name, `*?[\`)
}

// globFS is [iofs.Glob], except that a directory it has to search but
// cannot read fails it instead of matching nothing, that a wildcard does
// not match a name beginning with "." (see [matchedHidden]), so a hidden
// directory is not searched either, and that it stops with the error of ctx
// once ctx is done, checked before every directory and every batch of
// [readDirBatchSize] entries it reads.
func globFS(ctx context.Context, fsys iofs.FS, pattern string) ([]string, error) {
	if _, err := path.Match(pattern, ""); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !hasGlobMeta(pattern) {
		if _, err := iofs.Lstat(fsys, pattern); err != nil {
			if errors.Is(err, iofs.ErrNotExist) {
				return nil, nil
			}
			return nil, err
		}
		return []string{pattern}, nil
	}
	dir, name := path.Split(pattern)
	dir = path.Clean(dir)
	dirs := []string{dir}
	if hasGlobMeta(dir) {
		var err error
		if dirs, err = globFS(ctx, fsys, dir); err != nil {
			return nil, err
		}
	}
	var matches []string
	for _, d := range dirs {
		entries, err := readDirContext(ctx, fsys, d)
		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return nil, ctxErr
			}
			// A name the wildcard matched that is no directory, or one gone
			// since it was listed, has nothing to search.
			if info, statErr := iofs.Stat(fsys, d); errors.Is(err, iofs.ErrNotExist) || statErr == nil && !info.IsDir() {
				continue
			}
			return nil, err
		}
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), ".") && !namesDot(name) {
				continue
			}
			matched, err := path.Match(name, entry.Name())
			if err != nil {
				return nil, err
			}
			if matched {
				matches = append(matches, path.Join(d, entry.Name()))
			}
		}
	}
	return matches, nil
}

// readDirBatchSize is how many entries of a directory [readDirContext] reads
// at once.
const readDirBatchSize = 1024

// readDirContext is [iofs.ReadDir], sorted by name like it, but reading the
// directory in batches of [readDirBatchSize] entries and stopping with the
// error of ctx between them, so a huge directory does not hold a cancelled
// request until all of it is listed.
func readDirContext(ctx context.Context, fsys iofs.FS, dir string) ([]iofs.DirEntry, error) {
	f, err := fsys.Open(dir)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	dirFile, ok := f.(iofs.ReadDirFile)
	if !ok {
		return nil, &iofs.PathError{Op: "readdir", Path: dir, Err: errors.New("not implemented")}
	}
	var entries []iofs.DirEntry
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		batch, err := dirFile.ReadDir(readDirBatchSize)
		entries = append(entries, batch...)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
	}
	slices.SortFunc(entries, func(a, b iofs.DirEntry) int { return strings.Compare(a.Name(), b.Name()) })
	return entries, nil
}
