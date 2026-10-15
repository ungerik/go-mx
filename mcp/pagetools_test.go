package mcp_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"strings"
	"sync"
	"testing"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"

	"github.com/ungerik/go-mx"
	mxmcp "github.com/ungerik/go-mx/mcp"
	"github.com/ungerik/go-mx/web"
)

var (
	published = time.Date(2026, 6, 1, 8, 0, 0, 0, time.UTC)
	updated   = time.Date(2026, 8, 20, 17, 15, 0, 0, time.UTC)
)

// pages is a PageSource yielding copies of a fixed slice of pages, with the
// content only when asked for it, like a real source.
type pages []web.Page

func (p pages) Pages(ctx context.Context, withContent bool) iter.Seq2[*web.Page, error] {
	return func(yield func(*web.Page, error) bool) {
		for _, page := range p {
			if !withContent {
				page.Content = nil
			}
			if !yield(&page, nil) {
				return
			}
		}
	}
}

func markdownPage(path, title, description, content string) web.Page {
	return web.Page{
		Path:        path,
		Title:       title,
		Description: description,
		Published:   published,
		ContentType: mx.ContentTypeMarkdown,
		Content:     content,
	}
}

func testSite() *web.Site {
	return &web.Site{BaseURL: "https://example.com", Title: "Example", Lang: "en"}
}

func testPages() pages {
	faq := markdownPage("/en/faq", "FAQ", "Questions people ask about pricing", "## Pricing\n\nAll prices include VAT.\n")
	faq.LastUpdated = updated
	faq.Alternates = []web.Alternate{{HrefLang: "de", Path: "/de/faq"}}
	faqDE := markdownPage("/de/faq", "Häufige Fragen", "Fragen zu Preisen", "## Preise\n\nAlle Preise inklusive MwSt.\n")
	faqDE.Lang = "de"
	faqDE.Alternates = []web.Alternate{{HrefLang: "en", Path: "/en/faq"}}

	draft := markdownPage("/en/draft", "Pricing draft", "", "pricing pricing pricing")
	draft.Published = time.Time{}
	hidden := markdownPage("/en/hidden", "Hidden pricing", "", "pricing")
	hidden.NoIndex = true

	return pages{
		faq,
		faqDE,
		markdownPage("/en/about", "About us", "Who we are", "We have been selling things since 2020. Our pricing is fair."),
		markdownPage("/en-GB/about", "About us", "Who we are, in British English", "Colour and pricing."),
		markdownPage("/blog/hello", "Hello", "", "The first post."),
		markdownPage("/go/tour", "Go tour", "", "A tour of Go."),
		// Never returned: not public, or no Markdown alternate.
		draft,
		hidden,
		{Path: "/en/html", Title: "HTML pricing", Published: published, ContentType: mx.ContentTypeHTML, Content: "<p>pricing</p>"},
	}
}

// connect returns a client session to a server carrying the page tools.
func connect(t *testing.T, source web.PageSource, site *web.Site, opts mxmcp.PageToolsOptions) *sdk.ClientSession {
	t.Helper()
	ctx := context.Background()
	server := sdk.NewServer(&sdk.Implementation{Name: "test-server", Version: "v0.0.1"}, nil)
	require.NoError(t, mxmcp.AddPageTools(server, source, site, opts))

	serverTransport, clientTransport := sdk.NewInMemoryTransports()
	serverSession, err := server.Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	t.Cleanup(func() { serverSession.Close() })

	client := sdk.NewClient(&sdk.Implementation{Name: "test-client", Version: "v0.0.1"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	t.Cleanup(func() { session.Close() })
	return session
}

// call calls a tool and decodes its structured content into output, returning
// the result for its text content and error state.
func call(t *testing.T, session *sdk.ClientSession, tool string, arguments map[string]any, output any) *sdk.CallToolResult {
	t.Helper()
	result, err := session.CallTool(context.Background(), &sdk.CallToolParams{Name: tool, Arguments: arguments})
	require.NoError(t, err)
	if output != nil && !result.IsError {
		data, err := json.Marshal(result.StructuredContent)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(data, output))
	}
	return result
}

// text returns the text content of a tool result.
func text(t *testing.T, result *sdk.CallToolResult) string {
	t.Helper()
	require.Len(t, result.Content, 1)
	content, ok := result.Content[0].(*sdk.TextContent)
	require.True(t, ok, "content is %T", result.Content[0])
	return content.Text
}

type searchOutput struct {
	Results []mxmcp.PageInfo `json:"results"`
}

type listOutput struct {
	Pages []mxmcp.PageInfo `json:"pages"`
}

type getOutput struct {
	Page      mxmcp.PageInfo `json:"page"`
	Truncated bool           `json:"truncated"`
}

func urls(infos []mxmcp.PageInfo) []string {
	var u []string
	for _, info := range infos {
		u = append(u, info.URL)
	}
	return u
}

func TestAddPageToolsListsExactlyTheReadOnlyTools(t *testing.T) {
	session := connect(t, testPages(), testSite(), mxmcp.PageToolsOptions{})
	result, err := session.ListTools(context.Background(), nil)
	require.NoError(t, err)

	var names []string
	for _, tool := range result.Tools {
		names = append(names, tool.Name)
		// Exposing a site to agents must never hand them a way to change
		// it: every tool declares itself read-only.
		require.NotNil(t, tool.Annotations, tool.Name)
		require.True(t, tool.Annotations.ReadOnlyHint, tool.Name)
		require.Contains(t, tool.Description, "Example", tool.Name)
	}
	require.ElementsMatch(t, []string{"search_pages", "list_pages", "get_page"}, names)
}

func TestAddPageToolsErrors(t *testing.T) {
	server := sdk.NewServer(&sdk.Implementation{Name: "test-server", Version: "v0.0.1"}, nil)

	// Exposure is the caller's explicit act of passing a source.
	err := mxmcp.AddPageTools(server, nil, testSite(), mxmcp.PageToolsOptions{})
	require.ErrorContains(t, err, "source is nil")

	err = mxmcp.AddPageTools(nil, testPages(), testSite(), mxmcp.PageToolsOptions{})
	require.ErrorContains(t, err, "server is nil")

	err = mxmcp.AddPageTools(server, testPages(), nil, mxmcp.PageToolsOptions{})
	require.ErrorContains(t, err, "site is nil")

	err = mxmcp.AddPageTools(server, testPages(), &web.Site{BaseURL: "example.com"}, mxmcp.PageToolsOptions{})
	require.ErrorContains(t, err, "BaseURL")

	err = mxmcp.AddPageTools(server, testPages(), testSite(), mxmcp.PageToolsOptions{MaxResponseBytes: -1})
	require.ErrorContains(t, err, "MaxResponseBytes")

	// None of the failed calls registered anything.
	serverTransport, clientTransport := sdk.NewInMemoryTransports()
	serverSession, err := server.Connect(context.Background(), serverTransport, nil)
	require.NoError(t, err)
	defer serverSession.Close()
	session, err := sdk.NewClient(&sdk.Implementation{Name: "c", Version: "v0"}, nil).Connect(context.Background(), clientTransport, nil)
	require.NoError(t, err)
	defer session.Close()
	result, err := session.ListTools(context.Background(), nil)
	require.NoError(t, err)
	require.Empty(t, result.Tools)
}

func TestSearchPages(t *testing.T) {
	session := connect(t, testPages(), testSite(), mxmcp.PageToolsOptions{})

	var output searchOutput
	result := call(t, session, "search_pages", map[string]any{"query": "PRICING"}, &output)
	require.False(t, result.IsError, text(t, result))
	// The title and description outweigh a mention in the text. The draft,
	// the NoIndex page and the HTML page match too but are never returned.
	require.Equal(t,
		[]string{"https://example.com/en/faq", "https://example.com/en/about", "https://example.com/en-GB/about"},
		urls(output.Results),
	)
	faq := output.Results[0]
	require.Equal(t, mxmcp.PageInfo{
		Title:       "FAQ",
		Description: "Questions people ask about pricing",
		URL:         "https://example.com/en/faq",
		MarkdownURL: "https://example.com/en/faq.md",
		Language:    "en",
		LastUpdated: "2026-08-20T17:15:00Z",
		Snippet:     "## Pricing All prices include VAT.",
	}, faq)
	require.Equal(t, "2026-06-01T08:00:00Z", output.Results[1].LastUpdated, "falls back to Published")

	t.Run("every word has to match", func(t *testing.T) {
		var output searchOutput
		call(t, session, "search_pages", map[string]any{"query": "pricing fair"}, &output)
		require.Equal(t, []string{"https://example.com/en/about"}, urls(output.Results))
	})

	t.Run("snippet from the description", func(t *testing.T) {
		var output searchOutput
		call(t, session, "search_pages", map[string]any{"query": "who"}, &output)
		require.Equal(t, "Who we are", output.Results[0].Snippet)
	})

	t.Run("language", func(t *testing.T) {
		var output searchOutput
		call(t, session, "search_pages", map[string]any{"query": "preise", "language": "de"}, &output)
		require.Equal(t, []string{"https://example.com/de/faq"}, urls(output.Results))
		require.Equal(t, "de", output.Results[0].Language)

		// A language finds its regional variants too.
		call(t, session, "search_pages", map[string]any{"query": "who", "language": "EN"}, &output)
		require.Equal(t, []string{"https://example.com/en/about", "https://example.com/en-GB/about"}, urls(output.Results))

		call(t, session, "search_pages", map[string]any{"query": "who", "language": "en-gb"}, &output)
		require.Equal(t, []string{"https://example.com/en-GB/about"}, urls(output.Results))
	})

	t.Run("no match", func(t *testing.T) {
		var output searchOutput
		result := call(t, session, "search_pages", map[string]any{"query": "nonexistent"}, &output)
		require.False(t, result.IsError)
		require.NotNil(t, output.Results)
		require.Empty(t, output.Results)
	})

	t.Run("empty query", func(t *testing.T) {
		result := call(t, session, "search_pages", map[string]any{"query": "  "}, nil)
		require.True(t, result.IsError)
		require.Equal(t, "the query is empty", text(t, result))
	})

	t.Run("missing query", func(t *testing.T) {
		result := call(t, session, "search_pages", map[string]any{}, nil)
		require.True(t, result.IsError)
	})
}

func TestSearchPagesLimit(t *testing.T) {
	var many pages
	for i := range 30 {
		many = append(many, markdownPage(fmt.Sprintf("/notes/%02d", i), fmt.Sprintf("Note %d", i), "", "a note"))
	}
	session := connect(t, many, testSite(), mxmcp.PageToolsOptions{})

	tests := []struct {
		limit any
		want  int
	}{
		{limit: nil, want: mxmcp.DefaultSearchLimit},
		{limit: 0, want: mxmcp.DefaultSearchLimit},
		{limit: -5, want: mxmcp.DefaultSearchLimit},
		{limit: 3, want: 3},
		{limit: 100, want: mxmcp.MaxSearchLimit},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprint(tt.limit), func(t *testing.T) {
			arguments := map[string]any{"query": "note"}
			if tt.limit != nil {
				arguments["limit"] = tt.limit
			}
			var output searchOutput
			call(t, session, "search_pages", arguments, &output)
			require.Len(t, output.Results, tt.want)
			// Equally good matches keep their source order.
			require.Equal(t, "https://example.com/notes/00", output.Results[0].URL)
		})
	}
}

func TestSearchPagesSnippet(t *testing.T) {
	long := strings.Repeat("filler words ", 40) + "the İstanbul office NEEDLE is here " + strings.Repeat("more text ", 40)
	session := connect(t, pages{markdownPage("/long", "Long", "", long)}, testSite(), mxmcp.PageToolsOptions{})

	var output searchOutput
	call(t, session, "search_pages", map[string]any{"query": "needle"}, &output)
	require.Len(t, output.Results, 1)
	snippet := output.Results[0].Snippet
	// The snippet is cut around the match from the original text, even though
	// lower-casing "İ" changes its length in bytes.
	require.Contains(t, snippet, "İstanbul office NEEDLE is here")
	require.True(t, strings.HasPrefix(snippet, "…"), snippet)
	require.True(t, strings.HasSuffix(snippet, "…"), snippet)
	require.LessOrEqual(t, len([]rune(snippet)), 170)
}

func TestListPages(t *testing.T) {
	session := connect(t, testPages(), testSite(), mxmcp.PageToolsOptions{})

	var output listOutput
	result := call(t, session, "list_pages", nil, &output)
	require.False(t, result.IsError)
	require.Equal(t,
		[]string{
			"https://example.com/en/faq",
			"https://example.com/de/faq",
			"https://example.com/en/about",
			"https://example.com/en-GB/about",
			"https://example.com/blog/hello",
			"https://example.com/go/tour",
		},
		urls(output.Pages),
	)
	for _, page := range output.Pages {
		require.Empty(t, page.Snippet)
	}
	// The language comes from Page.Lang, else a path segment that is a
	// language tag, else the site. Neither "blog" nor "go" is a language.
	require.Equal(t, []string{"en", "de", "en", "en-GB", "en", "en"}, languages(output.Pages))

	call(t, session, "list_pages", map[string]any{"language": "de"}, &output)
	require.Equal(t, []string{"https://example.com/de/faq"}, urls(output.Pages))

	call(t, session, "list_pages", map[string]any{"language": "fr"}, &output)
	require.NotNil(t, output.Pages)
	require.Empty(t, output.Pages)
}

func TestListPagesLanguageOption(t *testing.T) {
	session := connect(t, testPages(), testSite(), mxmcp.PageToolsOptions{
		Language: func(page *web.Page) string { return "x-" + page.Title },
	})
	var output listOutput
	call(t, session, "list_pages", map[string]any{"language": "x-FAQ"}, &output)
	require.Equal(t, []string{"https://example.com/en/faq"}, urls(output.Pages))
}

func TestGetPage(t *testing.T) {
	site := testSite()
	source := testPages()
	session := connect(t, source, site, mxmcp.PageToolsOptions{})

	want, err := site.MarkdownAlternate(&source[0])
	require.NoError(t, err)

	// The page is found by every address the other tools hand out.
	for _, path := range []string{"/en/faq", "/en/faq.md", "https://example.com/en/faq", "https://example.com/en/faq.md"} {
		t.Run(path, func(t *testing.T) {
			var output getOutput
			result := call(t, session, "get_page", map[string]any{"path": path}, &output)
			require.False(t, result.IsError, text(t, result))
			require.Equal(t, string(want.Bytes()), text(t, result))
			require.False(t, output.Truncated)
			require.Equal(t, "https://example.com/en/faq.md", output.Page.MarkdownURL)
			require.Equal(t, "FAQ", output.Page.Title)
		})
	}
}

func TestGetPageNotFound(t *testing.T) {
	session := connect(t, testPages(), testSite(), mxmcp.PageToolsOptions{})

	// A page that does not exist, one that is not public and one without a
	// Markdown alternate all get the same answer, so a client learns nothing
	// about what it cannot read.
	for _, path := range []string{"/en/nope", "/en/draft", "/en/hidden", "/en/html", "https://other.example/en/faq"} {
		t.Run(path, func(t *testing.T) {
			result := call(t, session, "get_page", map[string]any{"path": path}, nil)
			require.True(t, result.IsError)
			require.Equal(t, mxmcp.ErrPageNotFound.Error(), text(t, result))
		})
	}

	result := call(t, session, "get_page", map[string]any{"path": ""}, nil)
	require.True(t, result.IsError)
	require.Equal(t, "the path is empty", text(t, result))
}

func TestGetPageTruncated(t *testing.T) {
	page := markdownPage("/big", "Big", "", strings.Repeat("A line of text that goes on.\n", 1000))
	session := connect(t, pages{page}, testSite(), mxmcp.PageToolsOptions{MaxResponseBytes: 2000})

	var output getOutput
	result := call(t, session, "get_page", map[string]any{"path": "/big"}, &output)
	require.False(t, result.IsError)
	require.True(t, output.Truncated)
	markdown := text(t, result)
	require.LessOrEqual(t, len(markdown), 2000)
	require.True(t, strings.HasPrefix(markdown, "# Big\n"))
	require.Contains(t, markdown, "A line of text that goes on.\n\n[Truncated: the page has ")
	require.True(t, strings.HasSuffix(markdown, "The whole page is at https://example.com/big.md]\n"), markdown)
}

func TestGetPageTruncatedDefault(t *testing.T) {
	// One rune of three bytes after another, so that a cut at the byte limit
	// would land inside one.
	page := markdownPage("/big", "Big", "", strings.Repeat("€", mxmcp.DefaultMaxResponseBytes))
	session := connect(t, pages{page}, testSite(), mxmcp.PageToolsOptions{})

	var output getOutput
	result := call(t, session, "get_page", map[string]any{"path": "/big"}, &output)
	require.True(t, output.Truncated)
	markdown := text(t, result)
	require.LessOrEqual(t, len(markdown), mxmcp.DefaultMaxResponseBytes)
	require.True(t, strings.HasSuffix(markdown, "]\n"))
	require.NotContains(t, markdown, "�")
}

// errorSource yields one good page and then fails.
type errorSource struct{}

func (errorSource) Pages(ctx context.Context, withContent bool) iter.Seq2[*web.Page, error] {
	return func(yield func(*web.Page, error) bool) {
		page := markdownPage("/ok", "OK", "", "fine")
		if !yield(&page, nil) {
			return
		}
		yield(nil, errors.New("database password expired for user admin"))
	}
}

// errorRecorder collects the errors passed to PageToolsOptions.OnError.
type errorRecorder struct {
	mu     sync.Mutex
	errors []error
}

func (r *errorRecorder) onError(ctx context.Context, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.errors = append(r.errors, err)
}

func (r *errorRecorder) recorded() []error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.errors
}

func TestPageToolsHideInternalErrors(t *testing.T) {
	recorder := new(errorRecorder)
	session := connect(t, errorSource{}, testSite(), mxmcp.PageToolsOptions{OnError: recorder.onError})

	calls := []struct {
		tool      string
		arguments map[string]any
	}{
		{"search_pages", map[string]any{"query": "fine"}},
		{"list_pages", nil},
		{"get_page", map[string]any{"path": "/ok"}},
	}
	for _, c := range calls {
		t.Run(c.tool, func(t *testing.T) {
			result := call(t, session, c.tool, c.arguments, nil)
			require.True(t, result.IsError)
			message := text(t, result)
			require.NotContains(t, message, "password")
			require.Contains(t, message, "unavailable")
		})
	}
	// The details go to the server's log instead.
	require.Len(t, recorder.recorded(), len(calls))
	require.ErrorContains(t, recorder.recorded()[0], "password")
}

func TestPageToolsSkipBrokenPages(t *testing.T) {
	// A page whose Markdown URL cannot be built is reported and left out
	// rather than taking every other page offline with it.
	broken := markdownPage("/broken.md", "Broken", "", "text")
	recorder := new(errorRecorder)
	session := connect(t, pages{broken, markdownPage("/ok", "OK", "", "text")}, testSite(), mxmcp.PageToolsOptions{OnError: recorder.onError})

	var output listOutput
	result := call(t, session, "list_pages", nil, &output)
	require.False(t, result.IsError)
	require.Equal(t, []string{"https://example.com/ok"}, urls(output.Pages))
	require.Len(t, recorder.recorded(), 1)
	require.ErrorContains(t, recorder.recorded()[0], ".md")

	// Without OnError the error is dropped, the page still left out.
	session = connect(t, pages{broken}, testSite(), mxmcp.PageToolsOptions{})
	call(t, session, "list_pages", nil, &output)
	require.Empty(t, output.Pages)
}

func TestPageToolsLeaveOutPagesWithoutAlternate(t *testing.T) {
	// Regression: a page whose URLs could be built but whose Markdown
	// alternate could not was listed, and then every get_page for it failed.
	page := markdownPage("/en/faq", "FAQ", "", "text")
	page.Alternates = []web.Alternate{{HrefLang: "de", Path: "/de"}, {HrefLang: "de", Path: "/de-2"}}
	recorder := new(errorRecorder)
	session := connect(t, pages{page, markdownPage("/en/ok", "OK", "", "text")}, testSite(), mxmcp.PageToolsOptions{OnError: recorder.onError})

	var output listOutput
	call(t, session, "list_pages", nil, &output)
	require.Equal(t, []string{"https://example.com/en/ok"}, urls(output.Pages))
	require.Len(t, recorder.recorded(), 1)
	require.ErrorContains(t, recorder.recorded()[0], "repeats")

	result := call(t, session, "get_page", map[string]any{"path": "/en/faq"}, nil)
	require.True(t, result.IsError)
	require.Equal(t, mxmcp.ErrPageNotFound.Error(), text(t, result))
}

func languages(infos []mxmcp.PageInfo) []string {
	var l []string
	for _, info := range infos {
		l = append(l, info.Language)
	}
	return l
}
