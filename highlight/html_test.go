package highlight

import (
	"context"
	"strings"
	"testing"

	"github.com/ungerik/go-mx"
)

// render renders a component with a non-indenting CheckedWriter and fails the
// test on any render error.
func render(t *testing.T, c mx.Component) string {
	t.Helper()
	var b strings.Builder
	if err := c.Render(context.Background(), mx.NewCheckedWriter(&b)); err != nil {
		t.Fatalf("render error: %v\npartial output:\n%s", err, b.String())
	}
	return b.String()
}

func TestComponent(t *testing.T) {
	out := render(t, Component("func main() {}\n"))
	wants := []string{
		`<pre class="hl">`,
		`<code>`,
		`<span class="hl-keyword">func</span>`,
		`<span class="hl-function">main</span>`,
		`</code></pre>`,
	}
	for _, want := range wants {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestComponentPreservesLayout(t *testing.T) {
	// The exact whitespace of the source must survive inside <pre>.
	src := "func f() {\n\tx := 1\n}\n"
	out := render(t, Component(src))
	if !strings.Contains(out, "{\n\t") {
		t.Errorf("indentation not preserved in:\n%s", out)
	}
	if !strings.HasSuffix(out, "}\n</code></pre>") {
		t.Errorf("trailing newline not preserved in:\n%s", out)
	}
}

func TestComponentEscapesText(t *testing.T) {
	out := render(t, Component(`x := "a<b>&c"`+"\n"))
	if strings.Contains(out, "<b>") {
		t.Errorf("string content was not escaped:\n%s", out)
	}
	if !strings.Contains(out, "&lt;b&gt;") {
		t.Errorf("expected escaped string content in:\n%s", out)
	}
}

func TestInline(t *testing.T) {
	out := render(t, Inline("x := 1\n"))
	if !strings.HasPrefix(out, `<code class="hl">`) {
		t.Errorf("Inline should start with <code class=\"hl\">: %s", out)
	}
	if strings.Contains(out, "<pre") {
		t.Errorf("Inline should not contain a <pre> wrapper: %s", out)
	}
}

func TestHTML(t *testing.T) {
	out, err := HTML("const x = 1\n")
	if err != nil {
		t.Fatalf("HTML error: %v", err)
	}
	if !strings.Contains(out, `<span class="hl-keyword">const</span>`) {
		t.Errorf("missing keyword span in:\n%s", out)
	}
	if !strings.Contains(out, `<span class="hl-number">1</span>`) {
		t.Errorf("missing number span in:\n%s", out)
	}
}

func TestCustomPrefixAndHighlighted(t *testing.T) {
	h := &Highlighter{
		Prefix:      "syn-",
		Highlighted: map[TokenClass]bool{ClassOperator: true},
	}
	out := render(t, h.Component("x := 1\n"))
	if !strings.Contains(out, `<pre class="syn">`) {
		t.Errorf("custom block class missing: %s", out)
	}
	if !strings.Contains(out, `<span class="syn-operator">:=</span>`) {
		t.Errorf("operator should be highlighted with custom prefix: %s", out)
	}
	if strings.Contains(out, "syn-keyword") {
		t.Errorf("keyword should not be highlighted when not selected: %s", out)
	}
}
