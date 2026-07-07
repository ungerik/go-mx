//go:generate go -C ../tools tool go-enum ../shadcn/$GOFILE

package shadcn

import (
	"fmt"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn/cva"
)

// Empty is a Go port of shadcn/ui's Empty, an empty-state layout block.
// Class strings are reconstructed from the post-July-2026 registry
// (empty.tsx skeleton + the style-vega.css .cn-empty-* rules) — see
// "Upstream restructure (July 2026)" in TODOS.md. Upstream's cn-font-heading
// utility on EmptyTitle expands to font-heading, a theme font variable this
// port's frozen new-york-v4 baseline does not define, and is dropped (the
// title inherits the body font).
//
// Compose: Empty(EmptyHeader(EmptyMedia(...), EmptyTitle(...),
// EmptyDescription(...)), EmptyContent(...)).
func Empty(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "empty",
		"flex w-full min-w-0 flex-1 flex-col items-center justify-center text-center text-balance gap-4 rounded-lg border-dashed p-12")
}

// EmptyHeader groups an [Empty]'s media, title and description.
func EmptyHeader(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "empty-header",
		"flex max-w-sm flex-col items-center gap-2")
}

// EmptyMediaVariant selects how an [EmptyMedia] frames its content.
type EmptyMediaVariant string //#enum

const (
	// EmptyMediaDefault renders the media without a frame.
	EmptyMediaDefault EmptyMediaVariant = "default"
	// EmptyMediaIcon frames an icon in a muted rounded square.
	EmptyMediaIcon EmptyMediaVariant = "icon"
)

// Valid indicates if e is any of the valid values for EmptyMediaVariant
func (e EmptyMediaVariant) Valid() bool {
	switch e {
	case
		EmptyMediaDefault,
		EmptyMediaIcon:
		return true
	}
	return false
}

// Validate returns an error if e is none of the valid values for EmptyMediaVariant
func (e EmptyMediaVariant) Validate() error {
	if !e.Valid() {
		return fmt.Errorf("invalid value %#v for type shadcn.EmptyMediaVariant", e)
	}
	return nil
}

// Enums returns all valid values for EmptyMediaVariant
func (EmptyMediaVariant) Enums() []EmptyMediaVariant {
	return []EmptyMediaVariant{
		EmptyMediaDefault,
		EmptyMediaIcon,
	}
}

// EnumStrings returns all valid values for EmptyMediaVariant as strings
func (EmptyMediaVariant) EnumStrings() []string {
	return []string{
		"default",
		"icon",
	}
}

// String implements the fmt.Stringer interface for EmptyMediaVariant
func (e EmptyMediaVariant) String() string {
	return string(e)
}

// emptyMediaVariants resolves an EmptyMedia's base + variant classes,
// declared the same way shadcn/ui's empty.tsx declares them with cva.
var emptyMediaVariants = cva.New(cva.Config{
	Base: "flex shrink-0 items-center justify-center [&_svg]:pointer-events-none [&_svg]:shrink-0 mb-2",
	Variants: map[string]map[string]string{
		"variant": {
			"default": "bg-transparent",
			"icon":    "bg-muted text-foreground flex size-10 shrink-0 items-center justify-center rounded-lg [&_svg:not([class*='size-'])]:size-6",
		},
	},
	DefaultVariants: map[string]string{"variant": "default"},
})

// EmptyMedia renders an [Empty]'s icon or image slot. variant may be "" for
// the default. The data-slot is "empty-icon", matching upstream's empty.tsx
// (which uses that name for the media part).
func EmptyMedia(variant EmptyMediaVariant, attribsChildren ...any) *mx.Element {
	if variant != EmptyMediaIcon {
		variant = EmptyMediaDefault
	}
	e := html.Div(attribsChildren...)
	e.Attribs = append(e.Attribs, html.DataAttr("variant", string(variant)))
	return finish(e, "empty-icon",
		Cn(emptyMediaVariants(map[string]string{"variant": string(variant)})))
}

// EmptyTitle renders an [Empty]'s title. Like upstream it is a <div>, not a
// heading element.
func EmptyTitle(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "empty-title",
		"text-lg font-medium tracking-tight")
}

// EmptyDescription renders an [Empty]'s description text. Like upstream it
// is a <div> (empty.tsx types it as a <p> but renders a <div>).
func EmptyDescription(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "empty-description",
		"text-muted-foreground [&>a]:underline [&>a]:underline-offset-4 [&>a:hover]:text-primary text-sm/relaxed")
}

// EmptyContent holds an [Empty]'s actions (buttons, links) below the header.
func EmptyContent(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "empty-content",
		"flex w-full max-w-sm min-w-0 flex-col items-center text-balance gap-4 text-sm")
}
