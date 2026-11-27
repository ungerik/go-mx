package main

import (
	"strings"
	"testing"
)

func TestCheckLinks(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"index.html": `<a href="html/index.html#top">ok</a> <a href="#here">self</a> <p id="here"></p>
<a href="https://example.org/x">external</a> <a href="?theme=dark">query</a> <img src="data:image/png;base64,x">
<a href="mailto:a@b.c">mail</a> <link href="style.css">
<pre><code>&lt;a href=&#34;nowhere.html&#34;&gt;</code></pre>`,
		"style.css":       "",
		"html/index.html": `<h1 id="top">html</h1><a href="../index.html#here">back</a> <a href="../doc.pdf#page=2">pdf</a>`,
		"doc.pdf":         "%PDF",
	})
	if err := checkLinks(dir); err != nil {
		t.Fatalf("a valid site failed: %v", err)
	}

	writeFiles(t, dir, map[string]string{
		"broken.html": `<a href="missing.html">1</a> <a href="html/">2</a> <a href="html">3</a>
<a href="index.html#nope">4</a> <a href="../outside.html">5</a> <a href="#gone">6</a>`,
	})
	err := checkLinks(dir)
	if err == nil {
		t.Fatal("broken links passed")
	}
	for _, want := range []string{
		`"missing.html": no such file`,
		`"html/": links a directory`,
		`"html": links a directory`,
		`"index.html#nope": no id "nope"`,
		`"../outside.html": leaves the site`,
		`"#gone": no id "gone"`,
	} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error misses %q", want)
		}
	}
}
