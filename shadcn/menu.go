package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// menuScript is the once-emitted client function for menu keyboard nav and
// open-focus. menuKeyNav is wired via onkeydown on each menu Content; menuOpen
// via ontoggle. The script handles:
//
//   - ArrowDown / ArrowUp: roving focus through [role^=menuitem] items.
//   - Home / End: jump to first / last item.
//   - Letter / digit keys: typeahead (focus first item whose text starts).
//   - Escape: close the containing popover.
//   - ArrowRight on a submenu trigger: open its popover, focus first item.
//   - ArrowLeft inside a submenu (data-submenu="true"): close, refocus parent.
//   - On popover open: focus the first menuitem.
const menuScript = /*js*/ `if(!window.menuKeyNav){window.menuKeyNav=function(ev){var menu=ev.currentTarget;var items=Array.prototype.slice.call(menu.querySelectorAll('[role^="menuitem"]:not([aria-disabled="true"])'));if(!items.length)return;var active=document.activeElement;var idx=items.indexOf(active);var key=ev.key;if(key==='ArrowDown'){ev.preventDefault();items[(idx+1)%items.length].focus();}else if(key==='ArrowUp'){ev.preventDefault();items[(idx-1+items.length)%items.length].focus();}else if(key==='Home'){ev.preventDefault();items[0].focus();}else if(key==='End'){ev.preventDefault();items[items.length-1].focus();}else if(key==='Escape'){ev.preventDefault();var p=menu.closest('[popover]');if(p)try{p.hidePopover();}catch(e){}}else if(key==='ArrowRight'&&active&&active.hasAttribute('popovertarget')){ev.preventDefault();var sub=document.getElementById(active.getAttribute('popovertarget'));if(sub){try{sub.showPopover();}catch(e){}setTimeout(function(){var f=sub.querySelector('[role^="menuitem"]:not([aria-disabled="true"])');if(f)f.focus();},0);}}else if(key==='ArrowLeft'){var pop=menu.closest('[popover]');if(pop&&pop.getAttribute('data-submenu')==='true'){ev.preventDefault();try{pop.hidePopover();}catch(e){}var t=document.querySelector('[popovertarget="'+pop.id+'"]');if(t)t.focus();}}else if(key.length===1&&/^[a-zA-Z0-9]$/.test(key)){var k=key.toLowerCase();var start=(idx+1)%items.length;for(var i=0;i<items.length;i++){var c=items[(start+i)%items.length];if((c.textContent||'').trim().toLowerCase().startsWith(k)){c.focus();break;}}}};window.menuOpen=function(ev){var p=ev.currentTarget;var open=ev.newState==='open';var trig=document.querySelector('[popovertarget="'+p.id+'"]');if(trig)trig.setAttribute('aria-expanded',open?'true':'false');if(open){var i=p.querySelector('[role^="menuitem"]:not([aria-disabled="true"])');if(i)i.focus();}};}`

// menuContentClasses is the shared shadcn menu content class set, used by
// DropdownMenuContent / ContextMenuContent / MenubarContent / SubContent.
// Radix-only z-50, max-h custom-property, origin custom-property classes
// are dropped (top-layer + CSS anchor positioning handle them).
const menuContentClasses = "bg-popover text-popover-foreground min-w-[8rem] overflow-x-hidden overflow-y-auto rounded-md border p-1 shadow-md"

// menuItemClasses styles a menu item. shadcn's data-[disabled]:* /
// data-[variant=destructive]:* selectors are kept since they target
// caller-supplied attributes, not Radix data-state.
const menuItemClasses = "focus:bg-accent focus:text-accent-foreground relative flex cursor-default items-center gap-2 rounded-sm px-2 py-1.5 text-sm outline-hidden select-none data-[disabled]:pointer-events-none data-[disabled]:opacity-50 [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4"

// menuCheckRadioItemClasses adds left padding for the indicator slot.
const menuCheckRadioItemClasses = menuItemClasses + " pl-8"

// menuLabelClasses styles a menu label heading.
const menuLabelClasses = "px-2 py-1.5 text-sm font-medium"

// menuSeparatorClasses styles a menu separator line.
const menuSeparatorClasses = "bg-border -mx-1 my-1 h-px"

// menuShortcutClasses styles a keyboard-shortcut hint on the right of an item.
const menuShortcutClasses = "ml-auto text-xs tracking-widest text-muted-foreground"

// menuContent builds the popover content <div> shared by DropdownMenuContent,
// ContextMenuContent, MenubarContent and the *SubContent variants. slot is
// the data-slot value; anchored=true wires CSS anchor positioning (used by
// dropdown / menubar / sub-menus); anchored=false skips it so the script can
// pixel-position at the cursor (used by [ContextMenu]). submenu marks this
// as a sub-popover (the menuKeyNav ArrowLeft handler then closes it and
// refocuses the parent trigger).
func menuContent(slot, popoverID string, side PopoverSide, anchored, submenu bool, attribsChildren ...any) *mx.Element {
	validateID(popoverID)
	e := html.Div(attribsChildren...)
	if e.AttribIndex("id") < 0 {
		e.Attribs = append(e.Attribs, html.ID(popoverID))
	}
	if e.AttribIndex("popover") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("popover", "auto"))
	}
	if e.AttribIndex("role") < 0 {
		e.Attribs = append(e.Attribs, html.Role("menu"))
	}
	if e.AttribIndex("tabindex") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("tabindex", "-1"))
	}
	if e.AttribIndex("onkeydown") < 0 {
		e.Attribs = append(e.Attribs, html.OnKeyDown("menuKeyNav(event)"))
	}
	if e.AttribIndex("ontoggle") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("ontoggle", "menuOpen(event)"))
	}
	if submenu {
		e.Attribs = append(e.Attribs, html.DataAttr("submenu", "true"))
	}
	if anchored {
		mergeStyle(e, popoverContentStyle(popoverID, side))
	}
	e.Children = append(e.Children, html.Script(mx.Raw(menuScript)))
	return finish(e, slot, menuContentClasses)
}

// menuItem builds a generic menu item <div role=menuitem tabindex=-1>. slot
// distinguishes Dropdown / Context / Menubar items.
func menuItem(slot string, attribsChildren ...any) *mx.Element {
	e := html.Div(attribsChildren...)
	if e.AttribIndex("role") < 0 {
		e.Attribs = append(e.Attribs, html.Role("menuitem"))
	}
	if e.AttribIndex("tabindex") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("tabindex", "-1"))
	}
	return finish(e, slot, menuItemClasses)
}

// menuCheckboxItem builds a checkbox-style menu item. checked drives
// aria-checked and whether the iconCheck indicator is rendered in the
// left-padding slot.
func menuCheckboxItem(slot string, checked bool, attribsChildren ...any) *mx.Element {
	e := html.Div(attribsChildren...)
	if e.AttribIndex("role") < 0 {
		e.Attribs = append(e.Attribs, html.Role("menuitemcheckbox"))
	}
	if e.AttribIndex("aria-checked") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("aria-checked", boolStr(checked)))
	}
	if e.AttribIndex("tabindex") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("tabindex", "-1"))
	}
	// Prepend the indicator span so it sits in the pl-8 slot.
	indicator := html.Span(html.Class("absolute left-2 flex size-3.5 items-center justify-center"))
	if checked {
		indicator.Children = mx.Components{iconCheck()}
	}
	e.Children = append(mx.Components{indicator}, e.Children...)
	return finish(e, slot, menuCheckRadioItemClasses)
}

// menuRadioItem builds a radio-style menu item. name + value identify the
// item within a [menuRadioGroup]; selected drives aria-checked and the
// iconCircle indicator. The caller's onclick (or hx-*) handles the actual
// selection state change.
func menuRadioItem(slot, name, value string, selected bool, attribsChildren ...any) *mx.Element {
	validateID(name)
	e := html.Div(attribsChildren...)
	if e.AttribIndex("role") < 0 {
		e.Attribs = append(e.Attribs, html.Role("menuitemradio"))
	}
	if e.AttribIndex("aria-checked") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("aria-checked", boolStr(selected)))
	}
	if e.AttribIndex("tabindex") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("tabindex", "-1"))
	}
	e.Attribs = append(e.Attribs,
		html.DataAttr("radio-group", name),
		html.DataAttr("value", value),
	)
	indicator := html.Span(html.Class("absolute left-2 flex size-3.5 items-center justify-center [&>svg]:size-2"))
	if selected {
		indicator.Children = mx.Components{iconCircle()}
	}
	e.Children = append(mx.Components{indicator}, e.Children...)
	return finish(e, slot, menuCheckRadioItemClasses)
}

// menuLabel renders a non-interactive label inside a menu.
func menuLabel(slot string, attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), slot, menuLabelClasses)
}

// menuSeparator renders a horizontal divider between menu sections.
func menuSeparator(slot string, attribsChildren ...any) *mx.Element {
	e := html.Div(attribsChildren...)
	if e.AttribIndex("role") < 0 {
		e.Attribs = append(e.Attribs, html.Role("separator"))
	}
	return finish(e, slot, menuSeparatorClasses)
}

// menuGroup renders a logical group of menu items (role="group").
func menuGroup(slot string, attribsChildren ...any) *mx.Element {
	e := html.Div(attribsChildren...)
	if e.AttribIndex("role") < 0 {
		e.Attribs = append(e.Attribs, html.Role("group"))
	}
	return finish(e, slot, "")
}

// menuRadioGroup renders a group of radio items sharing one name.
func menuRadioGroup(slot, name string, attribsChildren ...any) *mx.Element {
	validateID(name)
	e := html.Div(attribsChildren...)
	if e.AttribIndex("role") < 0 {
		e.Attribs = append(e.Attribs, html.Role("group"))
	}
	if e.AttribIndex("data-radio-group") < 0 {
		e.Attribs = append(e.Attribs, html.DataAttr("radio-group", name))
	}
	return finish(e, slot, "")
}

// menuShortcut renders a keyboard-shortcut hint floated to the right of an item.
func menuShortcut(slot string, attribsChildren ...any) *mx.Element {
	return finish(html.Span(attribsChildren...), slot, menuShortcutClasses)
}

// menuSubTrigger builds a menu item that opens a sub-popover. It uses
// popoverButton-equivalent attribs plus the menuitem role and a chevron-right
// indicator on the right.
func menuSubTrigger(slot, popoverID string, attribsChildren ...any) *mx.Element {
	validateID(popoverID)
	e := html.Div(attribsChildren...)
	if e.AttribIndex("role") < 0 {
		e.Attribs = append(e.Attribs, html.Role("menuitem"))
	}
	if e.AttribIndex("tabindex") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("tabindex", "-1"))
	}
	if e.AttribIndex("popovertarget") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("popovertarget", popoverID))
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
	mergeStyle(e, popoverAnchorStyle(popoverID))
	e.Children = append(e.Children, html.Span(html.Class("ml-auto"), iconChevronRight()))
	return finish(e, slot, menuItemClasses)
}
