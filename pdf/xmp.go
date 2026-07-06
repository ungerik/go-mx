package pdf

import (
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/xml"
)

// XMP namespace URIs of the schemas written by [XMPMetadata.XML].
const (
	xmpNSMeta          = "adobe:ns:meta/"
	xmpNSRDF           = "http://www.w3.org/1999/02/22-rdf-syntax-ns#"
	xmpNSDC            = "http://purl.org/dc/elements/1.1/"
	xmpNSPDF           = "http://ns.adobe.com/pdf/1.3/"
	xmpNSXMP           = "http://ns.adobe.com/xap/1.0/"
	xmpNSPDFAID        = "http://www.aiim.org/pdfa/ns/id/"
	xmpNSPDFAExtension = "http://www.aiim.org/pdfa/ns/extension/"
	xmpNSPDFASchema    = "http://www.aiim.org/pdfa/ns/schema#"
	xmpNSPDFAProperty  = "http://www.aiim.org/pdfa/ns/property#"
)

// FacturXNamespaceURI is the XMP namespace of the Factur-X / ZUGFeRD 2.x
// PDF/A extension schema.
const FacturXNamespaceURI = "urn:factur-x:pdfa:CrossIndustryDocument:invoice:1p0#"

// Factur-X / ZUGFeRD conformance levels for [FacturX.ConformanceLevel],
// naming the profile of the embedded invoice XML.
const (
	FacturXMinimum   = "MINIMUM"
	FacturXBasicWL   = "BASIC WL"
	FacturXBasic     = "BASIC"
	FacturXEN16931   = "EN 16931"
	FacturXExtended  = "EXTENDED"
	FacturXXRechnung = "XRECHNUNG"
)

// XMPMetadata describes a document-level XMP metadata packet. [XMPMetadata.XML]
// builds the packet, covering the Dublin Core, basic XMP, Adobe PDF and PDF/A
// identification schemas plus the Factur-X extension schema for ZUGFeRD/Factur-X
// hybrid invoices. Empty fields are omitted from the packet.
//
// The typed [Document.XMP] field is the usual way to use it: the document
// fills empty fields from its own metadata and keeps the PDF info dictionary
// consistent with the packet, as PDF/A validators require. For a raw
// [Renderer], pass the result of [XMPMetadata.XML] to
// [Renderer.SetXmpMetadata] and keep the info dictionary consistent yourself.
type XMPMetadata struct {
	Title       string    // dc:title
	Author      string    // dc:creator
	Subject     string    // dc:description
	Keywords    string    // pdf:Keywords
	CreatorTool string    // xmp:CreatorTool, the info dictionary's /Creator
	Producer    string    // pdf:Producer
	CreateDate  time.Time // xmp:CreateDate
	ModifyDate  time.Time // xmp:ModifyDate

	// PDFAPart and PDFAConformance identify the PDF/A profile the document
	// claims (pdfaid schema), e.g. part 3 conformance "B" for the PDF/A-3b
	// required by ZUGFeRD/Factur-X. A zero PDFAPart omits the identification;
	// an empty PDFAConformance defaults to "B".
	PDFAPart        int
	PDFAConformance string

	// FacturX, when non-nil, declares an embedded ZUGFeRD/Factur-X invoice.
	FacturX *FacturX
}

// FacturX declares the embedded XML invoice of a ZUGFeRD/Factur-X hybrid
// document in XMP: the PDF/A extension schema description that PDF/A requires
// for custom properties, and the properties themselves. The XML file must
// also be embedded as an [Attachment] whose Filename matches DocumentFileName
// and whose Relationship suits the profile ([AFRelationshipAlternative], or
// [AFRelationshipData] for MINIMUM and BASIC WL).
type FacturX struct {
	// DocumentType of the embedded XML, "INVOICE" (default) or "ORDER".
	DocumentType string
	// DocumentFileName is the name of the embedded XML attachment,
	// defaulting to "factur-x.xml".
	DocumentFileName string
	// Version of the Factur-X standard, defaulting to "1.0".
	Version string
	// ConformanceLevel is the profile of the embedded XML, one of the
	// FacturX… constants like [FacturXEN16931]. It must be set explicitly
	// because it has to match the embedded XML, which cannot be guessed.
	ConformanceLevel string
	// NamespaceURI of the extension schema, defaulting to
	// [FacturXNamespaceURI]. Order-X and other derivations use their own.
	NamespaceURI string
	// Prefix of the extension schema properties, defaulting to "fx".
	Prefix string
}

// withDefaults returns f with empty fields defaulted, or an error if a field
// without a sensible default is empty.
func (f *FacturX) withDefaults() (FacturX, error) {
	out := *f
	out.DocumentType = cmp.Or(out.DocumentType, "INVOICE")
	out.DocumentFileName = cmp.Or(out.DocumentFileName, "factur-x.xml")
	out.Version = cmp.Or(out.Version, "1.0")
	out.NamespaceURI = cmp.Or(out.NamespaceURI, FacturXNamespaceURI)
	out.Prefix = cmp.Or(out.Prefix, "fx")
	// The prefix is written into element and attribute names unescaped, so an
	// invalid one would silently produce a malformed packet.
	if !validXMLPrefix(out.Prefix) {
		return out, fmt.Errorf("pdf.FacturX.Prefix %q is not a valid XML namespace prefix", out.Prefix)
	}
	if out.ConformanceLevel == "" {
		return out, errors.New("pdf.FacturX.ConformanceLevel must be set to the profile of the embedded XML")
	}
	return out, nil
}

// validXMLPrefix reports whether s is a valid XML namespace prefix, an ASCII
// NCName: a letter or underscore, followed by letters, digits, '-', '_' or '.'
// (no colon). Prefixes are conventionally ASCII, so this deliberately rejects
// the wider Unicode NCName range rather than risk emitting a broken name.
func validXMLPrefix(s string) bool {
	if s == "" {
		return false
	}
	for i := range len(s) {
		c := s[i]
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c == '_':
			// valid as the first or a following character
		case i > 0 && (c >= '0' && c <= '9' || c == '-' || c == '.'):
			// valid only after the first character
		default:
			return false
		}
	}
	return true
}

// XML builds the XMP packet: an xpacket header, the metadata as RDF/XML,
// padding for in-place updates and the xpacket trailer. The result is ready
// for [Renderer.SetXmpMetadata].
//
// Characters that are not valid in XML 1.0 (C0 controls other than tab, LF,
// CR) are replaced with U+FFFD in the text fields, so a stray control byte
// in a metadata value cannot invalidate the whole packet.
func (m *XMPMetadata) XML() ([]byte, error) {
	san := *m
	san.Title = xmpText(m.Title)
	san.Author = xmpText(m.Author)
	san.Subject = xmpText(m.Subject)
	san.Keywords = xmpText(m.Keywords)
	san.CreatorTool = xmpText(m.CreatorTool)
	san.Producer = xmpText(m.Producer)
	san.PDFAConformance = xmpText(m.PDFAConformance)
	m = &san

	var descriptions []any

	if m.PDFAPart != 0 {
		descriptions = append(descriptions, rdfDescription("pdfaid", xmpNSPDFAID,
			xml.ElementNS("pdfaid", "part", strconv.Itoa(m.PDFAPart)),
			xml.ElementNS("pdfaid", "conformance", cmp.Or(m.PDFAConformance, "B")),
		))
	}

	var dc []any
	if m.Title != "" {
		dc = append(dc, xml.ElementNS("dc", "title", xmpLangAlt(m.Title)))
	}
	if m.Author != "" {
		dc = append(dc, xml.ElementNS("dc", "creator",
			xml.ElementNS("rdf", "Seq", xml.ElementNS("rdf", "li", m.Author))))
	}
	if m.Subject != "" {
		dc = append(dc, xml.ElementNS("dc", "description", xmpLangAlt(m.Subject)))
	}
	if len(dc) > 0 {
		descriptions = append(descriptions, rdfDescription("dc", xmpNSDC, dc...))
	}

	var pdfProps []any
	if m.Producer != "" {
		pdfProps = append(pdfProps, xml.ElementNS("pdf", "Producer", m.Producer))
	}
	if m.Keywords != "" {
		pdfProps = append(pdfProps, xml.ElementNS("pdf", "Keywords", m.Keywords))
	}
	if len(pdfProps) > 0 {
		descriptions = append(descriptions, rdfDescription("pdf", xmpNSPDF, pdfProps...))
	}

	var xmpProps []any
	if m.CreatorTool != "" {
		xmpProps = append(xmpProps, xml.ElementNS("xmp", "CreatorTool", m.CreatorTool))
	}
	if !m.CreateDate.IsZero() {
		xmpProps = append(xmpProps, xml.ElementNS("xmp", "CreateDate", xmpDate(m.CreateDate)))
	}
	if !m.ModifyDate.IsZero() {
		xmpProps = append(xmpProps, xml.ElementNS("xmp", "ModifyDate", xmpDate(m.ModifyDate)))
	}
	if len(xmpProps) > 0 {
		descriptions = append(descriptions, rdfDescription("xmp", xmpNSXMP, xmpProps...))
	}

	if m.FacturX != nil {
		f, err := m.FacturX.withDefaults()
		if err != nil {
			return nil, err
		}
		descriptions = append(descriptions,
			facturXExtensionSchema(f),
			rdfDescription(f.Prefix, f.NamespaceURI,
				xml.ElementNS(f.Prefix, "DocumentType", f.DocumentType),
				xml.ElementNS(f.Prefix, "DocumentFileName", f.DocumentFileName),
				xml.ElementNS(f.Prefix, "Version", f.Version),
				xml.ElementNS(f.Prefix, "ConformanceLevel", f.ConformanceLevel),
			))
	}

	packet := mx.Components{
		xml.ProcInst{Target: "xpacket", Data: "begin=\"\uFEFF\" id=\"W5M0MpCehiHzreSzNTczkc9d\""},
		xml.ElementNS("x", "xmpmeta",
			xml.XMLNSPrefix("x", xmpNSMeta),
			xml.ElementNS("rdf", "RDF",
				append([]any{xml.XMLNSPrefix("rdf", xmpNSRDF)}, descriptions...)...,
			),
		),
	}
	s, err := xml.String(packet)
	if err != nil {
		return nil, err
	}

	// ~2KB of trailing padding so tools can update the packet in place
	// (XMP Specification Part 1, 7.3.2), then the read-write xpacket trailer.
	const paddingLine = "                                                               \n"
	var b bytes.Buffer
	b.Grow(len(s) + 32*len(paddingLine) + 32)
	b.WriteString(s)
	b.WriteByte('\n')
	for range 32 {
		b.WriteString(paddingLine)
	}
	b.WriteString(`<?xpacket end="w"?>`)
	return b.Bytes(), nil
}

// xmpText returns s with characters that are not valid in XML 1.0 replaced
// by U+FFFD. The XML writer escapes markup characters but writes control
// characters through, and XML 1.0 forbids the C0 range except tab, LF and
// CR — one stray control byte in a metadata value (e.g. a producer read
// back from the info dictionary) would otherwise invalidate the packet.
func xmpText(s string) string {
	if !strings.ContainsFunc(s, invalidXMLChar) {
		return s
	}
	return strings.Map(func(r rune) rune {
		if invalidXMLChar(r) {
			return '�'
		}
		return r
	}, s)
}

func invalidXMLChar(r rune) bool {
	return (r < 0x20 && r != '\t' && r != '\n' && r != '\r') ||
		r == 0xFFFE || r == 0xFFFF
}

// rdfDescription builds an rdf:Description block that binds prefix to
// namespaceURI and holds the given property elements.
func rdfDescription(prefix, namespaceURI string, children ...any) *mx.Element {
	args := append([]any{
		xml.AttribNS("rdf", "about", ""),
		xml.XMLNSPrefix(prefix, namespaceURI),
	}, children...)
	return xml.ElementNS("rdf", "Description", args...)
}

// xmpLangAlt wraps text as the language-alternative array XMP requires for
// dc:title and dc:description.
func xmpLangAlt(text string) *mx.Element {
	return xml.ElementNS("rdf", "Alt",
		xml.ElementNS("rdf", "li", xml.XMLLang("x-default"), text))
}

// xmpDate formats tm as an XMP (ISO 8601) date with an explicit UTC offset,
// so it fully specifies the instant like the PDF info dictionary dates
// written alongside XMP metadata.
func xmpDate(tm time.Time) string {
	return tm.Format("2006-01-02T15:04:05-07:00")
}

// facturXExtensionSchema builds the rdf:Description declaring the Factur-X
// extension schema, which PDF/A requires for properties outside the
// predefined XMP schemas.
func facturXExtensionSchema(f FacturX) *mx.Element {
	property := func(name, description string) *mx.Element {
		return xml.ElementNS("rdf", "li",
			xml.AttribNS("rdf", "parseType", "Resource"),
			xml.ElementNS("pdfaProperty", "name", name),
			xml.ElementNS("pdfaProperty", "valueType", "Text"),
			xml.ElementNS("pdfaProperty", "category", "external"),
			xml.ElementNS("pdfaProperty", "description", description),
		)
	}
	return xml.ElementNS("rdf", "Description",
		xml.AttribNS("rdf", "about", ""),
		xml.XMLNSPrefix("pdfaExtension", xmpNSPDFAExtension),
		xml.XMLNSPrefix("pdfaSchema", xmpNSPDFASchema),
		xml.XMLNSPrefix("pdfaProperty", xmpNSPDFAProperty),
		xml.ElementNS("pdfaExtension", "schemas",
			xml.ElementNS("rdf", "Bag",
				xml.ElementNS("rdf", "li",
					xml.AttribNS("rdf", "parseType", "Resource"),
					xml.ElementNS("pdfaSchema", "schema", "Factur-X PDFA Extension Schema"),
					xml.ElementNS("pdfaSchema", "namespaceURI", f.NamespaceURI),
					xml.ElementNS("pdfaSchema", "prefix", f.Prefix),
					xml.ElementNS("pdfaSchema", "property",
						xml.ElementNS("rdf", "Seq",
							property("DocumentFileName", "The name of the embedded XML document"),
							property("DocumentType", "The type of the hybrid document in capital letters, e.g. INVOICE or ORDER"),
							property("Version", "The actual version of the standard applying to the embedded XML document"),
							property("ConformanceLevel", "The conformance level of the embedded XML document"),
						),
					),
				),
			),
		),
	)
}
