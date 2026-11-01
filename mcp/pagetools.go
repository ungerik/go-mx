package mcp

import (
	"bytes"
	"cmp"
	"context"
	"errors"
	"fmt"
	"iter"
	"math"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/domonda/go-errs"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/web"
)

const (
	// DefaultSearchLimit is the number of results search_pages returns when
	// the call does not ask for another limit.
	DefaultSearchLimit = 10

	// MaxSearchLimit is the most results search_pages returns, whatever limit
	// the call asks for.
	MaxSearchLimit = 25

	// MaxQueryBytes and MaxQueryWords bound a search_pages query, whose cost
	// grows with its words times the text of every public page: a query
	// longer than MaxQueryBytes or with more than MaxQueryWords distinct
	// words is refused rather than scanned. Repeated words count once.
	MaxQueryBytes = 256
	MaxQueryWords = 16

	// DefaultListLimit is the number of pages list_pages returns when the
	// call does not ask for another limit.
	DefaultListLimit = 50

	// MaxListLimit is the most pages list_pages returns, whatever limit the
	// call asks for. The pages after them are listed by further calls with
	// the next_cursor of the previous one.
	MaxListLimit = 200

	// DefaultMaxResponseBytes is the size get_page truncates a page at when
	// PageToolsOptions.MaxResponseBytes is zero.
	DefaultMaxResponseBytes = 64 << 10

	// snippetRunes is about how long a search_pages snippet is.
	snippetRunes = 160

	// maxLanguageBytes is the longest language filter accepted, the length
	// of the longest BCP 47 tag in practical use.
	maxLanguageBytes = 35

	// The search_pages score of a query word: a word in the title counts
	// most, one in the description more than one in the text, where every
	// occurrence counts up to maxTextScore. The cap stays below titleScore,
	// so a long page repeating a word cannot outrank the page about it.
	titleScore       = 10
	descriptionScore = 3
	maxTextScore     = titleScore - 1
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
	// report and filter by and get_page puts into the Markdown header. The
	// default is [web.Site.PageLang], Page.Lang else Site.Lang, the language
	// the page is rendered with, so the tools never contradict the page.
	Language func(*web.Page) string

	// OnError is called with the errors the tools do not show the client: a
	// failing source, a panic in a tool (the source, the Language hook or a
	// page), or a page whose URLs or Markdown alternate cannot be built and
	// is left out of the results. A call the client cancelled is no error
	// and is not reported. A broken page is reported again by every call
	// that searches or lists it or asks for it by its path, so a hook
	// writing to a log should deduplicate or rate-limit. Nil ignores them.
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
	Query    string `json:"query" jsonschema:"the words to search for in the title, description and text of the pages; a page has to contain every word; at most 16 words and 256 bytes"`
	Language string `json:"language,omitempty" jsonschema:"only return pages in this BCP 47 language, like en or de"`
	Limit    int    `json:"limit,omitempty" jsonschema:"the maximum number of results: 10 when not given, at most 25"`
}

type searchPagesOutput struct {
	Results []PageInfo `json:"results" jsonschema:"the matching pages, best match first"`
}

type listPagesInput struct {
	Language string `json:"language,omitempty" jsonschema:"only list pages in this BCP 47 language, like en or de"`
	Limit    int    `json:"limit,omitempty" jsonschema:"the maximum number of pages: 50 when not given, at most 200"`
	Cursor   string `json:"cursor,omitempty" jsonschema:"the next_cursor of the previous call, to list the pages after the ones it returned"`
}

type listPagesOutput struct {
	Pages      []PageInfo `json:"pages" jsonschema:"the public pages, in the order of the site"`
	NextCursor string     `json:"next_cursor,omitempty" jsonschema:"set if there are more pages; pass it as the cursor of another call to list them"`
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
//     results unless asked for another number, at most [MaxSearchLimit], and
//     refuses a query beyond [MaxQueryBytes] or [MaxQueryWords].
//   - list_pages(language?, limit?, cursor?) lists them in source order,
//     [DefaultListLimit] at a time unless asked for another number, at most
//     [MaxListLimit], with a next_cursor to pass to the call listing the
//     next ones. The cursor is the position in the list, so a page added or
//     removed between two calls moves the pages after it by one.
//   - get_page(path) returns the Markdown alternate of one page (see
//     [web.Site.MarkdownAlternate]), truncated at
//     PageToolsOptions.MaxResponseBytes.
//
// The language of search_pages and list_pages matches a page in it or in a
// regional variant of it, and is refused if it is no BCP 47 language tag.
//
// A page is public if it is [web.Page.Indexable] and has a Markdown source
// (see [web.Page.MarkdownSource]); no tool ever returns another one. Indexable
// is no access control, though: a page meant for logged-in readers only that
// search engines may still index, like one behind a paywall, has to be left
// out of the source passed here, or the tools hand its text to anyone. Its
// Markdown source is published as it is, HTML comments included, which an
// HTML renderer may drop from the page.
//
// The source is read with content on every call, so a site with an expensive
// source should pass a cached one. The site builds the URLs, and is usually
// the source as well. The markdown_url of a page and the note on a truncated
// one point at [web.Site.MarkdownAlternateURL], so the site should serve the
// alternates there, by setting web.Site.MarkdownAlternates for
// [web.Site.RegisterRoutes] or with [web.Site.ServeMarkdown].
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
		opts.Language = site.PageLang
	}
	tools := &pageTools{source: source, site: site, opts: opts}

	of := cmp.Or(site.Title, "the site")
	readOnly := &sdk.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true, OpenWorldHint: new(false)}
	sdk.AddTool(server, &sdk.Tool{
		Name:        "search_pages",
		Title:       "Search pages",
		Description: "Search the public pages of " + of + " for words in their title, description and text. Returns the best matches first, each with a snippet; read one with get_page.",
		Annotations: readOnly,
	}, recovered(tools, tools.searchPages))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_pages",
		Title:       "List pages",
		Description: "List the public pages of " + of + ", optionally only those in one language. If next_cursor is set, there are more: call again with it as the cursor. Read one with get_page.",
		Annotations: readOnly,
	}, recovered(tools, tools.listPages))
	sdk.AddTool(server, &sdk.Tool{
		Name:        "get_page",
		Title:       "Get page",
		Description: "Get the content of a public page of " + of + " as Markdown, starting with its title, description, URL, language and translations.",
		Annotations: readOnly,
	}, recovered(tools, tools.getPage))
	return nil
}

// recovered wraps a tool handler so that a panic in it — a bug in the source,
// the Language hook or a page — fails the call as unavailable and is reported
// to OnError, instead of crashing the process. The SDK calls tool handlers on
// goroutines of its own, where the per-request recover of net/http does not
// reach, so without this one page that makes the code panic would take the
// whole server down on every call that reads it.
func recovered[In, Out any](t *pageTools, handler sdk.ToolHandlerFor[In, Out]) sdk.ToolHandlerFor[In, Out] {
	return func(ctx context.Context, request *sdk.CallToolRequest, input In) (result *sdk.CallToolResult, output Out, err error) {
		defer func() {
			if p := recover(); p != nil {
				var zero Out
				result, output, err = nil, zero, t.unavailable(ctx, errs.Errorf("mcp page tool panicked: %v", p))
			}
		}()
		return handler(ctx, request, input)
	}
}

// pageTools implements the tools of [AddPageTools].
type pageTools struct {
	source web.PageSource
	site   *web.Site
	opts   PageToolsOptions
}

// publicPage is a page the tools may return, with what they need of it.
type publicPage struct {
	page *web.Page
	info PageInfo
}

// text returns the Markdown source of the page: string content as it is,
// []byte content copied. It is only read by search_pages, so the other tools
// copy no content at all.
func (p publicPage) text() string {
	switch content := p.page.Content.(type) {
	case string:
		return content
	case []byte:
		return string(content)
	}
	return ""
}

// lowerText returns the Markdown source in lower case, as strings.ToLower
// does: string content in lower case already is not copied at all, and
// []byte content is copied once, not into a string first and lowered again.
func (p publicPage) lowerText() string {
	switch content := p.page.Content.(type) {
	case string:
		return strings.ToLower(content)
	case []byte:
		var b strings.Builder
		b.Grow(len(content))
		for len(content) > 0 {
			if c := content[0]; c < utf8.RuneSelf {
				if 'A' <= c && c <= 'Z' {
					c += 'a' - 'A'
				}
				b.WriteByte(c)
				content = content[1:]
				continue
			}
			// An invalid byte becomes U+FFFD, as in strings.ToLower.
			r, size := utf8.DecodeRune(content)
			b.WriteRune(unicode.ToLower(r))
			content = content[size:]
		}
		return b.String()
	}
	return ""
}

// sourcePages iterates the pages of the source, read with content, ending
// with the error of a failing source or of ctx. A call the client gave up on
// stops between pages, rather than reading and searching the rest of the
// source for no one.
func (t *pageTools) sourcePages(ctx context.Context) iter.Seq2[*web.Page, error] {
	return func(yield func(*web.Page, error) bool) {
		for page, err := range t.source.Pages(ctx, true) {
			if err != nil {
				yield(nil, err)
				return
			}
			if err := ctx.Err(); err != nil {
				yield(nil, err)
				return
			}
			if !yield(page, nil) {
				return
			}
		}
	}
}

// publicPage returns page as a page the tools may return, with ok false if
// it is not public, does not match the language filter (see
// [languageFilter]) if that is not empty, or if its URLs or Markdown
// alternate cannot be built. Such a broken page is reported to OnError and
// left out, so that one broken page does not take every other page offline
// with it, and so that no tool lists a page get_page cannot return.
func (t *pageTools) publicPage(ctx context.Context, page *web.Page, language string) (publicPage, bool) {
	if !page.Indexable() || !page.HasMarkdownSource() {
		return publicPage{}, false
	}
	info, err := t.pageInfo(page)
	if err != nil {
		t.reportError(ctx, err)
		return publicPage{}, false
	}
	if !matchesLanguage(info.Language, language) {
		return publicPage{}, false
	}
	return publicPage{page: page, info: info}, true
}

// publicPages iterates the public pages of the source in source order (see
// [pageTools.publicPage]).
func (t *pageTools) publicPages(ctx context.Context, language string) iter.Seq2[publicPage, error] {
	return func(yield func(publicPage, error) bool) {
		for page, err := range t.sourcePages(ctx) {
			if err != nil {
				yield(publicPage{}, err)
				return
			}
			public, ok := t.publicPage(ctx, page, language)
			if !ok {
				continue
			}
			if !yield(public, nil) {
				return
			}
		}
	}
}

// pageInfo returns the description of a page in the tool results.
func (t *pageTools) pageInfo(page *web.Page) (PageInfo, error) {
	// The Markdown URL is only built for a page whose Markdown alternate can
	// be built, which is what fails for invalid alternates or no language.
	markdownURL, err := t.site.MarkdownAlternateURL(page)
	if err != nil {
		return PageInfo{}, err
	}
	canonical, err := t.site.PageURL(page)
	if err != nil {
		return PageInfo{}, err
	}
	info := PageInfo{
		Title:       page.Title,
		Description: page.Description,
		URL:         canonical,
		MarkdownURL: markdownURL,
		Language:    t.opts.Language(page),
	}
	if lastUpdated := page.LastModified(); !lastUpdated.IsZero() {
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

// unavailable reports err and returns the error the client sees instead. A
// call the client cancelled is no failure of the server to report.
func (t *pageTools) unavailable(ctx context.Context, err error) error {
	if ctxErr := ctx.Err(); ctxErr != nil && errors.Is(err, ctxErr) {
		return ctxErr
	}
	t.reportError(ctx, err)
	return errUnavailable
}

// queryWords returns the distinct lower-cased words of a search_pages query,
// or an error the client sees if the query is empty or too long to search.
func queryWords(query string) ([]string, error) {
	if len(query) > MaxQueryBytes {
		return nil, fmt.Errorf("the query is longer than %d bytes", MaxQueryBytes)
	}
	var words []string
	for word := range strings.FieldsSeq(strings.ToLower(query)) {
		if !slices.Contains(words, word) {
			words = append(words, word)
		}
	}
	switch {
	case len(words) == 0:
		return nil, errors.New("the query is empty")
	case len(words) > MaxQueryWords:
		return nil, fmt.Errorf("the query has more than %d words", MaxQueryWords)
	}
	return words, nil
}

func (t *pageTools) searchPages(ctx context.Context, _ *sdk.CallToolRequest, input searchPagesInput) (*sdk.CallToolResult, searchPagesOutput, error) {
	words, err := queryWords(input.Query)
	if err != nil {
		return nil, searchPagesOutput{}, err
	}
	language, err := languageFilter(input.Language)
	if err != nil {
		return nil, searchPagesOutput{}, err
	}
	limit := input.Limit
	if limit <= 0 {
		limit = DefaultSearchLimit
	}
	limit = min(limit, MaxSearchLimit)

	type match struct {
		page  publicPage
		score int
		at    int
	}
	var matches []match
	for page, err := range t.publicPages(ctx, language) {
		if err != nil {
			return nil, searchPagesOutput{}, t.unavailable(ctx, err)
		}
		if score, at := searchPage(page, words); score > 0 {
			matches = append(matches, match{page: page, score: score, at: at})
		}
	}
	// Stable, so that equally good matches stay in source order.
	slices.SortStableFunc(matches, func(a, b match) int { return cmp.Compare(b.score, a.score) })
	matches = matches[:min(len(matches), limit)]
	output := searchPagesOutput{Results: make([]PageInfo, len(matches))}
	// Snippets are only cut for the matches that are returned.
	for i, m := range matches {
		output.Results[i] = m.page.info
		if m.at < 0 {
			output.Results[i].Snippet = oneLine(m.page.info.Description)
		} else {
			output.Results[i].Snippet = excerpt(m.page.text(), m.at)
		}
	}
	return nil, output, nil
}

// searchPage scores how well the page matches the lower-cased words — zero if
// one of them is missing from the page, see [titleScore] for the weights —
// and returns the rune index of the first match in its Markdown source, or -1
// if only the title and description match.
func searchPage(page publicPage, words []string) (score, at int) {
	title := strings.ToLower(page.info.Title)
	description := strings.ToLower(page.info.Description)
	text := page.lowerText()
	first := -1
	for _, word := range words {
		wordScore := 0
		if strings.Contains(title, word) {
			wordScore += titleScore
		}
		if strings.Contains(description, word) {
			wordScore += descriptionScore
		}
		if i := strings.Index(text, word); i >= 0 {
			// Counted from the first match on rather than from the start.
			wordScore += min(1+strings.Count(text[i+len(word):], word), maxTextScore)
			if first < 0 || i < first {
				first = i
			}
		}
		if wordScore == 0 {
			return 0, -1
		}
		score += wordScore
	}
	if first < 0 {
		return score, -1
	}
	// Lower-casing maps every rune to exactly one rune, but not always to one
	// of the same byte length, so the offset is returned as a rune index.
	return score, utf8.RuneCountInString(text[:first])
}

// excerpt returns about [snippetRunes] runes of text around the rune index
// center, on one line, with an ellipsis where it cuts the text.
func excerpt(text string, center int) string {
	first := max(0, center-snippetRunes/3)
	start, end := 0, len(text)
	runes := 0
	for i := range text {
		if runes == first {
			start = i
		}
		if runes == first+snippetRunes {
			end = i
			break
		}
		runes++
	}
	snippet := oneLine(text[start:end])
	if start > 0 {
		snippet = "…" + snippet
	}
	if end < len(text) {
		snippet += "…"
	}
	return snippet
}

func (t *pageTools) listPages(ctx context.Context, _ *sdk.CallToolRequest, input listPagesInput) (*sdk.CallToolResult, listPagesOutput, error) {
	output := listPagesOutput{Pages: []PageInfo{}}
	language, err := languageFilter(input.Language)
	if err != nil {
		return nil, listPagesOutput{}, err
	}
	offset := 0
	if input.Cursor != "" {
		// The cursor is the position of the first page to list, bounded so
		// that adding the limit cannot overflow.
		offset, err = strconv.Atoi(input.Cursor)
		if err != nil || offset < 0 || offset > math.MaxInt32 {
			return nil, listPagesOutput{}, errors.New("the cursor is not a next_cursor list_pages returned")
		}
	}
	limit := input.Limit
	if limit <= 0 {
		limit = DefaultListLimit
	}
	limit = min(limit, MaxListLimit)

	index := 0 // of the next public page in the list
	more := false
	for page, err := range t.sourcePages(ctx) {
		if err != nil {
			return nil, listPagesOutput{}, t.unavailable(ctx, err)
		}
		// The rest of the source is still read after the last page listed,
		// so that a source failing half way fails every call alike, but
		// nothing more is built for it.
		if more {
			continue
		}
		// The pages before the cursor are checked like the listed ones, so
		// that the positions count the same pages in every call.
		public, ok := t.publicPage(ctx, page, language)
		if !ok {
			continue
		}
		switch {
		case index >= offset+limit:
			more = true
		case index >= offset:
			output.Pages = append(output.Pages, public.info)
		}
		index++
	}
	if more {
		output.NextCursor = strconv.Itoa(offset + limit)
	}
	return nil, output, nil
}

func (t *pageTools) getPage(ctx context.Context, _ *sdk.CallToolRequest, input getPageInput) (*sdk.CallToolResult, getPageOutput, error) {
	if input.Path == "" {
		return nil, getPageOutput{}, errors.New("the path is empty")
	}
	candidates := t.requestedPaths(input.Path)
	var found *publicPage
	for page, err := range t.sourcePages(ctx) {
		if err != nil {
			return nil, getPageOutput{}, t.unavailable(ctx, err)
		}
		// The rest of the source is still read after a match, so that a
		// source failing half way fails every call alike instead of only the
		// ones for pages behind the failure. Only a page whose path can be
		// the one asked for is checked and gets its URLs built.
		if found != nil || !slices.Contains(candidates, decodedPath(page.Path)) {
			continue
		}
		public, ok := t.publicPage(ctx, page, "")
		if !ok {
			continue
		}
		markdownPath, _ := web.MarkdownPath(page.Path)
		if input.Path == page.Path || input.Path == markdownPath ||
			input.Path == public.info.URL || input.Path == public.info.MarkdownURL {
			found = &public
		}
	}
	if found == nil {
		return nil, getPageOutput{}, ErrPageNotFound
	}
	doc, err := t.site.MarkdownAlternate(found.page)
	if err != nil {
		return nil, getPageOutput{}, t.unavailable(ctx, err)
	}
	// The header names the language the other tools report the page in,
	// which a PageToolsOptions.Language of the site's own may decide.
	doc.Lang = found.info.Language
	markdown, truncated := truncate(doc.Bytes(), t.opts.MaxResponseBytes, found.info.MarkdownURL)
	// The Markdown goes to the model as text, the metadata as structured
	// content, so the page is not sent twice — once as text and once
	// JSON-escaped — as it would be with the Markdown in the output struct.
	result := &sdk.CallToolResult{Content: []sdk.Content{&sdk.TextContent{Text: string(markdown)}}}
	return result, getPageOutput{Page: found.info, Truncated: truncated}, nil
}

// requestedPaths returns the decoded page paths the path argument of
// get_page can name: the argument itself, below the BaseURL if it is a URL
// of the site, and if that is a Markdown path also the page path it was
// derived from (see [web.MarkdownPath]). It is worked out once per call, so
// that each page of the source is only compared by its path, and a page
// passing that comparison is still matched exactly against the addresses
// the other tools return.
func (t *pageTools) requestedPaths(arg string) []string {
	p := arg
	if root, err := t.site.URL("/"); err == nil {
		if rest, ok := strings.CutPrefix(arg, strings.TrimSuffix(root, "/")); ok && strings.HasPrefix(rest, "/") {
			p = rest
		}
	}
	decoded, err := url.PathUnescape(p)
	if err != nil {
		// No page path has an invalid escape.
		return nil
	}
	if dir, ok := strings.CutSuffix(decoded, "/index.md"); ok {
		return []string{decoded, dir + "/"}
	}
	if page, ok := strings.CutSuffix(decoded, ".md"); ok {
		return []string{decoded, page}
	}
	return []string{decoded}
}

// decodedPath returns a page path with its percent escapes decoded, the way
// requestedPaths decodes the argument of get_page.
func decodedPath(p string) string {
	if !strings.Contains(p, "%") {
		return p
	}
	decoded, err := url.PathUnescape(p)
	if err != nil {
		return p
	}
	return decoded
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

// languageFilter returns the language argument of a tool call lower-cased for
// [matchesLanguage], or an error the client sees if it is no language tag.
// It is checked once per call, because it is compared with every page.
func languageFilter(language string) (string, error) {
	if language == "" {
		return "", nil
	}
	if len(language) > maxLanguageBytes || !html.ValidLangTag(language) {
		return "", errors.New("the language is not a BCP 47 language tag")
	}
	return strings.ToLower(language), nil
}

// matchesLanguage reports whether a page in language belongs to the results
// for the lower-cased filter: every page for an empty filter, else a page in
// the language itself or in a regional variant of it, so "en" finds "en-GB"
// too. Language tags are case-insensitive.
func matchesLanguage(language, filter string) bool {
	if filter == "" {
		return true
	}
	language = strings.ToLower(language)
	return strings.HasPrefix(language, filter) && (len(language) == len(filter) || language[len(filter)] == '-')
}

// oneLine collapses every run of whitespace in s, line breaks included, into
// one space.
func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
