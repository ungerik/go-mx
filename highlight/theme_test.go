package highlight

import (
	"strings"
	"testing"
)

func TestThemeCSS(t *testing.T) {
	css := LightTheme.CSS("")
	wants := []string{
		"pre.hl {",
		"background-color: #f6f8fa;",
		".hl-keyword { color: #d73a49 }",
		".hl-comment { color: #6a737d; font-style: italic }",
		".hl-string { color: #032f62 }",
	}
	for _, want := range wants {
		if !strings.Contains(css, want) {
			t.Errorf("missing %q in CSS:\n%s", want, css)
		}
	}
}

func TestThemeCSSDeterministic(t *testing.T) {
	first := LightTheme.CSS(DefaultPrefix)
	second := LightTheme.CSS(DefaultPrefix)
	if first != second {
		t.Error("CSS output is not deterministic")
	}
	// Rules are sorted by class name: builtin precedes comment precedes keyword.
	css := LightTheme.CSS("")
	iBuiltin := strings.Index(css, ".hl-builtin")
	iComment := strings.Index(css, ".hl-comment")
	iKeyword := strings.Index(css, ".hl-keyword")
	if !(iBuiltin >= 0 && iBuiltin < iComment && iComment < iKeyword) {
		t.Errorf("rules are not in sorted order: builtin=%d comment=%d keyword=%d", iBuiltin, iComment, iKeyword)
	}
}

func TestThemeCustomPrefix(t *testing.T) {
	css := LightTheme.CSS("syn-")
	if !strings.Contains(css, "pre.syn {") {
		t.Errorf("custom prefix block selector missing:\n%s", css)
	}
	if !strings.Contains(css, ".syn-keyword {") {
		t.Errorf("custom prefix token selector missing:\n%s", css)
	}
}

func TestThemeStyleElement(t *testing.T) {
	out := render(t, DarkTheme.StyleElement(""))
	if !strings.HasPrefix(out, "<style>") || !strings.HasSuffix(out, "</style>") {
		t.Errorf("StyleElement should render a <style> element: %s", out)
	}
	if !strings.Contains(out, ".hl-keyword { color: #ff7b72 }") {
		t.Errorf("dark theme keyword color missing in:\n%s", out)
	}
}

func TestStyleDecls(t *testing.T) {
	cases := []struct {
		style Style
		want  string
	}{
		{Style{}, ""},
		{Style{Color: "#fff"}, "color: #fff"},
		{Style{Color: "#fff", Bold: true}, "color: #fff; font-weight: bold"},
		{Style{Italic: true}, "font-style: italic"},
		{Style{Color: "#fff", Bold: true, Italic: true}, "color: #fff; font-weight: bold; font-style: italic"},
	}
	for _, c := range cases {
		if got := c.style.decls(); got != c.want {
			t.Errorf("Style%+v.decls() = %q, want %q", c.style, got, c.want)
		}
	}
}
