package highlight

import (
	"maps"
	"os"
	"strings"
	"testing"
	"testing/fstest"
)

// TestParseVSCodeThemeOneMonokai loads the theme file of the One Monokai
// extension (testdata/onemonokai-color-theme.json, version 0.5.2, MIT) and
// checks that the loader reproduces OneMonokaiTheme, whose colors were mapped
// from that file by hand. The one difference is the foreground: the file sets
// none, so the loaded theme inherits it, where OneMonokaiTheme takes the
// theme's documented plain color.
func TestParseVSCodeThemeOneMonokai(t *testing.T) {
	data, err := os.ReadFile("testdata/onemonokai-color-theme.json")
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParseVSCodeTheme(data)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "One Monokai" || got.Background != OneMonokaiTheme.Background || got.Foreground != "" {
		t.Errorf("name %q, background %q, foreground %q", got.Name, got.Background, got.Foreground)
	}
	if !maps.Equal(got.Styles, OneMonokaiTheme.Styles) {
		for class, want := range OneMonokaiTheme.Styles {
			if got.Styles[class] != want {
				t.Errorf("%s: got %+v, want %+v", class, got.Styles[class], want)
			}
		}
	}
}

func TestParseVSCodeTheme(t *testing.T) {
	theme, err := ParseVSCodeTheme([]byte(`{
		// JSONC: comments and trailing commas, as VS Code saves themes.
		"name": "Test",
		"colors": {"editor.background": "#000000", "editor.foreground": null, "foreground": "#eeeeee",},
		"tokenColors": [
			/* the global rule */ {"settings": {"foreground": "#111111"}},
			{"scope": "keyword", "settings": {"foreground": "#222222", "fontStyle": "bold"}},
			{"scope": "keyword.operator", "settings": {"foreground": "#333333"}},
			{"scope": "keyword.control, comment", "settings": {"fontStyle": "italic"}},
			{"scope": ["string"], "settings": {"foreground": "#444444"}},
			{"scope": ["string"], "settings": {"foreground": "#555555"}},
			{"scope": "source.go string", "settings": {"foreground": "#666666"}},
			{"scope": "constant.language - source.js", "settings": {"foreground": "#777777"}},
			{"scope": "comment", "settings": {"foreground": "#888888", "fontStyle": ""}},
		],
	}`))
	if err != nil {
		t.Fatal(err)
	}
	if theme.Background != "#000000" || theme.Foreground != "#111111" {
		t.Errorf("background %q, foreground %q: the global rule comes before the workbench foreground", theme.Background, theme.Foreground)
	}
	for class, want := range map[TokenClass]Style{
		// The color of "keyword", the font style of the more specific
		// "keyword.control".
		ClassKeyword: {Color: "#222222", Italic: true},
		// The more specific "keyword.operator" color, the font style of
		// "keyword": color and font style resolve separately.
		ClassOperator: {Color: "#333333", Bold: true},
		// Of two equally specific rules the later one wins; the descendant
		// selector "source.go string" is not matched.
		ClassString: {Color: "#555555"},
		// The later "comment" rule resets the italic of the earlier one.
		ClassComment: {Color: "#888888"},
	} {
		if got := theme.Styles[class]; got != want {
			t.Errorf("%s: got %+v, want %+v", class, got, want)
		}
	}
	// The exclusion selector is not matched, so nothing styles constants.
	if style, ok := theme.Styles[ClassConstant]; ok {
		t.Errorf("constant styled %+v by an exclusion selector", style)
	}
}

func TestParseVSCodeThemeFallbackForeground(t *testing.T) {
	theme, err := ParseVSCodeTheme([]byte(`{"colors": {"foreground": "#eeeeee"}}`))
	if err != nil {
		t.Fatal(err)
	}
	if theme.Foreground != "#eeeeee" {
		t.Errorf("foreground %q, want the workbench foreground", theme.Foreground)
	}
}

func TestParseVSCodeThemeErrors(t *testing.T) {
	for name, data := range map[string]string{
		"invalid JSON":   `{"name": }`,
		"tmTheme path":   `{"tokenColors": "./theme.tmTheme"}`,
		"include":        `{"include": "./base.json"}`,
		"scope type":     `{"tokenColors": [{"scope": 1, "settings": {}}]}`,
		"CSS in a color": `{"tokenColors": [{"scope": "keyword", "settings": {"foreground": "red; } body { display: none"}}]}`,
		"named color":    `{"colors": {"editor.background": "red"}}`,
	} {
		if _, err := ParseVSCodeTheme([]byte(data)); err == nil {
			t.Errorf("%s: no error", name)
		}
	}
}

func TestLoadVSCodeTheme(t *testing.T) {
	fsys := fstest.MapFS{
		"themes/base.json": {Data: []byte(`{
			"name": "Base",
			"colors": {"editor.background": "#000000", "editor.foreground": "#ffffff"},
			"tokenColors": [
				{"scope": "keyword", "settings": {"foreground": "#111111"}},
				{"scope": "string", "settings": {"foreground": "#222222"}}
			]
		}`)},
		"themes/child.json": {Data: []byte(`{
			"name": "Child",
			"include": "./base.json",
			"colors": {"editor.background": "#101010"},
			"tokenColors": [{"scope": "keyword", "settings": {"foreground": "#333333"}}]
		}`)},
		"loop/a.json": {Data: []byte(`{"include": "b.json"}`)},
		"loop/b.json": {Data: []byte(`{"include": "./a.json"}`)},
	}
	theme, err := LoadVSCodeTheme(fsys, "themes/child.json")
	if err != nil {
		t.Fatal(err)
	}
	// The including file wins its colors and, after the included rules,
	// a tie of equally specific rules; the rest comes from the base.
	if theme.Name != "Child" || theme.Background != "#101010" || theme.Foreground != "#ffffff" {
		t.Errorf("name %q, background %q, foreground %q", theme.Name, theme.Background, theme.Foreground)
	}
	if theme.Styles[ClassKeyword].Color != "#333333" || theme.Styles[ClassString].Color != "#222222" {
		t.Errorf("styles %+v", theme.Styles)
	}

	if _, err := LoadVSCodeTheme(fsys, "loop/a.json"); err == nil || !strings.Contains(err.Error(), "includes itself") {
		t.Errorf("include loop: %v", err)
	}
	if _, err := LoadVSCodeTheme(fsys, "themes/missing.json"); err == nil {
		t.Error("missing file: no error")
	}
}

func TestStripJSONC(t *testing.T) {
	got := string(stripJSONC([]byte(`{"a": "// not a comment, /* nor this */", "b": [1, 2, /* x */ ], "c": "\"", // end
}`)))
	// The trailing commas go, a block comment becomes a space, and the line
	// comment goes while the space before it stays.
	want := `{"a": "// not a comment, /* nor this */", "b": [1, 2   ], "c": "\""` + " \n}"
	if got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}
