package wordpress

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestSafeSlug(t *testing.T) {
	cases := map[string]string{
		"hello-world":       "hello-world",
		"Hello World":       "hello-world",
		"../../etc/passwd":  "etc-passwd",
		"..":                "",
		"":                  "",
		"  spaced  ":        "spaced",
		`a/b\c`:             "a-b-c",
		"café":              "caf",  // non-ASCII collapses; caller falls back to id
		"con":               "_con", // reserved Windows device name
		"multiple---dashes": "multiple-dashes",
		"%2e%2e%2fetc":      "2e-2e-2fetc",
		"trailing.dots...":  "trailing-dots",
	}
	for in, want := range cases {
		if got := safeSlug(in); got != want {
			t.Errorf("safeSlug(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSafeOutputPathContainment(t *testing.T) {
	dir := t.TempDir()
	// A hostile route must never resolve to a file outside dir. path.Clean
	// neutralizes the traversal, so the file lands inside; assert containment.
	for _, route := range []string{"/../../../etc/cron.d/x", "/..\\..\\windows", "/a/../../../../tmp/evil"} {
		full, err := safeOutputPath(dir, route)
		if err != nil {
			continue // rejected outright is also fine
		}
		if rel, _ := filepath.Rel(dir, full); rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			t.Errorf("route %q escaped output dir: %s", route, full)
		}
	}
	// A normal route lands inside.
	full, err := safeOutputPath(dir, "/hello-world/")
	if err != nil {
		t.Fatalf("safeOutputPath: %v", err)
	}
	if !strings.HasPrefix(full, dir) || !strings.HasSuffix(full, filepath.FromSlash("hello-world/index.html")) {
		t.Errorf("unexpected path: %s", full)
	}
}

func TestPermalinkCollision(t *testing.T) {
	site := &Site{
		Posts: []*Post{
			{ID: 1, Slug: "hello"},
			{ID: 2, Slug: "hello"}, // same slug, different post
		},
	}
	pl := buildPermalinks(site, PermalinkSlug, "")
	p1, p2 := pl.PostPath(site.Posts[0]), pl.PostPath(site.Posts[1])
	if p1 == p2 {
		t.Errorf("duplicate slugs collided to the same route: %q", p1)
	}
	if p1 != "/hello/" || p2 != "/hello-2/" {
		t.Errorf("got %q and %q, want /hello/ and /hello-2/", p1, p2)
	}
}

func TestPermalinkEmptySlugFallback(t *testing.T) {
	site := &Site{Posts: []*Post{{ID: 42, Slug: ""}}}
	pl := buildPermalinks(site, PermalinkSlug, "")
	if got := pl.PostPath(site.Posts[0]); got != "/post-42/" {
		t.Errorf("empty slug route = %q, want /post-42/", got)
	}
}

func TestPermalinkResolveInternal(t *testing.T) {
	site := &Site{Posts: []*Post{{ID: 10, Slug: "hello-world", Link: "https://example.com/2024/05/hello-world/"}}}
	pl := buildPermalinks(site, PermalinkSlug, "")
	if got := pl.resolve("https://example.com/?p=10"); got != "/hello-world/" {
		t.Errorf("resolve(?p=10) = %q, want /hello-world/", got)
	}
	if got := pl.resolve("https://example.com/2024/05/hello-world/"); got != "/hello-world/" {
		t.Errorf("resolve(permalink) = %q, want /hello-world/", got)
	}
	if got := pl.resolve("https://other.com/x"); got != "" {
		t.Errorf("external link should not resolve, got %q", got)
	}
}
