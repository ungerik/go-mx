package web

import (
	"fmt"
	"io/fs"
	"mime"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/ungerik/go-mx/html"
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

// LastModified returns when the content of the page last changed: LastUpdated,
// or Published if LastUpdated is zero.
func (p *Page) LastModified() time.Time {
	if p.LastUpdated.IsZero() {
		return p.Published
	}
	return p.LastUpdated
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

// robotsNoIndex is the robots directive a page that is not [Page.Indexable]
// is marked with, in its robots meta tag or X-Robots-Tag header.
const robotsNoIndex = "noindex, nofollow"

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
// Alternates is invalid: malformed, a translation pointing back at the page
// itself, repeating the rel, type and hreflang of another alternate, which
// would leave crawlers two answers for one question, or a text/markdown
// alternate without a language anywhere but at the page's [MarkdownPath],
// the Markdown version llms.txt and the MCP tools link. Another format of the
// page may share its URL, served by content negotiation (see
// [AcceptsMarkdown]).
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
		if alternate.IsTranslation() && !strings.EqualFold(alternate.HrefLang, HrefLangDefault) && sameSitePath(alternate.Path, p.Path) {
			return fmt.Errorf("page %q: alternate %d points at the page itself", p.Path, i)
		}
		key := alternate.key()
		if seen[key] {
			return fmt.Errorf("page %q: alternate %d repeats the rel, type and hreflang of an earlier one", p.Path, i)
		}
		seen[key] = true
		// The page's own Markdown version is the one at its MarkdownPath,
		// which llms.txt and the MCP tools link, or the page URL itself by
		// content negotiation. A link to anywhere else would send crawlers
		// and agents reading the head to another document than the rest.
		if alternate.isOwnMarkdown() && p.Path != "" && !sameSitePath(alternate.Path, p.Path) {
			markdownPath, err := MarkdownPath(p.Path)
			if err != nil {
				return fmt.Errorf("page %q: Markdown alternate %d: %w", p.Path, i, err)
			}
			if !sameSitePath(alternate.Path, markdownPath) {
				return fmt.Errorf("page %q: Markdown alternate %d is at %q, not at the page's Markdown path %q", p.Path, i, alternate.Path, markdownPath)
			}
		}
	}
	return nil
}

// sameSitePath reports whether two paths that passed validateSitePath are the
// same URL path once escaped the way [Site.URL] escapes them, so that "/café",
// "/caf%C3%A9" and "/caf%c3%a9", which a string comparison tells apart, count
// as one. Percent escapes are case-insensitive (RFC 3986, section 6.2.2.1),
// but EscapedPath keeps the case it was given.
func sameSitePath(a, b string) bool {
	if a == b {
		return true
	}
	// Escaping is one to one for a path without escapes of its own, so two
	// different ones can only meet if one of them holds a '%'.
	if !strings.Contains(a, "%") && !strings.Contains(b, "%") {
		return false
	}
	ua, errA := url.Parse(a)
	ub, errB := url.Parse(b)
	return errA == nil && errB == nil && upperEscapes(ua.EscapedPath()) == upperEscapes(ub.EscapedPath())
}

// percentEscapeRegexp matches one percent escape.
var percentEscapeRegexp = regexp.MustCompile(`%[0-9a-fA-F]{2}`)

// upperEscapes returns path with the hex digits of its percent escapes in
// upper case.
func upperEscapes(path string) string {
	return percentEscapeRegexp.ReplaceAllStringFunc(path, strings.ToUpper)
}

// validateLang returns an error if lang is not a BCP 47 language tag or is the
// hreflang keyword "x-default", which names no language.
func validateLang(lang string) error {
	if !html.ValidLangTag(lang) {
		return fmt.Errorf("%q is not a valid BCP 47 language tag", lang)
	}
	if strings.EqualFold(lang, HrefLangDefault) {
		return fmt.Errorf("%q is not a language", lang)
	}
	return nil
}
