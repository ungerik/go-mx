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
// Upstream's scroll-fade-b edge mask (from shadcn's own tailwind.css) and
// its autoscrolling and pending-scroll states (from the primitive) are not
// reproduced.
func MessageScrollerViewport(attribsChildren ...any) *mx.Element {
	e := html.Div(attribsChildren...)
	if e.AttribIndex(stickToBottomAttr) < 0 {
		e.Attribs = append(e.Attribs, StickToBottom)
	}
	e.Children = append(mx.Components{html.ScriptJS(stickToBottomScript)}, e.Children...)
	return finish(e, "message-scroller-viewport",
		"size-full min-h-0 min-w-0 [scrollbar-width:thin] [scrollbar-gutter:stable] overflow-y-auto overscroll-contain contain-content")
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

// MessageScrollerButton renders the "scroll to end" button of a
// [MessageScroller], centred over the viewport's bottom edge. It is shown
// while the viewport is not following new content, that is while it lacks
// StickToBottom's data-stuck, and scrolls it back to the bottom. variant and
// size may be "" for [ButtonSecondary] and [SizeIconSM], upstream's
// defaults. Without children it shows an arrow and the screen-reader text
// "Scroll to end" (localizable via [Labels]); a caller-supplied onclick
// replaces the scrolling.
//
// Upstream's direction="start" button, which scrolls to the top, has no
// counterpart: StickToBottom tracks only the bottom.
func MessageScrollerButton(variant ButtonVariant, size ButtonSize, attribsChildren ...any) *mx.Element {
	e := Button(variant.ValidOr(ButtonSecondary), size.ValidOr(SizeIconSM), attribsChildren...)
	if len(e.Children) == 0 {
		e.Children = mx.Components{
			iconArrowDown(),
			html.Span(html.Class("sr-only"), labelText(func(l Labels) string { return l.MessageScrollerButton })),
		}
	}
	if e.AttribIndex("onclick") < 0 {
		e.Attribs = append(e.Attribs, html.OnClick(messageScrollerButtonOnClick))
	}
	e.Attribs = append(e.Attribs, html.DataAttr("direction", "end"))
	return finish(e, "message-scroller-button", messageScrollerButtonClasses)
}

// messageScrollerButtonClasses is upstream's button class string with its
// data-[active=…] states rewritten to the viewport's data-stuck: active is
// the base state, and the group-has-[…data-stuck…] variant hides the button
// (upstream's data-[active=false]). Upstream's direction=start classes are
// dropped with that direction. border is added: upstream's border-border
// relies on the border width of its Button base, which this port's
// new-york-v4 Button does not have.
const messageScrollerButtonClasses = "absolute inset-s-1/2 bottom-4 -translate-x-1/2 rtl:translate-x-1/2 border border-border bg-background text-foreground transition-[translate,scale,opacity] duration-200 hover:bg-muted hover:text-foreground translate-y-0 scale-100 opacity-100 ease-[cubic-bezier(0.23,1,0.32,1)] group-has-[>[data-slot=message-scroller-viewport][data-stuck]]/message-scroller:pointer-events-none group-has-[>[data-slot=message-scroller-viewport][data-stuck]]/message-scroller:scale-95 group-has-[>[data-slot=message-scroller-viewport][data-stuck]]/message-scroller:opacity-0 group-has-[>[data-slot=message-scroller-viewport][data-stuck]]/message-scroller:duration-400 group-has-[>[data-slot=message-scroller-viewport][data-stuck]]/message-scroller:ease-[cubic-bezier(0.7,0,0.84,0)] group-has-[>[data-slot=message-scroller-viewport][data-stuck]]/message-scroller:translate-y-full"
