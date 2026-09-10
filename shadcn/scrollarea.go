package shadcn

import (
	"strconv"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// scrollAreaClasses is the base class set for ScrollArea. shadcn's component
// nests a Radix Root, Viewport, ScrollBar and Corner; this port flattens them
// to a single overflow <div>, so the class string combines Root's relative,
// Viewport's outline/ring focus handling, and a small set of arbitrary
// utilities that style the native scrollbar (the native equivalent of Radix's
// rendered ScrollBar element).
const scrollAreaClasses = "relative overflow-auto focus-visible:ring-ring/50 rounded-[inherit] transition-[color,box-shadow] outline-none focus-visible:ring-[3px] focus-visible:outline-1 [scrollbar-width:thin] [scrollbar-color:var(--color-border)_transparent] [&::-webkit-scrollbar]:size-2.5 [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-border"

// ScrollArea renders a shadcn/ui scroll area as a single <div> with overflow:
// auto and CSS-styled scrollbars. shadcn's Root/Viewport/ScrollBar/Corner
// structure collapses to one element here — a deliberate divergence, mirroring
// how AlertDialogOverlay is not ported (the native equivalent is a pseudo, not
// an element).
//
// shadcn's ScrollBar component is intentionally NOT exported by this package:
// in the native port the scrollbar is the ::-webkit-scrollbar pseudo-element
// (and Firefox's scrollbar-width / scrollbar-color), styled via the utility
// classes above. For finer control, override scrollbar styles in your own CSS.
// Pass [StickToBottom] (or [StickToBottomThreshold]) to make the area follow
// content that grows after the page was delivered; see those docs for what that
// adds to the markup.
func ScrollArea(attribsChildren ...any) *mx.Element {
	e := html.Div(attribsChildren...)
	if e.AttribIndex(stickToBottomAttr) >= 0 {
		// Ship the behavior with the component, like Tabs does with
		// tabsSelectScript: the script defines itself once per page and every
		// instance re-runs only the scan. As the last child it also runs after
		// the initial content is parsed, so the first scroll-to-bottom measures
		// the full height.
		e.Children = append(e.Children, html.ScriptJS(stickToBottomScript))
	}
	return finish(e, "scroll-area", scrollAreaClasses)
}

// stickToBottomAttr marks a [ScrollArea] as scroll-anchored. It is what the
// script selects on, and its value is the near-bottom threshold in pixels — an
// empty value means [StickToBottomDefaultThresholdPx].
const stickToBottomAttr = "data-stick-to-bottom"

// StickToBottomDefaultThresholdPx is how close to the bottom counts as
// "following" when [StickToBottom] is used without an explicit threshold. It
// trades two failure modes against each other: too small and a user who nudges
// the wheel stops receiving new content, too large and a user reading history
// gets yanked back down.
const StickToBottomDefaultThresholdPx = 48

// StickToBottom makes a [ScrollArea] follow content appended after the initial
// render — the behavior a chat transcript needs, where its absence reads as a
// bug rather than a missing feature — while leaving a user who has scrolled up
// where they are.
//
// It renders as a data attribute and makes [ScrollArea] emit a small inline
// script (once per page, like [Tabs]) that watches the element for DOM changes.
// The element is followed while its scroll position is within
// [StickToBottomDefaultThresholdPx] of the bottom; use [StickToBottomThreshold]
// for a different distance. Scrolling up stops the following, scrolling back
// down resumes it.
//
// The script tracks DOM mutations rather than htmx events, so it works for
// content arriving over an SSE stream ([mx.SSEResponse]), from an ordinary htmx
// swap, or from any other script. It wires up areas swapped in later via the
// htmx:load event, and needs no htmx to work on a normally loaded page.
const StickToBottom = mx.ConstAttrib(stickToBottomAttr + "=")

// StickToBottomThreshold is [StickToBottom] with an explicit near-bottom
// threshold in pixels: how far from the bottom a user may be and still count as
// following the content. A threshold of 0 follows only when scrolled exactly to
// the bottom. A negative value defers an error to render time.
func StickToBottomThreshold(px int) mx.Attrib {
	if px < 0 {
		return mx.ErrAttribf(stickToBottomAttr, "shadcn: StickToBottomThreshold needs a non-negative pixel distance, got %d", px)
	}
	return mx.NewAttrib(stickToBottomAttr, strconv.Itoa(px))
}

// stickToBottomScript is the once-emitted client behavior behind
// [StickToBottom]. Two details it depends on:
//
// Whether to follow is recorded on scroll, not measured when the content
// changes. By the time a mutation is observed the content has already grown, so
// a distance measured then says nothing about where the user was before it —
// with a large append every scroll position looks "far from the bottom".
//
// The definition is guarded by if(!window.mxStickToBottom) so several
// ScrollAreas on a page install it once, but the scan runs unguarded on every
// instance, which is what wires each element up at parse time.
//
// It is a var rather than a const only because it interpolates
// [StickToBottomDefaultThresholdPx], keeping that default in one place.
var stickToBottomScript = /*js*/ `if(!window.mxStickToBottom){window.mxStickToBottom=function(el){if(el.mxStuck)return;el.mxStuck=1;var t=parseInt(el.dataset.stickToBottom,10);if(!(t>=0))t=` + strconv.Itoa(StickToBottomDefaultThresholdPx) + `;var stuck=true;el.addEventListener('scroll',function(){stuck=el.scrollHeight-el.scrollTop-el.clientHeight<=t;},{passive:true});new MutationObserver(function(){if(stuck)el.scrollTop=el.scrollHeight;}).observe(el,{childList:true,subtree:true,characterData:true});el.scrollTop=el.scrollHeight;};window.mxStickToBottomScan=function(r){if(!r||!r.querySelectorAll)return;if(r.matches&&r.matches('[` + stickToBottomAttr + `]'))window.mxStickToBottom(r);r.querySelectorAll('[` + stickToBottomAttr + `]').forEach(window.mxStickToBottom);};document.addEventListener('DOMContentLoaded',function(){window.mxStickToBottomScan(document);});document.addEventListener('htmx:load',function(e){window.mxStickToBottomScan(e.target);});}window.mxStickToBottomScan(document);`
