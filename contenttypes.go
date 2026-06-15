package mx

import "strings"

// MIME type strings for use as Content-Type header values. Text-based types
// include the "; charset=utf-8" parameter.
const (
	// ContentTypePlainText is the MIME type for UTF-8 plain text.
	ContentTypePlainText = "text/plain; charset=utf-8"
	// ContentTypeMarkdown is the MIME type for UTF-8 Markdown.
	ContentTypeMarkdown = "text/markdown; charset=utf-8"
	// ContentTypeJavaScript is the MIME type for UTF-8 JavaScript.
	ContentTypeJavaScript = "text/javascript; charset=utf-8"
	// ContentTypeHTML is the MIME type for UTF-8 HTML.
	ContentTypeHTML = "text/html; charset=utf-8"
	// ContentTypeCSV is the MIME type for UTF-8 CSV.
	ContentTypeCSV = "text/csv; charset=utf-8"
	// ContentTypeJSON is the MIME type for UTF-8 JSON.
	ContentTypeJSON = "application/json; charset=utf-8"
	// ContentTypeXML is the MIME type for XML.
	ContentTypeXML = "application/xml"
	// ContentTypePDF is the MIME type for PDF documents.
	ContentTypePDF = "application/pdf"
	// ContentTypeZip is the MIME type for ZIP archives.
	ContentTypeZip = "application/zip"
	// ContentTypeOctetStream is the MIME type for arbitrary binary data.
	ContentTypeOctetStream = "application/octet-stream"
	// ContentTypeWWWFormURLEncoded is the MIME type for URL-encoded form data.
	ContentTypeWWWFormURLEncoded = "application/x-www-form-urlencoded"
	// ContentTypeMultipartFormData is the MIME type for multipart form data.
	ContentTypeMultipartFormData = "multipart/form-data"
	// ContentTypePNG is the MIME type for PNG images.
	ContentTypePNG = "image/png"
	// ContentTypeGIF is the MIME type for GIF images.
	ContentTypeGIF = "image/gif"
	// ContentTypeJPEG is the MIME type for JPEG images.
	ContentTypeJPEG = "image/jpeg"
	// ContentTypeTIFF is the MIME type for TIFF images.
	ContentTypeTIFF = "image/tiff"
)

// NormalizeContentTypeWithCharsetUTF8 lower-cases and trims contentType and, for
// the text-based MIME types that should be UTF-8 (text/plain, text/markdown,
// text/javascript, text/html, text/csv and application/json), appends
// "; charset=utf-8" if not already present. Other content types are returned
// trimmed and lower-cased without a charset parameter.
func NormalizeContentTypeWithCharsetUTF8(contentType string) string {
	contentType = strings.ToLower(strings.TrimSpace(contentType))
	switch contentType {
	case "text/plain", "text/markdown", "text/javascript", "text/html", "text/csv", "application/json":
		contentType += "; charset=utf-8"
	}
	return contentType
}
