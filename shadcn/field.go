//go:generate go -C ../tools tool go-enum ../shadcn/$GOFILE

package shadcn

import (
	"fmt"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn/cva"
)

// Field is a Go port of shadcn/ui's Field, the form-field layout system that
// replaced the Form parts in the post-July-2026 registry. Class strings are
// reconstructed from field.tsx + the style-vega.css .cn-field-* rules — see
// "Upstream restructure (July 2026)" in TODOS.md.
//
// Base UI selector rewrites in this port: `has-data-checked:` →
// `has-checked:` (our checkboxes/radios are native inputs), the
// `[role=checkbox],[role=radio]` alignment tweak → the data-slot attributes
// our controls emit, and `group-has-data-horizontal/field:` →
// `group-data-[orientation=horizontal]/field:` (our [Field] root carries
// data-orientation).
//
// Field carries no state of its own; the error/disabled looks key off
// author-set attributes: pass html.DataAttr("invalid", "true") to [Field]
// (turns the field's text destructive) and html.DataAttr("disabled", "true")
// (dims label and title) when rendering an errored or disabled field, and
// render a [FieldError] with the message. There is no react-hook-form
// equivalent — validation is a server concern.

// FieldOrientation selects how a [Field] lays out its label and control.
type FieldOrientation string //#enum

const (
	// FieldVertical stacks label above control (the default).
	FieldVertical FieldOrientation = "vertical"
	// FieldHorizontal puts label and control side by side.
	FieldHorizontal FieldOrientation = "horizontal"
	// FieldResponsive stacks on small containers and goes horizontal from
	// the @md container width of the enclosing [FieldGroup].
	FieldResponsive FieldOrientation = "responsive"
)

// Valid indicates if f is any of the valid values for FieldOrientation
func (f FieldOrientation) Valid() bool {
	switch f {
	case
		FieldVertical,
		FieldHorizontal,
		FieldResponsive:
		return true
	}
	return false
}

// Validate returns an error if f is none of the valid values for FieldOrientation
func (f FieldOrientation) Validate() error {
	if !f.Valid() {
		return fmt.Errorf("invalid value %#v for type shadcn.FieldOrientation", f)
	}
	return nil
}

// Enums returns all valid values for FieldOrientation
func (FieldOrientation) Enums() []FieldOrientation {
	return []FieldOrientation{
		FieldVertical,
		FieldHorizontal,
		FieldResponsive,
	}
}

// EnumStrings returns all valid values for FieldOrientation as strings
func (FieldOrientation) EnumStrings() []string {
	return []string{
		"vertical",
		"horizontal",
		"responsive",
	}
}

// String implements the fmt.Stringer interface for FieldOrientation
func (f FieldOrientation) String() string {
	return string(f)
}

// fieldVariants resolves a field's base + orientation classes, declared the
// same way shadcn/ui's field.tsx declares them with cva.
var fieldVariants = cva.New(cva.Config{
	Base: "group/field flex w-full data-[invalid=true]:text-destructive gap-3",
	Variants: map[string]map[string]string{
		"orientation": {
			"vertical": "flex-col *:w-full [&>.sr-only]:w-auto",
			"horizontal": "flex-row items-center has-[>[data-slot=field-content]]:items-start *:data-[slot=field-label]:flex-auto " +
				"has-[>[data-slot=field-content]]:[&>[data-slot=checkbox],&>[data-slot=radio-group-item]]:mt-px",
			"responsive": "flex-col *:w-full [&>.sr-only]:w-auto @md/field-group:flex-row @md/field-group:items-center @md/field-group:*:w-auto " +
				"@md/field-group:has-[>[data-slot=field-content]]:items-start @md/field-group:*:data-[slot=field-label]:flex-auto " +
				"@md/field-group:has-[>[data-slot=field-content]]:[&>[data-slot=checkbox],&>[data-slot=radio-group-item]]:mt-px",
		},
	},
	DefaultVariants: map[string]string{"orientation": "vertical"},
})

// normFieldOrientation maps an empty or unknown orientation to the default
// (vertical).
func normFieldOrientation(o FieldOrientation) FieldOrientation {
	switch o {
	case FieldHorizontal, FieldResponsive:
		return o
	default:
		return FieldVertical
	}
}

// Field groups one form field's label, control, description and error as a
// <div role="group">. orientation may be "" for the default (vertical). A
// caller-supplied role is left untouched. Compose it with [FieldLabel],
// [FieldContent], [FieldDescription] and [FieldError]; group several Fields
// with [FieldGroup] or [FieldSet].
func Field(orientation FieldOrientation, attribsChildren ...any) *mx.Element {
	o := normFieldOrientation(orientation)
	e := html.Div(attribsChildren...)
	if e.AttribIndex("role") < 0 {
		e.Attribs = append(e.Attribs, html.Role("group"))
	}
	e.Attribs = append(e.Attribs, html.DataAttr("orientation", string(o)))
	return finish(e, "field",
		Cn(fieldVariants(map[string]string{"orientation": string(o)})))
}

// FieldGroup stacks several [Field]s (or nested FieldGroups). It is also the
// container the responsive orientation queries (@container/field-group).
func FieldGroup(attribsChildren ...any) *mx.Element {
	// Upstream's data-[slot=checkbox-group]:gap-3 self-selector is dropped:
	// finish always sets this element's data-slot to "field-group", so it
	// could never match (same policy as the cn-font-heading drop in empty.go).
	return finish(html.Div(attribsChildren...), "field-group",
		"group/field-group @container/field-group flex w-full flex-col gap-7 *:data-[slot=field-group]:gap-4")
}

// FieldSet groups related [Field]s as a native <fieldset>, titled by a
// [FieldLegend].
func FieldSet(attribsChildren ...any) *mx.Element {
	return finish(html.FieldSet(attribsChildren...), "field-set",
		"flex flex-col gap-6 has-[>[data-slot=checkbox-group]]:gap-3 has-[>[data-slot=radio-group]]:gap-3")
}

// FieldLegendVariant selects a [FieldLegend]'s type scale.
type FieldLegendVariant string //#enum

const (
	// FieldLegendDefault is the default legend scale (the "legend" variant).
	FieldLegendDefault FieldLegendVariant = "legend"
	// FieldLegendLabel renders the legend at label size.
	FieldLegendLabel FieldLegendVariant = "label"
)

// Valid indicates if f is any of the valid values for FieldLegendVariant
func (f FieldLegendVariant) Valid() bool {
	switch f {
	case
		FieldLegendDefault,
		FieldLegendLabel:
		return true
	}
	return false
}

// Validate returns an error if f is none of the valid values for FieldLegendVariant
func (f FieldLegendVariant) Validate() error {
	if !f.Valid() {
		return fmt.Errorf("invalid value %#v for type shadcn.FieldLegendVariant", f)
	}
	return nil
}

// Enums returns all valid values for FieldLegendVariant
func (FieldLegendVariant) Enums() []FieldLegendVariant {
	return []FieldLegendVariant{
		FieldLegendDefault,
		FieldLegendLabel,
	}
}

// EnumStrings returns all valid values for FieldLegendVariant as strings
func (FieldLegendVariant) EnumStrings() []string {
	return []string{
		"legend",
		"label",
	}
}

// String implements the fmt.Stringer interface for FieldLegendVariant
func (f FieldLegendVariant) String() string {
	return string(f)
}

// FieldLegend renders a [FieldSet]'s <legend>. variant may be "" for the
// default scale; [FieldLegendLabel] matches the size of a [FieldLabel].
func FieldLegend(variant FieldLegendVariant, attribsChildren ...any) *mx.Element {
	if variant != FieldLegendLabel {
		variant = FieldLegendDefault
	}
	e := html.Element("legend", attribsChildren...)
	e.Attribs = append(e.Attribs, html.DataAttr("variant", string(variant)))
	return finish(e, "field-legend",
		"mb-3 font-medium data-[variant=label]:text-sm data-[variant=legend]:text-base")
}

// FieldContent stacks a [FieldTitle]/[FieldDescription] next to a control
// (e.g. beside a Checkbox in a horizontal [Field]), filling the free space.
func FieldContent(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "field-content",
		"group/field-content flex flex-1 flex-col leading-snug gap-1")
}

// fieldLabelClasses layers the Field-specific label classes over the shared
// labelClasses, merged once at package init (both operands are constants).
var fieldLabelClasses = Cn(labelClasses,
	"group/field-label peer/field-label flex w-fit has-[>[data-slot=field]]:w-full has-[>[data-slot=field]]:flex-col has-checked:bg-primary/5 has-checked:border-primary/30 dark:has-checked:border-primary/20 dark:has-checked:bg-primary/10 gap-2 leading-snug group-data-[disabled=true]/field:opacity-50 has-[>[data-slot=field]]:rounded-md has-[>[data-slot=field]]:border *:data-[slot=field]:p-3")

// FieldLabel is a [Label] for a [Field]'s control. It can also wrap a whole
// [Field] to make a selectable card (the has-[>[data-slot=field]]: and
// has-checked: classes) — e.g. a bordered radio choice.
func FieldLabel(attribsChildren ...any) *mx.Element {
	return finish(html.Label(attribsChildren...), "field-label", fieldLabelClasses)
}

// FieldLabelFor is a [FieldLabel] bound to the control with the given id, as
// a shortcut for FieldLabel(html.For(id), attribsChildren...).
func FieldLabelFor(id string, attribsChildren ...any) *mx.Element {
	return FieldLabel(append([]any{html.For(id)}, attribsChildren...)...)
}

// FieldTitle renders a label-look title where no <label> element fits. Like
// upstream it emits data-slot="field-label" so the [Field] layout classes
// treat it as the label.
func FieldTitle(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "field-label",
		"flex w-fit items-center gap-2 leading-snug text-sm font-medium group-data-[disabled=true]/field:opacity-50")
}

// FieldDescription renders helper text for a [Field].
func FieldDescription(attribsChildren ...any) *mx.Element {
	return finish(html.P(attribsChildren...), "field-description",
		"text-muted-foreground text-left text-sm leading-normal font-normal group-data-[orientation=horizontal]/field:text-balance last:mt-0 nth-last-2:-mt-1 [&>a]:underline [&>a]:underline-offset-4 [&>a:hover]:text-primary [[data-variant=legend]+&]:-mt-1.5")
}

// FieldSeparator renders a horizontal rule between [Field]s in a
// [FieldGroup]. Children (e.g. "Or continue with") are rendered as a
// centered inline chip over the rule.
func FieldSeparator(attribsChildren ...any) *mx.Element {
	e := html.Div(attribsChildren...)
	content := e.Children
	e.Children = []mx.Component{
		Separator(SeparatorHorizontal, html.Class("absolute inset-0 top-1/2")),
	}
	e.Attribs = append(e.Attribs, html.DataAttr("content", boolStr(len(content) > 0)))
	if len(content) > 0 {
		span := html.SpanClass("relative mx-auto block w-fit bg-background text-muted-foreground px-2")
		span.Attribs = append(span.Attribs, mx.NewAttrib("data-slot", "field-separator-content"))
		span.Children = content
		e.Children = append(e.Children, span)
	}
	return finish(e, "field-separator",
		"relative -my-2 h-5 text-sm group-data-[variant=outline]/field-group:-mb-2")
}

// FieldError renders a [Field]'s validation message as a <div role="alert">
// in the destructive color. Render it only when there is an error to show
// (upstream's errors-array dedup prop is React-form plumbing; in Go the
// caller passes the message as children). A caller-supplied role is left
// untouched.
func FieldError(attribsChildren ...any) *mx.Element {
	e := html.Div(attribsChildren...)
	if e.AttribIndex("role") < 0 {
		e.Attribs = append(e.Attribs, html.Role("alert"))
	}
	return finish(e, "field-error", "font-normal text-destructive text-sm")
}
