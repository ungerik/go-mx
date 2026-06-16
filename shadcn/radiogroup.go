package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// RadioGroup renders a shadcn/ui radio group as a <div role="radiogroup">. The
// group's name is a leading parameter so it can be validated and so the doc
// comment can flag the contract: a [RadioGroupItem] in this group must be
// passed the same name. The link is by attribute, not DOM nesting (the
// AlertDialog precedent), so any well-known go-mx composition pattern works.
//
// name must contain only letters, digits, '-' and '_'; an empty or invalid name
// panics, since it is interpolated into each item's name attribute and
// addressed by browsers as the radios' exclusive-selection group.
func RadioGroup(name string, attribsChildren ...any) *mx.Element {
	if err := validateID(name); err != nil {
		return mx.NewErrElement(err)
	}
	e := html.Div(attribsChildren...)
	if e.AttribIndex("role") < 0 {
		e.Attribs = append(e.Attribs, html.Role("radiogroup"))
	}
	if e.AttribIndex("data-name") < 0 {
		e.Attribs = append(e.Attribs, html.DataAttr("name", name))
	}
	return finish(e, "radio-group", "grid gap-3")
}

// radioGroupItemClasses is shadcn/ui's RadioGroupItem class set adapted to a
// styled void <input type="radio">. A native radio is a void element and cannot
// hold the Radix child Indicator/CircleIcon, so the dot is drawn with a before:
// pseudo-element on the appearance-none input — a deliberate divergence,
// mirroring the Switch thumb. data-[state=checked]:* is rewritten to checked:*
// throughout; the dot stays in the DOM and scales 0→100 on :checked.
const radioGroupItemClasses = "peer aspect-square size-4 shrink-0 appearance-none rounded-full border border-input bg-transparent shadow-xs text-primary outline-none transition-[color,box-shadow] inline-flex items-center justify-center focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] disabled:cursor-not-allowed disabled:opacity-50 aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive dark:bg-input/30 checked:border-primary" +
	" before:content-[''] before:block before:size-2 before:rounded-full before:bg-primary before:scale-0 checked:before:scale-100 before:transition-transform"

// RadioGroupItem renders one option in a [RadioGroup] as a styled void
// <input type="radio">. Both name and value are typed leading parameters
// because both are load-bearing (browsers group radios by name and submit the
// chosen value). name must match the [RadioGroup]'s name.
//
// Pass html.Checked to start selected, html.Disabled to disable, html.ID to
// link a [Label]. Caller-supplied type, name or value is left untouched.
// Children are not valid on a void element and are dropped.
func RadioGroupItem(name, value string, attribsChildren ...any) *mx.Element {
	if err := validateID(name); err != nil {
		return mx.NewErrElement(err)
	}
	e := html.Element("input", attribsChildren...)
	e.Children = nil // <input> is a void element
	if e.AttribIndex("type") < 0 {
		e.Attribs = append(e.Attribs, html.Type("radio"))
	}
	if e.AttribIndex("name") < 0 {
		e.Attribs = append(e.Attribs, html.Name(name))
	}
	if e.AttribIndex("value") < 0 {
		e.Attribs = append(e.Attribs, html.Value(value))
	}
	return finish(e, "radio-group-item", radioGroupItemClasses)
}
