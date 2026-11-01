//go:generate go -C ../tools tool go-enum ../shadcn/$GOFILE

package shadcn

import (
	"fmt"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn/cva"
)

// Marker is a Go port of shadcn/ui's Marker, a labelled row for status lines
// and dividers in a conversation ("Worked for 42s", "Rhea joined the chat").
// Class strings are reconstructed from the post-July-2026 registry
// (marker.tsx skeleton + the style-vega.css .cn-marker* rules) — see
// "Upstream restructure (July 2026)" in TODOS.md.

// MarkerVariant selects a [Marker]'s visual style.
type MarkerVariant string //#enum

const (
	// MarkerDefault is a plain row of muted text.
	MarkerDefault MarkerVariant = "default"
	// MarkerSeparator draws a horizontal rule on both sides of the content.
	MarkerSeparator MarkerVariant = "separator"
	// MarkerBorder draws a rule below the row.
	MarkerBorder MarkerVariant = "border"
)

// Valid indicates if m is any of the valid values for MarkerVariant
func (m MarkerVariant) Valid() bool {
	switch m {
	case
		MarkerDefault,
		MarkerSeparator,
		MarkerBorder:
		return true
	}
	return false
}

// Validate returns an error if m is none of the valid values for MarkerVariant
func (m MarkerVariant) Validate() error {
	if !m.Valid() {
		return fmt.Errorf("invalid value %#v for type shadcn.MarkerVariant", m)
	}
	return nil
}

// ValidOr returns m if it is any of the valid values for MarkerVariant, else defaultVal
func (m MarkerVariant) ValidOr(defaultVal MarkerVariant) MarkerVariant {
	if m.Valid() {
		return m
	}
	return defaultVal
}

// Enums returns all valid values for MarkerVariant
func (MarkerVariant) Enums() []MarkerVariant {
	return []MarkerVariant{
		MarkerDefault,
		MarkerSeparator,
		MarkerBorder,
	}
}

// EnumStrings returns all valid values for MarkerVariant as strings
func (MarkerVariant) EnumStrings() []string {
	return []string{
		"default",
		"separator",
		"border",
	}
}

// String implements the fmt.Stringer interface for MarkerVariant
func (m MarkerVariant) String() string {
	return string(m)
}

// markerVariants resolves a marker's base + variant classes, declared the
// same way shadcn/ui's marker.tsx declares them with cva.
var markerVariants = cva.New(cva.Config{
	Base: "group/marker relative flex w-full items-center gap-2 text-sm text-muted-foreground [a]:hover:text-foreground [a]:underline-offset-3 [a]:underline [&_svg:not([class*='size-'])]:size-4 min-h-4 text-left",
	Variants: map[string]map[string]string{
		"variant": {
			"default":   "",
			"separator": "before:h-px before:min-w-0 before:flex-1 before:bg-border after:h-px after:min-w-0 after:flex-1 after:bg-border before:mr-1 after:ml-1",
			"border":    "border-b border-border pb-2",
		},
	},
	DefaultVariants: map[string]string{"variant": "default"},
})

// Marker renders one marker row. variant may be "" for the default; an
// unknown value also resolves to the default. Compose it with [MarkerIcon]
// and [MarkerContent]; pass html.Role("status") for a row that reports
// progress.
//
// Upstream's render prop (React useRender) is replaced by the attributes: the
// row is a <div>, an <a> when given an href, and a <button> when given a type
// (pass html.Type("button")).
func Marker(variant MarkerVariant, attribsChildren ...any) *mx.Element {
	v := string(variant.ValidOr(MarkerDefault))
	e := html.Div(attribsChildren...)
	e.Name = renderTag(e, "div")
	e.Attribs = append(e.Attribs, html.DataAttr("variant", v))
	return finish(e, "marker", Cn(markerVariants(map[string]string{"variant": v})))
}

// MarkerIcon holds a [Marker]'s icon or [Spinner]. It is hidden from
// assistive technology, like upstream.
func MarkerIcon(attribsChildren ...any) *mx.Element {
	e := html.Span(attribsChildren...)
	if e.AttribIndex("aria-hidden") < 0 {
		e.Attribs = append(e.Attribs, html.Attrib("aria-hidden", "true"))
	}
	return finish(e, "marker-icon",
		"shrink-0 size-4 [&_svg:not([class*='size-'])]:size-4")
}

// MarkerContent holds a [Marker]'s text. In a [MarkerSeparator] row it
// shrinks to its text between the two rules.
func MarkerContent(attribsChildren ...any) *mx.Element {
	return finish(html.Span(attribsChildren...), "marker-content",
		"min-w-0 wrap-break-word group-data-[variant=separator]/marker:flex-none group-data-[variant=separator]/marker:text-center *:[a]:hover:text-foreground *:[a]:underline *:[a]:underline-offset-3")
}
