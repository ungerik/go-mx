// Package web assembles websites from the components of the html package.
//
// A [Page] is the metadata and content of a single page, independent of where
// it comes from and how it is rendered: a [PageSource] yields pages (from
// markdown files with front matter, a database, …) and a [PageRenderer] turns
// one into a complete [html.Document].
//
// [Site] holds what all pages of a website have in common — above all the
// BaseURL every absolute URL of the site is built from — and derives the
// site-level files search engines expect from its pages: a [Sitemap] listing
// every indexable page and a robots.txt ([Robots]) pointing at it. It also adds
// the per-page document metadata that a rendered page needs to be presentable
// in search results and link previews (canonical URL, description, Open Graph).
package web

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"maps"
	"net/http"
	"net/url"
	gopath "path"
	"slices"
	"strings"
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

var (
	_ PageSource   = new(Site)
	_ PageRenderer = new(Site)
)

// Site is the configuration a website has in common across all of its pages.
//
// It does not serve the pages themselves — that is what Routes and a
// [PageRenderer] are for — but it owns everything that is only knowable at the
// level of the whole site: the absolute URL of every page, which pages search
// engines should see, and the robots.txt and sitemap files derived from them.
//
//	site := &web.Site{
//	    BaseURL: "https://example.com",
//	    Title:   "Example",
//	    Lang:    "en",
//	    Sources: []web.PageSource{postsSource},
//	}
//	err := site.RegisterRoutes(mux)
type Site struct {
	// BaseURL is the absolute URL the site is served from, for example
	// "https://example.com" or "https://example.com/blog" for a site below a
	// path prefix. It must have an http or https scheme and a host and must not
	// have a query or fragment. A trailing slash is optional.
	//
	// Every absolute URL of the site is built from it, so it has to be the one
	// canonical origin of the site: the redirect target of all other spellings
	// (www vs. no www, http vs. https), not just whatever host is being served.
	BaseURL string

	// Title is the name of the whole site, used as the Open Graph og:site_name.
	Title string

	// Description is the fallback for the description of a page that has none.
	Description string

	// Lang is the BCP 47 language tag of the site, for example "en" or
	// "de-AT". It is used for the html lang attribute of pages whose document
	// does not set its own.
	Lang string

	// Author is the fallback for the author of a page that has none.
	Author string

	// Sources are the pages of the site, in the order they are iterated by
	// [Site.Pages]. They are what [Site.Sitemap] is built from.
	Sources []PageSource

	// Renderer renders a single page into an html.Document.
	// [DefaultPageRenderer] is used when it is nil.
	Renderer PageRenderer

	// Robots is the robots.txt served for the site. A nil Robots means an
	// allow-all robots.txt. [Site.RobotsTxt] adds the URL of the site's
	// sitemap to whatever is configured here.
	Robots *Robots

	// SitemapPath is the path the sitemap is served from,
	// [DefaultSitemapPath] when empty.
	SitemapPath string

	// Routes are the site's own routes, registered on the mux passed to
	// [Site.RegisterRoutes] together with the robots.txt and sitemap routes.
	Routes []mx.Route
}

// Validate returns an error if BaseURL is not an absolute http or https URL
// without query or fragment, if SitemapPath is not an absolute path, or if
// Robots does not validate.
func (s *Site) Validate() error {
	if _, err := s.baseURL(); err != nil {
		return err
	}
	// The sitemap path ends up as an http.ServeMux pattern, and ServeMux
	// panics on a malformed pattern or a duplicate registration instead of
	// returning an error. Both are caught here so a misconfigured Site fails
	// where it can be handled rather than crashing at startup.
	route, err := s.SitemapRoutePath()
	if err != nil {
		return err
	}
	if strings.ContainsAny(s.SitemapPath, "{} \t") {
		return fmt.Errorf("Site.SitemapPath %q contains the ServeMux pattern characters '{', '}' or whitespace", s.SitemapPath)
	}
	if route == RobotsTxtPath {
		return fmt.Errorf("Site.SitemapPath %q collides with the robots.txt path", s.SitemapPath)
	}
	if s.Robots != nil {
		return s.Robots.Validate()
	}
	return nil
}

// baseURL returns the parsed and validated BaseURL.
func (s *Site) baseURL() (*url.URL, error) {
	if s.BaseURL == "" {
		return nil, errors.New("Site.BaseURL is empty")
	}
	if err := validateAbsoluteURL(s.BaseURL); err != nil {
		return nil, fmt.Errorf("Site.BaseURL: %w", err)
	}
	u, err := url.Parse(s.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("Site.BaseURL: %w", err)
	}
	if u.RawQuery != "" || u.Fragment != "" {
		return nil, fmt.Errorf("Site.BaseURL %q has a query or fragment", s.BaseURL)
	}
	return u, nil
}

// SitemapPathOrDefault returns SitemapPath, or [DefaultSitemapPath] if it is
// empty. It is the path below the BaseURL, not necessarily the path below the
// root of the host — see [Site.SitemapRoutePath].
func (s *Site) SitemapPathOrDefault() string {
	if s.SitemapPath == "" {
		return DefaultSitemapPath
	}
	return s.SitemapPath
}

// SitemapRoutePath returns the path the sitemap is served from as seen from the
// root of the host: SitemapPath below the path prefix of the BaseURL. A site
// with a BaseURL of "https://example.com/blog" and the default sitemap path
// serves it at "/blog/sitemap.xml".
//
// [Site.RegisterRoutes] registers this path and [Site.RobotsTxt] advertises the
// URL built from it, so what the robots.txt points at is always what the site
// serves.
func (s *Site) SitemapRoutePath() (string, error) {
	base, err := s.baseURL()
	if err != nil {
		return "", err
	}
	sitemapPath := s.SitemapPathOrDefault()
	if err := validateSitePath(sitemapPath); err != nil {
		return "", fmt.Errorf("Site.SitemapPath: %w", err)
	}
	// JoinPath returns a relative path when the base URL has none of its own
	// (EscapedPath is "sitemap.xml", not "/sitemap.xml"), while a ServeMux
	// pattern has to be rooted.
	route := base.JoinPath(sitemapPath).EscapedPath()
	if !strings.HasPrefix(route, "/") {
		route = "/" + route
	}
	return route, nil
}

// URL returns the absolute URL of an absolute path within the site, joining it
// to the BaseURL. Unescaped characters are escaped, so the result is always a
// usable URL. A trailing slash of the path is preserved because "/blog" and
// "/blog/" are different URLs.
//
// See [validateSitePath] for what a path has to look like.
func (s *Site) URL(path string) (string, error) {
	base, err := s.baseURL()
	if err != nil {
		return "", err
	}
	if err := validateSitePath(path); err != nil {
		return "", err
	}
	return base.JoinPath(path).String(), nil
}

// validateSitePath returns an error if path is not usable as an absolute path
// within a site: it has to start with "/" and must not contain "..", which
// would resolve to a URL outside the site.
//
// It also rejects the two inputs that url.URL.JoinPath does not join but
// silently changes the meaning of. An invalid percent escape ("/100%-off")
// makes JoinPath return the base URL unchanged, which would give the page the
// home page's canonical URL and a duplicate sitemap entry. A query or fragment
// would be escaped into the path itself, turning "/a?b" into "/a%3Fb".
func validateSitePath(path string) error {
	switch {
	case path == "":
		return errors.New("path is empty")
	case !strings.HasPrefix(path, "/"):
		return fmt.Errorf("path %q does not start with '/'", path)
	case strings.ContainsAny(path, "?#"):
		return fmt.Errorf("path %q contains a query or fragment", path)
	}
	parsed, err := url.Parse(path)
	if err != nil {
		return fmt.Errorf("path %q is not a valid URL path: %w", path, err)
	}
	// The decoded path decides: "/%2e%2e/admin" carries no literal ".." but
	// means the same thing to anything that normalizes the path afterwards.
	for segment := range strings.SplitSeq(parsed.Path, "/") {
		if segment == ".." {
			return fmt.Errorf("path %q contains a '..' segment", path)
		}
	}
	// JoinPath cleans the path it joins, so "/blog//hello" and "/blog/./hello"
	// both resolve to "/blog/hello". Three spellings collapsing into one URL
	// means three pages claiming one canonical URL and one duplicated sitemap
	// entry, so a path that is not already clean is rejected instead of
	// silently rewritten. A trailing slash is meaningful and kept.
	cleaned := gopath.Clean(parsed.Path)
	if cleaned != "/" && strings.HasSuffix(parsed.Path, "/") {
		cleaned += "/"
	}
	if cleaned != parsed.Path {
		return fmt.Errorf("path %q is not in its canonical form %q", path, cleaned)
	}
	return nil
}

// PageURL returns the absolute canonical URL of the page. It is an error for
// the page to have no Path, because there is no URL to build then.
func (s *Site) PageURL(page *Page) (string, error) {
	u, err := s.URL(page.Path)
	if err != nil {
		return "", fmt.Errorf("page %q: %w", page.Title, err)
	}
	return u, nil
}

// Pages iterates the pages of all Sources in order, implementing the
// PageSource interface.
func (s *Site) Pages(ctx context.Context, withContent bool) iter.Seq2[*Page, error] {
	return func(yield func(*Page, error) bool) {
		for _, source := range s.Sources {
			for page, err := range source.Pages(ctx, withContent) {
				if !yield(page, err) {
					return
				}
			}
		}
	}
}

// Sitemap builds the sitemap of the site from the [Page.Indexable] pages of all
// Sources, sorted by URL so that a regenerated sitemap only differs where the
// site did. Pages that are drafts, scheduled for later or marked NoIndex are
// left out, because a sitemap should only list URLs that are meant to show up
// in search results.
//
// An indexable page without a Path is an error rather than a skipped entry: it
// cannot be pointed at, which is a bug in the PageSource and not something to
// discover as a silently incomplete sitemap.
func (s *Site) Sitemap(ctx context.Context) (*Sitemap, error) {
	sitemap := new(Sitemap)
	for page, err := range s.Pages(ctx, false) {
		if err != nil {
			return nil, err
		}
		if !page.Indexable() {
			continue
		}
		loc, err := s.PageURL(page)
		if err != nil {
			return nil, err
		}
		lastMod := page.LastUpdated
		if lastMod.IsZero() {
			lastMod = page.Published
		}
		sitemap.URLs = append(sitemap.URLs, SitemapURL{Loc: loc, LastMod: lastMod})
	}
	slices.SortFunc(sitemap.URLs, func(a, b SitemapURL) int {
		return strings.Compare(a.Loc, b.Loc)
	})
	if err := sitemap.Validate(); err != nil {
		return nil, err
	}
	return sitemap, nil
}

// RobotsTxt returns the robots.txt of the site: the Robots value with the
// absolute URL of the site's sitemap added, or an allow-all robots.txt
// referencing the sitemap when Robots is nil. The Robots field is not modified,
// so repeated calls return the same file.
func (s *Site) RobotsTxt() (*Robots, error) {
	sitemapURL, err := s.URL(s.SitemapPathOrDefault())
	if err != nil {
		return nil, err
	}
	if s.Robots == nil {
		return AllowAllRobots(sitemapURL), nil
	}
	robots := *s.Robots
	if !slices.Contains(robots.Sitemaps, sitemapURL) {
		robots.Sitemaps = append(slices.Clip(robots.Sitemaps), sitemapURL)
	}
	if err := robots.Validate(); err != nil {
		return nil, err
	}
	return &robots, nil
}

// RenderPage renders the page with the site's Renderer ([DefaultPageRenderer]
// when nil) and adds the site-level document metadata with
// [Site.AddPageMetadata]. It implements the PageRenderer interface, so a Site
// can be used wherever a single-page renderer is expected.
func (s *Site) RenderPage(ctx context.Context, page *Page) (html.Document, error) {
	renderer := s.Renderer
	if renderer == nil {
		renderer = DefaultPageRenderer
	}
	doc, err := renderer.RenderPage(ctx, page)
	if err != nil {
		return html.Document{}, err
	}
	err = s.AddPageMetadata(&doc, page)
	if err != nil {
		return html.Document{}, err
	}
	return doc, nil
}

// AddPageMetadata adds the document metadata that makes a rendered page
// presentable in search results and link previews. It is called by
// [Site.RenderPage] and can be called directly by a custom [PageRenderer].
//
// Everything the renderer already set is kept, so a page can always override
// the site-level default:
//
//   - the html lang attribute from Site.Lang and the <title> from Page.Title
//   - description and author meta tags, falling back to the site values
//   - the Open Graph tags describing the page, including the article dates for
//     a published page
//   - a canonical link and og:url, which are skipped for a page without a
//     Path because there is no URL to point at, and for a document that
//     already carries an og:url, which is taken as the renderer having
//     declared the page URL (and its canonical link) itself
//
// The one exception is the robots meta tag: a page that is not
// [Page.Indexable] is always marked "noindex, nofollow", so that drafts,
// scheduled pages and pages marked NoIndex cannot end up in search results
// because a renderer set a robots tag of its own.
func (s *Site) AddPageMetadata(doc *html.Document, page *Page) error {
	// The renderer may hand out a document that shares its maps with a
	// template or a previously rendered page. Writing this page's metadata
	// into them would leak it into every other page using the same maps, and
	// two concurrent requests writing them would crash with "concurrent map
	// writes", so the document gets its own copies first.
	doc.Meta = maps.Clone(doc.Meta)
	doc.MetaProperty = maps.Clone(doc.MetaProperty)

	// og:url is the marker for "the URL of this page has already been
	// declared": a document that has one keeps its canonical link as it is.
	// That makes AddPageMetadata idempotent, and it makes a renderer that
	// declares the page URL responsible for the canonical link too — a
	// deliberate coupling, because a second, independently derived canonical
	// link is worse than none: crawlers seeing two of them ignore both.
	pageURLSet := doc.MetaProperty["og:url"] != ""

	if doc.Lang == "" {
		doc.Lang = s.Lang
	}
	if doc.Title == "" {
		doc.Title = page.Title
	}
	description := firstNonEmpty(page.Description, s.Description)
	author := firstNonEmpty(page.Author, s.Author)
	setMetaIfEmpty(doc, "description", description)
	setMetaIfEmpty(doc, "author", author)
	if !page.Indexable() {
		setMeta(doc, "robots", "noindex, nofollow")
	}

	// https://ogp.me/
	ogType := "website"
	if !page.Published.IsZero() {
		ogType = "article"
	}
	setMetaPropertyIfEmpty(doc, "og:type", ogType)
	setMetaPropertyIfEmpty(doc, "og:title", doc.Title)
	setMetaPropertyIfEmpty(doc, "og:site_name", s.Title)
	setMetaPropertyIfEmpty(doc, "og:description", description)
	setMetaPropertyIfEmpty(doc, "og:locale", openGraphLocale(doc.Lang))
	// The article properties only mean anything for an og:type of "article",
	// so they follow the type the document ended up with, not the one derived
	// from the page.
	if doc.MetaProperty["og:type"] == "article" {
		setMetaPropertyIfEmpty(doc, "article:published_time", formatRFC3339(page.Published))
		setMetaPropertyIfEmpty(doc, "article:modified_time", formatRFC3339(page.LastUpdated))
		setMetaPropertyIfEmpty(doc, "article:author", author)
	}

	if page.Path == "" || pageURLSet {
		return nil
	}
	canonical, err := s.PageURL(page)
	if err != nil {
		return err
	}
	setMetaPropertyIfEmpty(doc, "og:url", canonical)
	link := html.Link(html.Rel("canonical"), html.HRef(canonical))
	if doc.HeadCustom == nil {
		doc.HeadCustom = link
	} else {
		doc.HeadCustom = mx.Components{doc.HeadCustom, link}
	}
	return nil
}

// RegisterRoutes registers the site's robots.txt and sitemap handlers together
// with the site's Routes on mux, returning the error from [Site.Validate] if
// the site is not configured completely enough to build absolute URLs.
//
// The handlers are registered at their path below the root of the host, so mux
// has to be the one serving that root. That is not a limitation of this
// package but of robots.txt: crawlers only ever read it from "/robots.txt", so
// a site below a path prefix still has to serve it from the root of the host.
func (s *Site) RegisterRoutes(mux *http.ServeMux) (err error) {
	// http.ServeMux reports a malformed pattern or a pattern that conflicts
	// with one already registered by panicking. Validate catches what it can
	// see, but not a route the caller registered before this call, so the
	// panic is turned into the error this method promises to return rather
	// than being left to take down the process at startup.
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("registering the routes of the site failed: %v", p)
		}
	}()
	if err := s.Validate(); err != nil {
		return err
	}
	sitemapRoute, err := s.SitemapRoutePath()
	if err != nil {
		return err
	}
	mux.HandleFunc("GET "+RobotsTxtPath, s.HandleRobotsTxt)
	mux.HandleFunc("GET "+sitemapRoute, s.HandleSitemapXML)
	for _, route := range s.Routes {
		route.Register(mux)
	}
	return nil
}

// HandleRobotsTxt writes the site's robots.txt as a plain text response,
// implementing http.HandlerFunc.
func (s *Site) HandleRobotsTxt(response http.ResponseWriter, request *http.Request) {
	robots, err := s.RobotsTxt()
	if err != nil {
		mx.RespondNonContextError(response, err)
		return
	}
	robots.HandleHTTP(response, request)
}

// HandleSitemapXML builds the site's sitemap and writes it as an XML response,
// implementing http.HandlerFunc. The sitemap is built from the Sources for
// every request; a site with expensive sources should cache it.
func (s *Site) HandleSitemapXML(response http.ResponseWriter, request *http.Request) {
	sitemap, err := s.Sitemap(request.Context())
	if err != nil {
		mx.RespondNonContextError(response, err)
		return
	}
	sitemap.HandleHTTP(response, request)
}

// setMeta sets a meta tag of the document, creating the map if needed and
// ignoring an empty content, which would render an empty meta tag.
func setMeta(doc *html.Document, name, content string) {
	if content == "" {
		return
	}
	if doc.Meta == nil {
		doc.Meta = make(map[string]string)
	}
	doc.Meta[name] = content
}

// setMetaIfEmpty is [setMeta] for a value that must not overwrite one that is
// already set.
func setMetaIfEmpty(doc *html.Document, name, content string) {
	if doc.Meta[name] == "" {
		setMeta(doc, name, content)
	}
}

// setMetaPropertyIfEmpty sets a property meta tag (Open Graph and friends) of
// the document unless it is already set, creating the map if needed and
// ignoring an empty content.
func setMetaPropertyIfEmpty(doc *html.Document, property, content string) {
	if content == "" || doc.MetaProperty[property] != "" {
		return
	}
	if doc.MetaProperty == nil {
		doc.MetaProperty = make(map[string]string)
	}
	doc.MetaProperty[property] = content
}

// openGraphLocale converts a BCP 47 language tag like "de-AT" to the
// language_TERRITORY form used by Open Graph ("de_AT").
func openGraphLocale(lang string) string {
	return strings.ReplaceAll(lang, "-", "_")
}

// formatRFC3339 formats a time as an RFC 3339 timestamp, returning an empty
// string for the zero time.
func formatRFC3339(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

// firstNonEmpty returns the first non-empty string of values or "".
func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
