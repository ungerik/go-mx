package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// Sidebar is a Go port of shadcn/ui's Sidebar — the most composite component in
// the set. shadcn drives its open/collapsed state with React context, a cookie
// and a keyboard shortcut, and swaps to a Sheet on mobile. This port keeps the
// markup and the expand/icon-collapse behavior natively:
//
//   - [SidebarProvider] is the flex wrapper carrying the --sidebar-width CSS
//     variables and a data-state ("expanded" | "collapsed") that the rest read
//     via the group/sidebar-wrapper group-data variant.
//   - One shared sidebarScript toggles data-state, persists it to the
//     `sidebar_state` cookie, restores it on load, and binds Cmd/Ctrl+B — the
//     native equivalent of shadcn's context + cookie + shortcut.
//   - Collapsing shrinks the sidebar to --sidebar-width-icon; labels clip
//     (overflow-hidden) leaving the leading icons, group labels fade, and
//     sub-menus hide.
//
// shadcn's floating/inset variants and the mobile-becomes-Sheet behavior are
// not reproduced; the offcanvas/icon collapse and the full part set are.
const (
	sidebarWidth     = "16rem"
	sidebarWidthIcon = "3rem"
)

const sidebarScript = /*js*/ `if(!window.sidebarToggle){window.sidebarToggle=function(){var w=document.querySelector('[data-slot=sidebar-wrapper]');var c=w.dataset.state==='collapsed';w.dataset.state=c?'expanded':'collapsed';document.cookie='sidebar_state='+c+';path=/;max-age=604800';};document.addEventListener('keydown',function(e){if((e.metaKey||e.ctrlKey)&&(e.key==='b'||e.key==='B')){e.preventDefault();window.sidebarToggle();}});var w=document.querySelector('[data-slot=sidebar-wrapper]');var m=document.cookie.match(/sidebar_state=(true|false)/);if(w&&m&&m[1]==='false')w.dataset.state='collapsed';}`

// SidebarProvider is the flex wrapper around the [Sidebar] and [SidebarInset].
// It defaults to data-state="expanded" and sets the sidebar width variables.
func SidebarProvider(attribsChildren ...any) *mx.Element {
	e := html.Div(attribsChildren...)
	if e.AttribIndex("data-state") < 0 {
		e.Attribs = append(e.Attribs, html.DataAttr("state", "expanded"))
	}
	if e.AttribIndex("style") < 0 {
		e.Attribs = append(e.Attribs, html.Style("--sidebar-width: "+sidebarWidth+"; --sidebar-width-icon: "+sidebarWidthIcon))
	}
	e.Children = append(e.Children, html.Script(mx.Raw(sidebarScript)))
	return finish(e, "sidebar-wrapper", "group/sidebar-wrapper flex min-h-svh w-full")
}

// Sidebar is the collapsible sidebar column.
func Sidebar(attribsChildren ...any) *mx.Element {
	return finish(html.Aside(attribsChildren...), "sidebar",
		"bg-sidebar text-sidebar-foreground flex h-full flex-col overflow-hidden border-r transition-[width] duration-200 ease-linear w-[var(--sidebar-width)] group-data-[state=collapsed]/sidebar-wrapper:w-[var(--sidebar-width-icon)]")
}

// SidebarTrigger toggles the sidebar between expanded and collapsed.
func SidebarTrigger(attribsChildren ...any) *mx.Element {
	e := html.Button(attribsChildren...)
	if e.AttribIndex("type") < 0 {
		e.Attribs = append(e.Attribs, html.Type("button"))
	}
	if e.AttribIndex("onclick") < 0 {
		e.Attribs = append(e.Attribs, html.OnClick("sidebarToggle()"))
	}
	if e.AttribIndex("aria-label") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("aria-label", "Toggle Sidebar"))
	}
	if len(e.Children) == 0 {
		e.Children = mx.Components{iconPanelLeft()}
	}
	return finish(e, "sidebar-trigger", ButtonClasses(ButtonGhost, SizeIcon)+" size-7")
}

// SidebarInset is the main content area beside the sidebar.
func SidebarInset(attribsChildren ...any) *mx.Element {
	return finish(html.Main(attribsChildren...), "sidebar-inset",
		"bg-background relative flex flex-1 flex-col overflow-auto")
}

// SidebarHeader is the top section of the sidebar.
func SidebarHeader(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "sidebar-header", "flex flex-col gap-2 p-2")
}

// SidebarContent is the scrollable middle of the sidebar.
func SidebarContent(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "sidebar-content",
		"flex min-h-0 flex-1 flex-col gap-2 overflow-auto p-2")
}

// SidebarFooter is the bottom section of the sidebar.
func SidebarFooter(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "sidebar-footer", "flex flex-col gap-2 p-2")
}

// SidebarSeparator is a divider sized to the sidebar's inner padding.
func SidebarSeparator(attribsChildren ...any) *mx.Element {
	e := html.Div(attribsChildren...)
	if e.AttribIndex("role") < 0 {
		e.Attribs = append(e.Attribs, html.Role("separator"))
	}
	return finish(e, "sidebar-separator", "bg-sidebar-border mx-2 h-px w-auto shrink-0")
}

// SidebarGroup is one labeled section within [SidebarContent].
func SidebarGroup(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "sidebar-group", "relative flex w-full min-w-0 flex-col p-2")
}

// SidebarGroupLabel labels a [SidebarGroup]; it fades out when collapsed.
func SidebarGroupLabel(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "sidebar-group-label",
		"text-sidebar-foreground/70 flex h-8 shrink-0 items-center rounded-md px-2 text-xs font-medium transition-opacity group-data-[state=collapsed]/sidebar-wrapper:opacity-0")
}

// SidebarGroupContent wraps a group's menu.
func SidebarGroupContent(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "sidebar-group-content", "w-full text-sm")
}

// SidebarMenu is the <ul> of menu items.
func SidebarMenu(attribsChildren ...any) *mx.Element {
	return finish(html.Element("ul", attribsChildren...), "sidebar-menu", "flex w-full min-w-0 flex-col gap-1")
}

// SidebarMenuItem is one <li> in a [SidebarMenu].
func SidebarMenuItem(attribsChildren ...any) *mx.Element {
	return finish(html.Element("li", attribsChildren...), "sidebar-menu-item", "group/menu-item relative")
}

// SidebarMenuButton is the clickable nav button. Mark the current item active
// with html.DataAttr("active", "true"). A leading icon stays visible when the
// sidebar collapses; the label clips.
func SidebarMenuButton(attribsChildren ...any) *mx.Element {
	e := html.Button(attribsChildren...)
	if e.AttribIndex("type") < 0 {
		e.Attribs = append(e.Attribs, html.Type("button"))
	}
	return finish(e, "sidebar-menu-button",
		"flex w-full items-center gap-2 overflow-hidden rounded-md p-2 text-left text-sm outline-hidden transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground data-[active=true]:bg-sidebar-accent data-[active=true]:font-medium [&>svg]:size-4 [&>svg]:shrink-0 group-data-[state=collapsed]/sidebar-wrapper:justify-center group-data-[state=collapsed]/sidebar-wrapper:[&>span]:hidden")
}

// SidebarMenuBadge is a count/indicator pinned to the right of a menu item.
func SidebarMenuBadge(attribsChildren ...any) *mx.Element {
	return finish(html.Span(attribsChildren...), "sidebar-menu-badge",
		"text-sidebar-foreground pointer-events-none absolute top-1.5 right-1 flex h-5 min-w-5 items-center justify-center rounded-md px-1 text-xs font-medium tabular-nums select-none group-data-[state=collapsed]/sidebar-wrapper:hidden")
}

// SidebarMenuSub is the nested <ul> of sub-items; hidden when collapsed.
func SidebarMenuSub(attribsChildren ...any) *mx.Element {
	return finish(html.Element("ul", attribsChildren...), "sidebar-menu-sub",
		"border-sidebar-border mx-3.5 flex min-w-0 flex-col gap-1 border-l px-2.5 py-0.5 group-data-[state=collapsed]/sidebar-wrapper:hidden")
}

// SidebarMenuSubItem is one <li> in a [SidebarMenuSub].
func SidebarMenuSubItem(attribsChildren ...any) *mx.Element {
	return finish(html.Element("li", attribsChildren...), "sidebar-menu-sub-item", "relative")
}

// SidebarMenuSubButton is a sub-item link. Mark the active one with
// html.DataAttr("active", "true").
func SidebarMenuSubButton(attribsChildren ...any) *mx.Element {
	return finish(html.A(attribsChildren...), "sidebar-menu-sub-button",
		"text-sidebar-foreground flex h-7 min-w-0 items-center gap-2 overflow-hidden rounded-md px-2 text-sm hover:bg-sidebar-accent hover:text-sidebar-accent-foreground data-[active=true]:bg-sidebar-accent data-[active=true]:font-medium [&>svg]:size-4 [&>svg]:shrink-0")
}
