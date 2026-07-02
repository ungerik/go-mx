package pdf

import (
	"bytes"
	"cmp"
	"context"
	"io"
	"net/http"
)

// ContentType is the MIME type of PDF output.
const ContentType = "application/pdf"

// Margins are the left, top and right page margins in document units. fpdf
// derives the bottom margin (the auto page-break trigger) from the top margin.
type Margins struct {
	Left, Top, Right float64
}

// Document is the top-level PDF builder, the analog of html.Document. It holds
// document metadata, page setup, a default font and the body components, and
// can render into an existing [Renderer] or produce a finished PDF directly.
//
// All zero-valued setup fields fall back to sensible defaults: A4 portrait in
// millimeters with a Helvetica 12pt font. The body is rendered after setup; it
// typically contains [Page] components, but any drawing primitive auto-starts
// the first page, so a single flow of text needs no explicit Page.
type Document struct {
	// Metadata written to the PDF info dictionary.
	Title    string
	Author   string
	Subject  string
	Keywords string
	Creator  string

	// Page setup. Zero values default to A4 portrait in millimeters.
	Orientation Orientation
	Unit        Unit
	PageSize    PageSize

	// Margins overrides the page margins when non-nil; nil keeps fpdf defaults.
	Margins *Margins

	// Default font applied before the body. Zero values use Helvetica 12pt.
	FontFamily string
	FontStyle  FontStyle
	FontSize   float64

	// Header and Footer, if set, render at the top and bottom of every page.
	// They run inside fpdf's page lifecycle with the context passed to Render.
	Header Component
	Footer Component

	// Body holds the page content.
	Body Component
}

// NewDocument creates a Document with the given title and body components.
func NewDocument(title string, body ...any) *Document {
	return &Document{
		Title: title,
		Body:  AsComponents(body...),
	}
}

// NewRenderer creates a [Renderer] configured from the document's page setup,
// applying the A4/portrait/millimeter defaults for any unset field.
func (d *Document) NewRenderer() *Renderer {
	orientation := cmp.Or(d.Orientation, OrientationPortrait)
	unit := cmp.Or(d.Unit, UnitMillimeter)
	size := cmp.Or(d.PageSize, PageSizeA4)
	return NewRenderer(orientation, unit, size)
}

// Render applies the document's metadata, margins, default font and
// header/footer to r, then renders the body. The Document is itself a
// [Component], so it can be embedded in a larger render.
func (d *Document) Render(ctx context.Context, r *Renderer) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	d.applySetup(ctx, r)
	if err := r.Error(); err != nil {
		return err
	}
	if d.Body != nil {
		if err := d.Body.Render(ctx, r); err != nil {
			return err
		}
	}
	return r.Error()
}

func (d *Document) applySetup(ctx context.Context, r *Renderer) {
	if d.Title != "" {
		r.SetTitle(d.Title, true)
	}
	if d.Author != "" {
		r.SetAuthor(d.Author, true)
	}
	if d.Subject != "" {
		r.SetSubject(d.Subject, true)
	}
	if d.Keywords != "" {
		r.SetKeywords(d.Keywords, true)
	}
	if d.Creator != "" {
		r.SetCreator(d.Creator, true)
	}
	if d.Margins != nil {
		r.SetMargins(d.Margins.Left, d.Margins.Top, d.Margins.Right)
	}

	family := cmp.Or(d.FontFamily, DefaultFontFamily)
	size := d.FontSize
	if size == 0 {
		size = DefaultFontSize
	}
	r.SetFont(family, string(d.FontStyle), size)

	// The renderer's header/footer callbacks have no error return, so a
	// component error would otherwise be lost: fold it into the renderer's
	// error state (kept until cleared) so it surfaces from Render's final
	// r.Error(). The callbacks are always set — nil clears them — so that a
	// document without Header/Footer does not inherit the callbacks of a
	// document previously rendered into the same renderer.
	if d.Header != nil {
		r.SetHeaderFunc(func() {
			if err := d.Header.Render(ctx, r); err != nil {
				r.SetError(err)
			}
		})
	} else {
		r.SetHeaderFunc(nil)
	}
	if d.Footer != nil {
		r.SetFooterFunc(func() {
			if err := d.Footer.Render(ctx, r); err != nil {
				r.SetError(err)
			}
		})
	} else {
		r.SetFooterFunc(nil)
	}
}

// Output renders the document to its own renderer and writes the PDF to w.
func (d *Document) Output(ctx context.Context, w io.Writer) error {
	r := d.NewRenderer()
	if err := d.Render(ctx, r); err != nil {
		return err
	}
	return r.Output(w)
}

// Bytes renders the document and returns the encoded PDF.
func (d *Document) Bytes(ctx context.Context) ([]byte, error) {
	var buf bytes.Buffer
	if err := d.Output(ctx, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// OutputFile renders the document and writes the PDF to the named file.
func (d *Document) OutputFile(ctx context.Context, filename string) error {
	r := d.NewRenderer()
	if err := d.Render(ctx, r); err != nil {
		return err
	}
	return r.OutputFileAndClose(filename)
}

// ServeHTTP renders the document and serves it as application/pdf, using the
// request context for cancellation. On error it responds with a short, generic
// 500 message and does not leak the underlying error.
func (d *Document) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	pdfBytes, err := d.Bytes(req.Context())
	if err != nil {
		http.Error(w, "failed to render PDF", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", ContentType)
	_, _ = w.Write(pdfBytes)
}
