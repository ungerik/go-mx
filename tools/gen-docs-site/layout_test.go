package main

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/highlight"
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

// TestCodeThemeCSS checks that the code blocks take the light theme's colors
// by default and One Monokai's below the .dark root, token rules included.
func TestCodeThemeCSS(t *testing.T) {
	css := codeThemeCSS()
	for _, want := range []string{
		"pre.hl {\n\tbackground-color: " + highlight.LightTheme.Background,
		".dark pre.hl {\n\tbackground-color: " + highlight.OneMonokaiTheme.Background,
		".hl-keyword { color: " + highlight.LightTheme.Styles[highlight.ClassKeyword].Color,
		".dark .hl-keyword { color: " + highlight.OneMonokaiTheme.Styles[highlight.ClassKeyword].Color,
	} {
		if !strings.Contains(css, want) {
			t.Errorf("code theme misses %q", want)
		}
	}
	// No dark rule may apply outside .dark.
	for line := range strings.Lines(css[len(highlight.LightTheme.CSS("")):]) {
		if strings.Contains(line, "{") && !strings.HasPrefix(line, ".dark ") {
			t.Errorf("dark rule without .dark: %q", line)
		}
	}
}
