package web_test

import (
	"context"
	"errors"
	"iter"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/web"
)

// testPages is a PageSource yielding a fixed slice of pages.
type testPages []*web.Page

func (p testPages) Pages(ctx context.Context, withContent bool) iter.Seq2[*web.Page, error] {
	return func(yield func(*web.Page, error) bool) {
		for _, page := range p {
			if !yield(page, nil) {
				return
			}
		}
	}
}

var (
	published = time.Date(2026, 6, 1, 8, 0, 0, 0, time.UTC)
	updated   = time.Date(2026, 8, 20, 17, 15, 0, 0, time.UTC)
	scheduled = time.Now().Add(24 * time.Hour)
)

func TestSiteURL(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		path    string
		want    string
	}{
		{name: "root", baseURL: "https://example.com", path: "/", want: "https://example.com/"},
		{name: "page", baseURL: "https://example.com", path: "/blog/hello", want: "https://example.com/blog/hello"},
		{name: "base with trailing slash", baseURL: "https://example.com/", path: "/blog", want: "https://example.com/blog"},
		{name: "base with path prefix", baseURL: "https://example.com/blog", path: "/hello", want: "https://example.com/blog/hello"},
		{
			// "/blog" and "/blog/" are different URLs, so a trailing slash
			// must survive into the canonical URL and the sitemap.
			name: "trailing slash preserved", baseURL: "https://example.com", path: "/blog/", want: "https://example.com/blog/",
		},
		{
			// The result has to be a usable URL even when the path was built
			// from a title that has not been slugified.
			name: "path escaped", baseURL: "https://example.com", path: "/a b", want: "https://example.com/a%20b",
		},
		{name: "empty base", baseURL: "", path: "/", want: ""},
		{name: "relative base", baseURL: "/blog", path: "/", want: ""},
		{name: "base with query", baseURL: "https://example.com?page=1", path: "/", want: ""},
		{name: "empty path", baseURL: "https://example.com", path: "", want: ""},
		{name: "relative path", baseURL: "https://example.com", path: "blog", want: ""},
		{
			// Resolving ".." would silently point the URL outside the site.
			name: "path with parent segment", baseURL: "https://example.com/blog", path: "/../etc", want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			site := &web.Site{BaseURL: tt.baseURL}
			u, err := site.URL(tt.path)
			if tt.want == "" {
				require.Error(t, err)
				require.Empty(t, u)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, u)
		})
	}
}

func TestSiteSitemap(t *testing.T) {
	site := &web.Site{
		BaseURL: "https://example.com",
		Sources: []web.PageSource{
			testPages{
				{Path: "/blog/second", Published: published, LastUpdated: updated},
				// Sorted after /blog/second although it is yielded first, so
				// that a regenerated sitemap only differs where the site did.
				{Path: "/", Published: published},
			},
			testPages{
				// Only pages that are meant to show up in search results
				// belong in a sitemap.
				{Path: "/blog/draft"},
				{Path: "/blog/scheduled", Published: scheduled},
				{Path: "/imprint", Published: published, NoIndex: true},
			},
		},
	}
	sitemap, err := site.Sitemap(context.Background())
	require.NoError(t, err)
	require.Equal(t,
		[]web.SitemapURL{
			{Loc: "https://example.com/", LastMod: published},
			{Loc: "https://example.com/blog/second", LastMod: updated},
		},
		sitemap.URLs,
	)
}

func TestSiteSitemapPageWithoutPath(t *testing.T) {
	// A published page that cannot be pointed at is a bug in the PageSource.
	// Reporting it beats shipping a sitemap that is quietly missing pages.
	site := &web.Site{
		BaseURL: "https://example.com",
		Sources: []web.PageSource{testPages{{Title: "Hello", Published: published}}},
	}
	_, err := site.Sitemap(context.Background())
	require.ErrorContains(t, err, "Hello")
}

func TestSiteRobotsTxt(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		// Without a Robots configuration the site still tells crawlers where
		// its sitemap is, which is the only way they find it without a
		// search console submission.
		site := &web.Site{BaseURL: "https://example.com"}
		robots, err := site.RobotsTxt()
		require.NoError(t, err)
		require.Equal(t,
			"User-agent: *\nDisallow:\n\nSitemap: https://example.com/sitemap.xml\n",
			robots.String(),
		)
	})

	t.Run("custom sitemap path", func(t *testing.T) {
		site := &web.Site{BaseURL: "https://example.com", SitemapPath: "/sitemap-index.xml"}
		robots, err := site.RobotsTxt()
		require.NoError(t, err)
		require.Contains(t, robots.String(), "Sitemap: https://example.com/sitemap-index.xml\n")
	})

	t.Run("configured robots keeps its rules", func(t *testing.T) {
		site := &web.Site{
			BaseURL: "https://example.com",
			Robots:  &web.Robots{Groups: []web.RobotsGroup{{Disallow: []string{"/admin/"}}}},
		}
		robots, err := site.RobotsTxt()
		require.NoError(t, err)
		require.Equal(t,
			"User-agent: *\nDisallow: /admin/\n\nSitemap: https://example.com/sitemap.xml\n",
			robots.String(),
		)

		// The Robots field is the site's configuration, not scratch space:
		// serving the robots.txt twice must not append the sitemap twice.
		robots, err = site.RobotsTxt()
		require.NoError(t, err)
		require.Equal(t,
			"User-agent: *\nDisallow: /admin/\n\nSitemap: https://example.com/sitemap.xml\n",
			robots.String(),
		)
		require.Empty(t, site.Robots.Sitemaps)
	})
}

func TestSiteRenderPage(t *testing.T) {
	site := &web.Site{
		BaseURL:     "https://example.com",
		Title:       "Example",
		Description: "A site about examples",
		Lang:        "de-AT",
		Author:      "Site Author",
	}
	page := &web.Page{
		Path:        "/blog/hello-world",
		Title:       "Hello World",
		Description: "The first post",
		Author:      "Page Author",
		Published:   published,
		LastUpdated: updated,
	}
	doc, err := site.RenderPage(context.Background(), page)
	require.NoError(t, err)

	require.Equal(t, "de-AT", doc.Lang)
	require.Equal(t, "Hello World", doc.Title)
	require.Equal(t,
		map[string]string{
			"description": "The first post",
			"author":      "Page Author",
		},
		doc.Meta,
	)
	require.Equal(t,
		map[string]string{
			"og:type":                "article",
			"og:title":               "Hello World",
			"og:site_name":           "Example",
			"og:description":         "The first post",
			"og:locale":              "de_AT",
			"og:url":                 "https://example.com/blog/hello-world",
			"article:published_time": "2026-06-01T08:00:00Z",
			"article:modified_time":  "2026-08-20T17:15:00Z",
			"article:author":         "Page Author",
		},
		doc.MetaProperty,
	)

	// The canonical link is what keeps the same page reachable under several
	// URLs from being indexed as duplicate content.
	require.Contains(t, renderDoc(t, &doc), `<link rel="canonical" href="https://example.com/blog/hello-world"/>`)
}

func TestSiteRenderPageFallbacks(t *testing.T) {
	site := &web.Site{
		BaseURL:     "https://example.com",
		Title:       "Example",
		Description: "A site about examples",
		Author:      "Site Author",
	}
	// A page without its own description and author falls back to the site's,
	// which is better than a page that shows up in search results with no
	// description at all.
	page := &web.Page{Path: "/", Title: "Home", Published: published}
	doc, err := site.RenderPage(context.Background(), page)
	require.NoError(t, err)

	require.Equal(t, "A site about examples", doc.Meta["description"])
	require.Equal(t, "Site Author", doc.Meta["author"])
	require.Equal(t, "A site about examples", doc.MetaProperty["og:description"])
}

func TestSiteRenderPageNotIndexable(t *testing.T) {
	site := &web.Site{BaseURL: "https://example.com"}
	tests := map[string]*web.Page{
		"draft":     {Path: "/draft", Title: "Draft"},
		"scheduled": {Path: "/scheduled", Title: "Scheduled", Published: scheduled},
		"noindex":   {Path: "/imprint", Title: "Imprint", Published: published, NoIndex: true},
	}
	for name, page := range tests {
		t.Run(name, func(t *testing.T) {
			doc, err := site.RenderPage(context.Background(), page)
			require.NoError(t, err)
			// A page that is not offered to crawlers in the sitemap can still
			// be found by following a link, so it has to say so itself.
			require.Equal(t, "noindex, nofollow", doc.Meta["robots"])
		})
	}
}

func TestSiteRenderPageKeepsRendererValues(t *testing.T) {
	site := &web.Site{
		BaseURL:     "https://example.com",
		Title:       "Example",
		Description: "A site about examples",
		Lang:        "en",
		// A renderer knows more about a page than the site-level defaults do,
		// so everything it sets survives.
		Renderer: web.PageRendererFunc(func(ctx context.Context, page *web.Page) (html.Document, error) {
			return html.Document{
				Lang:         "fr",
				Title:        "Renderer Title",
				Meta:         map[string]string{"description": "Renderer description"},
				MetaProperty: map[string]string{"og:type": "profile"},
				HeadCustom:   html.Link(html.Rel("alternate"), html.HRef("/feed.xml")),
			}, nil
		}),
	}
	page := &web.Page{Path: "/", Title: "Page Title", Description: "Page description", Published: published}
	doc, err := site.RenderPage(context.Background(), page)
	require.NoError(t, err)

	require.Equal(t, "fr", doc.Lang)
	require.Equal(t, "Renderer Title", doc.Title)
	require.Equal(t, "Renderer description", doc.Meta["description"])
	require.Equal(t, "profile", doc.MetaProperty["og:type"])
	// The article dates are not added to a document the renderer typed as
	// something other than an article.
	require.NotContains(t, doc.MetaProperty, "article:published_time")

	// Head content of the renderer is kept, the canonical link is added to it.
	rendered := renderDoc(t, &doc)
	require.Contains(t, rendered, `<link rel="alternate" href="/feed.xml"/>`)
	require.Contains(t, rendered, `<link rel="canonical" href="https://example.com/"/>`)
}

func TestSiteRenderPageWithoutPath(t *testing.T) {
	// A page rendered for a route that is not part of the site's page tree
	// has no canonical URL, which must not fail the render.
	site := &web.Site{BaseURL: "https://example.com", Title: "Example"}
	doc, err := site.RenderPage(context.Background(), &web.Page{Title: "Search results"})
	require.NoError(t, err)

	require.Equal(t, "Search results", doc.Title)
	require.NotContains(t, doc.MetaProperty, "og:url")
	require.NotContains(t, renderDoc(t, &doc), "canonical")
}

func TestSiteRegisterRoutes(t *testing.T) {
	site := &web.Site{
		BaseURL:     "https://example.com",
		SitemapPath: "/sitemap-posts.xml",
		Sources:     []web.PageSource{testPages{{Path: "/blog/hello", Published: published}}},
	}
	mux := http.NewServeMux()
	require.NoError(t, site.RegisterRoutes(mux))

	robotsResponse := httptest.NewRecorder()
	mux.ServeHTTP(robotsResponse, httptest.NewRequest(http.MethodGet, web.RobotsTxtPath, nil))
	require.Equal(t, http.StatusOK, robotsResponse.Code)
	require.Contains(t, robotsResponse.Body.String(), "Sitemap: https://example.com/sitemap-posts.xml")

	sitemapResponse := httptest.NewRecorder()
	mux.ServeHTTP(sitemapResponse, httptest.NewRequest(http.MethodGet, "/sitemap-posts.xml", nil))
	require.Equal(t, http.StatusOK, sitemapResponse.Code)
	require.Contains(t, sitemapResponse.Body.String(), "<loc>https://example.com/blog/hello</loc>")
}

func TestSiteRegisterRoutesInvalid(t *testing.T) {
	// Without a valid BaseURL the site cannot build a single absolute URL, so
	// registering its routes fails instead of serving a robots.txt and
	// sitemap that error on every request.
	mux := http.NewServeMux()
	require.Error(t, (&web.Site{}).RegisterRoutes(mux))
}

// renderDoc renders an html.Document to a string.
func renderDoc(t *testing.T, doc *html.Document) string {
	t.Helper()
	var b strings.Builder
	require.NoError(t, doc.Render(context.Background(), mx.NewCheckedWriter(&b)))
	return b.String()
}

func TestSiteValidate(t *testing.T) {
	// A sitemap path that is not absolute cannot be registered on a ServeMux
	// or joined to the BaseURL.
	site := &web.Site{BaseURL: "https://example.com", SitemapPath: "sitemap.xml"}
	require.ErrorContains(t, site.Validate(), "SitemapPath")

	// A robots.txt that does not validate would be served as a 500 on every
	// crawl, so it fails at configuration time instead.
	site = &web.Site{
		BaseURL: "https://example.com",
		Robots:  &web.Robots{Groups: []web.RobotsGroup{{UserAgents: []string{"Bad Bot"}}}},
	}
	require.Error(t, site.Validate())
	require.Error(t, site.RegisterRoutes(http.NewServeMux()))
}

// errorPages is a PageSource that fails while iterating.
type errorPages struct{}

func (errorPages) Pages(ctx context.Context, withContent bool) iter.Seq2[*web.Page, error] {
	return func(yield func(*web.Page, error) bool) {
		yield(nil, errors.New("source unavailable"))
	}
}

func TestSiteSitemapSourceError(t *testing.T) {
	// A source that fails half way would otherwise produce a sitemap that
	// silently drops every page after the failure.
	site := &web.Site{BaseURL: "https://example.com", Sources: []web.PageSource{errorPages{}}}
	_, err := site.Sitemap(context.Background())
	require.ErrorContains(t, err, "source unavailable")

	response := httptest.NewRecorder()
	site.HandleSitemapXML(response, httptest.NewRequest(http.MethodGet, web.DefaultSitemapPath, nil))
	require.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestSiteRenderPageRendererError(t *testing.T) {
	// The site adds metadata to a document, it does not paper over a renderer
	// that failed to produce one.
	site := &web.Site{
		BaseURL: "https://example.com",
		Renderer: web.PageRendererFunc(func(ctx context.Context, page *web.Page) (html.Document, error) {
			return html.Document{}, errors.New("template broken")
		}),
	}
	_, err := site.RenderPage(context.Background(), &web.Page{Path: "/"})
	require.ErrorContains(t, err, "template broken")
}

func TestSiteURLRejectsPathsJoinPathWouldMangle(t *testing.T) {
	site := &web.Site{BaseURL: "https://example.com"}

	// Regression: url.URL.JoinPath drops a path with an invalid percent escape
	// and returns the base URL unchanged, so "/100%-off" used to resolve to
	// "https://example.com" — the home page's canonical URL, and a duplicate
	// <loc> in the sitemap for two different pages.
	for _, path := range []string{"/100%-off", "/50%", "/a%zz"} {
		u, err := site.URL(path)
		require.Errorf(t, err, "path %q", path)
		require.Empty(t, u)
	}

	// A query or fragment is not part of a page path and JoinPath would escape
	// it into one ("/a?b" becomes "/a%3Fb").
	for _, path := range []string{"/a?b", "/a#b"} {
		_, err := site.URL(path)
		require.Errorf(t, err, "path %q", path)
	}

	// A percent escape that is valid still works.
	u, err := site.URL("/100%25-off")
	require.NoError(t, err)
	require.Equal(t, "https://example.com/100%25-off", u)
}

func TestSiteSitemapRoutePath(t *testing.T) {
	// Regression: the robots.txt advertised the sitemap below the BaseURL path
	// prefix while RegisterRoutes registered it at the root, so a crawler
	// following the advertised URL got a 404.
	site := &web.Site{BaseURL: "https://example.com/blog"}
	route, err := site.SitemapRoutePath()
	require.NoError(t, err)
	require.Equal(t, "/blog/sitemap.xml", route)

	robots, err := site.RobotsTxt()
	require.NoError(t, err)
	require.Contains(t, robots.String(), "Sitemap: https://example.com/blog/sitemap.xml")

	mux := http.NewServeMux()
	require.NoError(t, site.RegisterRoutes(mux))
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/blog/sitemap.xml", nil))
	require.Equal(t, http.StatusOK, response.Code)

	// A site at the root of a host keeps the rooted pattern a ServeMux needs.
	route, err = (&web.Site{BaseURL: "https://example.com"}).SitemapRoutePath()
	require.NoError(t, err)
	require.Equal(t, "/sitemap.xml", route)
}

func TestSiteValidateRejectsPanickingSitemapPaths(t *testing.T) {
	// Regression: RegisterRoutes has an error-returning contract, but these
	// paths reached http.ServeMux.HandleFunc and panicked there, taking down
	// startup instead of returning the error.
	tests := map[string]string{
		"robots.txt collision": web.RobotsTxtPath,
		"pattern wildcard":     "/sitemap/{id}",
		"unbalanced brace":     "/{",
		"whitespace":           "/sitemap .xml",
	}
	for name, sitemapPath := range tests {
		t.Run(name, func(t *testing.T) {
			site := &web.Site{BaseURL: "https://example.com", SitemapPath: sitemapPath}
			require.Error(t, site.Validate())
			require.Error(t, site.RegisterRoutes(http.NewServeMux()))
		})
	}
}

func TestSiteRenderPageDoesNotMutateRendererMaps(t *testing.T) {
	// Regression: AddPageMetadata wrote into the maps the renderer returned. A
	// renderer handing out a shared template document leaked one page's
	// metadata into every other page, and two concurrent requests writing the
	// same map crashed the process with "concurrent map writes".
	shared := html.Document{
		Meta:         map[string]string{"description": "Shared description"},
		MetaProperty: map[string]string{"og:type": "website"},
	}
	site := &web.Site{
		BaseURL: "https://example.com",
		Title:   "Example",
		Renderer: web.PageRendererFunc(func(ctx context.Context, page *web.Page) (html.Document, error) {
			return shared, nil
		}),
	}
	doc, err := site.RenderPage(context.Background(), &web.Page{
		Path: "/blog/hello", Title: "Hello", Published: published,
	})
	require.NoError(t, err)

	require.Equal(t, map[string]string{"description": "Shared description"}, shared.Meta)
	require.Equal(t, map[string]string{"og:type": "website"}, shared.MetaProperty)
	require.Equal(t, "https://example.com/blog/hello", doc.MetaProperty["og:url"])
}

func TestSiteAddPageMetadataIdempotent(t *testing.T) {
	// Adding the metadata twice must not produce two canonical links, which
	// would leave crawlers with contradicting canonical URLs.
	site := &web.Site{BaseURL: "https://example.com", Title: "Example"}
	page := &web.Page{Path: "/blog/hello", Title: "Hello", Published: published}
	doc, err := site.RenderPage(context.Background(), page)
	require.NoError(t, err)
	require.NoError(t, site.AddPageMetadata(&doc, page))

	rendered := renderDoc(t, &doc)
	require.Equal(t, 1, strings.Count(rendered, `rel="canonical"`))
}

func TestSiteURLRejectsEncodedTraversal(t *testing.T) {
	// Regression: the ".." check ran against the raw path, so an encoded
	// parent segment slipped through and could point a canonical URL outside
	// the site once something decoded and normalized the path.
	site := &web.Site{BaseURL: "https://example.com/blog"}
	for _, path := range []string{"/../admin", "/%2e%2e/admin", "/a/%2E%2E/b"} {
		_, err := site.URL(path)
		require.Errorf(t, err, "path %q", path)
	}

	// A file name that merely contains dots is not a traversal.
	u, err := site.URL("/release-1.2..3")
	require.NoError(t, err)
	require.Equal(t, "https://example.com/blog/release-1.2..3", u)
}

func TestSiteRegisterRoutesReturnsErrorInsteadOfPanicking(t *testing.T) {
	// Regression: a pattern the caller had already registered made
	// ServeMux.HandleFunc panic inside a method whose contract is to return an
	// error, taking down startup instead of being handled.
	mux := http.NewServeMux()
	mux.HandleFunc("GET "+web.RobotsTxtPath, func(http.ResponseWriter, *http.Request) {})

	site := &web.Site{BaseURL: "https://example.com"}
	err := site.RegisterRoutes(mux)
	require.ErrorContains(t, err, "registering the routes")
}

func TestSiteURLRejectsNonCanonicalPaths(t *testing.T) {
	// Regression: JoinPath cleans what it joins, so "/blog//hello" and
	// "/blog/./hello" both resolved to "/blog/hello". Three pages would have
	// claimed one canonical URL and produced one duplicated sitemap entry
	// while looking like three distinct pages in the source.
	site := &web.Site{BaseURL: "https://example.com"}
	for _, path := range []string{"/blog//hello", "/blog/./hello", "//hello", "/blog/hello/."} {
		_, err := site.URL(path)
		require.Errorf(t, err, "path %q", path)
	}

	// The canonical spellings still work, trailing slash and all.
	for path, want := range map[string]string{
		"/blog/hello":  "https://example.com/blog/hello",
		"/blog/hello/": "https://example.com/blog/hello/",
		"/":            "https://example.com/",
	} {
		u, err := site.URL(path)
		require.NoErrorf(t, err, "path %q", path)
		require.Equal(t, want, u)
	}
}
