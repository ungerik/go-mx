package wordpress

import (
	"io"
	"os"
	"strconv"

	"github.com/domonda/go-errs"
)

// Parse reads one WXR document from r into a [Site] and a [Report]. WordPress
// produces this XML under Tools → Export.
//
// A single malformed <item> is skipped and recorded in the Report rather than
// failing the whole import; the returned error is reserved for fatal conditions
// (input that is not a WXR export, a truncated document, an I/O failure).
func Parse(r io.Reader) (*Site, *Report, error) {
	site := &Site{}
	rep := &Report{}
	if err := parseWXR(r, site, rep); err != nil {
		return nil, nil, err
	}
	if err := validateWXR(site); err != nil {
		return nil, nil, err
	}
	finalize(site, rep)
	return site, rep, nil
}

// ParseFile reads a WXR document from the named file.
func ParseFile(name string) (*Site, *Report, error) {
	return ParseFiles(name)
}

// ParseFiles reads and merges a split multi-file WXR export (WordPress splits
// large exports into several files, each a complete RSS document repeating the
// authors and terms). Authors and terms are de-duplicated; posts, pages,
// attachments and menu items are concatenated.
func ParseFiles(names ...string) (*Site, *Report, error) {
	if len(names) == 0 {
		return nil, nil, errs.New("wordpress.ParseFiles: no input files given")
	}
	site := &Site{}
	rep := &Report{}
	for _, name := range names {
		f, err := os.Open(name)
		if err != nil {
			return nil, nil, errs.Errorf("opening WXR file %q: %w", name, err)
		}
		err = parseWXR(f, site, rep)
		_ = f.Close()
		if err != nil {
			return nil, nil, errs.Errorf("parsing WXR file %q: %w", name, err)
		}
		rep.SourceFiles = append(rep.SourceFiles, name)
	}
	if err := validateWXR(site); err != nil {
		return nil, nil, err
	}
	dedupe(site)
	finalize(site, rep)
	return site, rep, nil
}

// validateWXR rejects input that is not a WordPress export before it silently
// yields an empty Site — the worst failure mode. A real WXR always carries
// <wp:wxr_version>.
func validateWXR(site *Site) error {
	if site.WXRVersion == "" {
		return errs.New("not a WordPress eXtended RSS (WXR) export: missing <wp:wxr_version>. Export the file via WordPress → Tools → Export → \"All content\"")
	}
	return nil
}

func finalize(site *Site, rep *Report) {
	rep.Counts.Posts = len(site.Posts)
	rep.Counts.Pages = len(site.Pages)
	rep.Counts.Categories = len(site.Categories)
	rep.Counts.Tags = len(site.Tags)
	rep.Counts.Authors = len(site.Authors)
	rep.Counts.Attachments = len(site.Attachments)
	rep.Counts.MenuItems = len(site.MenuItems)
	for _, p := range site.Posts {
		rep.Counts.Comments += len(p.Comments)
	}
}

// dedupe removes the author/term/item duplicates a split export repeats across
// files, keeping the first occurrence.
func dedupe(site *Site) {
	site.Authors = dedupeBy(site.Authors, func(a *Author) string {
		if a.Login != "" {
			return a.Login
		}
		return "id:" + strconv.FormatInt(a.ID, 10)
	})
	site.Categories = dedupeBy(site.Categories, func(t *Term) string { return t.Slug })
	site.Tags = dedupeBy(site.Tags, func(t *Term) string { return t.Slug })
	site.Attachments = dedupeBy(site.Attachments, func(a *Attachment) int64 { return a.ID })
	site.Posts = dedupeBy(site.Posts, func(p *Post) int64 { return p.ID })
	site.Pages = dedupeBy(site.Pages, func(p *Page) int64 { return p.ID })
	site.MenuItems = dedupeBy(site.MenuItems, func(m *MenuItem) int64 { return m.ID })
}

func dedupeBy[T any, K comparable](items []T, key func(T) K) []T {
	if len(items) < 2 {
		return items
	}
	seen := make(map[K]struct{}, len(items))
	out := items[:0]
	for _, it := range items {
		k := key(it)
		if _, dup := seen[k]; dup {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, it)
	}
	return out
}
