package main

import (
	"testing"
)

// testResolver returns a resolver over a fixture repository and site
// directory with a few pages, a Go file and a generated gallery.
func testResolver(t *testing.T) (*resolver, map[string]*Page) {
	t.Helper()
	root, out := t.TempDir(), t.TempDir()
	writeFiles(t, root, map[string]string{
		"html/element.go":         "package html\n",
		"cmd/example-sse/main.go": "package main\n",
		"TODOS.md":                "# internal\n",
	})
	writeFiles(t, out, map[string]string{
		"shadcn/gallery/index.html":    "<html></html>",
		"examples/table-reference.pdf": "%PDF",
	})
	pages := map[string]*Page{}
	var list []*Page
	for _, source := range []string{
		"docs/index.md", "docs/html/index.md", "docs/html/tutorial.md", "docs/shadcn/index.md",
		"README.md", "html/README.md", "highlight/DESIGN.md",
	} {
		p := &Page{Source: source, Path: pagePath(source)}
		pages[source] = p
		list = append(list, p)
	}
	return newResolver(root, out, "https://example.org/go-mx/", list), pages
}

func TestResolve(t *testing.T) {
	r, pages := testResolver(t)
	for _, c := range []struct {
		from, dest, want string
	}{
		// Markdown files, as GitHub links them.
		{"README.md", "html/README.md", "html/index.html"},
		{"README.md", "html/README.md#elements", "html/index.html#elements"},
		{"README.md", "highlight/DESIGN.md", "highlight/design.html"},
		{"html/README.md", "../README.md", "../index.html"},
		{"README.md", "docs/html/tutorial.md", "../html/tutorial.html"},
		// The Jekyll forms the docs use: .html, a directory, no extension.
		{"docs/index.md", "html/tutorial.html", "html/tutorial.html"},
		{"docs/index.md", "html/", "html/index.html"},
		{"docs/index.md", "html/tutorial", "html/tutorial.html"},
		{"docs/html/tutorial.md", "../shadcn/", "../shadcn/index.html"},
		{"docs/html/tutorial.md", "index.html#reference", "index.html#reference"},
		// Generated files below docs/, also linked from outside docs/.
		{"docs/shadcn/index.md", "gallery/", "gallery/index.html"},
		{"html/README.md", "../docs/shadcn/gallery/", "../../shadcn/gallery/index.html"},
		{"docs/index.md", "examples/table-reference.pdf", "examples/table-reference.pdf"},
		// The repository's own GitHub URLs of a published file.
		{"docs/index.md", "https://github.com/ungerik/go-mx/blob/main/html/README.md#x", "repo/html/index.html#x"},
		{"docs/index.md", "https://github.com/ungerik/go-mx/tree/main/html", "repo/html/index.html"},
		// The published site's own URLs, as the READMEs link the docs.
		{"html/README.md", "https://example.org/go-mx/html/tutorial.html#step", "../../html/tutorial.html#step"},
		{"html/README.md", "https://example.org/go-mx/html/tutorial", "../../html/tutorial.html"},
		{"html/README.md", "https://example.org/go-mx/shadcn/gallery/", "../../shadcn/gallery/index.html"},
		{"html/README.md", "https://example.org/go-mx/", "../../index.html"},
		{"html/README.md", "https://example.org/other/", "https://example.org/other/"},
		// Other repository files go to GitHub.
		{"html/README.md", "element.go", repoURL + "blob/main/html/element.go"},
		{"html/README.md", "element.go#L10", repoURL + "blob/main/html/element.go#L10"},
		{"README.md", "cmd/example-sse", repoURL + "tree/main/cmd/example-sse"},
		{"README.md", "TODOS.md", repoURL + "blob/main/TODOS.md"},
		{"README.md", "/html/element.go", repoURL + "blob/main/html/element.go"},
		// Kept as written.
		{"README.md", "https://pkg.go.dev/github.com/ungerik/go-mx", "https://pkg.go.dev/github.com/ungerik/go-mx"},
		{"README.md", "https://github.com/ungerik/go-mx/blob/main/html/element.go", "https://github.com/ungerik/go-mx/blob/main/html/element.go"},
		{"README.md", "#install", "#install"},
		{"README.md", "mailto:x@example.com", "mailto:x@example.com"},
	} {
		from := pages[c.from]
		got, err := r.resolve(from, c.dest)
		if err != nil {
			t.Errorf("resolve(%s, %q): %v", c.from, c.dest, err)
			continue
		}
		// The links are relative to the page's file: README.md is
		// repo/index.html, html/README.md repo/html/index.html.
		if link := got.relativeTo(from.File()); link != c.want {
			t.Errorf("resolve(%s, %q) links %q, want %q", c.from, c.dest, link, c.want)
		}
	}
}

func TestResolveErrors(t *testing.T) {
	r, pages := testResolver(t)
	for _, dest := range []string{
		"missing.md",
		"../../outside.md",
		"docs/shadcn/missing-gallery/",
		"https://example.org/go-mx/renamed.html",
		"%zz",
	} {
		if _, err := r.resolve(pages["README.md"], dest); err == nil {
			t.Errorf("resolve(%q) did not fail", dest)
		}
	}
}

func TestTargetAbsolute(t *testing.T) {
	const base = "https://example.org/site/"
	for _, c := range []struct {
		t    target
		want string
	}{
		{target{File: "index.html"}, base},
		{target{File: "html/index.html", Fragment: "x"}, base + "html/#x"},
		{target{File: "html/tutorial.html"}, base + "html/tutorial"},
		{target{File: "examples/a.pdf"}, base + "examples/a.pdf"},
		{target{External: "https://x.org/"}, "https://x.org/"},
	} {
		if got := c.t.absolute(base); got != c.want {
			t.Errorf("%+v.absolute = %q, want %q", c.t, got, c.want)
		}
	}
}

func TestRelativeURL(t *testing.T) {
	for _, c := range []struct{ from, to, want string }{
		{"index.html", "html/tutorial.html", "html/tutorial.html"},
		{"html/tutorial.html", "index.html", "../index.html"},
		{"html/tutorial.html", "html/how-to.html", "how-to.html"},
		{"repo/pdf/pdftable/index.html", "repo/pdf/index.html", "../index.html"},
		{"html/index.html", "hx/index.html", "../hx/index.html"},
		{"a/b.html", "a b/c.html", "../a%20b/c.html"},
	} {
		if got := relativeURL(c.from, c.to); got != c.want {
			t.Errorf("relativeURL(%q, %q) = %q, want %q", c.from, c.to, got, c.want)
		}
	}
}
