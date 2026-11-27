package snapshot

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ReadPage returns the snapshot file name below dir, failing t if it is
// missing or carries no snapshot note.
func ReadPage(t testing.TB, dir, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(name)))
	if err != nil {
		t.Fatal(err)
	}
	page := string(data)
	if !strings.Contains(page, "data-mx-snapshot-note") {
		t.Errorf("%s has no snapshot note", name)
	}
	return page
}
