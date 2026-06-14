package wordpress

import (
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// Views renders the logical components of a [Site] as composable mx.Component
// values. Build it once with [Site.Views]; it carries the derived lookup indexes
// and permalink map so each constructor is O(1) and single-argument. Every
// constructor returns an mx.Component you can drop into your own go-mx page, so
// the importer doubles as an embeddable WordPress-content renderer.
type Views struct {
	site *Site
	opt  Options
	pl   *permalinks
	rep  *Report

	authorByLogin map[string]*Author
	catBySlug     map[string]*Term
	tagBySlug     map[string]*Term
	attachByID    map[int64]*Attachment
}

// Views builds the render context for a site. opt's zero value is valid.
func (s *Site) Views(opt Options) *Views {
	v := &Views{
		site: s, opt: opt,
		pl:            buildPermalinks(s, opt.Permalinks, opt.BasePath),
		rep:           &Report{},
		authorByLogin: make(map[string]*Author, len(s.Authors)),
		catBySlug:     make(map[string]*Term, len(s.Categories)),
		tagBySlug:     make(map[string]*Term, len(s.Tags)),
		attachByID:    make(map[int64]*Attachment, len(s.Attachments)),
	}
	for _, a := range s.Authors {
		v.authorByLogin[a.Login] = a
	}
	for _, t := range s.Categories {
		v.catBySlug[t.Slug] = t
	}
	for _, t := range s.Tags {
		v.tagBySlug[t.Slug] = t
	}
	for _, a := range s.Attachments {
		v.attachByID[a.ID] = a
	}
	finalize(s, v.rep)
	return v
}

// Report returns the diagnostics accumulated while rendering (content findings)
// plus the import counts.
func (v *Views) Report() *Report { return v.rep }

func (v *Views) siteTitle() string {
	if v.opt.SiteTitle != "" {
		return v.opt.SiteTitle
	}
	return v.site.Title
}

// Document wraps a page body in the site shell and an html.Document carrying the
// theme + prose CSS head fragment.
func (v *Views) Document(title string, body mx.Component) *html.Document {
	full := v.siteTitle()
	if title != "" && title != full {
		full = title + " — " + v.siteTitle()
	}
	doc := html.NewDocument(full, v.Shell(body))
	doc.Meta = map[string]string{"viewport": "width=device-width, initial-scale=1"}
	doc.HeadCustom = HeadComponents()
	return doc
}

// Shell is the header (site title + primary menu) / main / footer chrome. The
// primary nav wraps on small screens (no JS); deeper menu levels are flattened.
func (v *Views) Shell(content mx.Component) mx.Component {
	return mx.Components{
		html.A(html.HRef("#content"),
			html.Class("sr-only focus:not-sr-only focus:absolute focus:left-2 focus:top-2 focus:z-50 focus:rounded focus:bg-background focus:px-3 focus:py-2 focus:shadow"),
			"Skip to content"),
		html.Header(html.Class("border-b"),
			html.Div(html.Class("mx-auto max-w-5xl px-4 py-4 flex flex-wrap items-center justify-between gap-4"),
				html.A(html.HRef(v.pl.HomePath()), html.Class("text-lg font-semibold"), v.siteTitle()),
				v.MenuNav("primary"),
			),
		),
		html.Main(html.ID("content"), html.Class("px-4 py-8"), content),
		html.Footer(html.Class("border-t mt-16"),
			html.Div(html.Class("mx-auto max-w-5xl px-4 py-6 text-sm text-muted-foreground"),
				html.Textf("%s — rendered with go-mx", v.siteTitle()),
			),
		),
	}
}

// MenuNav renders the top level of the named WordPress menu as a horizontal,
// wrapping link row. Submenus are flattened in v1.
func (v *Views) MenuNav(menuSlug string) mx.Component {
	var items []*MenuItem
	for _, mi := range v.site.MenuItems {
		if mi.MenuSlug == menuSlug && mi.ParentItemID == 0 {
			items = append(items, mi)
		}
	}
	if len(items) == 0 {
		return nil
	}
	sortByOrder(items)
	links := make([]any, 0, len(items)+1)
	links = append(links, html.Attrib("aria-label", "Primary"), html.Class("flex flex-wrap items-center gap-4"))
	for _, mi := range items {
		links = append(links, html.A(html.HRef(v.menuHref(mi)), html.Class("text-sm hover:underline"), menuLabel(mi)))
	}
	return html.Nav(links...)
}

func (v *Views) menuHref(mi *MenuItem) string {
	switch {
	case mi.Type == "custom" && mi.URL != "":
		if safe, ok := (&walker{}).safeURL(mi.URL, false); ok {
			return safe
		}
		return "#"
	case mi.Object == "page":
		if r := v.pl.page[mi.ObjectID]; r != "" {
			return r
		}
	case mi.Object == "post":
		if r := v.pl.post[mi.ObjectID]; r != "" {
			return r
		}
	}
	return "#"
}

// Breadcrumb is a Home → current trail.
func (v *Views) Breadcrumb(current string) mx.Component {
	return shadcn.Breadcrumb(
		shadcn.BreadcrumbList(
			shadcn.BreadcrumbItem(shadcn.BreadcrumbLink(html.HRef(v.pl.HomePath()), "Home")),
			shadcn.BreadcrumbSeparator(),
			shadcn.BreadcrumbItem(shadcn.BreadcrumbPage(current)),
		),
	)
}

// Byline is the author + date + primary-category meta row.
func (v *Views) Byline(authorLogin string, published time.Time, primaryCat string) mx.Component {
	var parts []any
	if a := v.authorByLogin[authorLogin]; a != nil && a.DisplayName != "" {
		parts = append(parts, html.Span(a.DisplayName))
	}
	if !published.IsZero() {
		parts = append(parts, html.Time(html.Attrib("datetime", published.Format("2006-01-02")), published.Format("January 2, 2006")))
	}
	if t := v.catBySlug[primaryCat]; t != nil {
		parts = append(parts, html.A(html.HRef(v.pl.CategoryPath(primaryCat)), html.Class("hover:underline"), t.Name))
	}
	if len(parts) == 0 {
		return nil
	}
	return html.Div(append([]any{html.Class("flex flex-wrap items-center gap-x-2 gap-y-1 text-sm text-muted-foreground")}, dotJoin(parts)...)...)
}

// FeaturedImage renders an attachment by id, or nil if there is none.
func (v *Views) FeaturedImage(id int64) mx.Component {
	a := v.attachByID[id]
	if a == nil || a.URL == "" {
		return nil
	}
	alt := a.Alt
	if alt == "" {
		alt = a.Title
	}
	return html.Img(html.Src(a.URL), html.Alt(alt), html.Class("w-full rounded-lg"),
		html.Attrib("loading", "lazy"), html.Attrib("onerror", "this.style.display='none'"))
}

// Body renders post/page content through the safe content pipeline.
func (v *Views) Body(content ContentHTML, postID int64) mx.Component {
	return renderContent(content, postID, contentOptions{baseHost: hostOf(v.site.BaseURL), resolve: v.pl.resolve}, v.rep)
}

// TaxonomyBadges renders categories (secondary) and tags (outline) as linked badges.
func (v *Views) TaxonomyBadges(catSlugs, tagSlugs []string) mx.Component {
	var badges []any
	for _, s := range catSlugs {
		if t := v.catBySlug[s]; t != nil {
			badges = append(badges, html.A(html.HRef(v.pl.CategoryPath(s)), shadcn.Badge(shadcn.BadgeSecondary, t.Name)))
		}
	}
	for _, s := range tagSlugs {
		if t := v.tagBySlug[s]; t != nil {
			badges = append(badges, html.A(html.HRef(v.pl.TagPath(s)), shadcn.Badge(shadcn.BadgeOutline, t.Name)))
		}
	}
	if len(badges) == 0 {
		return nil
	}
	return html.Div(append([]any{html.Class("flex flex-wrap gap-2")}, badges...)...)
}

// PostView is the full single-post article.
func (v *Views) PostView(p *Post) mx.Component {
	var primaryCat string
	if len(p.CategorySlugs) > 0 {
		primaryCat = p.CategorySlugs[0]
	}
	article := []any{
		html.H1(html.Class("text-3xl font-bold tracking-tight"), p.Title),
	}
	article = appendComp(article, v.Byline(p.AuthorLogin, p.Published, primaryCat), "mt-3")
	article = appendComp(article, v.FeaturedImage(p.FeaturedImageID), "mt-6")
	article = append(article, html.Div(html.Class("mt-6"), v.Body(p.Content, p.ID)))
	article = appendComp(article, v.TaxonomyBadges(p.CategorySlugs, p.TagSlugs), "mt-8")
	article = appendComp(article, v.CommentThread(p.Comments), "")
	return html.Div(html.Class("mx-auto max-w-3xl"),
		v.Breadcrumb(p.Title),
		html.Article(append([]any{html.Class("mt-4")}, article...)...),
	)
}

// PageView is a single static page; it omits post metadata and lists child pages.
func (v *Views) PageView(p *Page) mx.Component {
	body := []any{
		html.H1(html.Class("text-3xl font-bold tracking-tight"), p.Title),
		html.Div(html.Class("mt-6"), v.Body(p.Content, p.ID)),
	}
	body = appendComp(body, v.childPages(p.ID), "mt-10")
	return html.Div(html.Class("mx-auto max-w-3xl"),
		v.Breadcrumb(p.Title),
		html.Article(append([]any{html.Class("mt-4")}, body...)...),
	)
}

func (v *Views) childPages(parentID int64) mx.Component {
	var kids []any
	for _, p := range v.site.Pages {
		if p.ParentID == parentID {
			kids = append(kids, html.LI(html.A(html.HRef(v.pl.PagePath(p)), html.Class("underline"), p.Title)))
		}
	}
	if len(kids) == 0 {
		return nil
	}
	return html.Div(
		html.H2(html.Class("text-lg font-semibold"), "In this section"),
		html.UL(append([]any{html.Class("mt-2 list-disc pl-6")}, kids...)...),
	)
}

// ArchiveView is a heading + post-card grid (or an empty state).
func (v *Views) ArchiveView(heading, description string, posts []*Post) mx.Component {
	head := []any{html.Class("mx-auto max-w-5xl"),
		html.H1(html.Class("text-3xl font-bold tracking-tight"), heading)}
	if description != "" {
		head = append(head, html.P(html.Class("mt-2 text-muted-foreground"), description))
	}
	if len(posts) == 0 {
		head = append(head, html.Div(html.Class("mt-8"),
			shadcn.Alert(shadcn.AlertDefault,
				shadcn.AlertTitle("No posts yet"),
				shadcn.AlertDescription("There’s nothing here yet."))))
		return html.Div(head...)
	}
	cards := make([]any, 0, len(posts)+1)
	cards = append(cards, html.Class("mt-8 grid gap-6 sm:grid-cols-2 lg:grid-cols-3"))
	for _, p := range posts {
		cards = append(cards, v.PostCard(p))
	}
	head = append(head, html.Div(cards...))
	return html.Div(head...)
}

// PostCard is one archive card: title-dominant, with date and a clamped excerpt.
func (v *Views) PostCard(p *Post) mx.Component {
	excerpt := p.Excerpt
	if excerpt == "" {
		excerpt = firstWords(stripTags(string(p.Content)), 30)
	}
	return html.A(html.HRef(v.pl.PostPath(p)), html.Class("block group"),
		shadcn.Card(html.Class("h-full transition-colors group-hover:bg-accent"),
			shadcn.CardHeader(
				shadcn.CardTitle(p.Title),
				shadcn.CardDescription(dateText(p.Published)),
			),
			shadcn.CardContent(html.P(html.Class("text-sm text-muted-foreground line-clamp-3"), excerpt)),
		),
	)
}

// CommentThread renders approved comments, threaded via ParentID, with the
// visual indentation capped and self-referential cycles guarded.
func (v *Views) CommentThread(comments []*Comment) mx.Component {
	byParent := map[int64][]*Comment{}
	var approved []*Comment
	for _, c := range comments {
		if !c.Approved || c.Type == "pingback" || c.Type == "trackback" {
			continue
		}
		byParent[c.ParentID] = append(byParent[c.ParentID], c)
		approved = append(approved, c)
	}
	if len(approved) == 0 {
		return nil
	}
	seen := map[int64]bool{}
	var renderOne func(c *Comment, depth int) mx.Component
	renderOne = func(c *Comment, depth int) mx.Component {
		seen[c.ID] = true
		item := []any{html.Class("mt-4 " + commentIndent(depth)),
			html.Div(html.Class("text-sm font-medium"), commentAuthor(c)),
			html.Div(html.Class("text-xs text-muted-foreground"), dateText(c.Date)),
			html.Div(html.Class("mt-1 text-sm"), c.Content),
		}
		if depth < 40 { // cap recursion against parent==self / cyclic threads
			for _, kid := range byParent[c.ID] {
				if !seen[kid.ID] {
					item = append(item, renderOne(kid, depth+1))
				}
			}
		}
		return html.Div(item...)
	}
	var body []any
	for _, c := range byParent[0] {
		if !seen[c.ID] {
			body = append(body, renderOne(c, 0))
		}
	}
	// Promote orphaned replies (parent deleted, unapproved, or cyclic) to top
	// level so no approved comment is silently lost.
	for _, c := range approved {
		if !seen[c.ID] {
			body = append(body, renderOne(c, 0))
		}
	}
	return html.Section(append([]any{html.Class("mt-12"),
		html.H2(html.Class("text-xl font-semibold"), "Comments")}, body...)...)
}

// NotFound is the 404 page body.
func (v *Views) NotFound() mx.Component {
	return html.Div(html.Class("mx-auto max-w-3xl text-center py-20"),
		html.H1(html.Class("text-4xl font-bold tracking-tight"), "Page not found"),
		html.P(html.Class("mt-4 text-muted-foreground"), "The page you’re looking for doesn’t exist."),
		html.A(html.HRef(v.pl.HomePath()), html.Class("mt-6 inline-block underline"), "Back to home"),
	)
}

// --- small helpers ---

func appendComp(s []any, c mx.Component, wrapClass string) []any {
	if c == nil {
		return s
	}
	if wrapClass == "" {
		return append(s, c)
	}
	return append(s, html.Div(html.Class(wrapClass), c))
}

func dotJoin(parts []any) []any {
	if len(parts) < 2 {
		return parts
	}
	out := make([]any, 0, len(parts)*2-1)
	for i, p := range parts {
		if i > 0 {
			out = append(out, html.Span(html.Class("opacity-50"), "·"))
		}
		out = append(out, p)
	}
	return out
}

func commentIndent(depth int) string {
	if depth <= 0 {
		return ""
	}
	if depth > 3 {
		depth = 3 // cap visual nesting so deep threads don't run off-screen
	}
	return "ml-6 border-l border-border pl-4"
}

func commentAuthor(c *Comment) string {
	if c.Author != "" {
		return c.Author
	}
	return "Anonymous"
}

func dateText(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("January 2, 2006")
}

func sortByOrder(items []*MenuItem) {
	for i := 1; i < len(items); i++ {
		for j := i; j > 0 && items[j-1].Order > items[j].Order; j-- {
			items[j-1], items[j] = items[j], items[j-1]
		}
	}
}

func menuLabel(mi *MenuItem) string {
	if mi.Title != "" {
		return mi.Title
	}
	return "Link"
}

var tagStripRe = regexp.MustCompile(`<[^>]*>`)

func stripTags(s string) string {
	return strings.Join(strings.Fields(tagStripRe.ReplaceAllString(s, " ")), " ")
}

func firstWords(s string, n int) string {
	words := strings.Fields(s)
	if len(words) <= n {
		return s
	}
	return strings.Join(words[:n], " ") + "…"
}

func hostOf(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(strings.ToLower(u.Host), "www.")
}
