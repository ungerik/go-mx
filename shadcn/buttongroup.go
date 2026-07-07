//go:generate go -C ../tools tool go-enum ../shadcn/$GOFILE

package shadcn

import (
	"fmt"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn/cva"
)

// ButtonGroup is a Go port of shadcn/ui's ButtonGroup. Class strings are
// reconstructed from the post-July-2026 registry (button-group.tsx skeleton
// + the style-vega.css .cn-button-group-* rules) — see "Upstream restructure
// (July 2026)" in TODOS.md.

// ButtonGroupOrientation selects a [ButtonGroup]'s axis.
type ButtonGroupOrientation string //#enum

const (
	// ButtonGroupHorizontal lays the group out as a row (the default).
	ButtonGroupHorizontal ButtonGroupOrientation = "horizontal"
	// ButtonGroupVertical lays the group out as a column.
	ButtonGroupVertical ButtonGroupOrientation = "vertical"
)

// Valid indicates if b is any of the valid values for ButtonGroupOrientation
func (b ButtonGroupOrientation) Valid() bool {
	switch b {
	case
		ButtonGroupHorizontal,
		ButtonGroupVertical:
		return true
	}
	return false
}

// Validate returns an error if b is none of the valid values for ButtonGroupOrientation
func (b ButtonGroupOrientation) Validate() error {
	if !b.Valid() {
		return fmt.Errorf("invalid value %#v for type shadcn.ButtonGroupOrientation", b)
	}
	return nil
}

// Enums returns all valid values for ButtonGroupOrientation
func (ButtonGroupOrientation) Enums() []ButtonGroupOrientation {
	return []ButtonGroupOrientation{
		ButtonGroupHorizontal,
		ButtonGroupVertical,
	}
}

// EnumStrings returns all valid values for ButtonGroupOrientation as strings
func (ButtonGroupOrientation) EnumStrings() []string {
	return []string{
		"horizontal",
		"vertical",
	}
}

// String implements the fmt.Stringer interface for ButtonGroupOrientation
func (b ButtonGroupOrientation) String() string {
	return string(b)
}

// buttonGroupVariants resolves a button group's base + orientation classes,
// declared the same way shadcn/ui's button-group.tsx declares them with cva.
// Upstream's select selectors target Radix's [data-slot=select-trigger]; this
// port's Select is a native <select data-slot="select">, so the w-fit rule is
// rewritten to that slot. Upstream's extra last-select rounding rule is gated
// on Radix's hidden <select aria-hidden> and dropped here — the native select
// is the last direct child, already rounded by the orientation variant's
// :not(:has(~[data-slot])) rule.
var buttonGroupVariants = cva.New(cva.Config{
	Base: "flex w-fit items-stretch *:focus-visible:relative *:focus-visible:z-10 [&>[data-slot=select]:not([class*='w-'])]:w-fit [&>input]:flex-1 has-[>[data-slot=button-group]]:gap-2",
	Variants: map[string]map[string]string{
		"orientation": {
			"horizontal": "*:data-slot:rounded-r-none [&>[data-slot]~[data-slot]]:rounded-l-none [&>[data-slot]~[data-slot]]:border-l-0 [&>[data-slot]:not(:has(~[data-slot]))]:rounded-r-md!",
			"vertical":   "flex-col *:data-slot:rounded-b-none [&>[data-slot]~[data-slot]]:rounded-t-none [&>[data-slot]~[data-slot]]:border-t-0 [&>[data-slot]:not(:has(~[data-slot]))]:rounded-b-md!",
		},
	},
	DefaultVariants: map[string]string{"orientation": "horizontal"},
})

// normButtonGroupOrientation maps an empty or unknown orientation to the
// default (horizontal).
func normButtonGroupOrientation(o ButtonGroupOrientation) ButtonGroupOrientation {
	if o == ButtonGroupVertical {
		return ButtonGroupVertical
	}
	return ButtonGroupHorizontal
}

// ButtonGroupClasses returns the merged base + orientation class string. It
// is the equivalent of shadcn/ui's exported buttonGroupVariants and mirrors
// [ButtonClasses]. An empty or unknown orientation resolves to horizontal.
func ButtonGroupClasses(orientation ButtonGroupOrientation) string {
	return Cn(buttonGroupVariants(map[string]string{
		"orientation": string(normButtonGroupOrientation(orientation)),
	}))
}

// ButtonGroup renders a group of visually joined buttons (and compatible
// controls such as [Input] or [Select]) as a <div role="group">. orientation
// may be "" for the default (horizontal). The joining works on the direct
// children's data-slot attributes, which every component in this package
// emits. A caller-supplied role is left untouched.
func ButtonGroup(orientation ButtonGroupOrientation, attribsChildren ...any) *mx.Element {
	o := normButtonGroupOrientation(orientation)
	e := html.Div(attribsChildren...)
	if e.AttribIndex("role") < 0 {
		e.Attribs = append(e.Attribs, html.Role("group"))
	}
	e.Attribs = append(e.Attribs, html.DataAttr("orientation", string(o)))
	return finish(e, "button-group", ButtonGroupClasses(o))
}

// ButtonGroupText renders static text (or an icon) inline with the buttons
// of a [ButtonGroup].
func ButtonGroupText(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "button-group-text",
		"flex items-center [&_svg]:pointer-events-none bg-muted gap-2 rounded-md border px-2.5 text-sm font-medium shadow-xs [&_svg:not([class*='size-'])]:size-4")
}

// buttonGroupSeparatorClasses are upstream's ButtonGroupSeparator classes on
// top of the shared separator classes, with Base UI's bare data-horizontal:/
// data-vertical: selectors rewritten to the data-[orientation=…]: attribute
// this port's separators emit.
const buttonGroupSeparatorClasses = "relative self-stretch data-[orientation=horizontal]:mx-px data-[orientation=horizontal]:w-auto data-[orientation=vertical]:my-px data-[orientation=vertical]:h-auto bg-input"

// ButtonGroupSeparator renders a separator line between buttons in a
// [ButtonGroup]. orientation may be "" for the default, which is vertical —
// the opposite of [Separator]'s default, matching upstream (a horizontal
// group needs a vertical rule). It is a [Separator] re-tagged with its own
// data-slot; the override classes ride in as a leading caller class so they
// win the merge (bg-input over bg-border, h-auto over h-full) while any
// caller class still wins over them.
func ButtonGroupSeparator(orientation SeparatorOrientation, attribsChildren ...any) *mx.Element {
	o := SeparatorVertical
	if orientation == SeparatorHorizontal {
		o = SeparatorHorizontal
	}
	return finish(
		Separator(o, append([]any{html.Class(buttonGroupSeparatorClasses)}, attribsChildren...)...),
		"button-group-separator", "")
}
