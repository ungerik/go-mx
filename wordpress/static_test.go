package wordpress

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func mustExist(t *testing.T, dir, rel string) string {
	t.Helper()
	full := filepath.Join(dir, filepath.FromSlash(rel))
	data, err := os.ReadFile(full)
	if err != nil {
		t.Fatalf("expected %s to exist: %v", rel, err)
	}
	return string(data)
}

func TestWriteStatic(t *testing.T) {
	site, parseRep, err := ParseFile("testdata/sample-wxr.xml")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	rep, err := WriteStatic(site, dir, Options{})
	if err != nil {
		t.Fatalf("WriteStatic: %v", err)
	}
	rep.InheritParse(parseRep)

	for _, rel := range []string{
		"index.html", "hello-world/index.html", "blocks/index.html",
		"about/index.html", "category/news/index.html", "tag/golang/index.html",
		"author/jane/index.html", "404.html", "import-report.json", "import-report/index.html",
	} {
		mustExist(t, dir, rel)
	}

	// The post body rendered through the prose pipeline.
	post := mustExist(t, dir, "hello-world/index.html")
	if !strings.Contains(post, `class="wp-content"`) {
		t.Error("post body not wrapped in .wp-content")
	}
	if !strings.Contains(post, "<strong>body</strong>") {
		t.Error("post body content missing")
	}

	// The Gutenberg post's unknown shortcode is recorded.
	found := false
	for _, f := range rep.UnknownShortcodes {
		if f.Name == "shortcode_unknown" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected the sample's [shortcode_unknown] to be reported, got %+v", rep.UnknownShortcodes)
	}

	// The JSON report is valid and non-empty.
	js := mustExist(t, dir, "import-report.json")
	if !strings.Contains(js, `"posts": 2`) {
		t.Errorf("import-report.json counts look wrong: %s", js)
	}
}

func TestWriteStaticTraversalContained(t *testing.T) {
	site := &Site{
		WXRVersion: "1.2", Title: "Evil Site",
		Posts: []*Post{{ID: 1, Title: "Evil", Slug: "../../../../tmp/wp-evil-marker", Status: StatusPublish, Content: "<p>x</p>"}},
	}
	dir := t.TempDir()
	if _, err := WriteStatic(site, dir, Options{}); err != nil {
		t.Fatalf("WriteStatic: %v", err)
	}
	// The traversal must have been neutralized: nothing written outside dir.
	if _, err := os.Stat("/tmp/wp-evil-marker/index.html"); err == nil {
		os.RemoveAll("/tmp/wp-evil-marker")
		t.Fatal("path traversal escaped the output directory")
	}
	// Every written file is inside dir.
	_ = filepath.WalkDir(dir, func(path string, _ os.DirEntry, _ error) error {
		if rel, _ := filepath.Rel(dir, path); strings.HasPrefix(rel, "..") {
			t.Errorf("file escaped dir: %s", path)
		}
		return nil
	})
}
