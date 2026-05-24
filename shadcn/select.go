package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// selectClasses is shadcn/ui's SelectTrigger class set adapted to a native
// <select>. The native control gets [appearance:base-select] so Chromium
// 130+ (and Safari Tech Preview) honor the full Tailwind look — including a
// stylable picker. Browsers without base-select (Firefox as of mid-2026)
// fall back to the native chrome — visually different but fully functional,
// a real form control, full a11y.
//
// shadcn's SelectTrigger/SelectValue/SelectContent/SelectItem/SelectScroll*
// abstractions are Radix wrappers; they collapse to nothing with a native
// <select> + <option>, and are not ported. Same precedent as
// AlertDialogOverlay and ScrollBar — see shadcn/README.md.
const selectClasses = "[appearance:base-select] border-input flex h-9 w-fit items-center justify-between gap-2 rounded-md border bg-transparent px-3 py-2 text-sm shadow-xs outline-none focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] disabled:cursor-not-allowed disabled:opacity-50 aria-invalid:border-destructive"

// selectOptionClasses styles individual <option>s. These rules only apply
// in browsers that honor appearance: base-select on the parent <select>;
// other browsers ignore them and show native option chrome.
const selectOptionClasses = "rounded-sm px-2 py-1.5 text-sm cursor-default outline-hidden checked:bg-accent checked:text-accent-foreground hover:bg-accent hover:text-accent-foreground disabled:opacity-50 disabled:pointer-events-none"

// Select renders a shadcn/ui select as a styled native <select>. Pass
// html.Name, html.Value, html.Required, html.Disabled, html.Multiple the
// normal way. Children are [SelectOption] and [SelectGroup].
//
// Form submission works as for any native <select>; no hidden field or
// script is needed.
func Select(attribsChildren ...any) *mx.Element {
	return finish(html.Select(attribsChildren...), "select", selectClasses)
}

// SelectGroup renders an option group inside a [Select] as a native
// <optgroup>. label is the group's heading (a typed parameter because the
// HTML attribute is load-bearing and the group is invalid without it).
func SelectGroup(label string, attribsChildren ...any) *mx.Element {
	e := html.OptGroup(attribsChildren...)
	if e.AttribIndex("label") < 0 {
		e.Attribs = append(e.Attribs, html.LabelAttr(label))
	}
	return finish(e, "select-group", "")
}

// SelectOption renders one item inside a [Select] or [SelectGroup] as a
// native <option>. value is the option's submitted value (typed because the
// HTML attribute is load-bearing); the option's display text is taken from
// the children. Pass html.Selected to mark the initial selection,
// html.Disabled to disable.
func SelectOption(value string, attribsChildren ...any) *mx.Element {
	e := html.Option(attribsChildren...)
	if e.AttribIndex("value") < 0 {
		e.Attribs = append(e.Attribs, html.Value(value))
	}
	return finish(e, "select-option", selectOptionClasses)
}
