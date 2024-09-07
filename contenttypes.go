package mx

import "strings"

const (
	ContentTypePlainText         = "text/plain; charset=utf-8"
	ContentTypeMarkdown          = "text/markdown; charset=utf-8"
	ContentTypeJavaScript        = "text/javascript; charset=utf-8"
	ContentTypeHTML              = "text/html; charset=utf-8"
	ContentTypeCSV               = "text/csv; charset=utf-8"
	ContentTypeJSON              = "application/json; charset=utf-8"
	ContentTypeXML               = "application/xml"
	ContentTypePDF               = "application/pdf"
	ContentTypeZip               = "application/zip"
	ContentTypeOctetStream       = "application/octet-stream"
	ContentTypeWWWFormURLEncoded = "application/x-www-form-urlencoded"
	ContentTypeMultipartFormData = "multipart/form-data"
	ContentTypePNG               = "image/png"
	ContentTypeGIF               = "image/gif"
	ContentTypeJPEG              = "image/jpeg"
	ContentTypeTIFF              = "image/tiff"
)

func NormalizeContentTypeWithCharsetUTF8(contentType string) string {
	contentType = strings.ToLower(strings.TrimSpace(contentType))
	switch contentType {
	case "text/plain", "text/markdown", "text/javascript", "text/html", "text/csv", "application/json":
		contentType += "; charset=utf-8"
	}
	return contentType
}
