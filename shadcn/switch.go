package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// switchClasses is shadcn/ui's Switch class set (root + thumb) collapsed onto a
// single styled void <input>. A native <input> cannot hold a child Thumb
// element, so the thumb is drawn with a before: pseudo-element on the
// appearance-none input — a deliberate divergence from shadcn's two-element
// Root/Thumb structure. Pseudo-elements render on checkbox/radio inputs once
// appearance is removed and act as flex items inside inline-flex items-center.
//
// data-state rewrites: data-[state=checked]:bg-primary → checked:bg-primary;
// data-[state=unchecked]:bg-input is the unprefixed base; the thumb's
// data-[state=checked]:translate-x-[calc(100%-2px)] becomes
// checked:before:translate-x-[calc(100%-2px)].
const switchClasses = "peer inline-flex h-[1.15rem] w-8 shrink-0 appearance-none items-center rounded-full border border-transparent bg-input shadow-xs transition-all outline-none focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] disabled:cursor-not-allowed disabled:opacity-50 checked:bg-primary dark:bg-input/80" +
	" before:content-[''] before:pointer-events-none before:block before:size-4 before:rounded-full before:bg-background before:ring-0 before:transition-transform before:translate-x-0 checked:before:translate-x-[calc(100%-2px)] dark:before:bg-foreground dark:checked:before:bg-primary-foreground"

// Switch renders a shadcn/ui switch as a styled void
// <input type="checkbox" role="switch">. The track is the input itself; the
// thumb is drawn with a CSS before: pseudo-element.
//
// Pass html.Checked to start in the on state, html.Name/html.Value for form
// submission, html.Disabled to disable. Caller-supplied type, role or onchange
// is left untouched. Children are not valid on a void element and are dropped.
func Switch(attribsChildren ...any) *mx.Element {
	e := html.Element("input", attribsChildren...)
	e.Children = nil // <input> is a void element
	if e.AttribIndex("type") < 0 {
		e.Attribs = append(e.Attribs, html.Type("checkbox"))
	}
	if e.AttribIndex("role") < 0 {
		e.Attribs = append(e.Attribs, html.Role("switch"))
	}
	return finish(e, "switch", switchClasses)
}
