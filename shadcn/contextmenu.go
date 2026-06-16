package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// contextMenuScript positions a popover at the cursor on right-click and
// opens it. CSS anchor positioning is element-relative, not cursor-relative,
// so the script sets inline position via clientX/clientY. Also resets the
// popover's default margin: auto and any position-anchor that might be set.
const contextMenuScript = /*js*/ `if(!window.contextMenuOpen){window.contextMenuOpen=function(ev,id){ev.preventDefault();var p=document.getElementById(id);if(!p)return;p.style.position='fixed';p.style.top=ev.clientY+'px';p.style.left=ev.clientX+'px';p.style.margin='0';p.style.positionAnchor='none';try{p.showPopover();}catch(e){}};}`

// ContextMenu wraps a right-click trigger area and its popover menu. It is
// purely structural; the trigger and content are linked by the menuID passed
// to [ContextMenuTrigger] and [ContextMenuContent].
//
// shadcn/ui's ContextMenu is Radix-driven (Provider + Root + Trigger + Portal
// + Content). This port replaces all of that with a native [popover] content
// element plus a tiny inline script that prevents the browser's native
// context menu, positions the popover at the cursor, and shows it. Keyboard
// navigation reuses the same menuKeyNav script as [DropdownMenu].
func ContextMenu(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "context-menu", "")
}

// ContextMenuTrigger renders the wrapping area whose right-click opens the
// menu. It is a <div> (not a button) so any content — a card, a list item, a
// canvas — can serve as the trigger zone. menuID is validated.
//
// Defaults: oncontextmenu="contextMenuOpen(event,'{menuID}')". The script is
// appended once as a child of the trigger.
func ContextMenuTrigger(menuID string, attribsChildren ...any) *mx.Element {
	if err := validateID(menuID); err != nil {
		return mx.NewErrElement(err)
	}
	e := html.Div(attribsChildren...)
	if e.AttribIndex("oncontextmenu") < 0 {
		e.Attribs = append(e.Attribs, html.OnContextMenu("contextMenuOpen(event,'"+menuID+"')"))
	}
	e.Children = append(e.Children, html.Script(mx.Raw(contextMenuScript)))
	return finish(e, "context-menu-trigger", "")
}

// ContextMenuContent renders the popover menu body. Unlike
// [DropdownMenuContent] it carries no CSS anchor positioning — the
// contextMenuOpen script sets top/left from the cursor at click time.
func ContextMenuContent(menuID string, attribsChildren ...any) *mx.Element {
	return menuContent("context-menu-content", menuID, "", false, false, attribsChildren...)
}

// ContextMenuItem renders one selectable item; same shape as [DropdownMenuItem].
func ContextMenuItem(attribsChildren ...any) *mx.Element {
	return menuItem("context-menu-item", attribsChildren...)
}

// ContextMenuLabel renders a non-interactive section heading.
func ContextMenuLabel(attribsChildren ...any) *mx.Element {
	return menuLabel("context-menu-label", attribsChildren...)
}

// ContextMenuSeparator renders a horizontal divider.
func ContextMenuSeparator(attribsChildren ...any) *mx.Element {
	return menuSeparator("context-menu-separator", attribsChildren...)
}

// ContextMenuGroup groups related items (role="group").
func ContextMenuGroup(attribsChildren ...any) *mx.Element {
	return menuGroup("context-menu-group", attribsChildren...)
}

// ContextMenuShortcut renders a keyboard-shortcut hint floated to the right.
func ContextMenuShortcut(attribsChildren ...any) *mx.Element {
	return menuShortcut("context-menu-shortcut", attribsChildren...)
}

// ContextMenuCheckboxItem renders a checkbox-style item; checked drives the
// indicator.
func ContextMenuCheckboxItem(checked bool, attribsChildren ...any) *mx.Element {
	return menuCheckboxItem("context-menu-checkbox-item", checked, attribsChildren...)
}

// ContextMenuRadioGroup groups a set of [ContextMenuRadioItem]s.
func ContextMenuRadioGroup(name string, attribsChildren ...any) *mx.Element {
	return menuRadioGroup("context-menu-radio-group", name, attribsChildren...)
}

// ContextMenuRadioItem renders one radio-style item.
func ContextMenuRadioItem(name, value string, selected bool, attribsChildren ...any) *mx.Element {
	return menuRadioItem("context-menu-radio-item", name, value, selected, attribsChildren...)
}

// ContextMenuSub wraps a sub-menu trigger and its sub-popover.
func ContextMenuSub(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "context-menu-sub", "")
}

// ContextMenuSubTrigger renders a menu item that opens its sub-popover.
func ContextMenuSubTrigger(subID string, attribsChildren ...any) *mx.Element {
	return menuSubTrigger("context-menu-sub-trigger", subID, attribsChildren...)
}

// ContextMenuSubContent renders the sub-menu body, anchored to its trigger
// (the parent ContextMenu's content is positioned at the cursor, but the
// sub-menu is anchored to its parent SubTrigger like a normal sub-popover).
func ContextMenuSubContent(subID string, attribsChildren ...any) *mx.Element {
	return menuContent("context-menu-sub-content", subID, PopoverRight, true, true, attribsChildren...)
}
