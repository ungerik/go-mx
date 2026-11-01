package shadcn

import (
	"context"
	"strings"
	"testing"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

func TestCodeBlock(t *testing.T) {
	out := render(t, CodeBlock("", "",
		CodeBlockHeader(
			CodeBlockTitle("main.go"),
			CodeBlockActions(CodeBlockCopyButton("")),
		),
		CodeBlockContent("package main"),
		CodeBlockFooter("Go"),
	))
	for _, want := range []string{
		`data-slot="code-block"`, `data-variant="default"`, `data-size="default"`,
		`data-slot="code-block-header"`, `data-slot="code-block-title"`,
		`data-slot="code-block-actions"`, `data-slot="code-block-copy-button"`,
		`<pre data-slot="code-block-content"`, `dir="ltr"`, `tabindex="0"`,
		`style="counter-reset: line 0"`,
		`<code data-slot="code-block-code"`, "package main</code></pre>",
		`data-slot="code-block-footer"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	for _, gone := range []string{"cn-", "scroll-fade", "scrollbar-thin"} {
		if strings.Contains(out, gone) {
			t.Errorf("unexpected %q in %s", gone, out)
		}
	}
}

func TestCodeBlockDestructiveSM(t *testing.T) {
	out := render(t, CodeBlock(CodeBlockDestructive, CodeBlockSizeSM, CodeBlockContent("err")))
	if !strings.Contains(out, "border-destructive/30") || !strings.Contains(out, `data-size="sm"`) {
		t.Errorf("destructive small block classes wrong: %s", out)
	}
}

func TestCodeBlockCopyButton(t *testing.T) {
	out := render(t, CodeBlockCopyButton("go get github.com/ungerik/go-mx"))
	for _, want := range []string{
		`data-value="go get github.com/ungerik/go-mx"`,
		`aria-label="Copy"`, `data-copy-label="Copy"`, `data-copied-label="Copied"`,
		`onclick="mxCodeBlockCopy(this)"`, "window.mxCodeBlockCopy",
		"lucide-copy", "lucide-check", "data-copied:[&amp;>.lucide-check]:block",
		`data-variant="ghost"`, `data-size="icon-xs"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	// Without a value the script copies the code block's text, so no
	// data-value may be emitted (an empty one would copy "").
	if out := render(t, CodeBlockCopyButton("")); strings.Contains(out, "data-value") {
		t.Errorf("empty value should emit no data-value: %s", out)
	}
	// A caller onclick replaces the copying instead of colliding with it.
	if out := render(t, CodeBlockCopyButton("x", html.OnClick("track()"))); strings.Contains(out, "mxCodeBlockCopy(this)") {
		t.Errorf("caller onclick should win: %s", out)
	}
}

func TestCodeBlockCopyButtonLabels(t *testing.T) {
	ctx := ContextWithLabels(context.Background(), Labels{CodeBlockCopy: "Kopieren", CodeBlockCopied: "Kopiert"})
	out := renderWith(t, ctx, CodeBlockCopyButton("x"))
	for _, want := range []string{`aria-label="Kopieren"`, `data-copied-label="Kopiert"`} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}

func TestCodeBlockContentLineNumbers(t *testing.T) {
	out := render(t, CodeBlockContent(CodeBlockLineNumbers, CodeBlockStartLine(10), "a\n\nc\n"))
	// A single text child is split into lines, the trailing newline dropped,
	// and the empty middle line holds a newline so a copied selection keeps it.
	if n := strings.Count(out, `data-slot="code-block-line"`); n != 3 {
		t.Errorf("want 3 lines, got %d: %s", n, out)
	}
	if !strings.Contains(out, `<span data-slot="code-block-line-content" class="min-w-0">`+"\n</span>") {
		t.Errorf("empty line should hold a newline: %s", out)
	}
	if !strings.Contains(out, `data-numbered=""`) || !strings.Contains(out, "counter-reset: line 9") {
		t.Errorf("numbering attributes wrong: %s", out)
	}
}

func TestCodeBlockContentMergesStyle(t *testing.T) {
	out := render(t, CodeBlockContent(html.Style("max-height: 16rem"), "x"))
	if !strings.Contains(out, `style="max-height: 16rem; counter-reset: line 0"`) {
		t.Errorf("caller style should be kept and extended: %s", out)
	}
}

func TestCodeBlockContentUnindented(t *testing.T) {
	// An indenting writer breaks the line before every element, but not
	// inside the <pre>, where the breaks would become part of the code.
	block := html.Div(CodeBlockContent(CodeBlockLineNumbers, "a\nb"))
	var b strings.Builder
	if err := block.Render(context.Background(), mx.NewCheckedWriter(&b).WithIndent("", "  ")); err != nil {
		t.Fatal(err)
	}
	out := b.String()
	start := strings.Index(out, "<pre")
	end := strings.Index(out, "</pre>")
	if start < 0 || end < 0 {
		t.Fatalf("no <pre> in %s", out)
	}
	if pre := out[start:end]; strings.Contains(pre, "\n") {
		t.Errorf("indentation inside <pre>: %s", pre)
	}
}

func TestCodeBlockLineVariants(t *testing.T) {
	added := render(t, CodeBlockLine(CodeBlockLineAdded, "x"))
	if !strings.Contains(added, "after:content-[&#39;+&#39;]") && !strings.Contains(added, "after:content-['+']") {
		t.Errorf("added line missing its + mark: %s", added)
	}
	removed := render(t, CodeBlockLine(CodeBlockLineRemoved, "x"))
	if !strings.Contains(removed, "bg-destructive/10") || !strings.Contains(removed, `data-variant="removed"`) {
		t.Errorf("removed line classes wrong: %s", removed)
	}
}

func TestCodeBlockContentWrapsLooseChildren(t *testing.T) {
	// Highlighted tokens are several children; each would become a row of
	// the <code> grid, so they are wrapped in one span.
	tokens := mx.Components{html.SpanClass("hl-keyword", "package"), mx.Text(" main")}
	out := render(t, CodeBlockContent(tokens))
	if !strings.Contains(out, `<code data-slot="code-block-code" class="`) ||
		!strings.Contains(out, `"><span><span class="hl-keyword">package</span> main</span></code>`) {
		t.Errorf("loose children should be wrapped in one span: %s", out)
	}
	// CodeBlockLines are the rows and stay direct children.
	lines := render(t, CodeBlockContent(CodeBlockLine("", "a"), CodeBlockLine("", "b")))
	if strings.Contains(lines, "<code data-slot=\"code-block-code\" class=\"grid min-h-lh min-w-max content-start in-data-numbered:grid-cols-[auto_1fr] in-data-wrap:min-w-0\"><span><span") {
		t.Errorf("lines must not be wrapped: %s", lines)
	}
}

func TestCodeBlockContentEmpty(t *testing.T) {
	// No children must still render <code></code>, not a void <code/>.
	if out := render(t, CodeBlockContent()); !strings.Contains(out, "></code></pre>") {
		t.Errorf("empty content should render an empty code element: %s", out)
	}
}

func TestCodeBlockCopyButtonMultilineValue(t *testing.T) {
	// The attribute escaping writes line breaks and tabs as character
	// references, so dataset.value reads back the value as written.
	out := render(t, CodeBlockCopyButton("func main() {\n\tfmt.Println(\"100%\")\n}"))
	want := `data-value="func main() {&#10;&#9;fmt.Println(&quot;100%&quot;)&#10;}"`
	if !strings.Contains(out, want) {
		t.Errorf("missing %s in %s", want, out)
	}
}

// TestCodeBlockCopyInBrowser runs the copy script in headless Chrome: a
// multi-line value reaches the clipboard as written, a block of numbered
// lines is copied without the numbers, the button shows the copied state
// and label, and both reset after two seconds.
func TestCodeBlockCopyInBrowser(t *testing.T) {
	body := mx.Components{
		CodeBlock("", "", html.ID("value"),
			CodeBlockActions(CodeBlockCopyButton("func main() {\n\tfmt.Println(\"100%\")\n}")),
			CodeBlockContent("ignored"),
		),
		CodeBlock("", "", html.ID("lines"),
			CodeBlockActions(CodeBlockCopyButton("")),
			CodeBlockContent(CodeBlockLineNumbers, "a := 1\n\nb := 2\n"),
		),
		CodeBlock("", "", html.ID("plain"),
			CodeBlockActions(CodeBlockCopyButton("")),
			CodeBlockContent("x := y"),
		),
	}
	const script = /*js*/ `
var copied=[];
Object.defineProperty(navigator,'clipboard',{configurable:true,value:{writeText:function(t){copied.push(t);return Promise.resolve();}}});
var btn=function(id){return document.querySelector('#'+id+' [data-slot=code-block-copy-button]');};
btn('value').click();
await wait(50);
out.value=copied[0];
out.copiedAttr=btn('value').hasAttribute('data-copied');
out.copiedLabel=btn('value').getAttribute('aria-label');
out.live=document.querySelector('[aria-live=polite]').textContent;
btn('lines').click();
await wait(50);
out.lines=copied[1];
btn('plain').click();
await wait(50);
out.plain=copied[2];
await wait(2100);
out.resetAttr=btn('value').hasAttribute('data-copied');
out.resetLabel=btn('value').getAttribute('aria-label');
out.resetLive=document.querySelector('[aria-live=polite]').textContent;
`
	got := runInChrome(t, body, script, nil)
	checkRecorded(t, got, map[string]any{
		"value":       "func main() {\n\tfmt.Println(\"100%\")\n}",
		"copiedAttr":  true,
		"copiedLabel": "Copied",
		"live":        "Copied",
		"lines":       "a := 1\n\nb := 2",
		"plain":       "x := y",
		"resetAttr":   false,
		"resetLabel":  "Copy",
		"resetLive":   "",
	})
}

// An empty line of highlighted tokens, as highlight's Lines returns it, is
// an empty line that keeps its newline.
func TestCodeBlockLineEmptyComponents(t *testing.T) {
	out := render(t, CodeBlockLine("", mx.Components{}))
	if !strings.Contains(out, "class=\"min-w-0\">\n</span>") {
		t.Errorf("an empty line must hold a newline: %s", out)
	}
}
