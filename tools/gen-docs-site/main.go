// Command gen-docs-site builds the go-mx documentation site that GitHub
// Pages serves at https://ungerik.github.io/go-mx/, rendered by go-mx itself.
//
//	go -C tools run ./gen-docs-site -out ../_site
//
// It renders every Markdown file of the repository as a page — the
// documentation under docs/, the READMEs, the CHANGELOG, the design notes —
// except the internal planning files (TODOS.md, V1-AUDIT.md) and the agent
// instructions (AGENTS.md, CLAUDE.md). The Markdown is converted by goldmark
// with the GitHub Flavored Markdown extensions, and the pages are built with
// the go-mx html and shadcn packages in the theme the component gallery uses,
// Go code highlighted server-side by the highlight package.
//
// Links between the Markdown files become relative links between the pages,
// and links to any other file of the repository go to GitHub, so the site
// works below any URL path and opened from disk; a link that resolves to
// nothing fails the build. The files under docs/ keep the URLs the earlier
// Jekyll site served (/html/tutorial, /shadcn/), the other Markdown files are
// published below /repo/.
//
// Next to the pages it writes the component gallery (go run
// ./cmd/shadcn-gallery), static snapshots of the example servers (their
// -snapshot mode), the reference PDFs of the pdf golden tests, and, through
// the web package, sitemap.xml, llms.txt and a Markdown alternate of every
// page. Finally it checks that every internal link and fragment of the
// written site resolves.
//
// The output directory is emptied first, which it only does for a directory
// holding the marker file of an earlier build (or an empty one), so a
// mistyped -out cannot delete anything else.
package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"path/filepath"
)

func main() {
	out := flag.String("out", "", "directory to write the site to (required)")
	root := flag.String("root", "", "repository root; found from the working directory when empty")
	baseURL := flag.String("base-url", "https://ungerik.github.io/go-mx/", "URL the site is published at, ending in a slash")
	gallery := flag.Bool("gallery", true, "write the component gallery")
	examples := flag.Bool("examples", true, "write the snapshots of the example servers")
	check := flag.Bool("check", true, "check every internal link of the written site")
	flag.Parse()

	if *out == "" {
		flag.Usage()
		os.Exit(2)
	}
	if *root == "" {
		found, err := findRoot()
		if err != nil {
			log.Fatal(err)
		}
		*root = found
	}
	err := build(context.Background(), config{
		Root:     *root,
		Out:      *out,
		BaseURL:  *baseURL,
		Gallery:  *gallery,
		Examples: *examples,
		Check:    *check,
		Log:      os.Stdout,
	})
	if err != nil {
		log.Fatal(err)
	}
}

// findRoot returns the repository root: the nearest directory at or above
// the working directory holding the go.work file of the repository.
func findRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("no go.work found above the working directory; pass -root")
		}
		dir = parent
	}
}
