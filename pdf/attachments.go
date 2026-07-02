// Copyright ©2023 The go-pdf Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package pdf

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
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

	objectNumber int // filled when content is included
}

// return the hex encoded checksum of `data`
func checksum(data []byte) string {
	sl := md5.Sum(data)
	return hex.EncodeToString(sl[:])
}

// Writes a compressed file like object as "/EmbeddedFile". Compressing is
// done with deflate. Includes length, compressed length and MD5 checksum.
func (r *Renderer) writeCompressedFileObject(content []byte) {
	lenUncompressed := len(content)
	sum := checksum(content)
	mem := xmem.compress(content)
	defer xmem.release(mem)
	compressed := mem.Bytes()
	lenCompressed := len(compressed)
	r.newobj()
	r.outf("<< /Type /EmbeddedFile /Length %d /Filter /FlateDecode /Params << /CheckSum <%s> /Size %d >> >>\n",
		lenCompressed, sum, lenUncompressed)
	r.putstream(compressed)
	r.out("endobj")
}

// Embed includes the content of `a`, and update its internal reference.
func (r *Renderer) embed(a *Attachment) {
	if a.objectNumber != 0 { // already embedded (objectNumber start at 2)
		return
	}
	oldState := r.state
	r.state = 1 // we write file content in the main buffer
	r.writeCompressedFileObject(a.Content)
	streamID := r.n
	r.newobj()
	r.outf("<< /Type /Filespec /F () /UF %s /EF << /F %d 0 R >> /Desc %s\n>>",
		r.textstring(utf8toutf16(a.Filename)),
		streamID,
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
	names := make([]string, len(r.attachments))
	for i, as := range r.attachments {
		names[i] = fmt.Sprintf("(Attachement%d) %d 0 R ", i+1, as.objectNumber)
	}
	nameTree := fmt.Sprintf("<< /Names [\n %s \n] >>", strings.Join(names, "\n"))
	return nameTree
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
