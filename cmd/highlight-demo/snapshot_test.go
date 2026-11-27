package main

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/internal/snapshot"
)

func TestWriteSnapshot(t *testing.T) {
	snap := &snapshot.Options{Dir: t.TempDir(), Command: "cmd/highlight-demo"}
	if err := writeSnapshot(snap); err != nil {
		t.Fatal(err)
	}
	light := snapshot.ReadPage(t, snap.Dir, "index.html")
	dark := snapshot.ReadPage(t, snap.Dir, "dark.html")
	if !strings.Contains(light, `href="dark.html"`) || !strings.Contains(dark, `href="index.html"`) {
		t.Error("the theme links do not point at the other file")
	}
	if strings.Contains(light+dark, "?theme=") {
		t.Error("a query-string theme link is left")
	}
	if light == dark {
		t.Error("both themes rendered the same page")
	}
}
