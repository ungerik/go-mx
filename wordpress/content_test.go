package wordpress

import (
	"context"
	"strings"
	"testing"

	"github.com/ungerik/go-mx"
)

func renderStr(t *testing.T, c mx.Component) string {
	t.Helper()
	var b strings.Builder
	if err := c.Render(context.Background(), mx.NewCheckedWriter(&b)); err != nil {
		t.Fatalf("render: %v", err)
	}
	return b.String()
}

func render(t *testing.T, body string) (string, *Report) {
	t.Helper()
	rep := &Report{}
	c := renderContent(ContentHTML(body), 1, contentOptions{}, rep)
	return renderStr(t, c), rep
}

func TestContentDropsScript(t *testing.T) {
	out, rep := render(t, `<p>hello</p><script>alert('xss')</script>`)
	if strings.Contains(out, "<script") || strings.Contains(out, "alert('xss')") {
		t.Errorf("script not removed: %s", out)
	}
	if !strings.Contains(out, "hello") {
		t.Errorf("legitimate content lost: %s", out)
	}
	if len(rep.DroppedHTML) == 0 {
		t.Error("dropped <script> not recorded in report")
	}
}

func TestContentBlocksJavascriptHref(t *testing.T) {
	out, rep := render(t, `<a href="javascript:alert(1)">click</a>`)
	if strings.Contains(out, "javascript:") {
		t.Errorf("javascript: URL not blocked: %s", out)
	}
	if !strings.Contains(out, "click") {
		t.Errorf("link text lost: %s", out)
	}
	if len(rep.BlockedURLs) == 0 {
		t.Error("blocked URL not recorded")
	}
}

func TestContentStripsEventHandlerAndStyle(t *testing.T) {
	out, _ := render(t, `<img src="/a.jpg" onerror="steal()" style="color:red">`)
	if strings.Contains(out, "steal()") {
		t.Errorf("input onerror handler not stripped: %s", out)
	}
	if strings.Contains(out, "color:red") || strings.Contains(out, "style=") {
		t.Errorf("inline style not stripped: %s", out)
	}
	if !strings.Contains(out, `loading="lazy"`) {
		t.Errorf("expected lazy loading added: %s", out)
	}
	if !strings.Contains(out, "this.style.display='none'") {
		t.Errorf("expected safe onerror fallback added: %s", out)
	}
}

func TestContentDemotesHeadings(t *testing.T) {
	out, _ := render(t, `<h1>Body Title</h1><h2>Sub</h2>`)
	if strings.Contains(out, "<h1") {
		t.Errorf("body <h1> not demoted (would duplicate the page H1): %s", out)
	}
	if !strings.Contains(out, "<h2") || !strings.Contains(out, "<h3") {
		t.Errorf("heading tree not shifted down by one: %s", out)
	}
}

func TestContentStripsGutenbergComments(t *testing.T) {
	out, _ := render(t, `<!-- wp:paragraph --><p>Block content.</p><!-- /wp:paragraph -->`)
	if strings.Contains(out, "wp:paragraph") || strings.Contains(out, "<!--") {
		t.Errorf("Gutenberg block comment not stripped: %s", out)
	}
	if !strings.Contains(out, "Block content.") {
		t.Errorf("inner HTML lost: %s", out)
	}
}

func TestContentRecordsPluginBlock(t *testing.T) {
	_, rep := render(t, `<!-- wp:acf/testimonial --><div>Quote</div><!-- /wp:acf/testimonial -->`)
	if len(rep.UnsupportedBlocks) != 1 || rep.UnsupportedBlocks[0].Name != "acf/testimonial" {
		t.Errorf("plugin block not recorded: %+v", rep.UnsupportedBlocks)
	}
}

func TestContentUnknownShortcode(t *testing.T) {
	out, rep := render(t, `<p>before [contact-form-7 id="99"] after</p>`)
	if strings.Contains(out, "[contact-form-7") {
		t.Errorf("shortcode delimiter not stripped: %s", out)
	}
	if !strings.Contains(out, "before") || !strings.Contains(out, "after") {
		t.Errorf("surrounding text lost: %s", out)
	}
	if len(rep.UnknownShortcodes) != 1 || rep.UnknownShortcodes[0].Name != "contact-form-7" {
		t.Errorf("unknown shortcode not recorded: %+v", rep.UnknownShortcodes)
	}
	if rep.UnknownShortcodes[0].PostIDs[0] != 1 {
		t.Errorf("finding not tied to source post: %+v", rep.UnknownShortcodes[0])
	}
}

func TestContentEscapedShortcode(t *testing.T) {
	out, _ := render(t, `<p>[[literal]]</p>`)
	if !strings.Contains(out, "[literal]") {
		t.Errorf("escaped [[shortcode]] should render as literal [shortcode]: %s", out)
	}
}

func TestContentDataURLs(t *testing.T) {
	// data:image is allowed; data:text/html is an XSS vector and must be blocked.
	out, _ := render(t, `<img src="data:image/png;base64,iVBOR"><a href="data:text/html,<script>x</script>">x</a>`)
	if !strings.Contains(out, "data:image/png") {
		t.Errorf("data:image should be allowed: %s", out)
	}
	if strings.Contains(out, "data:text/html") {
		t.Errorf("data:text/html should be blocked: %s", out)
	}
}

func TestContentWrapper(t *testing.T) {
	out, _ := render(t, `<p>x</p>`)
	if !strings.Contains(out, `class="wp-content"`) {
		t.Errorf("content not wrapped in .wp-content prose container: %s", out)
	}
}

func TestContentBlocksSvgAndExoticSchemes(t *testing.T) {
	// data:image/svg+xml can carry script — blocked even though it's an image
	// type; file:/blob: are not in the allowlist.
	out, _ := render(t, `<a href="data:image/svg+xml;base64,PHN2Zz4=">x</a>`+
		`<img src="data:image/svg+xml,foo">`+
		`<a href="file:///etc/passwd">f</a><a href="blob:abc">b</a>`)
	for _, bad := range []string{"svg+xml", "file:", "blob:"} {
		if strings.Contains(out, bad) {
			t.Errorf("%q not blocked: %s", bad, out)
		}
	}
	// A raster data image on <img src> is allowed.
	if out2, _ := render(t, `<img src="data:image/png;base64,iVBOR">`); !strings.Contains(out2, "data:image/png") {
		t.Errorf("raster data image should pass: %s", out2)
	}
	// mailto is fine on a link but not as an image source.
	if out3, _ := render(t, `<a href="mailto:a@b.com">m</a>`); !strings.Contains(out3, "mailto:a@b.com") {
		t.Errorf("mailto should pass on href: %s", out3)
	}
	if out4, _ := render(t, `<img src="mailto:a@b.com">`); strings.Contains(out4, "mailto:") {
		t.Errorf("mailto should be blocked on img src: %s", out4)
	}
}

func TestContentLoadingNotDuplicated(t *testing.T) {
	// The author's loading must win and we must not add a second one
	// (CheckedWriter rejects duplicate attributes → render abort).
	out, _ := render(t, `<img src="/a.png" loading="eager">`)
	if strings.Count(out, "loading=") != 1 {
		t.Errorf("expected exactly one loading attribute: %s", out)
	}
	if !strings.Contains(out, `loading="eager"`) {
		t.Errorf("author loading should win: %s", out)
	}
}

func TestContentTargetGetsNoopener(t *testing.T) {
	out, _ := render(t, `<a href="https://x.com" target="_blank">x</a>`)
	if !strings.Contains(out, "noopener") {
		t.Errorf("rel=noopener should be added for target: %s", out)
	}
}

func TestContentDeepNestingNoOverflow(t *testing.T) {
	// Must not stack-overflow; beyond maxNodeDepth nodes are dropped.
	deep := strings.Repeat("<div>", 5000) + "x" + strings.Repeat("</div>", 5000)
	_, _ = render(t, deep)
}

// TestContentNoPanicOnGarbage fuzz-style smoke: the walker must not panic on
// malformed or hostile input, and must never emit a script/handler.
func TestContentNoPanicOnGarbage(t *testing.T) {
	inputs := []string{
		`<p><b>unclosed`,
		`<<<>>><p>`,
		`<svg><script>alert(1)</script></svg>`,
		`<img src=x onerror=alert(1)>`,
		`<a href=" javascript:alert(1)">x</a>`,
		`<table><tr><td>[gallery ids="1,2"]</td></tr>`,
		strings.Repeat("<div>", 200) + "deep" + strings.Repeat("</div>", 200),
	}
	for _, in := range inputs {
		out, _ := render(t, in)
		if strings.Contains(out, "alert(1)") || strings.Contains(out, "<script") {
			t.Errorf("hostile content survived for %q: %s", in, out)
		}
	}
}
