package pdf

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

// renderEmbeddedFilesTree renders a minimal document with the given
// attachments and returns the embedded-files name tree it produces.
func renderEmbeddedFilesTree(t *testing.T, attachments []Attachment) (*Renderer, string) {
	t.Helper()
	doc := NewDocument("Attachments", Paragraph("body"))
	doc.Attachments = attachments
	r := doc.NewRenderer()
	if err := doc.Render(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := r.Output(&buf); err != nil {
		t.Fatal(err)
	}
	return r, r.getEmbeddedFiles()
}

// The "name (2)" disambiguation of repeated filenames must not collide with
// an attachment literally named "name (2)": the suffix search has to skip
// keys that are already taken, or the name tree would carry a duplicate key
// and readers would lose one of the files.
func TestGetEmbeddedFilesCraftedDuplicateKeys(t *testing.T) {
	r, tree := renderEmbeddedFilesTree(t, []Attachment{
		{Content: []byte("a"), Filename: "dup.txt"},
		{Content: []byte("b"), Filename: "dup.txt (2)"},
		{Content: []byte("c"), Filename: "dup.txt"},
	})
	for _, name := range []string{"dup.txt", "dup.txt (2)", "dup.txt (3)"} {
		key := r.textstring(utf8toutf16(name))
		if n := strings.Count(tree, key); n != 1 {
			t.Errorf("name tree contains key %q %d times, want once:\n%s", name, n, tree)
		}
	}
}

// Distinct invalid-UTF-8 filenames can decode to the same UTF-16BE key (both
// "\xff" and "\xfe" become U+FFFD), so deduplication must operate on the
// encoded key, not the raw name — or the name tree carries duplicate keys.
func TestGetEmbeddedFilesInvalidUTF8Keys(t *testing.T) {
	r, tree := renderEmbeddedFilesTree(t, []Attachment{
		{Content: []byte("a"), Filename: "\xff"},
		{Content: []byte("b"), Filename: "\xfe"},
	})
	for _, name := range []string{"�", "� (2)"} {
		key := r.textstring(utf8toutf16(name))
		if n := strings.Count(tree, key); n != 1 {
			t.Errorf("name tree contains key %q %d times, want once:\n%s", name, n, tree)
		}
	}
}

// An attachment without a filename still needs a unique name-tree key: it
// falls back to a generated per-index name. (The "Attachement" spelling is
// inherited from fpdf.)
func TestGetEmbeddedFilesEmptyFilenameFallback(t *testing.T) {
	r, tree := renderEmbeddedFilesTree(t, []Attachment{
		{Content: []byte("a")},
		{Content: []byte("b")},
	})
	for _, name := range []string{"Attachement1", "Attachement2"} {
		if !strings.Contains(tree, r.textstring(utf8toutf16(name))) {
			t.Errorf("name tree missing generated key %q:\n%s", name, tree)
		}
	}
}
