package shadcn

import (
	"slices"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/svg"
)

// Command is a Go port of shadcn/ui's Command (cmdk). cmdk filters its list
// client-side as you type; this port keeps that with one shared commandFilter
// script (the established inline-script pattern) that substring-matches each
// item's text on input, hides non-matching items and empty groups, and toggles
// [CommandEmpty].
//
// One item is selected at a time, marked data-selected="true" and
// aria-selected="true" as cmdk marks it: the first visible item after the
// page loads and after every filter run, then the one moved to with
// ArrowDown/ArrowUp/Home/End (without wrapping, as cmdk's default) or under
// the pointer. Enter clicks the selected item, or the first link or button
// inside it. Focus stays in the input, which points at the selected item
// with aria-activedescendant, so the script gives items and the list an id
// when they have none. An item with data-disabled="true" is never selected.
// Like cmdk, the keys are left alone while an IME composes text (isComposing,
// or keyCode 229 as Safari reports it) and when an earlier handler called
// preventDefault.
//
// Items inserted or removed after the page loads, such as server-side search
// results that htmx swaps into the [CommandList], are picked up by a
// MutationObserver: when the selected item is gone, the first item is
// selected again, and a replaced list gets its id and aria-controls back.
//
// cmdk's fuzzy ranking (sorting the matches by score) is not reproduced: the
// matches keep their order.
//
// The script is rendered inside every Command. Its first run defines
// window.commandFilter (the oninput handler of [CommandInput]); every run then
// wires up the Command it is in through document.currentScript, so a Command
// inserted by an htmx swap works too.
//
// For the ⌘K palette, place a Command inside a [Dialog] (DialogContent) — the
// CommandDialog recipe — see the gallery.
const commandScript = /*js*/ `if(!window.commandFilter){(function(){` +
	`var n=0;` +
	`var items=function(root){return Array.prototype.filter.call(root.querySelectorAll('[data-slot=command-item]'),function(it){return !it.hidden&&it.getAttribute('data-disabled')!=='true';});};` +
	`var sel=function(root,it,scroll){root.querySelectorAll('[data-slot=command-item][data-selected]').forEach(function(o){o.removeAttribute('data-selected');o.removeAttribute('aria-selected');});var input=root.querySelector('[data-slot=command-input]');if(!it){if(input)input.removeAttribute('aria-activedescendant');return;}if(!it.id)it.id='mx-command-item-'+(++n);it.setAttribute('data-selected','true');it.setAttribute('aria-selected','true');if(input)input.setAttribute('aria-activedescendant',it.id);if(scroll)it.scrollIntoView({block:'nearest'});};` +
	`window.commandFilter=function(input){var root=input.closest('[data-slot=command]');if(!root)return;var q=input.value.trim().toLowerCase();var any=false;root.querySelectorAll('[data-slot=command-item]').forEach(function(it){var m=it.textContent.toLowerCase().indexOf(q)>=0;it.hidden=!m;if(m)any=true;});root.querySelectorAll('[data-slot=command-group]').forEach(function(g){g.hidden=g.querySelectorAll('[data-slot=command-item]:not([hidden])').length===0;});var e=root.querySelector('[data-slot=command-empty]');if(e)e.hidden=any;sel(root,items(root)[0]||null,true);};` +
	`window.commandFilter.init=function(root){if(!root||root.mxCommand)return;root.mxCommand=true;` +
	`var link=function(){var input=root.querySelector('[data-slot=command-input]'),list=root.querySelector('[data-slot=command-list]');if(input&&list){if(!list.id)list.id='mx-command-list-'+(++n);input.setAttribute('aria-controls',list.id);}};link();` +
	`root.addEventListener('keydown',function(e){if(e.defaultPrevented||e.isComposing||e.keyCode===229)return;var all=items(root);if(!all.length)return;var cur=root.querySelector('[data-slot=command-item][data-selected=true]');var i=all.indexOf(cur),next;` +
	`switch(e.key){case 'ArrowDown':next=all[Math.min(i+1,all.length-1)];break;case 'ArrowUp':next=all[Math.max(i-1,0)];break;case 'Home':next=all[0];break;case 'End':next=all[all.length-1];break;` +
	`case 'Enter':if(cur){e.preventDefault();(cur.querySelector('a[href],button')||cur).click();}return;default:return;}e.preventDefault();sel(root,next,true);});` +
	`root.addEventListener('pointermove',function(e){var it=e.target.closest&&e.target.closest('[data-slot=command-item]');if(it&&root.contains(it)&&!it.hidden&&it.getAttribute('data-selected')!=='true'&&it.getAttribute('data-disabled')!=='true')sel(root,it,false);});` +
	`if(window.MutationObserver)new MutationObserver(function(){link();var cur=root.querySelector('[data-slot=command-item][data-selected=true]');if(!cur||cur.hidden||cur.getAttribute('data-disabled')==='true')sel(root,items(root)[0]||null,false);}).observe(root,{childList:true,subtree:true});` +
	`sel(root,items(root)[0]||null,false);};` +
	`})();}` +
	`window.commandFilter.init(document.currentScript&&document.currentScript.closest('[data-slot=command]'));`

// Command is the filterable command palette container.
func Command(attribsChildren ...any) *mx.Element {
	e := html.Div(append(attribsChildren, html.ScriptJS(commandScript))...)
	return finish(e, "command",
		"bg-popover text-popover-foreground flex h-full w-full flex-col overflow-hidden rounded-md")
}

// CommandInput renders the search box (with a leading search icon) that filters
// the list on input. Pass html.Placeholder("…") and other input
// attributes; they land on the <input>. It is a combobox controlling the
// [CommandList], as cmdk's input is: role="combobox", aria-autocomplete="list"
// and aria-expanded="true", with aria-controls and aria-activedescendant set
// by the script (see [Command]).
func CommandInput(attribs ...mx.Attrib) *mx.Element {
	input := html.VoidElement("input", append([]mx.Attrib{
		html.Type("text"),
		html.Role("combobox"),
		html.Attrib("aria-autocomplete", "list"),
		html.Attrib("aria-expanded", "true"),
		html.Attrib("autocomplete", "off"),
		html.Attrib("autocorrect", "off"),
		html.Attrib("spellcheck", "false"),
		html.OnInput("commandFilter(this)"),
	}, attribs...)...)
	input = finish(input, "command-input",
		"placeholder:text-muted-foreground flex h-10 w-full rounded-md bg-transparent py-3 text-sm outline-hidden disabled:cursor-not-allowed disabled:opacity-50")
	return finish(html.Div(
		icon("search", "size-4 shrink-0 opacity-50", svg.Circle(svg.CX(11), svg.CY(11), svg.R(8)), svg.Path(svg.D("m21 21-4.3-4.3"))),
		input,
	), "command-input-wrapper", "flex h-9 items-center gap-2 border-b px-3")
}

// CommandList is the scrollable results region, a listbox of the items.
func CommandList(attribsChildren ...any) *mx.Element {
	return finish(html.Div(append([]any{html.Role("listbox")}, attribsChildren...)...), "command-list",
		"max-h-[300px] scroll-py-1 overflow-x-hidden overflow-y-auto")
}

// CommandEmpty is shown by the filter when nothing matches. It starts hidden.
// Like cmdk's, it has role="presentation": a listbox may only contain
// options and groups, so its text is read without making it an entry.
func CommandEmpty(attribsChildren ...any) *mx.Element {
	e := html.Div(append([]any{html.Role("presentation")}, attribsChildren...)...)
	if e.AttribIndex("hidden") < 0 {
		e.Attribs = append(e.Attribs, html.Hidden)
	}
	return finish(e, "command-empty", "py-6 text-center text-sm")
}

// CommandGroup is a labeled set of items. heading may be "" for no label.
// The group is named by its heading through aria-label, so the visible
// heading is hidden from assistive technology, where it would otherwise be
// read as a stray entry of the listbox.
func CommandGroup(heading string, attribsChildren ...any) *mx.Element {
	attribs := []any{html.Role("group")}
	if heading != "" {
		attribs = append(attribs, html.Attrib("aria-label", heading))
	}
	e := html.Div(append(attribs, attribsChildren...)...)
	if heading != "" {
		// The heading is the group's first child, ahead of the caller's.
		e.Children = slices.Insert(e.Children, 0, mx.Component(html.Div(
			html.Class("text-muted-foreground px-2 py-1.5 text-xs font-medium"),
			html.Attrib("aria-hidden", "true"),
			heading,
		)))
	}
	return finish(e, "command-group", "text-foreground overflow-hidden p-1")
}

// CommandItem is one selectable row, an option of the [CommandList]. Its text
// is what the filter matches. The selected item is highlighted through
// upstream's data-[selected=true] classes, which the script sets (see
// [Command]); pass html.Attrib("data-disabled", "true") to dim an item and
// keep it from being selected.
func CommandItem(attribsChildren ...any) *mx.Element {
	return finish(html.Div(append([]any{html.Role("option")}, attribsChildren...)...), "command-item",
		"relative flex cursor-default items-center gap-2 rounded-sm px-2 py-1.5 text-sm outline-hidden select-none data-[disabled=true]:pointer-events-none data-[disabled=true]:opacity-50 data-[selected=true]:bg-accent data-[selected=true]:text-accent-foreground [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4 [&_svg:not([class*='text-'])]:text-muted-foreground")
}

// CommandSeparator divides groups. It has role="separator", as cmdk's does.
func CommandSeparator(attribsChildren ...any) *mx.Element {
	return finish(html.Div(append([]any{html.Role("separator")}, attribsChildren...)...), "command-separator", "bg-border -mx-1 h-px")
}

// CommandShortcut renders a right-aligned keyboard hint inside a [CommandItem].
func CommandShortcut(attribsChildren ...any) *mx.Element {
	return finish(html.Span(attribsChildren...), "command-shortcut",
		"text-muted-foreground ml-auto text-xs tracking-widest")
}
