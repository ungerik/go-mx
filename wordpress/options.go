package wordpress

// Options configures rendering and static export. The zero value is valid and
// produces sane defaults (slug permalinks, published-only, site root), so most
// callers can pass Options{}. Each field is mirrored by a cmd/wordpress-import
// CLI flag.
type Options struct {
	SiteTitle  string         // overrides Site.Title in the rendered shell
	Permalinks PermalinkStyle // URL/file layout; default PermalinkSlug
	BasePath   string         // URL sub-path prefix for hosting under e.g. /blog
	Statuses   []Status       // post/page statuses to include; default {StatusPublish}
}

// includeStatuses returns the effective status filter (default: publish only).
func (o Options) includeStatuses() map[Status]bool {
	list := o.Statuses
	if len(list) == 0 {
		list = []Status{StatusPublish}
	}
	m := make(map[Status]bool, len(list))
	for _, s := range list {
		m[s] = true
	}
	return m
}
