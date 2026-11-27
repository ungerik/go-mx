package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/adrg/frontmatter"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// Page is a Markdown file of the repository, rendered as one page of the site.
type Page struct {
	// Source is the path of the Markdown file relative to the repository
	// root, with slashes: "docs/html/tutorial.md", "highlight/README.md".
	Source string

	// Path is the URL path of the page within the site, see pagePath.
	Path string

	// Title is the title from the front matter, else the text of the
	// first level-one heading, else the file name.
	Title string

	// Description is the description from the front matter, else the text
	// of the first paragraph, shortened.
	Description string

	// Section is the title of the navigation section listing the page, see
	// pageSection.
	Section string

	// NavTitle is the page's entry in the navigation.
	NavTitle string

	// order sorts the pages within their section.
	order string

	// Body is the Markdown of the page without its front matter.
	Body []byte

	// hasH1 reports whether Body has a level-one heading of its own; a page
	// without one gets its Title as the heading.
	hasH1 bool
}

// File returns the path of the page's HTML file relative to the site root:
// "html/tutorial.html" for the path "/html/tutorial", "html/index.html" for
// "/html/".
func (p *Page) File() string {
	return fileOfPath(p.Path)
}

// fileOfPath returns the HTML file of the site path p, see [Page.File].
func fileOfPath(p string) string {
	p = strings.TrimPrefix(p, "/")
	if p == "" || strings.HasSuffix(p, "/") {
		return p + "index.html"
	}
	return p + ".html"
}

// excludedFiles are the Markdown files that are internal planning or agent
// instructions, not documentation, matched by base name anywhere in the tree.
var excludedFiles = []string{"TODOS.md", "V1-AUDIT.md", "AGENTS.md", "CLAUDE.md"}

// excludedDirs are directory names that hold no documentation: test fixtures
// and dependency trees. Directories beginning with "." are skipped too.
var excludedDirs = []string{"testdata", "node_modules", "vendor"}

// discoverPages finds the documentation in the repository at root: every
// Markdown file below it except the excluded ones and those in skipped
// directories, such as the output directory when it lies inside the
// repository. The pages are returned sorted by navigation order (see
// sortPages).
func discoverPages(root string, skip ...string) ([]*Page, error) {
	var pages []*Page
	err := filepath.WalkDir(root, func(file string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			name := entry.Name()
			if file != root && (strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") || slices.Contains(excludedDirs, name)) {
				return filepath.SkipDir
			}
			for _, s := range skip {
				if sameFile(file, s) {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if !strings.EqualFold(filepath.Ext(file), ".md") || slices.Contains(excludedFiles, entry.Name()) {
			return nil
		}
		rel, err := filepath.Rel(root, file)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		page, err := newPage(filepath.ToSlash(rel), data)
		if err != nil {
			return fmt.Errorf("%s: %w", rel, err)
		}
		pages = append(pages, page)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sortPages(pages)
	return pages, nil
}

// sameFile reports whether a and b are the same path once made absolute.
func sameFile(a, b string) bool {
	absA, errA := filepath.Abs(a)
	absB, errB := filepath.Abs(b)
	return errA == nil && errB == nil && absA == absB
}

// newPage builds the page of the Markdown file source (relative to the
// repository root) holding data.
func newPage(source string, data []byte) (*Page, error) {
	var meta struct {
		Title       string `yaml:"title"`
		Description string `yaml:"description"`
	}
	body, err := frontmatter.Parse(bytes.NewReader(data), &meta)
	if err != nil {
		return nil, fmt.Errorf("front matter: %w", err)
	}
	page := &Page{
		Source:      source,
		Path:        pagePath(source),
		Title:       strings.TrimSpace(meta.Title),
		Description: strings.TrimSpace(meta.Description),
		Body:        body,
	}
	h1, paragraph := outline(body)
	page.hasH1 = h1 != ""
	if page.Title == "" {
		page.Title = h1
	}
	if page.Title == "" {
		page.Title = path.Base(source)
	}
	if page.Description == "" {
		page.Description = shorten(paragraph, 200)
	}
	page.Section, page.NavTitle, page.order = pageSection(source, page.Title)
	return page, nil
}

// pagePath maps the repository path of a Markdown file to the URL path of
// its page.
//
// The files under docs/ keep the URLs the Jekyll site served them under, so
// links into the site keep working: "docs/html/tutorial.md" is
// "/html/tutorial", an index.md the page of its directory ("/html/"). Every
// other file is published below "/repo/", at its repository path in lower
// case, a README.md as the page of its directory: "README.md" is "/repo/",
// "pdf/pdftable/README.md" is "/repo/pdf/pdftable/" and "highlight/DESIGN.md"
// is "/repo/highlight/design".
func pagePath(source string) string {
	p := strings.TrimSuffix(source, path.Ext(source))
	if rest, ok := strings.CutPrefix(p, "docs/"); ok {
		return indexPath("/"+rest, "index")
	}
	return indexPath("/repo/"+strings.ToLower(p), "readme")
}

// indexPath turns a path whose last name is index into the path of its
// directory, with a trailing slash.
func indexPath(p, index string) string {
	if path.Base(p) == index {
		return strings.TrimSuffix(p, index)
	}
	return p
}

// overviewSection and repoSection are the titles of the navigation sections
// that are not named after a directory of docs/.
const (
	overviewSection = "Overview"
	repoSection     = "Repository"
)

// diataxisOrder orders the pages of a docs/ directory the way the Diátaxis
// framework the documentation follows reads them, with their navigation
// titles. Other pages follow, by title.
var diataxisOrder = []struct{ name, nav string }{
	{"index", "Overview"},
	{"tutorial", "Tutorial"},
	{"how-to", "How-to guides"},
	{"explanation", "Explanation"},
	{"reference", "Reference"},
}

// pageSection returns the navigation section, the navigation title and the
// sort key of the page of source titled title.
//
// The top level of docs/ is the Overview section, led by the home page; each
// directory below docs/ is a section of its own, titled by its path
// ("shadcn/magicui"), its pages in Diátaxis order. Everything else is the
// Repository section: the root README first, then the package READMEs and
// the other files of a directory by path, then the remaining files of the
// repository root such as the CHANGELOG.
func pageSection(source, title string) (section, nav, order string) {
	dir, file := path.Split(strings.TrimSuffix(source, path.Ext(source)))
	dir = strings.TrimSuffix(dir, "/")
	if rest, ok := strings.CutPrefix(dir+"/", "docs/"); ok {
		rest = strings.TrimSuffix(rest, "/")
		if rest == "" {
			if file == "index" {
				return overviewSection, "Home", "0"
			}
			return overviewSection, title, "1" + title
		}
		for i, d := range diataxisOrder {
			if file == d.name {
				return rest, d.nav, fmt.Sprintf("%d", i)
			}
		}
		return rest, title, "9" + title
	}
	isReadme := strings.EqualFold(file, "readme")
	switch {
	case dir == "" && isReadme:
		return repoSection, "README", "0"
	case dir == "":
		return repoSection, file + ".md", "2" + strings.ToLower(file)
	case isReadme:
		return repoSection, dir + "/", "1" + dir + "/"
	default:
		return repoSection, dir + "/" + file + ".md", "1" + dir + "/~" + strings.ToLower(file)
	}
}

// sortPages sorts pages for the navigation: the Overview section first, then
// the sections of docs/ by path, then the Repository section, each in the
// order pageSection gives.
func sortPages(pages []*Page) {
	rank := func(p *Page) string {
		switch p.Section {
		case overviewSection:
			return "0"
		case repoSection:
			return "2"
		}
		return "1" + p.Section
	}
	slices.SortStableFunc(pages, func(a, b *Page) int {
		if c := strings.Compare(rank(a), rank(b)); c != 0 {
			return c
		}
		return strings.Compare(a.order, b.order)
	})
}

// outline returns the text of the first level-one heading and of the first
// paragraph of the Markdown body, either empty if there is none.
func outline(body []byte) (h1, paragraph string) {
	doc := newMarkdown().Parser().Parse(text.NewReader(body))
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch n := n.(type) {
		case *ast.Heading:
			if n.Level == 1 && h1 == "" {
				h1 = plainText(n, body)
			}
			return ast.WalkSkipChildren, nil
		case *ast.Paragraph:
			if paragraph == "" {
				paragraph = plainText(n, body)
			}
			return ast.WalkSkipChildren, nil
		}
		return ast.WalkContinue, nil
	})
	return h1, paragraph
}

// plainText returns the text of the inline content of n, without markup and
// images, with line breaks as spaces.
func plainText(n ast.Node, source []byte) string {
	var b strings.Builder
	_ = ast.Walk(n, func(c ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch c := c.(type) {
		case *ast.Text:
			b.Write(c.Segment.Value(source))
			if c.SoftLineBreak() || c.HardLineBreak() {
				b.WriteByte(' ')
			}
		case *ast.String:
			b.Write(c.Value)
		case *ast.CodeSpan:
			for child := c.FirstChild(); child != nil; child = child.NextSibling() {
				if t, ok := child.(*ast.Text); ok {
					b.Write(t.Segment.Value(source))
				}
			}
			return ast.WalkSkipChildren, nil
		case *ast.RawHTML, *ast.Image:
			// An image's alt text is no prose: a paragraph of badges
			// has no text, so the description comes from the next one.
			return ast.WalkSkipChildren, nil
		}
		return ast.WalkContinue, nil
	})
	return strings.Join(strings.Fields(b.String()), " ")
}

// shorten cuts s to at most limit bytes at a word boundary, marking the cut
// with an ellipsis.
func shorten(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	cut := strings.LastIndexByte(s[:limit], ' ')
	if cut <= 0 {
		cut = limit
		for cut > 0 && !utf8.RuneStart(s[cut]) {
			cut--
		}
	}
	return strings.TrimRight(s[:cut], " ,;:.—-") + "…"
}
