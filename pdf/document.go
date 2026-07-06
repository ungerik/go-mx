package pdf

import (
	"bytes"
	"cmp"
	"context"
	"fmt"
	"io"
	"net/http"
	"slices"
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

	// Attachments are embedded as document-level files. Attachments with a
	// Relationship are also listed as PDF/A-3 associated files, as
	// ZUGFeRD/Factur-X hybrid invoices require.
	Attachments []Attachment

	// XMP, when non-nil, is built into the document's XMP metadata packet.
	// Empty XMP fields fall back to the document metadata above, and the
	// PDF info dictionary (including its dates) is kept consistent with the
	// packet, as PDF/A validators require. When XMP.FacturX is set, an
	// attachment with the declared DocumentFileName must be present.
	XMP *XMPMetadata

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
	if len(d.Attachments) > 0 {
		r.SetAttachments(d.Attachments)
		// Validate relationships up front so an invalid one surfaces from
		// Render rather than only from the later Output/Close that embeds them.
		for i := range d.Attachments {
			if rel := d.Attachments[i].Relationship; rel != "" && !rel.Valid() {
				r.SetError(fmt.Errorf("invalid AFRelationship %q for attachment %q", rel, d.Attachments[i].Filename))
			}
		}
	}
	if d.XMP != nil {
		d.applyXMP(r)
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

// applyXMP builds the XMP metadata packet from d.XMP, filling empty fields
// from the document metadata, and keeps the info dictionary consistent with
// the packet as PDF/A requires: every info-dictionary field the packet also
// carries (title, author, subject, keywords, creator, producer) and the
// fully-specified creation/modification instants are set on the renderer, so
// putinfo writes the same values the packet does. PDF/A validators reject a
// document whose /Info entry and its XMP property are both present but differ.
func (d *Document) applyXMP(r *Renderer) {
	m := *d.XMP
	m.Title = cmp.Or(m.Title, d.Title)
	m.Author = cmp.Or(m.Author, d.Author)
	m.Subject = cmp.Or(m.Subject, d.Subject)
	m.Keywords = cmp.Or(m.Keywords, d.Keywords)
	m.CreatorTool = cmp.Or(m.CreatorTool, d.Creator)
	// the engine default producer, written to /Info even when unset
	m.Producer = cmp.Or(m.Producer, "FPDF "+cnFpdfVersion)
	// Mirror the resolved metadata onto the renderer so /Info matches the XMP
	// packet. An empty field is left unset (putinfo omits it), which the XMP
	// packet does too, so the two stay consistent.
	if m.Title != "" {
		r.SetTitle(m.Title, true)
	}
	if m.Author != "" {
		r.SetAuthor(m.Author, true)
	}
	if m.Subject != "" {
		r.SetSubject(m.Subject, true)
	}
	if m.Keywords != "" {
		r.SetKeywords(m.Keywords, true)
	}
	if m.CreatorTool != "" {
		r.SetCreator(m.CreatorTool, true)
	}
	r.SetProducer(m.Producer, true)
	if m.CreateDate.IsZero() {
		m.CreateDate = timeOrNow(r.creationDate)
	}
	if m.ModifyDate.IsZero() {
		m.ModifyDate = r.modDate
		if m.ModifyDate.IsZero() {
			m.ModifyDate = m.CreateDate
		}
	}
	r.SetCreationDate(m.CreateDate)
	r.SetModificationDate(m.ModifyDate)

	// PDF/A mandates a minimum header version (1.4 for part 1, 1.7 for parts
	// 2 and 3); the engine defaults to 1.3, which validators reject outright.
	// Raise it to the version the declared part requires.
	if m.PDFAPart > 0 {
		want := pdfVers1_7
		if m.PDFAPart == 1 {
			want = pdfVers1_4
		}
		if r.pdfVersion < want {
			r.pdfVersion = want
		}
	}

	if m.FacturX != nil {
		f, err := m.FacturX.withDefaults()
		if err != nil {
			r.SetError(err)
			return
		}
		i := slices.IndexFunc(r.attachments, func(a Attachment) bool {
			return a.Filename == f.DocumentFileName
		})
		if i < 0 {
			r.SetError(fmt.Errorf("XMP Factur-X metadata references %q but no attachment has that filename", f.DocumentFileName))
			return
		}
		// The referenced XML must be a PDF/A-3 associated file: listed in the
		// catalog /AF array (which needs a Relationship) and typed with a MIME
		// /Subtype, or validators ignore the invoice.
		if a := r.attachments[i]; a.Relationship == "" || a.MIMEType == "" {
			r.SetError(fmt.Errorf("XMP Factur-X attachment %q must set Relationship and MIMEType to be a PDF/A-3 associated file", f.DocumentFileName))
			return
		}
	}

	packet, err := m.XML()
	if err != nil {
		r.SetError(err)
		return
	}
	r.SetXmpMetadata(packet)
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
