package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// DropdownMenu wraps a trigger and a popover menu. shadcn/ui's DropdownMenu
// is Radix-driven; this port replaces Radix with the native Popover API +
// CSS Anchor Positioning (see [Popover]) plus one shared inline script for
// keyboard navigation (see menu.go for menuKeyNav).
//
// Sub-menus, checkbox items, radio groups and keyboard shortcuts are all
// supported as separate parts; see the matching DropdownMenu* functions.
func DropdownMenu(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "dropdown-menu", "")
}

// DropdownMenuTrigger renders the button that toggles the menu. menuID is
// validated; it is the popovertarget the trigger toggles and the anchor-name
// referenced by [DropdownMenuContent].
//
// Defaults (overridable): type="button", popovertarget={menuID},
// popovertargetaction="toggle", aria-haspopup="menu", aria-expanded="false",
// plus the anchor-name style.
func DropdownMenuTrigger(menuID string, attribsChildren ...any) *mx.Element {
	if err := validateID(menuID); err != nil {
		return mx.NewErrElement(err)
	}
	e := html.Button(attribsChildren...)
	if e.AttribIndex("type") < 0 {
		e.Attribs = append(e.Attribs, html.Type("button"))
	}
	if e.AttribIndex("popovertarget") < 0 {
		e.Attribs = append(e.Attribs, html.PopoverTarget(menuID))
	}
	if e.AttribIndex("popovertargetaction") < 0 {
		e.Attribs = append(e.Attribs, html.PopoverTargetActionToggle)
	}
	if e.AttribIndex("aria-haspopup") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("aria-haspopup", "menu"))
	}
	if e.AttribIndex("aria-expanded") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("aria-expanded", "false"))
	}
	mergeStyle(e, popoverAnchorStyle(menuID))
	return finish(e, "dropdown-menu-trigger", "")
}

// DropdownMenuContent renders the popover menu body with id={menuID},
// popover="auto", role="menu", and the shared menuKeyNav script wired via
// onkeydown. side may be "" for the default (bottom).
func DropdownMenuContent(menuID string, side PopoverSide, attribsChildren ...any) *mx.Element {
	return menuContent("dropdown-menu-content", menuID, side, true, false, attribsChildren...)
}

// DropdownMenuItem renders one selectable menu entry as a <div role=menuitem
// tabindex=-1>. Callers attach an action with html.OnClick or hx.* attribs;
// the menu's keyboard script handles ArrowUp/Down/Home/End/typeahead/Escape
// across all items.
func DropdownMenuItem(attribsChildren ...any) *mx.Element {
	return menuItem("dropdown-menu-item", attribsChildren...)
}

// DropdownMenuLabel renders a non-interactive section heading.
func DropdownMenuLabel(attribsChildren ...any) *mx.Element {
	return menuLabel("dropdown-menu-label", attribsChildren...)
}

// DropdownMenuSeparator renders a horizontal divider between menu sections.
func DropdownMenuSeparator(attribsChildren ...any) *mx.Element {
	return menuSeparator("dropdown-menu-separator", attribsChildren...)
}

// DropdownMenuGroup groups related items together (role="group").
func DropdownMenuGroup(attribsChildren ...any) *mx.Element {
	return menuGroup("dropdown-menu-group", attribsChildren...)
}

// DropdownMenuShortcut renders a keyboard-shortcut hint floated to the right
// of a menu item.
func DropdownMenuShortcut(attribsChildren ...any) *mx.Element {
	return menuShortcut("dropdown-menu-shortcut", attribsChildren...)
}

// DropdownMenuCheckboxItem renders a checkbox-style menu item. checked
// drives aria-checked and whether a check-icon indicator renders on the
// left. The caller's onclick (or hx-*) handles the actual state change —
// the indicator reflects the rendered server-side state.
func DropdownMenuCheckboxItem(checked bool, attribsChildren ...any) *mx.Element {
	return menuCheckboxItem("dropdown-menu-checkbox-item", checked, attribsChildren...)
}

// DropdownMenuRadioGroup groups a set of [DropdownMenuRadioItem]s sharing
// one name. name is validated.
func DropdownMenuRadioGroup(name string, attribsChildren ...any) *mx.Element {
	return menuRadioGroup("dropdown-menu-radio-group", name, attribsChildren...)
}

// DropdownMenuRadioItem renders one radio-style menu item. name must match
// the enclosing [DropdownMenuRadioGroup]; value identifies this item;
// selected drives aria-checked and whether a dot indicator renders.
func DropdownMenuRadioItem(name, value string, selected bool, attribsChildren ...any) *mx.Element {
	return menuRadioItem("dropdown-menu-radio-item", name, value, selected, attribsChildren...)
}

// DropdownMenuSub wraps a sub-menu trigger and its sub-popover. Purely
// structural — like [DropdownMenu] but nested inside a parent menu's content.
func DropdownMenuSub(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "dropdown-menu-sub", "")
}

// DropdownMenuSubTrigger renders a menu item that opens its sub-popover.
// subID is validated and used as the popovertarget. The trigger gets a
// chevron-right indicator on the right; ArrowRight while focused opens the
// sub-popover, ArrowLeft inside the sub closes it (both via menuKeyNav).
func DropdownMenuSubTrigger(subID string, attribsChildren ...any) *mx.Element {
	return menuSubTrigger("dropdown-menu-sub-trigger", subID, attribsChildren...)
}

// DropdownMenuSubContent renders the sub-menu body. Same shape as
// [DropdownMenuContent] but marked data-submenu="true" so menuKeyNav's
// ArrowLeft handler refocuses the parent trigger. Defaults to side=right.
func DropdownMenuSubContent(subID string, attribsChildren ...any) *mx.Element {
	return menuContent("dropdown-menu-sub-content", subID, PopoverRight, true, true, attribsChildren...)
}
