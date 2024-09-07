package web

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

type Site struct {
	PageTemplate PageTemplate
	NotFoundPage html.Document
	Routes       []mx.Route
}
