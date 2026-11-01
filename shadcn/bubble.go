//go:generate go -C ../tools tool go-enum ../shadcn/$GOFILE

package shadcn

import (
	"fmt"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn/cva"
)

// Bubble is a Go port of shadcn/ui's Bubble, a chat bubble with seven color
// variants and optional reactions. Class strings are reconstructed from the
// post-July-2026 registry (bubble.tsx skeleton + the style-vega.css
// .cn-bubble* rules) — see "Upstream restructure (July 2026)" in TODOS.md.

// BubbleVariant selects a [Bubble]'s colors.
type BubbleVariant string //#enum

const (
	// BubbleDefault fills the bubble with the primary color.
	BubbleDefault BubbleVariant = "default"
	// BubbleSecondary fills the bubble with the secondary color.
	BubbleSecondary BubbleVariant = "secondary"
	// BubbleMuted fills the bubble with the muted color.
	BubbleMuted BubbleVariant = "muted"
	// BubbleTinted fills the bubble with a light tint of the primary color.
	BubbleTinted BubbleVariant = "tinted"
	// BubbleOutline draws a bordered bubble on the background color.
	BubbleOutline BubbleVariant = "outline"
	// BubbleGhost renders the content without a bubble, at full width: the
	// usual look of an assistant's reply.
	BubbleGhost BubbleVariant = "ghost"
	// BubbleDestructive tints the bubble with the destructive color, for a
	// failed message.
	BubbleDestructive BubbleVariant = "destructive"
)

// Valid indicates if b is any of the valid values for BubbleVariant
func (b BubbleVariant) Valid() bool {
	switch b {
	case
		BubbleDefault,
		BubbleSecondary,
		BubbleMuted,
		BubbleTinted,
		BubbleOutline,
		BubbleGhost,
		BubbleDestructive:
		return true
	}
	return false
}

// Validate returns an error if b is none of the valid values for BubbleVariant
func (b BubbleVariant) Validate() error {
	if !b.Valid() {
		return fmt.Errorf("invalid value %#v for type shadcn.BubbleVariant", b)
	}
	return nil
}

// ValidOr returns b if it is any of the valid values for BubbleVariant, else defaultVal
func (b BubbleVariant) ValidOr(defaultVal BubbleVariant) BubbleVariant {
	if b.Valid() {
		return b
	}
	return defaultVal
}

// Enums returns all valid values for BubbleVariant
func (BubbleVariant) Enums() []BubbleVariant {
	return []BubbleVariant{
		BubbleDefault,
		BubbleSecondary,
		BubbleMuted,
		BubbleTinted,
		BubbleOutline,
		BubbleGhost,
		BubbleDestructive,
	}
}

// EnumStrings returns all valid values for BubbleVariant as strings
func (BubbleVariant) EnumStrings() []string {
	return []string{
		"default",
		"secondary",
		"muted",
		"tinted",
		"outline",
		"ghost",
		"destructive",
	}
}

// String implements the fmt.Stringer interface for BubbleVariant
func (b BubbleVariant) String() string {
	return string(b)
}

// bubbleVariants resolves a bubble's base + variant classes, declared the
// same way shadcn/ui's bubble.tsx declares them with cva. The variants color
// the bubble's direct [BubbleContent] child; the hover colors apply when that
// child is an <a> or <button>.
var bubbleVariants = cva.New(cva.Config{
	Base: "group/bubble relative flex w-fit min-w-0 flex-col gap-1 data-[align=end]:self-end max-w-[80%] data-[variant=ghost]:max-w-full group-data-[align=end]/message:self-end",
	Variants: map[string]map[string]string{
		"variant": {
			"default":     "*:data-[slot=bubble-content]:bg-primary *:data-[slot=bubble-content]:text-primary-foreground [&>[data-slot=bubble-content]:is(button,a):hover]:bg-primary/80",
			"secondary":   "*:data-[slot=bubble-content]:bg-secondary *:data-[slot=bubble-content]:text-secondary-foreground [&>[data-slot=bubble-content]:is(button,a):hover]:bg-[color-mix(in_oklch,var(--secondary),var(--foreground)_5%)]",
			"muted":       "*:data-[slot=bubble-content]:bg-muted [&>[data-slot=bubble-content]:is(button,a):hover]:bg-[color-mix(in_oklch,var(--muted),var(--foreground)_5%)]",
			"tinted":      "*:data-[slot=bubble-content]:bg-[oklch(from_var(--primary)_0.93_calc(c*0.4)_h)] dark:*:data-[slot=bubble-content]:bg-[oklch(from_var(--primary)_0.3_calc(c*0.4)_h)] *:data-[slot=bubble-content]:text-foreground [&>[data-slot=bubble-content]:is(button,a):hover]:bg-[oklch(from_var(--primary)_0.88_calc(c*0.5)_h)] dark:[&>[data-slot=bubble-content]:is(button,a):hover]:bg-[oklch(from_var(--primary)_0.35_calc(c*0.5)_h)]",
			"outline":     "*:data-[slot=bubble-content]:bg-background *:data-[slot=bubble-content]:border-border [&>[data-slot=bubble-content]:is(button,a):hover]:bg-muted [&>[data-slot=bubble-content]:is(button,a):hover]:text-foreground dark:[&>[data-slot=bubble-content]:is(button,a):hover]:bg-input/30",
			"ghost":       "*:data-[slot=bubble-content]:rounded-none *:data-[slot=bubble-content]:bg-transparent *:data-[slot=bubble-content]:p-0 [&>[data-slot=bubble-content]:is(button,a):hover]:bg-muted [&>[data-slot=bubble-content]:is(button,a):hover]:text-foreground dark:[&>[data-slot=bubble-content]:is(button,a):hover]:bg-muted/50 border-none",
			"destructive": "*:data-[slot=bubble-content]:bg-destructive/10 dark:*:data-[slot=bubble-content]:bg-destructive/20 *:data-[slot=bubble-content]:text-destructive [&>[data-slot=bubble-content]:is(button,a):hover]:bg-destructive/20 dark:[&>[data-slot=bubble-content]:is(button,a):hover]:bg-destructive/30",
		},
	},
	DefaultVariants: map[string]string{"variant": "default"},
})

// BubbleGroup stacks consecutive [Bubble]s of one sender.
func BubbleGroup(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "bubble-group",
		"flex min-w-0 flex-col gap-2")
}

// Bubble renders one chat bubble holding a [BubbleContent] and optionally
// [BubbleReactions]. variant may be "" for [BubbleDefault] and align "" for
// [ChatAlignStart]; unknown values also resolve to those. Inside a
// [ChatAlignEnd] [Message] the bubble aligns to the end on its own. A bubble
// takes up to 80% of the width, a [BubbleGhost] all of it.
func Bubble(variant BubbleVariant, align ChatAlign, attribsChildren ...any) *mx.Element {
	v, a := string(variant.ValidOr(BubbleDefault)), string(align.ValidOr(ChatAlignStart))
	e := html.Div(attribsChildren...)
	e.Attribs = append(e.Attribs,
		html.DataAttr("variant", v),
		html.DataAttr("align", a),
	)
	return finish(e, "bubble", Cn(bubbleVariants(map[string]string{"variant": v})))
}

// BubbleContent renders a [Bubble]'s text. It must be the bubble's direct
// child, which the variant colors target.
//
// Upstream's render prop (React useRender) is replaced by the attributes: the
// content is a <div>, an <a> when given an href, and a <button> when given a
// type (pass html.Type("button")), for a bubble the user can click, such as a
// suggested reply; those two get a hover color and a focus ring.
func BubbleContent(attribsChildren ...any) *mx.Element {
	e := html.Div(attribsChildren...)
	e.Name = renderTag(e, "div")
	return finish(e, "bubble-content",
		"w-fit max-w-full min-w-0 overflow-hidden wrap-break-word [button]:text-left [button,a]:transition-colors rounded-xl border border-transparent px-3 py-2 text-sm leading-relaxed [button,a]:outline-none [button,a]:focus-visible:border-ring [button,a]:focus-visible:ring-3 [button,a]:focus-visible:ring-ring/50 group-data-[align=end]/bubble:self-end")
}

// BubbleReactionsSide is the edge of a [Bubble] its [BubbleReactions]
// overlap.
type BubbleReactionsSide string //#enum

const (
	// BubbleReactionsTop puts the reactions over the bubble's top edge.
	BubbleReactionsTop BubbleReactionsSide = "top"
	// BubbleReactionsBottom puts the reactions over the bubble's bottom edge.
	BubbleReactionsBottom BubbleReactionsSide = "bottom"
)

// Valid indicates if b is any of the valid values for BubbleReactionsSide
func (b BubbleReactionsSide) Valid() bool {
	switch b {
	case
		BubbleReactionsTop,
		BubbleReactionsBottom:
		return true
	}
	return false
}

// Validate returns an error if b is none of the valid values for BubbleReactionsSide
func (b BubbleReactionsSide) Validate() error {
	if !b.Valid() {
		return fmt.Errorf("invalid value %#v for type shadcn.BubbleReactionsSide", b)
	}
	return nil
}

// ValidOr returns b if it is any of the valid values for BubbleReactionsSide, else defaultVal
func (b BubbleReactionsSide) ValidOr(defaultVal BubbleReactionsSide) BubbleReactionsSide {
	if b.Valid() {
		return b
	}
	return defaultVal
}

// Enums returns all valid values for BubbleReactionsSide
func (BubbleReactionsSide) Enums() []BubbleReactionsSide {
	return []BubbleReactionsSide{
		BubbleReactionsTop,
		BubbleReactionsBottom,
	}
}

// EnumStrings returns all valid values for BubbleReactionsSide as strings
func (BubbleReactionsSide) EnumStrings() []string {
	return []string{
		"top",
		"bottom",
	}
}

// String implements the fmt.Stringer interface for BubbleReactionsSide
func (b BubbleReactionsSide) String() string {
	return string(b)
}

// bubbleReactionsVariants resolves the reactions' base + side + align
// classes, declared the same way shadcn/ui's bubble.tsx declares them with
// cva.
var bubbleReactionsVariants = cva.New(cva.Config{
	Base: "absolute z-10 flex w-fit items-center justify-center rounded-full ring-3 ring-card bg-muted shrink-0 gap-1 px-1.5 py-0.5 has-[button]:p-0 text-sm",
	Variants: map[string]map[string]string{
		"side": {
			"top":    "top-0 -translate-y-3/4",
			"bottom": "bottom-0 translate-y-3/4",
		},
		"align": {
			"start": "left-3",
			"end":   "right-3",
		},
	},
	DefaultVariants: map[string]string{"side": "bottom", "align": "end"},
})

// BubbleReactions renders a pill of reactions (emoji, counts, or small
// buttons) overlapping an edge of its [Bubble]. side may be "" for
// [BubbleReactionsBottom] and align "" for [ChatAlignEnd]; unknown values
// also resolve to those. The bubble needs room on that side, for example a
// margin, as the pill sits partly outside it.
func BubbleReactions(side BubbleReactionsSide, align ChatAlign, attribsChildren ...any) *mx.Element {
	s, a := string(side.ValidOr(BubbleReactionsBottom)), string(align.ValidOr(ChatAlignEnd))
	e := html.Div(attribsChildren...)
	e.Attribs = append(e.Attribs,
		html.DataAttr("align", a),
		html.DataAttr("side", s),
	)
	return finish(e, "bubble-reactions",
		Cn(bubbleReactionsVariants(map[string]string{"side": s, "align": a})))
}
