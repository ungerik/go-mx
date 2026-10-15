package web

import (
	"fmt"
	"io/fs"
	"mime"
	"regexp"
	"strings"
	"time"
)

// langTagRegexp matches the BCP 47 language-tag charset
// [A-Za-z]{1,8}(-[A-Za-z0-9]{1,8})*, the same check html.Document.Lang is
// rendered with. It is a charset check, not a registry lookup: what matters
// here is that a tag cannot break the attribute or Link header it ends up in.
var langTagRegexp = regexp.MustCompile(`^[A-Za-z]{1,8}(-[A-Za-z0-9]{1,8})*$`)

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

	// Lang is the BCP 47 language tag of the page, for example "de". It falls
	// back to Site.Lang when empty. A site in several languages needs it for
	// every page with translations, because the hreflang links of a page have
	// to name the page's own language too (see [Alternate]).
	Lang string

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

	// Alternates are the other representations of the page: translations,
	// other formats like a Markdown version, or a document describing it.
	// [Site.AddPageMetadata] renders them as <link> elements and
	// [Site.PageHeaders] as Link headers.
	Alternates []Alternate

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
//
// It is the one predicate for "public": every surface that publishes pages
// beyond the page itself — the sitemap, llms.txt, the MCP page tools — lists
// only pages it is true for.
func (p *Page) Indexable() bool {
	return !p.NoIndex && p.IsPublished()
}

// MarkdownSource returns the Markdown source of the page and true if the page
// has a Markdown alternate, which is the case if its ContentType is
// text/markdown (with any parameters, like the charset of
// mx.ContentTypeMarkdown) and its Content a string or []byte.
//
// Content is only loaded by a [PageSource] asked for it, so a page iterated
// without content never has a Markdown alternate.
func (p *Page) MarkdownSource() (source []byte, ok bool) {
	mediaType, _, err := mime.ParseMediaType(p.ContentType)
	if err != nil || mediaType != "text/markdown" {
		return nil, false
	}
	switch content := p.Content.(type) {
	case string:
		return []byte(content), true
	case []byte:
		return content, true
	}
	return nil, false
}

// Validate returns an error if the page's Path is set but not a valid path
// within a site, if Lang is set but not a language tag, or if one of its
// Alternates is invalid: malformed, pointing back at the page itself, or
// repeating the rel, type and hreflang of another alternate, which would leave
// crawlers two answers for one question.
//
// A page with translations additionally needs a language, which may come from
// the site; [Site.PageHeaders] and [Site.AddPageMetadata] check that too.
func (p *Page) Validate() error {
	if p.Path != "" {
		if err := validateSitePath(p.Path); err != nil {
			return fmt.Errorf("page %q: %w", p.Title, err)
		}
	}
	if p.Lang != "" {
		if err := validateLang(p.Lang); err != nil {
			return fmt.Errorf("page %q: Lang: %w", p.Path, err)
		}
	}
	seen := make(map[string]bool, len(p.Alternates))
	for i, alternate := range p.Alternates {
		if err := alternate.Validate(); err != nil {
			return fmt.Errorf("page %q: alternate %d: %w", p.Path, i, err)
		}
		// An x-default may well be the page itself — the version shown to a
		// visitor whose language has no translation of its own.
		if alternate.Path == p.Path && alternate.Type == "" && alternate.RelOrDefault() == RelAlternate &&
			!strings.EqualFold(alternate.HrefLang, HrefLangDefault) {
			return fmt.Errorf("page %q: alternate %d points at the page itself", p.Path, i)
		}
		key := alternate.key()
		if seen[key] {
			return fmt.Errorf("page %q: alternate %d repeats the rel, type and hreflang of an earlier one", p.Path, i)
		}
		seen[key] = true
	}
	return nil
}

// validateLang returns an error if lang is not a BCP 47 language tag or is the
// hreflang keyword "x-default", which names no language.
func validateLang(lang string) error {
	if !langTagRegexp.MatchString(lang) {
		return fmt.Errorf("%q is not a valid BCP 47 language tag", lang)
	}
	if strings.EqualFold(lang, HrefLangDefault) {
		return fmt.Errorf("%q is not a language", lang)
	}
	return nil
}
