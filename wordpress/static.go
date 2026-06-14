package wordpress

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"sort"

	"github.com/domonda/go-errs"
	"github.com/ungerik/go-mx"
)

// WriteStatic renders the whole site to static HTML files under dir and returns
// the import [Report] (counts plus the content diagnostics gathered while
// rendering). It also writes import-report.json and import-report.html next to
// the site. Every output path is containment-checked, so a hostile slug can
// never write outside dir.
//
// The output uses root-absolute links, so serve it from a web root (e.g.
// `python3 -m http.server` inside dir); set [Options.BasePath] to host it under a
// URL sub-path.
func WriteStatic(site *Site, dir string, opt Options) (*Report, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, errs.Errorf("creating output directory %q: %w", dir, err)
	}
	v := site.Views(opt)
	include := opt.includeStatuses()
	posts := sortedByDate(filterPostsByStatus(site.Posts, include))
	pages := filterPagesByStatus(site.Pages, include)

	write := func(route, title string, body mx.Component) error {
		full, err := safeOutputPath(dir, route)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return errs.Errorf("creating output dir: %w", err)
		}
		f, err := os.Create(full)
		if err != nil {
			return errs.Errorf("creating %q: %w", full, err)
		}
		// No indentation: the article body may contain <pre>, where injected
		// whitespace would corrupt code. Browsers ignore the formatting anyway.
		rerr := v.Document(title, body).Render(context.Background(), mx.NewCheckedWriter(f))
		cerr := f.Close()
		if rerr != nil {
			return errs.Errorf("rendering %q: %w", route, rerr)
		}
		return cerr
	}

	if err := write(v.pl.HomePath(), "", v.ArchiveView(v.siteTitle(), site.Tagline, posts)); err != nil {
		return v.rep, err
	}
	for _, p := range posts {
		if err := write(v.pl.PostPath(p), p.Title, v.PostView(p)); err != nil {
			return v.rep, err
		}
	}
	for _, pg := range pages {
		if err := write(v.pl.PagePath(pg), pg.Title, v.PageView(pg)); err != nil {
			return v.rep, err
		}
	}
	for _, t := range site.Categories {
		h := "Category: " + termTitle(t)
		if err := write(v.pl.CategoryPath(t.Slug), h, v.ArchiveView(h, t.Description, postsWithCategory(posts, t.Slug))); err != nil {
			return v.rep, err
		}
	}
	for _, t := range site.Tags {
		h := "Tag: " + termTitle(t)
		if err := write(v.pl.TagPath(t.Slug), h, v.ArchiveView(h, t.Description, postsWithTag(posts, t.Slug))); err != nil {
			return v.rep, err
		}
	}
	for _, a := range site.Authors {
		in := postsByAuthor(posts, a.Login)
		if a.Login == "" || len(in) == 0 {
			continue
		}
		h := "Author: " + authorName(a)
		if err := write(v.pl.AuthorPath(a.Login), h, v.ArchiveView(h, "", in)); err != nil {
			return v.rep, err
		}
	}
	if err := write("/404.html", "Page not found", v.NotFound()); err != nil {
		return v.rep, err
	}
	// Render the report page last, so the content findings are complete.
	if err := write("/import-report/", "Import report", v.ImportReportView()); err != nil {
		return v.rep, err
	}
	if data, err := v.rep.JSON(); err == nil {
		if err := os.WriteFile(filepath.Join(dir, "import-report.json"), data, 0o644); err != nil {
			return v.rep, errs.Errorf("writing import-report.json: %w", err)
		}
	}
	return v.rep, nil
}

func filterPostsByStatus(posts []*Post, include map[Status]bool) []*Post {
	out := make([]*Post, 0, len(posts))
	for _, p := range posts {
		if include[p.Status] {
			out = append(out, p)
		}
	}
	return out
}

func filterPagesByStatus(pages []*Page, include map[Status]bool) []*Page {
	out := make([]*Page, 0, len(pages))
	for _, p := range pages {
		if include[p.Status] {
			out = append(out, p)
		}
	}
	return out
}

func sortedByDate(posts []*Post) []*Post {
	out := append([]*Post(nil), posts...)
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Published.After(out[j].Published)
	})
	return out
}

func postsWithCategory(posts []*Post, slug string) []*Post {
	var out []*Post
	for _, p := range posts {
		if slices.Contains(p.CategorySlugs, slug) {
			out = append(out, p)
		}
	}
	return out
}

func postsWithTag(posts []*Post, slug string) []*Post {
	var out []*Post
	for _, p := range posts {
		if slices.Contains(p.TagSlugs, slug) {
			out = append(out, p)
		}
	}
	return out
}

func postsByAuthor(posts []*Post, login string) []*Post {
	var out []*Post
	for _, p := range posts {
		if p.AuthorLogin == login {
			out = append(out, p)
		}
	}
	return out
}

func termTitle(t *Term) string {
	if t.Name != "" {
		return t.Name
	}
	return t.Slug
}

func authorName(a *Author) string {
	if a.DisplayName != "" {
		return a.DisplayName
	}
	return a.Login
}
