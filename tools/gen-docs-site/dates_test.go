package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestGitDates(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("no git")
	}
	root := t.TempDir()
	git := func(env []string, args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		// The host's git configuration (commit signing, hooks of an init
		// template) must not decide whether the fixture commits succeed.
		cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1")
		cmd.Env = append(cmd.Env, env...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	at := func(date string) []string {
		return []string{"GIT_COMMITTER_DATE=" + date, "GIT_AUTHOR_DATE=" + date,
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.org", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.org"}
	}
	git(nil, "init", "-q")
	writeFiles(t, root, map[string]string{"a.md": "1", "b.md": "1"})
	git(nil, "add", ".")
	git(at("2026-01-01T10:00:00Z"), "commit", "-q", "-m", "first")
	writeFiles(t, root, map[string]string{"a.md": "2"})
	git(nil, "add", ".")
	git(at("2026-02-01T10:00:00Z"), "commit", "-q", "-m", "second")

	dates := gitDates(root)
	jan := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	feb := time.Date(2026, 2, 1, 10, 0, 0, 0, time.UTC)
	if d := dates["a.md"]; !d.Published.Equal(jan) || !d.LastUpdated.Equal(feb) {
		t.Errorf("a.md: %+v", d)
	}
	if d := dates["b.md"]; !d.Published.Equal(jan) || !d.LastUpdated.Equal(jan) {
		t.Errorf("b.md: %+v", d)
	}

	// A file git does not know falls back to its modification time.
	writeFiles(t, root, map[string]string{"new.md": "x"})
	mtime := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	if err := os.Chtimes(filepath.Join(root, "new.md"), mtime, mtime); err != nil {
		t.Fatal(err)
	}
	if d := datesOf(root, "new.md", dates); !d.Published.Equal(mtime) || !d.LastUpdated.Equal(mtime) {
		t.Errorf("new.md: %+v", d)
	}
	if d := datesOf(root, "missing.md", dates); !d.Published.IsZero() {
		t.Errorf("missing.md: %+v", d)
	}
}

func TestGitDatesWithoutRepository(t *testing.T) {
	if dates := gitDates(t.TempDir()); len(dates) != 0 {
		t.Errorf("dates outside a repository: %v", dates)
	}
}
