package main

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/internal/snapshot"
)

func TestWriteSnapshot(t *testing.T) {
	snap := &snapshot.Options{Dir: t.TempDir(), Command: "cmd/example-sse"}
	if err := writeSnapshot(snap, newHandler(false, false)); err != nil {
		t.Fatal(err)
	}
	page := snapshot.ReadPage(t, snap.Dir, "index.html")
	if strings.Contains(page, ` sse-connect="`) {
		t.Error("the snapshot still opens a stream")
	}
	// The whole transcript, then the reply with every token.
	for _, line := range transcript {
		if !strings.Contains(page, line.text) {
			t.Errorf("transcript line %q missing", line.text)
		}
	}
	if !strings.Contains(page, strings.Join(replyTokens, "")) {
		t.Error("the finished reply is missing")
	}
	if strings.Index(page, transcript[0].text) < strings.Index(page, `sse-swap="message"`) {
		t.Error("the transcript is not inside the element the stream appends to")
	}
}
