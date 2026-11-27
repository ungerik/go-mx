package main

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// fileDates is when a file was published and last changed.
type fileDates struct {
	Published, LastUpdated time.Time
}

// gitDates returns the dates of the Markdown files of the repository at root
// from its history, by repository path: Published the time of the first
// commit touching the file, LastUpdated that of the last one. Renames are not
// followed, so a moved file is published anew. Without git or a history, as
// in a shallow clone of one commit, the map is empty or holds that commit's
// time; see datesOf for the fallback.
func gitDates(root string) map[string]fileDates {
	dates := make(map[string]fileDates)
	cmd := exec.Command("git", "log", "--format=%x00%cI", "--name-only", "--", "*.md")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return dates
	}
	var commit time.Time
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		if stamp, ok := strings.CutPrefix(line, "\x00"); ok {
			commit, _ = time.Parse(time.RFC3339, stamp)
			continue
		}
		if line == "" || commit.IsZero() {
			continue
		}
		// git log lists the newest commit first: the first time a file
		// shows up is its last change, the last time its first commit.
		d, seen := dates[line]
		if !seen {
			d.LastUpdated = commit
		}
		d.Published = commit
		dates[line] = d
	}
	return dates
}

// datesOf returns the dates of the repository file source: those from dates,
// or for a file git knows nothing about (not committed yet, or no git) its
// modification time as both.
func datesOf(root, source string, dates map[string]fileDates) fileDates {
	if d, ok := dates[source]; ok {
		return d
	}
	info, err := os.Stat(filepath.Join(root, filepath.FromSlash(source)))
	if err != nil {
		return fileDates{}
	}
	t := info.ModTime().UTC().Truncate(time.Second)
	return fileDates{Published: t, LastUpdated: t}
}
