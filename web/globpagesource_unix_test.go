//go:build unix

package web_test

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/ungerik/go-fs"
	"github.com/ungerik/go-mx/web"
)

// A named pipe or a socket with a page extension is skipped without the
// source waiting on it. The file is opened once, without waiting for a
// writer of a pipe, and the open file is stat'ed, so a page file swapped for
// a pipe between a stat and an open by name cannot block the source either.
func TestGlobPageSourceSkipsPipesAndSockets(t *testing.T) {
	// A socket path has to fit into about a hundred bytes, which a test's
	// temporary directory may not leave.
	dir, err := os.MkdirTemp("", "mx")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(dir) })
	require.NoError(t, os.WriteFile(filepath.Join(dir, "hello.md"), []byte("---\ntitle: Hello\n---\npost\n"), 0o600))
	for _, name := range []string{"pipe.md", "pipe.html", "pipe.txt"} {
		require.NoError(t, syscall.Mkfifo(filepath.Join(dir, name), 0o600))
	}
	if listener, err := net.Listen("unix", filepath.Join(dir, "socket.md")); err == nil {
		t.Cleanup(func() { listener.Close() })
	}

	for _, withContent := range []bool{true, false} {
		done := make(chan []string)
		go func() {
			var titles []string
			for page, err := range (&web.GlobPageSource{Dir: fs.File(dir), Pattern: "*"}).Pages(context.Background(), withContent) {
				if err != nil {
					titles = append(titles, "error: "+err.Error())
					continue
				}
				titles = append(titles, page.Title)
			}
			done <- titles
		}()
		select {
		case titles := <-done:
			require.Equal(t, []string{"Hello"}, titles, "withContent %t", withContent)
		case <-time.After(10 * time.Second):
			t.Fatalf("the source blocked on a named pipe, withContent %t", withContent)
		}
	}
}
