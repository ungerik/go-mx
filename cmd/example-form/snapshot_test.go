package main

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/internal/snapshot"
)

func TestWriteSnapshot(t *testing.T) {
	snap := &snapshot.Options{Dir: t.TempDir(), Command: "cmd/example-form"}
	if err := writeSnapshot(snap, newHandler(newStore())); err != nil {
		t.Fatal(err)
	}

	form := snapshot.ReadPage(t, snap.Dir, "index.html")
	for _, want := range []string{
		`value="Ada Lovelace"`,
		`onsubmit="return false"`, // a static host has nothing to submit to
		`text/tailwindcss`,        // the snapshot adds the theme the bare form lacks
		`href="invalid.html"`,
	} {
		if !strings.Contains(form, want) {
			t.Errorf("index.html misses %q", want)
		}
	}

	invalid := snapshot.ReadPage(t, snap.Dir, "invalid.html")
	if n := strings.Count(invalid, `aria-invalid="true"`); n != 2 {
		t.Errorf("invalid.html marks %d fields invalid, want Name and Age", n)
	}
}
