package pdf

import (
	"bytes"
	"context"
	encxml "encoding/xml"
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

// facturXTestDocument builds a minimal ZUGFeRD/Factur-X style hybrid document
// with fixed dates so the output is deterministic.
func facturXTestDocument(fixed time.Time) (*Document, []byte) {
	invoiceXML := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<rsm:CrossIndustryInvoice xmlns:rsm="urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100">
  <rsm:ExchangedDocument/>
</rsm:CrossIndustryInvoice>`)
	doc := NewDocument("Invoice R2024-001",
		Paragraph("Invoice R2024-001"),
	)
	doc.Author = "ACME GmbH"
	doc.Subject = "Invoice R2024-001 for services"
	doc.Keywords = "invoice, factur-x"
	doc.Creator = "go-mx/pdf test"
	doc.Attachments = []Attachment{{
		Content:      invoiceXML,
		Filename:     "factur-x.xml",
		Description:  "Factur-X invoice",
		Relationship: AFRelationshipAlternative,
		MIMEType:     "text/xml",
		ModDate:      fixed,
	}}
	doc.XMP = &XMPMetadata{
		PDFAPart:   3,
		CreateDate: fixed,
		FacturX:    &FacturX{ConformanceLevel: FacturXEN16931},
	}
	return doc, invoiceXML
}

// The XMP packet must be well-formed XML wrapped in an xpacket, carrying the
// PDF/A identification, the document metadata, and — for ZUGFeRD/Factur-X —
// the extension schema declaration plus the fx: properties, because that is
// what validators check the embedded invoice XML against.
func TestXMPMetadataXML(t *testing.T) {
	fixed := time.Date(2024, 5, 6, 7, 8, 9, 0, time.UTC)
	m := &XMPMetadata{
		Title:           "Invoice R2024-001",
		Author:          "ACME GmbH",
		Subject:         "Invoice for services",
		Keywords:        "invoice, factur-x",
		CreatorTool:     "go-mx/pdf",
		Producer:        "go-mx/pdf producer",
		CreateDate:      fixed,
		ModifyDate:      fixed,
		PDFAPart:        3,
		PDFAConformance: "B",
		FacturX:         &FacturX{ConformanceLevel: FacturXEN16931},
	}
	packet, err := m.XML()
	if err != nil {
		t.Fatal(err)
	}

	// well-formed XML end to end (the decoder consumes the xpacket
	// processing instructions along with the elements)
	decoder := encxml.NewDecoder(bytes.NewReader(packet))
	for {
		_, err := decoder.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			t.Fatalf("XMP packet is not well-formed XML: %v\npacket:\n%s", err, packet)
		}
	}

	s := string(packet)
	if !strings.HasPrefix(s, `<?xpacket begin="`+"\uFEFF"+`" id="W5M0MpCehiHzreSzNTczkc9d"?>`) {
		t.Errorf("packet does not start with the xpacket header:\n%.100s", s)
	}
	if !strings.HasSuffix(s, `<?xpacket end="w"?>`) {
		t.Errorf("packet does not end with the xpacket trailer:\n%.100s", s[len(s)-100:])
	}
	for _, want := range []string{
		`<pdfaid:part>3</pdfaid:part>`,
		`<pdfaid:conformance>B</pdfaid:conformance>`,
		`<rdf:li xml:lang="x-default">Invoice R2024-001</rdf:li>`,
		`<rdf:li>ACME GmbH</rdf:li>`,
		`<pdf:Producer>go-mx/pdf producer</pdf:Producer>`,
		`<pdf:Keywords>invoice, factur-x</pdf:Keywords>`,
		`<xmp:CreatorTool>go-mx/pdf</xmp:CreatorTool>`,
		`<xmp:CreateDate>2024-05-06T07:08:09+00:00</xmp:CreateDate>`,
		`<xmp:ModifyDate>2024-05-06T07:08:09+00:00</xmp:ModifyDate>`,
		`<pdfaSchema:namespaceURI>` + FacturXNamespaceURI + `</pdfaSchema:namespaceURI>`,
		`<pdfaSchema:prefix>fx</pdfaSchema:prefix>`,
		`<pdfaProperty:name>ConformanceLevel</pdfaProperty:name>`,
		`xmlns:fx="` + FacturXNamespaceURI + `"`,
		`<fx:DocumentType>INVOICE</fx:DocumentType>`,
		`<fx:DocumentFileName>factur-x.xml</fx:DocumentFileName>`,
		`<fx:Version>1.0</fx:Version>`,
		`<fx:ConformanceLevel>EN 16931</fx:ConformanceLevel>`,
	} {
		if !strings.Contains(s, want) {
			t.Errorf("packet does not contain %q", want)
		}
	}
}

// Metadata values must be XML-escaped in the packet, or a title like
// "A < B & C" would produce a malformed packet that validators reject.
func TestXMPMetadataXML_escapesValues(t *testing.T) {
	m := &XMPMetadata{Title: `A < B & "C"`}
	packet, err := m.XML()
	if err != nil {
		t.Fatal(err)
	}
	if want := `A &lt; B &amp; &quot;C&quot;`; !strings.Contains(string(packet), want) {
		t.Errorf("packet does not contain escaped title %q:\n%s", want, packet)
	}
}

// The conformance level names the profile of the embedded XML, which cannot
// be guessed — silently defaulting it would produce metadata that contradicts
// the embedded invoice, so it must be a loud error.
func TestXMPMetadataXML_facturXConformanceLevelRequired(t *testing.T) {
	m := &XMPMetadata{FacturX: &FacturX{}}
	_, err := m.XML()
	if err == nil {
		t.Fatal("expected error for missing FacturX.ConformanceLevel")
	}
}

// A Document with XMP and an associated-file attachment must produce all the
// PDF structures that make a ZUGFeRD/Factur-X hybrid readable: the /AF
// catalog entry with /AFRelationship, the typed /Subtype and /Params of the
// embedded file, the XMP metadata stream, the trailer file identifier, and
// info dictionary dates that are consistent with the XMP dates (PDF/A
// validators compare them).
func TestDocumentFacturX(t *testing.T) {
	fixed := time.Date(2024, 5, 6, 7, 8, 9, 0, time.UTC)
	doc, _ := facturXTestDocument(fixed)

	r := doc.NewRenderer()
	if err := doc.Render(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := r.Output(&buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	for _, want := range []string{
		"/Type /Filespec",
		"/F (factur-x.xml) /UF (",
		"/AFRelationship /Alternative",
		"/Subtype /text#2Fxml",
		"/ModDate (D:20240506070809+00'00')",
		"/AF [",
		"/Type /Metadata /Subtype /XML",
		"<fx:ConformanceLevel>EN 16931</fx:ConformanceLevel>",
		"<pdfaid:part>3</pdfaid:part>",
		"/CreationDate (D:20240506070809+00'00')",
		"/ID [<",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("PDF output does not contain %q", want)
		}
	}
}

// The XMP Factur-X declaration promises readers an embedded file of that
// name; rendering a document that does not contain it must fail instead of
// silently producing a broken hybrid invoice.
func TestDocumentFacturX_missingAttachment(t *testing.T) {
	doc := NewDocument("Invoice", Paragraph("Invoice"))
	doc.XMP = &XMPMetadata{
		PDFAPart: 3,
		FacturX:  &FacturX{ConformanceLevel: FacturXEN16931},
	}
	err := doc.Render(context.Background(), doc.NewRenderer())
	if err == nil {
		t.Fatal("expected error for missing factur-x.xml attachment")
	}
	if !strings.Contains(err.Error(), "factur-x.xml") {
		t.Errorf("error does not name the missing file: %v", err)
	}
}

// A document that declares a PDF/A part must raise the file header to the
// version that part requires (1.7 for PDF/A-2/3). The engine defaults to 1.3,
// which every PDF/A validator rejects outright, so a Factur-X (PDF/A-3)
// document with a 1.3 header can never validate regardless of everything else.
func TestDocumentFacturX_headerVersion(t *testing.T) {
	fixed := time.Date(2024, 5, 6, 7, 8, 9, 0, time.UTC)
	doc, _ := facturXTestDocument(fixed)

	r := doc.NewRenderer()
	if err := doc.Render(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := r.Output(&buf); err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); !strings.HasPrefix(got, "%PDF-1.7") {
		t.Errorf("PDF/A-3 header is not 1.7, starts with %.8q", got)
	}
}

// The XML the Factur-X XMP references must be a real PDF/A-3 associated file
// (declared with an AFRelationship and a MIME type), not merely an attachment
// with the matching name. Without those it never lands in the catalog /AF
// array and the embedded stream lacks the required /Subtype, so validators
// ignore the invoice — the declaration must fail loudly instead.
func TestDocumentFacturX_attachmentMustBeAssociatedFile(t *testing.T) {
	fixed := time.Date(2024, 5, 6, 7, 8, 9, 0, time.UTC)
	doc, _ := facturXTestDocument(fixed)
	doc.Attachments[0].Relationship = "" // strip the associated-file markers
	doc.Attachments[0].MIMEType = ""

	err := doc.Render(context.Background(), doc.NewRenderer())
	if err == nil {
		t.Fatal("expected error for Factur-X attachment that is not an associated file")
	}
	if !strings.Contains(err.Error(), "factur-x.xml") {
		t.Errorf("error does not name the attachment: %v", err)
	}
}

// An invalid AFRelationship on a document attachment must surface from Render,
// not silently succeed there and only fail later at Output — a caller that
// checks Render's error and then trusts the renderer would otherwise ship a
// broken associated-file declaration.
func TestDocumentInvalidAFRelationship(t *testing.T) {
	doc := NewDocument("Doc", Paragraph("body"))
	doc.Attachments = []Attachment{{
		Content:      []byte("data"),
		Filename:     "data.bin",
		Relationship: AFRelationship("Bogus"),
	}}
	err := doc.Render(context.Background(), doc.NewRenderer())
	if err == nil {
		t.Fatal("expected error for invalid AFRelationship")
	}
	if !strings.Contains(err.Error(), "Bogus") {
		t.Errorf("error does not name the invalid relationship: %v", err)
	}
}

// The XMP fields exposed on Document.XMP override the plain document metadata,
// and PDF/A requires the info dictionary to carry the same values the XMP
// packet does. An XMP field set differently from the document field must win in
// BOTH places, or a validator sees /Info and XMP disagree.
func TestDocumentXMPInfoConsistency(t *testing.T) {
	doc := NewDocument("Document Title", Paragraph("body"))
	doc.Author = "Document Author"
	doc.XMP = &XMPMetadata{
		Title:  "XMP Title",
		Author: "XMP Author",
	}
	r := doc.NewRenderer()
	if err := doc.Render(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	if got, want := r.GetTitle(), utf8toutf16("XMP Title"); got != want {
		t.Errorf("info /Title = %q, want the XMP value %q", got, want)
	}
	if got, want := r.GetAuthor(), utf8toutf16("XMP Author"); got != want {
		t.Errorf("info /Author = %q, want the XMP value %q", got, want)
	}
	if s := string(r.xmp); !strings.Contains(s, "XMP Title") || !strings.Contains(s, "XMP Author") {
		t.Errorf("XMP packet does not carry the overridden metadata:\n%s", s)
	}
}

// A stray control byte in a metadata value (XML 1.0 forbids C0 controls
// other than tab/LF/CR) must not invalidate the whole packet: it is replaced
// with U+FFFD and the packet stays well-formed XML.
func TestXMPMetadataXML_sanitizesControlChars(t *testing.T) {
	m := &XMPMetadata{Title: "A\x01B", Producer: "x\x00y"}
	packet, err := m.XML()
	if err != nil {
		t.Fatal(err)
	}
	s := string(packet)
	if !strings.Contains(s, "A�B") || !strings.Contains(s, "x�y") {
		t.Errorf("control characters not replaced with U+FFFD:\n%s", s)
	}
	decoder := encxml.NewDecoder(bytes.NewReader(packet))
	for {
		if _, err := decoder.Token(); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			t.Fatalf("sanitized XMP packet is not well-formed XML: %v", err)
		}
	}
	// legal whitespace controls pass through untouched
	m = &XMPMetadata{Title: "A\tB"}
	packet, err = m.XML()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(packet), "A\tB") {
		t.Error("tab was sanitized, but XML 1.0 allows it")
	}
}

// Combining XMP metadata with document protection must fail loudly in either
// call order: the legacy RC4 encryption cannot exempt the metadata stream, so
// the combination would silently produce a file that can never be PDF/A and
// whose XMP packet no plaintext scanner finds.
func TestSetProtectionXMPConflict(t *testing.T) {
	// protection first, then XMP via Document.Render
	doc := NewDocument("Doc", Paragraph("body"))
	doc.XMP = &XMPMetadata{}
	r := doc.NewRenderer()
	r.SetProtection(0, "", "owner")
	err := doc.Render(context.Background(), r)
	if err == nil {
		t.Fatal("expected error combining SetProtection with XMP metadata")
	}
	if !strings.Contains(err.Error(), "SetProtection") {
		t.Errorf("error does not name the conflict: %v", err)
	}

	// XMP first, then protection on a raw renderer
	r = NewRenderer(OrientationPortrait, UnitMillimeter, PageSizeA4)
	r.SetXmpMetadata([]byte("<x/>"))
	r.SetProtection(0, "", "owner")
	if err := r.Error(); err == nil {
		t.Fatal("expected error setting protection with XMP metadata present")
	}

	// leaving XMP unset or clearing it under protection is no conflict: no
	// metadata stream will be emitted, so nothing gets encrypted
	r = NewRenderer(OrientationPortrait, UnitMillimeter, PageSizeA4)
	r.SetProtection(0, "", "owner")
	r.SetXmpMetadata(nil)
	if err := r.Error(); err != nil {
		t.Errorf("clearing XMP under protection must not error: %v", err)
	}
}

// The XMP guard on SetProtection must not get in the way of plain protection:
// a document without XMP metadata still has to encrypt and output normally.
func TestSetProtectionWithoutXMP(t *testing.T) {
	doc := NewDocument("Protected", Paragraph("body"))
	r := doc.NewRenderer()
	r.SetProtection(0, "user", "owner")
	if err := doc.Render(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := r.Output(&buf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "/Encrypt") {
		t.Error("output has no /Encrypt dictionary; protection was not applied")
	}
}

// A producer set on the renderer must survive applyXMP when neither the XMP
// nor the document names one: the packet has to carry a producer (because
// /Info always does), but it must be the caller's value, not the engine
// default overwriting it. Both setter encodings have to round-trip.
func TestDocumentXMPKeepsCallerProducer(t *testing.T) {
	doc := NewDocument("Doc", Paragraph("body"))
	doc.XMP = &XMPMetadata{}
	r := doc.NewRenderer()
	r.SetProducer("ACME Renderer 2.0", true)
	if err := doc.Render(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	if got, want := r.GetProducer(), utf8toutf16("ACME Renderer 2.0"); got != want {
		t.Errorf("info /Producer = %q, want the caller's %q", got, want)
	}
	if s := string(r.xmp); !strings.Contains(s, "<pdf:Producer>ACME Renderer 2.0</pdf:Producer>") {
		t.Errorf("XMP packet does not carry the caller producer:\n%s", s)
	}

	// a producer stored as Latin-1 bytes is decoded for the packet
	doc = NewDocument("Doc", Paragraph("body"))
	doc.XMP = &XMPMetadata{}
	r = doc.NewRenderer()
	r.SetProducer("Prod\xfccer", false)
	if err := doc.Render(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	if s := string(r.xmp); !strings.Contains(s, "<pdf:Producer>Prodücer</pdf:Producer>") {
		t.Errorf("XMP packet does not carry the decoded Latin-1 producer:\n%s", s)
	}

	// a Latin-1 producer that happens to start with the UTF-16 BOM bytes
	// "þÿ" must not be misdecoded as UTF-16 (the encoding is declared by the
	// setter, not sniffed)
	doc = NewDocument("Doc", Paragraph("body"))
	doc.XMP = &XMPMetadata{}
	r = doc.NewRenderer()
	r.SetProducer("\xfe\xffLegacy", false)
	if err := doc.Render(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	if s := string(r.xmp); !strings.Contains(s, "<pdf:Producer>þÿLegacy</pdf:Producer>") {
		t.Errorf("XMP packet mangled the fake-BOM Latin-1 producer:\n%s", s)
	}
}

// An explicit XMP.Producer must beat the renderer's /Info producer (including
// the construction-time engine default), in the packet and mirrored to /Info —
// swapping the fallback order would let /Info silently override the packet.
func TestDocumentXMPProducerOverridesRenderer(t *testing.T) {
	doc := NewDocument("Doc", Paragraph("body"))
	doc.XMP = &XMPMetadata{Producer: "XMP Producer"}
	r := doc.NewRenderer()
	r.SetProducer("Renderer Producer", true)
	if err := doc.Render(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	if s := string(r.xmp); !strings.Contains(s, "<pdf:Producer>XMP Producer</pdf:Producer>") {
		t.Errorf("XMP packet producer was overridden by the renderer value:\n%s", s)
	}
	if got, want := r.GetProducer(), utf8toutf16("XMP Producer"); got != want {
		t.Errorf("info /Producer = %q, want the XMP value %q", got, want)
	}
}

// Clearing the renderer producer must not leave the XMP packet without one:
// /Info always carries a producer, so when neither the XMP nor the renderer
// names one the packet falls back to the engine default.
func TestDocumentXMPEngineDefaultProducer(t *testing.T) {
	doc := NewDocument("Doc", Paragraph("body"))
	doc.XMP = &XMPMetadata{}
	r := doc.NewRenderer()
	r.SetProducer("", false) // clear the construction-time default
	if err := doc.Render(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	if s := string(r.xmp); !strings.Contains(s, "<pdf:Producer>FPDF "+cnFpdfVersion+"</pdf:Producer>") {
		t.Errorf("XMP packet does not fall back to the engine producer:\n%s", s)
	}
}

// A FacturX.Prefix that is not a valid XML name would be written verbatim into
// element and attribute names, silently producing a malformed packet; it must
// be a loud error instead.
func TestXMPMetadataXML_facturXInvalidPrefix(t *testing.T) {
	for _, prefix := range []string{"my fx", `p"><x`, "1fx", "fx:x"} {
		m := &XMPMetadata{FacturX: &FacturX{ConformanceLevel: FacturXEN16931, Prefix: prefix}}
		if _, err := m.XML(); err == nil {
			t.Errorf("expected error for invalid FacturX.Prefix %q", prefix)
		}
	}
	// a valid custom prefix still works
	m := &XMPMetadata{FacturX: &FacturX{ConformanceLevel: FacturXEN16931, Prefix: "zf"}}
	if _, err := m.XML(); err != nil {
		t.Errorf("valid custom prefix %q rejected: %v", "zf", err)
	}
}

// Two attachments that share a filename must still produce a valid embedded-
// files name tree: name-tree keys have to be unique, so the duplicate is
// disambiguated rather than emitted twice under the same key.
func TestDocumentEmbeddedFilesUniqueKeys(t *testing.T) {
	doc := NewDocument("Dup", Paragraph("body"))
	doc.Attachments = []Attachment{
		{Content: []byte("first"), Filename: "dup.txt"},
		{Content: []byte("second"), Filename: "dup.txt"},
	}
	r := doc.NewRenderer()
	if err := doc.Render(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := r.Output(&buf); err != nil {
		t.Fatal(err)
	}
	tree := r.getEmbeddedFiles()
	k1 := r.textstring(utf8toutf16("dup.txt"))
	k2 := r.textstring(utf8toutf16("dup.txt (2)"))
	if k1 == k2 {
		t.Fatal("test setup: keys are not distinct")
	}
	if !strings.Contains(tree, k1) {
		t.Errorf("name tree missing first key:\n%s", tree)
	}
	if !strings.Contains(tree, k2) {
		t.Errorf("name tree missing disambiguated key:\n%s", tree)
	}
}

// A non-ASCII attachment filename must be written into the filespec /F as its
// single-byte PDFDoc/Latin-1 form, not as raw UTF-8 bytes; the exact Unicode
// name is carried by /UF.
func TestAttachmentNonASCIIFilespec(t *testing.T) {
	doc := NewDocument("NonASCII", Paragraph("body"))
	doc.Attachments = []Attachment{{Content: []byte("x"), Filename: "faktüra.xml"}}
	r := doc.NewRenderer()
	if err := doc.Render(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := r.Output(&buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "/F (fakt\xfcra.xml)") {
		t.Errorf("/F is not PDFDoc-encoded (want single byte 0xFC for ü)")
	}
	if strings.Contains(out, "fakt\xc3\xbcra.xml") {
		t.Error("/F contains raw UTF-8 bytes for the filename")
	}
}
