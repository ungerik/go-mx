package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fixtureRepo writes a small repository with docs, READMEs, an excluded
// planning file and the reference PDFs copyPDFs needs.
func fixtureRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		"docs/index.md":         "---\ntitle: go-mx\n---\n\n# go-mx\n\nThe home page. Read the [tutorial](html/tutorial.html) and the [README](https://github.com/ungerik/go-mx/blob/main/README.md).\n",
		"docs/html/tutorial.md": "---\ntitle: Tutorial\n---\n\n# Tutorial\n\nBack [home](../) or to the [package](../../html/README.md#install).\n\n```go\npackage main\n```\n",
		"README.md":             "# go-mx\n\n[![badge](https://x.org/b.svg)](https://x.org)\n\nRoot readme. See [html](html/README.md) and [the plan](TODOS.md).\n",
		"html/README.md":        "# html\n\nThe html package.\n\n## Install\n\nSee [the docs](../docs/index.md) and [the code](element.go).\n",
		"html/element.go":       "package html\n",
		"TODOS.md":              "# internal\n",
	}
	for _, pdf := range referencePDFs {
		files[pdf.Source] = "%PDF-1.4"
	}
	writeFiles(t, root, files)
	return root
}

func TestBuild(t *testing.T) {
	root := fixtureRepo(t)
	out := filepath.Join(t.TempDir(), "site")
	cfg := config{Root: root, Out: out, BaseURL: "https://example.org/go-mx/", Check: true, Log: io.Discard}
	if err := build(t.Context(), cfg); err != nil {
		t.Fatal(err)
	}

	read := func(name string) string {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(out, filepath.FromSlash(name)))
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}
	if _, err := os.Stat(filepath.Join(out, "repo", "todos.html")); err == nil {
		t.Error("the excluded TODOS.md was published")
	}

	home := read("index.html")
	for _, want := range []string{
		"<title>go-mx</title>",
		`href="html/tutorial.html"`,
		`href="repo/index.html"`,
		`<link rel="canonical" href="https://example.org/go-mx/"`,
		`type="text/markdown"`,
		`href="index.md"`,
		`<style type="text/tailwindcss">`,
		`aria-current="page"`,
	} {
		if !strings.Contains(home, want) {
			t.Errorf("index.html misses %q", want)
		}
	}

	tutorial := read("html/tutorial.html")
	for _, want := range []string{
		"<title>Tutorial · go-mx</title>",
		`href="../index.html"`,
		`href="../repo/html/index.html#install"`,
		`class="hl"`,
	} {
		if !strings.Contains(tutorial, want) {
			t.Errorf("html/tutorial.html misses %q", want)
		}
	}

	readme := read("repo/index.html")
	if !strings.Contains(readme, `href="`+repoURL+`blob/main/TODOS.md"`) {
		t.Error("the link to the unpublished TODOS.md does not go to GitHub")
	}

	// The Markdown alternate links absolutely, as it is read on its own.
	alternate := read("html/tutorial.md")
	for _, want := range []string{
		"# Tutorial",
		"Canonical URL: https://example.org/go-mx/html/tutorial",
		"[home](https://example.org/go-mx/)",
		"[package](https://example.org/go-mx/repo/html/#install)",
	} {
		if !strings.Contains(alternate, want) {
			t.Errorf("html/tutorial.md misses %q", want)
		}
	}

	sitemap := read("sitemap.xml")
	for _, want := range []string{"<loc>https://example.org/go-mx/</loc>", "<loc>https://example.org/go-mx/html/tutorial</loc>", "<loc>https://example.org/go-mx/examples/</loc>"} {
		if !strings.Contains(sitemap, want) {
			t.Errorf("sitemap.xml misses %q", want)
		}
	}
	llms := read("llms.txt")
	for _, want := range []string{"# go-mx", "## html", "(https://example.org/go-mx/html/tutorial.md)", "## Repository"} {
		if !strings.Contains(llms, want) {
			t.Errorf("llms.txt misses %q", want)
		}
	}
	if strings.Contains(read("repo/index.md"), "badge") && !strings.Contains(llms, "Root readme.") {
		t.Error("the README's description is its badge row")
	}
	if _, err := os.Stat(filepath.Join(out, "robots.txt")); err == nil {
		t.Error("robots.txt was written below the site path, where crawlers never read it")
	}
	if !strings.Contains(read("examples/index.html"), `href="table-reference.pdf"`) {
		t.Error("the examples page does not link the reference PDFs")
	}

	// A second build replaces the first.
	if err := build(t.Context(), cfg); err != nil {
		t.Fatalf("rebuild: %v", err)
	}
}

func TestBuildFailsOnBrokenLinks(t *testing.T) {
	root := fixtureRepo(t)
	writeFiles(t, root, map[string]string{"docs/broken.md": "# Broken\n\n[x](nowhere.md) [y](../gone/)\n"})
	err := build(t.Context(), config{Root: root, Out: t.TempDir(), BaseURL: "https://example.org/", Log: io.Discard})
	if err == nil || !strings.Contains(err.Error(), "nowhere.md") || !strings.Contains(err.Error(), "gone") {
		t.Fatalf("broken links not reported: %v", err)
	}
}

func TestPrepareOut(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{"precious.txt": "keep"})
	if err := prepareOut(dir); err == nil {
		t.Fatal("a foreign directory was emptied")
	}
	if _, err := os.Stat(filepath.Join(dir, "precious.txt")); err != nil {
		t.Fatal("the foreign file is gone")
	}

	site := filepath.Join(t.TempDir(), "new", "site")
	if err := prepareOut(site); err != nil {
		t.Fatal(err)
	}
	writeFiles(t, site, map[string]string{"stale.html": "old"})
	if err := prepareOut(site); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(site, "stale.html")); err == nil {
		t.Error("a file of the earlier build survived")
	}
}

// TestBuildRepository builds the site of this repository the way the Pages
// workflow does, gallery and example snapshots included, and checks its
// links. It runs go run for each of those commands, so -short skips it.
func TestBuildRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("builds the whole site with go run; skipped with -short")
	}
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	err = build(t.Context(), config{
		Root:     root,
		Out:      filepath.Join(t.TempDir(), "site"),
		BaseURL:  "https://ungerik.github.io/go-mx/",
		Gallery:  true,
		Examples: true,
		Check:    true,
		Log:      io.Discard,
	})
	if err != nil {
		t.Fatal(err)
	}
}
