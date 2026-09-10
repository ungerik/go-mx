package shadcn

import (
	"context"
	"strings"
	"testing"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

func TestScrollAreaDefault(t *testing.T) {
	out := render(t, ScrollArea(html.Class("h-72 w-48"), "Long content here"))
	for _, want := range []string{
		`data-slot="scroll-area"`,
		"relative",
		"overflow-auto",
		"[scrollbar-width:thin]",
		"h-72 w-48",
		">Long content here<",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}

func TestScrollArea_NoStickToBottomScriptByDefault(t *testing.T) {
	// The script only ships with the areas that asked for the behavior, so a
	// page full of plain ScrollAreas carries no script at all.
	out := render(t, ScrollArea("content"))
	if strings.Contains(out, "<script") {
		t.Errorf("plain ScrollArea emitted a script: %s", out)
	}
	if strings.Contains(out, stickToBottomAttr) {
		t.Errorf("plain ScrollArea emitted %s: %s", stickToBottomAttr, out)
	}
}

func TestScrollArea_StickToBottomEmitsAttribAndScript(t *testing.T) {
	// The attribute is what the script selects on, so shipping one without the
	// other silently does nothing.
	out := render(t, ScrollArea(StickToBottom, "content"))
	for _, want := range []string{
		`data-stick-to-bottom=""`,
		"window.mxStickToBottom",
		"MutationObserver",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	// The script has to come after the content: it scrolls to the bottom when it
	// wires up, and would measure a partial height if it ran mid-parse.
	if strings.Index(out, ">content<") > strings.Index(out, "<script") {
		t.Errorf("script is emitted before the content: %s", out)
	}
}

func TestScrollArea_StickToBottomThreshold(t *testing.T) {
	// The threshold travels in the attribute value, so the same script serves
	// every instance and each reads its own tuning.
	out := render(t, ScrollArea(StickToBottomThreshold(96), "content"))
	if !strings.Contains(out, `data-stick-to-bottom="96"`) {
		t.Errorf("missing the threshold value in %s", out)
	}
	// 0 is meaningful (follow only when scrolled exactly to the bottom) and must
	// not be confused with "unset".
	out = render(t, ScrollArea(StickToBottomThreshold(0), "content"))
	if !strings.Contains(out, `data-stick-to-bottom="0"`) {
		t.Errorf("a zero threshold did not survive: %s", out)
	}
}

func TestScrollArea_NegativeThresholdIsAnError(t *testing.T) {
	// A negative distance has no meaning and would make the script fall back to
	// the default, hiding the mistake.
	var b strings.Builder
	err := ScrollArea(StickToBottomThreshold(-1), "content").
		Render(context.Background(), mx.NewCheckedWriter(&b))
	if err == nil {
		t.Errorf("a negative threshold rendered without an error: %s", b.String())
	}
}

func TestScrollArea_StickToBottomScriptGuardsItsDefinition(t *testing.T) {
	// Two anchored areas on one page must install the behavior once but wire up
	// both, which is why the definition is guarded and the scan is not.
	out := render(t, mx.Components{ScrollArea(StickToBottom, "a"), ScrollArea(StickToBottom, "b")})
	if got := strings.Count(out, "if(!window.mxStickToBottom)"); got != 2 {
		t.Fatalf("guard appears %d times, want 2 (once per emitted script)", got)
	}
	if got := strings.Count(out, "window.mxStickToBottomScan(document);</script>"); got != 2 {
		t.Errorf("the unguarded scan runs %d times, want once per instance", got)
	}
}
