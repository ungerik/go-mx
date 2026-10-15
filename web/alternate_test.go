package web_test

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ungerik/go-mx/web"
)

func TestAlternateValidate(t *testing.T) {
	valid := map[string]web.Alternate{
		"translation":         {HrefLang: "de", Path: "/de/faq"},
		"region subtag":       {HrefLang: "de-AT", Path: "/de-at/faq"},
		"x-default":           {HrefLang: web.HrefLangDefault, Path: "/"},
		"format":              {Type: "text/markdown", Path: "/en/faq.md"},
		"format with charset": {Type: "text/markdown; charset=utf-8", Path: "/en/faq.md"},
		"explicit rel":        {Rel: web.RelAlternate, Type: "application/rss+xml", Path: "/feed.xml"},
		// A description of the page is told apart by its relation, so it
		// needs neither a type nor a language.
		"describedby": {Rel: web.RelDescribedBy, Path: "/en/faq.json"},
	}
	for name, alternate := range valid {
		t.Run(name, func(t *testing.T) {
			require.NoError(t, alternate.Validate())
		})
	}

	invalid := map[string]struct {
		alternate web.Alternate
		wantErr   string
	}{
		"unknown rel":           {web.Alternate{Rel: "canonical", HrefLang: "de", Path: "/de"}, "rel"},
		"not a media type":      {web.Alternate{Type: "markdown", Path: "/a.md"}, "not a media type"},
		"quote in type":         {web.Alternate{Type: `text/markdown; x="a"`, Path: "/a.md"}, "quote"},
		"line break in type":    {web.Alternate{Type: "text/markdown\r\nX: y", Path: "/a.md"}, "control"},
		"invalid hreflang":      {web.Alternate{HrefLang: "de AT", Path: "/de"}, "hreflang"},
		"quote in hreflang":     {web.Alternate{HrefLang: `de"`, Path: "/de"}, "hreflang"},
		"no type nor hreflang":  {web.Alternate{Path: "/other"}, "neither a type nor a hreflang"},
		"relative path":         {web.Alternate{HrefLang: "de", Path: "de/faq"}, "path"},
		"empty path":            {web.Alternate{HrefLang: "de", Path: ""}, "path"},
		"path with query":       {web.Alternate{HrefLang: "de", Path: "/faq?lang=de"}, "path"},
		"path outside the site": {web.Alternate{HrefLang: "de", Path: "/../de"}, "path"},
	}
	for name, tt := range invalid {
		t.Run(name, func(t *testing.T) {
			require.ErrorContains(t, tt.alternate.Validate(), tt.wantErr)
		})
	}
}

func TestAlternateRelOrDefault(t *testing.T) {
	require.Equal(t, web.RelAlternate, web.Alternate{}.RelOrDefault())
	require.Equal(t, web.RelDescribedBy, web.Alternate{Rel: web.RelDescribedBy}.RelOrDefault())
}

func TestAlternateIsTranslation(t *testing.T) {
	require.True(t, web.Alternate{HrefLang: "de", Path: "/de"}.IsTranslation())
	require.True(t, web.Alternate{HrefLang: web.HrefLangDefault, Path: "/"}.IsTranslation())
	// The German Markdown version is a format alternate: hreflang links
	// point at pages of the same type only.
	require.False(t, web.Alternate{HrefLang: "de", Type: "text/markdown", Path: "/de.md"}.IsTranslation())
	require.False(t, web.Alternate{Rel: web.RelDescribedBy, HrefLang: "de", Path: "/de.json"}.IsTranslation())
	require.False(t, web.Alternate{Type: "text/markdown", Path: "/en.md"}.IsTranslation())
}

// link is the expected form of one page link, rendered both as the <link>
// element AddPageMetadata adds and as the Link header PageHeaders returns.
type link struct{ rel, typ, hrefLang, url string }

func (l link) element() string {
	s := `<link rel="` + l.rel + `"`
	if l.typ != "" {
		s += ` type="` + l.typ + `"`
	}
	if l.hrefLang != "" {
		s += ` hreflang="` + l.hrefLang + `"`
	}
	return s + ` href="` + l.url + `"/>`
}

func (l link) header() string {
	s := "<" + l.url + `>; rel="` + l.rel + `"`
	if l.typ != "" {
		s += `; type="` + l.typ + `"`
	}
	if l.hrefLang != "" {
		s += `; hreflang="` + l.hrefLang + `"`
	}
	return s
}

var linkElementRegexp = regexp.MustCompile(`<link [^>]*>`)

// requirePageLinks asserts that the page renders exactly the want links, in
// order, both into the HTML head and as Link headers.
func requirePageLinks(t *testing.T, site *web.Site, page *web.Page, want ...link) {
	t.Helper()

	var wantElements, wantHeaders []string
	for _, l := range want {
		wantElements = append(wantElements, l.element())
		wantHeaders = append(wantHeaders, l.header())
	}

	doc, err := site.RenderPage(context.Background(), page)
	require.NoError(t, err)
	require.Equal(t, wantElements, linkElementRegexp.FindAllString(renderDoc(t, &doc), -1), "<link> elements")

	header, err := site.PageHeaders(page)
	require.NoError(t, err)
	require.Equal(t, wantHeaders, header.Values("Link"), "Link headers")
}

func TestSitePageLinks(t *testing.T) {
	const base = "https://example.com"

	t.Run("one language", func(t *testing.T) {
		// A page without translations has no hreflang links at all, not even
		// one naming its own language: there is no set to be part of.
		site := &web.Site{BaseURL: base, Lang: "en", DefaultLang: "en"}
		page := &web.Page{
			Path:       "/en/faq",
			Published:  published,
			Alternates: []web.Alternate{{Type: "text/markdown", Path: "/en/faq.md"}},
		}
		requirePageLinks(t, site, page,
			link{rel: "canonical", url: base + "/en/faq"},
			link{rel: "alternate", typ: "text/markdown", url: base + "/en/faq.md"},
		)
	})

	t.Run("three languages", func(t *testing.T) {
		site := &web.Site{BaseURL: base, Lang: "en", DefaultLang: "en"}
		page := &web.Page{
			Path:      "/de/faq",
			Lang:      "de",
			Published: published,
			Alternates: []web.Alternate{
				{HrefLang: "en", Path: "/en/faq"},
				{HrefLang: "fr", Path: "/fr/faq"},
				{Type: "text/markdown", Path: "/de/faq.md"},
				{Rel: web.RelDescribedBy, Type: "application/ld+json", Path: "/de/faq.jsonld"},
			},
		}
		requirePageLinks(t, site, page,
			link{rel: "canonical", url: base + "/de/faq"},
			// Every version of a translated page has to name itself too,
			// or search engines ignore the set.
			link{rel: "alternate", hrefLang: "de", url: base + "/de/faq"},
			link{rel: "alternate", hrefLang: "en", url: base + "/en/faq"},
			link{rel: "alternate", hrefLang: "fr", url: base + "/fr/faq"},
			link{rel: "alternate", typ: "text/markdown", url: base + "/de/faq.md"},
			link{rel: "describedby", typ: "application/ld+json", url: base + "/de/faq.jsonld"},
			link{rel: "alternate", hrefLang: "x-default", url: base + "/en/faq"},
		)
	})

	t.Run("page in the default language", func(t *testing.T) {
		site := &web.Site{BaseURL: base, Lang: "en", DefaultLang: "en"}
		page := &web.Page{
			Path:       "/en/faq",
			Published:  published,
			Alternates: []web.Alternate{{HrefLang: "de", Path: "/de/faq"}},
		}
		requirePageLinks(t, site, page,
			link{rel: "canonical", url: base + "/en/faq"},
			link{rel: "alternate", hrefLang: "en", url: base + "/en/faq"},
			link{rel: "alternate", hrefLang: "de", url: base + "/de/faq"},
			link{rel: "alternate", hrefLang: "x-default", url: base + "/en/faq"},
		)
	})

	t.Run("no translation in the default language", func(t *testing.T) {
		// An x-default has to point at a version that exists. Pointing it at
		// the page itself instead would send every visitor without a
		// translation to whichever version they happened to land on.
		site := &web.Site{BaseURL: base, DefaultLang: "fr"}
		page := &web.Page{
			Path:       "/de/faq",
			Lang:       "de",
			Published:  published,
			Alternates: []web.Alternate{{HrefLang: "en", Path: "/en/faq"}},
		}
		requirePageLinks(t, site, page,
			link{rel: "canonical", url: base + "/de/faq"},
			link{rel: "alternate", hrefLang: "de", url: base + "/de/faq"},
			link{rel: "alternate", hrefLang: "en", url: base + "/en/faq"},
		)
	})

	t.Run("no default language", func(t *testing.T) {
		site := &web.Site{BaseURL: base}
		page := &web.Page{
			Path:       "/de/faq",
			Lang:       "de",
			Published:  published,
			Alternates: []web.Alternate{{HrefLang: "en", Path: "/en/faq"}},
		}
		requirePageLinks(t, site, page,
			link{rel: "canonical", url: base + "/de/faq"},
			link{rel: "alternate", hrefLang: "de", url: base + "/de/faq"},
			link{rel: "alternate", hrefLang: "en", url: base + "/en/faq"},
		)
	})

	t.Run("explicit x-default", func(t *testing.T) {
		// A language chooser is the classic x-default target, and it is not
		// one of the translations Site.DefaultLang could select.
		site := &web.Site{BaseURL: base, Lang: "en", DefaultLang: "en"}
		page := &web.Page{
			Path:      "/en/faq",
			Published: published,
			Alternates: []web.Alternate{
				{HrefLang: "de", Path: "/de/faq"},
				{HrefLang: "X-Default", Path: "/"},
			},
		}
		requirePageLinks(t, site, page,
			link{rel: "canonical", url: base + "/en/faq"},
			link{rel: "alternate", hrefLang: "en", url: base + "/en/faq"},
			link{rel: "alternate", hrefLang: "de", url: base + "/de/faq"},
			link{rel: "alternate", hrefLang: "X-Default", url: base + "/"},
		)
	})

	t.Run("without path", func(t *testing.T) {
		site := &web.Site{BaseURL: base, Lang: "en", DefaultLang: "en"}
		page := &web.Page{Title: "Search results", Alternates: []web.Alternate{{HrefLang: "de", Path: "/de/search"}}}
		requirePageLinks(t, site, page)
	})
}

func TestSitePageLinksErrors(t *testing.T) {
	tests := map[string]struct {
		site    *web.Site
		page    *web.Page
		wantErr string
	}{
		"translations without a language": {
			// Without a language the page cannot name itself in the set.
			site:    &web.Site{BaseURL: "https://example.com"},
			page:    &web.Page{Path: "/faq", Alternates: []web.Alternate{{HrefLang: "de", Path: "/de/faq"}}},
			wantErr: "no language",
		},
		"translation in the page's own language": {
			site:    &web.Site{BaseURL: "https://example.com", Lang: "en"},
			page:    &web.Page{Path: "/faq", Alternates: []web.Alternate{{HrefLang: "EN", Path: "/en/faq"}}},
			wantErr: "language",
		},
		"duplicate hreflang": {
			site: &web.Site{BaseURL: "https://example.com", Lang: "en"},
			page: &web.Page{Path: "/faq", Alternates: []web.Alternate{
				{HrefLang: "de", Path: "/de/faq"},
				{HrefLang: "DE", Path: "/de/faq-2"},
			}},
			wantErr: "repeats",
		},
		"invalid alternate path": {
			site:    &web.Site{BaseURL: "https://example.com", Lang: "en"},
			page:    &web.Page{Path: "/faq", Alternates: []web.Alternate{{HrefLang: "de", Path: "/de//faq"}}},
			wantErr: "canonical form",
		},
		"invalid site language": {
			// Site.Lang is only checked by Site.Validate, which is not
			// guaranteed to have run; in a Link header it could add
			// parameters of its own.
			site:    &web.Site{BaseURL: "https://example.com", Lang: `en"; rel="preload`},
			page:    &web.Page{Path: "/faq", Alternates: []web.Alternate{{HrefLang: "de", Path: "/de/faq"}}},
			wantErr: "language tag",
		},
		"translation pointing at the page itself": {
			site:    &web.Site{BaseURL: "https://example.com", Lang: "en"},
			page:    &web.Page{Path: "/faq", Alternates: []web.Alternate{{HrefLang: "de", Path: "/faq"}}},
			wantErr: "points at the page itself",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := tt.site.PageHeaders(tt.page)
			require.ErrorContains(t, err, tt.wantErr)

			// The HTML head fails the same way, so a page cannot be served
			// with links its headers would refuse.
			_, err = tt.site.RenderPage(context.Background(), tt.page)
			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestSitePageHeadersEscapeURL(t *testing.T) {
	// The URL parser accepts a quote and an angle bracket in the host of the
	// BaseURL, where they would end the URL reference of a Link header.
	site := &web.Site{BaseURL: `https://exa"mple>.com`}
	header, err := site.PageHeaders(&web.Page{Path: "/faq"})
	require.NoError(t, err)
	require.Equal(t, []string{`<https://exa%22mple%3E.com/faq>; rel="canonical"`}, header.Values("Link"))
}

func TestSiteAddPageMetadataAlternatesIdempotent(t *testing.T) {
	site := &web.Site{BaseURL: "https://example.com", Lang: "en", DefaultLang: "en"}
	page := &web.Page{
		Path:      "/en/faq",
		Published: published,
		Alternates: []web.Alternate{
			{HrefLang: "de", Path: "/de/faq"},
			{Type: "text/markdown", Path: "/en/faq.md"},
		},
	}
	doc, err := site.RenderPage(context.Background(), page)
	require.NoError(t, err)
	require.NoError(t, site.AddPageMetadata(&doc, page))

	rendered := renderDoc(t, &doc)
	require.Equal(t, 1, strings.Count(rendered, `rel="canonical"`))
	require.Equal(t, 1, strings.Count(rendered, `hreflang="de"`))
	require.Equal(t, 1, strings.Count(rendered, `hreflang="x-default"`))
	require.Equal(t, 1, strings.Count(rendered, `type="text/markdown"`))
}

func TestSiteRenderPageLang(t *testing.T) {
	// A page in another language than the site's says so in its html lang
	// attribute, the same language its own hreflang link names.
	site := &web.Site{BaseURL: "https://example.com", Lang: "en"}
	doc, err := site.RenderPage(context.Background(), &web.Page{Path: "/de/", Lang: "de", Published: published})
	require.NoError(t, err)
	require.Equal(t, "de", doc.Lang)
	require.Equal(t, "de", doc.MetaProperty["og:locale"])
}
