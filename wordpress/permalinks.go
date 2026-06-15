package wordpress

import (
	"net/url"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/domonda/go-errs"
)

// PermalinkStyle selects the URL/file layout for posts.
type PermalinkStyle string

const (
	// PermalinkSlug lays out posts by slug, e.g. /slug/.
	PermalinkSlug PermalinkStyle = "slug" // /slug/
	// PermalinkDated lays out posts by publication date and slug, e.g.
	// /2006/01/slug/.
	PermalinkDated PermalinkStyle = "dated" // /2006/01/slug/
	// PermalinkID lays out posts by numeric ID, e.g. /p/123/.
	PermalinkID PermalinkStyle = "id" // /p/123/
)

// reservedWindows are device names that cannot be path segments on Windows.
var reservedWindows = map[string]bool{
	"con": true, "prn": true, "aux": true, "nul": true,
	"com1": true, "com2": true, "com3": true, "com4": true, "com5": true,
	"com6": true, "com7": true, "com8": true, "com9": true,
	"lpt1": true, "lpt2": true, "lpt3": true, "lpt4": true, "lpt5": true,
	"lpt6": true, "lpt7": true, "lpt8": true, "lpt9": true,
}

// safeSlug reduces a WordPress slug (attacker-influenceable content) to a single
// path segment that is safe as both a URL component and a filename. It is ASCII
// lowercase [a-z0-9-_]; everything else (including "/", "\", "..", control
// characters and non-ASCII letters) becomes "-". An empty result means the
// caller must fall back to a synthetic id (e.g. post-123).
//
// Non-ASCII slugs collapse to "-" and thus fall back to the id form. This is a
// deliberate v1 safety choice (it sidesteps NFC normalization and
// case-insensitive-filesystem collisions); a future option can preserve them.
func safeSlug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	b.Grow(len(s))
	lastDash := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '_':
			b.WriteRune(r)
			lastDash = false
		case r == '-':
			if !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
		default:
			if !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	out := strings.Trim(b.String(), "-_")
	if out == "" {
		return ""
	}
	if reservedWindows[out] {
		out = "_" + out
	}
	return out
}

// safeOutputPath joins a URL route onto the output directory and verifies the
// result stays inside it — the guard against path traversal via a hostile slug.
func safeOutputPath(dir, route string) (string, error) {
	route = strings.ReplaceAll(route, "\\", "/")              // treat Windows separators as separators on every OS
	clean := path.Clean("/" + strings.TrimPrefix(route, "/")) // strips any ".."
	rel := strings.TrimPrefix(clean, "/")
	if !strings.HasSuffix(rel, ".html") {
		rel = strings.TrimSuffix(rel, "/") + "/index.html"
	}
	full := filepath.Join(dir, filepath.FromSlash(rel))
	back, err := filepath.Rel(dir, full)
	if err != nil || back == ".." || strings.HasPrefix(back, ".."+string(filepath.Separator)) {
		return "", errs.Errorf("unsafe output path: route %q escapes the output directory", route)
	}
	return full, nil
}

// permalinks assigns every renderable object a unique route, suffixing on
// collision so two same-slug posts never overwrite one file, and resolves
// internal source URLs to local routes for link rewriting.
type permalinks struct {
	style PermalinkStyle
	base  string // URL sub-path prefix, e.g. "/blog" ("" = site root)

	post   map[int64]string
	page   map[int64]string
	cat    map[string]string // category slug -> route
	tag    map[string]string // tag slug -> route
	author map[string]string // login -> route
	home   string

	used       map[string]bool  // allocated routes (lowercased, for case-insensitive FS)
	bySourceID map[int64]string // WordPress ?p=ID / page_id=ID -> route
	byLink     map[string]string
}

func buildPermalinks(site *Site, style PermalinkStyle, base string) *permalinks {
	if style == "" {
		style = PermalinkSlug
	}
	pl := &permalinks{
		style: style, base: strings.TrimRight(base, "/"),
		post: map[int64]string{}, page: map[int64]string{},
		cat: map[string]string{}, tag: map[string]string{}, author: map[string]string{},
		used: map[string]bool{}, bySourceID: map[int64]string{}, byLink: map[string]string{},
	}
	pl.home = pl.alloc(pl.base + "/")

	for _, p := range site.Posts {
		route := pl.alloc(pl.postRoute(p))
		pl.post[p.ID] = route
		pl.bySourceID[p.ID] = route
		if p.Link != "" {
			pl.byLink[normalizeLink(p.Link)] = route
		}
	}
	for _, p := range site.Pages {
		slug := slugOr(p.Slug, "page", p.ID)
		route := pl.alloc(pl.base + "/" + slug + "/")
		pl.page[p.ID] = route
		pl.bySourceID[p.ID] = route
		if p.Link != "" {
			pl.byLink[normalizeLink(p.Link)] = route
		}
	}
	for _, t := range site.Categories {
		pl.cat[t.Slug] = pl.alloc(pl.base + "/category/" + slugOr(t.Slug, "category", t.ID) + "/")
	}
	for _, t := range site.Tags {
		pl.tag[t.Slug] = pl.alloc(pl.base + "/tag/" + slugOr(t.Slug, "tag", t.ID) + "/")
	}
	for _, a := range site.Authors {
		if a.Login == "" {
			continue
		}
		pl.author[a.Login] = pl.alloc(pl.base + "/author/" + slugOr(a.Login, "author", a.ID) + "/")
	}
	return pl
}

func (pl *permalinks) postRoute(p *Post) string {
	slug := slugOr(p.Slug, "post", p.ID)
	switch pl.style {
	case PermalinkID:
		return pl.base + "/p/" + strconv.FormatInt(p.ID, 10) + "/"
	case PermalinkDated:
		if !p.Published.IsZero() {
			return pl.base + "/" + p.Published.Format("2006/01") + "/" + slug + "/"
		}
		fallthrough
	default:
		return pl.base + "/" + slug + "/"
	}
}

// alloc returns route, or a deterministically suffixed variant if route (case-
// insensitively) is already taken.
func (pl *permalinks) alloc(route string) string {
	key := strings.ToLower(route)
	if !pl.used[key] {
		pl.used[key] = true
		return route
	}
	trimmed := strings.TrimSuffix(route, "/")
	for i := 2; ; i++ {
		cand := trimmed + "-" + strconv.Itoa(i) + "/"
		k := strings.ToLower(cand)
		if !pl.used[k] {
			pl.used[k] = true
			return cand
		}
	}
}

// resolve maps a source-site URL to a local route, or "" if it is not an
// internal link the importer can resolve.
func (pl *permalinks) resolve(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return ""
	}
	if id := u.Query().Get("p"); id != "" {
		if r, ok := pl.bySourceID[parseID(id)]; ok {
			return r
		}
	}
	if id := u.Query().Get("page_id"); id != "" {
		if r, ok := pl.bySourceID[parseID(id)]; ok {
			return r
		}
	}
	if r, ok := pl.byLink[normalizeLink(raw)]; ok {
		return r
	}
	return ""
}

// PostPath returns the local route allocated for the given post.
func (pl *permalinks) PostPath(p *Post) string { return pl.post[p.ID] }

// PagePath returns the local route allocated for the given page.
func (pl *permalinks) PagePath(p *Page) string { return pl.page[p.ID] }

// CategoryPath returns the local route for the category with the given slug.
func (pl *permalinks) CategoryPath(slug string) string { return pl.cat[slug] }

// TagPath returns the local route for the tag with the given slug.
func (pl *permalinks) TagPath(slug string) string { return pl.tag[slug] }

// AuthorPath returns the local route for the author with the given login.
func (pl *permalinks) AuthorPath(login string) string { return pl.author[login] }

// HomePath returns the local route for the site home page.
func (pl *permalinks) HomePath() string { return pl.home }

func slugOr(slug, kind string, id int64) string {
	if s := safeSlug(slug); s != "" {
		return s
	}
	return kind + "-" + strconv.FormatInt(id, 10)
}

// normalizeLink trims a permalink to host+path for stable comparison.
func normalizeLink(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return strings.TrimRight(strings.TrimSpace(raw), "/")
	}
	host := strings.TrimPrefix(strings.ToLower(u.Host), "www.")
	return host + strings.TrimRight(u.Path, "/")
}
