package web

import (
	"context"
	"time"

	"github.com/ungerik/go-fs"
	"github.com/ungerik/go-mx"
)

// https://github.com/adrg/frontmatter
// https://github.com/yuin/goldmark

type Page struct {
	Title       string
	Author      string
	Type        string
	Tags        []string
	Created     time.Time
	LastUpdated time.Time
	Published   time.Time       // Zero time means not published
	Resources   []fs.FileReader // URLs or file paths
	ContentType string
	Content     []byte
}

type PageVisualizer interface {
	VisualizePage(ctx context.Context, page *Page) (mx.Component, error)
}

type ContentVisualizer interface {
	VisualizeContent(ctx context.Context, contentType string, content []byte) (mx.Component, error)
}
