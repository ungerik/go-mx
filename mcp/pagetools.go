package mcp

import (
	"bytes"
	"cmp"
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/domonda/go-errs"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ungerik/go-mx/web"
)

const (
	// DefaultSearchLimit is the number of results search_pages returns when
	// the call does not ask for another limit.
	DefaultSearchLimit = 10

	// MaxSearchLimit is the most results search_pages returns, whatever limit
	// the call asks for.
	MaxSearchLimit = 25

	// DefaultMaxResponseBytes is the size get_page truncates a page at when
	// PageToolsOptions.MaxResponseBytes is zero.
	DefaultMaxResponseBytes = 64 << 10

	// snippetRunes is about how long a search_pages snippet is.
	snippetRunes = 160
)

// ErrPageNotFound is the tool error get_page answers with for a path that is
// not a public page with a Markdown alternate. A page that does not exist and
// one that is not public get the same answer, so the error tells a client
// nothing about the pages it cannot read.
var ErrPageNotFound = errors.New("page not found")

// errUnavailable is the tool error for a failure on the side of the server.
// Its details are for the server's log (see PageToolsOptions.OnError), not for
// the client. It is a plain error rather than one of go-errs, whose message
// would carry the call stack to the client.
var errUnavailable = errors.New("the pages are unavailable, try again later")

// PageToolsOptions configure the tools registered by [AddPageTools]. The zero
// value is ready to use.
type PageToolsOptions struct {
	// MaxResponseBytes is the size get_page truncates the Markdown of a page
	// at, with a note saying so. [DefaultMaxResponseBytes] when zero.
	MaxResponseBytes int

	// Language returns the BCP 47 language tag of a page, which the tools
	// report and filter by. The default is Page.Lang, else the first path
	// segment if it is a language tag with a two-letter ISO 639-1 language
	// ("en", "de-AT"), else Site.Lang. The path heuristic knows neither
	// three-letter languages nor a site whose "/no/" or "/my/" is not
	// Norwegian or Burmese, so set Page.Lang or this function where it errs.
	Language func(*web.Page) string

	// OnError is called with the errors the tools do not show the client: a
	// failing source, or a page whose URLs or Markdown alternate cannot be
	// built and is left out of the results. Nil ignores them.
	OnError func(ctx context.Context, err error)
}

// PageInfo describes a page in the results of the page tools.
type PageInfo struct {
	Title       string `json:"title" jsonschema:"the title of the page"`
	Description string `json:"description,omitempty" jsonschema:"a summary of the page"`
	URL         string `json:"canonical_url" jsonschema:"the URL of the HTML page"`
	MarkdownURL string `json:"markdown_url" jsonschema:"the URL of the Markdown version of the page"`
	Language    string `json:"language,omitempty" jsonschema:"the BCP 47 language tag of the page"`
	LastUpdated string `json:"last_updated,omitempty" jsonschema:"when the content of the page last changed, as an RFC 3339 timestamp"`
	Snippet     string `json:"snippet,omitempty" jsonschema:"an excerpt of the page around a match of the search query"`
}

type searchPagesInput struct {
	Query    string `json:"query" jsonschema:"the words to search for in the title, description and text of the pages; a page has to contain every word"`
	Language string `json:"language,omitempty" jsonschema:"only return pages in this BCP 47 language, like en or de"`
	Limit    int    `json:"limit,omitempty" jsonschema:"the maximum number of results: 10 when not given, at most 25"`
}

type searchPagesOutput struct {
	Results []PageInfo `json:"results" jsonschema:"the matching pages, best match first"`
}

type listPagesInput struct {
	Language string `json:"language,omitempty" jsonschema:"only list pages in this BCP 47 language, like en or de"`
}

type listPagesOutput struct {
	Pages []PageInfo `json:"pages" jsonschema:"the public pages"`
}

type getPageInput struct {
	Path string `json:"path" jsonschema:"the path of the page, like /en/faq, or its canonical_url or markdown_url"`
}

type getPageOutput struct {
	Page      PageInfo `json:"page" jsonschema:"the page"`
	Truncated bool     `json:"truncated" jsonschema:"whether the Markdown was cut short; the whole page is at markdown_url"`
}

// AddPageTools adds three read-only tools to server that let an agent read
// the public pages of source, by their Markdown alternates:
//
//   - search_pages(query, language?, limit?) finds pages containing every word
//     of the query in their title, description or Markdown source, best match
//     first, with a snippet around a match. It returns [DefaultSearchLimit]
//     results unless asked for another number, at most [MaxSearchLimit].
//   - list_pages(language?) lists all of them.
//   - get_page(path) returns the Markdown alternate of one page (see
//     [web.Site.MarkdownAlternate]), truncated at
//     PageToolsOptions.MaxResponseBytes.
//
// A page is public if it is [web.Page.Indexable] and has a Markdown source
// (see [web.Page.MarkdownSource]); no tool ever returns another one. The
// source is read with content on every call, so a site with an expensive
// source should pass a cached one. The site builds the URLs, and is usually
// the source as well.
//
// Passing the source is what exposes it: a nil source registers nothing and
// returns an error, as does a nil server or a site that does not validate.
func AddPageTools(server *sdk.Server, source web.PageSource, site *web.Site, opts PageToolsOptions) error {
	switch {
	case server == nil:
		return errs.New("mcp.AddPageTools: server is nil")
	case source == nil:
		return errs.New("mcp.AddPageTools: source is nil, no pages to expose")
	case site == nil:
		return errs.New("mcp.AddPageTools: site is nil")
	case opts.MaxResponseBytes < 0:
		return errs.Errorf("mcp.AddPageTools: negative MaxResponseBytes %d", opts.MaxResponseBytes)
	}
	if err := site.Validate(); err != nil {
		return errs.Errorf("mcp.AddPageTools: %w", err)
	}
	if opts.MaxResponseBytes == 0 {
		opts.MaxResponseBytes = DefaultMaxResponseBytes
	}
	if opts.Language == nil {
		opts.Language = defaultLanguage(site)
	}
	tools := &pageTools{source: source, site: site, opts: opts}

	of := "the site"
	if site.Title != "" {
		of = site.Title
	}
	readOnly := &sdk.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true, OpenWorldHint: new(false)}
	sdk.AddTool(server, &sdk.Tool{
		Name:        "search_pages",
		Title:       "Search pages",
		Description: "Search the public pages of " + of + " for words in their title, description and text. Returns the best matches first, each with a snippet; read one with get_page.",
		Annotations: readOnly,
	}, tools.searchPages)
	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_pages",
		Title:       "List pages",
		Description: "List all public pages of " + of + ", optionally only those in one language. Read one with get_page.",
		Annotations: readOnly,
	}, tools.listPages)
	sdk.AddTool(server, &sdk.Tool{
		Name:        "get_page",
		Title:       "Get page",
		Description: "Get the content of a public page of " + of + " as Markdown, starting with its title, description, URL, language and translations.",
		Annotations: readOnly,
	}, tools.getPage)
	return nil
}

// pageTools implements the tools of [AddPageTools].
type pageTools struct {
	source web.PageSource
	site   *web.Site
	opts   PageToolsOptions
}

// publicPage is a page the tools may return, with what they need of it.
type publicPage struct {
	page     *web.Page
	info     PageInfo
	markdown []byte
}

// publicPages returns the public pages of the source in source order, keeping
// only those in language if it is not empty. A page whose URLs or Markdown
// alternate cannot be built is reported to OnError and left out, so that one
// broken page does not take every other page offline with it, and so that no
// tool lists a page get_page cannot return.
func (t *pageTools) publicPages(ctx context.Context, language string) ([]publicPage, error) {
	var pages []publicPage
	for page, err := range t.source.Pages(ctx, true) {
		if err != nil {
			return nil, err
		}
		if !page.Indexable() {
			continue
		}
		markdown, ok := page.MarkdownSource()
		if !ok {
			continue
		}
		info, err := t.pageInfo(page)
		if err != nil {
			t.reportError(ctx, err)
			continue
		}
		if !matchesLanguage(info.Language, language) {
			continue
		}
		pages = append(pages, publicPage{page: page, info: info, markdown: markdown})
	}
	return pages, nil
}

// pageInfo returns the description of a page in the tool results.
func (t *pageTools) pageInfo(page *web.Page) (PageInfo, error) {
	// PageHeaders builds the links a Markdown alternate is built from, which
	// is what fails for a page with invalid alternates or no language.
	if _, err := t.site.PageHeaders(page); err != nil {
		return PageInfo{}, err
	}
	canonical, err := t.site.PageURL(page)
	if err != nil {
		return PageInfo{}, err
	}
	markdownPath, err := web.MarkdownPath(page.Path)
	if err != nil {
		return PageInfo{}, err
	}
	markdownURL, err := t.site.URL(markdownPath)
	if err != nil {
		return PageInfo{}, err
	}
	lastUpdated := page.LastUpdated
	if lastUpdated.IsZero() {
		lastUpdated = page.Published
	}
	info := PageInfo{
		Title:       page.Title,
		Description: page.Description,
		URL:         canonical,
		MarkdownURL: markdownURL,
		Language:    t.opts.Language(page),
	}
	if !lastUpdated.IsZero() {
		info.LastUpdated = lastUpdated.Format(time.RFC3339)
	}
	return info, nil
}

// reportError passes an error the client does not see to OnError.
func (t *pageTools) reportError(ctx context.Context, err error) {
	if t.opts.OnError != nil {
		t.opts.OnError(ctx, err)
	}
}

// unavailable reports err and returns the error the client sees instead.
func (t *pageTools) unavailable(ctx context.Context, err error) error {
	t.reportError(ctx, err)
	return errUnavailable
}

func (t *pageTools) searchPages(ctx context.Context, _ *sdk.CallToolRequest, input searchPagesInput) (*sdk.CallToolResult, searchPagesOutput, error) {
	words := strings.Fields(strings.ToLower(input.Query))
	if len(words) == 0 {
		return nil, searchPagesOutput{}, errors.New("the query is empty")
	}
	limit := input.Limit
	if limit <= 0 {
		limit = DefaultSearchLimit
	}
	limit = min(limit, MaxSearchLimit)

	pages, err := t.publicPages(ctx, input.Language)
	if err != nil {
		return nil, searchPagesOutput{}, t.unavailable(ctx, err)
	}
	type match struct {
		info  PageInfo
		score int
	}
	var matches []match
	for _, page := range pages {
		score, snippet := searchPage(page, words)
		if score == 0 {
			continue
		}
		page.info.Snippet = snippet
		matches = append(matches, match{info: page.info, score: score})
	}
	// Stable, so that equally good matches stay in source order.
	slices.SortStableFunc(matches, func(a, b match) int { return cmp.Compare(b.score, a.score) })
	output := searchPagesOutput{Results: make([]PageInfo, 0, min(len(matches), limit))}
	for _, m := range matches[:min(len(matches), limit)] {
		output.Results = append(output.Results, m.info)
	}
	return nil, output, nil
}

// searchPage scores how well the page matches the lower-cased words — zero if
// one of them is missing from the page — and returns a snippet of the
// Markdown source around the first match in it, or the description if only
// the title and description match.
//
// A word in the title counts most and one in the description more than one in
// the text, where every occurrence counts, up to a cap so that a long page
// repeating a word does not outrank the page that is about it.
func searchPage(page publicPage, words []string) (score int, snippet string) {
	title := strings.ToLower(page.info.Title)
	description := strings.ToLower(page.info.Description)
	text := strings.ToLower(string(page.markdown))
	first := -1
	for _, word := range words {
		wordScore := 0
		if strings.Contains(title, word) {
			wordScore += 10
		}
		if strings.Contains(description, word) {
			wordScore += 3
		}
		if i := strings.Index(text, word); i >= 0 {
			wordScore += min(strings.Count(text, word), 10)
			if first < 0 || i < first {
				first = i
			}
		}
		if wordScore == 0 {
			return 0, ""
		}
		score += wordScore
	}
	if first < 0 {
		return score, oneLine(page.info.Description)
	}
	return score, excerpt(string(page.markdown), text, first)
}

// excerpt returns about [snippetRunes] runes of text around the byte offset
// at in lower, the lower-cased text, on one line, with an ellipsis where it
// cuts the text.
//
// The offset is converted through the rune index, because lower-casing maps
// every rune to exactly one rune but not always to one of the same byte
// length.
func excerpt(text, lower string, at int) string {
	runes := []rune(text)
	center := utf8.RuneCountInString(lower[:at])
	start := max(0, center-snippetRunes/3)
	end := min(len(runes), start+snippetRunes)
	snippet := oneLine(string(runes[start:end]))
	if start > 0 {
		snippet = "…" + snippet
	}
	if end < len(runes) {
		snippet += "…"
	}
	return snippet
}

func (t *pageTools) listPages(ctx context.Context, _ *sdk.CallToolRequest, input listPagesInput) (*sdk.CallToolResult, listPagesOutput, error) {
	pages, err := t.publicPages(ctx, input.Language)
	if err != nil {
		return nil, listPagesOutput{}, t.unavailable(ctx, err)
	}
	output := listPagesOutput{Pages: make([]PageInfo, len(pages))}
	for i, page := range pages {
		output.Pages[i] = page.info
	}
	return nil, output, nil
}

func (t *pageTools) getPage(ctx context.Context, _ *sdk.CallToolRequest, input getPageInput) (*sdk.CallToolResult, getPageOutput, error) {
	if input.Path == "" {
		return nil, getPageOutput{}, errors.New("the path is empty")
	}
	pages, err := t.publicPages(ctx, "")
	if err != nil {
		return nil, getPageOutput{}, t.unavailable(ctx, err)
	}
	i := slices.IndexFunc(pages, func(page publicPage) bool {
		markdownPath, _ := web.MarkdownPath(page.page.Path)
		return input.Path == page.page.Path || input.Path == markdownPath ||
			input.Path == page.info.URL || input.Path == page.info.MarkdownURL
	})
	if i < 0 {
		return nil, getPageOutput{}, ErrPageNotFound
	}
	page := pages[i]
	doc, err := t.site.MarkdownAlternate(page.page)
	if err != nil {
		return nil, getPageOutput{}, t.unavailable(ctx, err)
	}
	markdown, truncated := truncate(doc.Bytes(), t.opts.MaxResponseBytes, page.info.MarkdownURL)
	// The Markdown goes to the model as text, the metadata as structured
	// content, so the page is not sent twice — once as text and once
	// JSON-escaped — as it would be with the Markdown in the output struct.
	result := &sdk.CallToolResult{Content: []sdk.Content{&sdk.TextContent{Text: string(markdown)}}}
	return result, getPageOutput{Page: page.info, Truncated: truncated}, nil
}

// truncate cuts markdown to at most maxBytes, on a line break if there is one
// in the second half of what is kept, and appends a note saying where the
// whole page is. A maxBytes too small for the note returns just the note.
func truncate(markdown []byte, maxBytes int, markdownURL string) ([]byte, bool) {
	if len(markdown) <= maxBytes {
		return markdown, false
	}
	note := fmt.Sprintf("\n\n[Truncated: the page has %d bytes, more than one response holds. The whole page is at %s]\n", len(markdown), markdownURL)
	cut := max(0, maxBytes-len(note))
	for cut > 0 && !utf8.RuneStart(markdown[cut]) {
		cut--
	}
	if i := bytes.LastIndexByte(markdown[:cut], '\n'); i > cut/2 {
		cut = i
	}
	return append(markdown[:cut:cut], note...), true
}

// pathLanguageRegexp matches a first path segment that is taken for a
// language tag: a two-letter language with optional subtags, like "en",
// "de-AT" or "zh-Hant". Longer primary subtags are left out on purpose, as
// they would mistake "/faq/" or "/api/" for a language.
var pathLanguageRegexp = regexp.MustCompile(`^[A-Za-z]{2}(-[A-Za-z0-9]{2,8})*$`)

// iso639_1 are the two-letter ISO 639-1 language codes. Matching the primary
// subtag against them keeps "/go/", "/ai/" or "/ui/" from being taken for a
// language.
var iso639_1 = func() map[string]bool {
	codes := make(map[string]bool)
	for code := range strings.FieldsSeq(`
		aa ab ae af ak am an ar as av ay az ba be bg bh bi bm bn bo br bs ca ce
		ch co cr cs cu cv cy da de dv dz ee el en eo es et eu fa ff fi fj fo fr
		fy ga gd gl gn gu gv ha he hi ho hr ht hu hy hz ia id ie ig ii ik io is
		it iu ja jv ka kg ki kj kk kl km kn ko kr ks ku kv kw ky la lb lg li ln
		lo lt lu lv mg mh mi mk ml mn mr ms mt my na nb nd ne ng nl nn no nr nv
		ny oc oj om or os pa pi pl ps pt qu rm rn ro ru rw sa sc sd se sg si sk
		sl sm sn so sq sr ss st su sv sw ta te tg th ti tk tl tn to tr ts tt tw
		ty ug uk ur uz ve vi vo wa wo xh yi yo za zh zu`) {
		codes[code] = true
	}
	return codes
}()

// defaultLanguage returns the default PageToolsOptions.Language for site.
func defaultLanguage(site *web.Site) func(*web.Page) string {
	return func(page *web.Page) string {
		if page.Lang != "" {
			return page.Lang
		}
		segment, _, _ := strings.Cut(strings.TrimPrefix(page.Path, "/"), "/")
		if pathLanguageRegexp.MatchString(segment) && iso639_1[strings.ToLower(segment[:2])] {
			return segment
		}
		return site.Lang
	}
}

// matchesLanguage reports whether a page in language belongs to the results
// for the filter: every page for an empty filter, else a page in the language
// itself or in a regional variant of it, so "en" finds "en-GB" too. Language
// tags are case-insensitive.
func matchesLanguage(language, filter string) bool {
	if filter == "" {
		return true
	}
	language, filter = strings.ToLower(language), strings.ToLower(filter)
	return language == filter || strings.HasPrefix(language, filter+"-")
}

// oneLine collapses every run of whitespace in s, line breaks included, into
// one space.
func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
