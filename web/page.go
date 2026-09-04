package web

import (
	"io/fs"
	"time"
)

// Page holds the metadata and content of a single web page,
// independent of how it is sourced or rendered.
type Page struct {
	// Route      mx.Route
	PathValues map[string]any

	// Path is the absolute URL path of the page within the site, like
	// "/blog/hello-world". A [Site] needs it to build the page's canonical
	// URL: without it the page renders without a canonical link, and an
	// [Page.Indexable] page without it fails the sitemap it should be part of.
	Path string

	Title string
	// Description is the summary shown by search engines and link previews.
	// It falls back to Site.Description when empty.
	Description string
	Author      string
	Type        string
	Tags        []string
	NoIndex     bool // <meta name="robots" content="noindex, nofollow" />
	Created     time.Time
	LastUpdated time.Time
	Published   time.Time // Zero time means not published

	Resources   []fs.File // URLs or file paths
	ContentType string
	Content     any // Can be nil if only metadata is needed
}

// IsPublished reports whether the page is published at the current time, which
// is the case if Published is neither zero (a draft) nor in the future (a
// scheduled page).
func (p *Page) IsPublished() bool {
	return !p.Published.IsZero() && !p.Published.After(time.Now())
}

// Indexable reports whether the page should be offered to search engines: it
// has to be published and must not be marked NoIndex. It decides both whether
// [Site] lists the page in its sitemap and whether it renders a "noindex"
// robots meta tag for the page.
func (p *Page) Indexable() bool {
	return !p.NoIndex && p.IsPublished()
}
