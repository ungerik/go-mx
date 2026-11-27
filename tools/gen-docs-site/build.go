package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/web"
)

// config is what a build of the site needs.
type config struct {
	// Root is the repository root.
	Root string
	// Out is the directory the site is written to.
	Out string
	// BaseURL is the URL the site is published at, ending in a slash.
	BaseURL string
	// Gallery and Examples run the gallery and snapshot commands, which take
	// a while to compile; tests of the documentation pages leave them off.
	Gallery, Examples bool
	// Check checks every internal link and fragment of the written site.
	Check bool
	// Log receives the progress and the output of the commands run.
	Log io.Writer
}

// siteDescription is the description of the site as a whole.
const siteDescription = "Type-safe, component-based HTML generation for Go — with a shadcn/ui port."

// outMarker is the file that marks a directory as written by this command, so
// that a second build may empty it; any other directory that is not empty is
// refused rather than deleted.
const outMarker = ".gen-docs-site"

// build writes the whole site into cfg.Out.
func build(ctx context.Context, cfg config) error {
	if err := prepareOut(cfg.Out); err != nil {
		return err
	}
	if cfg.Gallery {
		if err := writeGallery(cfg.Root, cfg.Out, cfg.Log); err != nil {
			return err
		}
	}
	if cfg.Examples {
		if err := writeSnapshots(cfg.Root, cfg.Out, cfg.Log); err != nil {
			return err
		}
	}
	if err := copyPDFs(cfg.Root, cfg.Out); err != nil {
		return err
	}

	pages, err := discoverPages(cfg.Root, cfg.Out)
	if err != nil {
		return err
	}
	nav := buildNav(pages, examplesNav(cfg.Gallery, cfg.Examples))
	links := newResolver(cfg.Root, cfg.Out, cfg.BaseURL, pages)
	links.addFile(examplesFile)

	// Every page's links are resolved before anything is written, so a build
	// with broken links reports all of them at once.
	bodies := make([]string, len(pages))
	webPages := make([]*web.Page, len(pages))
	dates := gitDates(cfg.Root)
	// The examples page has no file of its own; it changes with the site.
	var siteDates fileDates
	var errs []error
	for i, page := range pages {
		bodies[i], err = renderMarkdown(page, links)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		d := datesOf(cfg.Root, page.Source, dates)
		if siteDates.Published.IsZero() || d.Published.Before(siteDates.Published) {
			siteDates.Published = d.Published
		}
		if d.LastUpdated.After(siteDates.LastUpdated) {
			siteDates.LastUpdated = d.LastUpdated
		}
		webPages[i], err = newWebPage(page, links, cfg.BaseURL, d)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if err := errors.Join(errs...); err != nil {
		return err
	}
	examplesPage := &web.Page{
		Path:        "/" + strings.TrimSuffix(examplesFile, "index.html"),
		Title:       "Examples",
		Description: "The component gallery, static snapshots of the example servers, and reference PDFs, all rendered by go-mx.",
		Type:        "Examples",
		ContentType: mx.ContentTypeHTML,
		Published:   siteDates.Published,
		LastUpdated: siteDates.LastUpdated,
	}
	site := &web.Site{
		BaseURL:     cfg.BaseURL,
		Title:       "go-mx",
		Description: siteDescription,
		Lang:        "en",
		Sources: []web.PageSource{web.PageSourceFunc(func(context.Context, bool) iter.Seq2[*web.Page, error] {
			return func(yield func(*web.Page, error) bool) {
				for _, p := range append(webPages, examplesPage) {
					if !yield(p, nil) {
						return
					}
				}
			}
		})},
		LLMsTxtOptions: &web.LLMsTxtOptions{
			Context: "The Markdown versions of the documentation pages of go-mx, a Go library that builds HTML, SVG, XML and PDF documents from type-safe components, with a port of shadcn/ui.",
		},
		MarkdownAlternates: true,
	}
	if err := site.Validate(); err != nil {
		return err
	}

	for i, page := range pages {
		markdownFile, err := markdownFileOf(page.Path)
		if err != nil {
			return err
		}
		doc := shell(nav, page.File(), page.Title, markdownContent(page, bodies[i], markdownFile))
		if err := writeDocument(site, doc, webPages[i], filepath.Join(cfg.Out, page.File())); err != nil {
			return fmt.Errorf("%s: %w", page.Source, err)
		}
		alternate, err := site.MarkdownAlternate(webPages[i])
		if err != nil {
			return err
		}
		if err := writeFile(filepath.Join(cfg.Out, markdownFile), alternate.Bytes()); err != nil {
			return err
		}
	}
	doc := shell(nav, examplesFile, examplesPage.Title, examplesContent(cfg.Gallery, cfg.Examples))
	if err := writeDocument(site, doc, examplesPage, filepath.Join(cfg.Out, examplesFile)); err != nil {
		return err
	}

	// The site files are written by the handlers that would serve them, so
	// they are what a server built on web.Site sends. robots.txt is left out:
	// crawlers read it only from the root of the host, and this site lives
	// below /go-mx/ of a host it does not own.
	for _, f := range []struct {
		name    string
		handler http.HandlerFunc
	}{
		{web.DefaultSitemapPath, site.HandleSitemapXML},
		{web.LLMsTxtPath, site.HandleLLMsTxt},
	} {
		if err := writeResponse(f.handler, filepath.Join(cfg.Out, strings.TrimPrefix(f.name, "/"))); err != nil {
			return fmt.Errorf("%s: %w", f.name, err)
		}
	}
	fmt.Fprintf(cfg.Log, "wrote %d documentation pages to %s\n", len(pages), cfg.Out)

	if cfg.Check {
		if err := checkLinks(cfg.Out); err != nil {
			return err
		}
		fmt.Fprintln(cfg.Log, "every internal link and fragment resolves")
	}
	return nil
}

// newWebPage returns the web.Page of page for the sitemap, llms.txt, the
// page metadata and the Markdown alternate, published and updated at dates. Its content is the Markdown with
// every link made absolute, because an alternate is read without the site
// around it.
func newWebPage(page *Page, links *resolver, baseURL string, dates fileDates) (*web.Page, error) {
	var errs []error
	content := rewriteMarkdownLinks(page.Body, func(dest string) string {
		t, err := links.resolve(page, dest)
		if err != nil {
			errs = append(errs, err)
			return dest
		}
		if t.File == "" && strings.HasPrefix(dest, "#") {
			return dest
		}
		return t.absolute(baseURL)
	})
	if err := errors.Join(errs...); err != nil {
		return nil, err
	}
	markdownPath, err := web.MarkdownPath(page.Path)
	if err != nil {
		return nil, err
	}
	return &web.Page{
		Path:        page.Path,
		Title:       page.Title,
		Description: page.Description,
		Type:        page.Section,
		ContentType: mx.ContentTypeMarkdown,
		Content:     content,
		Published:   dates.Published,
		LastUpdated: dates.LastUpdated,
		Alternates:  []web.Alternate{{Type: mx.ContentTypeMarkdown, Path: markdownPath}},
	}, nil
}

// markdownFileOf returns the site file of the Markdown alternate of the page
// at pagePath, see web.MarkdownPath.
func markdownFileOf(pagePath string) (string, error) {
	p, err := web.MarkdownPath(pagePath)
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(p, "/"), nil
}

// prepareOut makes dir an empty directory for the site. A directory that is
// not empty must hold the marker of an earlier build, which this one then
// replaces; anything else is refused, so a mistyped -out cannot delete a
// directory that is not a generated site.
func prepareOut(dir string) error {
	entries, err := os.ReadDir(dir)
	switch {
	case errors.Is(err, os.ErrNotExist):
	case err != nil:
		return err
	case len(entries) > 0:
		if _, err := os.Stat(filepath.Join(dir, outMarker)); err != nil {
			return fmt.Errorf("%s is not empty and holds no %s file of an earlier build; refusing to delete it", dir, outMarker)
		}
		if err := os.RemoveAll(dir); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, outMarker), []byte("written by go -C tools run ./gen-docs-site\n"), 0o644)
}

// writeDocument adds the page metadata of the site to doc (canonical link,
// description, Open Graph, the Markdown alternate) and writes it to file.
func writeDocument(site *web.Site, doc *html.Document, page *web.Page, file string) error {
	if err := site.AddPageMetadata(doc, page); err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := doc.Render(context.Background(), mx.NewCheckedWriter(&buf).WithIndent("", "  ")); err != nil {
		return err
	}
	return writeFile(file, buf.Bytes())
}

// writeResponse writes the body of handler's response to a GET request to
// file, failing on a status other than 200.
func writeResponse(handler http.HandlerFunc, file string) error {
	recorder := httptest.NewRecorder()
	handler(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	if recorder.Code != http.StatusOK {
		return fmt.Errorf("status %d: %s", recorder.Code, recorder.Body)
	}
	return writeFile(file, recorder.Body.Bytes())
}

// writeFile writes data to file, creating its directory.
func writeFile(file string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return err
	}
	return os.WriteFile(file, data, 0o644)
}
