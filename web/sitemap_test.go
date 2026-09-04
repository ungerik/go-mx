package web_test

import (
	"context"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/web"
	"github.com/ungerik/go-mx/xml"
)

func TestSitemapRender(t *testing.T) {
	sitemap := &web.Sitemap{
		URLs: []web.SitemapURL{
			{
				Loc:        "https://example.com/",
				LastMod:    time.Date(2026, 8, 25, 12, 30, 0, 0, time.UTC),
				ChangeFreq: web.ChangeFreqDaily,
				Priority:   1,
			},
			// Only Loc is required: every other element is left out rather
			// than rendered empty, which would fail schema validation.
			{Loc: "https://example.com/blog/hello-world"},
			// A query separator has to be escaped to keep the document
			// well-formed, which the mx writer does for the text of <loc>.
			{Loc: "https://example.com/search?q=a&b=c", Priority: 0.55},
		},
	}
	rendered, err := xml.String(sitemap)
	require.NoError(t, err)

	require.Equal(t,
		`<?xml version="1.0" encoding="UTF-8"?>`+"\n"+
			`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`+
			`<url>`+
			`<loc>https://example.com/</loc>`+
			`<lastmod>2026-08-25T12:30:00Z</lastmod>`+
			`<changefreq>daily</changefreq>`+
			`<priority>1.0</priority>`+
			`</url>`+
			`<url><loc>https://example.com/blog/hello-world</loc></url>`+
			`<url>`+
			`<loc>https://example.com/search?q=a&amp;b=c</loc>`+
			`<priority>0.55</priority>`+
			`</url>`+
			`</urlset>`,
		rendered,
	)
}

func TestSitemapValidate(t *testing.T) {
	tests := []struct {
		name  string
		url   web.SitemapURL
		valid bool
	}{
		{
			name:  "absolute URL",
			url:   web.SitemapURL{Loc: "https://example.com/"},
			valid: true,
		},
		{
			// A sitemap is fetched without a page context, so a relative
			// reference has nothing to resolve against.
			name:  "relative URL",
			url:   web.SitemapURL{Loc: "/blog/"},
			valid: false,
		},
		{
			name:  "empty URL",
			url:   web.SitemapURL{},
			valid: false,
		},
		{
			// A typo in a changefreq keyword is silently dropped by crawlers,
			// so the enum catches it while the value is still in Go.
			name:  "unknown change frequency",
			url:   web.SitemapURL{Loc: "https://example.com/", ChangeFreq: "often"},
			valid: false,
		},
		{
			name:  "priority above one",
			url:   web.SitemapURL{Loc: "https://example.com/", Priority: 1.5},
			valid: false,
		},
		{
			name:  "negative priority",
			url:   web.SitemapURL{Loc: "https://example.com/", Priority: -0.5},
			valid: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.url.Validate()
			if tt.valid {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)

			// An invalid entry must not render as markup at all: the error is
			// deferred to render time, not written out as a broken sitemap
			// that a crawler would silently reject.
			rendered, err := xml.String(&web.Sitemap{URLs: []web.SitemapURL{tt.url}})
			require.Error(t, err)
			require.Empty(t, rendered)
		})
	}
}

func TestSitemapHandleHTTP(t *testing.T) {
	sitemap := &web.Sitemap{URLs: []web.SitemapURL{{Loc: "https://example.com/"}}}
	response := httptest.NewRecorder()
	sitemap.HandleHTTP(response, httptest.NewRequest(http.MethodGet, web.DefaultSitemapPath, nil))

	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, mx.ContentTypeXML, response.Header().Get("Content-Type"))
	require.Equal(t,
		`<?xml version="1.0" encoding="UTF-8"?>`+"\n"+
			"<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n"+
			"  <url>\n"+
			"    <loc>https://example.com/</loc>\n"+
			"  </url>\n"+
			"</urlset>",
		response.Body.String(),
	)
}

func TestSitemapIndexRender(t *testing.T) {
	index := &web.SitemapIndex{
		Sitemaps: []web.SitemapIndexEntry{
			{
				Loc:     "https://example.com/sitemap-posts.xml",
				LastMod: time.Date(2026, 8, 25, 0, 0, 0, 0, time.UTC),
			},
			{Loc: "https://example.com/sitemap-pages.xml"},
		},
	}
	rendered, err := xml.String(index)
	require.NoError(t, err)

	require.Equal(t,
		`<?xml version="1.0" encoding="UTF-8"?>`+"\n"+
			`<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`+
			`<sitemap>`+
			`<loc>https://example.com/sitemap-posts.xml</loc>`+
			`<lastmod>2026-08-25T00:00:00Z</lastmod>`+
			`</sitemap>`+
			`<sitemap><loc>https://example.com/sitemap-pages.xml</loc></sitemap>`+
			`</sitemapindex>`,
		rendered,
	)

	// The index has the same requirements for its entry URLs as a sitemap.
	_, err = xml.String(&web.SitemapIndex{Sitemaps: []web.SitemapIndexEntry{{Loc: "sitemap.xml"}}})
	require.Error(t, err)
}

func TestChangeFreqEnum(t *testing.T) {
	// go-enum generated validation and value listing.
	require.True(t, web.ChangeFreqDaily.Valid())
	require.False(t, web.ChangeFreq("often").Valid())
	require.Error(t, web.ChangeFreq("often").Validate())
	require.Equal(t, "daily", web.ChangeFreqDaily.String())
	require.Equal(t,
		[]string{"always", "hourly", "daily", "weekly", "monthly", "yearly", "never"},
		web.ChangeFreq("").EnumStrings(),
	)
}

func TestSitemapMaxURLs(t *testing.T) {
	// A sitemap over the protocol limit is rejected as a whole by search
	// engines, so shipping one means losing every URL in it, not just the
	// excess. Split it across a SitemapIndex instead.
	sitemap := &web.Sitemap{URLs: make([]web.SitemapURL, web.MaxSitemapURLs+1)}
	require.ErrorContains(t, sitemap.Validate(), "50000")

	index := &web.SitemapIndex{Sitemaps: make([]web.SitemapIndexEntry, web.MaxSitemapURLs+1)}
	require.ErrorContains(t, index.Validate(), "50000")
}

func TestSitemapMaxLocLength(t *testing.T) {
	tooLong := "https://example.com/" + strings.Repeat("x", web.MaxSitemapLocLength)
	require.ErrorContains(t, web.SitemapURL{Loc: tooLong}.Validate(), "2048")
	require.ErrorContains(t, web.SitemapIndexEntry{Loc: tooLong}.Validate(), "2048")
}

func TestSitemapIndexHandleHTTP(t *testing.T) {
	index := &web.SitemapIndex{
		Sitemaps: []web.SitemapIndexEntry{{Loc: "https://example.com/sitemap-posts.xml"}},
	}
	response := httptest.NewRecorder()
	index.HandleHTTP(response, httptest.NewRequest(http.MethodGet, web.DefaultSitemapPath, nil))

	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, mx.ContentTypeXML, response.Header().Get("Content-Type"))
	require.Contains(t, response.Body.String(), "<loc>https://example.com/sitemap-posts.xml</loc>")

	// A deferred validation error surfaces as a 500 rather than a truncated
	// document that a crawler would try to parse.
	broken := &web.SitemapIndex{Sitemaps: []web.SitemapIndexEntry{{Loc: "sitemap.xml"}}}
	brokenResponse := httptest.NewRecorder()
	broken.HandleHTTP(brokenResponse, httptest.NewRequest(http.MethodGet, web.DefaultSitemapPath, nil))
	require.Equal(t, http.StatusInternalServerError, brokenResponse.Code)
}

func TestSitemapRejectsNaNPriority(t *testing.T) {
	// Regression: every comparison against NaN is false, so a NaN priority
	// passed the 0.0-to-1.0 range check and rendered as "<priority>NaN.0</priority>",
	// which invalidates the whole sitemap for a crawler.
	url := web.SitemapURL{Loc: "https://example.com/", Priority: math.NaN()}
	require.Error(t, url.Validate())

	rendered, err := xml.String(&web.Sitemap{URLs: []web.SitemapURL{url}})
	require.Error(t, err)
	require.Empty(t, rendered)
}

func TestSitemapRenderWritesNothingWhenInvalid(t *testing.T) {
	// Regression: the deferred error lived in the root element, so the XML
	// declaration was already written by the time rendering failed. A render
	// straight into a file or socket left a truncated document behind that
	// looked like a sitemap to whatever read it next.
	var b strings.Builder
	sitemap := &web.Sitemap{URLs: []web.SitemapURL{{Loc: "/relative"}}}
	require.Error(t, sitemap.Render(context.Background(), mx.NewCheckedWriter(&b)))
	require.Empty(t, b.String())

	b.Reset()
	index := &web.SitemapIndex{Sitemaps: []web.SitemapIndexEntry{{Loc: "/relative"}}}
	require.Error(t, index.Render(context.Background(), mx.NewCheckedWriter(&b)))
	require.Empty(t, b.String())
}

func TestSitemapMaxBytes(t *testing.T) {
	// The 50 MiB limit bites before the 50000 URL limit for long URLs, and a
	// sitemap over it is rejected as a whole by search engines.
	long := "https://example.com/" + strings.Repeat("x", web.MaxSitemapLocLength-len("https://example.com/"))
	urls := make([]web.SitemapURL, 30000)
	for i := range urls {
		urls[i] = web.SitemapURL{Loc: long}
	}
	sitemap := &web.Sitemap{URLs: urls}
	require.Less(t, len(sitemap.URLs), web.MaxSitemapURLs)
	require.ErrorContains(t, sitemap.Validate(), "bytes")
}

func TestSitemapDocumentWritesNothingWhenInvalid(t *testing.T) {
	// Regression: the invalid document still carried the XML declaration, so
	// rendering it wrote that declaration before the deferred error surfaced.
	var b strings.Builder
	sitemap := &web.Sitemap{URLs: []web.SitemapURL{{Loc: "/relative"}}}
	require.Error(t, sitemap.Document().Render(context.Background(), mx.NewCheckedWriter(&b)))
	require.Empty(t, b.String())

	b.Reset()
	index := &web.SitemapIndex{Sitemaps: []web.SitemapIndexEntry{{Loc: "/relative"}}}
	require.Error(t, index.Document().Render(context.Background(), mx.NewCheckedWriter(&b)))
	require.Empty(t, b.String())
}

func TestSitemapMaxBytesCountsEscaping(t *testing.T) {
	// Regression: the size estimate counted the raw URL, but "&" is written as
	// "&amp;" — five bytes, not one. A sitemap of query URLs passed the check
	// and still rendered far over the 50 MiB limit.
	loc := "https://example.com/?" + strings.Repeat("&", 1000)
	urls := make([]web.SitemapURL, 12000)
	for i := range urls {
		urls[i] = web.SitemapURL{Loc: loc}
	}
	require.ErrorContains(t, (&web.Sitemap{URLs: urls}).Validate(), "bytes")
}
