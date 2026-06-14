package highlight

import (
	"go/parser"
	"strings"
	"testing"
)

func TestGoSourceIsValidGo(t *testing.T) {
	srcs := []string{
		"func main() {}\n",
		"package p\n\nimport \"fmt\"\n\nfunc f(a int) { fmt.Println(a) }\n",
		"x := \"quotes \\\" and \\t tabs\"\n",
		"not valid go @#$ still works\n",
	}
	for _, src := range srcs {
		out := GoSource(src)
		// The result is a Go expression; it must parse.
		if _, err := parser.ParseExpr(out); err != nil {
			t.Errorf("GoSource(%q) is not a valid Go expression: %v\n%s", src, err, out)
		}
	}
}

func TestGoSourceContent(t *testing.T) {
	out := GoSource("func main() {}\n")
	wants := []string{
		`html.Pre(html.Class("hl"),`,
		`html.Code(`,
		`html.Span(html.Class("hl-keyword"), "func"),`,
		`html.Span(html.Class("hl-function"), "main"),`,
	}
	for _, want := range wants {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestGoSourceIsFormatted(t *testing.T) {
	out := GoSource("func main() {}\n")
	// gofmt indents nested calls with tabs and puts each child on its own line.
	if !strings.Contains(out, "\n\thtml.Code(") {
		t.Errorf("output is not gofmt-indented:\n%s", out)
	}
	if strings.HasSuffix(out, "\n") {
		t.Errorf("formatted expression should not have a trailing newline:\n%q", out)
	}
}

func TestGoSourceMergesPlainText(t *testing.T) {
	// "fmt.Println(" -> the leading "fmt." should be a single merged string
	// literal, not three separate plain tokens.
	out := GoSource("fmt.Println(x)\n")
	if !strings.Contains(out, `"fmt.",`) {
		t.Errorf("adjacent plain tokens were not merged:\n%s", out)
	}
}
