//go:generate go -C ../tools tool go-enum ../shadcn/$GOFILE

package shadcn

import (
	"fmt"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// ToggleGroupType selects exclusive- vs independent-pressed behavior. shadcn's
// Radix Root exposes the same choice as a type="single"/"multiple" prop.
type ToggleGroupType string //#enum

const (
	// ToggleGroupSingle allows at most one item in the group to be pressed at a time (the default).
	ToggleGroupSingle ToggleGroupType = "single"
	// ToggleGroupMultiple allows any number of items in the group to be pressed independently.
	ToggleGroupMultiple ToggleGroupType = "multiple"
)

// Valid indicates if t is any of the valid values for ToggleGroupType
func (t ToggleGroupType) Valid() bool {
	switch t {
	case
		ToggleGroupSingle,
		ToggleGroupMultiple:
		return true
	}
	return false
}

// Validate returns an error if t is none of the valid values for ToggleGroupType
func (t ToggleGroupType) Validate() error {
	if !t.Valid() {
		return fmt.Errorf("invalid value %#v for type shadcn.ToggleGroupType", t)
	}
	return nil
}

// Enums returns all valid values for ToggleGroupType
func (ToggleGroupType) Enums() []ToggleGroupType {
	return []ToggleGroupType{
		ToggleGroupSingle,
		ToggleGroupMultiple,
	}
}

// EnumStrings returns all valid values for ToggleGroupType as strings
func (ToggleGroupType) EnumStrings() []string {
	return []string{
		"single",
		"multiple",
	}
}

// String implements the fmt.Stringer interface for ToggleGroupType
func (t ToggleGroupType) String() string {
	return string(t)
}

// toggleGroupClickScript is the once-emitted client function that drives
// item presses. It reads the parent group's data-type at click time so the
// same handler covers both modes: in single mode it clears every sibling's
// aria-pressed and sets this one; in multiple mode it flips this one in place.
// Items opt out by passing any hx-* attribute (see [ToggleGroupItem]) — then
// htmx drives the state change instead.
const toggleGroupClickScript = /*js*/ `if(!window.toggleGroupClick){window.toggleGroupClick=function(btn){var g=btn.closest('[data-toggle-group]');if(!g)return;if(g.getAttribute('data-type')==='single'){g.querySelectorAll('[data-slot="toggle-group-item"]').forEach(function(b){b.setAttribute('aria-pressed','false');});btn.setAttribute('aria-pressed','true');}else{btn.setAttribute('aria-pressed',btn.getAttribute('aria-pressed')==='true'?'false':'true');}};}`

// normToggleGroupType maps an empty or unknown type to single (matching Radix
// which has no implicit default — but a single-mode group is the more common
// shadcn use, so it makes a defensible fallback).
func normToggleGroupType(t ToggleGroupType) string {
	if t == ToggleGroupMultiple {
		return string(ToggleGroupMultiple)
	}
	return string(ToggleGroupSingle)
}

// ToggleGroup renders a shadcn/ui toggle group as a <div role="group">. id is
// a stable identifier (validated) used as the data-toggle-group attribute that
// the shared toggleGroupClick script targets via Element.closest. variant and
// size are reused by [ToggleGroupItem]; an empty value resolves to the default.
//
// The group emits its variant/size as data-variant / data-size so shadcn's
// data-[variant=outline]:* selectors (carried on both the root and the items)
// continue to match. The toggleGroupClick script is appended once per group
// instance, guarded with if(!window.toggleGroupClick).
func ToggleGroup(groupType ToggleGroupType, variant ToggleVariant, size ToggleSize, id string, attribsChildren ...any) *mx.Element {
	if err := validateID(id); err != nil {
		return mx.NewErrElement(err)
	}
	e := html.Div(append(attribsChildren, html.ScriptJS(toggleGroupClickScript))...)
	if e.AttribIndex("role") < 0 {
		e.Attribs = append(e.Attribs, html.Role("group"))
	}
	e.Attribs = append(e.Attribs,
		html.DataAttr("type", normToggleGroupType(groupType)),
		html.DataAttr("variant", normToggleVariant(variant)),
		html.DataAttr("size", normToggleSize(size)),
		html.DataAttr("toggle-group", id),
	)
	return finish(e, "toggle-group", "group/toggle-group flex w-fit items-center rounded-md data-[variant=outline]:shadow-xs")
}

// toggleGroupItemJoinClasses are the join classes shadcn applies on top of
// toggleVariants() for items packed against each other in a group.
const toggleGroupItemJoinClasses = "min-w-0 flex-1 shrink-0 rounded-none shadow-none first:rounded-l-md last:rounded-r-md focus:z-10 focus-visible:z-10 data-[variant=outline]:border-l-0 data-[variant=outline]:first:border-l"

// ToggleGroupItem renders one item in a [ToggleGroup] as a <button>, styled
// with [ToggleClasses] plus join classes. groupID must match the [ToggleGroup]
// id (validated; emitted as data-toggle-group on the item too, for symmetry
// and CSS targeting). value identifies this item.
//
// variant and size should match the [ToggleGroup]'s — shadcn's React context
// flows them down; in this port the caller passes them explicitly. Both may be
// "" for the default.
//
// Defaults (overridable): type="button", aria-pressed="false", data-variant,
// data-size, data-toggle-group-value, data-toggle-group. When no caller
// onclick and no htmx attribute are present, ToggleGroupItem adds a default
// onclick that calls the shared toggleGroupClick(this); pass any hx.*
// attribute (e.g. hx.Post(...)) to drive the press server-side instead.
func ToggleGroupItem(groupID, value string, variant ToggleVariant, size ToggleSize, attribsChildren ...any) *mx.Element {
	if err := validateID(groupID); err != nil {
		return mx.NewErrElement(err)
	}
	e := html.Button(attribsChildren...)
	if e.AttribIndex("type") < 0 {
		e.Attribs = append(e.Attribs, html.Type("button"))
	}
	if e.AttribIndex("aria-pressed") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("aria-pressed", "false"))
	}
	e.Attribs = append(e.Attribs,
		html.DataAttr("variant", normToggleVariant(variant)),
		html.DataAttr("size", normToggleSize(size)),
		html.DataAttr("toggle-group", groupID),
		html.DataAttr("toggle-group-value", value),
	)
	if e.AttribIndex("onclick") < 0 && !hasHX(e) {
		e.Attribs = append(e.Attribs, html.OnClick("toggleGroupClick(this)"))
	}
	return finish(e, "toggle-group-item", Cn(ToggleClasses(variant, size), toggleGroupItemJoinClasses))
}
