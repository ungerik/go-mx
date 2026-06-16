package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// menubarScript implements the menubar's click-to-switch-without-clicking
// behavior. Hovering a trigger while any sibling menu is already open closes
// the open one and opens this one — matching the OS menubar idiom shadcn
// users expect. The script reads :popover-open and walks
// [data-slot="menubar-trigger"] siblings inside the same [role="menubar"].
const menubarScript = /*js*/ `if(!window.menubarHover){window.menubarHover=function(btn){var bar=btn.closest('[role="menubar"]');if(!bar)return;var triggers=Array.prototype.slice.call(bar.querySelectorAll('[data-slot="menubar-trigger"]'));var anyOpen=triggers.some(function(t){var id=t.getAttribute('popovertarget');var p=id?document.getElementById(id):null;return p&&p.matches(':popover-open');});if(!anyOpen)return;triggers.forEach(function(t){var id=t.getAttribute('popovertarget');var p=id?document.getElementById(id):null;if(p&&p.matches(':popover-open'))try{p.hidePopover();}catch(e){}});var myId=btn.getAttribute('popovertarget');var myP=myId?document.getElementById(myId):null;if(myP)try{myP.showPopover();}catch(e){}};}`

// menubarClasses is shadcn/ui's Menubar root class set, transcribed verbatim.
const menubarClasses = "bg-background flex h-9 items-center gap-1 rounded-md border p-1 shadow-xs"

// menubarTriggerClasses is shadcn/ui's MenubarTrigger class set with the
// Radix data-[state=open]:* rewritten to aria-expanded:* (the menuOpen
// script flips aria-expanded when the popover toggles).
const menubarTriggerClasses = "focus:bg-accent focus:text-accent-foreground aria-expanded:bg-accent aria-expanded:text-accent-foreground flex cursor-default items-center rounded-sm px-2 py-1 text-sm font-medium outline-hidden select-none"

// Menubar renders a coordinated row of menus as a <div role="menubar">. Place
// [MenubarMenu]s inside it. The menubarHover script is appended once so
// hovering a trigger while another menu is open switches to it.
func Menubar(attribsChildren ...any) *mx.Element {
	e := html.Div(attribsChildren...)
	if e.AttribIndex("role") < 0 {
		e.Attribs = append(e.Attribs, html.Role("menubar"))
	}
	e.Children = append(e.Children, html.Script(mx.Raw(menubarScript)))
	return finish(e, "menubar", menubarClasses)
}

// MenubarMenu wraps one menu (trigger + content) inside a [Menubar]. It is
// purely structural.
func MenubarMenu(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "menubar-menu", "")
}

// MenubarTrigger renders the button for one [MenubarMenu]. menuID is
// validated; it is the popovertarget and the anchor-name for [MenubarContent].
//
// Defaults (overridable): type="button", popovertarget={menuID},
// popovertargetaction="toggle", aria-haspopup="menu", aria-expanded="false",
// onmouseenter="menubarHover(this)", plus the anchor-name style.
func MenubarTrigger(menuID string, attribsChildren ...any) *mx.Element {
	if err := validateID(menuID); err != nil {
		return mx.NewErrElement(err)
	}
	e := html.Button(attribsChildren...)
	if e.AttribIndex("type") < 0 {
		e.Attribs = append(e.Attribs, html.Type("button"))
	}
	if e.AttribIndex("popovertarget") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("popovertarget", menuID))
	}
	if e.AttribIndex("popovertargetaction") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("popovertargetaction", "toggle"))
	}
	if e.AttribIndex("aria-haspopup") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("aria-haspopup", "menu"))
	}
	if e.AttribIndex("aria-expanded") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("aria-expanded", "false"))
	}
	if e.AttribIndex("onmouseenter") < 0 {
		e.Attribs = append(e.Attribs, html.OnMouseEnter("menubarHover(this)"))
	}
	mergeStyle(e, popoverAnchorStyle(menuID))
	return finish(e, "menubar-trigger", menubarTriggerClasses)
}

// MenubarContent renders the popover menu body for one [MenubarMenu]. side
// may be "" for the default (bottom).
func MenubarContent(menuID string, side PopoverSide, attribsChildren ...any) *mx.Element {
	return menuContent("menubar-content", menuID, side, true, false, attribsChildren...)
}

// MenubarItem renders one selectable item; same shape as [DropdownMenuItem].
func MenubarItem(attribsChildren ...any) *mx.Element {
	return menuItem("menubar-item", attribsChildren...)
}

// MenubarLabel renders a non-interactive section heading.
func MenubarLabel(attribsChildren ...any) *mx.Element {
	return menuLabel("menubar-label", attribsChildren...)
}

// MenubarSeparator renders a horizontal divider.
func MenubarSeparator(attribsChildren ...any) *mx.Element {
	return menuSeparator("menubar-separator", attribsChildren...)
}

// MenubarGroup groups related items (role="group").
func MenubarGroup(attribsChildren ...any) *mx.Element {
	return menuGroup("menubar-group", attribsChildren...)
}

// MenubarShortcut renders a keyboard-shortcut hint floated to the right.
func MenubarShortcut(attribsChildren ...any) *mx.Element {
	return menuShortcut("menubar-shortcut", attribsChildren...)
}

// MenubarCheckboxItem renders a checkbox-style item; checked drives the indicator.
func MenubarCheckboxItem(checked bool, attribsChildren ...any) *mx.Element {
	return menuCheckboxItem("menubar-checkbox-item", checked, attribsChildren...)
}

// MenubarRadioGroup groups a set of [MenubarRadioItem]s.
func MenubarRadioGroup(name string, attribsChildren ...any) *mx.Element {
	return menuRadioGroup("menubar-radio-group", name, attribsChildren...)
}

// MenubarRadioItem renders one radio-style item.
func MenubarRadioItem(name, value string, selected bool, attribsChildren ...any) *mx.Element {
	return menuRadioItem("menubar-radio-item", name, value, selected, attribsChildren...)
}

// MenubarSub wraps a sub-menu trigger and its sub-popover.
func MenubarSub(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "menubar-sub", "")
}

// MenubarSubTrigger renders a menu item that opens its sub-popover.
func MenubarSubTrigger(subID string, attribsChildren ...any) *mx.Element {
	return menuSubTrigger("menubar-sub-trigger", subID, attribsChildren...)
}

// MenubarSubContent renders the sub-menu body, anchored to its parent SubTrigger.
func MenubarSubContent(subID string, attribsChildren ...any) *mx.Element {
	return menuContent("menubar-sub-content", subID, PopoverRight, true, true, attribsChildren...)
}
