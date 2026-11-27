package main

import (
	"strings"
	"testing"

	"github.com/yuin/goldmark/ast"
)

func TestRenderMarkdown(t *testing.T) {
	r, pages := testResolver(t)
	page := pages["html/README.md"]
	page.Body = []byte(strings.Join([]string{
		"# html",
		"",
		"See the [tutorial](../docs/html/tutorial.md#step-1) and [element.go](element.go).",
		"",
		"## Elements & `Attribs`",
		"",
		"## Elements & `Attribs`",
		"",
		"| a | b |",
		"|---|---|",
		"| 1 | ~~2~~ |",
		"",
		"- [x] done",
		"",
		"Visit https://example.org now.[^1]",
		"",
		"[^1]: A footnote.",
		"",
		"```go",
		`fmt.Println("<b>")`,
		"```",
		"",
		"```sh",
		"echo '<b>'",
		"```",
		"",
		"<details><summary>raw</summary>kept</details>",
	}, "\n"))
	got, err := renderMarkdown(page, r)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`<a href="../../html/tutorial.html#step-1">tutorial</a>`,
		`<a href="` + repoURL + `blob/main/html/element.go">element.go</a>`,
		// GitHub's heading ids, a repeat numbered.
		`<h2 id="elements--attribs">`,
		`<h2 id="elements--attribs-1">`,
		"<table>", "<del>2</del>", `type="checkbox"`,
		`<a href="https://example.org">https://example.org</a>`,
		`class="footnotes"`,
		// Go highlighted server-side, its markup escaped.
		`<pre class="hl">`, `<span class="hl-string">&quot;&lt;b&gt;&quot;</span>`,
		// Other languages plain, still escaped.
		`<code class="language-sh">echo &apos;&lt;b&gt;&apos;`,
		"<details><summary>raw</summary>kept</details>",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("rendered HTML misses %q", want)
		}
	}
	if t.Failed() {
		t.Log(got)
	}

	page.Body = []byte("[broken](nowhere.md) and [also](gone.md)")
	_, err = renderMarkdown(page, r)
	if err == nil || !strings.Contains(err.Error(), "nowhere.md") || !strings.Contains(err.Error(), "gone.md") {
		t.Errorf("broken links not all reported: %v", err)
	}
}

func TestGitHubIDs(t *testing.T) {
	ids := newGitHubIDs()
	for _, c := range []struct{ text, want string }{
		{"Step 3: Serve it", "step-3-serve-it"},
		{"Build with `html`", "build-with-html"},
		{"snake_case and dash-case", "snake_case-and-dash-case"},
		{"Über Größe", "über-größe"},
		{"A — B", "a--b"},
		{"Step 3: Serve it", "step-3-serve-it-1"},
		{"Step 3: Serve it", "step-3-serve-it-2"},
		{"!!!", "section"},
	} {
		if got := string(ids.Generate([]byte(c.text), ast.KindHeading)); got != c.want {
			t.Errorf("Generate(%q) = %q, want %q", c.text, got, c.want)
		}
	}
	ids.Put([]byte("taken"))
	if got := string(ids.Generate([]byte("Taken"), ast.KindHeading)); got != "taken-1" {
		t.Errorf("an id set explicitly was reused: %q", got)
	}
}
