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

// Any data-direction="start" attrib makes the scroll-to-start button, not
// only the MessageScrollerStart constant.
func TestMessageScrollerButtonDirectionAttrib(t *testing.T) {
	out := render(t, MessageScrollerButton("", "", html.DataAttr("direction", "start")))
	if !strings.Contains(out, `data-direction="start"`) || strings.Contains(out, `data-direction="end"`) || !strings.Contains(out, "scrollTo({top:0") {
		t.Errorf("want the scroll-to-start button: %s", out)
	}
}

// TestMessageScrollerInBrowser runs StickToBottom in a MessageScroller in
// headless Chrome: it opens pinned to the latest message, follows an append
// while at the bottom, leaves a reader who scrolled up where they are, and
// the button brings them back and resumes following.
//
// Under headless Chrome's virtual time no animation frame runs, so a smooth
// scroll does not advance and a scroll does not fire its scroll event. The
// page makes smooth scrolls instant and fires the scroll events itself.
func TestMessageScrollerInBrowser(t *testing.T) {
	items := make([]any, 0, 31)
	items = append(items, html.Style("display:block"))
	for i := range 30 {
		items = append(items, MessageScrollerItem(html.Style("display:block;height:40px"), "message ", i))
	}
	body := MessageScroller(html.ID("ms"), html.Style("position:relative;height:200px;display:block"),
		MessageScrollerViewport(html.Style("height:200px;overflow-y:auto"),
			MessageScrollerContent(items...),
		),
		MessageScrollerButton("", ""),
		MessageScrollerButton("", "", MessageScrollerStart),
	)
	const script = /*js*/ `
var scrollTo=Element.prototype.scrollTo;
Element.prototype.scrollTo=function(o){if(o&&o.behavior==='smooth')o=Object.assign({},o,{behavior:'instant'});return scrollTo.call(this,o);};
var v=document.querySelector('#ms [data-slot=message-scroller-viewport]');
var c=document.querySelector('#ms [data-slot=message-scroller-content]');
var gap=function(){return Math.round(v.scrollHeight-v.clientHeight-v.scrollTop);};
var add=function(){var d=document.createElement('div');d.style.height='40px';d.textContent='new';c.appendChild(d);};
await wait(100);
out.openGap=gap();
out.openStuck=v.hasAttribute('data-stuck');
out.openAtStart=v.hasAttribute('data-at-start');
add();
await wait(100);
out.followGap=gap();
v.scrollTop=0;
v.dispatchEvent(new Event('scroll'));
await wait(100);
out.upStuck=v.hasAttribute('data-stuck');
out.atStart=v.hasAttribute('data-at-start');
add();
await wait(100);
out.upTop=v.scrollTop;
document.querySelector('#ms [data-slot=message-scroller-button]').click();
v.dispatchEvent(new Event('scroll'));
await wait(100);
out.buttonGap=gap();
out.buttonStuck=v.hasAttribute('data-stuck');
add();
await wait(100);
out.resumeGap=gap();
document.querySelector('#ms [data-direction=start]').click();
v.dispatchEvent(new Event('scroll'));
await wait(100);
out.startTop=v.scrollTop;
out.startAtStart=v.hasAttribute('data-at-start');
`
	got := runInChrome(t, body, script, nil)
	checkRecorded(t, got, map[string]any{
		"openGap":      float64(0),
		"openStuck":    true,
		"followGap":    float64(0),
		"upStuck":      false,
		"upTop":        float64(0),
		"buttonGap":    float64(0),
		"buttonStuck":  true,
		"resumeGap":    float64(0),
		"openAtStart":  false,
		"atStart":      true,
		"startTop":     float64(0),
		"startAtStart": true,
	})
}
