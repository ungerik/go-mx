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
		"",
		"    y := m[k](indented.md)",
		"",
		"- a list item with a fence:",
		"",
		"    ```go",
		"    x := ForEach[int](xs, f)",
		"    ```",
		"",
		"[![badge](b.svg)](target.md) and [wrapped",
		"text](w.md)",
		"",
		"A code span `wrapped(",
		"across)` lines and [after it](span.md); `not",
		"](a-link.md)` either.",
	}, "\n")
	var seen, images []string
	got := string(rewriteMarkdownLinks([]byte(source), func(dest string, image bool) string {
		seen = append(seen, dest)
		if image {
			images = append(images, dest)
		}
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
		"",
		// Code as goldmark sees it: an indented block, and a fence in a
		// list item indented by four spaces.
		"    y := m[k](indented.md)",
		"",
		"- a list item with a fence:",
		"",
		"    ```go",
		"    x := ForEach[int](xs, f)",
		"    ```",
		"",
		"[![badge](<b.svg>)](<target.md>) and [wrapped",
		"text](<w.md>)",
		"",
		// A code span wrapped across lines, as goldmark sees it.
		"A code span `wrapped(",
		"across)` lines and [after it](<span.md>); `not",
		"](a-link.md)` either.",
	}, "\n")
	if got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
	if strings.Join(seen, " ") != "a.md i.png c.md#x https://x.org/a_(b) ref.md f.md a b.md after.md b.svg target.md w.md span.md" {
		t.Errorf("rewrote %q", seen)
	}
	if strings.Join(images, " ") != "i.png b.svg" {
		t.Errorf("images %q", images)
	}
}
