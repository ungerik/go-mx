// Copyright ©2023 The go-pdf Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package pdf

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"slices"
	"strings"
	"time"
)

// Attachment defines a content to be included in the pdf, in one
// of the following ways :
//   - associated with the document as a whole : see SetAttachments()
//   - accessible via a link localized on a page : see AddAttachmentAnnotation()
type Attachment struct {
	Content []byte

	// Filename is the displayed name of the attachment
	Filename string

	// Description is only displayed when using AddAttachmentAnnotation(),
	// and might be modified by the pdf reader.
	Description string

	// Relationship, when non-empty, declares the attachment as a PDF/A-3
	// associated file with this /AFRelationship, listed in the document
	// catalog's /AF array. ZUGFeRD/Factur-X invoices embed their XML with
	// [AFRelationshipAlternative] (or [AFRelationshipData] for the MINIMUM
	// and BASIC WL profiles).
	Relationship AFRelationship

	// MIMEType, when non-empty, is written as the embedded file's /Subtype
	// (e.g. "text/xml"), required for PDF/A-3 associated files.
	MIMEType string

	// ModDate is the modification date written to the embedded file's
	// /Params dictionary. A zero value falls back to the document's
	// creation date (or the current time if that is unset too).
	ModDate time.Time

	objectNumber int // filled when content is included
}

// return the hex encoded checksum of `data`
func checksum(data []byte) string {
	sl := md5.Sum(data)
	return hex.EncodeToString(sl[:])
}

// Writes a compressed file like object as "/EmbeddedFile". Compressing is
// done with deflate. Includes length, compressed length, MD5 checksum and
// modification date, plus the MIME /Subtype when the attachment declares one.
func (r *Renderer) writeCompressedFileObject(a *Attachment) {
	lenUncompressed := len(a.Content)
	sum := checksum(a.Content)
	mem := xmem.compress(a.Content)
	defer xmem.release(mem)
	compressed := mem.Bytes()
	lenCompressed := len(compressed)
	r.newobj()
	r.out("<< /Type /EmbeddedFile")
	if a.MIMEType != "" {
		r.outf("/Subtype /%s", escapeName(a.MIMEType))
	}
	modDate := a.ModDate
	if modDate.IsZero() {
		modDate = timeOrNow(r.creationDate)
	}
	r.outf("/Length %d /Filter /FlateDecode", lenCompressed)
	r.outf("/Params << /CheckSum <%s> /Size %d /ModDate %s >> >>",
		sum, lenUncompressed, r.textstring(pdfDate(modDate, true)))
	r.putstream(compressed)
	r.out("endobj")
}

// Embed includes the content of `a`, and update its internal reference.
func (r *Renderer) embed(a *Attachment) {
	if a.objectNumber != 0 { // already embedded (objectNumber start at 2)
		return
	}
	if a.Relationship != "" && !a.Relationship.Valid() {
		r.SetError(fmt.Errorf("invalid AFRelationship %q for attachment %q", a.Relationship, a.Filename))
		return
	}
	oldState := r.state
	r.state = 1 // we write file content in the main buffer
	r.writeCompressedFileObject(a)
	streamID := r.n
	r.newobj()
	r.out("<< /Type /Filespec")
	// PDF/A-3 requires both /F and /UF, and /UF also in the /EF dictionary
	r.outf("/F %s /UF %s", r.textstring(pdfDocEncode(a.Filename)), r.textstring(utf8toutf16(a.Filename)))
	if a.Relationship != "" {
		r.outf("/AFRelationship /%s", a.Relationship)
	}
	r.outf("/EF << /F %d 0 R /UF %d 0 R >> /Desc %s\n>>",
		streamID, streamID,
		r.textstring(utf8toutf16(a.Description)))
	r.out("endobj")
	a.objectNumber = r.n
	r.state = oldState
}

// SetAttachments writes attachments as embedded files (document attachment).
// These attachments are global, see AddAttachmentAnnotation() for a link
// anchored in a page. Note that only the last call of SetAttachments is
// useful, previous calls are discarded. Be aware that not all PDF readers
// support document attachments. See the SetAttachment example for a
// demonstration of this method.
func (r *Renderer) SetAttachments(as []Attachment) {
	r.attachments = as
}

// embed current attachments. store object numbers
// for later use by getEmbeddedFiles()
func (r *Renderer) putAttachments() {
	for i, a := range r.attachments {
		r.embed(&a)
		r.attachments[i] = a
	}
}

// return /EmbeddedFiles tree name catalog entry.
func (r *Renderer) getEmbeddedFiles() string {
	type nameRef struct {
		key          string // the written key, UTF-16BE encoded
		objectNumber int
	}
	refs := make([]nameRef, len(r.attachments))
	seen := make(map[string]int, len(r.attachments))
	for i, a := range r.attachments {
		name := a.Filename
		if name == "" {
			name = fmt.Sprintf("Attachement%d", i+1)
		}
		// name-tree keys must be unique; disambiguate repeated filenames
		seen[name]++
		if n := seen[name]; n > 1 {
			name = fmt.Sprintf("%s (%d)", name, n)
		}
		refs[i] = nameRef{utf8toutf16(name), a.objectNumber}
	}
	// name-tree keys must be sorted by the byte value of the written strings,
	// which are UTF-16BE, so compare on that encoding not the raw UTF-8 name
	slices.SortFunc(refs, func(a, b nameRef) int {
		return strings.Compare(a.key, b.key)
	})
	names := make([]string, len(refs))
	for i, ref := range refs {
		names[i] = fmt.Sprintf("%s %d 0 R ", r.textstring(ref.key), ref.objectNumber)
	}
	nameTree := fmt.Sprintf("<< /Names [\n %s \n] >>", strings.Join(names, "\n"))
	return nameTree
}

// getAssociatedFiles returns the catalog /AF array listing the filespecs of
// attachments declared with an AFRelationship (PDF/A-3 associated files),
// or "" if there are none.
func (r *Renderer) getAssociatedFiles() string {
	var refs []string
	for _, a := range r.attachments {
		if a.Relationship != "" && a.objectNumber != 0 {
			refs = append(refs, fmt.Sprintf("%d 0 R", a.objectNumber))
		}
	}
	if len(refs) == 0 {
		return ""
	}
	return "[" + strings.Join(refs, " ") + "]"
}

// ---------------------------------- Annotations ----------------------------------

type annotationAttach struct {
	*Attachment

	x, y, w, h float64 // fpdf coordinates (y diff and scaling done)
}

// AddAttachmentAnnotation puts a link on the current page, on the rectangle
// defined by `x`, `y`, `w`, `h`. This link points towards the content defined
// in `a`, which is embedded in the document. Note than no drawing is done by
// this method : a method like `Cell()` or `Rect()` should be called to
// indicate to the reader that there is a link here. Requiring a pointer to an
// Attachment avoids useless copies in the resulting pdf: attachment pointing
// to the same data will have their content only be included once, and be
// shared amongst all links. Be aware that not all PDF readers support
// annotated attachments. See the AddAttachmentAnnotation example for a
// demonstration of this method.
func (r *Renderer) AddAttachmentAnnotation(a *Attachment, x, y, w, h float64) {
	if a == nil {
		return
	}
	r.pageAttachments[r.page] = append(r.pageAttachments[r.page], annotationAttach{
		Attachment: a,
		x:          x * r.k, y: r.hPt - y*r.k, w: w * r.k, h: h * r.k,
	})
}

// embed current annotations attachments. store object numbers
// for later use by putAttachmentAnnotationLinks(), which is
// called for each page.
func (r *Renderer) putAnnotationsAttachments() {
	// avoid duplication
	m := map[*Attachment]bool{}
	for _, l := range r.pageAttachments {
		for _, an := range l {
			if m[an.Attachment] { // already embedded
				continue
			}
			r.embed(an.Attachment)
		}
	}
}

func (r *Renderer) putAttachmentAnnotationLinks(out *bytes.Buffer, page int) {
	for _, an := range r.pageAttachments[page] {
		x1, y1, x2, y2 := an.x, an.y, an.x+an.w, an.y-an.h
		as := fmt.Sprintf("<< /Type /XObject /Subtype /Form /BBox [%.2f %.2f %.2f %.2f] /Length 0 >>",
			x1, y1, x2, y2)
		as += "\nstream\nendstream"

		fmt.Fprintf(out, "<< /Type /Annot /Subtype /FileAttachment /Rect [%.2f %.2f %.2f %.2f] /Border [0 0 0]\n",
			x1, y1, x2, y2)
		fmt.Fprintf(out, "/Contents %s ", r.textstring(utf8toutf16(an.Description)))
		fmt.Fprintf(out, "/T %s ", r.textstring(utf8toutf16(an.Filename)))
		fmt.Fprintf(out, "/AP << /N %s>>", as)
		fmt.Fprintf(out, "/FS %d 0 R >>\n", an.objectNumber)
	}
}
