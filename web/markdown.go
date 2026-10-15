package web

// https://llmstxt.org/#format

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"net/http"
	gopath "path"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/domonda/go-errs"
	"github.com/ungerik/go-mx"
)

// MarkdownPath returns the path of the Markdown alternate of the page at
// pagePath: ".md" appended to the path, or "index.md" for a path ending in a
// slash, so "/en/faq" becomes "/en/faq.md", "/en/" becomes "/en/index.md" and
// "/" becomes "/index.md".
//
// It is an error for pagePath to already end in ".md", which would make the
// page and its alternate the same URL, to end in "/index", whose alternate
// would be that of the directory page ("/en/index" and "/en/" both mapping to
// "/en/index.md"), or not to be a valid path within a site.
func MarkdownPath(pagePath string) (string, error) {
	if err := validateSitePath(pagePath); err != nil {
		return "", err
	}
	if strings.HasSuffix(strings.ToLower(pagePath), ".md") {
		return "", fmt.Errorf("path %q already ends in .md", pagePath)
	}
	if strings.HasSuffix(pagePath, "/") {
		return pagePath + "index.md", nil
	}
	if gopath.Base(pagePath) == "index" {
		return "", fmt.Errorf("path %q ends in /index, whose Markdown path would be that of %q", pagePath, gopath.Dir(pagePath))
	}
	return pagePath + ".md", nil
}

// Translation is a version of a page in another language.
type Translation struct {
	// Lang is the BCP 47 language tag of the translation.
	Lang string
	// URL is the absolute URL of the translation.
	URL string
}

// MarkdownSection is a section a caller adds to a [MarkdownDocument], rendered
// as a "## Title" heading followed by the Body Markdown.
type MarkdownSection struct {
	// Title is the heading of the section. Line breaks are rendered as
	// spaces, because a heading ends at the end of its line. An empty Title
	// renders the Body without a heading.
	Title string
	// Body is the Markdown content of the section, rendered verbatim.
	Body string
}

// MarkdownDocument is the Markdown alternate of a page: the page's own
// Markdown source, preceded by a header that makes the document make sense on
// its own — for a reader that fetched it without the HTML page around it, like
// a crawler or an AI agent:
//
//	# Title
//
//	> Description
//
//	- Canonical URL: https://example.com/en/faq
//	- Language: en
//	- Last updated: 2026-08-20
//	- Translations: [de](https://example.com/de/faq)
//
//	## Section title
//
//	Section body
//
//	---
//
//	The page's Markdown source, verbatim.
//
// [Site.MarkdownAlternate] builds one from a page. Its fields are rendered
// only when the document is written, so a caller can still add Sections — the
// facts about a page that only the caller knows — after building it.
type MarkdownDocument struct {
	// Title is rendered as the H1 heading, Description as a block quote.
	// Empty values are left out.
	Title       string
	Description string

	// URL is the canonical URL of the page, the HTML version the document is
	// an alternate of.
	URL string

	// Lang is the BCP 47 language tag of the page.
	Lang string

	// LastUpdated is the time the page content last changed, rendered as an
	// RFC 3339 date. The zero time is left out.
	LastUpdated time.Time

	// Translations are the versions of the page in other languages.
	Translations []Translation

	// Sections are rendered after the metadata and before the page's
	// Markdown source.
	Sections []MarkdownSection

	// Source is the Markdown source of the page, rendered verbatim after a
	// horizontal rule.
	Source []byte
}

// Bytes renders the document.
func (d *MarkdownDocument) Bytes() []byte {
	var b bytes.Buffer
	if d.Title != "" {
		b.WriteString("# " + markdownLine(d.Title) + "\n\n")
	}
	if d.Description != "" {
		// CommonMark ends a line at a lone carriage return too, so it has to
		// get its "> " prefix like any other line break, or the rest of the
		// description would leave the block quote and could be a heading.
		description := strings.NewReplacer("\r\n", "\n", "\r", "\n").Replace(d.Description)
		for line := range strings.Lines(strings.TrimSpace(description)) {
			line = strings.TrimRight(line, " \t\r\n")
			if line == "" {
				b.WriteString(">\n")
			} else {
				b.WriteString("> " + line + "\n")
			}
		}
		b.WriteByte('\n')
	}

	var metadata []string
	if d.URL != "" {
		metadata = append(metadata, "Canonical URL: "+d.URL)
	}
	if d.Lang != "" {
		metadata = append(metadata, "Language: "+d.Lang)
	}
	if !d.LastUpdated.IsZero() {
		metadata = append(metadata, "Last updated: "+d.LastUpdated.Format(time.DateOnly))
	}
	if len(d.Translations) > 0 {
		translations := make([]string, len(d.Translations))
		for i, translation := range d.Translations {
			translations[i] = "[" + translation.Lang + "](" + markdownURL(translation.URL) + ")"
		}
		metadata = append(metadata, "Translations: "+strings.Join(translations, ", "))
	}
	for _, item := range metadata {
		b.WriteString("- " + item + "\n")
	}
	if len(metadata) > 0 {
		b.WriteByte('\n')
	}

	for _, section := range d.Sections {
		if title := markdownLine(section.Title); title != "" {
			b.WriteString("## " + title + "\n\n")
		}
		if body := strings.TrimSpace(section.Body); body != "" {
			b.WriteString(body + "\n\n")
		}
	}

	b.WriteString("---\n\n")
	b.Write(d.Source)
	return b.Bytes()
}

// WriteTo writes the rendered document to w, implementing io.WriterTo.
func (d *MarkdownDocument) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(d.Bytes())
	return int64(n), err
}

// MarkdownAlternate returns the Markdown alternate of the page: its Markdown
// source (see [Page.MarkdownSource]) with a header carrying its title,
// description, canonical URL, language, last update — LastUpdated, falling back
// to Published — and the URLs of its translations.
//
// It is an error for the page to have no Markdown source or no Path, or to
// fail [Page.Validate]. Whether the page should be published at all is the
// caller's decision: unlike the sitemap, this does not check
// [Page.Indexable], so a handler serving the alternates of a site has to,
// just as it does for the HTML pages.
func (s *Site) MarkdownAlternate(page *Page) (*MarkdownDocument, error) {
	source, ok := page.MarkdownSource()
	if !ok {
		return nil, errs.Errorf("page %q has no Markdown alternate", page.Path)
	}
	canonical, err := s.PageURL(page)
	if err != nil {
		return nil, err
	}
	links, err := s.pageLinks(page)
	if err != nil {
		return nil, err
	}
	lastUpdated := page.LastUpdated
	if lastUpdated.IsZero() {
		lastUpdated = page.Published
	}
	doc := &MarkdownDocument{
		Title:       page.Title,
		Description: page.Description,
		URL:         canonical,
		Lang:        s.pageLang(page),
		LastUpdated: lastUpdated,
		Source:      source,
	}
	for _, link := range links {
		isTranslation := link.rel == RelAlternate && link.typ == "" && link.hrefLang != ""
		if !isTranslation || link.url == canonical || strings.EqualFold(link.hrefLang, HrefLangDefault) {
			continue
		}
		doc.Translations = append(doc.Translations, Translation{Lang: link.hrefLang, URL: link.url})
	}
	return doc, nil
}

// ServeMarkdown writes the Markdown alternate of the page (see
// [Site.MarkdownAlternate]) as a text/markdown response.
//
// The response carries a canonical Link header pointing at the HTML page, so
// search engines index that one rather than the alternate, and "Vary: Accept",
// because a site serving the alternate under the page's own URL to a client
// that asks for Markdown (see [AcceptsMarkdown]) has to keep caches from
// handing it to everyone else. A page that is not [Page.Indexable] is marked
// "X-Robots-Tag: noindex, nofollow", the header form of its robots meta tag.
//
// The etag and lastMod are the caller's revision of the page, or empty and
// zero for none. An etag that is not quoted yet is quoted. Conditional and
// range requests are answered by [http.ServeContent], so a client presenting
// the current etag gets a 304.
func (s *Site) ServeMarkdown(response http.ResponseWriter, request *http.Request, page *Page, etag string, lastMod time.Time) {
	doc, err := s.MarkdownAlternate(page)
	if err != nil {
		mx.RespondNonContextError(response, err)
		return
	}
	header := response.Header()
	header.Set("Content-Type", mx.ContentTypeMarkdown)
	header.Add("Link", pageLink{rel: "canonical", url: doc.URL}.headerValue())
	header.Add("Vary", "Accept")
	if etag != "" {
		header.Set("ETag", quoteETag(etag))
	}
	if !page.Indexable() {
		header.Set("X-Robots-Tag", "noindex, nofollow")
	}
	http.ServeContent(response, request, "", lastMod, bytes.NewReader(doc.Bytes()))
}

// AcceptsMarkdown reports whether the request's Accept header prefers
// text/markdown over text/html: it has to name text/markdown explicitly, with a
// higher quality value than text/html gets, either directly or through a
// "text/*" or "*/*" wildcard. A missing quality value is 1.
//
// At equal quality the client's order decides: text/markdown listed before
// text/html, or named while text/html is only reached through a wildcard,
// wins. RFC 9110 gives the order no meaning, but agents rank equals by it —
// "text/markdown, text/html, */*" asks for Markdown first.
//
// A browser's Accept header ends in "*/*" and names no Markdown, so it gets
// HTML; a client that sends "text/markdown" gets Markdown. Wrap the HTML
// handler of a page with it to serve the alternate under the page's own URL:
//
//	if web.AcceptsMarkdown(request) {
//	    site.ServeMarkdown(response, request, page, etag, lastMod)
//	    return
//	}
func AcceptsMarkdown(request *http.Request) bool {
	// The quality of the most specific range matching a type wins
	// (RFC 9110, section 12.5.1), so text/html is tracked per specificity.
	// -1 means not listed.
	var markdownQ, htmlQ, textQ, anyQ float64 = -1, -1, -1, -1
	markdownFirst := false
	for _, value := range request.Header.Values("Accept") {
		for mediaRange := range strings.SplitSeq(value, ",") {
			mediaType, q, ok := parseMediaRange(mediaRange)
			if !ok {
				continue
			}
			switch mediaType {
			case "text/markdown":
				if markdownQ < 0 && htmlQ < 0 {
					markdownFirst = true
				}
				markdownQ = max(markdownQ, q)
			case "text/html":
				htmlQ = max(htmlQ, q)
			case "text/*":
				textQ = max(textQ, q)
			case "*/*":
				anyQ = max(anyQ, q)
			}
		}
	}
	if markdownQ <= 0 {
		return false
	}
	htmlNamed := htmlQ >= 0
	switch {
	case !htmlNamed && textQ >= 0:
		htmlQ = textQ
	case !htmlNamed:
		htmlQ = anyQ
	}
	if markdownQ != htmlQ {
		return markdownQ > htmlQ
	}
	return markdownFirst || !htmlNamed
}

// parseMediaRange parses one media range of an Accept header into its
// lower-cased media type and quality value. A malformed quality value makes
// the range unusable rather than guessing at what the client meant.
func parseMediaRange(mediaRange string) (mediaType string, q float64, ok bool) {
	mediaType, params, _ := strings.Cut(mediaRange, ";")
	mediaType = strings.ToLower(strings.TrimSpace(mediaType))
	if mediaType == "" {
		return "", 0, false
	}
	q = 1
	for param := range strings.SplitSeq(params, ";") {
		name, value, _ := strings.Cut(param, "=")
		if !strings.EqualFold(strings.TrimSpace(name), "q") {
			continue
		}
		// ParseFloat accepts "NaN", which every comparison is false for.
		parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err != nil || math.IsNaN(parsed) || parsed < 0 || parsed > 1 {
			return "", 0, false
		}
		q = parsed
	}
	return mediaType, q, true
}

// quoteETag returns etag as an entity tag: unchanged if it already is a
// strong ("…") or weak (W/"…") one, otherwise quoted as a strong one, because
// [http.ServeContent] cannot match an unquoted tag and would never answer 304.
func quoteETag(etag string) string {
	if strings.HasPrefix(etag, `"`) || strings.HasPrefix(etag, `W/"`) {
		return etag
	}
	return `"` + etag + `"`
}

// markdownLine returns s on a single line, with every run of whitespace —
// line breaks included — collapsed into one space and other control
// characters dropped. A heading, a list item and the title of a link all end
// at a line break.
func markdownLine(s string) string {
	s = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && !unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)
	return strings.Join(strings.Fields(s), " ")
}

// markdownURL percent-encodes the parentheses of a URL, which is the same URL
// to every client but cannot end the destination of a Markdown link early.
func markdownURL(u string) string {
	return strings.NewReplacer("(", "%28", ")", "%29").Replace(u)
}
