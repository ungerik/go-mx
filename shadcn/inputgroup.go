//go:generate go -C ../tools tool go-enum ../shadcn/$GOFILE

package shadcn

import (
	"fmt"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn/cva"
)

// InputGroup is a Go port of shadcn/ui's InputGroup, an input with attached
// addons (prefix/suffix icons, text, buttons). Class strings are
// reconstructed from the post-July-2026 registry (input-group.tsx skeleton +
// the style-vega.css .cn-input-group-* rules) — see "Upstream restructure
// (July 2026)" in TODOS.md.
//
// The group carries the border and focus ring; the control inside must be an
// [InputGroupInput] or [InputGroupTextarea] (borderless variants of [Input] /
// [Textarea] with data-slot="input-group-control", which the group's
// has-[…]: state classes key off).

// InputGroup renders the group wrapper as a <div role="group">. A caller-
// supplied role is left untouched.
func InputGroup(attribsChildren ...any) *mx.Element {
	e := html.Div(attribsChildren...)
	if e.AttribIndex("role") < 0 {
		e.Attribs = append(e.Attribs, html.Role("group"))
	}
	return finish(e, "input-group",
		"group/input-group relative flex w-full min-w-0 items-center outline-none has-[>textarea]:h-auto border-input dark:bg-input/30 has-[[data-slot=input-group-control]:focus-visible]:border-ring has-[[data-slot=input-group-control]:focus-visible]:ring-ring/50 has-[[data-slot][aria-invalid=true]]:ring-destructive/20 has-[[data-slot][aria-invalid=true]]:border-destructive dark:has-[[data-slot][aria-invalid=true]]:ring-destructive/40 h-9 rounded-md border shadow-xs transition-[color,box-shadow] in-data-[slot=combobox-content]:focus-within:border-inherit in-data-[slot=combobox-content]:focus-within:ring-0 has-[[data-slot=input-group-control]:focus-visible]:ring-3 has-[[data-slot][aria-invalid=true]]:ring-3 has-[>[data-align=block-end]]:h-auto has-[>[data-align=block-end]]:flex-col has-[>[data-align=block-start]]:h-auto has-[>[data-align=block-start]]:flex-col has-[>[data-align=block-end]]:[&>input]:pt-3 has-[>[data-align=block-start]]:[&>input]:pb-3 has-[>[data-align=inline-end]]:[&>input]:pr-1.5 has-[>[data-align=inline-start]]:[&>input]:pl-1.5")
}

// InputGroupAddonAlign selects where an [InputGroupAddon] sits in the group:
// inline at the start/end of the control's row, or as a block row above/
// below it.
type InputGroupAddonAlign string //#enum

const (
	// InputGroupAddonInlineStart places the addon before the control (the default).
	InputGroupAddonInlineStart InputGroupAddonAlign = "inline-start"
	// InputGroupAddonInlineEnd places the addon after the control.
	InputGroupAddonInlineEnd InputGroupAddonAlign = "inline-end"
	// InputGroupAddonBlockStart places the addon as a full-width row above the control.
	InputGroupAddonBlockStart InputGroupAddonAlign = "block-start"
	// InputGroupAddonBlockEnd places the addon as a full-width row below the control.
	InputGroupAddonBlockEnd InputGroupAddonAlign = "block-end"
)

// Valid indicates if i is any of the valid values for InputGroupAddonAlign
func (i InputGroupAddonAlign) Valid() bool {
	switch i {
	case
		InputGroupAddonInlineStart,
		InputGroupAddonInlineEnd,
		InputGroupAddonBlockStart,
		InputGroupAddonBlockEnd:
		return true
	}
	return false
}

// Validate returns an error if i is none of the valid values for InputGroupAddonAlign
func (i InputGroupAddonAlign) Validate() error {
	if !i.Valid() {
		return fmt.Errorf("invalid value %#v for type shadcn.InputGroupAddonAlign", i)
	}
	return nil
}

// Enums returns all valid values for InputGroupAddonAlign
func (InputGroupAddonAlign) Enums() []InputGroupAddonAlign {
	return []InputGroupAddonAlign{
		InputGroupAddonInlineStart,
		InputGroupAddonInlineEnd,
		InputGroupAddonBlockStart,
		InputGroupAddonBlockEnd,
	}
}

// EnumStrings returns all valid values for InputGroupAddonAlign as strings
func (InputGroupAddonAlign) EnumStrings() []string {
	return []string{
		"inline-start",
		"inline-end",
		"block-start",
		"block-end",
	}
}

// String implements the fmt.Stringer interface for InputGroupAddonAlign
func (i InputGroupAddonAlign) String() string {
	return string(i)
}

// inputGroupAddonVariants resolves an addon's base + align classes, declared
// the same way shadcn/ui's input-group.tsx declares them with cva.
var inputGroupAddonVariants = cva.New(cva.Config{
	Base: "flex cursor-text items-center justify-center select-none text-muted-foreground h-auto gap-2 py-1.5 text-sm font-medium group-data-[disabled=true]/input-group:opacity-50 [&>kbd]:rounded-[calc(var(--radius)-5px)] [&>svg:not([class*='size-'])]:size-4",
	Variants: map[string]map[string]string{
		"align": {
			"inline-start": "order-first pl-2 has-[>button]:-ml-1 has-[>kbd]:ml-[-0.15rem]",
			"inline-end":   "order-last pr-2 has-[>button]:-mr-1 has-[>kbd]:mr-[-0.15rem]",
			"block-start":  "order-first w-full justify-start px-2.5 pt-2 group-has-[>input]/input-group:pt-2 [.border-b]:pb-2",
			"block-end":    "order-last w-full justify-start px-2.5 pb-2 group-has-[>input]/input-group:pb-2 [.border-t]:pt-2",
		},
	},
	DefaultVariants: map[string]string{"align": "inline-start"},
})

// inputGroupAddonFocus is the addon's default onclick: clicking the addon
// focuses the group's control, unless the click landed on a button inside
// the addon (the port of input-group.tsx's onClick handler; upstream queries
// only 'input', which misses [InputGroupTextarea] groups).
const inputGroupAddonFocus = "if(!event.target.closest('button'))this.parentElement.querySelector('input,textarea')?.focus()"

// InputGroupAddon renders an addon (icon, [InputGroupText], [Kbd],
// [InputGroupButton], …) inside an [InputGroup]. align may be "" for the
// default (inline-start). Clicking the addon focuses the group's control;
// pass your own html.OnClick to replace that.
func InputGroupAddon(align InputGroupAddonAlign, attribsChildren ...any) *mx.Element {
	switch align {
	case InputGroupAddonInlineEnd, InputGroupAddonBlockStart, InputGroupAddonBlockEnd:
	default:
		align = InputGroupAddonInlineStart
	}
	e := html.Div(attribsChildren...)
	if e.AttribIndex("role") < 0 {
		e.Attribs = append(e.Attribs, html.Role("group"))
	}
	if e.AttribIndex("onclick") < 0 {
		e.Attribs = append(e.Attribs, html.OnClick(inputGroupAddonFocus))
	}
	e.Attribs = append(e.Attribs, html.DataAttr("align", string(align)))
	return finish(e, "input-group-addon",
		Cn(inputGroupAddonVariants(map[string]string{"align": string(align)})))
}

// InputGroupButtonSize selects an [InputGroupButton]'s size — its own scale,
// smaller than [ButtonSize], sized to sit inside the group's h-9 frame.
type InputGroupButtonSize string //#enum

const (
	// InputGroupButtonXS is the extra-small input-group button size (the default).
	InputGroupButtonXS InputGroupButtonSize = "xs"
	// InputGroupButtonSM is the small input-group button size.
	InputGroupButtonSM InputGroupButtonSize = "sm"
	// InputGroupButtonIconXS is the extra-small square size for an icon-only button.
	InputGroupButtonIconXS InputGroupButtonSize = "icon-xs"
	// InputGroupButtonIconSM is the small square size for an icon-only button.
	InputGroupButtonIconSM InputGroupButtonSize = "icon-sm"
)

// Valid indicates if i is any of the valid values for InputGroupButtonSize
func (i InputGroupButtonSize) Valid() bool {
	switch i {
	case
		InputGroupButtonXS,
		InputGroupButtonSM,
		InputGroupButtonIconXS,
		InputGroupButtonIconSM:
		return true
	}
	return false
}

// Validate returns an error if i is none of the valid values for InputGroupButtonSize
func (i InputGroupButtonSize) Validate() error {
	if !i.Valid() {
		return fmt.Errorf("invalid value %#v for type shadcn.InputGroupButtonSize", i)
	}
	return nil
}

// Enums returns all valid values for InputGroupButtonSize
func (InputGroupButtonSize) Enums() []InputGroupButtonSize {
	return []InputGroupButtonSize{
		InputGroupButtonXS,
		InputGroupButtonSM,
		InputGroupButtonIconXS,
		InputGroupButtonIconSM,
	}
}

// EnumStrings returns all valid values for InputGroupButtonSize as strings
func (InputGroupButtonSize) EnumStrings() []string {
	return []string{
		"xs",
		"sm",
		"icon-xs",
		"icon-sm",
	}
}

// String implements the fmt.Stringer interface for InputGroupButtonSize
func (i InputGroupButtonSize) String() string {
	return string(i)
}

// inputGroupButtonVariants resolves the classes an [InputGroupButton] layers
// over [ButtonClasses], declared the same way shadcn/ui's input-group.tsx
// declares them with cva. The "sm" size adds nothing beyond the shared
// classes (verbatim from style-vega.css, which defines no
// cn-input-group-button-size-sm rule).
var inputGroupButtonVariants = cva.New(cva.Config{
	Base: "flex items-center shadow-none gap-2 text-sm",
	Variants: map[string]map[string]string{
		"size": {
			"xs":      "h-6 gap-1 rounded-[calc(var(--radius)-5px)] px-1.5 [&>svg:not([class*='size-'])]:size-3.5",
			"sm":      "",
			"icon-xs": "size-6 rounded-[calc(var(--radius)-5px)] p-0 has-[>svg]:p-0",
			"icon-sm": "size-8 p-0 has-[>svg]:p-0",
		},
	},
	DefaultVariants: map[string]string{"size": "xs"},
})

// normInputGroupButtonSize maps an empty or unknown size to the default (xs).
func normInputGroupButtonSize(s InputGroupButtonSize) string {
	switch s {
	case InputGroupButtonSM, InputGroupButtonIconXS, InputGroupButtonIconSM:
		return string(s)
	default:
		return string(InputGroupButtonXS)
	}
}

// InputGroupButton renders a button inside an [InputGroupAddon]. variant may
// be "" for the default, which is [ButtonGhost] (not [ButtonDefault] —
// matching upstream, a subdued button suits the inside of an input). size
// may be "" for the default (xs). Like [Button] it defaults to
// type="button"; pass html.Type to override.
func InputGroupButton(variant ButtonVariant, size InputGroupButtonSize, attribsChildren ...any) *mx.Element {
	if variant == "" {
		variant = ButtonGhost
	}
	s := normInputGroupButtonSize(size)
	e := html.Button(attribsChildren...)
	if e.AttribIndex("type") < 0 {
		e.Attribs = append(e.Attribs, html.Type("button"))
	}
	e.Attribs = append(e.Attribs,
		html.DataAttr("variant", string(variant)),
		html.DataAttr("size", s),
	)
	return finish(e, "button", Cn(
		ButtonClasses(variant, SizeDefault),
		inputGroupButtonVariants(map[string]string{"size": s}),
	))
}

// InputGroupText renders static text inside an [InputGroupAddon] as a
// <span>. Upstream emits no data-slot here; this port emits
// data-slot="input-group-text" for consistency (every component in this
// package tags its slot).
func InputGroupText(attribsChildren ...any) *mx.Element {
	return finish(html.Span(attribsChildren...), "input-group-text",
		"flex items-center [&_svg]:pointer-events-none text-muted-foreground gap-2 text-sm [&_svg:not([class*='size-'])]:size-4")
}

// inputGroupInputClasses / inputGroupTextareaClasses layer the borderless
// overrides over the plain control classes, merged once at package init
// (both operands are constants).
var (
	inputGroupInputClasses    = Cn(inputClasses, "flex-1 rounded-none border-0 bg-transparent shadow-none ring-0 focus-visible:ring-0 aria-invalid:ring-0 dark:bg-transparent")
	inputGroupTextareaClasses = Cn(textareaClasses, "flex-1 rounded-none border-0 bg-transparent py-2 shadow-none ring-0 focus-visible:ring-0 aria-invalid:ring-0 dark:bg-transparent resize-none")
)

// InputGroupInput renders the group's text input: an [Input] restyled
// borderless (the [InputGroup] wrapper carries the border and focus ring)
// with data-slot="input-group-control".
func InputGroupInput(attribs ...mx.Attrib) *mx.Element {
	return finish(html.VoidElement("input", attribs...), "input-group-control",
		inputGroupInputClasses)
}

// InputGroupTextarea renders the group's textarea: a [Textarea] restyled
// borderless (the [InputGroup] wrapper carries the border and focus ring)
// with data-slot="input-group-control".
func InputGroupTextarea(attribsChildren ...any) *mx.Element {
	return finish(html.TextArea(attribsChildren...), "input-group-control",
		inputGroupTextareaClasses)
}
