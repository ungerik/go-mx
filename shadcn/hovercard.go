package shadcn

import (
	"strconv"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// hoverCardScript opens / closes a popover on hover with configurable
// delays. A single timer is kept per popover id so a quick mouse-out
// followed by mouse-in cancels the pending hide. The shape mirrors
// tooltipShow/Hide but adds setTimeout / clearTimeout for delays.
const hoverCardScript = /*js*/ `if(!window.hoverCardShow){window.hoverCardTimers={};window.hoverCardShow=function(t,id,d){clearTimeout(window.hoverCardTimers[id]);window.hoverCardTimers[id]=setTimeout(function(){var p=document.getElementById(id);if(p)try{p.showPopover();}catch(e){}},d);};window.hoverCardHide=function(t,id,d){clearTimeout(window.hoverCardTimers[id]);window.hoverCardTimers[id]=setTimeout(function(){var p=document.getElementById(id);if(p)try{p.hidePopover();}catch(e){}},d);};}`

// Default delays match shadcn/ui's HoverCard React defaults.
const (
	hoverCardDefaultOpenDelayMs  = 700
	hoverCardDefaultCloseDelayMs = 300
)

// hoverCardContentClasses is shadcn/ui's HoverCardContent class set with the
// Radix-only animation / slide / z-index / origin classes dropped (same
// pattern as Popover and Tooltip).
const hoverCardContentClasses = "bg-popover text-popover-foreground w-64 rounded-md border p-4 shadow-md outline-hidden"

// HoverCard wraps a trigger and its hover-card popover. It is purely
// structural; the trigger and content are linked by the id passed to
// [HoverCardTrigger] and [HoverCardContent].
//
// shadcn/ui's HoverCard is Radix-driven with openDelay / closeDelay props.
// This port replaces Radix with a native [popover] content element plus one
// shared inline script that opens/closes the popover with timers — same
// approach as Tooltip but with delays.
func HoverCard(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "hover-card", "inline-block")
}

// HoverCardTrigger renders the wrapping element that fires the hover card.
// openDelayMs / closeDelayMs are the timer values in milliseconds; pass 0 for
// either to select the shadcn defaults (700ms open, 300ms close). hoverCardID
// is validated.
//
// As with TooltipTrigger, the trigger is a <span> so any element can sit
// inside without nesting buttons. focusin/focusout (not focus/blur) are
// used so descendant focus events bubble.
func HoverCardTrigger(hoverCardID string, openDelayMs, closeDelayMs int, attribsChildren ...any) *mx.Element {
	if err := validateID(hoverCardID); err != nil {
		return mx.NewErrElement(err)
	}
	if openDelayMs <= 0 {
		openDelayMs = hoverCardDefaultOpenDelayMs
	}
	if closeDelayMs <= 0 {
		closeDelayMs = hoverCardDefaultCloseDelayMs
	}
	openStr := strconv.Itoa(openDelayMs)
	closeStr := strconv.Itoa(closeDelayMs)
	e := html.Span(attribsChildren...)
	if e.AttribIndex("aria-describedby") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("aria-describedby", hoverCardID))
	}
	if e.AttribIndex("onmouseover") < 0 {
		e.Attribs = append(e.Attribs, html.OnMouseOver("hoverCardShow(this,'"+hoverCardID+"',"+openStr+")"))
	}
	if e.AttribIndex("onmouseout") < 0 {
		e.Attribs = append(e.Attribs, html.OnMouseOut("hoverCardHide(this,'"+hoverCardID+"',"+closeStr+")"))
	}
	if e.AttribIndex("onfocusin") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("onfocusin", "hoverCardShow(this,'"+hoverCardID+"',"+openStr+")"))
	}
	if e.AttribIndex("onfocusout") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("onfocusout", "hoverCardHide(this,'"+hoverCardID+"',"+closeStr+")"))
	}
	// Name the trigger as the CSS anchor for HoverCardContent's position-anchor;
	// without it the content has no anchor and falls back to the viewport corner.
	mergeStyle(e, popoverAnchorStyle(hoverCardID))
	return finish(e, "hover-card-trigger", "")
}

// HoverCardContent renders the hover-card body as a <div popover="auto"
// id={hoverCardID}> with the shared anchor-position style and the
// hoverCardScript appended once. side may be "" for the default (bottom).
//
// Also wires onmouseover/onmouseout/onfocusin/onfocusout on the content
// itself so the popover stays open while the cursor is over it (and closes
// shortly after the cursor leaves).
func HoverCardContent(hoverCardID string, side PopoverSide, openDelayMs, closeDelayMs int, attribsChildren ...any) *mx.Element {
	if err := validateID(hoverCardID); err != nil {
		return mx.NewErrElement(err)
	}
	if openDelayMs <= 0 {
		openDelayMs = hoverCardDefaultOpenDelayMs
	}
	if closeDelayMs <= 0 {
		closeDelayMs = hoverCardDefaultCloseDelayMs
	}
	openStr := strconv.Itoa(openDelayMs)
	closeStr := strconv.Itoa(closeDelayMs)
	e := html.Div(append(attribsChildren, html.ScriptJS(hoverCardScript))...)
	if e.AttribIndex("id") < 0 {
		e.Attribs = append(e.Attribs, html.ID(hoverCardID))
	}
	if e.AttribIndex("popover") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("popover", "auto"))
	}
	if e.AttribIndex("role") < 0 {
		e.Attribs = append(e.Attribs, html.Role("dialog"))
	}
	// Re-fire the timers when the cursor enters the content so quick
	// trigger-to-content travel doesn't close the card.
	if e.AttribIndex("onmouseover") < 0 {
		e.Attribs = append(e.Attribs, html.OnMouseOver("hoverCardShow(this,'"+hoverCardID+"',"+openStr+")"))
	}
	if e.AttribIndex("onmouseout") < 0 {
		e.Attribs = append(e.Attribs, html.OnMouseOut("hoverCardHide(this,'"+hoverCardID+"',"+closeStr+")"))
	}
	mergeStyle(e, popoverContentStyle(hoverCardID, side))
	return finish(e, "hover-card-content", hoverCardContentClasses)
}
