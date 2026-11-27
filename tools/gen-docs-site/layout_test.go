package main

import (
	"strings"
	"testing"
)

func TestBuildNav(t *testing.T) {
	pages := []*Page{
		{Source: "docs/index.md", Path: "/", Section: overviewSection, NavTitle: "Home"},
		{Source: "docs/html/index.md", Path: "/html/", Section: "html", NavTitle: "Overview"},
		{Source: "README.md", Path: "/repo/", Section: repoSection, NavTitle: "README"},
	}
	nav := buildNav(pages, examplesNav(true, true))
	var titles []string
	for _, s := range nav {
		titles = append(titles, s.Title)
	}
	// Examples go after the documentation and before the repository files.
	if got := strings.Join(titles, ","); got != "Overview,html,Examples,Repository" {
		t.Errorf("sections %s", got)
	}
	exampleCommands := examples
	examples := nav[2].Items
	if examples[0].File != "shadcn/gallery/index.html" || examples[1].File != "examples/index.html" || len(examples) != 2+len(exampleCommands) {
		t.Errorf("examples %+v", examples)
	}

	// Without repository pages the examples still get a section, last.
	nav = buildNav(pages[:2], examplesNav(false, false))
	if last := nav[len(nav)-1]; last.Title != "Examples" || len(last.Items) != 1 {
		t.Errorf("last section %+v", last)
	}
}
