package main

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// repoURL is the GitHub URL of the repository; a link to a file the site does
// not publish goes to its blob or tree URL below it.
const repoURL = "https://github.com/ungerik/go-mx/"

// target is what a link in a page's Markdown points to, resolved: a file of
// the site or a URL outside it.
type target struct {
	// File is the path of the site file relative to the site root, such as
	// "html/tutorial.html" or "shadcn/gallery/index.html". Empty for an
	// external target.
	File string

	// External is the URL of a target outside the site, kept as written or
	// pointing to the repository on GitHub.
	External string

	// Fragment is the fragment of the link, escaped, without the "#", kept
	// for a site file.
	Fragment string
}

// relativeTo returns the URL of t as a link in the site file from: a
// relative URL for a site file, so the site works below any URL path and
// opened from disk, else the external URL.
func (t target) relativeTo(from string) string {
	if t.File == "" {
		return t.External
	}
	u := relativeURL(from, t.File)
	if t.Fragment != "" {
		u += "#" + t.Fragment
	}
	return u
}

// absolute returns the URL of t below baseURL, which ends in a slash, for the
// Markdown alternates, which are read without the site around them. A page
// file is linked by its page path, as its canonical URL is.
func (t target) absolute(baseURL string) string {
	if t.File == "" {
		return t.External
	}
	file := t.File
	if strings.HasSuffix(file, "index.html") {
		file = strings.TrimSuffix(file, "index.html")
	} else {
		file = strings.TrimSuffix(file, ".html")
	}
	u := baseURL + file
	if t.Fragment != "" {
		u += "#" + t.Fragment
	}
	return u
}

// relativeURL returns the URL of the site file to relative to the site file
// from, both relative to the site root with slashes: "../hx/index.html" from
// "html/tutorial.html" to "hx/index.html".
func relativeURL(from, to string) string {
	var fromDirs []string
	if dir := path.Dir(from); dir != "." {
		fromDirs = strings.Split(dir, "/")
	}
	toNames := strings.Split(to, "/")
	common := 0
	for common < len(fromDirs) && common < len(toNames)-1 && fromDirs[common] == toNames[common] {
		common++
	}
	rel := strings.Repeat("../", len(fromDirs)-common) + strings.Join(toNames[common:], "/")
	return (&url.URL{Path: rel}).EscapedPath()
}

// resolver resolves the links of the pages' Markdown, see [resolver.resolve].
type resolver struct {
	// root is the repository root, to find a linked file that is no page.
	root string
	// out is the site directory, to find a linked file that is generated,
	// like the gallery.
	out string
	// baseURL is the URL the site is published at, ending in a slash.
	baseURL string
	// pages are the pages by their Source, files the site files of the
	// pages.
	pages map[string]*Page
	files map[string]bool
}

func newResolver(root, out, baseURL string, pages []*Page) *resolver {
	r := &resolver{root: root, out: out, baseURL: baseURL, pages: make(map[string]*Page, len(pages)), files: make(map[string]bool, len(pages))}
	for _, p := range pages {
		r.pages[p.Source] = p
		r.files[p.File()] = true
	}
	return r
}

// resolve returns the target of the link dest in the Markdown of the page
// from. A link that is relative to the Markdown file is read as GitHub reads
// it, a path in the repository, and resolved, in this order, to:
//
//   - the page of a published Markdown file, also when it is linked as the
//     HTML file the Jekyll site made of it ("tutorial.html"), without an
//     extension, or as a directory with an index.md or README.md;
//   - a file the site generates (the gallery under docs/shadcn/gallery/, the
//     examples), found in the site directory at the path a file under docs/
//     is published at;
//   - any other file or directory of the repository, linked on GitHub.
//
// A link that resolves to none of them is an error, so a broken link fails
// the build. A link to the repository's own files on GitHub
// (https://github.com/ungerik/go-mx/blob/main/…) becomes a link to the page
// when the site publishes that file. A link to the published site itself
// (below baseURL, as the READMEs link the docs so the link works on GitHub)
// becomes a relative link to the file it names, and is an error if the site
// has no such file. Other absolute URLs, fragments and query-only links are
// kept as written.
func (r *resolver) resolve(from *Page, dest string) (target, error) {
	u, err := url.Parse(dest)
	if err != nil {
		return target{}, fmt.Errorf("link %q: %w", dest, err)
	}
	if u.Scheme != "" || u.Host != "" {
		if rest, ok := r.sitePath(u); ok {
			file, ok := r.siteFile(rest)
			if !ok {
				return target{}, fmt.Errorf("link %q in %s: the site has no file for %q", dest, from.Source, "/"+rest)
			}
			return target{File: file, Fragment: u.EscapedFragment()}, nil
		}
		if page := r.githubPage(u); page != nil {
			return target{File: page.File(), Fragment: u.EscapedFragment()}, nil
		}
		return target{External: dest}, nil
	}
	if u.Path == "" {
		// "#fragment" or "?query": relative to the page itself.
		return target{External: dest}, nil
	}
	repoPath := u.Path
	if !strings.HasPrefix(repoPath, "/") {
		repoPath = path.Join(path.Dir(from.Source), repoPath)
	} else {
		repoPath = strings.TrimPrefix(path.Clean(repoPath), "/")
	}
	if repoPath == ".." || strings.HasPrefix(repoPath, "../") {
		return target{}, fmt.Errorf("link %q leaves the repository", dest)
	}
	if repoPath == "." {
		repoPath = ""
	}
	if page := r.pageAt(repoPath); page != nil {
		return target{File: page.File(), Fragment: u.EscapedFragment()}, nil
	}
	if file, ok := r.generated(repoPath); ok {
		return target{File: file, Fragment: u.EscapedFragment()}, nil
	}
	info, err := os.Stat(filepath.Join(r.root, filepath.FromSlash(repoPath)))
	if err != nil {
		return target{}, fmt.Errorf("link %q in %s: %s is neither a page, a generated file nor a file of the repository", dest, from.Source, repoPath)
	}
	kind := "blob"
	if info.IsDir() {
		kind = "tree"
	}
	external := repoURL + kind + "/main/" + repoPath
	if u.Fragment != "" {
		external += "#" + u.EscapedFragment()
	}
	return target{External: external}, nil
}

// pageAt returns the page published for the repository path p, or nil: the
// page of the Markdown file p, of p with ".html" replaced by or ".md" added,
// or of the index.md or README.md of the directory p.
func (r *resolver) pageAt(p string) *Page {
	candidates := []string{p}
	if base, ok := strings.CutSuffix(p, ".html"); ok {
		candidates = append(candidates, base+".md")
	}
	dir := p
	if dir != "" {
		candidates = append(candidates, p+".md")
		dir += "/"
	}
	candidates = append(candidates, dir+"index.md", dir+"README.md")
	for _, c := range candidates {
		if page := r.pages[c]; page != nil {
			return page
		}
	}
	return nil
}

// generated returns the site file a repository path below docs/ is published
// as when the generator writes it rather than taking it from a Markdown file
// (one already in the site directory, or one registered with addFile): a
// file, or the index.html of a directory.
func (r *resolver) generated(repoPath string) (string, bool) {
	sitePath, ok := strings.CutPrefix(repoPath, "docs/")
	if !ok {
		return "", false
	}
	for _, file := range []string{sitePath, path.Join(sitePath, "index.html")} {
		if r.files[file] {
			return file, true
		}
		info, err := os.Stat(filepath.Join(r.out, filepath.FromSlash(file)))
		if err == nil && !info.IsDir() {
			return file, true
		}
	}
	return "", false
}

// addFile registers a site file the generator writes after the links are
// resolved, such as the examples page.
func (r *resolver) addFile(file string) { r.files[file] = true }

// sitePath returns the path of u below baseURL, without the leading slash,
// and whether u is below baseURL at all.
func (r *resolver) sitePath(u *url.URL) (string, bool) {
	if r.baseURL == "" {
		return "", false
	}
	base, err := url.Parse(r.baseURL)
	if err != nil || u.Scheme != base.Scheme || u.Host != base.Host {
		return "", false
	}
	return strings.CutPrefix(u.Path, base.Path)
}

// siteFile returns the file of the site at the URL path rest (relative to
// the site root) as GitHub Pages resolves it: a page or a generated file, a
// directory as its index.html, a path without extension as its .html file.
func (r *resolver) siteFile(rest string) (string, bool) {
	var candidates []string
	if rest == "" || strings.HasSuffix(rest, "/") {
		candidates = []string{rest + "index.html"}
	} else {
		candidates = []string{rest, rest + ".html", rest + "/index.html"}
	}
	for _, c := range candidates {
		if r.files[c] {
			return c, true
		}
		info, err := os.Stat(filepath.Join(r.out, filepath.FromSlash(c)))
		if err == nil && !info.IsDir() {
			return c, true
		}
	}
	return "", false
}

// githubPage returns the page published for a link to the repository on
// GitHub at u, such as https://github.com/ungerik/go-mx/blob/main/html/README.md,
// or nil if u links somewhere else or to a file the site does not publish.
func (r *resolver) githubPage(u *url.URL) *Page {
	if u.Scheme != "https" || u.Host != "github.com" {
		return nil
	}
	rest, ok := strings.CutPrefix(u.Path, "/ungerik/go-mx/")
	if !ok {
		return nil
	}
	for _, prefix := range []string{"blob/main/", "tree/main/"} {
		if p, ok := strings.CutPrefix(rest, prefix); ok {
			p = strings.TrimSuffix(p, "/")
			// Only the Markdown file itself or a directory with a README:
			// "blob/main/html" is not how GitHub links a page.
			if page := r.pages[p]; page != nil {
				return page
			}
			if prefix == "tree/main/" {
				return r.pages[path.Join(p, "README.md")]
			}
		}
	}
	return nil
}
