package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/cmd/shadcn-gallery/examples"
	"github.com/ungerik/go-mx/html"
)

// writeStatic renders the whole gallery to dir as static HTML files — the same
// pages the server serves: the index at <dir>/index.html and each component at
// <dir>/components/<slug>/index.html. The pages link to each other with
// relative URLs naming those files, so dir works served from any URL path and
// opened from disk.
func writeStatic(reg *Registry, dir string) error {
	n := 0
	write := func(relPath string, doc *html.Document) error {
		n++
		return writePage(filepath.Join(dir, relPath), doc)
	}

	if err := write("index.html", page(reg, "", "Components", indexContent(reg))); err != nil {
		return err
	}
	for i := range reg.Docs {
		d := &reg.Docs[i]
		if err := write(filepath.Join("components", d.Slug, "index.html"),
			page(reg, d.Slug, d.Title, componentContent(d))); err != nil {
			return err
		}
	}

	// The fragments are all generated, so a fragment an example no longer
	// serves must not linger from an earlier export.
	if err := os.RemoveAll(filepath.Join(dir, "fragments")); err != nil {
		return err
	}
	for _, f := range examples.Fragments() {
		if err := writeFragment(filepath.Join(dir, "fragments", filepath.FromSlash(f.Path)), f.Func()); err != nil {
			return err
		}
	}

	fmt.Printf("wrote %d pages and %d fragments to %s\n", n, len(examples.Fragments()), dir)
	return nil
}

// writeFragment renders comp to path, as the server's /fragments route
// does, creating parent directories as needed.
func writeFragment(path string, comp mx.Component) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return comp.Render(context.Background(), mx.NewCheckedWriter(f))
}

// writePage renders doc to path, creating parent directories as needed. It uses
// the same indenting writer as the server's [html.Document.HandleHTTP], so the
// static output matches the served pages byte for byte.
func writePage(path string, doc *html.Document) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := mx.NewCheckedWriter(f).WithIndent("", "  ")
	return doc.Render(context.Background(), w)
}
