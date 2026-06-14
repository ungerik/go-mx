package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/svg"
)

// Command is a Go port of shadcn/ui's Command (cmdk). cmdk filters its list
// client-side as you type; this port keeps that with one shared commandFilter
// script (the established inline-script pattern) that substring-matches each
// item's text on input, hides non-matching items and empty groups, and toggles
// [CommandEmpty]. cmdk's fuzzy ranking and arrow-key navigation are not
// reproduced — substring filtering plus hover/click selection.
//
// For the ⌘K palette, place a Command inside a [Dialog] (DialogContent) — the
// CommandDialog recipe — see the gallery.
const commandScript = /*js*/ `if(!window.commandFilter){window.commandFilter=function(input){var root=input.closest('[data-slot=command]');if(!root)return;var q=input.value.trim().toLowerCase();var any=false;root.querySelectorAll('[data-slot=command-item]').forEach(function(it){var m=it.textContent.toLowerCase().indexOf(q)>=0;it.hidden=!m;if(m)any=true;});root.querySelectorAll('[data-slot=command-group]').forEach(function(g){g.hidden=g.querySelectorAll('[data-slot=command-item]:not([hidden])').length===0;});var e=root.querySelector('[data-slot=command-empty]');if(e)e.hidden=any;};}`

// Command is the filterable command palette container.
func Command(attribsChildren ...any) *mx.Element {
	e := html.Div(attribsChildren...)
	e.Children = append(e.Children, html.Script(mx.Raw(commandScript)))
	return finish(e, "command",
		"bg-popover text-popover-foreground flex h-full w-full flex-col overflow-hidden rounded-md")
}

// CommandInput renders the search box (with a leading search icon) that filters
// the list on input. Pass html.Attrib("placeholder", "…") and other input
// attributes; they land on the <input>.
func CommandInput(attribsChildren ...any) *mx.Element {
	input := html.Element("input", append([]any{html.Type("text"), html.OnInput("commandFilter(this)")}, attribsChildren...)...)
	input.Children = nil // <input> is a void element; avoid a stray </input>
	input = finish(input, "command-input",
		"placeholder:text-muted-foreground flex h-10 w-full rounded-md bg-transparent py-3 text-sm outline-hidden disabled:cursor-not-allowed disabled:opacity-50")
	return finish(html.Div(
		icon("search", "size-4 shrink-0 opacity-50", svg.Circle(svg.CX(11), svg.CY(11), svg.R(8)), svg.Path(svg.D("m21 21-4.3-4.3"))),
		input,
	), "command-input-wrapper", "flex h-9 items-center gap-2 border-b px-3")
}

// CommandList is the scrollable results region.
func CommandList(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "command-list",
		"max-h-[300px] scroll-py-1 overflow-x-hidden overflow-y-auto")
}

// CommandEmpty is shown by the filter when nothing matches. It starts hidden.
func CommandEmpty(attribsChildren ...any) *mx.Element {
	e := html.Div(attribsChildren...)
	if e.AttribIndex("hidden") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("hidden", ""))
	}
	return finish(e, "command-empty", "py-6 text-center text-sm")
}

// CommandGroup is a labeled set of items. heading may be "" for no label.
func CommandGroup(heading string, attribsChildren ...any) *mx.Element {
	children := make([]any, 0, len(attribsChildren)+1)
	if heading != "" {
		children = append(children, html.Div(
			html.Class("text-muted-foreground px-2 py-1.5 text-xs font-medium"), heading))
	}
	children = append(children, attribsChildren...)
	return finish(html.Div(children...), "command-group", "text-foreground overflow-hidden p-1")
}

// CommandItem is one selectable row. Its text is what the filter matches.
func CommandItem(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "command-item",
		"hover:bg-accent hover:text-accent-foreground relative flex cursor-default items-center gap-2 rounded-sm px-2 py-1.5 text-sm outline-hidden select-none [&_svg]:size-4 [&_svg]:shrink-0")
}

// CommandSeparator divides groups.
func CommandSeparator(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "command-separator", "bg-border -mx-1 h-px")
}

// CommandShortcut renders a right-aligned keyboard hint inside a [CommandItem].
func CommandShortcut(attribsChildren ...any) *mx.Element {
	return finish(html.Span(attribsChildren...), "command-shortcut",
		"text-muted-foreground ml-auto text-xs tracking-widest")
}
