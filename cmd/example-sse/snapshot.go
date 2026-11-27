package main

import (
	"net/http"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/internal/snapshot"
)

// writeSnapshot writes the example as one static page for the documentation
// site: the page handler serves, showing the state the stream ends in, the
// whole transcript with the reply complete.
//
// A static host cannot stream, and a browser's EventSource refuses a file
// that only looks like an event stream, so the snapshot drops the
// sse-connect attribute and renders the end state into the transcript
// itself, where the stream would have appended it.
func writeSnapshot(snap *snapshot.Options, handler http.Handler) error {
	page, err := snapshot.Get(handler, "/")
	if err != nil {
		return err
	}
	page, err = snapshot.Edit(page, ` sse-connect="/stream"`, "", 1)
	if err != nil {
		return err
	}
	end, err := snapshot.Render(endState())
	if err != nil {
		return err
	}
	page, err = snapshot.InsertAfterStartTags(page, `sse-swap="`+eventMessage+`"`, end, 1)
	if err != nil {
		return err
	}
	return snap.Write("index.html", page)
}

// endState is the transcript after the stream finished: every message, then
// the reply with all of its tokens appended.
func endState() mx.Component {
	comps := make(mx.Components, 0, len(transcript)+1)
	for _, line := range transcript {
		comps = append(comps, message(line.speaker, line.text))
	}
	return append(comps, reply(replyTokens...))
}
