package wordpress

import (
	"strings"
	"testing"
	"time"
)

func TestParseSample(t *testing.T) {
	site, rep, err := ParseFile("testdata/sample-wxr.xml")
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}

	if site.WXRVersion != "1.2" {
		t.Errorf("WXRVersion = %q, want 1.2", site.WXRVersion)
	}
	// <image><title> must not shadow the channel <title>.
	if site.Title != "Example Blog" {
		t.Errorf("Title = %q, want Example Blog", site.Title)
	}
	if site.Tagline != "Just another WordPress site" {
		t.Errorf("Tagline = %q", site.Tagline)
	}
	if site.Language != "en-US" {
		t.Errorf("Language = %q, want en-US", site.Language)
	}
	if site.BaseURL != "https://example.com" {
		t.Errorf("BaseURL = %q", site.BaseURL)
	}

	wantCounts := Counts{Posts: 2, Pages: 1, Categories: 1, Tags: 1, Authors: 1, Attachments: 1, MenuItems: 1, Comments: 1}
	if rep.Counts != wantCounts {
		t.Errorf("Counts = %+v, want %+v", rep.Counts, wantCounts)
	}

	post := findPost(site, 10)
	if post == nil {
		t.Fatal("post 10 not found")
	}
	// THE regression: content:encoded and excerpt:encoded share the local name
	// "encoded" and must not collide.
	if !strings.Contains(string(post.Content), "<strong>body</strong>") {
		t.Errorf("post.Content = %q, want the body HTML", post.Content)
	}
	if post.Excerpt != "A short excerpt." {
		t.Errorf("post.Excerpt = %q, want %q (content/excerpt namespace collision)", post.Excerpt, "A short excerpt.")
	}
	if post.AuthorLogin != "jane" {
		t.Errorf("post.AuthorLogin = %q, want jane", post.AuthorLogin)
	}
	if post.Status != StatusPublish {
		t.Errorf("post.Status = %q, want publish", post.Status)
	}
	if got := post.Published.UTC(); !got.Equal(time.Date(2024, 5, 1, 10, 0, 0, 0, time.UTC)) {
		t.Errorf("post.Published = %v, want 2024-05-01 10:00:00 UTC", got)
	}
	if !eqStrings(post.CategorySlugs, []string{"news"}) {
		t.Errorf("post.CategorySlugs = %v, want [news]", post.CategorySlugs)
	}
	if !eqStrings(post.TagSlugs, []string{"golang"}) {
		t.Errorf("post.TagSlugs = %v, want [golang]", post.TagSlugs)
	}
	if post.FeaturedImageID != 20 {
		t.Errorf("post.FeaturedImageID = %d, want 20", post.FeaturedImageID)
	}
	if len(post.Comments) != 1 {
		t.Fatalf("post.Comments = %d, want 1", len(post.Comments))
	}
	c := post.Comments[0]
	if c.Author != "Bob" || !c.Approved || c.ParentID != 0 {
		t.Errorf("comment = %+v, want Bob/approved/top-level", c)
	}

	// Author resolution target.
	if len(site.Authors) != 1 || site.Authors[0].Login != "jane" || site.Authors[0].DisplayName != "Jane Doe" {
		t.Errorf("authors = %+v", site.Authors)
	}

	// Category term definition (slug is the join key).
	if len(site.Categories) != 1 || site.Categories[0].Slug != "news" || site.Categories[0].Name != "News" {
		t.Errorf("categories = %+v", site.Categories)
	}

	page := site.Pages[0]
	if page.ID != 5 || page.Slug != "about" || page.ParentID != 0 {
		t.Errorf("page = %+v", page)
	}

	att := site.Attachments[0]
	if att.ID != 20 || att.Alt != "A featured image" || !strings.HasSuffix(att.URL, "featured.jpg") {
		t.Errorf("attachment = %+v", att)
	}

	mi := site.MenuItems[0]
	if mi.MenuSlug != "primary" || mi.Object != "page" || mi.ObjectID != 5 || mi.Type != "post_type" {
		t.Errorf("menu item = %+v", mi)
	}
}

func TestParseNotWXR(t *testing.T) {
	// Valid XML, valid RSS even — but not a WordPress export (no wp:wxr_version).
	const notWXR = `<?xml version="1.0"?><rss version="2.0"><channel><title>Plain RSS</title></channel></rss>`
	_, _, err := Parse(strings.NewReader(notWXR))
	if err == nil {
		t.Fatal("expected an error for non-WXR input, got nil (silent empty site is the worst failure mode)")
	}
	if !strings.Contains(err.Error(), "WXR") {
		t.Errorf("error = %q, want it to mention WXR + how to export", err)
	}
}

func TestParseLatin1(t *testing.T) {
	// ISO-8859-1 declared; 0xE9 is 'é' in Latin-1. Exercises charsetReader.
	data := "<?xml version=\"1.0\" encoding=\"ISO-8859-1\"?>" +
		"<rss version=\"2.0\" xmlns:wp=\"http://wordpress.org/export/1.2/\">" +
		"<channel><title>Caf\xe9</title><wp:wxr_version>1.2</wp:wxr_version></channel></rss>"
	site, _, err := Parse(strings.NewReader(data))
	if err != nil {
		t.Fatalf("Parse latin1: %v", err)
	}
	if site.Title != "Café" {
		t.Errorf("Title = %q, want Café (latin1 → UTF-8)", site.Title)
	}
}

func TestParseWindows1252(t *testing.T) {
	// 0x92 is a right single quote (U+2019) in Windows-1252, but U+0092 (a C1
	// control) in Latin-1. Decoding cp1252 as Latin-1 would corrupt it.
	data := "<?xml version=\"1.0\" encoding=\"windows-1252\"?>" +
		"<rss xmlns:wp=\"http://wordpress.org/export/1.2/\"><channel>" +
		"<title>It\x92s</title><wp:wxr_version>1.2</wp:wxr_version></channel></rss>"
	site, _, err := Parse(strings.NewReader(data))
	if err != nil {
		t.Fatalf("Parse cp1252: %v", err)
	}
	if site.Title != "It’s" {
		t.Errorf("Title = %q, want %q (windows-1252 smart quote)", site.Title, "It’s")
	}
}

func TestParseUnsupportedCharset(t *testing.T) {
	data := "<?xml version=\"1.0\" encoding=\"Shift_JIS\"?>" +
		"<rss xmlns:wp=\"http://wordpress.org/export/1.2/\"><channel>" +
		"<wp:wxr_version>1.2</wp:wxr_version></channel></rss>"
	_, _, err := Parse(strings.NewReader(data))
	if err == nil {
		t.Fatal("expected an error for an unsupported charset")
	}
	if !strings.Contains(err.Error(), "UTF-8") {
		t.Errorf("error = %q, want an actionable 're-export as UTF-8' message", err)
	}
}

func findPost(site *Site, id int64) *Post {
	for _, p := range site.Posts {
		if p.ID == id {
			return p
		}
	}
	return nil
}

func eqStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
