package shadcn

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestBubble(t *testing.T) {
	out := render(t, BubbleGroup(
		Bubble("", "", BubbleContent("Hello"), BubbleReactions("", "", "👍")),
	))
	for _, want := range []string{
		`data-slot="bubble-group"`, `data-slot="bubble"`, `data-variant="default"`,
		`data-align="start"`, `<div data-slot="bubble-content"`,
		`data-slot="bubble-reactions"`, `data-side="bottom"`, `data-align="end"`,
		"*:data-[slot=bubble-content]:bg-primary", "bottom-0", "right-3",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	if strings.Contains(out, "cn-") {
		t.Errorf("unresolved cn-* token leaked into output: %s", out)
	}
}

func TestBubbleVariants(t *testing.T) {
	for _, v := range BubbleVariant("").Enums() {
		out := render(t, Bubble(v, ChatAlignEnd, BubbleContent("x")))
		if !strings.Contains(out, `data-variant="`+string(v)+`"`) || !strings.Contains(out, `data-align="end"`) {
			t.Errorf("%s: missing data attributes: %s", v, out)
		}
	}
	ghost := render(t, Bubble(BubbleGhost, "", BubbleContent("x")))
	if !strings.Contains(ghost, "*:data-[slot=bubble-content]:p-0") {
		t.Errorf("ghost variant should drop the content padding: %s", ghost)
	}
}

func TestBubbleContentRenderTag(t *testing.T) {
	if out := render(t, BubbleContent(html.Type("button"), "Yes")); !strings.HasPrefix(out, "<button ") {
		t.Errorf("type should render a <button>: %s", out)
	}
	if out := render(t, BubbleContent(html.HRef("/doc"), "Doc")); !strings.HasPrefix(out, "<a ") {
		t.Errorf("href should render an <a>: %s", out)
	}
}

func TestBubbleReactionsTopStart(t *testing.T) {
	out := render(t, BubbleReactions(BubbleReactionsTop, ChatAlignStart, "x"))
	if !strings.Contains(out, "-translate-y-3/4") || !strings.Contains(out, "left-3") || strings.Contains(out, "right-3") {
		t.Errorf("top/start reactions classes wrong: %s", out)
	}
}
