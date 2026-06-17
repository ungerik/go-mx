package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// tabsSelectScript is the once-emitted client function that drives tab
// switching. It scopes to a single Tabs instance by [data-tabs="{id}"] so
// multiple Tabs can coexist on a page without colliding. Triggers opt out by
// passing any hx-* attribute (see [TabsTrigger]) — then htmx drives the swap
// instead.
const tabsSelectScript = /*js*/ `if(!window.tabsSelect){window.tabsSelect=function(id,value){var r=document.querySelector('[data-tabs="'+id+'"]');if(!r)return;r.querySelectorAll('[data-slot="tabs-trigger"]').forEach(function(t){var on=t.getAttribute('data-tabs-value')===value;t.setAttribute('aria-selected',on);t.setAttribute('tabindex',on?'0':'-1');});r.querySelectorAll('[data-slot="tabs-content"]').forEach(function(p){if(p.getAttribute('data-tabs-value')===value)p.removeAttribute('hidden');else p.setAttribute('hidden','');});};}`

// Tabs renders a shadcn/ui tabs container as a <div>. id is a stable identifier
// that scopes the shared tabsSelect script to this Tabs instance, so several
// Tabs can coexist without coordination; it is validated with the shared
// validateID rules.
//
// shadcn/ui's Tabs is Radix-driven. This port keeps the same accessible
// markup — role="tablist", role="tab", role="tabpanel", aria-selected,
// aria-controls, aria-labelledby — and drives the switching with one short
// inline <script> emitted once per Tabs (guarded with if(!window.tabsSelect)).
// Triggers opt into htmx by passing an hx-* attribute; see [TabsTrigger].
func Tabs(id string, attribsChildren ...any) *mx.Element {
	if err := validateID(id); err != nil {
		return mx.NewErrElement(err)
	}
	e := html.Div(append(attribsChildren, html.ScriptJS(tabsSelectScript))...)
	if e.AttribIndex("data-tabs") < 0 {
		e.Attribs = append(e.Attribs, html.DataAttr("tabs", id))
	}
	return finish(e, "tabs", "flex flex-col gap-2")
}

// TabsList renders the row of triggers as a <div role="tablist">.
func TabsList(attribsChildren ...any) *mx.Element {
	e := html.Div(attribsChildren...)
	if e.AttribIndex("role") < 0 {
		e.Attribs = append(e.Attribs, html.Role("tablist"))
	}
	return finish(e, "tabs-list",
		"bg-muted text-muted-foreground inline-flex h-9 w-fit items-center justify-center rounded-lg p-[3px]")
}

// tabsTriggerClasses is shadcn/ui's TabsTrigger class string with the Radix
// data-[state=active]:* selectors rewritten to aria-selected:* (Tailwind's
// aria-selected variant expands to [aria-selected="true"], which is what the
// trigger carries when active).
const tabsTriggerClasses = "aria-selected:bg-background dark:aria-selected:text-foreground focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:outline-ring dark:aria-selected:border-input dark:aria-selected:bg-input/30 text-foreground dark:text-muted-foreground inline-flex h-[calc(100%-1px)] flex-1 items-center justify-center gap-1.5 rounded-md border border-transparent px-2 py-1 text-sm font-medium whitespace-nowrap transition-[color,box-shadow] focus-visible:ring-[3px] focus-visible:outline-1 disabled:pointer-events-none disabled:opacity-50 aria-selected:shadow-sm [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4"

// TabsTrigger renders one tab control as a <button role="tab">. tabsID must
// match the [Tabs] id; value is this tab's identifier (used to address the
// matching [TabsContent]); active marks the initially-selected tab.
//
// Defaults (overridable): type="button", role="tab", aria-selected, aria-controls,
// id, data-tabs-value, and tabindex="-1" for inactive triggers (roving
// tabindex). When no caller onclick is supplied and no htmx attribute is
// present, TabsTrigger adds a default onclick that calls the shared
// tabsSelect(tabsID, value) — pass any hx.* attribute (e.g. hx.Get(...)) to
// drive the panel swap server-side instead.
func TabsTrigger(tabsID, value string, active bool, attribsChildren ...any) *mx.Element {
	if err := validateID(tabsID); err != nil {
		return mx.NewErrElement(err)
	}
	e := html.Button(attribsChildren...)
	if e.AttribIndex("type") < 0 {
		e.Attribs = append(e.Attribs, html.Type("button"))
	}
	if e.AttribIndex("role") < 0 {
		e.Attribs = append(e.Attribs, html.Role("tab"))
	}
	if e.AttribIndex("aria-selected") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("aria-selected", boolStr(active)))
	}
	if e.AttribIndex("aria-controls") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("aria-controls", tabsID+"-panel-"+value))
	}
	if e.AttribIndex("id") < 0 {
		e.Attribs = append(e.Attribs, html.ID(tabsID+"-tab-"+value))
	}
	if e.AttribIndex("data-tabs-value") < 0 {
		e.Attribs = append(e.Attribs, html.DataAttr("tabs-value", value))
	}
	if !active && e.AttribIndex("tabindex") < 0 {
		e.Attribs = append(e.Attribs, html.TabIndex(-1))
	}
	if e.AttribIndex("onclick") < 0 && !hasHX(e) {
		e.Attribs = append(e.Attribs, html.OnClick("tabsSelect('"+tabsID+"','"+value+"')"))
	}
	return finish(e, "tabs-trigger", tabsTriggerClasses)
}

// TabsContent renders one tab panel as a <div role="tabpanel">. tabsID must
// match the [Tabs] id; value must match a [TabsTrigger]'s value; active makes
// this panel initially visible. Inactive panels are emitted with the boolean
// hidden attribute; the tabsSelect script flips it on switch.
func TabsContent(tabsID, value string, active bool, attribsChildren ...any) *mx.Element {
	if err := validateID(tabsID); err != nil {
		return mx.NewErrElement(err)
	}
	e := html.Div(attribsChildren...)
	if e.AttribIndex("role") < 0 {
		e.Attribs = append(e.Attribs, html.Role("tabpanel"))
	}
	if e.AttribIndex("id") < 0 {
		e.Attribs = append(e.Attribs, html.ID(tabsID+"-panel-"+value))
	}
	if e.AttribIndex("aria-labelledby") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("aria-labelledby", tabsID+"-tab-"+value))
	}
	if e.AttribIndex("data-tabs-value") < 0 {
		e.Attribs = append(e.Attribs, html.DataAttr("tabs-value", value))
	}
	if e.AttribIndex("tabindex") < 0 {
		e.Attribs = append(e.Attribs, html.TabIndex(0))
	}
	if !active && e.AttribIndex("hidden") < 0 {
		e.Attribs = append(e.Attribs, html.Hidden)
	}
	return finish(e, "tabs-content", "flex-1 outline-none")
}
