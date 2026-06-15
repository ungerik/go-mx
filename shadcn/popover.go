package shadcn

import (
	"context"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// PopoverSide selects on which side of its trigger a popover-positioned
// content element renders. The default for "" is PopoverBottom.
//
// Side placement is realized with CSS Anchor Positioning (anchor-name +
// position-anchor + position-area). Browsers with anchor positioning
// (Chromium 125+, Safari 26+) place the content next to its trigger; browsers
// without it (Firefox as of mid-2026) render the popover at its default
// centered position — still dismissible and functional, just not anchored.
// See shadcn/README.md for the rationale.
type PopoverSide string // TODO use go-enum

const (
	// PopoverTop renders the content above its trigger.
	PopoverTop PopoverSide = "top"
	// PopoverRight renders the content to the right of its trigger.
	PopoverRight PopoverSide = "right"
	// PopoverBottom renders the content below its trigger (the default).
	PopoverBottom PopoverSide = "bottom"
	// PopoverLeft renders the content to the left of its trigger.
	PopoverLeft PopoverSide = "left"
)

// normPopoverSide maps an empty or unknown side to the default (bottom).
func normPopoverSide(s PopoverSide) PopoverSide {
	switch s {
	case PopoverTop, PopoverRight, PopoverLeft:
		return s
	default:
		return PopoverBottom
	}
}

// popoverAnchorStyle is the trigger-side style fragment that names this
// element as a CSS anchor for its associated popover content.
func popoverAnchorStyle(name string) string {
	return "anchor-name: --" + name
}

// popoverContentStyle is the content-side style fragment that anchors the
// popover to its trigger and places it on the requested side with a 4px gap.
// Uses CSS position-area; Chrome 129+ and Safari 26+ understand it.
func popoverContentStyle(name string, side PopoverSide) string {
	base := "position-anchor: --" + name + "; position-area: "
	switch normPopoverSide(side) {
	case PopoverTop:
		return base + "top; margin-bottom: 4px"
	case PopoverRight:
		return base + "right; margin-left: 4px"
	case PopoverLeft:
		return base + "left; margin-right: 4px"
	default:
		return base + "bottom; margin-top: 4px"
	}
}

// popoverContentClasses is shadcn/ui's PopoverContent class set with the
// Radix-only z-index, animation and slide-from-side classes dropped — a
// native [popover] renders in the top layer (no z-index needed) and CSS
// anchor positioning handles side placement (no slide-from-* transforms).
const popoverContentClasses = "bg-popover text-popover-foreground w-72 rounded-md border p-4 shadow-md outline-hidden"

// Popover wraps a trigger and its popover content. It is purely structural;
// the trigger and content are linked by the popover id passed to
// [PopoverTrigger] and [PopoverContent], not by DOM nesting.
//
// shadcn/ui's Popover is Radix-driven. This port replaces Radix with the
// native HTML Popover API (popover attribute + popovertarget) and CSS Anchor
// Positioning, so no client-side framework is needed for open / close /
// light-dismiss / Escape-to-close / top-layer rendering.
func Popover(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "popover", "")
}

// PopoverTrigger renders the button that toggles the popover with the given
// id via popovertarget. popoverID is validated; it is interpolated into the
// trigger's popovertarget attribute and the anchor-name CSS, so it must be a
// safe HTML id.
//
// Defaults (overridable via the AttribIndex < 0 idiom): type="button",
// popovertarget={popoverID}, popovertargetaction="toggle",
// aria-haspopup="dialog", aria-expanded="false", and the anchor-name style.
// Pass any html.* attributes (or html.Class for styling) the normal way; the
// anchor-name fragment is merged into a caller-supplied style.
func PopoverTrigger(popoverID string, attribsChildren ...any) *mx.Element {
	validateID(popoverID)
	e := html.Button(attribsChildren...)
	if e.AttribIndex("type") < 0 {
		e.Attribs = append(e.Attribs, html.Type("button"))
	}
	if e.AttribIndex("popovertarget") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("popovertarget", popoverID))
	}
	if e.AttribIndex("popovertargetaction") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("popovertargetaction", "toggle"))
	}
	if e.AttribIndex("aria-haspopup") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("aria-haspopup", "dialog"))
	}
	if e.AttribIndex("aria-expanded") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("aria-expanded", "false"))
	}
	mergeStyle(e, popoverAnchorStyle(popoverID))
	return finish(e, "popover-trigger", "")
}

// PopoverContent renders the popover body with id={popoverID}, popover="auto"
// (so the browser handles open / close / light-dismiss / Escape) and the
// CSS anchor-position style. side may be "" for the default (bottom).
func PopoverContent(popoverID string, side PopoverSide, attribsChildren ...any) *mx.Element {
	validateID(popoverID)
	e := html.Div(attribsChildren...)
	if e.AttribIndex("id") < 0 {
		e.Attribs = append(e.Attribs, html.ID(popoverID))
	}
	if e.AttribIndex("popover") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("popover", "auto"))
	}
	if e.AttribIndex("role") < 0 {
		e.Attribs = append(e.Attribs, html.Role("dialog"))
	}
	mergeStyle(e, popoverContentStyle(popoverID, side))
	return finish(e, "popover-content", popoverContentClasses)
}

// mergeStyle appends fragment to the element's style attribute. If no style
// is present a new one is created; otherwise fragment is concatenated to the
// existing value with a "; " separator. Mirrors how the other components
// treat default attributes — caller-supplied content stays, the component's
// additions are appended.
func mergeStyle(e *mx.Element, fragment string) {
	if fragment == "" {
		return
	}
	if i := e.AttribIndex("style"); i >= 0 {
		// Concatenate; finish() will not dedupe style because it's a single
		// attribute name from the caller's perspective. We rewrite it here.
		existing, _ := e.Attribs[i].AttribValue(context.Background())
		e.Attribs[i] = html.Style(existing + "; " + fragment)
		return
	}
	e.Attribs = append(e.Attribs, html.Style(fragment))
}
