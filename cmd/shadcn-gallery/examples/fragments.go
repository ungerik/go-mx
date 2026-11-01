package examples

import (
	"time"

	"github.com/ungerik/go-mx"
)

// FragmentBase is the URL path the gallery serves its fragments under, such
// as "/fragments" or "/go-mx/shadcn/gallery/fragments". main sets it before
// the first example renders.
var FragmentBase = "/fragments"

// Fragment is a piece of markup the gallery serves on its own, for an
// example that loads more markup after the page, such as a calendar loading
// another month. The static export writes it to <dir>/fragments/<Path>, so
// the gallery needs no server for it.
type Fragment struct {
	Path string
	Func func() mx.Component
}

// fragments is filled by the examples that serve fragments, in a fixed
// order, so the static export is deterministic.
var fragments []Fragment

// Fragments returns every fragment of the examples.
func Fragments() []Fragment {
	return fragments
}

// fragmentURL returns the URL of the fragment at path.
func fragmentURL(path string) string {
	return FragmentBase + "/" + path
}

// monthURL returns the MonthURL of a calendar whose months are the
// fragments <name>/2006-01.html.
func monthURL(name string) func(time.Time) string {
	return func(month time.Time) string {
		return fragmentURL(name + "/" + month.Format("2006-01") + ".html")
	}
}

// addMonthFragments registers the fragments <name>/2006-01.html of the
// months from first to last, rendered by calendar, which is what the
// calendar's [monthURL] fetches.
func addMonthFragments(name string, first, last time.Time, calendar func(month time.Time) mx.Component) {
	for m := first; !m.After(last); m = m.AddDate(0, 1, 0) {
		month := m
		fragments = append(fragments, Fragment{
			Path: name + "/" + month.Format("2006-01") + ".html",
			Func: func() mx.Component { return calendar(month) },
		})
	}
}
