//go:generate go -C ../tools tool go-enum ../shadcn/$GOFILE

package shadcn

import (
	"fmt"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn/cva"
)

// Item is a Go port of shadcn/ui's Item, a generic media-object / list-item
// layout block. Class strings are reconstructed from the post-July-2026
// registry (item.tsx skeleton + the style-vega.css .cn-item-* rules) — see
// "Upstream restructure (July 2026)" in TODOS.md. Upstream's render prop
// (React useRender) is Go composition here: Item always renders a <div>;
// wrap it in an <a> or restyle via caller classes for other shapes.

// ItemVariant selects an [Item]'s visual style.
type ItemVariant string //#enum

const (
	// ItemDefault is the default item style with a transparent border.
	ItemDefault ItemVariant = "default"
	// ItemOutline styles the item with a visible border.
	ItemOutline ItemVariant = "outline"
	// ItemMuted styles the item with a muted background.
	ItemMuted ItemVariant = "muted"
)

// Valid indicates if i is any of the valid values for ItemVariant
func (i ItemVariant) Valid() bool {
	switch i {
	case
		ItemDefault,
		ItemOutline,
		ItemMuted:
		return true
	}
	return false
}

// Validate returns an error if i is none of the valid values for ItemVariant
func (i ItemVariant) Validate() error {
	if !i.Valid() {
		return fmt.Errorf("invalid value %#v for type shadcn.ItemVariant", i)
	}
	return nil
}

// Enums returns all valid values for ItemVariant
func (ItemVariant) Enums() []ItemVariant {
	return []ItemVariant{
		ItemDefault,
		ItemOutline,
		ItemMuted,
	}
}

// EnumStrings returns all valid values for ItemVariant as strings
func (ItemVariant) EnumStrings() []string {
	return []string{
		"default",
		"outline",
		"muted",
	}
}

// String implements the fmt.Stringer interface for ItemVariant
func (i ItemVariant) String() string {
	return string(i)
}

// ItemSize selects an [Item]'s padding and gap scale.
type ItemSize string //#enum

const (
	// ItemSizeDefault is the default item size.
	ItemSizeDefault ItemSize = "default"
	// ItemSizeSM is the small item size.
	ItemSizeSM ItemSize = "sm"
	// ItemSizeXS is the extra-small item size.
	ItemSizeXS ItemSize = "xs"
)

// Valid indicates if i is any of the valid values for ItemSize
func (i ItemSize) Valid() bool {
	switch i {
	case
		ItemSizeDefault,
		ItemSizeSM,
		ItemSizeXS:
		return true
	}
	return false
}

// Validate returns an error if i is none of the valid values for ItemSize
func (i ItemSize) Validate() error {
	if !i.Valid() {
		return fmt.Errorf("invalid value %#v for type shadcn.ItemSize", i)
	}
	return nil
}

// Enums returns all valid values for ItemSize
func (ItemSize) Enums() []ItemSize {
	return []ItemSize{
		ItemSizeDefault,
		ItemSizeSM,
		ItemSizeXS,
	}
}

// EnumStrings returns all valid values for ItemSize as strings
func (ItemSize) EnumStrings() []string {
	return []string{
		"default",
		"sm",
		"xs",
	}
}

// String implements the fmt.Stringer interface for ItemSize
func (i ItemSize) String() string {
	return string(i)
}

// itemVariants resolves an item's base + variant + size classes, declared the
// same way shadcn/ui's item.tsx declares them with cva.
var itemVariants = cva.New(cva.Config{
	Base: "group/item flex w-full flex-wrap items-center transition-colors duration-100 outline-none focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 [a]:transition-colors [a]:hover:bg-muted rounded-md border text-sm",
	Variants: map[string]map[string]string{
		"variant": {
			"default": "border-transparent",
			"outline": "border-border",
			"muted":   "bg-muted/50 border-transparent",
		},
		"size": {
			"default": "gap-3.5 px-4 py-3.5",
			"sm":      "gap-2.5 px-3 py-2.5",
			"xs":      "gap-2 px-2.5 py-2 in-data-[slot=dropdown-menu-content]:p-0",
		},
	},
	DefaultVariants: map[string]string{"variant": "default", "size": "default"},
})

// normItemVariant maps an empty or unknown variant to the default.
func normItemVariant(v ItemVariant) string {
	switch v {
	case ItemOutline, ItemMuted:
		return string(v)
	default:
		return string(ItemDefault)
	}
}

// normItemSize maps an empty or unknown size to the default.
func normItemSize(s ItemSize) string {
	switch s {
	case ItemSizeSM, ItemSizeXS:
		return string(s)
	default:
		return string(ItemSizeDefault)
	}
}

// Item renders one media-object row. variant and size may be "" for the
// defaults. Compose it with [ItemMedia], [ItemContent] (holding [ItemTitle]
// and [ItemDescription]), [ItemActions], [ItemHeader] and [ItemFooter].
func Item(variant ItemVariant, size ItemSize, attribsChildren ...any) *mx.Element {
	v, s := normItemVariant(variant), normItemSize(size)
	e := html.Div(attribsChildren...)
	e.Attribs = append(e.Attribs,
		html.DataAttr("variant", v),
		html.DataAttr("size", s),
	)
	return finish(e, "item",
		Cn(itemVariants(map[string]string{"variant": v, "size": s})))
}

// ItemGroup renders a list of [Item]s as a <div role="list">. A caller-
// supplied role is left untouched.
func ItemGroup(attribsChildren ...any) *mx.Element {
	e := html.Div(attribsChildren...)
	if e.AttribIndex("role") < 0 {
		e.Attribs = append(e.Attribs, html.Role("list"))
	}
	// Upstream's bare has-data-[size=sm]: selectors would match any
	// descendant with a data-size — including this port's Button, which
	// (unlike upstream's) always emits one — so they are scoped to items.
	return finish(e, "item-group",
		"group/item-group flex w-full flex-col gap-4 has-[[data-slot=item][data-size=sm]]:gap-2.5 has-[[data-slot=item][data-size=xs]]:gap-2")
}

// ItemSeparator renders a horizontal [Separator] between items in an
// [ItemGroup].
func ItemSeparator(attribsChildren ...any) *mx.Element {
	return finish(Separator(SeparatorHorizontal, attribsChildren...),
		"item-separator", "my-2")
}

// ItemMediaVariant selects how an [ItemMedia] frames its content.
type ItemMediaVariant string //#enum

const (
	// ItemMediaDefault renders the media without a frame.
	ItemMediaDefault ItemMediaVariant = "default"
	// ItemMediaIcon sizes the media for an icon.
	ItemMediaIcon ItemMediaVariant = "icon"
	// ItemMediaImage frames an image in a rounded square that follows the item size.
	ItemMediaImage ItemMediaVariant = "image"
)

// Valid indicates if i is any of the valid values for ItemMediaVariant
func (i ItemMediaVariant) Valid() bool {
	switch i {
	case
		ItemMediaDefault,
		ItemMediaIcon,
		ItemMediaImage:
		return true
	}
	return false
}

// Validate returns an error if i is none of the valid values for ItemMediaVariant
func (i ItemMediaVariant) Validate() error {
	if !i.Valid() {
		return fmt.Errorf("invalid value %#v for type shadcn.ItemMediaVariant", i)
	}
	return nil
}

// Enums returns all valid values for ItemMediaVariant
func (ItemMediaVariant) Enums() []ItemMediaVariant {
	return []ItemMediaVariant{
		ItemMediaDefault,
		ItemMediaIcon,
		ItemMediaImage,
	}
}

// EnumStrings returns all valid values for ItemMediaVariant as strings
func (ItemMediaVariant) EnumStrings() []string {
	return []string{
		"default",
		"icon",
		"image",
	}
}

// String implements the fmt.Stringer interface for ItemMediaVariant
func (i ItemMediaVariant) String() string {
	return string(i)
}

// itemMediaVariants resolves an ItemMedia's base + variant classes, declared
// the same way shadcn/ui's item.tsx declares them with cva.
var itemMediaVariants = cva.New(cva.Config{
	Base: "flex shrink-0 items-center justify-center [&_svg]:pointer-events-none gap-2 group-has-data-[slot=item-description]/item:translate-y-0.5 group-has-data-[slot=item-description]/item:self-start",
	Variants: map[string]map[string]string{
		"variant": {
			"default": "bg-transparent",
			"icon":    "[&_svg:not([class*='size-'])]:size-4",
			"image":   "size-10 overflow-hidden rounded-sm group-data-[size=sm]/item:size-8 group-data-[size=xs]/item:size-6 [&_img]:size-full [&_img]:object-cover",
		},
	},
	DefaultVariants: map[string]string{"variant": "default"},
})

// ItemMedia renders an [Item]'s leading icon, image or avatar slot. variant
// may be "" for the default.
func ItemMedia(variant ItemMediaVariant, attribsChildren ...any) *mx.Element {
	if variant != ItemMediaIcon && variant != ItemMediaImage {
		variant = ItemMediaDefault
	}
	e := html.Div(attribsChildren...)
	e.Attribs = append(e.Attribs, html.DataAttr("variant", string(variant)))
	return finish(e, "item-media",
		Cn(itemMediaVariants(map[string]string{"variant": string(variant)})))
}

// ItemContent holds an [Item]'s [ItemTitle] and [ItemDescription], filling
// the space between media and actions.
func ItemContent(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "item-content",
		"flex flex-1 flex-col [&+[data-slot=item-content]]:flex-none gap-1 group-data-[size=xs]/item:gap-0")
}

// ItemTitle renders an [Item]'s title. Like upstream it is a <div>, not a
// heading element.
func ItemTitle(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "item-title",
		"line-clamp-1 flex w-fit items-center gap-2 text-sm leading-snug font-medium underline-offset-4")
}

// ItemDescription renders an [Item]'s description text as a <p>.
func ItemDescription(attribsChildren ...any) *mx.Element {
	return finish(html.P(attribsChildren...), "item-description",
		"line-clamp-2 font-normal [&>a]:underline [&>a]:underline-offset-4 [&>a:hover]:text-primary text-muted-foreground text-left text-sm leading-normal group-data-[size=xs]/item:text-xs")
}

// ItemActions holds an [Item]'s trailing actions (buttons, switches, …).
func ItemActions(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "item-actions",
		"flex items-center gap-2")
}

// ItemHeader renders a full-width row above an [Item]'s media and content.
func ItemHeader(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "item-header",
		"flex basis-full items-center justify-between gap-2")
}

// ItemFooter renders a full-width row below an [Item]'s media and content.
func ItemFooter(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "item-footer",
		"flex basis-full items-center justify-between gap-2")
}
