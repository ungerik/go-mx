package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// navigationMenuTriggerClasses is shadcn/ui's NavigationMenuTrigger class set
// with the Radix data-[state=open]:* rewritten to aria-expanded:* (the
// menuOpen script flips aria-expanded when the popover toggles).
const navigationMenuTriggerClasses = "group inline-flex h-9 w-max items-center justify-center rounded-md bg-background px-4 py-2 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:text-accent-foreground aria-expanded:bg-accent/50 disabled:pointer-events-none disabled:opacity-50 outline-none"

// navigationMenuContentClasses is the popover content panel — wider than a
// regular menu since NavigationMenu content tends to be rich (grids of links,
// previews).
const navigationMenuContentClasses = "bg-popover text-popover-foreground w-auto min-w-[15rem] rounded-md border p-2 shadow-md outline-hidden"

// navigationMenuLinkClasses is shadcn/ui's NavigationMenuLink class set
// (matching the Trigger's hover / focus look). data-active drives the
// "current page" look — controlled by the active bool on [NavigationMenuLink].
const navigationMenuLinkClasses = "block select-none space-y-1 rounded-md p-3 leading-none no-underline outline-hidden transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:text-accent-foreground data-[active=true]:bg-accent/50"

// NavigationMenu renders a site-navigation landmark as a <nav>. Place a
// [NavigationMenuList] inside it. shadcn/ui's NavigationMenuViewport (a
// shared content area shared across items) and NavigationMenuIndicator (the
// arrow that tracks the active trigger) are not ported: each item's content
// is its own popover, and active styling is per-link via the active bool on
// [NavigationMenuLink].
func NavigationMenu(attribsChildren ...any) *mx.Element {
	e := html.Nav(attribsChildren...)
	if e.AttribIndex("aria-label") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("aria-label", "Main"))
	}
	return finish(e, "navigation-menu", "relative flex max-w-max flex-1 items-center justify-center")
}

// NavigationMenuList renders the horizontal list of nav items as a <ul>.
func NavigationMenuList(attribsChildren ...any) *mx.Element {
	return finish(html.UL(attribsChildren...), "navigation-menu-list",
		"group flex flex-1 list-none items-center justify-center gap-1")
}

// NavigationMenuItem renders one nav item as an <li>. Inside it goes either a
// direct [NavigationMenuLink] or a [NavigationMenuTrigger] +
// [NavigationMenuContent] pair for items with a dropdown.
func NavigationMenuItem(attribsChildren ...any) *mx.Element {
	return finish(html.LI(attribsChildren...), "navigation-menu-item", "")
}

// NavigationMenuTrigger renders the button that opens an item's content
// popover. navID is validated; it is the popovertarget and the anchor-name
// for [NavigationMenuContent]. A chevron-down icon is appended that rotates
// when the popover is open (group-aria-expanded:rotate-180).
func NavigationMenuTrigger(navID string, attribsChildren ...any) *mx.Element {
	validateID(navID)
	e := html.Button(attribsChildren...)
	if e.AttribIndex("type") < 0 {
		e.Attribs = append(e.Attribs, html.Type("button"))
	}
	if e.AttribIndex("popovertarget") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("popovertarget", navID))
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
	mergeStyle(e, popoverAnchorStyle(navID))
	chevron := icon("chevron-down",
		"relative top-[1px] ml-1 size-3 transition duration-300 [button[aria-expanded=true]>&]:rotate-180",
		svgPath("m6 9 6 6 6-6"))
	e.Children = append(e.Children, chevron)
	return finish(e, "navigation-menu-trigger", navigationMenuTriggerClasses)
}

// NavigationMenuContent renders the popover content panel for one nav item.
// side may be "" for the default (bottom). The menuOpen ontoggle handler is
// reused so the trigger's aria-expanded flips for the chevron rotation.
func NavigationMenuContent(navID string, side PopoverSide, attribsChildren ...any) *mx.Element {
	validateID(navID)
	e := html.Div(attribsChildren...)
	if e.AttribIndex("id") < 0 {
		e.Attribs = append(e.Attribs, html.ID(navID))
	}
	if e.AttribIndex("popover") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("popover", "auto"))
	}
	if e.AttribIndex("role") < 0 {
		e.Attribs = append(e.Attribs, html.Role("menu"))
	}
	if e.AttribIndex("ontoggle") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("ontoggle", "menuOpen(event)"))
	}
	mergeStyle(e, popoverContentStyle(navID, side))
	e.Children = append(e.Children, html.Script(mx.Raw(menuScript)))
	return finish(e, "navigation-menu-content", navigationMenuContentClasses)
}

// NavigationMenuLink renders a direct navigation link as an <a>. active marks
// the current page: it sets data-active="true" (for the .data-[active=true]:
// styling) and aria-current="page". Pass html.HRef for the target.
func NavigationMenuLink(active bool, attribsChildren ...any) *mx.Element {
	e := html.A(attribsChildren...)
	if e.AttribIndex("data-active") < 0 {
		e.Attribs = append(e.Attribs, html.DataAttr("active", boolStr(active)))
	}
	if active && e.AttribIndex("aria-current") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("aria-current", "page"))
	}
	return finish(e, "navigation-menu-link", navigationMenuLinkClasses)
}
