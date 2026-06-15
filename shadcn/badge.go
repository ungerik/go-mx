package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn/cva"
)

// BadgeVariant selects a badge's visual style. Class strings are transcribed
// verbatim from shadcn/ui's badge.tsx (new-york-v4, Tailwind v4).
type BadgeVariant string // TODO use go-enum

const (
	// BadgeDefault is the default badge style, using the primary color.
	BadgeDefault BadgeVariant = "default"
	// BadgeSecondary styles the badge with the secondary color.
	BadgeSecondary BadgeVariant = "secondary"
	// BadgeDestructive styles the badge to signal an error or destructive condition.
	BadgeDestructive BadgeVariant = "destructive"
	// BadgeOutline styles the badge with a transparent background and visible border.
	BadgeOutline BadgeVariant = "outline"
)

// badgeVariants resolves a badge's base + variant classes, declared the same
// way shadcn/ui's badge.tsx declares them with cva.
var badgeVariants = cva.New(cva.Config{
	Base: "inline-flex items-center justify-center rounded-md border px-2 py-0.5 text-xs font-medium w-fit whitespace-nowrap shrink-0 [&>svg]:size-3 gap-1 [&>svg]:pointer-events-none focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive transition-[color,box-shadow] overflow-hidden",
	Variants: map[string]map[string]string{
		"variant": {
			"default":     "border-transparent bg-primary text-primary-foreground [a&]:hover:bg-primary/90",
			"secondary":   "border-transparent bg-secondary text-secondary-foreground [a&]:hover:bg-secondary/90",
			"destructive": "border-transparent bg-destructive text-white [a&]:hover:bg-destructive/90 focus-visible:ring-destructive/20 dark:focus-visible:ring-destructive/40 dark:bg-destructive/60",
			"outline":     "text-foreground [a&]:hover:bg-accent [a&]:hover:text-accent-foreground",
		},
	},
	DefaultVariants: map[string]string{"variant": "default"},
})

// BadgeClasses returns the merged base + variant badge class string. It is the
// equivalent of shadcn/ui's exported badgeVariants and mirrors [ButtonClasses]:
// use it to give the badge look to another element. An empty or unknown
// variant resolves to the default.
func BadgeClasses(variant BadgeVariant) string {
	return Cn(badgeVariants(map[string]string{"variant": normBadgeVariant(variant)}))
}

// normBadgeVariant maps an empty or unknown variant to the default.
func normBadgeVariant(v BadgeVariant) string {
	switch v {
	case BadgeSecondary, BadgeDestructive, BadgeOutline:
		return string(v)
	default:
		return string(BadgeDefault)
	}
}

// Badge renders a shadcn/ui badge as a <span>. variant may be "" for the
// default. To render a badge as a link, nest it in an <a> — the [a&]: variant
// classes target exactly that case.
func Badge(variant BadgeVariant, attribsChildren ...any) *mx.Element {
	return finish(html.Span(attribsChildren...), "badge", BadgeClasses(variant))
}
