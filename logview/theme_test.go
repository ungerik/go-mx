package logview

import (
	"strings"
	"testing"
)

func TestThemeCSSIsDeterministic(t *testing.T) {
	// The stylesheet is normally rendered into a page on every request, and map
	// iteration order would otherwise make two identical pages differ.
	first := DarkTheme.CSS("")
	for range 20 {
		if DarkTheme.CSS("") != first {
			t.Fatal("Theme.CSS is not deterministic across calls")
		}
	}
}

func TestThemeCSSCoversTheEmittedClasses(t *testing.T) {
	css := DarkTheme.CSS("")
	for _, want := range []string{
		".log-view {",
		".log-toolbar {",
		".log-filter {",
		".log-pause {",
		".log-badge {",
		".log-lines {",
		".log-line {",
		".log-multiline {",
		".log-error {",
		".log-key {",
		".log-string {",
		".log-number {",
		".log-bool {",
		".log-null {",
		".log-time {",
		".log-level-error {",
		".log-level-undefined {",
	} {
		if !strings.Contains(css, want) {
			t.Errorf("missing rule %q in:\n%s", want, css)
		}
	}
}

func TestThemeCSSHidesPausedAndFilteredLines(t *testing.T) {
	// Pausing and filtering hide lines through these two attributes rather than
	// the hidden property, whose UA-stylesheet rule any author display
	// declaration beats — a flex log row would stay visible.
	css := DarkTheme.CSS("")
	for _, want := range []string{
		".log-line[data-mx-log-pending]",
		".log-line[data-mx-log-nomatch]",
		"display: none !important;",
	} {
		if !strings.Contains(css, want) {
			t.Errorf("missing %q in:\n%s", want, css)
		}
	}
}

func TestThemeCSSPrefix(t *testing.T) {
	// The prefix is a CSS() argument rather than a Theme field so markup and
	// stylesheet cannot silently disagree about which classes they use.
	css := DarkTheme.CSS("mylog-")
	if !strings.Contains(css, ".mylog-line {") {
		t.Errorf("custom prefix not applied:\n%s", css)
	}
	if strings.Contains(css, ".log-line {") {
		t.Errorf("default prefix leaked into a custom-prefix stylesheet:\n%s", css)
	}
	// The behavior attributes are independent of the prefix, so one script
	// serves views rendered with different prefixes.
	if !strings.Contains(css, "[data-mx-log-pending]") {
		t.Errorf("behavior attribute changed with the prefix:\n%s", css)
	}
}

func TestStyleDecls(t *testing.T) {
	for _, tc := range []struct {
		style Style
		want  string
	}{
		{Style{}, ""},
		{Style{Color: "#fff"}, "color: #fff"},
		{Style{Background: "#000"}, "background-color: #000"},
		{Style{Bold: true, Italic: true}, "font-weight: bold; font-style: italic"},
		{Style{Underline: true}, "text-decoration: underline"},
		{Style{Strikethrough: true}, "text-decoration: line-through"},
		// Both decorations are one CSS property, so they have to combine rather
		// than emit the declaration twice, where the second would win.
		{Style{Underline: true, Strikethrough: true}, "text-decoration: underline line-through"},
	} {
		if got := tc.style.decls(); got != tc.want {
			t.Errorf("Style%+v decls = %q, want %q", tc.style, got, tc.want)
		}
	}
}

func TestThemeCSSSkipsIllegalLevelKeys(t *testing.T) {
	// A theme key is a class token. Accepting one with a space would reopen the
	// hole that keeping level values out of the class attribute closes.
	theme := Theme{Levels: map[string]LevelStyle{
		"ok":                    {Style: Style{Color: "#fff"}},
		"bad key":               {Style: Style{Color: "#fff"}},
		"bad>key":               {Style: Style{Color: "#fff"}},
		"":                      {Style: Style{Color: "#fff"}},
		strings.Repeat("x", 33): {Style: Style{Color: "#fff"}},
	}}
	css := theme.CSS("")
	if !strings.Contains(css, ".log-level-ok {") {
		t.Errorf("a legal level key was dropped:\n%s", css)
	}
	for _, unwanted := range []string{"bad key", "bad>key", strings.Repeat("x", 33)} {
		if strings.Contains(css, unwanted) {
			t.Errorf("illegal level key %q reached the stylesheet:\n%s", unwanted, css)
		}
	}
}

func TestLevelToken(t *testing.T) {
	theme := Theme{Levels: map[string]LevelStyle{"info": {}, "bad key": {}}}
	for _, tc := range []struct{ value, want string }{
		{"info", "info"},
		{"INFO", "info"},
		{" Info ", "info"},
		{"warn", LevelUndefined},    // not defined by this theme
		{"bad key", LevelUndefined}, // defined, but not a legal class token
		{"", LevelUndefined},
	} {
		if got := theme.levelToken(tc.value); got != tc.want {
			t.Errorf("levelToken(%q) = %q, want %q", tc.value, got, tc.want)
		}
	}
}

func TestZeroThemeIsDark(t *testing.T) {
	// A zero Config has to produce a usable view, and a log panel is normally
	// dark even on a light page.
	if (&Config{}).theme().Name != DarkTheme.Name {
		t.Error("the zero Config did not select DarkTheme")
	}
	// A theme that only says "no colors" is still the caller's choice.
	custom := Theme{Name: "none"}
	if (&Config{Theme: custom}).theme().Name != "none" {
		t.Error("an explicit theme was overridden by the default")
	}
}
