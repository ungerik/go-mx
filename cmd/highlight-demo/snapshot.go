package main

import (
	"net/http"

	"github.com/ungerik/go-mx/internal/snapshot"
)

// writeSnapshot writes the demo page in both themes for the documentation
// site, as handle serves them: index.html in the light theme and dark.html in
// the dark one. The theme switch is a link with a query string, which a static
// host ignores, so each page's link is pointed at the other file.
func writeSnapshot(snap *snapshot.Options) error {
	for _, p := range []struct{ target, name, link, file string }{
		{"/", "index.html", "?theme=dark", "dark.html"},
		{"/?theme=dark", "dark.html", "?theme=light", "index.html"},
	} {
		page, err := snapshot.Get(http.HandlerFunc(handle), p.target)
		if err != nil {
			return err
		}
		page, err = snapshot.Edit(page, `href="`+p.link+`"`, `href="`+p.file+`"`, 1)
		if err != nil {
			return err
		}
		if err := snap.Write(p.name, page); err != nil {
			return err
		}
	}
	return nil
}
