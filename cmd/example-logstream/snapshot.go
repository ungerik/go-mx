package main

import (
	"math/rand/v2"
	"time"

	"github.com/ungerik/go-mx/internal/snapshot"
)

// snapshotLines is how many generated lines the static snapshot shows, enough
// to scroll and to come across a stack trace and a level the theme does not
// define.
const snapshotLines = 80

// writeSnapshot writes the example as static pages for the documentation
// site: index.html with one view and split.html with two, both the pages the
// page handler serves, holding the same batch of lines generated from seed.
//
// A static host cannot stream, and a browser's EventSource refuses a file
// that only looks like an event stream, so the snapshot drops the
// sse-connect attribute and puts the lines into each view as the backlog the
// stream would have sent. It also drops the view's search parameter, which
// would make a filter reconnect to the server with the query: on the static
// page the filter only hides the lines on screen. Pause and the line cap work
// as they do live.
func writeSnapshot(snap *snapshot.Options, seed uint64) error {
	lines, err := snapshot.Render(view.Lines(snapshotBatch(seed)...))
	if err != nil {
		return err
	}
	handler := newHandler(newSource(), false)
	for _, p := range []struct {
		target, name string
		panes        int
	}{
		{"/", "index.html", 1},
		{"/?panes=2", "split.html", 2},
	} {
		page, err := snapshot.Get(handler, p.target)
		if err != nil {
			return err
		}
		edits := []struct {
			old, new string
			n        int
		}{
			{` sse-connect="/stream"`, "", p.panes},
			{` data-mx-log-search="` + view.SearchParam + `"`, "", p.panes},
			// The split link toggles between the two pages.
			{`href="/?panes=2"`, `href="split.html"`, 2 - p.panes},
			{`href="/"`, `href="index.html"`, p.panes - 1},
		}
		for _, e := range edits {
			if page, err = snapshot.Edit(page, e.old, e.new, e.n); err != nil {
				return err
			}
		}
		page, err = snapshot.InsertAfterStartTags(page, `data-mx-log-lines=""`, lines, p.panes)
		if err != nil {
			return err
		}
		if err := snap.Write(p.name, page); err != nil {
			return err
		}
	}
	return nil
}

// snapshotBatch generates snapshotLines lines from seed, as the generator
// goroutine would, with time stamps from a fixed clock that advances a little
// per line, so the snapshot is the same on every build.
func snapshotBatch(seed uint64) []string {
	clock := time.Date(2026, 11, 2, 9, 30, 0, 0, time.UTC)
	now = func() time.Time {
		clock = clock.Add(370 * time.Millisecond)
		return clock
	}
	defer func() { now = time.Now }()

	rnd := rand.New(rand.NewPCG(seed, seed+1))
	var lines []string
	for len(lines) < snapshotLines {
		lines = append(lines, generate(rnd)...)
	}
	return lines[:snapshotLines]
}
