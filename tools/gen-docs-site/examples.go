package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// galleryDir is where the component gallery is published, the URL the
// Jekyll site served it under.
const galleryDir = "shadcn/gallery"

// examplesDir is where the example snapshots and PDFs are published.
const examplesDir = "examples"

// example is one example command published as a static snapshot.
type example struct {
	// Name is the directory below examplesDir.
	Name string
	// Command is the package directory of the command.
	Command string
	// Title and Description introduce it on the examples page.
	Title       string
	Description string
	// Pages are the snapshot files the command writes, the first one the
	// entry page, with their link texts.
	Pages []navItem
}

// examples are the example commands with a -snapshot mode.
var examples = []example{
	{
		Name:        "form",
		Command:     "cmd/example-form",
		Title:       "Reflected form",
		Description: "mx.ReflectFormHandler renders a Profile struct through shadcn.FieldDecider: every field kind, a section, a repeatable editor, and the inline errors of an invalid submission.",
		Pages:       []navItem{{"Form", "index.html"}, {"After an invalid submission", "invalid.html"}},
	},
	{
		Name:        "sse",
		Command:     "cmd/example-sse",
		Title:       "Server-sent events chat",
		Description: "A chat transcript streamed with mx.SSEResponse and the htmx SSE extension, a reply appended token by token out of band to a KeyedID target. The snapshot shows the transcript the stream ends with.",
		Pages:       []navItem{{"Finished transcript", "index.html"}},
	},
	{
		Name:        "logstream",
		Command:     "cmd/example-logstream",
		Title:       "Log stream",
		Description: "A simulated production log in a logview.View: JSON records, nested values, stack traces and plain lines, with filter, pause and a split view. The snapshot holds a fixed batch of generated lines.",
		Pages:       []navItem{{"One view", "index.html"}, {"Split view", "split.html"}},
	},
	{
		Name:        "highlight",
		Command:     "cmd/highlight-demo",
		Title:       "Syntax highlighting",
		Description: "The highlight package colors a Go file and shows the go-mx source highlight.GoSource generates to build the same markup.",
		Pages:       []navItem{{"Light theme", "index.html"}, {"Dark theme", "dark.html"}},
	},
}

// referencePDF is a reference PDF of a golden test, published as a PDF
// example.
type referencePDF struct {
	// Source is the path in the repository, Name the file name in the site.
	Source, Name string
	// Title and Description introduce it on the examples page; Test is the
	// golden test that renders it.
	Title, Description, Test string
}

// referencePDFs are the golden reference PDFs that show what the pdf
// packages draw. The parity PDFs of the legacy fpdf stack are left out: they
// show the old renderer, not this one.
var referencePDFs = []referencePDF{
	{
		Source:      "pdf/testdata/reference.pdf",
		Name:        "pdf-reference.pdf",
		Title:       "pdf: text, shapes and layout",
		Description: "The golden document of the pdf package: text flow, positioned text, rectangles, circles, lines and polygons.",
		Test:        "pdf/golden_test.go",
	},
	{
		Source:      "pdf/testdata/svg_reference.pdf",
		Name:        "svg-reference.pdf",
		Title:       "pdf.SVG: svg element trees as vector graphics",
		Description: "svg package elements drawn natively by pdf.SVG: shapes, paths, gradients, transforms, text and viewport scaling.",
		Test:        "pdf/svggolden_test.go",
	},
	{
		Source:      "pdf/pdftable/testdata/table_reference.pdf",
		Name:        "table-reference.pdf",
		Title:       "pdftable: data tables",
		Description: "Fixed, weighted and measured columns, wrapped cells, the style cascade, repeated header rows across page breaks, and every grid rule combination.",
		Test:        "pdf/pdftable/golden_test.go",
	},
}

// runGo runs the go command with args in the repository root dir, its
// output going to log.
func runGo(root string, log io.Writer, args ...string) error {
	fmt.Fprintf(log, "go %s\n", strings.Join(args, " "))
	cmd := exec.Command("go", args...)
	cmd.Dir = root
	cmd.Stdout = log
	cmd.Stderr = log
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go %s: %w", strings.Join(args, " "), err)
	}
	return nil
}

// writeGallery writes the component gallery into the site with the command
// the gallery's README documents, its sidebar linking back to the site.
func writeGallery(root, out string, log io.Writer) error {
	dir, err := filepath.Abs(filepath.Join(out, galleryDir))
	if err != nil {
		return err
	}
	home := relativeURL(galleryDir+"/index.html", "index.html")
	return runGo(root, log, "run", "./cmd/shadcn-gallery", "-out", dir, "-static-highlight", "-home", home)
}

// writeSnapshots writes the static snapshot of every example into the site.
func writeSnapshots(root, out string, log io.Writer) error {
	for _, ex := range examples {
		dir, err := filepath.Abs(filepath.Join(out, examplesDir, ex.Name))
		if err != nil {
			return err
		}
		if err := runGo(root, log, "run", "./"+ex.Command, "-snapshot", dir, "-snapshot-index", "../index.html"); err != nil {
			return err
		}
	}
	return nil
}

// copyPDFs copies the reference PDFs into the site.
func copyPDFs(root, out string) error {
	for _, pdf := range referencePDFs {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(pdf.Source)))
		if err != nil {
			return err
		}
		dst := filepath.Join(out, examplesDir, pdf.Name)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(dst, data, 0o644); err != nil {
			return err
		}
	}
	return nil
}

// examplesFile is the site file of the examples page.
var examplesFile = path.Join(examplesDir, "index.html")

// examplesNav is the Examples section of the navigation: the gallery, the
// examples page and every snapshot.
func examplesNav(withGallery, withExamples bool) navSection {
	s := navSection{Title: "Examples"}
	if withGallery {
		s.Items = append(s.Items, navItem{"Component gallery", galleryDir + "/index.html"})
	}
	s.Items = append(s.Items, navItem{"All examples", examplesFile})
	if withExamples {
		for _, ex := range examples {
			s.Items = append(s.Items, navItem{ex.Title, path.Join(examplesDir, ex.Name, ex.Pages[0].File)})
		}
	}
	return s
}

// examplesContent is the main column of the examples page: the gallery, a
// card per snapshot with its live command and source, and the reference
// PDFs.
func examplesContent(withGallery, withExamples bool) mx.Component {
	from := examplesFile
	link := func(file string) string { return relativeURL(from, file) }
	linkClass := html.Class("font-medium underline underline-offset-4")
	var cards mx.Components
	if withExamples {
		for _, ex := range examples {
			var pages mx.Components
			for i, p := range ex.Pages {
				variant := shadcn.ButtonOutline
				if i == 0 {
					variant = shadcn.ButtonDefault
				}
				pages = append(pages, html.AHRef(link(path.Join(examplesDir, ex.Name, p.File)),
					html.Class(shadcn.ButtonClasses(variant, shadcn.SizeSM)), p.Title))
			}
			cards = append(cards, shadcn.Card(
				shadcn.CardHeader(
					shadcn.CardTitle(ex.Title),
					shadcn.CardDescription(ex.Description),
				),
				shadcn.CardContent(
					html.DivClass("flex flex-wrap gap-2", pages),
					html.PClass("mt-4 text-sm text-muted-foreground",
						"Live: ", html.Code(html.Class("rounded bg-muted px-1.5 py-0.5 font-mono text-xs"), "go run ./"+ex.Command),
						" · ", html.AHRef(repoURL+"tree/main/"+ex.Command, linkClass, "Source"),
					),
				),
			))
		}
	}
	var pdfs mx.Components
	for _, pdf := range referencePDFs {
		pdfs = append(pdfs, html.LI(html.Class("mt-3"),
			html.AHRef(link(path.Join(examplesDir, pdf.Name)), linkClass, pdf.Title),
			html.Span(html.Class("text-muted-foreground"), " — ", pdf.Description, " Rendered by "),
			html.AHRef(repoURL+"blob/main/"+pdf.Test, linkClass, html.Code(pdf.Test)),
			html.Span(html.Class("text-muted-foreground"), "."),
		))
	}
	return html.Article(
		html.H1Class("text-3xl font-bold tracking-tight", "Examples"),
		html.PClass("mt-4 leading-7",
			"Everything on this site is rendered by go-mx. The examples below are servers, which a static site cannot run, so each one is published as a snapshot: "+
				"the HTML its handler serves, written at build time, with a note on top and the command that runs it live."),
		mx.If(withGallery,
			shadcn.Card(html.Class("mt-8"),
				shadcn.CardHeader(
					shadcn.CardTitle("Component gallery"),
					shadcn.CardDescription("Every shadcn/ui and Magic UI component of the shadcn package, each example a live preview next to the Go source that renders it."),
				),
				shadcn.CardContent(
					html.AHRef(link(galleryDir+"/index.html"), html.Class(shadcn.ButtonClasses(shadcn.ButtonDefault, shadcn.SizeSM)), "Open the gallery"),
				),
			),
		),
		mx.If(len(cards) > 0,
			html.H2Class("mt-12 text-2xl font-semibold tracking-tight", "Example servers"),
			html.DivClass("mt-6 grid gap-6", cards),
		),
		html.H2Class("mt-12 text-2xl font-semibold tracking-tight", "PDF documents"),
		html.PClass("mt-4 leading-7",
			"The reference documents of the golden tests of the pdf packages, rendered by go-mx's own PDF engine."),
		html.UL(html.Class("mt-2"), pdfs),
	)
}
