//go:generate go -C ../tools tool go-enum ../web/$GOFILE

package web

// https://developers.google.com/search/docs/crawling-indexing/sitemaps/build-sitemap
// https://www.sitemaps.org/protocol.html

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/xml"
)

const (
	// DefaultSitemapPath is the conventional path of a site's sitemap. Unlike
	// [RobotsTxtPath] the path is not mandated, but it is not free either: a
	// sitemap only covers URLs at or below its own directory unless it was
	// submitted through a search console, so a sitemap under "/private/"
	// cannot list the whole site. Keep it at the root of the site (see
	// [Site.SitemapPath]).
	DefaultSitemapPath = "/sitemap.xml"

	// SitemapNamespace is the XML namespace of the sitemap protocol 0.9, the
	// current version.
	SitemapNamespace = "http://www.sitemaps.org/schemas/sitemap/0.9"

	// MaxSitemapURLs is the maximum number of URLs one sitemap file may
	// contain. Split a bigger site into several sitemaps and reference them
	// from a [SitemapIndex].
	MaxSitemapURLs = 50000

	// MaxSitemapLocLength is the maximum length of a sitemap <loc> URL.
	MaxSitemapLocLength = 2048

	// MaxSitemapBytes is the maximum uncompressed size of one sitemap file.
	// It is the limit that bites first for a site with long URLs: 50000 URLs
	// of maximum length would be twice this size.
	MaxSitemapBytes = 50 << 20
)

// ChangeFreq tells crawlers how often the content of a page is expected to
// change. It is a hint that search engines are free to ignore — Google does —
// so leaving it empty (which omits the <changefreq> element) is a reasonable
// default.
type ChangeFreq string //#enum

const (
	// ChangeFreqAlways describes a page that changes on every access.
	ChangeFreqAlways ChangeFreq = "always"
	// ChangeFreqHourly describes a page that changes hourly.
	ChangeFreqHourly ChangeFreq = "hourly"
	// ChangeFreqDaily describes a page that changes daily.
	ChangeFreqDaily ChangeFreq = "daily"
	// ChangeFreqWeekly describes a page that changes weekly.
	ChangeFreqWeekly ChangeFreq = "weekly"
	// ChangeFreqMonthly describes a page that changes monthly.
	ChangeFreqMonthly ChangeFreq = "monthly"
	// ChangeFreqYearly describes a page that changes yearly.
	ChangeFreqYearly ChangeFreq = "yearly"
	// ChangeFreqNever describes an archived page that will not change again.
	ChangeFreqNever ChangeFreq = "never"
)

// Valid indicates if c is any of the valid values for ChangeFreq
func (c ChangeFreq) Valid() bool {
	switch c {
	case
		ChangeFreqAlways,
		ChangeFreqHourly,
		ChangeFreqDaily,
		ChangeFreqWeekly,
		ChangeFreqMonthly,
		ChangeFreqYearly,
		ChangeFreqNever:
		return true
	}
	return false
}

// Validate returns an error if c is none of the valid values for ChangeFreq
func (c ChangeFreq) Validate() error {
	if !c.Valid() {
		return fmt.Errorf("invalid value %#v for type web.ChangeFreq", c)
	}
	return nil
}

// Enums returns all valid values for ChangeFreq
func (ChangeFreq) Enums() []ChangeFreq {
	return []ChangeFreq{
		ChangeFreqAlways,
		ChangeFreqHourly,
		ChangeFreqDaily,
		ChangeFreqWeekly,
		ChangeFreqMonthly,
		ChangeFreqYearly,
		ChangeFreqNever,
	}
}

// EnumStrings returns all valid values for ChangeFreq as strings
func (ChangeFreq) EnumStrings() []string {
	return []string{
		"always",
		"hourly",
		"daily",
		"weekly",
		"monthly",
		"yearly",
		"never",
	}
}

// String implements the fmt.Stringer interface for ChangeFreq
func (c ChangeFreq) String() string {
	return string(c)
}

// SitemapURL is one <url> entry of a [Sitemap].
type SitemapURL struct {
	// Loc is the absolute URL of the page, including the scheme and host.
	// It is the only required field.
	Loc string

	// LastMod is the time the page was last modified, rendered as an RFC 3339
	// timestamp. The zero time omits the element. It is the one entry field
	// search engines actually use, so it should only be set when it reflects a
	// real content change — a timestamp that moves on every deploy teaches
	// crawlers to ignore it.
	LastMod time.Time

	// ChangeFreq is an optional hint about how often the page changes.
	// The empty value omits the element.
	ChangeFreq ChangeFreq

	// Priority is the optional priority of this page relative to the other
	// pages of the same site, from 0.1 (lowest) to 1.0 (highest). Zero omits
	// the element, which means the protocol default of 0.5. It has no effect
	// on the ranking between sites and is ignored by Google.
	Priority float64
}

// Validate returns an error if Loc is not an absolute http or https URL, or if
// ChangeFreq or Priority are outside their allowed value ranges.
func (u SitemapURL) Validate() error {
	if err := validateSitemapLoc(u.Loc); err != nil {
		return err
	}
	if u.ChangeFreq != "" {
		if err := u.ChangeFreq.Validate(); err != nil {
			return err
		}
	}
	// NaN has to be excluded explicitly: every comparison against it is false,
	// so a range check alone would let it through and render as "NaN.0".
	if math.IsNaN(u.Priority) || u.Priority < 0 || u.Priority > 1 {
		return fmt.Errorf("priority %v is outside the range 0.0 to 1.0", u.Priority)
	}
	return nil
}

// element builds the <url> element of a validated SitemapURL. The child order
// follows the sitemap schema, which declares them as a sequence.
func (u SitemapURL) element() *mx.Element {
	children := mx.Components{xml.Element("loc", u.Loc)}
	if !u.LastMod.IsZero() {
		children = append(children, xml.Element("lastmod", u.LastMod.Format(time.RFC3339)))
	}
	if u.ChangeFreq != "" {
		children = append(children, xml.Element("changefreq", string(u.ChangeFreq)))
	}
	if u.Priority != 0 {
		children = append(children, xml.Element("priority", formatSitemapPriority(u.Priority)))
	}
	return xml.Element("url", children)
}

// Sitemap is a sitemap XML file listing the URLs of a site that should be
// crawled. It is a [mx.Component] rendering the complete XML document, so it
// can be written to a file with [xml.String] for a statically generated site or
// served with [Sitemap.HandleHTTP].
//
// A sitemap does not decide what gets indexed, it only helps crawlers discover
// pages they might otherwise miss, and tells them via <lastmod> what is worth
// re-fetching. Build one from the pages of a [Site] with [Site.Sitemap].
type Sitemap struct {
	// URLs are the sitemap entries, rendered in the given order. There may be
	// at most [MaxSitemapURLs] of them.
	URLs []SitemapURL
}

// Validate returns an error if the sitemap holds more than [MaxSitemapURLs]
// entries or if any entry is invalid.
func (s *Sitemap) Validate() error {
	if len(s.URLs) > MaxSitemapURLs {
		return fmt.Errorf("sitemap has %d URLs, more than the maximum of %d", len(s.URLs), MaxSitemapURLs)
	}
	size := sitemapURLSetBytes
	for i, u := range s.URLs {
		if err := u.Validate(); err != nil {
			return fmt.Errorf("sitemap URL %d: %w", i, err)
		}
		size += u.renderedBytes()
		if size > MaxSitemapBytes {
			return fmt.Errorf("sitemap is larger than the maximum of %d bytes", MaxSitemapBytes)
		}
	}
	return nil
}

// Document returns the sitemap as an XML document. An invalid sitemap yields a
// document whose root defers the error from [Sitemap.Validate] to render time
// (see [mx.NewErrElement]), so an invalid value can never render as markup.
// Such a document carries no XML declaration either, so rendering it writes
// nothing at all before failing.
func (s *Sitemap) Document() *xml.Document {
	if err := s.Validate(); err != nil {
		return &xml.Document{Root: mx.NewErrElement(err)}
	}
	children := make(mx.Components, len(s.URLs))
	for i, u := range s.URLs {
		children[i] = u.element()
	}
	return xml.NewDocument(xml.Element("urlset", xml.XMLNS(SitemapNamespace), children))
}

// Render writes the sitemap XML document, implementing [mx.Component]. It
// validates before writing anything, so a render straight into a file or a
// socket cannot leave a truncated document behind.
func (s *Sitemap) Render(ctx context.Context, w mx.Writer) error {
	if err := s.Validate(); err != nil {
		return err
	}
	return s.Document().Render(ctx, w)
}

// HandleHTTP renders the sitemap and writes it as an indented XML response,
// implementing http.HandlerFunc.
func (s *Sitemap) HandleHTTP(response http.ResponseWriter, request *http.Request) {
	s.Document().HandleHTTP(response, request)
}

// SitemapIndexEntry is one <sitemap> entry of a [SitemapIndex].
type SitemapIndexEntry struct {
	// Loc is the absolute URL of the sitemap file.
	Loc string
	// LastMod is the time that sitemap file was last modified, rendered as an
	// RFC 3339 timestamp. The zero time omits the element.
	LastMod time.Time
}

// Validate returns an error if Loc is not an absolute http or https URL.
func (e SitemapIndexEntry) Validate() error {
	return validateSitemapLoc(e.Loc)
}

// element builds the <sitemap> element of a validated SitemapIndexEntry.
func (e SitemapIndexEntry) element() *mx.Element {
	children := mx.Components{xml.Element("loc", e.Loc)}
	if !e.LastMod.IsZero() {
		children = append(children, xml.Element("lastmod", e.LastMod.Format(time.RFC3339)))
	}
	return xml.Element("sitemap", children)
}

// SitemapIndex is a sitemap index file listing the URLs of several sitemap
// files. It is needed for sites with more than [MaxSitemapURLs] pages and
// useful to split a site into separately updated sitemaps, for example one per
// section. Like [Sitemap] it is a [mx.Component] rendering the complete XML
// document.
type SitemapIndex struct {
	// Sitemaps are the index entries, rendered in the given order. There may
	// be at most [MaxSitemapURLs] of them.
	Sitemaps []SitemapIndexEntry
}

// Validate returns an error if the index holds more than [MaxSitemapURLs]
// entries or if any entry is invalid.
func (x *SitemapIndex) Validate() error {
	if len(x.Sitemaps) > MaxSitemapURLs {
		return fmt.Errorf("sitemap index has %d sitemaps, more than the maximum of %d", len(x.Sitemaps), MaxSitemapURLs)
	}
	size := sitemapIndexBytes
	for i, e := range x.Sitemaps {
		if err := e.Validate(); err != nil {
			return fmt.Errorf("sitemap index entry %d: %w", i, err)
		}
		size += e.renderedBytes()
		if size > MaxSitemapBytes {
			return fmt.Errorf("sitemap index is larger than the maximum of %d bytes", MaxSitemapBytes)
		}
	}
	return nil
}

// Document returns the sitemap index as an XML document, deferring a
// validation error to render time like [Sitemap.Document].
func (x *SitemapIndex) Document() *xml.Document {
	if err := x.Validate(); err != nil {
		return &xml.Document{Root: mx.NewErrElement(err)}
	}
	children := make(mx.Components, len(x.Sitemaps))
	for i, e := range x.Sitemaps {
		children[i] = e.element()
	}
	return xml.NewDocument(xml.Element("sitemapindex", xml.XMLNS(SitemapNamespace), children))
}

// Render writes the sitemap index XML document, implementing [mx.Component].
// Like [Sitemap.Render] it validates before writing anything.
func (x *SitemapIndex) Render(ctx context.Context, w mx.Writer) error {
	if err := x.Validate(); err != nil {
		return err
	}
	return x.Document().Render(ctx, w)
}

// HandleHTTP renders the sitemap index and writes it as an indented XML
// response, implementing http.HandlerFunc.
func (x *SitemapIndex) HandleHTTP(response http.ResponseWriter, request *http.Request) {
	x.Document().HandleHTTP(response, request)
}

const (
	// sitemapDeclarationBytes is the XML declaration and the line break the
	// writer puts after it.
	sitemapDeclarationBytes = len(`<?xml version="1.0" encoding="UTF-8"?>`) + 1

	// sitemapURLSetBytes is the rendered size of an empty <urlset> document,
	// sitemapIndexBytes the one of an empty <sitemapindex> document.
	sitemapURLSetBytes = sitemapDeclarationBytes + len(`<urlset xmlns="`+SitemapNamespace+`"></urlset>`)
	sitemapIndexBytes  = sitemapDeclarationBytes + len(`<sitemapindex xmlns="`+SitemapNamespace+`"></sitemapindex>`)

	// sitemapIndentBytesPerElement is what one element costs on top of its
	// compact markup when the document is written indented, the form
	// [Sitemap.HandleHTTP] serves: a line break plus up to six spaces. Counting
	// it makes the size an upper bound for both the compact and the indented
	// form, so the estimate can only be too careful, never too permissive.
	sitemapIndentBytesPerElement = 7
)

// renderedBytes returns the number of bytes the entry adds to a sitemap: the
// markup of the elements that are actually rendered plus their escaped
// content, so [Sitemap.Validate] can enforce [MaxSitemapBytes] without
// rendering the document. The URL is measured escaped because that is what
// ends up in the file — an "&" in a query costs five bytes, not one, and
// counting it as one would put the estimate below the real size and let an
// oversized sitemap through.
func (u SitemapURL) renderedBytes() int {
	size := len(`<url><loc></loc></url>`) + len(xml.Escape(u.Loc)) + 2*sitemapIndentBytesPerElement
	if !u.LastMod.IsZero() {
		size += len(`<lastmod></lastmod>`) + len(time.RFC3339) + sitemapIndentBytesPerElement
	}
	if u.ChangeFreq != "" {
		size += len(`<changefreq></changefreq>`) + len(u.ChangeFreq) + sitemapIndentBytesPerElement
	}
	if u.Priority != 0 {
		size += len(`<priority></priority>`) + len(formatSitemapPriority(u.Priority)) + sitemapIndentBytesPerElement
	}
	return size
}

// renderedBytes returns the number of bytes the entry adds to a sitemap index,
// counted like [SitemapURL.renderedBytes].
func (e SitemapIndexEntry) renderedBytes() int {
	size := len(`<sitemap><loc></loc></sitemap>`) + len(xml.Escape(e.Loc)) + 2*sitemapIndentBytesPerElement
	if !e.LastMod.IsZero() {
		size += len(`<lastmod></lastmod>`) + len(time.RFC3339) + sitemapIndentBytesPerElement
	}
	return size
}

// validateSitemapLoc returns an error if loc is not an absolute http or https
// URL, or if it is longer than a sitemap <loc> may be. It is the shared
// requirement of a [SitemapURL] and a [SitemapIndexEntry].
func validateSitemapLoc(loc string) error {
	if err := validateAbsoluteURL(loc); err != nil {
		return err
	}
	if len(loc) > MaxSitemapLocLength {
		return fmt.Errorf("URL %q is longer than the maximum of %d characters", loc, MaxSitemapLocLength)
	}
	return nil
}

// formatSitemapPriority formats a priority as a decimal with at least one
// fractional digit, the notation used throughout the sitemap protocol.
func formatSitemapPriority(priority float64) string {
	s := strconv.FormatFloat(priority, 'f', -1, 64)
	if !strings.Contains(s, ".") {
		s += ".0"
	}
	return s
}
