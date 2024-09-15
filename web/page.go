package web

import (
	"io/fs"
	"time"
)

type Page struct {
	// Route      mx.Route
	PathValues map[string]any

	Title       string
	Author      string
	Type        string
	Tags        []string
	Created     time.Time
	LastUpdated time.Time
	Published   time.Time // Zero time means not published

	Resources   []fs.File // URLs or file paths
	ContentType string
	Content     any // Can be nil if only metadata is needed
}
