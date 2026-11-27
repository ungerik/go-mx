package main

import (
	"slices"
	"strings"
	"testing"

	"github.com/ungerik/go-mx/internal/snapshot"
)

func TestWriteSnapshot(t *testing.T) {
	snap := &snapshot.Options{Dir: t.TempDir(), Command: "cmd/example-logstream"}
	if err := writeSnapshot(snap, 1); err != nil {
		t.Fatal(err)
	}
	for _, p := range []struct {
		name, otherLink string
		panes           int
	}{
		{"index.html", `href="split.html"`, 1},
		{"split.html", `href="index.html"`, 2},
	} {
		page := snapshot.ReadPage(t, snap.Dir, p.name)
		// The attributes, not the names: the view's script mentions both.
		if strings.Contains(page, ` sse-connect="`) || strings.Contains(page, ` data-mx-log-search="`) {
			t.Errorf("%s still reaches for the server", p.name)
		}
		if !strings.Contains(page, p.otherLink) {
			t.Errorf("%s misses the link %s", p.name, p.otherLink)
		}
		if n := strings.Count(page, `data-mx-log-lines=""`); n != p.panes {
			t.Errorf("%s has %d views, want %d", p.name, n, p.panes)
		}
	}
}

func TestSnapshotBatchIsDeterministic(t *testing.T) {
	a, b := snapshotBatch(1), snapshotBatch(1)
	if len(a) != snapshotLines || !slices.Equal(a, b) {
		t.Fatalf("two batches of seed 1 differ or have the wrong length %d", len(a))
	}
	if !strings.Contains(strings.Join(a, "\n"), `"2026-11-02T09:30:00.370Z"`) {
		t.Error("no line is stamped by the fixed clock")
	}
	if now().Year() == 2026 && now().Month() == 11 && now().Day() == 2 {
		t.Error("the fixed clock was not restored")
	}
}
