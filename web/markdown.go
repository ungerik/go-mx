package web

// https://llmstxt.org/#format

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
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
	// The checks look at the decoded path, the one a server routes by:
	// "/en/faq%2Emd" is "/en/faq.md" there, and "/en/%69ndex" is "/en/index".
	parsed, err := url.Parse(pagePath)
	if err != nil {
		return "", fmt.Errorf("path %q: %w", pagePath, err)
	}
	if strings.HasSuffix(strings.ToLower(parsed.Path), ".md") {
		return "", fmt.Errorf("path %q already ends in .md", pagePath)
	}
	if strings.HasSuffix(pagePath, "/") {
		return pagePath + "index.md", nil
	}
	if gopath.Base(parsed.Path) == "index" {
		return "", fmt.Errorf("path %q ends in /index, whose Markdown path would be that of %q", pagePath, gopath.Dir(parsed.Path))
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

	// Translations are the versions of the page in other languages. One
	// whose Lang leaves no usable link text is left out.
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
		description := lineBreakNormalizer.Replace(d.Description)
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

	// Every metadata item is one line of a list. The fields are the
	// caller's to set, so each is put on one line before it is written.
	var metadata []string
	if u := markdownLine(d.URL); u != "" {
		metadata = append(metadata, "Canonical URL: "+u)
	}
	if lang := markdownLine(d.Lang); lang != "" {
		metadata = append(metadata, "Language: "+lang)
	}
	if !d.LastUpdated.IsZero() {
		metadata = append(metadata, "Last updated: "+d.LastUpdated.Format(time.DateOnly))
	}
	var translations []string
	for _, translation := range d.Translations {
		if lang := markdownLinkText(translation.Lang); lang != "" {
			translations = append(translations, "["+lang+"]("+markdownURL(markdownLine(translation.URL))+")")
		}
	}
	if len(translations) > 0 {
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
// to Published — and the URLs of its translations. The source is published as
// it is, HTML comments included, which an HTML renderer may drop from the page.
//
// It is an error for the page to have no Markdown source or no Path, to fail
// [Page.Validate], or to fail the checks of its links that need the site: a
// BaseURL, and for a page with translations a valid language of its own (from
// Page.Lang or Site.Lang) that none of them repeats. Whether the page should be published at all is the
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
	// A page whose hreflang links are broken gets no alternate either.
	err = s.validatePageLinks(page)
	if err != nil {
		return nil, err
	}
	doc := &MarkdownDocument{
		Title:       page.Title,
		Description: page.Description,
		URL:         canonical,
		Lang:        s.PageLang(page),
		LastUpdated: page.LastModified(),
		Source:      source,
	}
	for _, alternate := range page.Alternates {
		if !alternate.IsTranslation() || strings.EqualFold(alternate.HrefLang, HrefLangDefault) {
			continue
		}
		u, err := s.URL(alternate.Path)
		if err != nil {
			return nil, err
		}
		doc.Translations = append(doc.Translations, Translation{Lang: alternate.HrefLang, URL: u})
	}
	return doc, nil
}

// MarkdownAlternateURL returns the absolute URL of the page's Markdown
// alternate, at the [MarkdownPath] of its Path. It fails for a page without a
// valid Path and for one whose links cannot be built, as
// [Site.MarkdownAlternate] does, so the alternate can be served wherever this
// URL is listed. Whether the page has a Markdown source is not checked.
func (s *Site) MarkdownAlternateURL(page *Page) (string, error) {
	err := s.validatePageLinks(page)
	if err != nil {
		return "", err
	}
	markdownPath, err := MarkdownPath(page.Path)
	if err != nil {
		return "", fmt.Errorf("page %q: %w", page.Title, err)
	}
	return s.URL(markdownPath)
}

// ServeMarkdown writes the Markdown alternate of the page (see
// [Site.MarkdownAlternate]) as a text/markdown response, with the headers and
// the conditional-request handling of [Site.ServeMarkdownDocument]. It serves
// a page as it is; to add Sections to the alternate, or change its other
// fields, build it with [Site.MarkdownAlternate] and serve the result with
// [Site.ServeMarkdownDocument] instead.
//
// A page without a Markdown source has no alternate to serve and gets a 404;
// any other failure of [Site.MarkdownAlternate] is a 500.
func (s *Site) ServeMarkdown(response http.ResponseWriter, request *http.Request, page *Page, etag string, lastMod time.Time) {
	if _, ok := page.MarkdownSource(); !ok {
		http.NotFound(response, request)
		return
	}
	doc, err := s.MarkdownAlternate(page)
	if err != nil {
		mx.RespondNonContextError(response, err)
		return
	}
	s.ServeMarkdownDocument(response, request, page, doc, etag, lastMod)
}

// ServeMarkdownDocument writes doc, the Markdown alternate of the page, as a
// text/markdown response. It is [Site.ServeMarkdown] for a document the caller
// built with [Site.MarkdownAlternate] and then changed, typically by adding
// Sections; a page served as it is goes through [Site.ServeMarkdown].
//
// The response carries a canonical Link header pointing at doc.URL, the HTML
// page, so search engines index that one rather than the alternate (none for
// an empty doc.URL, which the body leaves out too), and "Vary: Accept",
// because a site serving the alternate under the page's own URL to a client
// that asks for Markdown (see [AcceptsMarkdown]) has to keep caches from
// handing it to everyone else. A page that is not [Page.Indexable] is marked
// "X-Robots-Tag: noindex, nofollow", the header form of its robots meta tag.
// That is all the page is used for: everything else comes from doc.
//
// The etag and lastMod are the caller's revision of the document, or empty and
// zero for none; for a document with Sections of its own they have to change
// when those do. The etag is quoted if it is not yet and marked "-md" inside
// the quotes: under content negotiation the alternate shares its URL with the
// HTML page, and a cache revalidating both with one tag could answer a
// Markdown client with the HTML. Conditional and range requests are answered
// by [http.ServeContent], so a client presenting the current tag gets a 304.
// Headers set before the call, like Cache-Control, are left as they are, and
// the request is used as given, so a caller that wants only the etag to
// answer a conditional request passes a clone without If-Modified-Since.
//
// A nil doc is a bug in the caller and gets a 500, like a failure of
// [Site.MarkdownAlternate] in [Site.ServeMarkdown].
func (s *Site) ServeMarkdownDocument(response http.ResponseWriter, request *http.Request, page *Page, doc *MarkdownDocument, etag string, lastMod time.Time) {
	if doc == nil {
		mx.RespondNonContextError(response, errs.Errorf("page %q: no Markdown document to serve", page.Path))
		return
	}
	header := response.Header()
	header.Set("Content-Type", mx.ContentTypeMarkdown)
	// "<>" would name the alternate itself as canonical.
	if doc.URL != "" {
		header.Add("Link", pageLink{rel: "canonical", url: doc.URL}.headerValue())
	}
	// The caller may have set it already, for the HTML response it would
	// have answered with instead (see AcceptsMarkdown).
	if !headerListContains(header.Values("Vary"), "Accept") {
		header.Add("Vary", "Accept")
	}
	if etag != "" {
		header.Set("ETag", markdownETag(etag))
	}
	if !page.Indexable() {
		header.Set("X-Robots-Tag", robotsNoIndex)
	}
	http.ServeContent(response, request, "", lastMod, bytes.NewReader(doc.Bytes()))
}

// headerListContains reports whether element is one of the comma-separated
// elements of the header values, compared case-insensitively.
func headerListContains(values []string, element string) bool {
	for _, value := range values {
		for e := range strings.SplitSeq(value, ",") {
			if strings.EqualFold(strings.TrimSpace(e), element) {
				return true
			}
		}
	}
	return false
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
// handler of a page with it to serve the alternate under the page's own URL.
// Both answers then depend on the Accept header, so the HTML response needs
// "Vary: Accept" as much as the Markdown one, or a cache that kept a browser's
// HTML hands it to the next agent asking for Markdown. A CDN that ignores Vary
// for HTML defeats this either way; link agents to the alternate's own URL
// (see [Site.MarkdownAlternateURL]) behind one:
//
//	response.Header().Add("Vary", "Accept")
//	if _, ok := page.MarkdownSource(); ok && web.AcceptsMarkdown(request) {
//	    site.ServeMarkdown(response, request, page, etag, lastMod)
//	    return
//	}
//
// A page without a Markdown source has no alternate, so it answers with HTML
// whatever the client prefers.
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

// markdownETag returns the entity tag of a Markdown alternate for the
// caller's etag: a strong ("…") or weak (W/"…") tag with "-md" added inside
// its quotes, anything else quoted as a strong one first, because
// [http.ServeContent] cannot match an unquoted tag and would never answer 304.
func markdownETag(etag string) string {
	weak := strings.HasPrefix(etag, `W/`)
	opaque := strings.TrimPrefix(etag, `W/`)
	if len(opaque) >= 2 && strings.HasPrefix(opaque, `"`) && strings.HasSuffix(opaque, `"`) {
		opaque = opaque[1 : len(opaque)-1]
	} else {
		// Not a tag yet: all of it, quotes removed, becomes the opaque part.
		weak = false
		opaque = strings.ReplaceAll(etag, `"`, "")
	}
	tag := `"` + opaque + `-md"`
	if weak {
		return `W/` + tag
	}
	return tag
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

// lineBreakNormalizer turns CRLF and lone CR line breaks into LF.
var lineBreakNormalizer = strings.NewReplacer("\r\n", "\n", "\r", "\n")

// markdownURLEscaper percent-encodes the parentheses, spaces and backslashes
// of a URL.
var markdownURLEscaper = strings.NewReplacer("(", "%28", ")", "%29", " ", "%20", `\`, "%5C")

// markdownLinkText makes s usable as the text of a Markdown link: on one line,
// with square brackets as parentheses, backticks as apostrophes, "<" as "‹"
// and no trailing backslash, so it can neither end, escape nor swallow the
// link. It returns "" if nothing usable is left.
func markdownLinkText(s string) string {
	// Spaces and backslashes are trimmed together, so `a \ \` leaves
	// no backslash at the end.
	s = markdownLinkTextReplacer.Replace(markdownLine(s))
	return strings.TrimRight(s, ` \`)
}

// markdownLinkTextReplacer replaces what [markdownLinkText] may not hold with
// look-alikes.
var markdownLinkTextReplacer = strings.NewReplacer("[", "(", "]", ")", "`", "'", "<", "‹")

// markdownURL percent-encodes the parentheses, spaces and backslashes of a
// URL, which is the same URL to every client but cannot end or escape the
// destination of a Markdown link early.
func markdownURL(u string) string {
	return markdownURLEscaper.Replace(u)
}
