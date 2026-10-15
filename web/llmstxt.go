package web

// https://llmstxt.org/

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ungerik/go-mx"
)

// LLMsTxtPath is the path an llms.txt file is served from, at the root of the
// site like robots.txt.
const LLMsTxtPath = "/llms.txt"

// LLMsTxt is an llms.txt file (https://llmstxt.org): a Markdown index of a
// site written for language models, which have little use for navigation,
// scripts and layout and a limited context to read a site with. It names the
// site, summarizes it and links the pages worth reading in sections, ideally
// their Markdown alternates (see [Site.MarkdownAlternate]):
//
//	# Title
//
//	> Summary
//
//	Context
//
//	## Section
//
//	- [Link title](https://example.com/page.md): Note
//
// A section titled "Optional" has a meaning of its own: its links may be
// skipped by a reader that needs a shorter context.
//
// Like [Robots], it validates before writing, and for the same reason: the
// format is line based, so a line break smuggled into a value would add a
// heading or a link nobody wrote. [Site.LLMsTxt] builds one from the pages of
// a site.
type LLMsTxt struct {
	// Title is the name of the site or project, the one required field.
	Title string

	// Summary is a short summary of the site, rendered as a block quote.
	Summary string

	// Context is a paragraph of further information on how to read the site,
	// rendered after the summary.
	Context string

	// Sections list the links, each under a "## Title" heading.
	Sections []LLMsSection
}

// LLMsSection is one section of an [LLMsTxt]: a heading and a list of links.
type LLMsSection struct {
	// Title is the heading of the section.
	Title string
	// Links are rendered as a list in the given order.
	Links []LLMsLink
}

// LLMsLink is one link of an [LLMsSection], rendered as
// "- [Title](URL): Note".
type LLMsLink struct {
	// Title is the text of the link.
	Title string
	// URL is an absolute http or https URL or a path starting with "/".
	URL string
	// Note is an optional description of what is behind the link.
	Note string
}

// Validate returns an error if the Title is missing, or if a field holds what
// would change the structure of the rendered file: a line break in any field,
// a context starting with "#", an empty section or link title, a bracket,
// backtick or trailing backslash in a link title, which would end, escape or
// swallow the link text, a parenthesis or whitespace in a URL, which would end
// the link destination early, or a URL that is neither an absolute http or
// https URL nor a path within the site.
func (l *LLMsTxt) Validate() error {
	if strings.TrimSpace(l.Title) == "" {
		return fmt.Errorf("llms.txt title is empty")
	}
	if err := validateLLMsLine("title", l.Title); err != nil {
		return err
	}
	if err := validateLLMsLine("summary", l.Summary); err != nil {
		return err
	}
	if err := validateLLMsLine("context", l.Context); err != nil {
		return err
	}
	// The context is the one value rendered at the start of a line of its
	// own, where a "#" makes it a heading — and a "## " heading a section.
	if strings.HasPrefix(strings.TrimLeft(l.Context, " \t"), "#") {
		return fmt.Errorf("llms.txt context %q starts with '#', which makes it a heading", l.Context)
	}
	for i, section := range l.Sections {
		if err := section.Validate(); err != nil {
			return fmt.Errorf("llms.txt section %d: %w", i, err)
		}
	}
	return nil
}

// Validate returns an error if the section title is empty or not a single
// line, or if one of its links is invalid.
func (s LLMsSection) Validate() error {
	if strings.TrimSpace(s.Title) == "" {
		return fmt.Errorf("section title is empty")
	}
	if err := validateLLMsLine("section title", s.Title); err != nil {
		return err
	}
	for i, link := range s.Links {
		if err := link.Validate(); err != nil {
			return fmt.Errorf("link %d: %w", i, err)
		}
	}
	return nil
}

// Validate returns an error if the link title is empty or would break the
// Markdown link, if the URL is not usable as its destination, or if the note
// is not a single line.
func (l LLMsLink) Validate() error {
	if strings.TrimSpace(l.Title) == "" {
		return fmt.Errorf("link title is empty")
	}
	if err := validateLLMsLine("link title", l.Title); err != nil {
		return err
	}
	// The llmstxt.org reference parser reads the title up to the first "]",
	// whatever escaping CommonMark would allow, and a "[" would open a link
	// inside the link text.
	if strings.ContainsAny(l.Title, "[]") {
		return fmt.Errorf("link title %q contains a square bracket", l.Title)
	}
	// A backtick opens a code span, which takes precedence over the link and
	// can run on into the note, swallowing the "](URL)" in between.
	if strings.Contains(l.Title, "`") {
		return fmt.Errorf("link title %q contains a backtick", l.Title)
	}
	if strings.HasSuffix(l.Title, `\`) {
		return fmt.Errorf("link title %q ends in a backslash, which escapes the closing bracket", l.Title)
	}
	if err := validateLLMsURL(l.URL); err != nil {
		return err
	}
	return validateLLMsLine("link note", l.Note)
}

// validateLLMsLine returns an error if value contains a line break or another
// control character.
func validateLLMsLine(field, value string) error {
	if strings.ContainsFunc(value, func(r rune) bool { return r < ' ' && r != '\t' || r == 0x7F }) {
		return fmt.Errorf("llms.txt %s %q contains a line break or control character", field, value)
	}
	return nil
}

// validateLLMsURL returns an error if rawURL is not an absolute http or https
// URL or a path within the site, or holds a character that would end the
// destination of a Markdown link early. A "(" and ")" can be percent-encoded
// as %28 and %29 without changing the URL.
func validateLLMsURL(rawURL string) error {
	switch {
	case rawURL == "":
		return fmt.Errorf("link URL is empty")
	case strings.ContainsFunc(rawURL, func(r rune) bool { return r <= ' ' || r == 0x7F }):
		return fmt.Errorf("link URL %q contains whitespace or control characters", rawURL)
	case strings.ContainsAny(rawURL, "()<>"):
		return fmt.Errorf("link URL %q contains a parenthesis or angle bracket (percent-encode it)", rawURL)
	// "//host/path" starts with a slash too, but is a URL on another host.
	case strings.HasPrefix(rawURL, "/") && !strings.HasPrefix(rawURL, "//"):
		if _, err := url.Parse(rawURL); err != nil {
			return fmt.Errorf("link URL %q: %w", rawURL, err)
		}
		return nil
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("link URL %q: %w", rawURL, err)
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
		return fmt.Errorf("link URL %q is neither an absolute http or https URL nor a path starting with '/'", rawURL)
	}
	return nil
}

// MarshalText renders the llms.txt file, implementing encoding.TextMarshaler.
// It returns the error from [LLMsTxt.Validate] instead of writing a file whose
// structure would differ from what the value describes.
func (l *LLMsTxt) MarshalText() ([]byte, error) {
	if err := l.Validate(); err != nil {
		return nil, err
	}
	var b bytes.Buffer
	b.WriteString("# " + l.Title + "\n")
	if l.Summary != "" {
		b.WriteString("\n> " + l.Summary + "\n")
	}
	if l.Context != "" {
		b.WriteString("\n" + l.Context + "\n")
	}
	for _, section := range l.Sections {
		b.WriteString("\n## " + section.Title + "\n")
		if len(section.Links) > 0 {
			b.WriteByte('\n')
		}
		for _, link := range section.Links {
			b.WriteString("- [" + link.Title + "](" + link.URL + ")")
			if link.Note != "" {
				b.WriteString(": " + link.Note)
			}
			b.WriteByte('\n')
		}
	}
	return b.Bytes(), nil
}

// WriteTo writes the llms.txt file to w, implementing io.WriterTo.
func (l *LLMsTxt) WriteTo(w io.Writer) (int64, error) {
	text, err := l.MarshalText()
	if err != nil {
		return 0, err
	}
	n, err := w.Write(text)
	return int64(n), err
}

// String renders the llms.txt file, returning a "!ERROR: " line instead of the
// file content if the value does not validate. It is meant for debugging and
// tests; use [LLMsTxt.MarshalText] to handle the error.
func (l *LLMsTxt) String() string {
	text, err := l.MarshalText()
	if err != nil {
		return "!ERROR: " + err.Error()
	}
	return string(text)
}

// HandleHTTP writes the llms.txt file as a text/markdown response,
// implementing http.HandlerFunc. The ETag is a hash of the content, so a
// client re-reading an unchanged file gets a 304. On a validation error it
// responds with a generic 500 status via [mx.RespondNonContextError].
func (l *LLMsTxt) HandleHTTP(response http.ResponseWriter, request *http.Request) {
	text, err := l.MarshalText()
	if err != nil {
		mx.RespondNonContextError(response, err)
		return
	}
	hash := sha256.Sum256(text)
	response.Header().Set("Content-Type", mx.ContentTypeMarkdown)
	response.Header().Set("ETag", `"`+hex.EncodeToString(hash[:16])+`"`)
	http.ServeContent(response, request, "", time.Time{}, bytes.NewReader(text))
}

// DefaultLLMsSectionTitle is the section title [Site.LLMsTxt] lists pages
// without a Page.Type under.
const DefaultLLMsSectionTitle = "Pages"

// LLMsTxtOptions configure the llms.txt built by [Site.LLMsTxt].
type LLMsTxtOptions struct {
	// Title of the llms.txt, Site.Title when empty.
	Title string

	// Summary of the llms.txt, Site.Description when empty.
	Summary string

	// Context is the optional paragraph after the summary.
	Context string

	// SectionTitles maps a Page.Type to the title of the section listing the
	// pages of that type. A type without an entry is used as the title
	// itself, and pages without a type are listed under
	// [DefaultLLMsSectionTitle] unless the map has an entry for "".
	SectionTitles map[string]string
}

// LLMsTxt builds an llms.txt listing the Markdown alternates of the site's
// pages: every [Page.Indexable] page with a Markdown source (see
// [Page.MarkdownSource]), linked by the URL of its alternate (see
// [MarkdownPath]), with its title and description. The pages are grouped into
// one section per Page.Type, in the order the types first appear in the
// Sources, and listed in source order within a section. Types that
// SectionTitles maps to the same title share one section.
//
// It reads the Sources with content, because only the content tells whether a
// page has a Markdown source. A site with expensive sources should cache the
// result.
//
// Page titles and descriptions are site content rather than the structure of
// the file, so instead of failing the whole file they are made to fit: line
// breaks become spaces, other control characters are dropped, square brackets
// in a title become parentheses and backticks apostrophes, and the
// parentheses of a URL are percent-encoded. A page without a usable title is
// listed by its path.
//
// A public page that has no Path, or whose Markdown alternate could not be
// built, fails the whole file like it fails the sitemap: it is a bug in the
// source, not something to discover as a silently incomplete index.
//
// It is the easy default. A site that wants other sections, notes or links
// builds its own [LLMsTxt].
func (s *Site) LLMsTxt(ctx context.Context, opts LLMsTxtOptions) (*LLMsTxt, error) {
	llms := &LLMsTxt{
		Title:   firstNonEmpty(opts.Title, s.Title),
		Summary: markdownLine(firstNonEmpty(opts.Summary, s.Description)),
		Context: opts.Context,
	}
	sectionIndex := make(map[string]int)
	for page, err := range s.Pages(ctx, true) {
		if err != nil {
			return nil, err
		}
		if !page.Indexable() {
			continue
		}
		if _, ok := page.MarkdownSource(); !ok {
			continue
		}
		// A page whose alternate cannot be built would be linked to a 500.
		if _, err := s.pageLinks(page); err != nil {
			return nil, err
		}
		markdownPath, err := MarkdownPath(page.Path)
		if err != nil {
			return nil, fmt.Errorf("page %q: %w", page.Title, err)
		}
		alternateURL, err := s.URL(markdownPath)
		if err != nil {
			return nil, err
		}
		title, ok := opts.SectionTitles[page.Type]
		if !ok {
			title = firstNonEmpty(page.Type, DefaultLLMsSectionTitle)
		}
		// Sections are told apart by their title, so types mapped to the
		// same title share one section instead of repeating the heading.
		i, ok := sectionIndex[title]
		if !ok {
			i = len(llms.Sections)
			sectionIndex[title] = i
			llms.Sections = append(llms.Sections, LLMsSection{Title: title})
		}
		llms.Sections[i].Links = append(llms.Sections[i].Links, LLMsLink{
			Title: firstNonEmpty(llmsLinkTitle(page.Title), llmsLinkTitle(page.Path)),
			URL:   markdownURL(alternateURL),
			Note:  markdownLine(page.Description),
		})
	}
	if err := llms.Validate(); err != nil {
		return nil, err
	}
	return llms, nil
}

// llmsLinkTitle makes a page title usable as the text of an [LLMsLink]: on
// one line, with square brackets as parentheses, backticks as apostrophes and
// no trailing backslash. It returns "" if nothing usable is left.
func llmsLinkTitle(title string) string {
	title = strings.NewReplacer("[", "(", "]", ")", "`", "'").Replace(markdownLine(title))
	return strings.TrimSpace(strings.TrimRight(title, `\`))
}
