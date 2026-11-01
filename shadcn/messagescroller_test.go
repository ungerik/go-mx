package shadcn

import (
	"context"
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestMessageScroller(t *testing.T) {
	out := render(t, MessageScroller(html.Class("h-96"),
		MessageScrollerViewport(
			MessageScrollerContent(MessageScrollerItem("one"), MessageScrollerItem("two")),
		),
		MessageScrollerButton("", ""),
	))
	for _, want := range []string{
		`data-slot="message-scroller"`, `data-slot="message-scroller-viewport"`,
		`data-stick-to-bottom=""`, "window.mxStickToBottom",
		`data-slot="message-scroller-content"`, `data-slot="message-scroller-item"`,
		`data-slot="message-scroller-button"`, `data-direction="end"`,
		`data-variant="secondary"`, `data-size="icon-sm"`, "lucide-arrow-down",
		`<span class="sr-only">Scroll to end</span>`, "h-96",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	// The button hides on StickToBottom's data-stuck instead of the
	// primitive's data-active, which nothing here sets.
	if strings.Contains(out, "data-[active=") || !strings.Contains(out, "group-has-[>[data-slot=message-scroller-viewport][data-stuck]]/message-scroller:opacity-0") {
		t.Errorf("button state selectors not rewritten: %s", out)
	}
	for _, gone := range []string{"cn-", "scroll-fade", "scrollbar-thin", "data-autoscrolling", "data-pending-scroll"} {
		if strings.Contains(out, gone) {
			t.Errorf("unexpected %q in %s", gone, out)
		}
	}
}

func TestMessageScrollerViewportScriptFirst(t *testing.T) {
	// The script goes first, as in ScrollArea, so :last-child still matches
	// the last content row.
	out := render(t, MessageScrollerViewport(MessageScrollerContent("x")))
	_, after, _ := strings.Cut(out, ">")
	if !strings.HasPrefix(after, "<script") {
		t.Errorf("script should be the viewport's first child: %s", out)
	}
}

func TestMessageScrollerViewportThreshold(t *testing.T) {
	out := render(t, MessageScrollerViewport(StickToBottomThreshold(80), "x"))
	if strings.Count(out, "data-stick-to-bottom=") != 1 || !strings.Contains(out, `data-stick-to-bottom="80"`) {
		t.Errorf("caller threshold should replace the default: %s", out)
	}
}

func TestMessageScrollerButtonLabel(t *testing.T) {
	ctx := ContextWithLabels(context.Background(), Labels{MessageScrollerButton: "Zum Ende"})
	if out := renderWith(t, ctx, MessageScrollerButton("", "")); !strings.Contains(out, "Zum Ende") {
		t.Errorf("label not localized: %s", out)
	}
	// Caller children replace the default icon and text.
	if out := render(t, MessageScrollerButton(ButtonOutline, SizeSM, "New messages")); strings.Contains(out, "lucide-arrow-down") || !strings.Contains(out, "New messages") {
		t.Errorf("caller children should replace the default: %s", out)
	}
}
