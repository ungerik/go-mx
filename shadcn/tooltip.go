package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// tooltipScript wires hover and focus triggering for tooltips. The HTML
// Popover API has no declarative hover-to-open attribute (it's all
// click/popovertarget); this script bridges that gap. try/catch wraps
// showPopover/hidePopover, which throw when called on an already-open or
// already-closed popover (e.g. quick re-entry of the trigger area).
const tooltipScript = /*js*/ `if(!window.tooltipShow){window.tooltipShow=function(t,id){var p=document.getElementById(id);if(p)try{p.showPopover();}catch(e){}};window.tooltipHide=function(t,id){var p=document.getElementById(id);if(p)try{p.hidePopover();}catch(e){}};}`

// tooltipContentClasses is shadcn/ui's TooltipContent class set with the
// Radix-only animation / slide / z-index / transform-origin classes removed.
// The native [popover] handles the open / close lifecycle and renders in the
// top layer; CSS anchor positioning handles placement.
const tooltipContentClasses = "bg-primary text-primary-foreground w-fit rounded-md px-3 py-1.5 text-xs text-balance"

// Tooltip wraps a trigger and its tooltip popover. It is purely structural;
// the trigger and content are linked by the tooltip id passed to
// [TooltipTrigger] and [TooltipContent].
//
// shadcn/ui's Tooltip is Radix-driven (Provider + Root + Trigger + Portal +
// Content + Arrow). This port replaces Radix with a native [popover] content
// element plus one shared inline script that opens the popover on
// mouseover/focusin and closes it on mouseout/focusout. Provider, Portal and
// Arrow are not ported.
func Tooltip(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "tooltip", "inline-block")
}

// TooltipTrigger renders the wrapping element that fires the tooltip on
// hover or focus. It is a <span> rather than a <button> so the caller can
// put any element (a button, a piece of text, an icon) inside without
// nesting a button-in-a-button. tooltipID is validated.
//
// Defaults (overridable): aria-describedby={tooltipID}, the four event
// handlers (onmouseover/onmouseout/onfocusin/onfocusout) wired to the shared
// tooltipShow/tooltipHide script, and the anchor-name style that TooltipContent
// anchors to. focusin/focusout (not focus/blur) are used because they bubble —
// so any descendant getting focus opens the tooltip.
func TooltipTrigger(tooltipID string, attribsChildren ...any) *mx.Element {
	if err := validateID(tooltipID); err != nil {
		return mx.NewErrElement(err)
	}
	e := html.Span(attribsChildren...)
	if e.AttribIndex("aria-describedby") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("aria-describedby", tooltipID))
	}
	if e.AttribIndex("onmouseover") < 0 {
		e.Attribs = append(e.Attribs, html.OnMouseOver("tooltipShow(this,'"+tooltipID+"')"))
	}
	if e.AttribIndex("onmouseout") < 0 {
		e.Attribs = append(e.Attribs, html.OnMouseOut("tooltipHide(this,'"+tooltipID+"')"))
	}
	if e.AttribIndex("onfocusin") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("onfocusin", "tooltipShow(this,'"+tooltipID+"')"))
	}
	if e.AttribIndex("onfocusout") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("onfocusout", "tooltipHide(this,'"+tooltipID+"')"))
	}
	// Name the trigger as the CSS anchor for TooltipContent's position-anchor;
	// without it the content has no anchor and falls back to the viewport corner.
	mergeStyle(e, popoverAnchorStyle(tooltipID))
	return finish(e, "tooltip-trigger", "")
}

// TooltipContent renders the tooltip body as a <div popover="auto"
// id={tooltipID}> with the shared anchor-position style and the shared
// tooltipScript appended once. side may be "" for the default (top —
// matching shadcn's default tooltip placement).
func TooltipContent(tooltipID string, side PopoverSide, attribsChildren ...any) *mx.Element {
	if err := validateID(tooltipID); err != nil {
		return mx.NewErrElement(err)
	}
	if side == "" {
		side = PopoverTop
	}
	e := html.Div(append(attribsChildren, html.ScriptJS(tooltipScript))...)
	if e.AttribIndex("id") < 0 {
		e.Attribs = append(e.Attribs, html.ID(tooltipID))
	}
	if e.AttribIndex("popover") < 0 {
		e.Attribs = append(e.Attribs, html.PopoverAuto)
	}
	if e.AttribIndex("role") < 0 {
		e.Attribs = append(e.Attribs, html.Role("tooltip"))
	}
	mergeStyle(e, popoverContentStyle(tooltipID, side))
	return finish(e, "tooltip-content", tooltipContentClasses)
}
