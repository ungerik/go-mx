package main

import (
	"strings"
	"testing"
)

func TestRewriteMarkdownLinks(t *testing.T) {
	source := strings.Join([]string{
		"See [a](a.md) and ![img](i.png \"title\") and [`code`](c.md#x).",
		"A [paren](https://x.org/a_(b)) link and <https://auto.org>.",
		"Not a link: `[x](inline.md)` in code.",
		"[ref]: ref.md",
		"[^note]: [inside a footnote](f.md)",
		"[angle](<a b.md>)",
		"```go",
		"x := m[k](y.md)",
		"```",
		"~~~~",
		"[fenced](z.md)",
		"~~~~",
		"[after](after.md)",
		"empty []() and ](",
	}, "\n")
	var seen []string
	got := string(rewriteMarkdownLinks([]byte(source), func(dest string) string {
		seen = append(seen, dest)
		return "<" + dest + ">"
	}))
	want := strings.Join([]string{
		"See [a](<a.md>) and ![img](<i.png> \"title\") and [`code`](<c.md#x>).",
		"A [paren](<https://x.org/a_(b)>) link and <https://auto.org>.",
		"Not a link: `[x](inline.md)` in code.",
		"[ref]: <ref.md>",
		"[^note]: [inside a footnote](<f.md>)",
		"[angle](<<a b.md>>)",
		"```go",
		"x := m[k](y.md)",
		"```",
		"~~~~",
		"[fenced](z.md)",
		"~~~~",
		"[after](<after.md>)",
		"empty []() and ](",
	}, "\n")
	if got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
	if strings.Join(seen, " ") != "a.md i.png c.md#x https://x.org/a_(b) ref.md f.md a b.md after.md" {
		t.Errorf("rewrote %q", seen)
	}
}
