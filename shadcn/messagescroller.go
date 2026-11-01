package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// MessageScroller is a Go port of shadcn/ui's MessageScroller, the scrolling
// frame of a chat transcript: it follows new messages while the reader is at
// the bottom, leaves a reader who scrolled up where they are, and offers a
// button back to the latest message. Class strings are reconstructed from
// the post-July-2026 registry (message-scroller.tsx skeleton + the
// style-vega.css .cn-message-scroller* rules) — see "Upstream restructure
// (July 2026)" in TODOS.md.
//
// Upstream wraps its @shadcn/react/message-scroller primitive. Here the
// following is [StickToBottom], which the viewport always carries, and the
// button is shown by CSS from StickToBottom's data-stuck attribute: no
// script beyond StickToBottom's own. Compose:
//
//	MessageScroller(
//		MessageScrollerViewport(
//			MessageScrollerContent(MessageScrollerItem(…), …),
//		),
//		MessageScrollerButton("", ""),
//	)
//
// Give the MessageScroller a height (or put it in a flex column that gives it
// one), since the viewport scrolls within it.
func MessageScroller(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "message-scroller",
		"group/message-scroller relative flex size-full min-h-0 flex-col overflow-hidden")
}

// MessageScrollerViewport renders the scrolling element of a
// [MessageScroller]. It carries [StickToBottom] — pass
// [StickToBottomThreshold] for another near-bottom distance — and emits
// StickToBottom's script as its first child, as [ScrollArea] does.
//
// Upstream's scroll-fade-b edge mask, from shadcn's own tailwind.css, is
// emitted only with [UseUtilities]; its autoscrolling and pending-scroll
// states (from the primitive) are not reproduced.
func MessageScrollerViewport(attribsChildren ...any) *mx.Element {
	e := html.Div(attribsChildren...)
	if e.AttribIndex(stickToBottomAttr) < 0 {
		e.Attribs = append(e.Attribs, StickToBottom)
	}
	e.Children = append(mx.Components{html.ScriptJS(stickToBottomScript)}, e.Children...)
	return finish(e, "message-scroller-viewport",
		"size-full min-h-0 min-w-0 [scrollbar-width:thin] [scrollbar-gutter:stable] overflow-y-auto overscroll-contain contain-content"+utilityClasses(" scroll-fade-b", ""))
}

// MessageScrollerContent stacks a [MessageScroller]'s items, at least as tall
// as the viewport.
func MessageScrollerContent(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "message-scroller-content",
		"flex h-max min-h-full flex-col gap-8")
}

// MessageScrollerItem holds one entry of a [MessageScrollerContent],
// typically a [Message]. Items scrolled out of view skip rendering
// (content-visibility: auto, with a 10rem placeholder height), which keeps a
// long transcript cheap.
func MessageScrollerItem(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "message-scroller-item",
		"min-w-0 shrink-0 [contain-intrinsic-size:auto_10rem] [content-visibility:auto]")
}

// messageScrollerButtonOnClick scrolls the button's viewport to the bottom;
// StickToBottom's scroll listener then resumes following there.
const messageScrollerButtonOnClick = /*js*/ `var v=this.closest('[data-slot=message-scroller]').querySelector('[data-slot=message-scroller-viewport]');v.scrollTo({top:v.scrollHeight,behavior:'smooth'})`

// messageScrollerStartButtonOnClick scrolls the button's viewport to the
// top.
const messageScrollerStartButtonOnClick = /*js*/ `this.closest('[data-slot=message-scroller]').querySelector('[data-slot=message-scroller-viewport]').scrollTo({top:0,behavior:'smooth'})`

// MessageScrollerStart makes a [MessageScrollerButton] upstream's
// direction="start" button: centred over the viewport's top edge, shown
// while the viewport is not scrolled to the top (StickToBottom's
// data-at-start), and scrolling it there, for a transcript read from the
// top.
const MessageScrollerStart = mx.ConstAttrib("data-direction=start")

// MessageScrollerButton renders the "scroll to end" button of a
// [MessageScroller], centred over the viewport's bottom edge. It is shown
// while the viewport is not following new content, that is while it lacks
// StickToBottom's data-stuck, and scrolls it back to the bottom. With
// [MessageScrollerStart] it is the "scroll to start" button instead. variant
// and size may be "" for [ButtonSecondary] and [SizeIconSM], upstream's
// defaults. Without children it shows an arrow and the screen-reader text
// "Scroll to end" or "Scroll to start" (localizable via [Labels]); a
// caller-supplied onclick replaces the scrolling. Hidden, it is also
// invisible, so the keyboard does not tab onto it.
func MessageScrollerButton(variant ButtonVariant, size ButtonSize, attribsChildren ...any) *mx.Element {
	e := Button(variant.ValidOr(ButtonSecondary), size.ValidOr(SizeIconSM), attribsChildren...)
	start := false
	if i := e.AttribIndex("data-direction"); i >= 0 {
		start = e.Attribs[i] == MessageScrollerStart
		e.Attribs = append(e.Attribs[:i], e.Attribs[i+1:]...)
	}
	if len(e.Children) == 0 {
		label := func(l Labels) string { return l.MessageScrollerButton }
		if start {
			label = func(l Labels) string { return l.MessageScrollerStartButton }
		}
		e.Children = mx.Components{
			iconArrowDown(),
			html.Span(html.Class("sr-only"), labelText(label)),
		}
	}
	onclick, direction, classes := messageScrollerButtonOnClick, "end", messageScrollerButtonClasses
	if start {
		onclick, direction, classes = messageScrollerStartButtonOnClick, "start", messageScrollerStartButtonClasses
	}
	if e.AttribIndex("onclick") < 0 {
		e.Attribs = append(e.Attribs, html.OnClick(onclick))
	}
	e.Attribs = append(e.Attribs, html.DataAttr("direction", direction))
	return finish(e, "message-scroller-button", classes)
}

// messageScrollerButtonClasses is upstream's button class string with its
// data-[active=…] states rewritten to the viewport's data-stuck: active is
// the base state, and the group-has-[…data-stuck…] variant hides the button
// (upstream's data-[active=false]). border is added: upstream's
// border-border relies on the border width of its Button base, which this
// port's new-york-v4 Button does not have. visibility joins the transition
// and the hidden state, so a hidden button leaves the tab order once it has
// faded out (upstream's stays focusable).
const messageScrollerButtonClasses = "absolute inset-s-1/2 bottom-4 -translate-x-1/2 rtl:translate-x-1/2 border border-border bg-background text-foreground transition-[translate,scale,opacity,visibility] transition-discrete duration-200 hover:bg-muted hover:text-foreground visible translate-y-0 scale-100 opacity-100 ease-[cubic-bezier(0.23,1,0.32,1)] group-has-[>[data-slot=message-scroller-viewport][data-stuck]]/message-scroller:invisible group-has-[>[data-slot=message-scroller-viewport][data-stuck]]/message-scroller:pointer-events-none group-has-[>[data-slot=message-scroller-viewport][data-stuck]]/message-scroller:scale-95 group-has-[>[data-slot=message-scroller-viewport][data-stuck]]/message-scroller:opacity-0 group-has-[>[data-slot=message-scroller-viewport][data-stuck]]/message-scroller:duration-400 group-has-[>[data-slot=message-scroller-viewport][data-stuck]]/message-scroller:ease-[cubic-bezier(0.7,0,0.84,0)] group-has-[>[data-slot=message-scroller-viewport][data-stuck]]/message-scroller:translate-y-full"

// messageScrollerStartButtonClasses is upstream's direction=start variant of
// [messageScrollerButtonClasses]: over the top edge, its arrow turned up,
// hidden while the viewport has StickToBottom's data-at-start.
const messageScrollerStartButtonClasses = "absolute inset-s-1/2 top-4 -translate-x-1/2 rtl:translate-x-1/2 border border-border bg-background text-foreground transition-[translate,scale,opacity,visibility] transition-discrete duration-200 hover:bg-muted hover:text-foreground [&_svg]:rotate-180 visible translate-y-0 scale-100 opacity-100 ease-[cubic-bezier(0.23,1,0.32,1)] group-has-[>[data-slot=message-scroller-viewport][data-at-start]]/message-scroller:invisible group-has-[>[data-slot=message-scroller-viewport][data-at-start]]/message-scroller:pointer-events-none group-has-[>[data-slot=message-scroller-viewport][data-at-start]]/message-scroller:scale-95 group-has-[>[data-slot=message-scroller-viewport][data-at-start]]/message-scroller:opacity-0 group-has-[>[data-slot=message-scroller-viewport][data-at-start]]/message-scroller:duration-400 group-has-[>[data-slot=message-scroller-viewport][data-at-start]]/message-scroller:ease-[cubic-bezier(0.7,0,0.84,0)] group-has-[>[data-slot=message-scroller-viewport][data-at-start]]/message-scroller:-translate-y-full"
