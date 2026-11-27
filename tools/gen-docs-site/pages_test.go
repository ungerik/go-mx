package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeFiles writes files (slash path → content) below dir.
func writeFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		file := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestPagePath(t *testing.T) {
	for source, want := range map[string]string{
		// docs/ keeps the URLs the Jekyll site served.
		"docs/index.md":                 "/",
		"docs/why-go-mx.md":             "/why-go-mx",
		"docs/html/index.md":            "/html/",
		"docs/html/tutorial.md":         "/html/tutorial",
		"docs/shadcn/magicui/how-to.md": "/shadcn/magicui/how-to",
		"docs/web/knowledge-layer.md":   "/web/knowledge-layer",
		// Everything else lives below /repo/, lower case, a README as its
		// directory.
		"README.md":                    "/repo/",
		"html/README.md":               "/repo/html/",
		"pdf/pdftable/README.md":       "/repo/pdf/pdftable/",
		"highlight/DESIGN.md":          "/repo/highlight/design",
		"CHANGELOG.md":                 "/repo/changelog",
		"cmd/shadcn-gallery/README.md": "/repo/cmd/shadcn-gallery/",
	} {
		if got := pagePath(source); got != want {
			t.Errorf("pagePath(%q) = %q, want %q", source, got, want)
		}
	}
}

func TestFileOfPath(t *testing.T) {
	for p, want := range map[string]string{
		"/":               "index.html",
		"/html/":          "html/index.html",
		"/html/tutorial":  "html/tutorial.html",
		"/repo/changelog": "repo/changelog.html",
	} {
		if got := fileOfPath(p); got != want {
			t.Errorf("fileOfPath(%q) = %q, want %q", p, got, want)
		}
	}
}

func TestNewPage(t *testing.T) {
	page, err := newPage("docs/html/tutorial.md", []byte("---\ntitle: Build a page\ndescription: Short.\n---\n\n# Heading\n\nFirst paragraph.\n"))
	if err != nil {
		t.Fatal(err)
	}
	if page.Title != "Build a page" || page.Description != "Short." || !page.hasH1 {
		t.Errorf("front matter not used: %+v", page)
	}
	if strings.Contains(string(page.Body), "title:") {
		t.Error("the front matter is left in the body")
	}
	if page.Section != "html" || page.NavTitle != "Tutorial" {
		t.Errorf("section %q, nav %q", page.Section, page.NavTitle)
	}

	// Without front matter: the H1 and the first paragraph.
	page, err = newPage("svg/README.md", []byte("# svg\n\n[![badge](https://x/b.svg)](https://x)\n\nSVG elements for `go-mx`,\nmapped like html.\n"))
	if err != nil {
		t.Fatal(err)
	}
	if page.Title != "svg" {
		t.Errorf("title %q", page.Title)
	}
	if page.Description != "SVG elements for go-mx, mapped like html." {
		t.Errorf("description %q", page.Description)
	}

	// No heading at all: the file name.
	page, err = newPage("NOTES.md", []byte("Just text.\n"))
	if err != nil {
		t.Fatal(err)
	}
	if page.Title != "NOTES.md" || page.hasH1 {
		t.Errorf("title %q, hasH1 %v", page.Title, page.hasH1)
	}

	if _, err := newPage("bad.md", []byte("---\ntitle: [unclosed\n---\n")); err == nil {
		t.Error("broken front matter did not fail")
	}
}

func TestPageSection(t *testing.T) {
	for _, c := range []struct{ source, section, nav string }{
		{"docs/index.md", overviewSection, "Home"},
		{"docs/why-go-mx.md", overviewSection, "T"},
		{"docs/hx/how-to.md", "hx", "How-to guides"},
		{"docs/shadcn/magicui/explanation.md", "shadcn/magicui", "Explanation"},
		{"docs/web/knowledge-layer.md", "web", "T"},
		{"README.md", repoSection, "README"},
		{"CHANGELOG.md", repoSection, "CHANGELOG.md"},
		{"pdf/pdftable/README.md", repoSection, "pdf/pdftable/"},
		{"highlight/DESIGN.md", repoSection, "highlight/DESIGN.md"},
	} {
		section, nav, _ := pageSection(c.source, "T")
		if section != c.section || nav != c.nav {
			t.Errorf("pageSection(%q) = %q, %q, want %q, %q", c.source, section, nav, c.section, c.nav)
		}
	}
}

func TestDiscoverPages(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		"README.md":             "# go-mx\n",
		"CHANGELOG.md":          "# Changelog\n",
		"TODOS.md":              "# internal\n",
		"AGENTS.md":             "# agents\n",
		"docs/index.md":         "# Home\n",
		"docs/why.md":           "# Why\n",
		"docs/html/how-to.md":   "# How\n",
		"docs/html/index.md":    "# html\n",
		"docs/html/tutorial.md": "# Tut\n",
		"html/README.md":        "# html pkg\n",
		"html/DESIGN.md":        "# design\n",
		"html/testdata/x.md":    "# fixture\n",
		".github/x.md":          "# hidden\n",
		"_site/index.md":        "# output\n",
		"out/stale.md":          "# skipped output dir\n",
	})
	pages, err := discoverPages(root, filepath.Join(root, "out"))
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, p := range pages {
		got = append(got, p.Source)
	}
	want := []string{
		"docs/index.md", "docs/why.md",
		"docs/html/index.md", "docs/html/tutorial.md", "docs/html/how-to.md",
		"README.md", "html/README.md", "html/DESIGN.md", "CHANGELOG.md",
	}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Errorf("pages\n got %v\nwant %v", got, want)
	}
}

func TestShorten(t *testing.T) {
	if got := shorten("short", 10); got != "short" {
		t.Errorf("got %q", got)
	}
	if got := shorten("one two three four", 10); got != "one two…" {
		t.Errorf("got %q", got)
	}
	// No space to cut at: the cut must not split a rune.
	if got := shorten("ääääää", 5); got != "ää…" {
		t.Errorf("got %q", got)
	}
}
